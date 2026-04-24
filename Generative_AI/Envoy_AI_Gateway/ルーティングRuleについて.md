# Envoy AI Gateway - ルーティング詳細

> **バージョン情報**（2026年4月時点）
> - Envoy AI Gateway: **v0.5.0**（2026-01-23 リリース、Current / Latest）
> - Envoy Gateway: **v1.7.2**（2026-04-17 リリース）
> - 次期 v0.6 は 2026 Q2（4月末〜5月）頃リリース見込み

> Envoy AI Gateway v0.5 時点の仕様をベースに、`AIGatewayRoute` のルーティング機能を網羅的に整理する。`AIGatewayRoute` は内部的に Gateway API の `HTTPRoute` を自動生成する仕組みなので、Envoy Gateway 本体の `BackendTrafficPolicy` と組み合わせて高度なトラフィック制御を実現する。

## 1. AIGatewayRoute の基本構造

`AIGatewayRoute` は以下の階層構造で定義される。

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: my-route
  namespace: envoy-ai-gateway-system
spec:
  parentRefs:              # 接続先の Gateway（v0.3+ 必須。旧 targetRefs は非推奨）
    - name: envoy-ai-gateway
      namespace: envoy-gateway-system
      kind: Gateway
      group: gateway.networking.k8s.io
  rules:                   # ルーティングルールのリスト
    - matches: [...]       # 条件
      backendRefs: [...]   # 条件にマッチしたときの転送先バックエンド
      timeouts: {...}      # タイムアウト
      modelsOwnedBy: ""    # /models エンドポイント用
      modelsCreatedAt: ""  # /models エンドポイント用
  llmRequestCosts: [...]   # トークンコスト計測（rate limit 連携用）
```

### 1.1 `parentRefs` と `targetRefs` の違い（v0.3 以降）

古いドキュメントやサンプルコードには `AIGatewayRoute.spec.targetRefs` を使っている例があるが、**v0.3 以降は非推奨で、`parentRefs` に移行されている**。v0.3 リリースノートより：

> "AIGatewayRoute's targetRefs Pattern: The targetRefs pattern is no longer supported for AIGatewayRoute. Existing configurations will continue to work but should be migrated to parentRefs."

新規構成では必ず `parentRefs` を使う。既存の `targetRefs` は動作するが v0.4 以降で削除される可能性があるため早めに移行すべき。

### 1.2 AIServiceBackend の API Schema（参考）

ルーティング先の `AIServiceBackend.spec.schema.name` で指定できるのは以下（v0.5 時点）：

| Schema | 用途 |
|---|---|
| `OpenAI` | OpenAI API 互換（OpenAI Platform、Together, Groq, DeepSeek 等）|
| `AWSBedrock` | AWS Bedrock（Converse API 形式）|
| `AWSAnthropic` | AWS Bedrock 上で Anthropic Messages API 形式（v0.4+）|
| `Anthropic` | Anthropic ネイティブ API（v0.4+）|
| `AzureOpenAI` | Azure OpenAI |
| `GCPVertexAI` | GCP Vertex AI（Gemini 等）|
| `GCPAnthropic` | GCP Vertex AI 上の Claude |
| `Cohere` | Cohere |

AI Gateway は `AIGatewayRoute` の入力 schema（通常 OpenAI）と `AIServiceBackend` の出力 schema の間で自動変換を行う。これが AI Gateway の核心機能の一つ。

## 2. ルールマッチング (`matches`)

### 2.1 ヘッダーマッチのみサポート

`AIGatewayRoute.rules[].matches` は **HTTP ヘッダーマッチのみ**対応。パスマッチやメソッドマッチは `AIGatewayRoute` からは直接指定できない（パス別の振り分けは Endpoint ごとに Gateway 側で行う）。

```yaml
rules:
  - matches:
      - headers:
          - type: Exact                  # Exact / RegularExpression
            name: x-ai-eg-model          # ヘッダー名
            value: claude-sonnet-4-6     # ヘッダー値
```

### 2.2 `x-ai-eg-model` という特別なヘッダー

`x-ai-eg-model` は Envoy AI Gateway が**自動的に付与する特別なヘッダー**。リクエストボディの `model` フィールドを ExtProc が抽出して、このヘッダーにセットしてからルーティング判定が走る。

つまり、クライアントは通常の OpenAI 形式でリクエストを送るだけで良い：

```json
{
  "model": "claude-sonnet-4-6",   // ← ここが x-ai-eg-model に自動コピーされる
  "messages": [...]
}
```

クライアントは `x-ai-eg-model` ヘッダーを自分で付ける必要はない。

### 2.3 複数条件の AND / OR

一つの `matches` エントリに複数のヘッダーを書くと **AND** 条件：

```yaml
matches:
  - headers:
      - type: Exact
        name: x-ai-eg-model
        value: claude-sonnet-4-6
      - type: Exact                      # AND（両方マッチで成立）
        name: x-tenant-id
        value: premium
