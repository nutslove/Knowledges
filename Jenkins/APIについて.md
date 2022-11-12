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