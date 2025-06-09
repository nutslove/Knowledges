# コマンドの実行
- `os/exec`パッケージを使ってOSコマンドを実行し、戻り値・標準出力・標準エラー出力を取得
- `CommandContext`

# signal(e.g. SIGTERM、SIGINT)の受信
- 参考URL
  - https://oohira.github.io/gobyexample-jp/signals.html
- `os/signal`パッケージを使ってUNIXシグナルを受信し、適切な処理を行う
- Go のシグナル通知は、チャネルに`os.Signal`値を`signal.Notify`で送信することで行う
  - これらの通知を受信するためのチャネルを作る
  - このチャネルはバッファリングされることにご注意
- シグナルを受信するためのチャネルを作成し、`signal.Notify`関数を使ってシグナルをチャネルに送信

### サンプルコード1
- `sig := <-sigs`を別のgoroutineで受信することで、メインゴルーチンがブロックされないようにする。もしgoroutineを使わずに直接`sig := <-sigs`を書くと、メインgoroutineがそこで完全にブロックされてしまう。つまり、シグナルが来るまで何も処理できない状態になる。  
```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    sigs := make(chan os.Signal, 1)
    
    // signal.Notify は、指定されたシグナル通知を受信するために、 与えられたチャネルを登録します。
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    done := make(chan bool, 1)

    go func() {
        sig := <-sigs
        fmt.Println()
        fmt.Println(sig)
        done <- true
    }()
   
   // プログラムはシグナルを受信するまで
   // (前述の done に値を送信するゴルーチンで知らされる) 待機した後、終了します。
    fmt.Println("awaiting signal")
    <-done
    fmt.Println("exiting")
}
```
### サンプルコード２
- `wg.Wait()`と`close(ch)`を別goroutineにしている理由は、デットロックを防ぐため。  
- 例えば、以下のように別goroutineにしない場合、何らかの理由で`ch := make(chan CodebaseAnalysisResponse, len(repos))`で作成したチャネルがフルになった場合、goroutine内の `ch <- result` がブロック → goroutineが完了できないため、`wg.Done()` が呼ばれない → メインgorouineは`wg.Wait()`で永遠に待機 → デットロックになる。
```go
// ❌ 危険なパターン
for _, gitRepo := range repos {
    go func(repo string) {
        // ... 処理 ...
        ch <- result // チャネルに結果を送信
    }(gitRepo)
}

wg.Wait()    // すべてのgoroutineの完了を待つ
close(ch)    // チャネルを閉じる

// 結果を収集
for result := range ch {
    // ...
}
```

