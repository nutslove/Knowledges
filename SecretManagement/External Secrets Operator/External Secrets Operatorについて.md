- https://external-secrets.io/latest/  
    > External Secrets Operator is a Kubernetes operator that integrates external secret management systems like AWS Secrets Manager, HashiCorp Vault, Google Secrets Manager, Azure Key Vault, IBM Cloud Secrets Manager, CyberArk Conjur and many more. The operator reads information from external APIs and automatically injects the values into a Kubernetes Secret.
- https://external-secrets.io/latest/introduction/overview/  
    > The External Secrets Operator extends Kubernetes with Custom Resources, which define where secrets live and how to synchronize them. The controller fetches secrets from an external API and creates Kubernetes secrets. If the secret from the external API changes, the controller will reconcile the state in the cluster and update the secrets accordingly.

## install
- https://external-secrets.io/latest/introduction/getting-started/
### Helmでインストール
```shell
helm repo add external-secrets https://charts.external-secrets.io

helm install external-secrets \
   external-secrets/external-secrets \
    -n external-secrets \
    --create-namespace \
```

### k8sマニフェストファイルからインストール
```shell
kubectl apply -k "https://raw.githubusercontent.com/external-secrets/external-secrets/<replace_with_your_version>/deploy/crds/bundle.yaml"
```

## 設定
- HashiCorp Vaultとの組み合わせ設定
    - https://external-secrets.io/v0.5.7/provider-hashicorp-vault/
#### `SecretStore`の作成
- `SecretStore`にSecret Store(e.g. AWS Secrets Manager、Vaultなど)
- Secret StoreとしてHashiCorp Vaultを使う例
    ```yaml
    ---
    apiVersion: external-secrets.io/v1alpha1
    kind: SecretStore
    metadata:
      name: vault-backend
    spec:
      provider:
        vault:
          server: "http://<vaultのエンドポイント>:8200"
          path: "secret"
          version: "v2"
          auth:
            tokenSecretRef:
              name: vault-token
              key: token
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: vault-token
      data:
        token: <BASE64_ENCODED_VAULT_ROOT_TOKEN> ## vault init時に出力されるInitial Root Tokenをbase64でencodingした値
    ```
- `ExternalSecret`でVaultに登録されているSecretを参照(取得)
    ```yaml
    apiVersion: external-secrets.io/v1alpha1
    kind: ExternalSecret
    metadata:
      name: my-external-secret
    spec:
      refreshInterval: "1h" ## External Secrets OperatorがSecretStoreからsecretを再取得する間隔
      secretStoreRef:
        name: vault-backend ## SecretStoreのmetadata.nameの値
        kind: SecretStore
      target:
        name: my-secret ## Secretリソースmy-secretとして管理される
        creationPolicy: Owner
      data:
        - secretKey: my-key1 ## KubernetesのSecretリソース内のkey名（Pod側で指定）
          remoteRef:
            key: path/to/secret ## Vaultのパス
            property: my-access-key ## my-access-keyの値が取得される（Vault内のsecretのkey名）
        - secretKey: my-key2 ## KubernetesのSecretリソース内のkey名（Pod側で指定）
          remoteRef:
            key: path/to/secret ## Vaultのパス
            property: my-secret-key ## my-secret-keyの値が取得される（Vault内のsecretのkey名）
    ```
    - **Secrets Engineのversionが2の場合、pathのEngineの次に`/data/`を追加する必要がある。例えばSecrets Engineが`secret/`で、パスが`secret/minio/config`の場合、`secret/data/minio/config`にする必要がある**
    - `creationPolicy`には、`Secret`リソースをどのように管理するかを指定
        - `Owner`: ExternalSecretが作成したSecretリソースの所有者として振る舞い、ExternalSecretが削除されると、関連するSecretリソースも削除される。
        - `Merge`: ExternalSecretがSecretリソースを作成または更新し、既存のキーと値を維持する。
        - `None`: Secretリソースの作成や更新を行わず、既存のリソースを変更しない。
- Pod(Deployment)でSecretを環境変数として参照
    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: my-pod
    spec:
      containers:
      - name: my-container
        image: my-image
        env:
        - name: envname1 ## 環境変数名
          valueFrom:
            secretKeyRef:
              name: my-secret ## ExternalSecretのspec.target.nameの値
              key: my-key1 ## ExternalSecretのspec.data.secretKeyの値
        - name: envname2 ## 環境変数名
          valueFrom:
            secretKeyRef:
              name: my-secret ## ExternalSecretのspec.target.nameの値
              key: my-key2 ## ExternalSecretのspec.data.secretKeyの値
    ```