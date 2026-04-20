# Envoy Gateway / Envoy AI Gateway コンポーネント整理

> [!CAUTION]
> 2026年4月時点の情報なので、今後変更される可能性があります。
> - 対象バージョン: 
>   - Envoy AI Gateway v0.5.x
>   - Envoy Gateway v1.6.x+

## 全体像

Envoy AI Gateway は Envoy Gateway の上に構築された **拡張（extension）**。独立した製品ではなく、Envoy Gateway の Pod ライフサイクルに相乗りする形で動作する。

登場する Pod は 3 種類：

| レイヤ | Pod | 作成タイミング | 作成主体 |
|---|---|---|---|
| Control plane | Envoy Gateway Controller | Helm インストール時 | `gateway-helm` |
| Control plane | AI Gateway Controller | Helm インストール時 | `ai-gateway-helm` |
| Data plane | Envoy Proxy Pod（+ ExtProc サイドカー） | `Gateway` CR apply 時に動的生成 | Envoy Gateway Controller + AI Gateway Webhook |

> [!NOTE]
> `Gateway` CR が apply された瞬間に Envoy Gateway Controller が Envoy Proxy Pod を生成するが、その際に AI Gateway Controller の MutatingWebhook が割り込んで ExtProc サイドカーを注入するイメージ。以降、プロバイダ API 変換やトークンカウントはこのサイドカーが担当し、AIGatewayRoute 等による AI 固有のルーティング設定は AI Gateway Controller の Extension Server が Envoy Gateway に提供する。

## 前提知識: xDS とは

xDS = 「x Discovery Service」の総称。Envoy Proxy が設定情報を **動的に gRPC API 経由で受け取る仕組み**。静的 YAML ではなくコントロールプレーンから pull してくる形になっているため、設定変更時に Envoy の再起動が不要。

「x」は具体化して下記の API に分かれる：

| 略称 | 正式名 | 配信するもの |
|---|---|---|
| LDS | Listener Discovery Service | リスナー（待ち受けポート、プロトコル） |
| RDS | Route Discovery Service | HTTP ルーティングルール |
| CDS | Cluster Discovery Service | アップストリームクラスタ（バックエンド群）の定義 |
| EDS | Endpoint Discovery Service | 各クラスタの実エンドポイント（IP:port）一覧 |
| SDS | Secret Discovery Service | TLS 証明書・秘密鍵 |
| ADS | Aggregated Discovery Service | 上記を単一 gRPC ストリームで束ねたもの（順序保証あり） |

実運用では ADS で束ねて配信するのが一般的。

本ドキュメント内の文脈：
- **Envoy Gateway Controller** は `Gateway` / `HTTPRoute` 等の Kubernetes リソースを xDS 設定に変換し、管理下の Envoy Proxy Pod に gRPC で配信する。
- **AI Gateway Controller の Extension Server** はこの xDS 生成プロセスに割り込んで、AI 固有の設定（per-backend upstream filter 等）を追加する役割を担う。

公式仕様: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol

---

## Control Plane コンポーネント

### 1. Envoy Gateway Controller

Envoy Gateway プロジェクト本体のコントローラ。Kubernetes Gateway API の実装。

| 項目 | 内容 |
|---|---|
| Namespace | `envoy-gateway-system` |
| Helm チャート | `oci://docker.io/envoyproxy/gateway-helm` |
| 必要バージョン | v1.6.x 以上（AI Gateway v0.5 連携時） |
| 主な監視リソース | `GatewayClass`, `Gateway`, `HTTPRoute`, `GRPCRoute`, `EnvoyProxy`, `ClientTrafficPolicy`, `BackendTrafficPolicy`, `SecurityPolicy`, `EnvoyExtensionPolicy`, `Backend` |
| 主な役割 | Gateway API リソースを監視し、Envoy Proxy の Deployment / Service / ConfigMap を生成。xDS で Envoy Proxy に設定配信 |
| 設定ファイル | `EnvoyGateway` CR（通常は Helm values から生成される ConfigMap） |

#### AI Gateway 連携に必須の values 設定

AI Gateway v0.3+ では、Envoy Gateway インストール時に AI Gateway 公式の `envoy-gateway-values.yaml` を `-f` で渡すことで拡張登録を行う。

公式ファイル: `https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-values.yaml`

中身の要点：

