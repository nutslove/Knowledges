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