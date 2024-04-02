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