```yaml
config:
  envoyGateway:
    extensionApis:
      enableEnvoyPatchPolicy: true   # 推奨
      enableBackend: true            # 必須（AI プロバイダへの FQDN 接続用）
    extensionManager:
      hooks:
        xdsTranslator:
          translation:
            listener: {includeAll: true}
            route:    {includeAll: true}
            cluster:  {includeAll: true}
            secret:   {includeAll: true}
          post:
            - Translation
            - Cluster
            - Route
      service:
        fqdn:
          hostname: ai-gateway-controller.envoy-ai-gateway-system.svc.cluster.local
          port: 1063
```

---

### 2. AI Gateway Controller

Envoy Gateway の拡張として動作する、AI 特化機能のコントローラ。

| 項目 | 内容 |
|---|---|
| Namespace | `envoy-ai-gateway-system` |
| Helm チャート | `oci://docker.io/envoyproxy/ai-gateway-helm` |
| CRD チャート | `oci://docker.io/envoyproxy/ai-gateway-crds-helm` |
| Deployment 名 | `ai-gateway-controller` |
| Service | ポート 9443（mutating-webhook）, 1063（grpc）, 8080（http-metrics） |
| 主な監視リソース | `AIGatewayRoute`, `AIServiceBackend`, `BackendSecurityPolicy`, `GatewayConfig`, `MCPRoute`, `QuotaPolicy` |
| 主な役割 | AI 固有リソースの監視 + **MutatingWebhook による ExtProc サイドカー注入** + Envoy Gateway Controller への拡張 xDS 情報の提供（Extension Server） |
| 依存 | Envoy Gateway が先にインストール済みで、かつ AI Gateway 用の values で起動していること |

#### 2 つの重要な仕組み

1. **Extension Server**: Envoy Gateway Controller が xDS 生成時に呼び出す gRPC サービス（port 1063）。AIGatewayRoute 等を Envoy 設定に変換する。
2. **MutatingWebhook**: Envoy Proxy Pod が作られる瞬間に割り込んで、ExtProc コンテナを Pod spec に注入する。`objectSelector` は `app.kubernetes.io/managed-by: envoy-gateway` で Envoy Gateway 管理の Pod のみを対象とする。

#### 実際の values.yaml キー（ai-gateway-helm）

```yaml
controller:
  replicaCount: 1                    # デフォルト 1
  leaderElection:
    enabled: true                    # デフォルト true（複数レプリカ時の split-brain 対策）
  logLevel: info
  watch:
    namespaces: []                   # 空 = 全 namespace 監視
    cacheSyncTimeout: 2m
  serviceAccount:
    create: true
    annotations: {}                  # IRSA 等をここに
  mutatingWebhook:
    tlsCertSecretName: self-signed-cert-for-mutating-webhook
    certManager:
      enable: false                  # production では true + cert-manager 推奨
  image:
    repository: docker.io/envoyproxy/ai-gateway-controller

extProc:
  image:
    repository: docker.io/envoyproxy/ai-gateway-extproc
  logLevel: info
  enableRedaction: false             # debug ログ時にプロンプト等をマスク（production debug 時は true）
  extraEnvVars: []

endpointConfig:
  rootPrefix: "/"                    # AI Gateway が生成するルートの共通接頭辞
  openai: ""                         # → /v1/...
  cohere: "/cohere"                  # → /cohere/v2/...
  anthropic: "/anthropic"            # → /anthropic/v1/...
```

---

## Data Plane コンポーネント

### 3. Envoy Proxy Pod（メインコンテナ）

実際にクライアントからのリクエストを受けて LLM プロバイダに転送するデータプレーン本体。

| 項目 | 内容 |
|---|---|
| Namespace | **`envoy-gateway-system`**（`envoy-ai-gateway-system` ではない） |
| 作成タイミング | `Gateway` CR が apply された瞬間 |
| 作成主体 | Envoy Gateway Controller が Deployment / Service を生成 |
| 中身 | Envoy Proxy バイナリ + xDS クライアント |
| カスタマイズ | `EnvoyProxy` CR を Gateway の `infrastructure.parametersRef` で紐付け |

#### よく触るカスタマイズポイント

- **NLB アノテーション**: `EnvoyProxy.spec.provider.kubernetes.envoyService.annotations` に `service.beta.kubernetes.io/aws-load-balancer-*` を設定
- **リソース**: `EnvoyProxy.spec.provider.kubernetes.envoyDeployment.container.resources`
- **バッファ上限**: `ClientTrafficPolicy` の `connection.bufferLimit` でデフォルト 32KB から 50MB 程度に引き上げ（AI レスポンス用）