```

別々の `matches` エントリに分けると **OR** 条件（いずれかにマッチすればルール成立）：

```yaml
matches:
  - headers:                             # パターン A
      - type: Exact
        name: x-ai-eg-model
        value: claude-sonnet-4-6
  - headers:                             # パターン B（OR）
      - type: Exact
        name: x-ai-eg-model
        value: claude-opus-4-7
```

### 2.4 ルール順序による優先度

同一 `AIGatewayRoute` 内で複数ルールが定義されている場合の選択規則は Gateway API の `HTTPRoute` の仕様に従う：

1. より具体的なマッチ（ヘッダー数が多い等）が優先
2. 具体性が同じなら、定義順で最初にマッチしたルールが使われる

## 3. backendRefs の配分ロジック

`backendRefs` は 1 つのルールに対して複数のバックエンドを指定でき、**`weight` と `priority` の 2 軸**で配分が決まる。

### 3.1 `weight`（重み付け配分）

`weight` は **同一 priority 内での比率配分**を指定する。Gateway API の `BackendRef.weight` と完全に同じ意味。

| フィールド | 型 | デフォルト | 意味 |
|---|---|---|---|
| `weight` | integer | `1` | 同一 priority 内での比率配分 |

### 3.2 `priority`（優先度 / フェイルオーバー）

`priority` は **Envoy のプライオリティベースロードバランシング**にマッピングされる。

| フィールド | 型 | デフォルト | 意味 |
|---|---|---|---|
| `priority` | integer | `0` | 数値が **小さいほど優先**。primary = 0、fallback = 1 |

公式ドキュメントの注意書き：

> "Priority is the priority of the backend. This sets the priority on the underlying endpoints. ... Note: This will override the `fallback` property of the underlying Envoy Gateway Backend"

つまり `priority` を指定すると、Envoy Gateway の `Backend` CRD が持つ `fallback` プロパティを上書きする。

### 3.3 `weight` と `priority` の組み合わせ

この 2 軸を理解すると、以下のパターンを自在に組める：

| パターン | 構成 | 挙動 |
|---|---|---|
| **A/B テスト / Active-Active** | 同じ `priority`、異なる `weight` | 重み比でリクエストを振り分ける |
| **フェイルオーバー / Active-Passive** | 異なる `priority` | primary が unhealthy になると fallback に切り替わる |
| **ハイブリッド** | 複数 `priority` + `priority` 内で `weight` 指定 | 優先層内で重み付け、不調時に次層へ |

## 4. パーセンテージルーティング（`weight` ベース）

### 4.1 基本形：50/50 の振り分け

`weight` のみを指定して、`priority` を省略（またはすべて `0`）することで、指定の比率でリクエストが振り分けられる。

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: ab-test-route
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
              value: claude-sonnet-4-6
      backendRefs:
        - name: envoy-ai-gateway-aws                                    # AWS Bedrock
          modelNameOverride: global.anthropic.claude-sonnet-4-6
          weight: 50                                                     # 50%
        - name: envoy-ai-gateway-gcp                                    # GCP Vertex AI
          modelNameOverride: claude-sonnet-4@20250514
          weight: 50                                                     # 50%
```

このとき両バックエンドとも `priority` のデフォルト値 `0` を共有しているため、**同一プライオリティ内の重み配分**となる。クライアントは統一されたモデル名 `claude-sonnet-4-6` を送るだけで、Gateway が AWS / GCP に振り分ける。これが公式で言う **Model Name Virtualization** の具体例。

### 4.2 カナリアデプロイの例（90/10）

新プロバイダへの切替を段階的に進める場合：

```yaml
backendRefs:
  - name: openai-backend            # 既存プロバイダ
    weight: 90
  - name: azure-openai-backend      # 新プロバイダ（カナリア）
    weight: 10
```

問題なければ 90/10 → 50/50 → 10/90 → 0/100 と段階的に移行する。

### 4.3 N 分割の例（3 分割）

3 つ以上のバックエンドでも可能：

```yaml
backendRefs:
  - name: openai-backend
    weight: 40
  - name: azure-openai-backend
    weight: 40
  - name: bedrock-backend
    weight: 20
```

**重要**: `weight` は **比率**であって**パーセント**ではない。`weight` の合計は 100 である必要はない（例: `weight: 2` + `weight: 3` は 40%/60% の意味）。ただし分かりやすさのため合計 100 にするのが慣例。

### 4.4 同一プロバイダ内でのモデル重み付け（モデル間 A/B）

`modelNameOverride` を使うと、**同じバックエンド内で異なるモデル**への重み付けも可能：

```yaml
backendRefs:
  - name: openai-backend
    modelNameOverride: gpt-5-nano       # 70% を gpt-5-nano
    weight: 70
  - name: openai-backend
    modelNameOverride: gpt-5-mini       # 30% を gpt-5-mini
    weight: 30
```

