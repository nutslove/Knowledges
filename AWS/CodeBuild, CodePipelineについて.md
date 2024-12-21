<!-- TOC -->

- [CodeBuild](#codebuild)
- [CodePipeline](#codepipeline)

<!-- /TOC -->
# CodeBuild
- defaultではCodeBuildはコードリポジトリ(e.g. GitHub)のルートディレクトリから`buildspec.yml`を探して実行する。  
  - 設定で別のファイル名やルートディレクトリ以外のディレクトリ（e.g. `terraform/buildspec/buildspec-plan.yml`など）を指定することもできる
- **GithubリポジトリのPushやPull Requestの生成 / マージなどを契機にCodeBuildをトリガーすることもできる**
  - プロジェクトの「プロジェクトの設定」タブの「ソース」のところの「プライマリソースのウェブフックイベント」で設定できる
  - **そのためには「デフォルトソース認証情報を管理」からGithubに接続設定をしておく必要がある**
  - Githubなどへの接続状況は「設定」タブの「接続」で確認できる
  - **Github側でもアプリをインストールする必要がある**
    - 「Settings」→「Integrations」タブの「Applications」 →「Installed Github Apps」タブ →「Visit Marketplace」にて「AWS Connector for Github」をインストール
    - https://docs.github.com/en/apps/using-github-apps/reviewing-and-modifying-installed-github-apps#navigating-to-the-github-app-you-want-to-review-or-modify


# CodePipeline