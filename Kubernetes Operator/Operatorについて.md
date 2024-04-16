- `Custom Resource`と`Custom Controller`で構成されている

## ■ `Custom Resource`
- APIの拡張
- `CustomResourceDefinition`（CRD）で定義
- `CustomResourceDefinition`のマニフェストファイルをkubectlなどで適用すると、
  kube-apiserverに新しいAPIとして登録され、`kubectl get <CRDで定義したリソース名>`などが打てるようになる

## ■ `Custom Controller`
- kube-apiserverを通じてetcd上の`Custom Resource`を実態とあるべき姿を監視して、実態とあるべき姿に差分がある場合、あるべき姿になるように制御する

## Custom Operatorの開発
- `Operator SDK`と`KubeBuilder`２つのフレームワークがある

## Operator SDK
- Operator Frameworkのコンポーネント
- Go/Ansible/Helm、３つの開発方法がある
  - https://sdk.operatorframework.io/docs/overview/#workflow
- install
  - https://sdk.operatorframework.io/docs/installation/