新モデル評価時に、本番トラフィックの一部を新モデルに流して比較評価する用途に使える。

## 5. フォールバック（`priority` ベース）

### 5.1 基本形：Primary + Fallback

`priority` の数値が **小さいほど優先**。primary backend を `priority: 0`、fallback を `priority: 1` に設定する：

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: AIGatewayRoute
metadata:
  name: fallback-route
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
              value: claude-sonnet-4-6
      backendRefs:
        - name: envoy-ai-gateway-aws                                    # Primary
          modelNameOverride: global.anthropic.claude-sonnet-4-6
          priority: 0
        - name: envoy-ai-gateway-gcp                                    # Fallback
          modelNameOverride: claude-sonnet-4@20250514
          priority: 1
```

Primary が以下の条件を満たすと、Fallback に自動的に切り替わる：
- **コネクション失敗**（connect-failure）
- **5xx レスポンス**
- **レート制限（429）**
- **タイムアウト**

### 5.2 多段フォールバック

`priority` を 0 → 1 → 2 と増やすことで、多段フォールバックが可能：

```yaml
backendRefs:
  - name: provider-primary                 # 最優先
    priority: 0
  - name: provider-fallback-1              # 第一フォールバック
    priority: 1
  - name: provider-fallback-2              # 第二フォールバック
    priority: 2
```

### 5.3 同一プロバイダ内での安価モデルへのフォールバック

`modelNameOverride` を組み合わせると、**同じプロバイダ内で「高いモデル → 安いモデル」へのフォールバック**も可能。レート制限に達した際のダウングレード戦略に使える：

```yaml
rules:
  - matches:
      - headers:
          - type: Exact
            name: x-ai-eg-model
            value: gpt-5-nano
    backendRefs:
      - name: openai-backend                 # Primary: gpt-5-nano
        # modelNameOverride なし → リクエストの model 値 (gpt-5-nano) がそのまま使われる
        priority: 0
      - name: openai-backend                 # Fallback: gpt-5-nano-mini
        modelNameOverride: gpt-5-nano-mini
        priority: 1
```

### 5.4 フォールバックを正しく動作させる BackendTrafficPolicy 設定

**重要**: `priority` だけを指定しても、**`BackendTrafficPolicy` の retry 設定を正しく行わないとフォールバックは期待通り動作しない**。

公式ドキュメントの推奨設定は以下：

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: provider-fallback
  namespace: envoy-ai-gateway-system
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute                        # AIGatewayRoute が自動生成した HTTPRoute
      name: fallback-route                   # AIGatewayRoute と同名
  retry:
    numAttemptsPerPriority: 1                # ★ 最重要: 各 priority で 1 回だけ試行
    numRetries: 5                            # 全体のリトライ予算
    perRetry:
      backOff:
        baseInterval: 100ms
        maxInterval: 10s
      timeout: 30s
    retryOn:
      httpStatusCodes:
        - 500
      triggers:
        - connect-failure
        - retriable-status-codes
```

#### 5.4.1 `numAttemptsPerPriority` が最重要

**`numAttemptsPerPriority` は Envoy Gateway v1.5.1 で追加されたフィールド**で、**これがないと fallback が期待通り動かない**。公式ドキュメント（Envoy Gateway）の定義：

> "NumAttemptsPerPriority defines the number of requests (initial attempt + retries) that should be sent to the same priority before switching to a different one. If not specified or set to 0, all requests are sent to the highest priority that is healthy."

- `numAttemptsPerPriority: 1` を指定すると、**primary で 1 回試行 → 失敗したら即 fallback に切り替え** という挙動になる
- これを指定しないと、**primary のリトライを全部消費するまで fallback に行かない**（primary が本当に死んでないと fallback されない）
- 公式の provider-fallback 例でも明示的にコメントされている：
  > "This ensures that only one attempt is made per priority. For example, if the primary backend fails, it will not retry on the same backend."

#### 5.4.2 `numRetries` と `numAttemptsPerPriority` の関係

| フィールド | 意味 | デフォルト |
|---|---|---|
| `numRetries` | **全体**のリトライ回数（初回試行以外）| 2 |
| `numAttemptsPerPriority` | **同じ priority** 内で何回試行するか | 0（= 同 priority 内でリトライし続ける）|

設定例の動作イメージ（`numAttemptsPerPriority: 1, numRetries: 5`、primary が毎回失敗する場合）：

1. **priority 0** (primary) で 1 回試行 → 失敗
2. `numAttemptsPerPriority: 1` に達したので **priority 1** (fallback) に切り替え
3. fallback で 1 回試行 → 失敗なら priority 2 があればそちらへ
4. priority が尽きるか、numRetries=5 回に達するまで続く

#### 5.4.3 ありがちな間違い

