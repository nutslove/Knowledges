# Envoy AI Gateway - AI プロバイダ認証方式

> [!NOTE]
> **対象バージョン**: Envoy AI Gateway v0.5.x
> **確認日**: 2026-04-21
> **情報ソース**: 公式 API Reference（v0.5 / latest 両方）、公式 getting-started、リリースノート v0.3/v0.4/v0.5

---

## 共通の仕組み

### BackendSecurityPolicy CRD

AI プロバイダへのアップストリーム認証は全て `BackendSecurityPolicy` CRD で設定する。`targetRefs` で `AIServiceBackend`（または InferencePool）を指定して紐付ける。

| フィールド | 役割 |
|---|---|
| `spec.type` | 認証タイプを enum で指定（下記参照） |
| `spec.targetRefs` | このポリシーを適用する `AIServiceBackend` 等の参照 |
| `spec.<type>` | タイプ別の詳細設定（`apiKey` / `awsCredentials` / `azureCredentials` / `gcpCredentials` 等） |

### 認証タイプ一覧（`spec.type` enum 値）

v0.5 時点で定義されている値：

| `type` | 対象プロバイダ |
|---|---|
| `APIKey` | OpenAI、その他 OpenAI 互換 API（Mistral 等） |
| `AWSCredentials` | AWS Bedrock |
| `AzureAPIKey` | Azure OpenAI（API Key 方式） |
| `AzureCredentials` | Azure OpenAI（Entra ID 方式） |
| `GCPCredentials` | GCP Vertex AI |
| `AnthropicAPIKey` | Anthropic 本家 API（AWS / GCP 経由でない直接連携） |

### Keyless 認証の設計思想

公式 `upstream-auth` capabilities ページからの抜粋:

- **AWS Bedrock**: OIDC integration with AWS STS（一時クレデンシャル発行）
- **Azure OpenAI**: Entra ID による短命 access token
- **GCP Vertex AI**: GCP Workload Identity Federation + Google STS

いずれも Gateway Controller が自動でトークンをリフレッシュする設計になっている。

---

## AWS Bedrock

### 認証フィールド

```yaml
spec:
  type: AWSCredentials
  awsCredentials:
    region: us-east-1          # 必須
    credentialsFile:            # オプション
      secretRef:
        name: aws-secret
      profile: default
    oidcExchangeToken:          # オプション
      awsRoleArn: arn:aws:iam::123:role/MyRole
      oidc:
        issuer: ...
```

### サポートされる認証方式

**API Reference の原文**:
> "When neither CredentialsFile nor OIDCExchangeToken is specified, the AWS SDK's default credential chain will be used. This automatically supports various authentication methods in the following order:
> 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)
> 2. EKS Pod Identity - automatically rotates credentials for pods in EKS clusters
> 3. IAM Roles for Service Accounts (IRSA) - injects credentials via mounted service account tokens
> 4. EC2 instance metadata (IAM instance roles)
> 5. ECS task roles"

つまり認証方式は **実質 3 パターン** に分類できる：

| # | 方式 | `credentialsFile` | `oidcExchangeToken` | 実体 |
|---|---|---|---|---|
| 1 | **Static Credentials** | 設定する | 空 | Secret 内の credentials ファイルを使用 |
| 2 | **OIDC → STS** | 空 | 設定する | 外部 OIDC provider → STS → 一時クレデンシャル |
| 3 | **Default Credential Chain** | 空 | 空 | AWS SDK の自動検出（下記 5 種類を順にトライ） |

Default Credential Chain 内の内訳:
- 環境変数（`AWS_ACCESS_KEY_ID` 等）
- **EKS Pod Identity**
- **IRSA (IAM Roles for Service Accounts)**
- EC2 IAM Instance Profile
- ECS Task Role

> [!IMPORTANT]
> **「EKS Pod Identity、IRSA、Static Credentials の 3 つ」という分類は少し粗い**。
> 正確には「Static / OIDC 交換 / Default Credential Chain の 3 パターン」であり、Pod Identity と IRSA は Default Chain 内の 2 選択肢という関係。

### 推奨

Kubernetes 環境では **Default Credential Chain**（Pod Identity または IRSA）が推奨。自動ローテーションが効き、Secret 管理不要。

---

## Azure OpenAI

### 認証タイプ 2 種類

Azure は他のクラウドと異なり、`spec.type` レベルで 2 つの独立したタイプに分かれる：

| `spec.type` | 用途 |
|---|---|
| `AzureAPIKey` | API Key 方式（シンプル、開発向け） |
| `AzureCredentials` | Entra ID 方式（Enterprise 向け） |

