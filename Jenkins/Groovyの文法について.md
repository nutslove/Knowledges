##### 部分文字列検索
-  


##### 複数の戻り値をやり取りする方法
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

##### Pipelineの中でawkコマンドを使う方法
- 他にもできるかもだけど、私の場合shを`"""`で囲みawkは`''`で囲んで`$`の前にエスケープ文字を入れたらできた
- 例
  ~~~groovy
  def ip_addr = sh(script: """ip a | grep -E 'inet .* eth' | awk '{print \$2}' | cut -d'.' -f 1,2""", returnStdout: true).trim()
  ~~~

##### Parametersについて
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