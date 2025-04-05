## External Secrets Operator（ESO）とは
- https://external-secrets.io/latest/  
  ![](./image/introduction.jpg)
- **`ExternalSecret`というESOのCRを作成すると、ESOが`ExternalSecret`で指定している外部のSecret Store (e.g. AWS SecretManager、Hashicorp Vaultなど) に保存されているシークレットをk8sの`Secret`リソースに変換してくれる**

## install
- https://external-secrets.io/latest/introduction/getting-started/

### Helmで`values.yml`を用いてデプロイする場合
- `values.yml`の全量
  - https://github.com/external-secrets/external-secrets/blob/main/deploy/charts/external-secrets/values.yaml
- `values.yml`の例  
  ```yaml
  global:
    nodeSelector:
      karpenter.sh/nodepool: arm64-nodepool
      karpenter.sh/capacity-type: on-demand
  serviceAccount:
    create: true
    name: "eso-service-account"
  ```
- インストール  
  ```shell
  helm install external-secrets --values=values.yml external-secrets/external-secrets -n external-secrets --create-namespace
  ```

---

## Components
- https://external-secrets.io/latest/api/components/  
  ![](./image/components.jpg)

## CR（Custom Resource）
- 以下の4つの主要なCRが作成できる
  - **`ExternalSecret`**
    - 外部のシークレットストアから特定のシークレットを取得し、Kubernetesの`Secret`リソースとして同期するための設定を定義
    - `ExternalSecret`リソースを作成すると、`Secret`リソースが作成される
  - **`ClusterExternalSecret`**
    - 複数のnamespaceで使える`ExternalSecret`
    - `namespaceSelectors`で作成するnamespaceを指定(リスト)
    - https://external-secrets.io/latest/api/clusterexternalsecret/  
      > The `ClusterExternalSecret` is a cluster scoped resource that can be used to manage `ExternalSecret` resources in specific namespaces.
      >
      > With `namespaceSelectors` you can select namespaces in which the ExternalSecret should be created. If there is a conflict with an existing resource the controller will error out.
  - **`SecretStore`**
    - 外部シークレットストアの接続情報を定義
    - 特定の`namespace`でのみ使用可能
  - **`ClusterSecretStore`**
    - 外部シークレットストアの接続情報を定義
    - クラスタ全体で使用可能

---

## AWS Secrets Managerとの連携例
- 参照URL
  - https://aws.amazon.com/jp/blogs/news/leverage-aws-secrets-stores-from-eks-fargate-with-external-secrets-operator/
    - IRSAでの設定例
  - https://external-secrets.io/latest/provider/aws-secrets-manager/
  - https://techstep.hatenablog.com/entry/2023/02/02/182127

### `SecretStore`
- EKS WorkerNodeのIAMロールでSecrets Managerからシークレットを取得する例（IRSAやシークレットキーで取得する場合は追加の設定が必要）  

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: monitoring
spec:
  provider:
    aws:
      service: SecretsManager
      region: ap-northeast-1
```

### `ExternalSecret`
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: rds-auth
spec:
  refreshInterval: 1h # SecretStoreプロバイダーから再度値を読み込む間隔を指定
  secretStoreRef: # 参照する SecretStore を指定
    name: aws-secrets-manager
    kind: SecretStore
  target: # ExternalSecret から作成されるKubernetes Secretを指定する
    name: rdsauth # 作成されるSecretリソース名
    creationPolicy: Owner
  data: # 取得する秘匿情報を指定する
  - secretKey: senaki-rds-host # 作成するSecretのkey名. PodやDeployment側で指定する
    remoteRef:
      key: senaki-rds-auth # AWS Secrets Managerに登録したシークレット名
      property: host # AWS Secrets Managerに登録したシークレットキー名
  - secretKey: senaki-rds-engine
    remoteRef:
      key: senaki-rds-auth
      property: engine
  - secretKey: senaki-rds-username
    remoteRef:
      key: senaki-rds-auth
      property: username
  - secretKey: senaki-rds-password
    remoteRef:
      key: senaki-rds-auth
      property: password
```

### Secretを使う側（`Deployment`）
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  labels:
    app: grafana
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
        - name: grafana
          image: grafana/grafana:11.4.0
          imagePullPolicy: IfNotPresent
          env:
            - name: GF_DATABASE_TYPE
              valueFrom:
                secretKeyRef:
                  name: rdsauth
                  key: senaki-rds-engine
            - name: GF_DATABASE_HOST
              valueFrom:
                secretKeyRef:
                  name: rdsauth
                  key: senaki-rds-host
            - name: GF_DATABASE_NAME
              value: "grafana_sk"
            - name: GF_DATABASE_USER
              valueFrom:
                secretKeyRef:
                  name: rdsauth
                  key: senaki-rds-username
            - name: GF_DATABASE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: rdsauth
                  key: senaki-rds-password
            - name: AWS_DEFAULT_REGION
              value: "ap-northeast-1"
          ports:
            - containerPort: 3000
              name: http-grafana
              protocol: TCP
```
---

## Template機能
- https://external-secrets.io/v0.15.1/guides/templating/
- 以下の場合はTemplate機能を使えば解決できる
  - `ExternalSecret`（変換される`Secret`）の一部のデータは外部のSecretStore（e.g. AWS SecretManager）から取得するのではなく直接記述したい場合
  - 変換される`Secret`のAnnotationやLabelに任意の値を設定したい場合
- 例
  - `url`や`insecure`などの項目は別に秘密性はなく、べた書きしても良いのでべた書きして、秘密性が高い`password`などは外部SecretStoreから取得して設定する。
  - 変換される`Secret`に設定したいLabelやAnnotationは`spec.target.template.templateFrom`フィールドに、Labelの場合は`- target: Labels`で、Annotationの場合は`- target: Annotations`で設定する
  ```yaml
  apiVersion: external-secrets.io/v1beta1
  kind: ExternalSecret
  metadata:
    name: lee-repo
    namespace: argocd
  spec:
    refreshInterval: 1h
    secretStoreRef:
      name: aws-secrets-manager
      kind: ClusterSecretStore
    target:
      name: lee-repo
      creationPolicy: Owner
      template:
        engineVersion: v2
        templateFrom:
      # - target: Annotations
        - target: Labels
          literal: "argocd.argoproj.io/secret-type: repository"
        data:
          url: https://github.com/nutslove/IaC.git
          insecure: "false"
          username: "{{ .username }}"
          password: "{{ .password }}"
    data:
    - secretKey: username
      remoteRef:
        key: argocd-repo-creds
        property: username
    - secretKey: password
      remoteRef:
        key: argocd-repo-creds
        property: password
  ```