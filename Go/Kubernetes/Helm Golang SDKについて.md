## 代表的なPackage
- https://helm.sh/ko/docs/topics/advanced/#go-sdk  
  > This is a list of the most commonly used packages with a simple explanation about each one:
  >
  > - **`pkg/action`**: Contains the main “client” for performing Helm actions. This is the same package that the CLI is using underneath the hood. If you just need to perform basic Helm commands from another Go program, this package is for you
  > - **`pkg/{chart,chartutil}`**: Methods and helpers used for loading and manipulating charts
  > - **`pkg/cli`** and its subpackages: Contains all the handlers for the standard Helm environment variables and its subpackages contain output and values file handling
  > **`pkg/release`**: Defines the Release object and statuses

## 利用の流れ
- `cli.New()`で`*EnvSettings`をインスタンス化
- `settings.SetNamespace("<対象namespace>")`でターゲットnamespaceを指定
- `*EnvSettings.RepositoryConfig`でリポジトリファイルのパスを設定後、リポジトリファイルを読み込む
  - *リポジトリファイル*はHelmリポジトリのエントリを管理するためのYAMLファイル
    - `HELM_REPOSITORY_CONFIG`環境変数で設定したパス、なければ`$HOME/.config/helm/repositories.yaml`がパスとして設定される
    - リポジトリファイルの例  
      ```yaml
      apiVersion: v1
      generated: "2021-08-01T12:00:00Z"
      repositories:
        - name: bitnami
          url: https://charts.bitnami.com/bitnami
        - name: opensearch
          url: https://opensearch-project.github.io/helm-charts/
      ```
- リポジトリファイルがない場合は`repo.NewFile()`でリポジトリファイルの構造体を新規作成し、すでにリポジトリファイルがある場合は`repo.LoadFile()`でリポジトリファイルを読み込む
- `repo.NewChartRepository()`でチャートリポジトリを作成し、`DownloadIndexFile()`でインデックスファイルをダウンロードする
  - *チャートリポジトリ*
    - Helmチャートが保存されている場所で、Helmクライアントは特定のリポジトリからチャートを検索してダウンロードすることができる。  
      リポジトリは通常、HTTPサーバー(e.g. Github)を介してホスティングされ、複数のチャートとそのバージョンを含むインデックスファイルを提供する
    - `helm` CLIの場合の例  
      `helm repo add opensearch https://opensearch-project.github.io/helm-charts/`
  - *インデックスファイル*
    - チャートリポジトリ内のチャートのメタデータを含むファイルで、Helmクライアントは、インデックスファイルを使用してチャートリポジトリ内のチャートを検索し、必要なチャートをダウンロードする
    - ダウンロードされたインデックスファイルは通常`$HOME/.cache/helm/repository/`配下にある
    - インデックスファイルの例
      ```yaml
      apiVersion: v1
      entries:
        opensearch:
        - apiVersion: v2
          appVersion: 2.14.0
          created: "2024-05-14T22:54:27.269895895Z"
          description: A Helm chart for OpenSearch
          digest: e09b9d82cf94ef43eebba4c71fe0a1aeef899b6a5bde6fa1c8454fe2d8ea2406
          home: https://opensearch.org
          maintainers:
          - name: DandyDeveloper
          - name: bbarani
          - name: gaiksaya
          - name: peterzhuamazon
          - name: prudhvigodithi
          - name: TheAlgo
          name: opensearch
          sources:
          - https://github.com/opensearch-project/opensearch
          - https://github.com/opensearch-project/helm-charts
          type: application
          urls:
          - https://github.com/opensearch-project/helm-charts/releases/download/opensearch-2.20.0/opensearch-2.20.0.tgz
          version: 2.20.0
        - apiVersion: v2
          appVersion: 2.13.0
          created: "2024-04-03T03:57:48.493940954Z"
          description: A Helm chart for OpenSearch
          digest: 61b5d932d64b738a49bef2ba5783d61893c296b165a81d1b95d3514d68f02500
          home: https://opensearch.org
          maintainers:
          - name: DandyDeveloper
          - name: bbarani
          - name: gaiksaya
          - name: peterzhuamazon
          - name: prudhvigodithi
          - name: TheAlgo
          name: opensearch
          sources:
          - https://github.com/opensearch-project/opensearch
          - https://github.com/opensearch-project/helm-charts
          type: application
          urls:
          - https://github.com/opensearch-project/helm-charts/releases/download/opensearch-2.19.0/opensearch-2.19.0.tgz
          version: 2.19.0
      generated: "2024-06-09T00:45:48.12533375Z"
      ```
    - `helm` CLIでインデックスファイルの更新  
      `helm repo update`
  - チャートリポジトリとインデックスファイルの関係 (`*.tgz`が実際のチャートファイル)
    ```
    チャートリポジトリ (e.g. https://example.com/charts/)
      ├── index.yaml
      ├── chart1-1.0.0.tgz
      ├── chart1-1.1.0.tgz
      ├── chart2-2.0.0.tgz
      └── chart2-2.1.0.tgz
    ```

