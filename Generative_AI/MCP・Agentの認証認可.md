# 概要
MCP (Model Context Protocol) や A2A (Agent-to-Agent) プロトコルにおける、ユーザーや他のAgentからのアクセスに対する認証・認可の設計についてまとめる。

王道パターンは **「IdPで認証 → JWTのclaimsを使ってRBAC/ABACで認可」** で、必要に応じてPolicy Engine (OPA/OpenFGA等) で細粒度認可を追加する。

---

# 基本アーキテクチャ

## ベース構成: OAuth 2.1 + OIDC

MCPの公式仕様(2025年6月のAuthorization spec改定以降)では、以下が推奨されている。

- MCPサーバーは **Resource Server** として振る舞う
- 認証はIdP (Entra ID, Auth0, Cognito, Keycloak等) に委譲
- **Resource Indicators (RFC 8707)** でトークンのaudienceを縛る

## JWT claimsでRBAC/ABAC

| トークン種別 | 取得フロー | 制御に使う情報 |
|------------|----------|-------------|
| ユーザートークン | Authorization Code + PKCE | ユーザーのロール/グループclaim |
| M2Mトークン | Client Credentials | scope, client_id |

---

# 重要な概念

## RBAC vs ABAC

**RBAC (Role-Based Access Control)**
- 「ロール」で制御
- 例: `role=admin なら全API許可`

**ABAC (Attribute-Based Access Control)**
- 「属性」で制御
- 例: `department=platform かつ env=staging かつ time<18:00 なら許可`
- JWTのclaimsに `role` 以外にも `department`, `team`, `region`, `clearance_level` など任意の属性を入れて、組み合わせて細かいポリシーを書ける
- 実務ではRBACをベースに条件式で拡張するハイブリッドが主流

## Resource Indicators (RFC 8707)

OAuthトークン取得時に、**「このトークンはどのAPI (リソース) 向けか」を明示する仕組み**。

```
POST /oauth/token
grant_type=authorization_code
resource=https://mcp-loki.example.com
resource=https://mcp-grafana.example.com
```

IdPは発行するJWTの `aud` (audience) claimにresource値を入れる。受け取ったMCPサーバーは `aud` を検証して、自分宛じゃないトークンを拒否できる。

**なぜ重要か**: これがないと、悪意あるMCPサーバーAが受け取ったユーザートークンを別のMCPサーバーBに転送して悪用できる (**token passthrough攻撃**)。MCP仕様で明示的に禁止されている。

## Confused Deputy問題と Token Exchange

MCPサーバーが下流APIを呼ぶときに、ユーザーのトークンをそのまま転送するのはNG。
**On-Behalf-Of (OBO) フロー** や **Token Exchange (RFC 8693)** で別トークンに変換するのが推奨。

- `act` claim (actor) に Client Agent の身元
- `sub` claim にユーザーの身元
- Server Agentは「誰のために誰が呼んでいるか」両方検証できる

## Tool-level RBAC

MCPはツール単位で権限を分けたいケースが多いので、scopeをツール名に対応させる設計が現実的。
例: `mcp:tool:read_loki`, `mcp:tool:write_grafana`

---

# A2A (Agent-to-Agent) プロトコル

Googleが2025年4月に発表、2025年6月23日にLinux FoundationがA2Aプロジェクトを正式発足させ、その配下で標準化が進んでいる。
MCPが「Agent ↔ Tool」なのに対し、A2Aは「Agent ↔ Agent」の通信。

## コアコンセプト: Agent Card

各AgentはJSONの「名刺」を `/.well-known/agent-card.json` (RFC 8615準拠) で公開する。
※ `/.well-known/agent.json` はv0.2.x以前の古いパス。現行SDK/仕様では `agent-card.json` に変更されており、古いパスのままだとクライアントのディスカバリが壊れる。