```yaml
# ❌ 間違い: numAttemptsPerPriority なしで numRetries: 1
retry:
  numRetries: 1

# この場合、primary で 1 回だけリトライして終わり。fallback には絶対行かない
```

```yaml
# ❌ 間違い: numAttemptsPerPriority なしで numRetries が大きい
retry:
  numRetries: 5
  retryOn: {...}

# この場合、primary で 5 回リトライを続けてしまう。fallback に切り替わらない
```

```yaml
# ✅ 正しい
retry:
  numAttemptsPerPriority: 1
  numRetries: 5
  retryOn: {...}
```

### 5.5 `weight` + `priority` の組み合わせ：階層化配分

```yaml
backendRefs:
  # 優先層：OpenAI 70% / Azure OpenAI 30% で active-active
  - name: openai-backend
    priority: 0
    weight: 70
  - name: azure-openai-backend
    priority: 0
    weight: 30
  # フォールバック層：両方失敗したら Bedrock
  - name: bedrock-backend
    priority: 1
    weight: 1
```

通常時は OpenAI と Azure に 70/30 で配分され、両方とも失敗した場合だけ Bedrock にフォールバック。

## 6. Model Name Virtualization

### 6.1 `modelNameOverride` の役割

`modelNameOverride` は各 `backendRef` ごとに指定可能なフィールドで、**リクエストの `model` フィールドをバックエンド固有の値に書き換える**。

```yaml
backendRefs:
  - name: aws-backend
    modelNameOverride: anthropic.claude-sonnet-4-20250514-v1:0
    weight: 50
  - name: gcp-backend
    modelNameOverride: claude-sonnet-4@20250514
    weight: 50
```

クライアント側は統一されたモデル名（`claude-sonnet-4-6` 等）を使い続けながら、Gateway が内部でプロバイダ固有のモデル名に変換する。

### 6.2 プロバイダ別のモデル名解決動作

各プロバイダは異なる「モデル名解決」戦略を採用している。これを理解すると `modelNameOverride` の使い方が適切に判断できる：

| プロバイダ | モデル名解決戦略 | レスポンス内 model フィールド |
|---|---|---|
| **OpenAI** | Automatic Routing（`gpt-5-nano` → 実際は `gpt-5-nano-2025-08-07`）| 実際に使われた model 名が返る |
| **AWS Bedrock** | Static Execution（完全固定。例: `anthropic.claude-sonnet-4-20250514-v1:0`）| model フィールドなし（Converse API）|
| **AWS Bedrock (AWSAnthropic)** | Static Execution（同上）| model フィールドなし |
| **GCP Vertex AI** | URL パス内のバージョン固定（例: `gemini-1.5-pro-002`）| model フィールドなし |
| **GCP Anthropic** | URL パス内のバージョン固定（例: `claude-sonnet-4@20250514`）| model フィールドなし |
| **Azure OpenAI** | **リクエストの `model` フィールドは無視**、URL 内の deployment 名で決まる | 実際に使われた model 名が返る |

`Azure OpenAI` が特殊で、body の `model` 値は完全に無視される。そのため `modelNameOverride` で書き換えても Azure 側には影響せず、Azure は URL 内の deployment 名で決まる。`Backend` の FQDN 設計で調整が必要。

### 6.3 使いどころ

1. **同一モデルを複数プロバイダに展開**
2. **プロバイダ間の weight 配分**
3. **モデルのバージョン切替**
4. **ダウングレードフォールバック**

### 6.4 用語の整理

- `x-ai-eg-model` = クライアントが送る論理モデル名（ルーティングキー）
- `modelNameOverride` = プロバイダに送る実際のモデル名
- この 2 つの分離が **Virtualization** の本質

## 7. クロスネームスペース参照

v0.4+ で `backendRefs` に `namespace` フィールドが追加され、別 namespace の `AIServiceBackend` を参照できるようになった。

```yaml
backendRefs:
  - name: shared-openai-backend
    namespace: platform-shared              # 別 namespace を指定
    modelNameOverride: gpt-4
```

**要件**: 参照先の namespace に `ReferenceGrant` リソースを作成して、クロス namespace 参照を明示的に許可する必要がある。

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: allow-ai-gateway-access
  namespace: platform-shared                 # 参照先 namespace
spec:
  from:
    - group: aigateway.envoyproxy.io
      kind: AIGatewayRoute
      namespace: tenant-a                    # 参照元 namespace
  to:
    - group: aigateway.envoyproxy.io
      kind: AIServiceBackend
```

マルチテナント構成や、プラットフォームチームが提供する共通バックエンドを各テナントから使う場合に有効。

## 8. タイムアウト設定 (`timeouts`)

ルールレベルで `HTTPRouteTimeouts` が指定可能：

```yaml
rules:
  - matches: [...]
    backendRefs: [...]
    timeouts:
      request: 300s                          # リクエスト全体（リトライ含む）
      backendRequest: 60s                    # バックエンドへの 1 回の試行
