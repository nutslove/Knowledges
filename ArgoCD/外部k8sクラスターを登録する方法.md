- 外部のEKSクラスターを登録するための`Secret`の例
  ```yaml
  apiVersion: v1
  kind: Secret
  metadata:
    name: <任意のSecret名>
    namespace: argocd
    labels:
      argocd.argoproj.io/secret-type: cluster
  type: Opaque
  stringData:
    name: eks-cluster
    server: <EKSクラスターのAPIサーバのURL>
    config: |
      {
        "bearerToken": "<対象クラスター上のServiceAccountのトークン>",
        "tlsClientConfig": {
          "insecure": false,
          "caData": "<クラスターと安全に通信するためのSSL証明書>"
        }
      }
  ```
  - `metadata.labels`の`argocd.argoproj.io/secret-type: cluster`は、
    ArgoCDがSecretをクラスター情報として認識するために、このラベルを設定することが推奨されている
  - `stringData.name`には任意の値を指定できる（EKSクラスター名と一致しなくても良い）
  - `bearerToken`にはターゲットクラスター上のすべての権限(Role)を持っているServiceAccountのSecretで作成されるトークンの値を指定
  - `caData`にはEKSクラスターと暗号化通信をするための証明書で、kubeconfig(`~/.kube/config`)の`certificate-authority-data`の値を指定