認証は、A2A初期(v0.1系)の `authentication` フィールドではなく、**現行仕様(v0.3.0)ではOpenAPI 3.0準拠の `securitySchemes` + `security`** で表現する。
`security` は「OR of ANDs」(例: OAuth または (APIKey かつ mTLS)) を表す。

```json
{
  "protocolVersion": "0.3.0",
  "name": "RCA Agent",
  "url": "https://rca-agent.kinto.jp/a2a",
  "capabilities": { "streaming": true, "pushNotifications": true },
  "securitySchemes": {
    "oauth2": {
      "type": "oauth2",
      "flows": {
        "authorizationCode": {
          "authorizationUrl": "https://login.microsoftonline.com/.../authorize",
          "tokenUrl": "https://login.microsoftonline.com/.../token",
          "scopes": {
            "rca:read": "Read RCA results",
            "rca:execute": "Execute RCA workflow"
          }
        }
      }
    }
  },
  "security": [
    { "oauth2": ["rca:read", "rca:execute"] }
  ]
}
```

## 認証フローのパターン

### 1. Client Agent → Server Agent (同一組織内)

- Client Agentが自分のService Principal/Workload Identityで Client Credentials Flowでトークン取得
- JWTには `client_id`, `scope`, カスタムclaim (部署、環境など)
- Server AgentはJWTを検証してRBAC/ABACで認可

### 2. User → Client Agent → Server Agent (ユーザー主体)

- ユーザートークンをそのまま転送するのはNG (Confused Deputy)
- **OAuth 2.0 Token Exchange (RFC 8693)** でClient Agentが「ユーザー＋自分」の複合トークンを取得

### 3. Cross-organization A2A

- Federated identity (OIDC Federation, SPIFFE/SPIRE) で組織を跨ぐ
- mTLSをトランスポート層で併用するパターンも多い

## A2A特有のRBAC/ABAC設計

- **Skill単位の認可**: Agent Cardには `skills` 配列があるが、A2A仕様の `AgentSkill` に「必要scope」の専用フィールドはない。認可はagentレベルの `security`/`securitySchemes` で表現し、skill単位のscope制御は設計パターンとして実装する (scope名をskillに対応させる等)
- **Task lifecycle**: A2AはTask単位で長時間実行されるので、Taskの作成・取得・キャンセル権限を分離
- **Streaming/Push notification**: SSE/Webhookのコールバック先URLの検証 (SSRF対策含む)

---

# Gateway選択肢

## Envoy AI GatewayのA2A対応状況

Envoy AI Gatewayは **2026年6月23日に v1.0 GA に到達**し、`MCPRoute` CRD が v1 stable に昇格。MCP実装は2025-06-18仕様準拠。
一方、**A2Aネイティブ対応は依然として限定的**でロードマップ段階。A2Aは「HTTPとして通せばJWT検証はできる」レベルにとどまり、A2A固有のskill単位RBACのような深い制御は別ツールに任せる形になる。

## A2Aネイティブ対応ゲートウェイ