---

### 4. ExtProc サイドカー（AI 処理担当）

Envoy Proxy Pod に同居する、AI 固有のリクエスト/レスポンス処理を担うサイドカー。

| 項目 | 内容 |
|---|---|
| Namespace | Envoy Proxy Pod と同じ（`envoy-gateway-system`） |
| 注入方法 | AI Gateway の MutatingWebhook が Pod 作成時に挿入 |
| イメージ | `docker.io/envoyproxy/ai-gateway-extproc` |
| Envoy との通信 | **Unix Domain Socket (UDS)**（ネットワーク gRPC ではない。v0.2 以降でサイドカー + UDS 方式に変更） |
| 主な処理 | モデル名ベースのルーティング判定、プロバイダ API スキーマ変換（OpenAI ↔ Bedrock 等）、トークンカウント、プロバイダ認証情報の付与、トークン使用量のメトリクス発行 |
| リソース等の設定 | `GatewayConfig` CRD（v0.5 新規）経由で Gateway ごとに指定 |

#### GatewayConfig での ExtProc 設定例（v0.5 の新方式）

```yaml
apiVersion: aigateway.envoyproxy.io/v1alpha1
kind: GatewayConfig
metadata:
  name: my-gateway-config
spec:
  extProc:
    kubernetes:
      resources:
        requests: {cpu: 100m, memory: 128Mi}
        limits:   {cpu: 500m, memory: 512Mi}
      env:
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: http://otel-collector:4317
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ai-gateway
  annotations:
    aigateway.envoyproxy.io/gateway-config: my-gateway-config
```

なお v0.4 以前の `AIGatewayRoute.spec.filterConfig.externalProcessor.resources` は非推奨で v0.6 で削除予定。

---

## Helm チャート一覧

| チャート | 代表バージョン（2026-04） | 内容 |
|---|---|---|
| `envoyproxy/gateway-crds-helm` | v1.7.2 | Envoy Gateway 用 CRD（Gateway API + Envoy Gateway 独自） |
| `envoyproxy/gateway-helm` | v1.7.2（v1.6.x+ が AI Gateway の要件） | Envoy Gateway Controller 本体 |
| `envoyproxy/ai-gateway-crds-helm` | v0.5.x | AI Gateway 用 CRD |
| `envoyproxy/ai-gateway-helm` | v0.5.x | AI Gateway Controller + Webhook + ExtProc イメージ |

### インストール手順（Helm 直接実行）

```bash
# 1. Envoy Gateway を AI Gateway 用 values で起動
helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.7.2 \
  --namespace envoy-gateway-system --create-namespace \
  -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-values.yaml

# 2. AI Gateway CRD
helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \
  --version v0.5.0 \
  --namespace envoy-ai-gateway-system --create-namespace

# 3. AI Gateway 本体
helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.5.0 \
  --namespace envoy-ai-gateway-system
```

### アドオン（必要な場合のみ）

Rate Limiting / InferencePool は別途 addon values ファイルを重ね掛けする：

```bash
helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.7.2 \
  --namespace envoy-gateway-system --create-namespace \
  -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-values.yaml \
  -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/examples/token_ratelimit/envoy-gateway-values-addon.yaml \
  -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/examples/inference-pool/envoy-gateway-values-addon.yaml
```

### インストール手順（ArgoCD Application）

ArgoCD の Application リソースとして定義し、GitOps で管理する方式。values.yaml を Git リポジトリで管理し、OCI Helm チャートを参照する構成が一般的。

#### リポジトリ構成例

```
platform-config/                          # ArgoCD が参照する Git リポジトリ
├── apps/
│   ├── envoy-gateway-crds.yaml           # Application マニフェスト
│   ├── envoy-gateway.yaml
│   ├── ai-gateway-crds.yaml
│   └── ai-gateway.yaml
└── values/
    ├── envoy-gateway/
    │   └── values.yaml                   # AI Gateway 連携用 values + 自社カスタマイズ
    └── ai-gateway/
        └── values.yaml
```

`values/envoy-gateway/values.yaml` は AI Gateway 公式の `envoy-gateway-values.yaml` の中身をコピーして、自社のリソース設定等を追記したもの：

