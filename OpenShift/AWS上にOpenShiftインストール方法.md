## OpenShiftインストール
- IPIとUPIという２種類のインストール方法がある

### IPI
#### 必要な構成
- ２つのそれぞれのSubnetに１つのPrivate Subnetと１つのPublic Subnetが必要
- OpenShiftで使うドメインはRoute53でドメイン登録が必要
  - https://docs.aws.amazon.com/ja_jp/Route53/latest/DeveloperGuide/domain-register.html#domain-register-procedure-section