### AzureAPIKey（API Key 方式）

```yaml
spec:
  type: AzureAPIKey
  azureAPIKey:
    secretRef:
      name: azure-openai-key
```

Secret の key 名は `apiKey`。値は `api-key` ヘッダーに注入される。

### AzureCredentials（Entra ID 方式）

```yaml
spec:
  type: AzureCredentials
  azureCredentials:
    clientID: <Azure App Client ID>    # 必須
    tenantID: <Azure AD Tenant ID>     # 必須
    clientSecretRef:                    # オプション（排他 A）
      name: azure-client-secret
    oidcExchangeToken:                  # オプション（排他 B）
      oidc:
        issuer: ...
```

**API Reference の制約**:
> "Only one of ClientSecretRef or OIDCExchangeToken must be specified. Credentials will not be generated if neither are set."

つまり `clientSecretRef` と `oidcExchangeToken` のどちらか **一方のみ必須**。

| サブ方式 | 設定 | 用途 |
|---|---|---|
| **Client Secret** | `clientID` + `tenantID` + `clientSecretRef` | Service Principal の client secret（OAuth 2.0 client credentials flow） |
| **OIDC Federation** | `clientID` + `tenantID` + `oidcExchangeToken` | Workload Identity Federation（Kubernetes SA トークンを Entra ID と federate） |

### 公式 getting-started の記述

Azure OpenAI 接続ページ:
> "There are two ways to do the Azure OpenAI authentication: Microsoft Entra ID and API Key. We will use Microsoft Entra ID to authenticate an application to use the Azure OpenAI service."

→ 公式は **Entra ID（AzureCredentials）を推奨**、API Key は補助的。

### 現状の制約

> [!WARNING]
> **AKS 専用の Workload Identity タイプは v0.5 では未対応**。v0.5 リリースノートの Future Work に「**Azure/AKS workload identity**」と明記されており、今後対応予定。現状 AKS から使うには `oidcExchangeToken` に AKS の OIDC issuer を指定する形で Workload Identity Federation として組むのが最も近い。

### AWS との比較

| 項目 | AWS | Azure |
|---|---|---|
| 自動検出モード | ✅ Default Credential Chain あり | ❌ なし（明示設定必須） |
| Secret-less な方式 | IRSA / Pod Identity（自動） | OIDC Federation（明示設定）|

---

## GCP Vertex AI

### 認証フィールド

```yaml
spec:
  type: GCPCredentials
  gcpCredentials:
    projectName: my-project      # 必須
    region: us-central1          # 必須
    credentialsFile:             # オプション（排他 A）
      secretRef:
        name: gcp-sa-key
    workloadIdentityFederationConfig:  # オプション（排他 B）
      projectID: "123456"
      workloadIdentityPoolName: my-pool
      workloadIdentityProviderName: my-provider
      serviceAccountImpersonation:
        serviceAccountName: my-gcp-sa
      oidcExchangeToken:
        oidc:
          issuer: ...
```

### サポートされる認証方式

v0.3+ で GCP Vertex AI が production 対応。認証は **2 択**:

| 方式 | 設定フィールド | 実体 |
|---|---|---|
| **Service Account Key File (Static)** | `credentialsFile.secretRef` | JSON キーファイルを Secret に保存して access token 生成 |
| **Workload Identity Federation (Keyless)** | `workloadIdentityFederationConfig` | 外部 OIDC provider トークンを Google STS で交換、Service Account を impersonate |

### v0.3 リリースノートの原文

> "GCP Vertex AI Authentication with Service Account Key or Workload Identity Federation."

### 公式 getting-started の推奨

> "Consider using GCP Workload Identity (Federation)/IAM roles and limited-scope credentials for production environments."

### Secret のキー名

- `credentialsFile.secretRef`: Secret の key 名は **`service_account.json`**

### 現状の制約

> [!WARNING]
> **GCP では「GKE Workload Identity」の専用タイプは存在しない**。AWS の Default Credential Chain 相当の自動検出モード（ADC の自動使用など）もない。必ず `credentialsFile` か `workloadIdentityFederationConfig` のどちらかを明示的に設定する必要がある。
> GKE 上で動かす場合でも、`workloadIdentityFederationConfig` に GKE cluster の OIDC issuer を指定して Workload Identity Federation として組む形になる。

---

## 3 プロバイダー比較マトリクス

### 認証方式サマリ

