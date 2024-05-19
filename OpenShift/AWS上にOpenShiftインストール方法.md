# OpenShiftインストール方式
- IPIとUPIという２種類のインストール方法がある

## IPI
### 必要な構成
- ２つのそれぞれのSubnetに１つのPrivate Subnetと１つのPublic Subnetが必要
- OpenShiftで使うドメインはRoute53でドメイン登録が必要
  - https://docs.aws.amazon.com/ja_jp/Route53/latest/DeveloperGuide/domain-register.html#domain-register-procedure-section

### OpenShiftクラスターインストール
- `install-config.yaml`を用意したうえで、`openshift-install create cluster`コマンドを実行
  - アンインストールは`openshift-install destroy cluster`

#### インストール後、Ingress証明書の更新
- 作成した証明書を使ってtlsのsecretを作成
  ```shell
  oc create secret tls router-certs --cert=${CERTDIR}/fullchain.pem --key=${CERTDIR}/key.pem -n openshift-ingress
  ```
- 作成したSecretをdefault ingresscontrollerの`spec.defaultCertificate`フィールドに追加
  ```shell
  oc patch ingresscontroller default -n openshift-ingress-operator --type=merge --patch='{"spec": { "defaultCertificate": { "name": "router-certs" }}}'
  ```