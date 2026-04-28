# Envoy AI Gateway - Azure OpenAI 連携手順

> Envoy AI Gateway を Azure OpenAI に Microsoft Entra ID（OAuth 2.0 Client Credentials Grant）で接続する設定手順をまとめる。Client Secret は AWS Secrets Manager に保管し、External Secrets Operator (ESO) 経由で Kubernetes Secret として展開する。

## 構成概要

```
[ クライアント ]
       │
       ▼
[ Envoy AI Gateway (EKS) ]
       │ ① Client Secret を K8s Secret から読み取り
       │ ② Entra ID にトークンリクエスト（Client Credentials Grant）
       ▼
[ Microsoft Entra ID ]
       │ ③ Service Principal を認証
       │ ④ アクセストークン発行（1 時間有効）
       ▼
[ Envoy AI Gateway ]
       │ ⑤ Bearer トークンを付与
       ▼
[ Azure OpenAI ]


別ライン:
[ AWS Secrets Manager ] ──── ESO（1 時間ごとに同期）────▶ [ K8s Secret ]
```

## 認証方式の選定

Azure OpenAI への認証方式は複数あるが、本手順では **Microsoft Entra ID + Client Secret 方式** を採用する。理由：

| 方式 | 評価 | 採用判断 |
|---|---|---|
| **Entra ID + Client Secret** | 公式 Getting Started で動作保証されている、RBAC でスコープ制御可能、監査ログで主体識別可能 | ✅ 採用 |
| Entra ID + Workload Identity Federation | 静的 Secret 不要で最も安全、ただし v0.5 時点で Envoy AI Gateway 側の公式サンプルが乏しい | 将来的な移行先候補 |
| API Key | 静的長命キー、RBAC なし、リソース全体への無制限アクセス | ❌ 非推奨 |

Client Secret 自体は AWS Secrets Manager で管理し、ESO で K8s に同期することで、マニフェストへの直書きを回避する。

## 全体の流れ

1. **Azure 側**: App Registration の作成 → Client Secret 発行 → Service Principal への RBAC 権限付与
2. **AWS 側**: Secrets Manager に Client Secret を保存
3. **Kubernetes 側**: ExternalSecret + BackendSecurityPolicy 等のマニフェストを apply

---

## 1. Azure 側の設定

### 1.1 前提条件

- Azure サブスクリプションへのアクセス権
- 対象 Azure OpenAI リソース（既存または新規作成）
- Azure OpenAI リソースに **カスタムサブドメイン名が設定済み**（Entra ID 認証には必須）
- 利用したいモデル（例: GPT-4o, GPT-5.4 系）が Azure AI Foundry でデプロイ済み

### 1.2 必要な権限

App Registration の作成と Service Principal への RBAC 割り当てには、以下のいずれかの権限が必要：

| 操作 | 必要なロール |
|---|---|
| App Registration の作成 | Microsoft Entra ID の管理者ロール（Application Developer 等） |
| Azure OpenAI への RBAC 割り当て | **Owner** / **User Access Administrator** / **Role Based Access Control Administrator** のいずれか |

`Contributor` ロールでは RBAC 割り当てができない点に注意。`Microsoft.Authorization/roleAssignments/write` 権限を持つロールが必要。

### 1.3 Step 1: App Registration の作成

#### Azure Portal での手順

1. Azure Portal → **Microsoft Entra ID** → **App registrations** → **New registration**
2. 以下を入力：
   - **Name**: 任意（例: `envoy-ai-gateway`）
   - **Supported account types**: "Accounts in this organizational directory only (Single tenant)"
   - **Redirect URI**: 設定不要
3. 登録後、Overview ページで以下 2 つの値を取得：
   - **Application (client) ID**
   - **Directory (tenant) ID**

#### Azure CLI での手順

```bash
APP_NAME="envoy-ai-gateway"

# App Registration 作成
az ad app create --display-name "${APP_NAME}"

# Client ID 取得
CLIENT_ID=$(az ad app list --display-name "${APP_NAME}" --query "[0].appId" -o tsv)

# Service Principal 作成（RBAC 割り当てに必須）
az ad sp create --id ${CLIENT_ID}

# Tenant ID 取得
TENANT_ID=$(az account show --query tenantId -o tsv)

echo "CLIENT_ID=${CLIENT_ID}"
echo "TENANT_ID=${TENANT_ID}"
```

### 1.4 Step 2: Client Secret の発行