- サンプルコード  
```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"bytes"
	"fmt"
	"os/exec"
	"os"
	"net/http"
	"time"
	"log/slog"
	"golang.org/x/sync/semaphore"
	"os/signal"
	"syscall"
	"github.com/goccy/go-yaml"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"context"
	"encoding/json"
	"strconv"
	"sync"
)


const (
	WorkDir = "/home/ec2-user/codebase"
)

var (
	ConfigFile = fmt.Sprintf("%s/config.yaml", WorkDir)
	secretName = os.Getenv("SECRETMANAGER_SECRET_ID")
	MaxGoroutines = os.Getenv("MAX_GOROUTINES")
)

type SecretData struct {
	GithubToken string `json:"github_token"`
}

type ConfigMap map[string]map[string]string

type Codebase struct {
	GitRepo string
	Branch string
}

type CodebaseAnalysisRequest struct {
	GitRepos []string `json:"git_repos"` // 分析対象のGitリポジトリ
	Branch   string   `json:"branch"`    // 分析対象のブランチ
	Data     string   `json:"data"`      // 分析対象のデータ(エラーログなど)
}

type CodebaseAnalysisResponse struct {
	Response map[string]string `json:"response"` // key: git_repo, value: analysis result
}

func getGithubToken() SecretData {
	config, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion("ap-northeast-1"))
	if err != nil {
		slog.Error("Error occurred while loading AWS config", "error", err)
		return SecretData{}
	}

	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.Background(), input)
	if err != nil {
		slog.Error("Error occurred while getting Github token from AWS Secret Manager", "error", err)
		return SecretData{}
	}

	var secretString string = *result.SecretString
	var secretData SecretData
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		slog.Error("Secret unmarshal error: ", "error", err)
		return SecretData{}
	}
	return secretData
}

func RunCodebaseAnalysis(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel() // すべてのgoroutineにキャンセルを通知して終了させる
	}()

	if secretName == "" {
		slog.Error("Environment variable SECRETMANAGER_SECRET_ID is not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Environment variable SECRETMANAGER_SECRET_ID is not set"})
		return
	}
	var req CodebaseAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("InvalidRequest", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"InvalidRequest": err.Error()})
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}

	response := CodebaseAnalysisResponse{
		Response: make(map[string]string),
	}
	config, configErr := loadConfigFromFile()
	if configErr != nil {
		slog.Error("Failed to load config file", "error", configErr)
		// Configファイル読み込みに失敗した場合は、ローカルのGithubリポジトリを利用する
	}
	startTime := time.Now()
	MaxGoroutines, err := strconv.Atoi(MaxGoroutines)
	if err != nil {
		slog.Error("Failed to convert MaxGoroutines to int", "error", err)
		MaxGoroutines = 2
	}
	sem := semaphore.NewWeighted(int64(MaxGoroutines))
	repos := req.GitRepos

	ch := make(chan CodebaseAnalysisResponse, len(repos))
	var wg sync.WaitGroup
	wg.Add(len(repos))
	
	for _, gitRepo := range repos {
		go func(repo string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				slog.Warn("Context canceled, stopping analysis for repository", "repository", repo)
				ch <- CodebaseAnalysisResponse{
					Response: map[string]string{
						repo: fmt.Sprintf("Context canceled, stopping analysis for repository: %s", repo),
					},
				}
				return
			default:
				if err := sem.Acquire(ctx, 1); err != nil {
					slog.Error("Failed to acquire semaphore", "error", err)
					ch <- CodebaseAnalysisResponse{
						Response: map[string]string{
							repo: fmt.Sprintf("Failed to acquire semaphore: %v", err),
						},
					}
					return
				}
				defer sem.Release(1)
				codebase := Codebase{
					GitRepo: repo,
					Branch: req.Branch,
				}
				
				localHash := config[repo]["hash"]
				remoteHash, err := codebase.getRemoteGitRepoHash()
				if err != nil {
					slog.Error("Failed to get remote git repo hash", "repository", repo, "error", err)
				}

				slog.Info("localHash", "localHash", localHash)
				slog.Info("remoteHash", "remoteHash", remoteHash)
				localHash = "test" // TODO: あとで削除する
				if localHash != remoteHash && !dirExists(fmt.Sprintf("%s/%s", WorkDir, repo)) {
					err := codebase.cloneGitRepo()
					slog.Info("git clone completed", "repository", repo)
					if err != nil {
						slog.Error("Failed to clone git repo", "repository", repo, "error", err)
						ch <- CodebaseAnalysisResponse{
							Response: map[string]string{
								repo: fmt.Sprintf("Failed to clone git repo: %s, error: %v", repo, err),
							},
						}
						return
					}
				} else if localHash != remoteHash && dirExists(fmt.Sprintf("%s/%s", WorkDir, repo)) {
					err := codebase.pullGitRepo()
					slog.Info("git pull completed", "repository", repo)
					if err != nil {
						slog.Error("Failed to pull git repo", "repository", repo, "error", err)
						ch <- CodebaseAnalysisResponse{
							Response: map[string]string{
								repo: fmt.Sprintf("Failed to pull git repo: %s, error: %v", repo, err),
							},
						}
						return
					}
				}

				analysisResult, err := codebase.analyzeCodebase(ctx, req.Data)
				// analysisResult, err := codebase.analyzeCodebase("Servlet.service() for servlet [dispatcherServlet] in context with path [] threw exception [Handler dispatch failed; nested exception is java.lang.OutOfMemoryError: Requested array size exceeds VM limit] with root cause")
				if err != nil {
					slog.Error("Error occurred during codebase analysis", "repository", repo, "error", err)
					ch <- CodebaseAnalysisResponse{
						Response: map[string]string{
							repo: fmt.Sprintf("Error occurred during codebase analysis for repository: %s, error: %v", repo, err),
						},
					}
					return
				}
				ch <- CodebaseAnalysisResponse{
					Response: map[string]string{
						repo: analysisResult,
					},
				}
			}
		}(gitRepo)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	// 安全な結果収集 (Context cancelを考慮)
	collected := 0
	for collected < len(repos) {
		select {
		case <-ctx.Done():
			break
		case result, ok := <-ch:
			if !ok {
				// チャネルが閉じられたら、もう結果がないので、ループを抜ける
				break
			}
			for k, v := range result.Response {
				response.Response[k] = v
			}
			collected++
		}
	}

	slog.Info("Analysis Result", "response", response)
	slog.Info("Total Duration", "duration", time.Since(startTime))
	c.JSON(http.StatusOK, gin.H{"analysis_results": response})
}

func loadConfigFromFile() (ConfigMap, error) {
	yamlFile, err := os.ReadFile(ConfigFile)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return nil, err
	}

	var config ConfigMap
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		slog.Error("Failed to unmarshal config file", "error", err)
		return nil, err
	}
	return config, nil
}

func (codebase *Codebase) getRemoteGitRepoHash() (string, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: fmt.Sprintf("origin-%s", codebase.GitRepo),
		URLs: []string{codebase.GitRepo},
	})
	
	refs, err := remote.List(&git.ListOptions{
		Auth: &gitHttp.BasicAuth{
			Username: "joonkilee", // TODO: あとでGithub Appのものに変更する
			Password: getGithubToken().GithubToken,
		},
	})
	if err != nil {
		return "", err
	}

	lastCommitHash := ""
	for _, ref := range refs {
		if ref.Name().String() == fmt.Sprintf("refs/heads/%s", codebase.Branch) {
			lastCommitHash = ref.Hash().String()
			break
		}
		if ref.Name() == plumbing.HEAD {
			lastCommitHash = ref.Hash().String()
			break
		}
		if ref.Name().String() == "refs/heads/master" {
			lastCommitHash = ref.Hash().String()
			break
		}
	}
	return lastCommitHash, nil
}

func (codebase *Codebase) cloneGitRepo() error {
	_, err := git.PlainClone(fmt.Sprintf("%s/%s", WorkDir, codebase.GitRepo), false, &git.CloneOptions{
		URL: getRepoUrl(codebase.GitRepo),
		Auth: &gitHttp.BasicAuth{
			Username: "joonkilee", // TODO: あとでGithub Appのものに変更する
			Password: getGithubToken().GithubToken,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (codebase *Codebase) pullGitRepo() error {	
	repo, err := git.PlainOpen(fmt.Sprintf("%s/%s", WorkDir, codebase.GitRepo))
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth: &gitHttp.BasicAuth{
			Username: "joonkilee", // TODO: あとでGithub Appのものに変更する
			Password: getGithubToken().GithubToken,
		},
	})
	
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			// すでに最新の場合はエラーではない
			return nil
		}
		return err
	}

	return nil
}

func (codebase *Codebase) analyzeCodebase(ctx context.Context, data string) (string, error) {
	query := fmt.Sprintf("Is there any code in this repository related to the following error? If so, identify the relevant code portions and propose where and how to modify the code to fix the error.\nError: %s", data)
	cmd := exec.CommandContext(ctx, "claude", "--verbose", "-p", query)
	cmd.Dir = fmt.Sprintf("%s/%s", WorkDir, codebase.GitRepo)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	start := time.Now()
	slog.Info("Starting analysis", "repository", codebase.GitRepo)
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() != nil {
			slog.Warn("Analysis canceled due to context", "repository", codebase.GitRepo, "duration", duration)
			return "", ctx.Err()
		}
		slog.Error("Analysis error", "stderr", stderr.String(), "error", err)
		return "", err
	}
	slog.Info("Analysis completed", "repository", codebase.GitRepo, "duration", duration)
	
	return stdout.String(), nil
}

func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if !info.IsDir() {
		return false
	}
	return true
}

func getRepoUrl(repo string) string {
	repoUrl := fmt.Sprintf("https://github.com/kinto-dev/%s.git", repo)
	return repoUrl
}
```