| ツール | 特徴 |
|-------|-----|
| **agentgateway** | Rust製フルスクラッチ(Envoyベースではない)、MCP/A2A両方をネイティブ理解。元々Solo.io製だが2025年9月にLinux Foundation (Agentic AI Foundation) へ寄贈済み |
| **Kuadrant MCP Gateway** | A2A対応を設計中 (Issue #766) |
| **LiteLLM Agent Gateway** | A2A JSON-RPCに対応、管理UIあり |

## ゲートウェイ系の選択肢一覧

### APIゲートウェイ系 (MCP/A2Aの前段に立てる)

- **Envoy Gateway + Envoy AI Gateway** — `SecurityPolicy` でJWT検証 + RBAC、`BackendTrafficPolicy` でレート制限。ABACまでやるならext_authzでOPA連携
- **Kong Gateway** — JWT/OAuth2/OPAプラグイン、MCP向けプラグインも出始め
- **Apache APISIX** — 同様のJWT/OAuth/OPAプラグイン、軽量
- **Traefik** — ForwardAuthでOAuth2-Proxy/Authelia連携

### 専用Auth Gateway / Identity-Aware Proxy

- **OAuth2-Proxy + Dex/Keycloak** — 古典的だが堅実
- **Pomerium** — Zero Trust的なIAP、policyでABAC表現可能
- **Ory Oathkeeper + Ory Keto** — Oathkeeperがリクエスト認可、KetoがZanzibarベースのfine-grained authz
- **Cloudflare Access** — マネージドIAP

### Service Mesh (Workload Identity + mTLS)

- **Istio** — `AuthorizationPolicy` でJWT claimベースのRBAC、`PeerAuthentication` でmTLS
- **Linkerd** — mTLS自動、authorization policy
- **SPIFFE/SPIRE** — Workload Identityの標準

### IdP統合製品

- **Auth0 FGA** — OpenFGA商用版
- **Okta Fine Grained Authorization**
- **WorkOS** — B2B SaaS向け、AuthKitでMCP対応進行中
- **Keycloak + Authorization Services** — OSS、UMA 2.0対応

### マネージド

- **AWS Bedrock AgentCore Gateway** — Inbound/Outbound Auth

---

# Policy Engine

## PDPとPEPの役割分担

| 役割 | 説明 | 例 |
|-----|-----|----|
| **PDP (Policy Decision Point)** | 判定する人。「許可していい?」と聞かれて「Yes/No」を返す | OPA, OpenFGA, Cedar |
| **PEP (Policy Enforcement Point)** | 実際に強制する人。PDPに問い合わせて結果に従ってリクエストを通す/拒否する | Gateway, Service Mesh, アプリコード |

**PDPとPEPは必ずペアで必要**。OPAやOpenFGAを使うときには誰がPEPになるかを決める必要がある。

## PEPの選択肢

1. **Gateway** (Envoy AI Gateway, agentgateway等) — ext_authzでPDPに問い合わせ
2. **Service Mesh** (Istio) — AuthorizationPolicyの外部認可
3. **アプリケーションコード** — MCPサーバー自身がコード内でOPAを呼ぶ
4. **Kubernetes Admission Controller** — クラスタレベルの制御
5. **API Server側のミドルウェア** — FastAPIのミドルウェアとしてOPAを呼ぶ

## Policy Engineの種類比較

### OPA (Open Policy Agent)

- **ABAC寄り**: 属性ベースで条件式を評価
- JWT claimをそのまま入力にして判定可能
- IdPの属性だけで判定が完結することが多い
- CNCF Graduated、エコシステム成熟
- Rego言語

```rego
allow if {
    "sre" in input.user.groups
    input.request.cluster == "prod"
}
```

### OpenFGA

- **ReBAC (Relationship-Based)**: リレーション (関係性) で判定
- 「ユーザーAはドキュメントXのowner」みたいな関係を別途FGAに保存
- JWTから取れる属性だけでは不十分で、FGAに「誰がどのリソースとどう関係しているか」のデータが必要
- Google Zanzibarインスパイア

```
user:lee.joonki  is  owner  of  document:design-doc-2025
user:lee.joonki  is  viewer of  folder:platform-team
```

### その他

- **Cedar (AWS)** — AVP (Verified Permissions) でマネージド提供
- **Casbin** — 軽量、多言語SDK

## OPA vs OpenFGA: どっちを選ぶか

### OPAが向くケース (ABAC寄り)

- 「SREチーム & 営業時間内 & 本番環境」みたいな**条件式**で表現できる
- ユーザー属性とリクエスト属性の組み合わせで決まる
- ツール単位の認可 (どのツールを呼べるか)

### OpenFGAが向くケース (ReBAC寄り)

- 「このユーザーがこの特定リソースに対して権限を持つ」というリソース単位
- Google Drive的な「ファイルを誰に共有したか」が動的に変わる
- マルチテナントSaaSで顧客が自分のデータを管理
- RAGで「このユーザーがアクセスできるドキュメントだけ検索結果に含める」
- データ単位の認可 (どのドキュメント、どのクラスタにアクセスできるか)

### 推奨アプローチ

**OPA優先、必要に応じてOpenFGAを併用**

1. まずOPA一本で開始
2. リソース単位の動的権限管理が必要になったらOpenFGAを追加
3. 両方混ぜることもよくある: OPAでツール呼び出し可否 → OpenFGAでリソースアクセス可否

---

# IdPとPolicy Engineの関係

## 役割分担

| レイヤー | 役割 | 例 |
|---------|------|-----|
| IdP | 認証 (AuthN): 「この人は誰か」 | Entra ID, Auth0, Keycloak |
| IdP | 属性ストア: ユーザーのrole/group/attribute保持 | 同上 |
| JWT | 属性の運搬: IdPが発行、claimsに属性を載せる | - |
| Policy Engine (PDP) | 認可 (AuthZ): 「この人はこれをやっていい?」判定 | OPA, OpenFGA, Cedar |
| PEP | 強制: 判定結果に従って通す/拒否 | Gateway, App |

## 全体フロー

```
[1] IdPで認証 + 属性付与
       ↓
[2] JWTに属性が載って発行される
       ↓
[3] Client → PEP (Gateway or App) にJWT付きでリクエスト
       ↓
[4] PEP → PDP (OPA/OpenFGA) に判定問い合わせ
       ↓
[5] PDPが Yes/No 返答
       ↓
[6] PEPが許可/拒否を強制
```

## OPAとOpenFGAでの違い

**OPA**: JWT claim → ポリシー判定 で完結することが多い。追加データストアは必須ではない。

**OpenFGA**: JWTから取れるのは「あなたは○○さんです」というID情報まで。
そのIDがどのリソースとどんな関係を持っているかは、**FGA側のデータベースに入れる必要がある**。

---

# Gatewayあり/なしの比較

## Gatewayなし (各MCPサーバー/A2A Agentで個別実装)

各サーバーで以下を全部やる必要がある:

1. JWT検証 (署名、aud、exp等)
2. OPA/OpenFGAへの問い合わせコード
3. ポリシー違反時のエラーハンドリング
4. OPA/OpenFGAクライアントのサイドカーやライブラリ組み込み

```python
@app.middleware("http")
async def authz_middleware(request, call_next):
    claims = verify_jwt(request)
    allowed = await fga_client.check(
        user=f"user:{claims['sub']}",
        relation="can_call",
        object=f"tool:{extract_tool_name(request)}"
    )
    if not allowed:
        return JSONResponse(403, {"error": "forbidden"})
    return await call_next(request)
```

**向くケース**:
- MCPサーバーが1〜2個しかない小規模構成
- ポリシーがサーバーごとに大きく異なる
- PoC的に動かしたい

**デメリット**:
- サーバーが増えるたびに同じコードを書く
- ポリシー変更時に各サーバーの再デプロイが必要になる場合がある
- 言語が混在すると実装の重複が大変
- 監査ログ、レート制限、メトリクスも各サーバーで実装

## Gatewayあり

認証認可ロジックが一箇所に集約される。各MCPサーバー/A2A Agentは自分のビジネスロジックに集中できる。

```
[Client]
    ↓ JWT
[Gateway]                        ← ここで全部やる
  - JWT検証
  - OPA/OpenFGA問い合わせ
  - 拒否処理
    ↓ (許可されたリクエストのみ転送)
[MCP Server A]   ← 認可ロジック不要
[MCP Server B]
[A2A Agent C]
```

**向くケース**:
- MCPサーバー/A2A Agentが複数ある
- 認可ポリシーを中央集権的に管理したい
- 認可以外の横断的関心事 (レート制限、ロギング、トレース) もまとめたい
- 多言語のサーバーが混在

**デメリット**:
- Gateway自体の運用コスト
- ネットワークホップが1つ増える
- 単一障害点になりうる (HA構成必須)

## 多層防御パターン (実務で多い)

セキュリティが厳しい環境では、Gatewayでも各サーバーでも両方やる多層防御が一般的。

```
[Gateway]
  - 粗い粒度の認可 (このユーザーはこのMCPサーバーにアクセス可?)
  - レート制限、認証検証
    ↓
[MCP Server]
  - 細かい粒度の認可 (このユーザーはこの特定リソースにアクセス可?)
  - ビジネスロジックレベルの認可
```

OpenFGAだと特に**リソース単位の判定はMCPサーバー側に置くことが多い** (Gatewayはリクエストが来た時点でリソースIDを知らないことが多いため)。

---

# Policy Engineの介入ポイント (Envoyベース)

## なぜPolicy Engineを分離するのか

Gatewayの組み込みRBACは「JWT claimとAPIパスのマッチング」程度しかできないことが多い。
以下のような複雑な条件は書きにくい:

- 「SREチームかつ営業時間内なら本番Loki可、それ以外は staging のみ」
- 「このRCA Agentがこのユーザーのために動いているとき、対象クラスタはユーザーが所属する組織のものに限る」
- 「過去5分で10回以上ツール呼び出しがあれば、コスト系ツールはブロック」

判定ロジックを外部のPolicy Engineに切り出す。

## 介入ポイント: ext_authz (External Authorization)

Envoyの `ext_authz` フィルタで、リクエスト処理の途中で外部サービスに「これ通していい?」と問い合わせ。

```
1. Client → Gateway:
   POST /mcp/tools/call
   Authorization: Bearer eyJhbGc...
   Body: { "tool": "query_loki", "args": {"cluster": "prod"} }

2. Gateway (Envoy):
   - JWT signature検証 (IdPの公開鍵で)
   - aud claim検証 (Resource Indicator)
   - 基本的なscope確認
   → ここまでは Gateway 単体で完結

3. Gateway → OPA (ext_authz):
   {
     "input": {
       "user": { "sub": "lee.joonki", "groups": ["sre", "platform"] },
       "request": { "tool": "query_loki", "cluster": "prod" },
       "context": { "time": "14:30", "source_ip": "10.0.1.5" }
     }
   }

4. OPA:
   - Regoポリシー評価
   → { "allow": true }

5. OPA → Gateway: 200 OK (allow)

6. Gateway → MCP Server: リクエスト転送

7. MCP Server → Tool 実行
```

## 介入できる場所のバリエーション

1. **ext_authz (リクエスト前)** — 一番一般的、「通すか拒否するか」判定
2. **ext_proc (リクエスト/レスポンス処理中)** — JSONボディの中身を見て判定や書き換え。MCPはJSON-RPC、A2Aも標準トランスポートがJSON-RPC 2.0 (v0.3以降はgRPC / HTTP+JSON(REST) もオプションで選択可) なので、ボディを見るこっちが重要
3. **Lua/Wasm filter** — 軽量な変換やヘッダ操作
4. **Application層** — MCPサーバー自身がOPAをライブラリとして埋め込む

A2Aの「skill単位の認可」やMCPの「tool単位の認可」は、リクエストボディの中の `method` や `tool_name` を見ないと判定できないので、**ext_procが本質的に必要** (gRPCトランスポートを使う場合はext_procでのプロトコルパースが別途必要になる点に注意)。

## デプロイパターン

```
[Pod: MCP Server]
  - container: mcp-server
  - sidecar: opa (localhost:8181)

[Pod: Envoy AI Gateway]
  - container: envoy
  - ext_authz → opa の Service (Cluster内通信)
  - ext_proc → mcp-protocol-parser
```

OPAは「サイドカーで各MCPサーバーに付ける」か「中央集権的にDeploymentで動かす」かの選択がある。
- レイテンシ重視 → サイドカー
- ポリシー管理重視 → 中央集権

実務では中央のOPAから各サイドカーOPAにポリシーをbundleで配布する形が多い。

---

# OPAポリシーの書き方

## ルールの分割粒度

OPAは「エンドポイントごと」ではなく「ポリシーパッケージ」単位で書く。1つのポリシーで複数エンドポイントをカバーすることも、エンドポイントごとに分けることも可能。

## パターン1: 1つのポリシーで全エンドポイントを判定

```rego
package mcp.authz

default allow := false

# ツール一覧取得は全員可
allow if {
    input.request.path == "/mcp/tools/list"
}

# Lokiクエリツールはsreグループのみ
allow if {
    input.request.path == "/mcp/tools/call"
    input.request.body.tool == "query_loki"
    "sre" in input.user.groups
}

# Grafana設定変更ツールはplatform-adminのみ
allow if {
    input.request.path == "/mcp/tools/call"
    input.request.body.tool == "update_grafana_alert"
    "platform-admin" in input.user.roles
}

# 本番環境への操作は営業時間内のみ
allow if {
    input.request.path == "/mcp/tools/call"
    input.request.body.args.env == "prod"
    "sre" in input.user.groups
    business_hours
}

business_hours if {
    hour := time.clock(time.now_ns())[0]
    hour >= 9
    hour < 18
}
```

メリット: 全体を一望できる、共通ルールを書きやすい
デメリット: ファイルが肥大化する、複数チームで管理しづらい

## パターン2: 機能ごとにパッケージ分割

```rego
# policies/mcp/loki.rego
package mcp.loki

default allow := false

allow if {
    "sre" in input.user.groups
    input.request.body.args.cluster in allowed_clusters
}

allowed_clusters := ["staging", "dev"] if {
    not "sre-prod" in input.user.groups
}

allowed_clusters := ["staging", "dev", "prod"] if {
    "sre-prod" in input.user.groups
}
```

```rego
# policies/mcp/main.rego (エントリポイント)
package mcp.authz

import data.mcp.loki
import data.mcp.grafana

allow if {
    input.request.body.tool == "query_loki"
    loki.allow
}

allow if {
    startswith(input.request.body.tool, "grafana_")
    grafana.allow
}
```

メリット: モジュール化、チーム別に管理可能、テストしやすい
デメリット: 全体把握がやや難しい

## パターン3: A2A Agent向けskill単位

```rego
package a2a.rca_agent

default allow := false

# RCA実行skillはSREのみ
allow if {
    input.request.skill == "execute_rca"
    "sre" in input.user.groups
}

# 結果閲覧skillは関係チーム全員
allow if {
    input.request.skill == "view_rca_result"
    input.user.team in ["sre", "platform", "dev-lead"]
}
```

## 推奨ディレクトリ構成

```
policies/
├── common/
│   ├── jwt_helpers.rego        # JWT claim抽出のヘルパー
│   └── time_helpers.rego       # 営業時間判定など
├── mcp/
│   ├── main.rego               # エントリポイント、ツール名でルーティング
│   ├── observability.rego      # Loki/Tempo/Thanos系ツール
│   ├── grafana.rego            # Grafana管理系ツール
│   ├── aws.rego                # AWS CLI実行系ツール
│   └── github.rego             # GitHub操作系ツール
├── a2a/
│   ├── rca_agent.rego          # RCA Agent
│   └── security_agent.rego     # セキュリティスキャンAgent
└── tests/
    ├── mcp_test.rego           # OPAは単体テスト機能を内蔵
    └── a2a_test.rego
```

OPA bundleで配布して、Gateway側のsidecar OPAが定期的にpullする運用が標準的。

## ext_authzが受け取るinputの実例

```json
{
  "attributes": {
    "request": {
      "http": {
        "method": "POST",
        "path": "/mcp/tools/call",
        "headers": {
          "authorization": "Bearer eyJhbGc...",
          "x-forwarded-for": "10.0.1.5"
        },
        "body": "{\"tool\":\"query_loki\",\"args\":{\"cluster\":\"prod\"}}"
      }
    },
    "source": {
      "address": {"socketAddress": {"address": "10.0.1.5"}}
    }
  },
  "parsed_path": ["mcp", "tools", "call"],
  "parsed_body": {"tool": "query_loki", "args": {"cluster": "prod"}},
  "parsed_jwt": {
    "sub": "lee.joonki@kinto.jp",
    "groups": ["sre", "platform"],
    "scope": "mcp:tools:execute"
  }
}
```

## OpenFGAの場合のルール定義

OpenFGAは「ポリシー (条件式)」ではなく「Authorization Model」というスキーマを定義する。

```
model
  schema 1.1

type user

type mcp_tool
  relations
    define caller: [user, team#member]

type team
  relations
    define member: [user]
    define admin: [user]
```

リレーションデータを別途登録:
```
team:sre#member ⊃ user:lee.joonki
mcp_tool:query_loki#caller ⊃ team:sre#member
```

判定:
```
check(user:lee.joonki, caller, mcp_tool:query_loki) → true
```

**OPAが「ロジック中心」なのに対し、OpenFGAは「データ中心」**。

---

# 推奨構成 (現環境ベース)

既存の Envoy AI Gateway + Entra ID 構成をベースに、A2A/MCP両対応 & ABAC化する場合:

```
[Agent/User]
  ↓ JWT (Entra ID, aud=mcp-rca)
[Envoy AI Gateway]
  - SecurityPolicy: JWT検証, 基本scope確認 (粗い粒度のRBAC)
  - ext_proc → mcp-router (tool名抽出)
  - ext_authz → OPA (ABAC判定)
  ↓
[MCP Server / A2A Agent]
  - 細かい粒度の認可 (リソース単位、必要ならOpenFGA併用)
```

## 段階的な導入ステップ

1. **フェーズ1**: OPA一本で開始
   - Envoy AI Gateway の ext_authz → OPA sidecar
   - JWT claims (Entra ID) + リクエスト属性で判定
   - ツール/サーバー単位のRBAC + 部署/環境などのABAC

2. **フェーズ2**: 必要に応じてOpenFGA追加
   - RCA Agentでユーザーが見られるクラスタを動的管理
   - マルチテナントAI Gatewayで顧客ごとのリソース分離

3. **フェーズ3**: A2Aネイティブ対応が必要なら
   - agentgateway (Linux Foundation / 旧Solo.io) への移行検討
   - もしくは LiteLLM Agent Gateway

## 判断基準まとめ

- MCPサーバーが5個以上並ぶ想定 → **Gateway導入はほぼ必須**
- 単一の小規模アプリ → アプリ内蔵でも可
- A2Aプロトコルレベルの深い制御が必要 → Envoy AI Gateway以外を検討
- リソース単位の動的認可 → OpenFGA併用
- 条件式ベースの認可で十分 → OPA単体

---

# 参考リンク

- MCP Authorization Spec: https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization
- A2A Protocol (公式サイト): https://a2a-protocol.org/
- A2A Protocol (GitHub): https://github.com/a2aproject/A2A
- OAuth 2.1: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1
- RFC 8707 (Resource Indicators): https://datatracker.ietf.org/doc/html/rfc8707
- RFC 8693 (Token Exchange): https://datatracker.ietf.org/doc/html/rfc8693
- OPA: https://www.openpolicyagent.org/
- OpenFGA: https://openfga.dev/
- Envoy AI Gateway: https://aigateway.envoyproxy.io/
- agentgateway: https://agentgateway.dev/