```

### 8.1 デフォルト値

公式ドキュメントより：

> "If this field is not set, or the timeout.requestTimeout is nil, Envoy AI Gateway defaults to set 60s for the request timeout as opposed to 15s of the Envoy Gateway's default value."

- **Envoy Gateway デフォルト**: 15 秒
- **Envoy AI Gateway デフォルト**: 60 秒（LLM 応答を考慮）

### 8.2 ストリーミングレスポンス時の注意

SSE ストリーミング（`stream: true`）を使う場合、Claude Opus 4.7 の extended thinking 等で数分間無音になることがある。そうした用途では **30m 以上**に設定するのが安全：

```yaml
timeouts:
  request: 30m                               # 長時間の推論に対応
  backendRequest: 30m
```

## 9. ヘッダー / ボディ変換

### 9.1 `headerMutation` と `bodyMutation` の階層

v0.4+ で `AIServiceBackend` に、v0.5+ で `AIGatewayRoute.rules[].backendRefs[]` にも `headerMutation` と `bodyMutation` が追加された。

| レベル | リソース | スコープ |
|---|---|---|
| **Backend レベル** | `AIServiceBackend.spec.headerMutation` / `bodyMutation` | このバックエンドへのリクエスト全てに適用 |
| **Route レベル** | `AIGatewayRoute.rules[].backendRefs[].headerMutation` / `bodyMutation` | 特定のルール + 特定のバックエンドの組み合わせのみに適用 |

**優先順位**: 競合する操作については **Route レベルが Backend レベルを上書き**。競合しない操作は両方とも適用される。

### 9.2 `headerMutation` の使い方

```yaml
headerMutation:
  set:                                       # 追加 / 上書き
    - name: x-tenant-tier
      value: premium
  remove:                                    # 削除
    - x-internal-debug
```

最大 **16 エントリ**まで。ヘッダー名は大文字小文字を区別しない。

### 9.3 `bodyMutation` の使い方

リクエストボディの **トップレベル JSON フィールド**のみ編集可能（ネストされたフィールドは非対応）。

```yaml
bodyMutation:
  set:
    - path: service_tier                     # トップレベルのみ
      value: '"scale"'                       # 文字列は '"..."' で囲む
    - path: max_tokens
      value: "8192"                          # 数値はそのまま
    - path: temperature
      value: "0.7"
  remove:
    - internal_tracking_id
```

**value の記法**:
- 文字列 → `'"scale"'`（内側に二重引用符、外側に一重引用符）
- 数値 / 真偽値 → `"8192"` / `"true"`
- オブジェクト / 配列 → `'{"key": "value"}'` / `'[1,2,3]'`

### 9.4 実用例：Bedrock の service_tier を route 別に設定

```yaml
rules:
  # 低レイテンシ重視
  - matches:
      - headers:
          - type: Exact
            name: x-ai-eg-model
            value: claude-sonnet-4-6-realtime
    backendRefs:
      - name: envoy-ai-gateway-aws
        modelNameOverride: global.anthropic.claude-sonnet-4-6
        bodyMutation:
          set:
            - path: service_tier
              value: '"priority"'            # 優先ティア（高い）
  # コスト重視（バッチ処理）
  - matches:
      - headers:
          - type: Exact
            name: x-ai-eg-model
            value: claude-sonnet-4-6-batch
    backendRefs:
      - name: envoy-ai-gateway-aws
        modelNameOverride: global.anthropic.claude-sonnet-4-6
        bodyMutation:
          set:
            - path: service_tier
              value: '"flex"'                # 柔軟ティア（安い）
```

## 10. `/models` エンドポイント用メタデータ

### 10.1 `modelsOwnedBy`

OpenAI 互換の `GET /v1/models` エンドポイントを叩いたときの `owned_by` フィールドの値を指定する。この情報はルーティング動作には影響せず、単に **API レスポンスの整形用**。

```yaml
rules:
  - matches:
      - headers:
          - type: Exact
            name: x-ai-eg-model
            value: claude-sonnet-4-6
    backendRefs: [...]
    modelsOwnedBy: my-organization            # /models で表示される owned_by
```

デフォルトは `Envoy AI Gateway`。

### 10.2 `modelsCreatedAt`

同様に、`/models` レスポンスの `created` フィールド用のタイムスタンプ。

```yaml
rules:
  - matches: [...]
    backendRefs: [...]
    modelsCreatedAt: "2026-02-18T00:00:00Z"  # RFC 3339 形式