#### サンプルコード  
```go
package utilities

import (
  "fmt"
  "log"
  "os"
  "path/filepath"

  "helm.sh/helm/v3/pkg/action"
  "helm.sh/helm/v3/pkg/chart"
  "helm.sh/helm/v3/pkg/chart/loader"
  "helm.sh/helm/v3/pkg/cli"
  "helm.sh/helm/v3/pkg/getter"
  "helm.sh/helm/v3/pkg/repo"
)

var (
  repoEntry = &repo.Entry{
    Name: "opensearch",
    URL:  "https://opensearch-project.github.io/helm-charts/",
  }
)

func OpenSearchHelmSetting(releaseName string, actionType string) (*action.Install, *action.Uninstall, *chart.Chart) {
  // Helm CLI設定の取得
  settings := cli.New()
  // settings.Debug = true

  // Namespaceを設定
  settings.SetNamespace("opensearch")

  // リポジトリファイルのパスを設定
  repoFile := settings.RepositoryConfig

  // リポジトリファイルを読み込むか、存在しない場合は新規作成する
  var r *repo.File
  if _, err := os.Stat(repoFile); os.IsNotExist(err) {
    r = repo.NewFile()
  } else {
    var err error
    r, err = repo.LoadFile(repoFile)
    if err != nil {
      log.Fatalf("Failed to load repo file: %v", err)
    }
  }

  // リポジトリファイルにリポジトリを追加する
  if !r.Has(repoEntry.Name) {
    r.Update(repoEntry)

    // リポジトリファイルを保存するディレクトリを作成
    if err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm); err != nil {
      log.Fatalf("Failed to create directory for repo file: %v", err)
    }

    // リポジトリファイルを保存する
    if err := r.WriteFile(repoFile, 0644); err != nil {
      log.Fatalf("Failed to write repo file: %v", err)
    }
  }

  // チャートリポジトリを作成し、インデックスファイルをダウンロードする
  chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(settings))
  if err != nil {
    log.Fatalf("Failed to create new chart repository: %v", err)
  }
  _, err = chartRepo.DownloadIndexFile()
  if err != nil {
    log.Fatalf("Failed to download index file: %v", err)
  }

  // Helm設定の初期化
  actionConfig := new(action.Configuration)
  if err := actionConfig.Init(settings.RESTClientGetter(), "opensearch", "secret", func(format string, v ...interface{}) {
    fmt.Sprintf(format, v...)
  }); err != nil {
    log.Fatalf("Failed to initialize Helm configuration: %v", err)
  }

  var installClient *action.Install
  var uninstallClient *action.Uninstall
  if actionType == "install" {
    installClient = action.NewInstall(actionConfig)
    // インストールクライアントの設定
    installClient.Namespace = "opensearch"
    installClient.ReleaseName = releaseName
    installClient.CreateNamespace = true
    installClient.Wait = true
    installClient.Timeout = 900
    uninstallClient = nil
  } else if actionType == "uninstall" {
    uninstallClient = action.NewUninstall(actionConfig)
    installClient = nil
  }

  // チャートのパスを見つける
  chartName := "opensearch/opensearch"
  chartPathOptions := action.ChartPathOptions{}
  chartPath, err := chartPathOptions.LocateChart(chartName, settings)
  if err != nil {
    log.Fatalf("Failed to locate chart: %v", err)
  }

  // チャートをロードする
  chart, err := loader.Load(chartPath)
  if err != nil {
    log.Fatalf("Failed to load chart: %v", err)
  }

  return installClient, uninstallClient, chart
}
```