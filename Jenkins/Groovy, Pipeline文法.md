### 基本的なPipelineの書き方
- https://www.jenkins.io/doc/book/pipeline/syntax/

### List
- 参考URL
  - https://koji-k.github.io/groovy-tutorial/collection/list.html
  - https://gist.github.com/aeg/3741669
  - https://www.tutorialspoint.com/groovy/groovy_lists.htm

- 空のList作成
  - `def <変数名> = []`
- Listに要素追加
  - `<List名>.push("<追加する要素>")`
    - Listの前から追加
  - `<List名>.add("<追加する要素>")`
    - Listの後ろから追加
- Listの要素数確認
  - `<List名>.size()`
- Listの中に特定の文字列を含む要素があるか確認(検索)
  - `<List名>.findAll{it=~/<regex(存在するか確認したい文字列)>/}`
  - 例
    ~~~groovy
    def ERROR_USER_EXIST_FLAG = "false"
    for (error_user_type in ERROR_USER_TYPE_LIST) {
      if (ERROR_USER_MAP.get(error_user_type).findAll{it=~/[a-zA-Z0-9]/}) {
        ERROR_USER_EXIST_FLAG = "true"
      }
    }
    ~~~
  - 参考URL
    - https://gist.github.com/kanemu/433317
    - https://koji-k.github.io/groovy-tutorial/collection/list.html

### Map
- 参考URL
  - https://koji-k.github.io/groovy-tutorial/collection/map.html
  - https://www.tutorialspoint.com/groovy/groovy_maps_get.htm

- 空のMap作成
  - `def <変数名> = [:]`
- MapにKey,Value追加
  - `<Map名>.put("<Key名>","<Value名>")`
- MapからValueを取得
  - `<Map名>.get("<Key名>")`

### 複数の戻り値をやり取りする方法
- 関数を呼び出す側
  - 戻り値を受け取る変数を`()`で囲む必要がある
      ~~~groovy
      (envName, teststr) = GetEnvName3(ENV_MAP,"test")
      ~~~

- 戻り値を返す側
  - 戻り値を`[]`で囲む必要がある
      ~~~groovy
      return [ENVIRONMENT_NAME, teststrresponse]
      ~~~

### Pipelineの中でawkコマンドを使う方法
- 他にもできるかもだけど、私の場合shを`"""`で囲みawkは`''`で囲んで`$`の前にエスケープ文字を入れたらできた
- 例
  ~~~groovy
  def ip_addr = sh(script: """ip a | grep -E 'inet .* eth' | awk '{print \$2}' | cut -d'.' -f 1,2""", returnStdout: true).trim()
  ~~~

### 文字列置換
- `replace`と`replaceAll`両方ともすべての該当する文字列を変換する
  - 例
    ~~~groovy
    a = "HERO HERO"
    println a.replace("ER","ELL") -----> "HELLO HELLO"と出力される
    println a.replaceAll("ER","ELL") --> "HELLO HELLO"と出力される
    ~~~
- `replaceFirst`と`replaceLast`で最初/最後の文字列だけ置換することも可能
- `//`で囲んでregexを使うことも可能
  - 例
    ~~~groovy
    def mphone = "1+555-555-5555"
    result = mphone.replaceFirst(/^1/, "")
    ~~~
  - 参考URL
    - https://stackoverflow.com/questions/9788983/simple-groovy-replace-using-regex

### 文字列から配列に変換
- `split`で区切り文字で分割してList化
- 例
  ~~~groovy
  List ERROR_USRLIST = ERROR_USERLIST.split(",")
  ~~~

### 文字列(`String`)を数字型(`Int`もしくは`Long`)に変換する方法
- `<String>.toInteger()`や`Integer.parseInt(<String>)`などがあるけど、Jenkinsではscriptsecurityに引っかかったりして使えなかった。
- **`<String> as Integer` or `<String> as Long`** は使える
- デカい数字の場合は`Int`の代わりに`Long`を使うこと

### Parametersについて
- https://www.jenkins.io/doc/book/pipeline/syntax/#available-parameters
- JenkinsのGUIにてPipelineのParameterを手動で追加しなくても、  
  Jenkinsfileの方で下記のように`parameters`に定義しておくと自動でPipeline(GUI)に反映される
  ~~~groovy
  pipeline {
    agent {
        kubernetes {
            yaml """
                apiVersion: v1
                kind: Pod
                spec:
                    containers:
                    - name: jnlp
                      image: jenkins/inbound-agent:4.10-3
                    - name: ansible
                      image: nutslove/ansible:2.9
                      imagePullPolicy: Always
                """
            }
    }
    parameters {
        string(name: 'JOB_TYPE', defaultValue: '', description: 'Job Type to exec')
        string(name: 'AWS_ACCOUNT', defaultValue: '', description: 'AWS ACCOUNTS')
        string(name: 'AWS_IAM_USERS', defaultValue: '', description: 'AWS IAM Users')
    }
    stages {
        ・
        ・
    }
  }
  ~~~

### ファイルの扱い
- 参考URL
  - https://www.jenkins.io/doc/pipeline/steps/workflow-basic-steps/#writefile-write-file-to-workspace
  - https://www.jenkins.io/doc/pipeline/examples/#archive-build-output-artifacts

- __ファイル作成__
  - `writeFile`
    1. `writeFile file: "<生成するファイル名>", text: "ファイルに書き込む内容"`
    2. `writeFile(file: "<生成するファイル名>", text: "ファイルに書き込む内容")`
  - 例
    ~~~groovy
    writeFile file: "output/usefulfile.txt", text: "This file is useful, need to archive it."
    writeFile(file: "aws_cli/exec/${POLICY_FILE_NAME}", text: "${REPLACED_IAM_POLICY_STR}")
    ~~~
- __ファイル読み込み__
  - `readFile`
    1. `readFile(file: "<読み込むファイル名>")`
    2. `readFile file: "<読み込むファイル名>"`

### Pipelineの中の`sh`で異常終了について
- Jenkinsは`sh`(shell)をDefaultで`/bin/sh -xe`で実行する
  - `-x`: means to print every command executed
  - `-e`: means to exit with failure if any of the commands in the script failed  
          → つまりRCが0以外の場合は異常終了する
- `sh`を0以外のRCの時も異常終了しないようにする方法と
  1. `set +e`を設定
  2. `command || exit 0` or `command || true`
- 参考URL
  - https://stackoverflow.com/questions/14392349/dont-fail-jenkins-build-if-execute-shell-fails

### postセクションも各stageごとに定義できる