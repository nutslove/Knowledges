- APIにはmodern API (from Jenkins 2.138.1) とlagecy API (to Jenkins 2.138.1) の2種類がある
  - 最新バージョンのJenkinsでもlagecy APIは利用可能だけどmodern APIを使うことが推奨されている
  - 参考URL
    - https://docs.cloudbees.com/docs/cloudbees-ci-kb/latest/client-and-managed-masters/how-to-generate-change-an-apitoken
- APIはJenkinsのGUI上で作成することもできるが、`Jenkinsの管理` - `スクリプトコンソール`にて以下を直接実行することでも作成できる  
  → 画面の結果にAPI Tokenが表示されるので押さえておくこと！
  ~~~groovy
  import hudson.model.*
  import jenkins.model.*
  import jenkins.security.*
  import jenkins.security.apitoken.*

  // script parameters
  def userName = 'admin'
  def tokenName = 'kb-token'

  def user = User.get(userName, false)
  def apiTokenProperty = user.getProperty(ApiTokenProperty.class)
  def result = apiTokenProperty.tokenStore.generateNewToken(tokenName)
  user.save()

  return result.plainValue
  ~~~

## APIでJobをキックする時、Jobにパラメータを渡す方法
- Jobにパラメータを渡す方法はいくつかある
  1. json形式で渡す
     - 例
       ~~~
       curl -i -X POST -u "<USER>:<TOKEN>" http://localhost:8080/job/<JOB名>/build?delay=0sec --data-urlencode json='{"parameter": [{"name":"id", "value":"123"}, {"name":"verbosity", "value":"high"}]}'
       ~~~
     - **Jobで定義されているすべてのパラメータを渡さないとJobがじっこうされない。**  
       **なので使わないパラメータでもJobで定義されているパラメータはvalueを空欄にして渡さないといけない。**
  2. クエリ文字列(URLパラメータ)で渡す
     - 例
       ~~~
       curl -i -X POST -u "<USER>:<TOKEN>" http://localhost:8080/job/<JOB名>/buildWithParameters?<パラメータ名>=<値>&<パラメータ名>=<値>[<パラメータ名>=<値>,・・・]&delay=0sec"
       ~~~
     - **Jobで定義されているすべてのパラメータを渡さなくてもエラーにならず実行される**
- その他にも色々あるっぽい
  - jsonファイルを指定する方法 (testapi.jsonがローカルにあるとした場合)
    - e.g. `curl -v –user apiuser:apitoken -X POST http://10.10.10.100:8080/job/testjob/build –data “@testapi.json” -H “Accept: application/json”`
  - `--data`で1つずつパラメータを指定する方法
    - `curl JENKINS_URL/job/JOB_NAME/buildWithParameters --user USER:TOKEN --data id=123 --data verbosity=high`
  - `--form`でパラメータが定義されているファイルごとに渡す方法
    - `curl JENKINS_URL/job/JOB_NAME/buildWithParameters --user USER:PASSWORD --form FILE_LOCATION_AS_SET_IN_JENKINS=@PATH_TO_FILE`
- 参考URL
  - https://www.jenkins.io/doc/book/using/remote-access-api/
  - https://wiki.jenkins.io/display/JENKINS/Remote+access+API
  - https://beginnersforum.net/blog/2019/11/28/jenkin-paramerized-job-api-json/