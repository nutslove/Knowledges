## Project
- 関連するリソースとユーザーをグループ化するための主要な組織単位
- Kubernetesのnamespaceに相当するものだが、追加の機能が提供されている
  - namespaceを拡張した感じ

### Projectの特徴
- **リソースの分離**
  - Projectはリソース (e.g. Pod、Service、Deployment、ConfigMap、Secret、・・・) を分離し、複数のユーザーやチームが同じクラスタ上で作業できるようにしている。各Projectは独立しており、他のProjectのリソースにはアクセスできない。
- **アクセス制御**
  - OpenShiftでは、Projectごとにアクセス権を細かく設定できる。これにより、特定のユーザーに対してProject内のリソースへのアクセスを制限したり、特定の操作を許可したりすることができる。
- **クォータとリミット**
  - 管理者は、Projectごとにリソース使用量のクォータを設定できる。これにより、Projectが使用できるリソースの最大量を制御し、クラスターのリソースを適切に分配することができる。
- **ラベリングとアノテーション**
  - Projectにはラベルやアノテーションを付けることができ、これにより、リソースの管理や識別を簡単に行うことができる。
- **ネットワークの分離**
  - **デフォルト**では、OpenShiftはすべてのProject間で完全なネットワーク分離を提供し、**異なるProject間の通信はできない**。
    - Network Policyではなく、OpenShiftのSDNによって定義されているらしい
  - `Network Policy`で異なるProject間の通信を許可することもできる
    - https://www.redhat.com/en/blog/network-policies-controlling-cross-project-communication-on-openshift

## User
- ユーザには3つのタイプがある
1. Regular
   - Project (namespace) に割り当てられ、そのProject内でアプリケーションのデプロイや管理を行うことができるユーザ
   - 通常はクラスター管理者によって作成される
2. System
   - OpenShiftのインフラコンポーネントや内部サービスによって使用される特別なユーザーアカウント
   - `system:` Prefixが付く（e.g. `system:admin`）
3. ServiceAccount
   - `system:serviceaccount:` Prefixが付く（e.g. `system:serviceaccount:monitoring:scrapinguser`）
     - `monitoring`の部分はnamespace

## OAuth Server
- OpenShiftクラスターにはOAuth Serverが含まれていて、ユーザがOpenShiftを操作したいときの流れは以下の通り。
  1. ユーザが最初にOAuthサーバにアクセストークンを要求する
  2. OAuthサーバが連携しているIdP (Identity Provider) にそのユーザが登録されているか確認を行う
  3. 確認出来たらユーザにアクセストークンを発行する
  4. ユーザは発行されたアクセストークンを使ってAPI Serverにリクエストを送る
### OpenShift4のOAuthで使えるIdP一覧
- HTPasswd
  - ユーザ名とパスワードが記載されている`.htpasswd`ファイルを利用
- Keystone
  - OpenStack Keystone v3サーバを利用
- LDAP
- Basic認証
- Request Header
  - `X-Remote-User`などのリクエストヘッダー値を利用
- OpenID Connect
- etc.

#### OAuthとは
- https://qiita.com/TakahikoKawasaki/items/e37caf50776e00e733be

## Openshift API Server
- OpenShiftはKubernetesから拡張されており、Kubernetesは持っていないAPIも備えている
- OpenShiftはkube-apiserverとは別で、openshift-apiserverも持っていて、OpenShiftにしかないリソースに対するAPIはopenshift-apiserverに送られる

## Route
- OpenShift固有の概念で、外部トラフィックをOpenShiftクラスタ内のServiceにルーティングするためのResource
- クラスタ外から特定のServiceにアクセスするための(インターネットからアクセス可能な)パブリックURLが提供される
- 以下の特徴がある
  - **パブリックDNS名の割り当て**
    - Routeを作成すると、指定したサービスにアクセスするためのパブリックDNS名が割り当てられる
  - **TLS/SSLサポート**
  - **パスベースのルーティング**
    - 特定のパスへのリクエストを特定のServiceにルーティングすることが可能。これにより、同じドメイン名を使用して複数のServiceにアクセスすることができる
  - **ロードバランシング**
    - 複数のPod間でトラフィックを分散させることができる

## ImageStream
- OpenShift独自のAPIの１つで、コンテナイメージの参照を抽象化し、バージョン管理、自動化、アクセス制御、共有を容易にする機能
- ImageStreamがコンテナイメージを持っているわけではなく、あくまでコンテナイメージレジストリへのポインタを持っている感じ
#### ImageStreamの特徴
1. **イメージの抽象化**
   - ImageStreamは、実際のコンテナイメージを指す論理的なポインタのようなもの。これにより、実際のイメージが変更されても、Image Streamを参照しているDeploymentsやDeploymentConfigなどは変更する必要がない。
2. **イメージのバージョン管理**
   - ImageStreamを使用すると、1つのイメージに対して複数のバージョン(タグ)を管理できる。例えば、「latest」や「v1.0」などのタグを付けることができる。
3. **イメージの自動ビルドとデプロイ**
   - ImageStreamは、GitリポジトリやDockerfileの変更を検知し、自動的にビルドとデプロイを行うように設定できる。これにより、継続的インテグレーション・デリバリー(CI/CD)のプロセスを簡素化できる。
4. **イメージのアクセス制御**
   - ImageStreamを使用すると、プロジェクト内でのイメージのアクセス制御を行うことができる。例えば、特定のユーザーやグループにのみイメージの読み取りや書き込み権限を与えることができる。
5. **クラスター内でのイメージ共有**
   - ImageStreamを使用すると、クラスター内の他のプロジェクトからイメージを参照できる。これにより、イメージの共有と再利用が容易になる。
#### ImageStreamの例
```yaml
apiVersion: image.openshift.io/v1
kind: ImageStream
metadata:
  name: my-app --> ImageStreamの名前
spec:
  lookupPolicy:
    local: false
  tags: --> spec.tagsにImageStreamが追跡するタグを定義
    - name: latest
      from: --> 参照する実際のDockerイメージを指定
        kind: DockerImage
        name: docker.io/my-org/my-app:latest
      importPolicy:
        scheduled: true
    - name: v1.0
      from:
        kind: DockerImage
        name: docker.io/my-org/my-app:v1.0
      importPolicy:
        scheduled: true
    - name: v2.0
      from:
        kind: DockerImage
        name: docker.io/my-org/my-app:v2.0
      importPolicy:
        scheduled: true --> importPolicy.scheduledをtrueにすることで、定期的にDockerイメージの変更を確認し、ImageStreamを更新するようにしている
```
- 上記と以下の`Deployment`を組み合わせることで`my-app:latest`イメージの中身が変わったときに自動的にDeploymentのPodが作り直される
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: my-app
  spec:
    template:
      spec:
        containers:
          - name: my-app
            image: my-app:latest --> ★直接コンテナイメージ(docker.io/my-org/my-app:latest)を指定するのではなく、ImageStreamの<metadata.name>:<spec.tags.name>を指定
  ```