```

デフォルトは `AIGatewayRoute` 自体の作成タイムスタンプ。

### 10.3 用途

OpenAI 互換クライアントが `/v1/models` を叩いてモデル一覧を取得するとき、これらのフィールドが以下のように表示される：

```json
{
  "object": "list",
  "data": [
    {
      "id": "claude-sonnet-4-6",
      "object": "model",
      "created": 1739836800,
      "owned_by": "my-organization"
    }
  ]
}
```

## 11. LLM リクエストコスト計測 (`llmRequestCosts`)

### 11.1 基本概念

`llmRequestCosts` は、**リクエストごとのトークン消費量を Envoy の dynamic metadata に記録する**ための設定。記録された値は、`BackendTrafficPolicy.rateLimit` でトークンベースのレート制限に使える。

```yaml
spec:
  llmRequestCosts:
    - metadataKey: llm_input_token
      type: InputToken
    - metadataKey: llm_output_token
      type: OutputToken
    - metadataKey: llm_total_token
      type: TotalToken
    - metadataKey: llm_cached_input_token         # v0.5+
      type: CachedInputToken
    - metadataKey: llm_cache_creation_input_token # v0.5+
      type: CacheCreationInputToken
```

### 11.2 `LLMRequestCostType` の種類

| タイプ | 意味 |
|---|---|
| `InputToken` | 入力トークン数 |
| `OutputToken` | 出力トークン数 |
| `TotalToken` | 合計トークン数 |
| `CachedInputToken` | キャッシュから読んだトークン数（v0.5+）|
| `CacheCreationInputToken` | キャッシュ作成用トークン数（v0.5+）|
| `ReasoningToken` | 推論トークン数（o1/Claude extended thinking 等）|
| `CEL` | CEL 式で独自計算 |

### 11.3 CEL 式での柔軟なコスト計算

`type: CEL` を使うと、独自のコスト式を定義できる。利用可能な変数：

- `model`: リクエストの model 名（string）
- `backend`: バックエンド名 `name.namespace`（string）
- `input_tokens`, `output_tokens`, `total_tokens`: トークン数（uint）
- `cached_input_tokens`, `cache_creation_input_tokens`, `reasoning_tokens`: 同上

CEL 式の例：

```yaml
llmRequestCosts:
  - metadataKey: weighted_cost
    type: CEL
    cel: |
      model == 'claude-opus-4-7' 
        ? input_tokens * 3 + output_tokens * 15
        : input_tokens + output_tokens * 3
```

プロバイダごとの料金差を式に反映させ、総量ではなく**金額ベース**でレート制限できる。

### 11.4 トークンベースのレート制限との組み合わせ

`BackendTrafficPolicy.rateLimit` と組み合わせると、例えば「**テナントごとに 1 時間あたり 10,000 トークン**」といった制限が実現できる：

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: token-rate-limit
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: my-route                               # AIGatewayRoute と同名
  rateLimit:
    type: Global
    global:
      rules:
        - clientSelectors:
            - headers:
                - name: x-tenant-id
                  type: Distinct                   # テナント別
          limit:
            requests: 10000                        # 10,000 トークン/hour
            unit: Hour
          cost:
            request:
              from: Number
              number: 0                            # リクエスト時点では消費しない
            response:
              from: Metadata
              metadata:
                namespace: io.envoy.ai_gateway    # 固定
                key: llm_total_token               # llmRequestCosts で定義した key
```

メタデータ名前空間は必ず `io.envoy.ai_gateway`。

### 11.5 Global Cost（v0.5+）の継承構造

v0.5+ で追加された `GatewayConfig.spec.globalLLMRequestCosts` を使うと、**Gateway 全体でデフォルトのコスト式を定義**できる。`AIGatewayRoute.spec.llmRequestCosts` があればそちらで上書き、なければグローバル値が使われる。

同じメトリクスキーを複数ルートで使いつつ、特定ルート（premium tier 等）でのみ異なる式を適用する、というユースケースで便利。

### 11.6 v0.5+ の新機能: `QuotaPolicy`

v0.5+ では、より高機能なレート制限用 CRD として `QuotaPolicy` が導入された。`llmRequestCosts` + `BackendTrafficPolicy.rateLimit` に比べて、**モデル別のクォータ**や **CEL 式によるコスト計算**が一つのリソース内で完結する。

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: QuotaPolicy
metadata:
  name: tenant-quota
  namespace: envoy-ai-gateway-system
spec:
  targetRefs:
    - group: aigateway.envoyproxy.io
      kind: AIServiceBackend
      name: envoy-ai-gateway-aws
  # 全モデル共通のクォータ
  serviceQuota:
    costExpression: "input_tokens + output_tokens * 6"
    quota:
      limit: 1000000
      duration: 1h
  # モデル別のクォータ上書き
  perModelQuotas:
    - modelName: global.anthropic.claude-opus-4-7
      quota:
        costExpression: "input_tokens * 3 + output_tokens * 15"
        mode: Exclusive
        defaultBucket:
          limit: 100000
          duration: 1h
        bucketRules:
          - clientSelectors:
              - headers:
                  - name: x-tenant-tier
                    value: premium
            quota:
              limit: 500000
              duration: 1h
