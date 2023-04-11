- URL
  - https://plugins.jenkins.io/ansicolor/
  - https://stackoverflow.com/questions/67832590/nice-way-to-display-terraform-plan-output-in-jenkins
  - https://plugins.jenkins.io/ansible/

## Jenkinsパイプライン上でAnsible実行結果に色を付ける方法
- AnsiColorプラグインをインストール
- `environment{}`内に`ANSIBLE_FORCE_COLOR = true`を設定し、`options { ansiColor('xterm') }`を追加
- 例
  ~~~groovy
  pipeline {
    environment {
      HTTP_PROXY = 'http://xx.xx.xx.xx:xx'
      HTTPS_PROXY = 'http://xx.xx.xx.xx:xx'
      ANSIBLE_FORCE_COLOR = true
    }
  }
  options {
    ansiColor('xterm')
  }
  aent {
    ・
    ・
  }
  stages {
    ・
    ・
  }
  ~~~

## Jenkinsパイプライン上でTerraform実行結果に色を付ける方法
- AnsiColorプラグインをインストール
- Pipelineで`options { ansiColor('xterm') }`を入れるだけ