```yaml
# AI Gateway 連携に必須の部分（公式 envoy-gateway-values.yaml より）
config:
  envoyGateway:
    extensionApis:
      enableEnvoyPatchPolicy: true
      enableBackend: true
    extensionManager:
      hooks:
        xdsTranslator:
          translation:
            listener: {includeAll: true}
            route:    {includeAll: true}
            cluster:  {includeAll: true}
            secret:   {includeAll: true}
          post: [Translation, Cluster, Route]
      service:
        fqdn:
          hostname: ai-gateway-controller.envoy-ai-gateway-system.svc.cluster.local
          port: 1063

# 自社カスタマイズ
deployment:
  replicas: 2
  envoyGateway:
    resources:
      requests: {cpu: 100m, memory: 256Mi}
      limits:   {cpu: 500m, memory: 1Gi}
```

#### Application マニフェスト

**① Envoy Gateway CRDs**（sync wave 0）

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: envoy-gateway-crds
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "0"
spec:
  project: platform
  source:
    repoURL: docker.io/envoyproxy
    chart: gateway-crds-helm
    targetRevision: v1.7.2
    helm:
      parameters:
      - {name: crds.gatewayAPI.enabled,  value: "true"}
      - {name: crds.gatewayAPI.channel,  value: "standard"}
      - {name: crds.envoyGateway.enabled, value: "true"}
  destination:
    server: https://kubernetes.default.svc
    namespace: envoy-gateway-system
  syncPolicy:
    automated: {prune: true, selfHeal: true}
    syncOptions:
    - CreateNamespace=true
    - ServerSideApply=true        # CRD サイズ対策（必須）
```

**② Envoy Gateway Controller**（sync wave 1、values は Git 側を参照）

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: envoy-gateway
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  project: platform
  sources:
  - repoURL: docker.io/envoyproxy
    chart: gateway-helm
    targetRevision: v1.7.2
    helm:
      valueFiles:
      - $values/values/envoy-gateway/values.yaml
  - repoURL: https://github.com/<org>/platform-config.git
    targetRevision: main
    ref: values
  destination:
    server: https://kubernetes.default.svc
    namespace: envoy-gateway-system
  syncPolicy:
    automated: {prune: true, selfHeal: true}
    syncOptions:
    - CreateNamespace=true
```

**③ AI Gateway CRDs**（sync wave 0）

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ai-gateway-crds
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "0"
spec:
  project: platform
  source:
    repoURL: docker.io/envoyproxy
    chart: ai-gateway-crds-helm
    targetRevision: v0.5.0
  destination:
    server: https://kubernetes.default.svc
    namespace: envoy-ai-gateway-system
  syncPolicy:
    automated: {prune: true, selfHeal: true}
    syncOptions:
    - CreateNamespace=true
    - ServerSideApply=true
```

**④ AI Gateway Controller**（sync wave 2）

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ai-gateway
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  project: platform
  sources:
  - repoURL: docker.io/envoyproxy
    chart: ai-gateway-helm
    targetRevision: v0.5.0
    helm:
      valueFiles:
      - $values/values/ai-gateway/values.yaml
  - repoURL: https://github.com/<org>/platform-config.git
    targetRevision: main
    ref: values
  destination:
    server: https://kubernetes.default.svc
    namespace: envoy-ai-gateway-system
  syncPolicy:
    automated: {prune: true, selfHeal: true}
    syncOptions:
    - CreateNamespace=true
```

#### App-of-Apps でまとめる場合

4 つの Application を個別に apply するのが面倒なら、親 Application を 1 つ作って `apps/` 配下を監視させる：

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: envoy-ai-gateway-stack
  namespace: argocd