```

既存の BackendTrafficPolicy ベースの方法と `QuotaPolicy` は併用可能だが、**新規構築では `QuotaPolicy` の使用が推奨**される方向（ただし v0.5 時点ではまだ新しい機能なので、本番投入は慎重に）。

## 12. InferencePool を使った動的ルーティング

v0.3+ で追加された機能で、**自前ホストの LLM（vLLM 等）に対して、KV-cache 使用率や queue depth など AI 特有のメトリクスに基づいて動的にルーティング**する。

```yaml
rules:
  - matches:
      - headers:
          - type: Exact
            name: x-ai-eg-model
            value: meta-llama/Llama-3.1-8B-Instruct
    backendRefs:
      - group: inference.networking.k8s.io        # ← AIServiceBackend ではない（注意）
        kind: InferencePool                        # Gateway API Inference Extension
        name: vllm-llama3-8b-instruct
```

### 12.1 API Group の注意

Gateway API Inference Extension の API group は以下の 2 通りがドキュメント内に登場する：
- `inference.networking.k8s.io`（公式 API Reference の AIGatewayRouteRuleBackendRef 定義に記載）
- `inference.networking.x-k8s.io`（v0.5 リリースノートの `BackendSecurityPolicy.targetRefs` の記載、v0.3 Endpoint Picker blog の HTTPRoute 例等）

Gateway API Inference Extension 自体のバージョン（v0.5.1 / v1.0）により API group が異なる場合があるので、使用しているバージョンに合わせて指定する。実装前に必ず `kubectl get crd | grep inference` で確認。

### 12.2 制約事項

公式ドキュメントより：

> "When referencing InferencePool resources:
> - Only one InferencePool backend is allowed per rule
> - Cannot mix InferencePool with AIServiceBackend references in the same rule
> - Fallback behavior is handled by the InferencePool's endpoint picker"

- **1 ルールにつき InferencePool は 1 つだけ**
- **AIServiceBackend と混在不可**
- **フォールバックは InferencePool の Endpoint Picker に委ねる**（`priority` / `weight` / `modelNameOverride` / `headerMutation` / `bodyMutation` はすべて無視される）

## 13. 参考：自動生成されるリソース

`AIGatewayRoute` を apply すると、内部で以下の Kubernetes リソースが自動生成される：

| 生成リソース | 名前規則 | 役割 |
|---|---|---|
| `HTTPRoute` (Gateway API) | `AIGatewayRoute` と同名 | 実際のルーティング定義 |
| `HTTPRouteFilter` (Envoy Gateway) | `ai-eg-host-rewrite-${AIGatewayRoute.Name}` | ホスト名書き換え |

すべて `AIGatewayRoute` と同じ namespace に生成される。したがって、`BackendTrafficPolicy` を後付けでアタッチしたい場合は、`AIGatewayRoute` と同名の `HTTPRoute` を `targetRefs` で指定する。

**この動作は将来変更される可能性がある（implementation detail）**ことが公式ドキュメントで明記されている点に注意。

## 14. トラブルシューティングのポイント

### 14.1 ルールがマッチしない

1. **リクエストボディの `model` 値を確認**
   - `model` フィールドが `x-ai-eg-model` にコピーされるので、この値と `matches` が一致している必要がある
2. **`AIGatewayRoute` と `Gateway` が正しく紐付いているか**
   - `parentRefs.namespace` が `Gateway` の実際の namespace と一致しているか
3. **`Gateway.listeners[].allowedRoutes.namespaces.from` が `All` または適切に設定されているか**
   - 別 namespace に `AIGatewayRoute` がある場合特に重要

### 14.2 フォールバックが発生しない

**最も多い原因は `numAttemptsPerPriority` の未設定**（セクション 5.4 参照）。チェックリスト：

1. **`numAttemptsPerPriority: 1` が `BackendTrafficPolicy.retry` に設定されているか**
   - これが未設定だと、primary のリトライだけで完結してしまい fallback に行かない
2. **`priority` が同じ値だと weight 配分になるだけでフォールバックしない**
   - 数値を必ず異なる値にする
3. **`retryOn.triggers` / `retryOn.httpStatusCodes` が実際の失敗条件を含んでいるか**
   - AWS Bedrock の ThrottlingException は 429 として返る。`triggers: [retriable-status-codes]` と `httpStatusCodes: [429, 500, 502, 503, 504]` の両方を設定しておくと安全

### 14.3 `bodyMutation` が効かない

1. **ネストフィールドは非対応**
   - `path: generationConfig.temperature` のような `.` を含む path は動かない
2. **`value` の JSON 記法に注意**
   - 文字列は `'"scale"'`（外側一重引用符、内側二重引用符）
3. **最大 16 エントリ制限**
4. **Azure OpenAI は `model` フィールドを無視する**
   - Azure OpenAI に対して `bodyMutation` で `model` を書き換えても、Azure 側は URL の deployment 名しか見ないので効かない

### 14.4 `llmRequestCosts` が rate limit に反映されない

1. **`BackendTrafficPolicy.rateLimit.cost.response.metadata.namespace` が `io.envoy.ai_gateway` になっているか**
2. **`key` が `llmRequestCosts[].metadataKey` と一致しているか**
3. **`BackendTrafficPolicy.targetRefs` が自動生成された `HTTPRoute`（AIGatewayRoute 同名）を指しているか**

## 15. バージョン別対応状況のまとめ

| 機能 | v0.1 | v0.2 | v0.3 | v0.4 | v0.5 |
|---|---|---|---|---|---|
| `matches` (headers) | ✅ | ✅ | ✅ | ✅ | ✅ |
| `weight` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `priority`（フォールバック）| ⚠️ | ✅ | ✅ | ✅ | ✅ |
| `modelNameOverride` | ❌ | ❌ | ✅ | ✅ | ✅ |
| `parentRefs`（`targetRefs` 非推奨化）| ❌ | ❌ | ✅ | ✅ | ✅ |
| 複数 `AIGatewayRoute` / Gateway | ❌ | ✅ | ✅ | ✅ | ✅ |
| クロスネームスペース参照 | ❌ | ❌ | ❌ | ✅ | ✅ |
| `headerMutation` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `bodyMutation` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `InferencePool` | ❌ | ❌ | ✅ | ✅ | ✅ |
| `GatewayConfig` / Global costs | ❌ | ❌ | ❌ | ❌ | ✅ |
| `CachedInputToken` / `CacheCreationInputToken` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `QuotaPolicy`（新レート制限 API）| ❌ | ❌ | ❌ | ❌ | ✅ |
| `numAttemptsPerPriority`（Envoy Gateway v1.5.1+）| ❌ | ❌ | ❌ | ⚠️ | ✅ |

### 15.1 リリースサイクル

Envoy AI Gateway は **約 3 ヶ月サイクル**でリリースされている：

| バージョン | リリース日 |
|---|---|
| v0.2 | 2025 年 6 月 |
| v0.3 | 2025 年 8 月 21 日 |
| v0.4 | 2025 年 11 月 7 日 |
| v0.5 | 2026 年 1 月 23 日（**現在の最新**）|
| **v0.6** | 2026 年 Q2（4 月末〜5 月頃と推測）|

### 15.2 v0.6 で予定されている機能（v0.5 リリースノートの "Future Work" より）

- **Batch inference APIs** — OpenAI Batch API / Bedrock CreateModelInvocationJob 対応
- **Advanced caching strategies** — prompt cache key / 保持期間制御
- **Upstream provider quota policies** — プロバイダ側のクォータ管理 API
- **AWS Bedrock InvokeModel API support** — Claude / GPT の追加対応
- **Gemini embeddings** — Gemini のエンベディング対応
- **Azure / AKS workload identity** — Azure 側の Pod Identity 対応

## 16. API バージョン（v0.5 時点）

| リソース | 利用可能バージョン | 推奨 |
|---|---|---|
| `AIGatewayRoute` | `v1alpha1`, `v1beta1` | `v1beta1` が公式推奨だが、CRD のインストール状況によって v1alpha1 のみ動作する環境もある |
| `AIServiceBackend` | `v1alpha1`, `v1beta1` | `v1beta1` 推奨 |
| `BackendSecurityPolicy` | `v1alpha1`, `v1beta1` | `v1beta1` 推奨 |
| `GatewayConfig` | `v1alpha1`, `v1beta1` | `v1beta1` 推奨 |
| `MCPRoute` | `v1alpha1` のみ | - |
| `QuotaPolicy` | `v1alpha1` のみ | - |

apply してエラーになる場合は CRD のインストール状況を確認：

```bash
kubectl get crd aigatewayroutes.aigateway.envoyproxy.io -o yaml | grep -A 5 versions
```

`v1beta1` が一覧になければ古い CRD なので Helm chart のアップグレードが必要。

---

## まとめ：主要なルーティングパターン

1. **通常運用**: 1 モデル → 1 プロバイダ（priority 0 固定、weight 1）
2. **高可用性が必要**: 複数プロバイダに `priority` で primary/fallback 設定 + **`BackendTrafficPolicy` で `numAttemptsPerPriority: 1` を必ず設定**
3. **コスト最適化**: 同一プロバイダ内で安価モデルへのダウングレードフォールバック
4. **A/B テスト / カナリアリリース**: `weight` での配分
5. **テナント別クォータ**: `llmRequestCosts` + `BackendTrafficPolicy.rateLimit`（v0.5+ なら `QuotaPolicy` も検討）
6. **リトライ動作のカスタマイズ**: `BackendTrafficPolicy.retry`（自動生成 HTTPRoute に attach）
7. **モデル名の抽象化**: `modelNameOverride` でクライアントコードから provider 固有名を隠蔽