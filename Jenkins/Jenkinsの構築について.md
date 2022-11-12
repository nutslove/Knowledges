- `/usr/share/jenkins/ref/init.groovy.d/`内にJenkins起動時に組み込みたい設定(初期ユーザ作成等)を記載したgroovyファイルを置けばJenkins起動時に反映される
  - ユーザ作成の例  
    ※(Dockerfileの)環境変数に初期ユーザのIDを`JENKINS_USER`に,PWを`JENKINS_PASS`に設定した前提
    ~~~groovy
    import hudson.model.*
    import jenkins.model.*
    import hudson.security.*
    import jenkins.security.apitoken.*

    def env = System.getenv()

    def jenkins = Jenkins.getInstance()
    if(!(jenkins.getSecurityRealm() instanceof HudsonPrivateSecurityRealm))
        jenkins.setSecurityRealm(new HudsonPrivateSecurityRealm(false))

    if(!(jenkins.getAuthorizationStrategy() instanceof GlobalMatrixAuthorizationStrategy))
        jenkins.setAuthorizationStrategy(new GlobalMatrixAuthorizationStrategy())

    def user = jenkins.getSecurityRealm().createAccount(env.JENKINS_USER, env.JENKINS_PASS)
    user.save()

    jenkins.getAuthorizationStrategy().add(Jenkins.ADMINISTER, env.JENKINS_USER)

    jenkins.save()
    ~~~

- JenkinsをEKS上にPodとして構築する場合、Jenkins HomeとしてEFSを使うことでJenkinsをAZレベル冗長構成にすることができる
  - https://aws.amazon.com/jp/blogs/storage/deploying-jenkins-on-amazon-eks-with-amazon-efs/
  - https://medium.com/@CloudifyOps/setting-up-jenkins-high-availability-with-efs-in-the-backend-ce970c55da87
- API Tokenを利用したAPIの実行もNLB経由でできることを実機確認済