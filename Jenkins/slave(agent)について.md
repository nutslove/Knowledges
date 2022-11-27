- Jenkins slave(agent)はjavaが実行できる環境である必要がある（つまりJavaのインストールが必要）

- Jenkins masterからslave(agent)への接続には以下の2つの方法等がある
  1. ssh
  2. JNLP

- JNLPとは
  - Java Network Launch Protocol
  - Java Web Startとも呼ばれる
    - https://cloudbees.techmatrix.jp/blog/struggle-story-about-ci-6/
    - Java Web Startとは (https://www.klab.com/jp/blog/tech/2021/20210805-jenkins.html)
      > Java アプリケーションを Web ブラウザーからワンクリックで起動し、クライアントのマシンに デプロイできるシステムです。 Java Web Start は、ユーザーに面倒なインストール作業を強いることなく、全自動でインストールから起動までを代行してくれます。
  - Kubernetes PluginではJNLP方式が使われる
    - 1つのPod内にJNLPコンテナと実際にエージェントとして使うコンテナが動く(マルチコンテナ)
    - https://plugins.jenkins.io/kubernetes/
  - sshはmasterからagentに接続してagentを起動する方式だが、  
    JNLPはagentの方からmaster(50000port)に接続する方式
    - なのでmaster側にてagentから50000portへの接続を許可する設定が必要