#### Azure Portal での手順

1. 作成した App Registration を開く
2. 左メニュー **Certificates & secrets** → **Client secrets** タブ → **New client secret**
3. 以下を入力：
   - **Description**: 任意
   - **Expires**: 有効期限を選択（社内ポリシーに従う）
4. **Add** ボタンを押す
5. 表示された **Value** 列の値をコピー

> ⚠️ **重要**: Value は一度画面を離れると二度と表示されない。必ずこのタイミングで取得すること。Value の隣にある "Secret ID" は別物。

#### Azure CLI での手順

```bash
CLIENT_SECRET=$(az ad app credential reset \
  --id ${CLIENT_ID} \
  --display-name "envoy-ai-gateway-secret" \
  --years 1 \
  --query password -o tsv)

echo "CLIENT_SECRET=${CLIENT_SECRET}"
```

### 1.5 Step 3: Azure OpenAI への RBAC 権限付与

App Registration を作っただけでは Azure OpenAI を呼び出せない。Service Principal に対象 Azure OpenAI リソースへの **Cognitive Services OpenAI User** ロールを付与する。

#### Azure Portal での手順

1. 対象の **Azure OpenAI リソース** を開く
2. 左メニュー **Access control (IAM)** → **Add** → **Add role assignment**
3. Role タブで **Cognitive Services OpenAI User** を選択 → Next
4. Members タブで：
   - **Assign access to**: "User, group, or service principal"
   - **+ Select members** → Step 1 で作った App Registration の名前を検索して選択
5. **Review + assign** で割り当て

> 「Add role assignment」ボタンが非活性の場合は、自分のユーザーに RBAC 割り当て権限がない。1.2 を参照。

#### Azure CLI での手順

```bash
RESOURCE_GROUP="<your-resource-group>"
AOAI_RESOURCE_NAME="<your-azure-openai-resource-name>"

# Service Principal の Object ID 取得
SP_OBJECT_ID=$(az ad sp show --id ${CLIENT_ID} --query id -o tsv)

# Azure OpenAI リソースの Resource ID 取得
AOAI_RESOURCE_ID=$(az cognitiveservices account show \
  --name ${AOAI_RESOURCE_NAME} \
  --resource-group ${RESOURCE_GROUP} \
  --query id -o tsv)

# ロール割り当て
az role assignment create \
  --assignee-object-id ${SP_OBJECT_ID} \
  --assignee-principal-type ServicePrincipal \
  --role "Cognitive Services OpenAI User" \
  --scope ${AOAI_RESOURCE_ID}
```

> ⚠️ ロール割り当ては伝播に **最大 5 分** かかる。直後にテストして 401/403 が出ても、少し待つと解消する場合がある。

### 1.6 Step 4: Azure OpenAI リソースの情報を確認

マニフェストに使う情報を取得する：

| 項目 | 取得元 | 用途 |
|---|---|---|
| エンドポイントホスト名 | Azure OpenAI リソース → **Keys and Endpoint** ページ | Backend FQDN（例: `<your-resource>.openai.azure.com`）|
| デプロイメント名 | Azure AI Foundry → **Deployments** | AIGatewayRoute の `x-ai-eg-model` 値と一致させる |
| API version | Azure AI Foundry → 各モデル → **Chat Playground** → **View code** | AIServiceBackend の `schema.version` |

#### API version の確認方法

最も確実な方法：

1. Azure AI Foundry（https://ai.azure.com）にアクセス
2. 対象モデルのデプロイメントを選択
3. **Chat Playground** または **Models + endpoints** で **View code** をクリック
4. 表示されたサンプルコード内の `api_version="..."` の値を確認

参考：公式の API version reference ページ
https://learn.microsoft.com/en-us/azure/ai-services/openai/reference

### 1.7 Azure 側設定のまとめ

この時点で以下が揃っている状態：

- App Registration の **Application (client) ID**
- Entra ID の **Directory (tenant) ID**
- App Registration の **Client Secret Value**
- Azure OpenAI リソースの **エンドポイントホスト名**
- Azure AI Foundry の **デプロイメント名**
- 利用する **API version**
- Service Principal に **Cognitive Services OpenAI User** ロール付与済み

---

## 2. AWS 側の設定

### 2.1 AWS Secrets Manager への Client Secret 保存

```bash
aws secretsmanager create-secret \
  --name <your-secret-name> \
  --description "Azure Entra ID Client Secret for Envoy AI Gateway" \
  --secret-string '{"azure-client-secret":"<実際の Client Secret 値>"}' \
  --region ap-northeast-1
```