spec:
  project: platform
  source:
    repoURL: https://github.com/<org>/platform-config.git
    targetRevision: main
    path: apps
    directory:
      recurse: true
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated: {prune: true, selfHeal: true}
```

この親 App を一度だけ apply すれば、以降 Git の変更が自動的に反映される。

#### 運用上の注意

- **Sync wave の設計**: CRD（wave 0）→ Envoy Gateway Controller（wave 1）→ AI Gateway Controller（wave 2）の順で起動する。AI Gateway Controller は Envoy Gateway の Extension Server 登録を前提とするため、後段に置く。
- **ServerSideApply=true は CRD で必須**: gateway-crds-helm の CRD は 2MB を超えるため、client-side apply では `metadata.annotations: Too long` エラーで失敗する。
- **OCI リポジトリの事前登録**: `docker.io/envoyproxy` は匿名アクセス可能だが、ArgoCD によっては `argocd repo add --type helm --enable-oci` で事前登録が必要な場合がある。
- **values.yaml の同期**: Git 側の values を変更すると ArgoCD が自動で Helm テンプレートを再生成して差分適用する。Envoy Gateway の場合は ConfigMap 更新に伴い Deployment の rollout が必要なこともあるので、必要に応じて `kubectl rollout restart` で明示的に再起動する。
- **targetRevision の固定**: production では `targetRevision: v1.7.2` のように固定バージョンを指定すること。`HEAD` や branch 指定だとチャート提供側の更新に引きずられて意図しない変更が入る。

---

## 主要 CRD 一覧

### Kubernetes Gateway API 標準（API group: `gateway.networking.k8s.io`）

| CRD | 役割 |
|---|---|
| `GatewayClass` | Gateway 実装（Envoy Gateway）の宣言 |
| `Gateway` | リスナー定義（port / protocol / TLS） |
| `HTTPRoute` | 標準的な HTTP ルーティング |

### Envoy Gateway 拡張（API group: `gateway.envoyproxy.io`）

| CRD | 役割 |
|---|---|
| `EnvoyProxy` | データプレーンの挙動（リソース、アノテーション、ログ等） |
| `ClientTrafficPolicy` | クライアント側のバッファ、タイムアウト、TLS |
| `BackendTrafficPolicy` | バックエンド側の接続プール、リトライ、サーキットブレーカー |
| `SecurityPolicy` | JWT / OIDC / CORS 等の認証認可 |
| `Backend` | クラスタ外 FQDN を宛先として定義（要 `enableBackend: true`） |
| `EnvoyExtensionPolicy` | Wasm / Lua 等の拡張フィルタ |
| `EnvoyPatchPolicy` | xDS を直接パッチする高度な手段 |

### Envoy AI Gateway 拡張（API group: `aigateway.envoyproxy.io`）

| CRD | API バージョン | 役割 |
|---|---|---|
| `AIGatewayRoute` | v1alpha1（非推奨）/ v1beta1 | AI 固有ルーティング（モデル名マッチ、フェイルオーバー、トークンレート制限）。v0.4+ で `schema` フィールドは不要に |
| `AIServiceBackend` | v1alpha1（非推奨）/ v1beta1 | LLM プロバイダの定義（OpenAI / Bedrock / Azure / Anthropic / GCP Vertex AI 等） |
| `BackendSecurityPolicy` | v1alpha1（非推奨）/ v1beta1 | プロバイダ認証。v0.3+ で `targetRefs` 方式に変更（旧 `AIServiceBackend.backendSecurityPolicyRef` は非推奨） |
| `GatewayConfig` | v1alpha1（非推奨）/ v1beta1 | **v0.5 新規**。Gateway 単位での ExtProc 設定（リソース、env 等） |
| `MCPRoute` | v1alpha1 のみ | **v0.4 新規**。MCP サーバへのルーティング（OAuth、ツールフィルタ等） |
| `QuotaPolicy` | v1alpha1 のみ | 推論サービス向けのトークンクォータ設定 |

**ストレージバージョン移行**: v0.5 で v1beta1 が追加されたが、既存リソースの etcd 上のストレージバージョンは自動移行されない。新規は `v1beta1` を使い、既存は `kubectl apply` で再適用すること。

---

## Gateway CR apply 時のライフサイクル

```
[ユーザー] kubectl apply -f gateway.yaml
    │
    ▼
[Envoy Gateway Controller]
    ├─ Gateway CR を検知
    └─ Envoy Proxy Deployment / Service / ConfigMap を生成
         │
         ▼
[Kubernetes API Server]
    └─ Pod 作成前に MutatingWebhook を呼び出し
         │
         ▼
[AI Gateway MutatingWebhook]
    └─ Pod spec に ExtProc サイドカーを注入
         │
         ▼
[Envoy Proxy Pod 起動]
    ├─ Envoy が xDS で Envoy Gateway Controller から設定取得
    │   （このとき Envoy Gateway は Extension Server:1063 経由で AI Gateway に問い合わせて xDS を加工）
    └─ Envoy と ExtProc が Unix Domain Socket で接続

――― 以降、AIGatewayRoute apply 時 ―――

[ユーザー] kubectl apply -f ai-route.yaml
    │
    ▼
