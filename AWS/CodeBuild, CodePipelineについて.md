<!-- TOC -->

- [CodeBuild](#codebuild)
  - [`buildspec.yml`の構成](#buildspecymlの構成)
- [CodePipeline](#codepipeline)

<!-- /TOC -->
# CodeBuild
- ビルドのたびにCodeBuildがDockerコンテナを作って、その中でビルドする
- defaultではCodeBuildはコードリポジトリ(e.g. GitHub)のルートディレクトリから`buildspec.yml`を探して実行する。  
  - 設定で別のファイル名やルートディレクトリ以外のディレクトリ（e.g. `terraform/buildspec/buildspec-plan.yml`など）を指定することもできる
- **GithubリポジトリのPushやPull Requestの生成 / マージなどを契機にCodeBuildをトリガーすることもできる**
  - プロジェクトの「プロジェクトの設定」タブの「ソース」のところの「プライマリソースのウェブフックイベント」で設定できる
  - **そのためには「デフォルトソース認証情報を管理」からGithubに接続設定をしておく必要がある**
  - Githubなどへの接続状況は「設定」タブの「接続」で確認できる
  - **Github側でもアプリをインストールする必要がある**
    - 「Settings」→「Integrations」タブの「Applications」 →「Installed Github Apps」タブ →「Visit Marketplace」にて「AWS Connector for Github」をインストール
    - https://docs.github.com/en/apps/using-github-apps/reviewing-and-modifying-installed-github-apps#navigating-to-the-github-app-you-want-to-review-or-modify
- 実行するOSを選択できる（e.g. Amazon Linux、Ubuntu）
- デフォルトのDockerイメージの代わりにカスタムDockerイメージを使うこともできる
- ビルドするコンピューティングリソースを選択できる（e.g. 3GBメモリ/2vCPU、7GBメモリ/4vCPUなど）
- デフォルトではCodeBuildはVPC内のリソースにアクセスできないが、設定すればアクセスできる
  - https://docs.aws.amazon.com/codebuild/latest/userguide/vpc-support.html
- artifact(生成物)をS3にアップロードできる

## `buildspec.yml`の構成
- https://docs.aws.amazon.com/ja_jp/codebuild/latest/userguide/build-spec-ref.html
- 全体例  
  ```yaml
  version: 0.2 # 必須。バージョンを指定 (現在は 0.2 を使用) 

  run-as: Linux-user-name # 必須ではない。実行ユーザ（Linux）

  env: # 必須ではない
    shell: bash | /bin/sh
    variables: # 環境変数
      key: "value"
      key: "value"
    parameter-store: # Systems Manager Parameter Storeに保存されている保存されているカスタム環境変数を取得
      key: "value"
      key: "value"
    exported-variables: 
      - variable
      - variable
    secrets-manager: # Secrets Managerに保存されているカスタム環境変数を取得
      key: secret-id:json-key:version-stage:version-id
    git-credential-helper: no | yes

  proxy: # 必須ではない
    upload-artifacts: no | yes
    logs: no | yes

  batch: # 必須ではない（https://docs.aws.amazon.com/ja_jp/codebuild/latest/userguide/batch-build-buildspec.html）
    fast-fail: false | true # 1つ以上のビルドタスクが失敗した場合のバッチビルドの動作を指定（falseの場合途中で失敗してもすべてのビルドが完了する）
    # build-list:
    # build-matrix:
    # build-graph:
          
  phases:
    install: # 必須ではない。依存パッケージのインストールやセットアップ
      run-as: Linux-user-name
      on-failure: ABORT | CONTINUE # 必須ではない。フェーズ中に障害が発生した場合に実行するアクションを指定
      runtime-versions: # ランタイムバージョンの指定
        runtime: version
        runtime: version
      commands:
        - command
        - command
      finally:
        - command
        - command
      
    pre_build: # 必須ではない。ビルドの前に実行するコマンドがある場合記述
      run-as: Linux-user-name
      on-failure: ABORT | CONTINUE
      commands:
        - command
        - command
      finally:
        - command
        - command
      
    build: # 必須。ビルドの実行手順
      run-as: Linux-user-name
      on-failure: ABORT | CONTINUE
      commands:
        - command
        - command
      finally:
        - command
        - command
      
    post_build: # 必須ではない。ビルド後に実行するコマンドがある場合記述
      run-as: Linux-user-name
      on-failure: ABORT | CONTINUE
      commands:
        - command
        - command
      finally:
        - command
        - command
      
  reports: # 必須ではない。レポートの設定
    report-group-name-or-arn:
      files: # レポートによって生成されたテスト結果の生データを含む場所を表す
        - location
        - location
      base-directory: location
      discard-paths: no | yes
      file-format: report-format
  artifacts: # 必須ではない。ビルド成果物の設定（S3にアップロードする）
    files:
      - location
      - location
    name: artifact-name
    discard-paths: no | yes
    base-directory: location
    exclude-paths: excluded paths
    enable-symlinks: no | yes
    s3-prefix: prefix
    secondary-artifacts:
      artifactIdentifier:
        files:
          - location
          - location
        name: secondary-artifact-name
        discard-paths: no | yes
        base-directory: location
      artifactIdentifier:
        files:
          - location
          - location
        discard-paths: no | yes
        base-directory: location
  cache: # 必須ではない。キャッシュの設定
    paths:
      - path
      - path
  ```

- `runtime-versions`に複数のランタイムを指定できる
  - 例
    ```yaml
    phases:
      install:
        runtime-versions:
          java: corretto8
          python: 3.x
          ruby: "$MY_RUBY_VAR"
    ```

# CodePipeline