JSON 形式で保存しておくと、ESO で `property` フィールドを使って取り出しやすくなる。複数の秘密情報を 1 つの Secret にまとめることも可能。

### 2.2 ESO 用の IAM 権限

External Secrets Operator が AWS Secrets Manager にアクセスできるように、ESO の ServiceAccount に紐付いた IAM Role に以下の権限を付与する。

#### IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": [
        "arn:aws:secretsmanager:ap-northeast-1:<account-id>:secret:<your-secret-name>*"
      ]
    }
  ]
}
```

#### Pod Identity Association（推奨）または IRSA

ESO 用 IAM Role と ESO の ServiceAccount を紐付ける。Pod Identity を使う場合：

```bash
aws eks create-pod-identity-association \
  --cluster-name <cluster-name> \
  --namespace external-secrets \
  --service-account external-secrets \
  --role-arn arn:aws:iam::<account-id>:role/<eso-role-name>
```

ServiceAccount 名と namespace は、ESO を Helm でインストールした際の値に合わせる（デフォルトは `external-secrets` namespace の `external-secrets` SA）。

---

## 3. Kubernetes マニフェスト

### 3.1 前提

- ESO（External Secrets Operator）が EKS クラスタにインストール済み
- `SecretStore` または `ClusterSecretStore` が `aws-secretsmanager-store` という名前で定義済み
- Envoy AI Gateway がインストール済み（CRD およびコントローラ）
- Gateway / EnvoyProxy / GatewayClass などの基本リソースがすでに存在

### 3.2 SecretStore の例（未作成の場合）

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secretsmanager-store
  namespace: envoy-ai-gateway-system
spec:
  provider:
    aws:
      service: SecretsManager
      region: ap-northeast-1
      # auth フィールドを指定しない場合、ESO Pod の SA に紐付いた IAM 認証が使われる
```

クラスタ全体で共有する場合は `ClusterSecretStore` を使う。

### 3.3 Azure OpenAI 連携用マニフェスト

#### ExternalSecret: Client Secret を AWS Secrets Manager から同期

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: envoy-ai-gateway-azure-client-secret
  namespace: envoy-ai-gateway-system
spec:
  refreshInterval: 1h                              # 1 時間ごとに同期
  secretStoreRef:
    name: aws-secretsmanager-store
    kind: SecretStore                               # ClusterSecretStore を使う場合は ClusterSecretStore
  target:
    name: envoy-ai-gateway-azure-client-secret     # 生成される K8s Secret の名前
    creationPolicy: Owner
  data:
    - secretKey: client-secret                     # ★ K8s Secret 内のキー名(固定: client-secret)
      remoteRef:
        key: <your-secret-name>                    # AWS Secrets Manager の Secret 名
        property: azure-client-secret              # JSON 内のフィールド名
```

> 重要: `secretKey: client-secret` は **Envoy AI Gateway の仕様で固定**。これ以外の名前にすると Controller が認識しない。

#### AIServiceBackend: Azure OpenAI バックエンドの定義

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIServiceBackend
metadata:
  name: envoy-ai-gateway-azure
  namespace: envoy-ai-gateway-system
spec:
  schema:
    name: AzureOpenAI
    version: <api-version>                        # 例: 2024-12-01-preview
  backendRef:
    name: envoy-ai-gateway-azure
    kind: Backend
    group: gateway.envoyproxy.io
```

#### BackendSecurityPolicy: Entra ID 認証設定

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: BackendSecurityPolicy
metadata:
  name: envoy-ai-gateway-azure
  namespace: envoy-ai-gateway-system
spec:
  targetRefs:
    - group: aigateway.envoyproxy.io
      kind: AIServiceBackend
      name: envoy-ai-gateway-azure
  type: AzureCredentials
  azureCredentials:
    clientID: "<azure-client-id>"                 # App Registration の Application (client) ID
    tenantID: "<azure-tenant-id>"                 # Entra ID の Directory (tenant) ID
    clientSecretRef:
      name: envoy-ai-gateway-azure-client-secret  # ExternalSecret が生成する K8s Secret
      namespace: envoy-ai-gateway-system
```

#### Backend: Azure OpenAI のエンドポイント

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: envoy-ai-gateway-azure
  namespace: envoy-ai-gateway-system
spec:
  endpoints:
    - fqdn:
        hostname: <your-resource>.openai.azure.com
        port: 443
```

