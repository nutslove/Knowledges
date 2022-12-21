- 参考URL
  - https://www.jenkins.io/doc/book/pipeline/shared-libraries/
  - https://swet.dena.com/entry/2021/01/18/200000

### Jenkins設定
- `Jenkinsの管理` → `システムの設定` → `Global Pipeline Libraries`

|  設定項目名  |  設定内容  |  必須/任意  |  設定例  |  備考  |
| ---- | ---- | ---- | ---- | ---- |
|  Name  |  ライブラリ名.<br>Pipelineでインポートする際に利用  |  必須  |  id-ope  |    |
|  Default version  |  Gitのブランチ名やタグを指定  |  任意  |  master  |    |
|  Retrieval method  |  利用するSCM(Source Code Management)  |  必須  |  Modern SCM  |    |
|  Source Code Management |  利用するSCMサービス  |  必須  |  Git  |    |
|  Project Repository  |  Git URL  |  必須  |  HTTPS/SSH Git URL  |    |
|  Credentials  |  Git認証に使うCredential  |  必須  |  設定済みのCredentialから選択  |    |
|  Behaviors  |  Gitに対する動作  |  必須  |  Discover branches  |    |
|  Library Path (optional)  |  src,resources,varsディレクトリがあるパス  |  任意  |  shared-libraries  |    |

### Directory構造
- src
  - パッケージに分けて独自のクラスを定義可能
- vars
  - pipeline jobで利用可能な変数定義（.groovy）とヘルプファイル（.txt）  
  - xxx.groovyのファイル名を変数として利用可能でありそのファイル内に定義されているメソッドを {ファイル名}.{メソッド名} で呼び出し可能
- resources
  - groovyではないファイル（xxx.jsonやxxx.sh等）を格納  
  - libraryResourceを利用してメソッド内で利用が可能

### 使い方



### vars
- vars内の関数と呼び方に2つのやり方がある
- https://www.jenkins.io/doc/book/pipeline/shared-libraries/#defining-global-variables
  1. __vars側は`def 変数名`で定義し、Jenkinsfileの方で`ファイル名.関数名`で呼び出す__  
       - vars内groovy
        ~~~groovy
        // vars/log.groovy
        def info(message) {
          echo "INFO: ${message}"
        }

        def warning(message) {
          echo "WARNING: ${message}"
        }
        ~~~
       - 呼び出し側
        ~~~groovy
        // Jenkinsfile
        @Library('utils') _

        log.info 'Starting'
        log.warning 'Nothing to do!'
        ~~~
      > **Note**  
      > vars内変数の呼び出しは`script`ブロック内でしか使えない.  
      > 下の例で言うとコメントアウトされている`log.info`はエラーとなる.
      > ~~~groovy
      > @Library('utils') _
      > pipeline {
      >   agent none
      >   stages {
      >     stage ('Example') {
      >        steps {
      >            // log.info 'Starting' 
      >            script { 
      >                log.info 'Starting'
      >                log.warning 'Nothing to do!'
      >            }
      >        }
      >     }
      >   }
      >}
      > ~~~

      > **Note**  
      > `@Library`の後に`_`が必要！

  2. __vars側の関数は`call`固定で、Jenkinsfileの方で`ファイル名`で呼び出す__
      - vars内groovy
        ~~~groovy

        ~~~ 
      - 呼び出し側

### vars内でstageを使う方法
1. vars呼び出し側ですでに`pipeline`、`agent`、`stages`等を定義している場合
   - vars側には`stage`だけ定義
   - `pipeline`や`agent`,`stages`,`steps`等が入るとエラーになる
      ~~~groovy
      def call(AWS_ACCOUNT,AWS_IAM_USERS) {
        stage("AWS IAM User作成") {
          script {

            try {
              def IAM_USERS_LIST =  AWS_IAM_USERS.split(",")

              for ( iam_usr in IAM_USERS_LIST ) {
                IAM_USER_RETURN_CODE = this.steps.sh(script: "aws iam get-user --user-name ${iam_usr} 2> /dev/null 1> /dev/null", returnStatus: true)
                if (IAM_USER_RETURN_CODE != 0){
                  sh """
                    aws iam create-user --user-name ${iam_usr}
                    aws iam create-login-profile --user-name ${iam_usr} --password パスワード --password-reset-required
                    aws iam add-user-to-group --user-name ${iam_usr} --group-name Default-Policy-Group
                  """
                  println "IAM USER[" + iam_usr + "] is created"
                } else {
                  println "IAM USER[" + iam_usr + "] already exist"
                }
              }
            } catch (e) {
              throw e
            }

          }
        }
      }
      ~~~

2. vars呼び出し側に何もなく、関数のみで呼び出している場合
   - https://www.jenkins.io/doc/book/pipeline/shared-libraries/#defining-declarative-pipelines
   - `pipeline`から`agent`,`stages`,`steps`等、すべてvars側に定義

### vars内で`node`で実行エージェントを指定する方法
- `node(<ノード名>) { script { ・・・処理・・・ } }`
- 例
  ~~~groovy
  def call(envConfigInformaion) {
    stage("TEST"){
      node('cicd') {
        script {
          try {

          ・・・処理・・・

          } catch (e) {
            throw e
          }
        }
      }
    }
  }
  ~~~
- 参考URL
  - https://www.jenkins.io/blog/2020/10/21/a-sustainable-pattern-with-shared-library/