| プロバイダー | Static Credentials | Keyless (Federation) | 自動検出モード |
|---|---|---|---|
| **AWS Bedrock** | `credentialsFile` (credentials file in Secret) | `oidcExchangeToken` (OIDC → STS → AssumeRoleWithWebIdentity) | ✅ **あり**: 両方未指定で Default Credential Chain 起動（Pod Identity / IRSA / EC2 / ECS / env 自動判定） |
| **Azure OpenAI** | `AzureAPIKey` (API Key) または `azureCredentials.clientSecretRef` (Service Principal) | `azureCredentials.oidcExchangeToken` (OIDC → Entra ID access token) | ❌ **なし**（明示設定必須） |
| **GCP Vertex AI** | `gcpCredentials.credentialsFile` (Service Account Key JSON) | `gcpCredentials.workloadIdentityFederationConfig` (OIDC → Google STS → SA impersonation) | ❌ **なし**（明示設定必須） |

### CRD 構造の違い

| プロバイダー | 認証タイプ数 | 構造 |
|---|---|---|
| AWS | 1 個（`AWSCredentials`）| `awsCredentials` 内で `credentialsFile` と `oidcExchangeToken` を選択、または両方空で自動検出 |
| Azure | 2 個（`AzureAPIKey`、`AzureCredentials`）| tier が 2 段階。API Key は独立、Entra ID 系は `AzureCredentials` 内で細分 |
| GCP | 1 個（`GCPCredentials`）| `gcpCredentials` 内で `credentialsFile` と `workloadIdentityFederationConfig` を選択（排他） |

---

## 実用アドバイス

### production 向けの選び方

| プロバイダー | EKS 上で動かす場合 | AKS 上で動かす場合 | GKE 上で動かす場合 | その他 K8s（on-prem 等）|
|---|---|---|---|---|
| AWS Bedrock | **Default Chain + Pod Identity / IRSA**（推奨） | OIDC → STS（AKS OIDC issuer 経由） | OIDC → STS（GKE OIDC issuer 経由）| Static または OIDC |
| Azure OpenAI | OIDC Federation（EKS OIDC issuer 経由） | **OIDC Federation**（AKS OIDC issuer 経由）※ v0.5 時点で専用タイプは未対応 | OIDC Federation（GKE OIDC issuer 経由）| Client Secret または API Key |
| GCP Vertex AI | **Workload Identity Federation**（EKS OIDC issuer 経由） | Workload Identity Federation（AKS OIDC issuer 経由）| **Workload Identity Federation**（GKE OIDC issuer 経由）| SA Key File または WIF |

### 汎用的なベストプラクティス

- **Secret は最小権限で**: `BackendSecurityPolicy` が参照する Secret は該当の namespace に限定
- **Credential rotation**: 静的キーは定期ローテーション、可能なら Keyless 方式へ移行
- **Environment separation**: dev / staging / prod で別プロジェクト・別ロールに分ける
- **監査ログ**: プロバイダ側（CloudTrail / Azure Monitor / GCP Cloud Logging）で Gateway からのアクセスを監査

---

## 参考リンク（公式一次ソース）

### API Reference（フィールド定義の一次ソース）
- v0.5 API Reference: https://aigateway.envoyproxy.io/docs/0.5/api/
- latest API Reference: https://aigateway.envoyproxy.io/docs/api/

### Capabilities
- Upstream Authentication: https://aigateway.envoyproxy.io/docs/capabilities/security/upstream-auth/
- Connecting to AI Providers: https://aigateway.envoyproxy.io/docs/capabilities/llm-integrations/connect-providers/

### Getting Started（実際の YAML 例）
- Connect Azure OpenAI: https://aigateway.envoyproxy.io/docs/getting-started/connect-providers/azure-openai/
- Connect GCP VertexAI: https://aigateway.envoyproxy.io/docs/getting-started/connect-providers/gcp-vertexai/

### リリースノート
- v0.3 (GCP Vertex AI production support): https://aigateway.envoyproxy.io/release-notes/v0.3/
- v0.4 (Anthropic direct / AWSAnthropic): https://aigateway.envoyproxy.io/release-notes/v0.4/
- v0.5 (Future: Azure/AKS workload identity): https://aigateway.envoyproxy.io/release-notes/v0.5/

### サンプル YAML（GitHub）
- Azure OpenAI: https://github.com/envoyproxy/ai-gateway/blob/main/examples/basic/azure_openai.yaml
- GCP Vertex AI: https://github.com/envoyproxy/ai-gateway/blob/main/examples/basic/gcp_vertex.yaml
- AWS Bedrock: https://github.com/envoyproxy/ai-gateway/tree/main/examples/basic

---

## 改訂履歴

| 日付 | バージョン | 内容 |
|---|---|---|
| 2026-04-21 | v0.5.x ベースで初版作成 | AWS / Azure / GCP の認証方式を公式 API Reference と照合して整理 |