#### BackendTLSPolicy: TLS 検証

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: envoy-ai-gateway-azure-tls
  namespace: envoy-ai-gateway-system
spec:
  targetRefs:
    - group: gateway.envoyproxy.io
      kind: Backend
      name: envoy-ai-gateway-azure
  validation:
    wellKnownCACertificates: System
    hostname: <your-resource>.openai.azure.com    # Backend と同じ
```

#### AIGatewayRoute: ルーティングルール

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: envoy-ai-gateway
  namespace: envoy-ai-gateway-system
spec:
  parentRefs:
    - name: envoy-ai-gateway
      namespace: envoy-gateway-system
      kind: Gateway
      group: gateway.networking.k8s.io
  rules:
    - matches:
        - headers:
            - type: Exact
              name: x-ai-eg-model
              value: <deployment-name>            # ★ Azure AI Foundry の deployment 名
      backendRefs:
        - name: envoy-ai-gateway-azure
```

> 重要: `value` は **Azure AI Foundry の Deployments ページに表示される deployment 名と完全一致** させる必要がある。Azure OpenAI は body の `model` フィールドではなく URL パスの deployment 名で判定するため。

複数モデルを使う場合は rules を複数並べる。

### 3.4 マニフェスト内の置換ポイント

| プレースホルダ | 取得元 |
|---|---|
| `<your-secret-name>` | AWS Secrets Manager の Secret 名 |
| `azure-client-secret` | AWS Secrets Manager の JSON 内のフィールド名（任意の名前で OK）|
| `<api-version>` | Azure AI Foundry の View code で確認した値 |
| `<azure-client-id>` | Entra ID App Registration の Application (client) ID |
| `<azure-tenant-id>` | Entra ID の Directory (tenant) ID |
| `<your-resource>.openai.azure.com` | Azure OpenAI リソースの Keys and Endpoint ページのホスト名 |
| `<deployment-name>` | Azure AI Foundry の Deployments ページの deployment 名 |

---

## 4. apply と動作確認

### 4.1 apply

```bash
kubectl apply -f azure-openai-config.yaml
```

### 4.2 ExternalSecret の同期確認

```bash
kubectl get externalsecret -n envoy-ai-gateway-system

# READY=True、SYNCED=True になっていれば成功
```

```bash
kubectl describe externalsecret -n envoy-ai-gateway-system envoy-ai-gateway-azure-client-secret
```

### 4.3 K8s Secret の生成確認

```bash
kubectl get secret -n envoy-ai-gateway-system envoy-ai-gateway-azure-client-secret -o yaml

# data.client-secret に base64 エンコードされた値が入っているはず
```

### 4.4 BackendSecurityPolicy の状態確認

```bash
kubectl get backendsecuritypolicy -n envoy-ai-gateway-system

# STATUS=Accepted になっていれば成功
```

### 4.5 動作確認

```bash
curl -H "Content-Type: application/json" \
  -d '{
    "model": "<deployment-name>",
    "messages": [{"role": "user", "content": "Hi."}]
  }' \
  $GATEWAY_URL/v1/chat/completions
```

---

## 5. トラブルシューティング

### 5.1 ExternalSecret が SYNCED にならない

```bash
kubectl describe externalsecret -n envoy-ai-gateway-system envoy-ai-gateway-azure-client-secret
kubectl logs -n external-secrets deployment/external-secrets
```

よくある原因：

- **IAM 権限不足**: ESO の SA に紐付いた IAM Role に `secretsmanager:GetSecretValue` 権限がない
- **Secret 名のミスマッチ**: `remoteRef.key` の値が AWS Secrets Manager 上の実際の Secret 名と一致していない
- **JSON のフィールド名のミスマッチ**: `remoteRef.property` の値が Secret 内 JSON のキー名と一致していない
- **SecretStore / ClusterSecretStore 種別の指定ミス**: `ClusterSecretStore` を `kind: SecretStore` で指定している

### 5.2 401 Unauthorized / 403 Forbidden

- **RBAC 伝播待ち**: ロール割り当て直後は最大 5 分かかる
- **ロール不足**: Service Principal に **Cognitive Services OpenAI User** が付与されていない
- **Client Secret の有効期限切れ**: Entra ID Portal で確認・再発行
- **Client ID / Tenant ID のタイプミス**: ハイフンの位置まで含めて確認
- **カスタムサブドメイン未設定**: Azure OpenAI リソースにカスタムサブドメインが設定されていない

