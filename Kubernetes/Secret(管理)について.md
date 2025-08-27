- k8s `Secret`はBase64エンコーディングなので簡単にデコーディングできてしまうため、別のツールと組み合わせてSecretを安全に管理する必要がある

## Secretについて
- https://kubernetes.io/ja/docs/concepts/configuration/secret/
- PodがSecretを参照させる場合、PodのSpecに `volumes` としてSecretをマウントする方法と、`env` として環境変数で参照する方法がある
  - `volumes` としてマウントする場合、Pod内のファイルとしてSecretの内容が利用できる
  - `env` として参照する場合、Pod内の環境変数としてSecretの内容が利用できる

> [!CAUTION]  
> `volumes`としてマウントする場合は、Secretが更新されたときに自動的にPod内に反映されるが、`env`として参照させている場合は自動的にPodに反映されず、Podの再起動が必要！

## Secret管理ツール
1. HashiCorp Vault
2. Secrets Store CSI Driver
   - https://secrets-store-csi-driver.sigs.k8s.io/
   - KubernetesのCSI（Container Storage Interface）に基づいたプロジェクトで、外部の秘密ストア（Vault、Azure Key Vault、Google Secret Managerなど）から秘密情報を取得し、KubernetesのSecretsとして利用することができる。
     - https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/manage-secrets.html
   - これにより、KubernetesのSecretsを直接使用する代わりに、外部のセキュリティが強化された秘密ストアを活用できる。
3. Kubernetes External Secrets
   - https://external-secrets.io/latest/
   - https://github.com/external-secrets/external-secrets
   - Kubernetes外部の秘密管理システム（例：AWS Secrets Manager、Azure Key Vault）と連携し、それらのシステムに格納された秘密情報をKubernetesのSecretsとして同期することができる。
4. Bitnami Sealed Secrets
   - https://github.com/bitnami-labs/sealed-secrets
   - https://sealed-secrets.netlify.app/
   - Cluster内の **Controller(Operator)** とクライアント側Utilityである **kubeseal** で構成されていて、kubesealで暗号化してControllerの方で複合化する仕組みっぽい