[AI Gateway Controller]
    ├─ AIGatewayRoute を検知
    └─ 内部的に HTTPRoute + HTTPRouteFilter を生成（ai-eg-host-rewrite-*）
         │
         ▼
[Envoy Gateway Controller]
    └─ 上記 HTTPRoute を Envoy 設定に変換する過程で Extension Server を呼び出し
         │
         ▼
[AI Gateway Controller (Extension Server)]
    └─ AI 固有設定（per-backend upstream filter 等）を xDS に追加して返す
         │
         ▼
[Envoy Proxy] 最終 xDS を受領して配信開始
```

---

## 運用上のハマりどころ

### ① Envoy Proxy Pod が存在しないと焦る

Helm インストール直後に `kubectl get pods -n envoy-gateway-system` すると Envoy Gateway Controller Pod しかいない。正常動作。`Gateway` CR を apply して初めてデータプレーン Pod が起動する。

### ② Webhook 証明書の失効で Pod が起動しなくなる

ExtProc 注入は MutatingWebhook 経由のため、AI Gateway の webhook TLS 証明書が壊れると **新規 Envoy Proxy Pod が一切起動できなくなる**（既存 Pod は動き続ける）。

公式チャートには自己署名証明書が埋め込まれているが production 非推奨。`controller.mutatingWebhook.certManager.enable: true` で cert-manager 連携を有効化するのが望ましい。

### ③ ExtProc ログは別 namespace に見えるがそこにない

AI Gateway Controller は `envoy-ai-gateway-system` namespace にあるが、ExtProc サイドカーは Envoy Proxy Pod（`envoy-gateway-system`）に同居している。

```bash
# Envoy Proxy Pod を探す
kubectl get pods -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=<gateway-name>

# サイドカーのログを取得（コンテナ名は kubectl describe で要確認）
kubectl describe pod -n envoy-gateway-system <envoy-proxy-pod>
kubectl logs -n envoy-gateway-system <envoy-proxy-pod> -c <extproc-container-name>
```

### ④ ClientTrafficPolicy のバッファ上限

デフォルトの 32KB では AI のレスポンス（大きな出力や画像入力）に不十分。50MB 程度への引き上げが公式サンプルの推奨値。

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: ai-buffer
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: ai-gw
  connection:
    bufferLimit: 50Mi
```

### ⑤ `v0.0.0-latest` タグは production 非推奨

公式ドキュメントは `--version v0.0.0-latest` を例示しているが、このタグは継続的に上書きされるため予期せぬ変更を被る。production では必ず `v0.5.0` のような固定バージョンを指定する。

### ⑥ v0.2 以前からのアップグレード

CRD 所有権の移管のため `--take-ownership` フラグが必要：

```bash
helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \
  --version v0.5.0 --namespace envoy-ai-gateway-system --take-ownership
helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \
  --version v0.5.0 --namespace envoy-ai-gateway-system
```

また、v0.1-v0.2 時代の `envoy-gateway-config/redis.yaml` + `config.yaml` を手動適用していた場合は、v0.3+ では不要なので整理すること。

### ⑦ Envoy Gateway Observability 設定のドリフト問題

OTel sink や Prometheus metrics の有効化手順は ConfigMap を直接編集する形で案内されている箇所があるが、Helm values に対応パラメータがないため、ArgoCD / Flux で管理すると Helm upgrade のたびにドリフトが発生する可能性がある。運用方針を事前に決めておく。

---

## 参考リンク（公式一次ソース）

- Envoy AI Gateway Docs (latest): https://aigateway.envoyproxy.io/docs/
- Envoy AI Gateway v0.5 Installation: https://aigateway.envoyproxy.io/docs/0.5/getting-started/installation
- Envoy AI Gateway Compatibility Matrix: https://aigateway.envoyproxy.io/docs/0.5/compatibility
- Envoy AI Gateway Release Notes: https://aigateway.envoyproxy.io/release-notes/
- Envoy Gateway Docs: https://gateway.envoyproxy.io/
- Envoy Gateway Helm values 全リファレンス: https://github.com/envoyproxy/gateway/blob/main/charts/gateway-helm/values.tmpl.yaml
- Envoy AI Gateway Helm values 全リファレンス: https://github.com/envoyproxy/ai-gateway/blob/main/manifests/charts/ai-gateway-helm/values.yaml
- AI Gateway 用 Envoy Gateway values（公式）: https://github.com/envoyproxy/ai-gateway/blob/main/manifests/envoy-gateway-values.yaml