### 5.3 404 Not Found

- **デプロイメント名と `model` 値の不一致**: Azure AI Foundry の Deployment 名と `x-ai-eg-model` の値が完全一致しているか確認
- **モデルが未デプロイ**: Foundry でモデルがデプロイ済みか確認
- **エンドポイント URL の typo**: `Backend.endpoints[].fqdn.hostname` を再確認

### 5.4 Gateway Controller のログ確認

```bash
kubectl logs -n envoy-ai-gateway-system deployment/ai-gateway-controller
```

よくあるエラーメッセージ：

| エラー | 原因 |
|---|---|
| `AADSTS7000215: Invalid client secret provided` | Client Secret が間違っている、または Secret ID と Value を取り違えている |
| `AADSTS700016: Application with identifier '<id>' was not found` | Client ID 間違い、またはテナント間違い |
| `AADSTS50034: The user account does not exist in tenant` | Tenant ID 間違い |

---

## 6. ローテーション運用

### 6.1 自動ローテーションの仕組み

```
1. AWS Secrets Manager で値を更新
2. ESO が refreshInterval（1 時間）後に K8s Secret を更新
3. Envoy AI Gateway Controller が K8s Secret 更新を検知
4. 新しい Client Secret で Entra ID から新しいアクセストークンを取得
```

### 6.2 即時反映したい場合

```bash
kubectl annotate externalsecret -n envoy-ai-gateway-system \
  envoy-ai-gateway-azure-client-secret \
  force-sync=$(date +%s) --overwrite
```

### 6.3 Client Secret の手動ローテーション手順

1. Entra ID Portal で App Registration → Certificates & secrets → **新しい Client Secret を発行**
2. 新しい Value を AWS Secrets Manager に保存：
   ```bash
   aws secretsmanager put-secret-value \
     --secret-id <your-secret-name> \
     --secret-string '{"azure-client-secret":"<new-value>"}'
   ```
3. ESO の自動同期を待つ、または `force-sync` annotation で即時反映
4. 動作確認後、Entra ID 側の **古い Client Secret を削除**

複数の Client Secret を並行発行できるため、新旧両方が有効な状態で切り替えれば無停止でローテーション可能。

---

## 7. 設定チェックリスト

### Azure 側

- [ ] App Registration を作成済み
- [ ] Service Principal が作成済み（`az ad sp create` 実行済み）
- [ ] Client Secret を発行済み（Value をメモ済み）
- [ ] 対象 Azure OpenAI リソースに **Cognitive Services OpenAI User** ロール付与済み
- [ ] Azure OpenAI リソースにカスタムサブドメインが設定済み
- [ ] 利用するモデルを Azure AI Foundry でデプロイ済み
- [ ] 使用する API version を Foundry の View code で確認済み

### AWS 側

- [ ] AWS Secrets Manager に Client Secret を保存済み
- [ ] ESO 用 IAM Role に `secretsmanager:GetSecretValue` 権限を付与済み
- [ ] ESO の ServiceAccount と IAM Role を Pod Identity / IRSA で紐付け済み

### Kubernetes 側

- [ ] ESO がインストール済み
- [ ] SecretStore または ClusterSecretStore が定義済み
- [ ] Envoy AI Gateway がインストール済み
- [ ] Gateway / EnvoyProxy / GatewayClass が存在
- [ ] ExternalSecret apply 後、K8s Secret が `client-secret` キーで生成されている
- [ ] BackendSecurityPolicy の status が `Accepted`
- [ ] AIGatewayRoute の `x-ai-eg-model` 値が deployment 名と一致

---

## 参考リンク

- [Envoy AI Gateway - Connect Azure OpenAI](https://aigateway.envoyproxy.io/docs/getting-started/connect-providers/azure-openai/)
- [Envoy AI Gateway - API Reference](https://aigateway.envoyproxy.io/docs/api/)
- [Microsoft - OAuth 2.0 Client Credentials Grant Flow](https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-client-creds-grant-flow)
- [Azure OpenAI - Authentication](https://learn.microsoft.com/en-us/azure/ai-services/openai/reference#authentication)
- [Azure OpenAI - API version reference](https://learn.microsoft.com/en-us/azure/ai-services/openai/reference)
- [External Secrets Operator - AWS Secrets Manager](https://external-secrets.io/latest/provider/aws-secrets-manager/)