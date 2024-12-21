# 目次 <!-- omit in toc -->
<!-- TOC -->
- [CodeBuild](#codebuild)
- [CodePipeline](#codepipeline)
<!-- /TOC -->

# CodeBuild
- defaultではCodeBuildはコードリポジトリ(e.g. GitHub)のルートディレクトリから`buildspec.yml`を探して実行する。  
  - 設定で別のファイル名やルートディレクトリ以外のディレクトリ（e.g. `terraform/buildspec/buildspec-plan.yml`など）を指定することもできる


# CodePipeline