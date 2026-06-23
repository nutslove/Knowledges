# Envoy AI Gateway のメトリクス収集

> [!CAUTION]
> 2026/06/19時点の情報。
> - 対象バージョン:
>   - Envoy AI Gateway v0.7.x
>   - Envoy Gateway v1.8.x（Envoy Proxy v1.38.x）
>   - OpenTelemetry GenAI Semantic Conventions準拠
>
> AI Gateway独自メトリクスは比較的新しい領域で、v0.3〜v0.7にかけて急速に拡充されている。今後のバージョンで追加・変更される可能性が高い。

---

## 概要

Envoy AI Gatewayの監視は、**3つのレイヤを別々に押さえる**必要がある。それぞれ収集元・ポート・収集機構が異なる。

| レイヤ | 何を測るか | 提供元 | 公開元 |
|---|---|---|---|
| **AI/LLM固有** | トークン使用量、TTFT、モデル別レイテンシ | ExtProcサイドカー | OTLP push（Prometheus exporterも可） |
| **Envoy Proxy（データプレーン）** | リクエスト数、HTTPステータス、コネクション、クラスタヘルス | Envoy Proxy本体 | `/stats/prometheus` (admin API) |
| **Envoy Gateway / AI Gateway Controller（コントロールプレーン）** | xDS配信状況、リソース処理、Webhookイベント | Goコントローラ | `/metrics` (Prometheus形式) |

> [!NOTE]
> AI/LLM固有メトリクスはExtProcサイドカー（AI Gateway固有コンポーネント）が生成し、OpenTelemetry SDKで送出する。Envoy Proxy本体の `stats` システムとは別経路。
> よって「Envoyの `/stats/prometheus` を見ているだけ」では **トークン使用量やGenAIメトリクスは取れない**。OTLP/Prometheus exporterの設定が別途必要。

---

## 1. AI/LLM固有メトリクス（最重要）

### 1.1 概要

AI Gateway v0.7時点で、**[OpenTelemetry GenAI Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)** に準拠した4つのコアメトリクスを出力する。
収集は **デフォルトで有効**、Prometheus exporter経由でscrape可能。OTLP gRPC pushも設定可能。

#### どのPodの・どのポートから取れるか

**ExtProc サイドカー**（Envoy Proxy Pod に AI Gateway Controller のMutatingWebhookで注入される `ai-gateway-extproc` コンテナ）が出している。Envoy Proxy本体ではない点に注意。

| 項目 | 値 | 備考 |
|---|---|---|
| 出力主体 | ExtProcサイドカー | Envoy Proxy Pod内に同居（同じPodの別コンテナ） |
| ネームスペース | `envoy-gateway-system` | Envoy Proxy Podと同じ |
| Pod selector | `gateway.envoyproxy.io/owning-gateway-name=<Gateway名>` | データプレーンPodと共通 |
| ポート | **1064** | ExtProc admin server (`-adminPort` フラグのデフォルト) |
| パス | `/metrics` | Prometheus形式 |
| 同居機能 | `/health` も同ポートで提供 | ヘルスチェック用 |

> [!NOTE]
> ExtProcの **gRPC ext_proc 通信** は別ポート（`-extProcAddr` のデフォルト `:1063`、UDSも可）で、メトリクスとは無関係。
> Envoy Proxy本体のadmin（`/stats/prometheus`）の **19001** ともポートが違うので混同しないこと。

確認方法:

```bash
# Envoy Proxy Pod（中にExtProcサイドカーが居る）を特定
export ENVOY_POD=$(kubectl get pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=<gateway-name> \
  -o jsonpath='{.items[0].metadata.name}')

# ExtProcの1064 を port-forward（コンテナ指定は不要、Pod共有なので）
kubectl port-forward -n envoy-gateway-system pod/$ENVOY_POD 1064:1064
curl localhost:1064/metrics | grep gen_ai
```

実装ソース: [`cmd/extproc/mainlib/main.go`](https://github.com/envoyproxy/ai-gateway/blob/main/cmd/extproc/mainlib/main.go)（`adminPort` デフォルト 1064 を定義）、[`cmd/extproc/mainlib/admin.go`](https://github.com/envoyproxy/ai-gateway/blob/main/cmd/extproc/mainlib/admin.go)（`mux.Handle("/metrics", promhttp.HandlerFor(...))`）。

対応エンドポイント（v0.7）:
- `/v1/chat/completions`（streaming / non-streaming）
- `/v1/completions`（legacy text completions）
- `/v1/embeddings`
- `/v1/responses`（Azure OpenAI Responses API）
- `/v1/audio/transcriptions`、`/v1/audio/translations`
- `/cohere/v2/rerank`
- `/anthropic/v1/messages`（streaming / non-streaming）

### 1.2 コアメトリクス4本

すべて `Float64Histogram` 型。実装は `internal/metrics/genai.go`。

| メトリクス名 | 単位 | 説明 |
|---|---|---|
| `gen_ai.client.token.usage` | `{token}` | 処理されたトークン数。`gen_ai.token.type` ラベルで input / output / cached_input / cache_creation_input / reasoning を区別 |
| `gen_ai.server.request.duration` | `s` | リクエストヘッダ受信完了 〜 レスポンスボディ送出完了までのレイテンシ（ExtProcフィルタ内で計測） |
| `gen_ai.server.time_to_first_token` | `s` | ヘッダ受信完了 〜 レスポンス1トークン目到着までのレイテンシ（**TTFT**） |
| `gen_ai.server.time_per_output_token` | `s` | 連続するトークン／チャンク間のレイテンシ（**TPOT** = Time Per Output Token） |

> [!IMPORTANT]
> `time_to_first_token` と `time_per_output_token` はストリーミング系エンドポイントでのみ意味を持つ。non-streamingでは記録されない／全体レイテンシに包含される。

### 1.3 共通ラベル（attributes）

全AIメトリクスに自動付与:

| ラベル | 例 | 用途 |
|---|---|---|
| `gen_ai.operation.name` | `chat` / `completion` / `embeddings` / `messages` / `image_generation` / `responses` / `speech` / `transcription` / `translation` / `rerank` | オペレーション種別 |
| `gen_ai.provider.name` | `openai` / `azure.openai` / `aws.bedrock` / `aws.anthropic` / `gcp.vertex_ai` / `gcp.anthropic` / `anthropic` / `cohere` | プロバイダ識別 |
| `gen_ai.original.model` | `gpt-4o` | クライアントが指定した元のモデル名（仮想モデル名でもrawの値） |
| `gen_ai.request.model` | `gpt-4o-2024-08-06` | 実際にバックエンドへ送ったモデル名（モデル名書き換え後） |
| `gen_ai.response.model` | `gpt-4o-2024-08-06` | レスポンスで返ってきたモデル名 |
| `gen_ai.token.type` | `input` / `output` / `cached_input` / `cache_creation_input` / `reasoning` | `token.usage` のみ。**`reasoning` は v0.6で追加された thinking トークン専用** |
| `error.type` | `_OTHER` 等 | エラー時のみ付与 |

### 1.4 カスタムラベル（リクエストヘッダ → attribute）

`controller.metricsRequestHeaderAttributes`（メトリクス専用） / `controller.requestHeaderAttributes`（メトリクス・トレース・アクセスログ共通）で、任意のHTTPヘッダをメトリクスのラベルに昇格できる。

```yaml
# Helm values.yaml
controller:
  # 共通（metrics + spans + access logs）
  requestHeaderAttributes:
    x-team-id: team.id
    x-environment: deployment.environment
  # メトリクス専用（カーディナリティを抑えたい場合）
  metricsRequestHeaderAttributes:
    x-tenant-id: tenant.id
```

> [!WARNING]
> **カーディナリティに注意**。`session.id`（`agent-session-id` ヘッダ）はデフォルトで **メトリクスには付与されない**（トレース・アクセスログには付く）。
> user_id、session_id、request_id等のユニーク値をメトリクスに入れると、Prometheusのシリーズ数が爆発する。テナント／チーム単位のラベルに留めること。

### 1.5 ヒストグラムバケット

`internal/metrics/genai.go` で定義されているdurationヒストグラムのバケット境界（参考。バージョンにより変動）:

```
0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10  (秒)
```

LLMリクエストは通常数百ミリ秒〜数十秒のオーダーなので、p95/p99を見るには上限側のバケット（5s / 10s）が重要。長尺reasoningモデル（Opus 4.7 xhigh等）を扱う場合はバケット拡張を検討（カスタムビューが必要）。

### 1.6 出力形式とエクスポート

- **デフォルト**: ExtProcサイドカーがPrometheus exporterを起動し、scrape可能なエンドポイントを公開
- **OTLP push**: 環境変数 `OTEL_EXPORTER_OTLP_ENDPOINT` を設定するとOTLP gRPC/HTTPでpush可能（SkyWalking OAP、Tempo、Phoenix等の任意のOTel collectorへ送れる）

OTLP pushの設定例（SkyWalkingモニタリングのケース）:

| 環境変数 | 値の例 | 説明 |
|---|---|---|
| `OTEL_SERVICE_NAME` | `my-ai-gateway` | サービス識別子 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://oap:11800` | OTel collectorエンドポイント |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` | プロトコル |
| `OTEL_METRICS_EXPORTER` | `otlp` | メトリクスpush有効化 |
| `OTEL_LOGS_EXPORTER` | `otlp` | アクセスログpush有効化 |
| `OTEL_METRIC_EXPORT_INTERVAL` | `5000` | push間隔（ms） |
| `OTEL_RESOURCE_ATTRIBUTES` | `service.layer=ENVOY_AI_GATEWAY` | リソース属性 |

これらは `GatewayConfig` CRDの `extProcContainer.env` に設定する（v0.6以降）:

```yaml
apiVersion: aigateway.envoyproxy.io/v1beta1
kind: GatewayConfig
metadata:
  name: prod-gateway-config
  namespace: envoy-gateway-system
spec:
  extProcContainer:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://otel-collector.monitoring.svc.cluster.local:4317
      - name: OTEL_METRICS_EXPORTER
        value: otlp
```

`Gateway` CRからは annotation `aigateway.envoyproxy.io/gateway-config: prod-gateway-config` で参照する。

---

## 2. MCP関連メトリクス

v0.6で MCPRoute CRD が安定化、それに伴いMCPサーバ向けのメトリクスも追加された。実装は `internal/metrics/mcp_metrics.go`。

| メトリクス名 | 型 | 単位 | 説明 |
|---|---|---|---|
| `mcp.request.duration` | Histogram | `s` | MCPリクエスト処理時間 |
| `mcp.method.count` | Counter | - | 呼び出されたMCPメソッド総数 |
| `mcp.initialization.duration` | Histogram | `s` | MCP初期化（handshake）所要時間 |
| `mcp.capabilities.negotiated` | Counter | - | ネゴシエートされたcapability数 |
| `mcp.progress.notifications` | Counter | - | 送信されたprogress通知数 |

### ラベル

| ラベル | 値の例 |
|---|---|
| `mcp.method.name` | `tools/list` / `tools/call` / `resources/read` 等のJSON-RPCメソッド名 |
| `status` | `success` / `failed` / `error` |
| `error.type` | `unsupported_protocol_version` / `invalid_json_rpc` / `unsupported_method` 等 |
| `capability.type` | `tools` / `resources` / `prompts` / `sampling` / `roots` / `experimental` / `elicitation` / `completions` / `logging` |
| `capability.side` | `client` / `server` |
| `mcp.backend` | アップストリームMCPバックエンド識別子 |

MCPツールフィルタや認可フィルタ（v0.7で追加）の動作確認には `mcp.method.count{status="failed"}` の監視が有効。

---

## 3. Envoy Proxy（データプレーン）メトリクス

AI Gateway固有ではなく、**Envoy Proxy本体が出力する標準のstats**。AI Gatewayでも同様に取得できる。

### 3.1 公開ポートとパス

Envoy ProxyのAdmin APIから公開される:

- ポート: **19001**（Envoy Gatewayデフォルト）
- パス: `/stats/prometheus`

```bash
# port-forwardで確認
export ENVOY_POD=$(kubectl get pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=<gateway-name> \
  -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n envoy-gateway-system pod/$ENVOY_POD 19001:19001
curl localhost:19001/stats/prometheus
```

### 3.2 押さえるべき主要カテゴリ

| カテゴリ | プレフィックス | 例 | 用途 |
|---|---|---|---|
| **HTTPリクエスト** | `envoy_http_*` | `envoy_http_downstream_rq_total{envoy_response_code_class="5"}` | エラー率、QPS |
| **クラスタ（upstream）** | `envoy_cluster_*` | `envoy_cluster_upstream_rq_time_bucket`、`envoy_cluster_upstream_cx_active` | アップストリーム（OpenAI / Bedrock / Vertex 等）のレイテンシ・接続数 |
| **リスナー** | `envoy_listener_*` | `envoy_listener_downstream_cx_total` | ダウンストリーム接続数 |
| **サーバ全体** | `envoy_server_*` | `envoy_server_memory_allocated`、`envoy_server_uptime` | プロキシ自体の健全性 |
| **ヘルスチェック** | `envoy_cluster_health_check_*` | `envoy_cluster_health_check_failure` | バックエンドのヘルス |
| **TLS** | `envoy_listener_ssl_*` | `envoy_listener_ssl_handshake` | TLSハンドシェイク状況 |

### 3.3 AI Gateway特有の観点で重要なEnvoyメトリクス

LLMバックエンド向けに特に重要なもの:

- `envoy_cluster_upstream_rq_timeout` — タイムアウト発生数（長尺レスポンスのチューニング指標）
- `envoy_cluster_upstream_cx_connect_timeout` — 接続タイムアウト
- `envoy_cluster_upstream_rq_retry` — リトライ発生数
- `envoy_cluster_upstream_rq_pending_active` — キュー詰まり（プロバイダのレート制限ヒット時に増える）
- `envoy_cluster_circuit_breakers_default_rq_open` — サーキットブレーカ作動

### 3.4 設定方法（EnvoyProxy CRD）

Prometheus exporterはデフォルトで有効。明示的に制御する場合:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: ai-gateway-proxy-config
  namespace: envoy-gateway-system
spec:
  telemetry:
    metrics:
      prometheus:
        disable: false   # デフォルト false（有効）
      sinks:
        - type: OpenTelemetry
          openTelemetry:
            host: otel-collector.monitoring.svc.cluster.local
            port: 4317
            protocol: grpc
```

`Gateway` の `spec.infrastructure.parametersRef` で参照する。

---

## 4. コントロールプレーンメトリクス

### 4.1 Envoy Gateway Controller

`envoy-gateway-system` ネームスペースの `envoy-gateway` Pod、ポート **19001** の `/metrics` で公開（Goランタイムメトリクス + 独自メトリクス）。

| カテゴリ | 例 | 用途 |
|---|---|---|
| **Watchable** | `watchable_depth`、`watchable_subscribe_duration_seconds`、`watchable_publish_total`、`watchable_panics_recovered_total` | 内部リソース処理キューの状況。`runner` ラベルで `gateway-api`、`xds-server`、`xds-translator` 等を区別 |
| **Status Updater** | `status_update_total`、`status_update_duration_seconds` | Kubernetesリソースのstatus更新頻度 |
| **xDS Server** | `xds_snapshot_create_total`、`xds_snapshot_update_total`、`xds_stream_duration_seconds`（`nodeID` ラベル付き） | xDS配信状況 |
| **Infrastructure Manager** | `resource_apply_total`、`resource_apply_duration_seconds`、`resource_delete_total` | Envoy Proxy Pod等の生成・削除 |
| **Wasm cache** | `wasm_cache_entries`、`wasm_cache_lookup_total` | Wasmフィルタキャッシュ |
| **Topology Injector** | `topology_injector_webhook_events_total` | ExtProc注入Webhookのイベント |

### 4.2 AI Gateway Controller

`envoy-ai-gateway-system` ネームスペースの `ai-gateway-controller` Podも `/metrics` を公開（Goランタイム + controller-runtime標準）。
独自メトリクスはまだ少なく、主にcontroller-runtime由来:

- `controller_runtime_reconcile_total{controller="aigatewayroute",result="success|error|requeue"}`
- `controller_runtime_reconcile_time_seconds_bucket`
- `workqueue_depth{name="..."}`、`workqueue_adds_total`
- `rest_client_requests_total`（K8s API呼び出し）

### 4.3 Prometheus / OpenTelemetry sink設定

`EnvoyGateway` ConfigMapで制御:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
    telemetry:
      metrics:
        prometheus:
          disable: false
        sinks:
          - type: OpenTelemetry
            openTelemetry:
              host: otel-collector.monitoring.svc.cluster.local
              port: 4317
              protocol: grpc
```

変更後は `kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system` が必要。

---

## 5. レートリミット関連メトリクス

`QuotaPolicy`（v0.6でCRD追加、v0.7でenforcement開始）や従来の `BackendTrafficPolicy` のレートリミット動作には、Envoy Gateway同梱の **rate-limit service** が使われる。
rate-limit Pod（`envoy-gateway-system` ネームスペース）のポート **19001** の `/metrics` で公開。

| メトリクス | 説明 |
|---|---|
| `ratelimit_service_rate_limit_total_hits` | 評価された全リクエスト数（allowed + denied） |
| `ratelimit_service_rate_limit_over_limit` | 制限超過で拒否されたリクエスト |
| `ratelimit_service_rate_limit_within_limit` | 制限内で許可されたリクエスト |
| `ratelimit_service_rate_limit_near_limit` | 制限に近づいているが許可されたリクエスト |
| `ratelimit_service_rate_limit_shadow_mode` | dry-runモードで評価されたリクエスト |

ラベル: `domain`（rate-limit domain名）、`key1` / `key2`（descriptorの最初の2キー）。

> [!CAUTION]
> descriptor値にクライアントIPやuser_idのようなhigh cardinality値を入れると、Prometheusのシリーズ数が爆発する。トークンベースのレート制限では `LLMRequestCost` で生成するbucket keyの設計に注意。

AI Gateway固有の使い方として、**トークンベースクォータ**（`LLMRequestCost.type: TotalToken / InputToken / OutputToken / ReasoningToken / CEL`）の効き具合は、これらのメトリクスとAIメトリクス（`gen_ai.client.token.usage`）を**組み合わせて監視**する必要がある。

---

## 6. 監視構成の推奨パターン

### 6.1 最小構成（Prometheus pull）

すべて `/metrics` または `/stats/prometheus` を **Prometheus / VictoriaMetrics でscrape** する構成。
追加コンポーネント不要で立ち上げが速い。

scrape対象一覧:

| # | 対象 | コンテナ | ネームスペース | ポート | パス |
|---|---|---|---|---|---|
| 1 | Envoy Proxy（データプレーン） | `envoy`（main） | `envoy-gateway-system` | 19001 | `/stats/prometheus` |
| 2 | **ExtProcサイドカー（AI/LLMメトリクス）** | `ai-gateway-extproc` | **同上Pod** | **1064**（`-adminPort` デフォルト） | `/metrics` |
| 3 | Envoy Gateway Controller | `envoy-gateway` | `envoy-gateway-system` | 19001 | `/metrics` |
| 4 | AI Gateway Controller | `controller` | `envoy-ai-gateway-system` | 8080（controller-runtimeデフォルト） | `/metrics` |
| 5 | Rate Limit Service | `envoy-ratelimit` | `envoy-gateway-system` | 19001 | `/metrics` |

> [!IMPORTANT]
> #1 と #2 は **同一Pod内の別コンテナ**。Kubernetesでは同一Podがネットワークnamespaceを共有するため、Pod IPに対して19001 と 1064 の両方をscrapeする形になる。Prometheus上ではjobを分けて2回scrapeする構成が素直。

#### 6.1.1 `prometheus.yml` の scrape_configs 例

```yaml
scrape_configs:

  # ─────────────────────────────────────────────
  # 1. Envoy Proxy（データプレーン）
  #    Envoy Proxy Pod のmainコンテナ 19001 → /stats/prometheus
  # ─────────────────────────────────────────────
  - job_name: 'envoy-proxy-dataplane'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [envoy-gateway-system]
    metrics_path: /stats/prometheus
    scrape_interval: 30s
    scrape_timeout: 10s
    relabel_configs:
      # AI Gateway 配下の Envoy Proxy Pod のみ対象
      - source_labels: [__meta_kubernetes_pod_label_gateway_envoyproxy_io_owning_gateway_name]
        action: keep
        regex: .+
      # mainコンテナの 19001 ポートに限定
      - source_labels: [__meta_kubernetes_pod_container_name, __meta_kubernetes_pod_container_port_number]
        action: keep
        regex: 'envoy;19001'
      # 出力ラベル整形
      - source_labels: [__meta_kubernetes_pod_label_gateway_envoyproxy_io_owning_gateway_name]
        target_label: gateway_name
      - source_labels: [__meta_kubernetes_pod_label_gateway_envoyproxy_io_owning_gateway_namespace]
        target_label: gateway_namespace
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
    # カーディナリティ抑制：必要なEnvoyメトリクスファミリだけ残す例
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'envoy_(cluster|http|listener|server)_.*'
        action: keep

  # ─────────────────────────────────────────────
  # 2. ExtProc サイドカー（AI/LLM メトリクス）★最重要
  #    Envoy Proxy Pod の ai-gateway-extproc コンテナ 1064 → /metrics
  # ─────────────────────────────────────────────
  - job_name: 'envoy-ai-gateway-extproc'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [envoy-gateway-system]
    metrics_path: /metrics
    scrape_interval: 30s
    scrape_timeout: 10s
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_gateway_envoyproxy_io_owning_gateway_name]
        action: keep
        regex: .+
      # ExtProc サイドカーコンテナの 1064 ポートに限定
      - source_labels: [__meta_kubernetes_pod_container_name, __meta_kubernetes_pod_container_port_number]
        action: keep
        regex: 'ai-gateway-extproc;1064'
      - source_labels: [__meta_kubernetes_pod_label_gateway_envoyproxy_io_owning_gateway_name]
        target_label: gateway_name
      - source_labels: [__meta_kubernetes_pod_label_gateway_envoyproxy_io_owning_gateway_namespace]
        target_label: gateway_namespace
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod

  # ─────────────────────────────────────────────
  # 3. Envoy Gateway Controller（コントロールプレーン）
  # ─────────────────────────────────────────────
  - job_name: 'envoy-gateway-controller'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [envoy-gateway-system]
    metrics_path: /metrics
    scrape_interval: 30s
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_control_plane]
        action: keep
        regex: envoy-gateway
      - source_labels: [__meta_kubernetes_pod_container_port_number]
        action: keep
        regex: '19001'
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod

  # ─────────────────────────────────────────────
  # 4. AI Gateway Controller（コントロールプレーン）
  # ─────────────────────────────────────────────
  - job_name: 'ai-gateway-controller'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [envoy-ai-gateway-system]
    metrics_path: /metrics
    scrape_interval: 30s
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        action: keep
        regex: ai-gateway-controller
      - source_labels: [__meta_kubernetes_pod_container_port_number]
        action: keep
        regex: '8080'
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod

  # ─────────────────────────────────────────────
  # 5. Rate Limit Service
  # ─────────────────────────────────────────────
  - job_name: 'envoy-ratelimit'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [envoy-gateway-system]
    metrics_path: /metrics
    scrape_interval: 30s
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        action: keep
        regex: envoy-ratelimit
      - source_labels: [__meta_kubernetes_pod_container_port_number]
        action: keep
        regex: '19001'
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
```

> [!NOTE]
> Pod selectorのラベル（`gateway.envoyproxy.io/owning-gateway-name`、`control-plane=envoy-gateway`、`app.kubernetes.io/name=ai-gateway-controller` 等）と containerName（`envoy`、`ai-gateway-extproc`、`envoy-ratelimit`）は、Helmチャートのバージョン・values設定で変わる可能性がある。実環境で `kubectl get pods -n envoy-gateway-system --show-labels` と `kubectl get pod <pod> -o jsonpath='{.spec.containers[*].name}'` を必ず確認すること。

### 6.2 OTLP push構成（推奨：マルチクラスタ / SkyWalking / Phoenix等を使う場合）

- AI/LLM固有メトリクス → ExtProc から OTLP push（OTel collector → Prometheus / SkyWalking OAP / Arize Phoenix）
- Envoy Proxyデータプレーン → `EnvoyProxy.spec.telemetry.metrics.sinks` で OTel collector へ push
- コントロールプレーン → `EnvoyGateway.telemetry.metrics.sinks` で OTel collector へ push

OTel collectorで一元集約 → 任意のバックエンド（Prometheus / Cortex / Mimir / SkyWalking）へ流す形が、運用上もっとも柔軟。

### 6.3 SLI/SLO設計の出発点

LLMトラフィック向けのSLI候補:

| SLI | クエリ例 |
|---|---|
| **可用性** | `1 - sum(rate(envoy_cluster_upstream_rq_total{envoy_response_code_class="5"}[5m])) / sum(rate(envoy_cluster_upstream_rq_total[5m]))` |
| **p95 全体レイテンシ** | `histogram_quantile(0.95, sum by (le, gen_ai_provider_name) (rate(gen_ai_server_request_duration_seconds_bucket[5m])))` |
| **p95 TTFT**（ストリーミングUX） | `histogram_quantile(0.95, sum by (le, gen_ai_request_model) (rate(gen_ai_server_time_to_first_token_seconds_bucket[5m])))` |
| **TPOT**（出力レート） | `histogram_quantile(0.95, sum by (le, gen_ai_request_model) (rate(gen_ai_server_time_per_output_token_seconds_bucket[5m])))` |
| **トークン消費レート** | `sum by (gen_ai_provider_name, gen_ai_request_model, gen_ai_token_type) (rate(gen_ai_client_token_usage_tokens_sum[5m]))` |
| **コスト換算（独自記録ルールで重み付け）** | `sum by (tenant_id) (rate(gen_ai_client_token_usage_tokens_sum{gen_ai_token_type="output"}[5m]) * 0.000015)` 等 |
| **クォータヒット率** | `sum(rate(ratelimit_service_rate_limit_over_limit[5m])) / sum(rate(ratelimit_service_rate_limit_total_hits[5m]))` |

> [!NOTE]
> Prometheusに取り込まれる際、OpenTelemetryメトリクス名のドット `.` はアンダースコア `_` に変換される。
> 例: `gen_ai.client.token.usage` → `gen_ai_client_token_usage`（unit付きで `gen_ai_client_token_usage_tokens` になる場合あり、exporter設定次第）。

### 6.4 アラート設計の出発点

| 観点 | 条件例 |
|---|---|
| プロバイダ5xx率 | `envoy_cluster_upstream_rq_total{envoy_response_code_class="5"}` 比率が5分で >1% |
| プロバイダタイムアウト | `rate(envoy_cluster_upstream_rq_timeout[5m]) > 0` が継続 |
| TTFT劣化 | `gen_ai.server.time_to_first_token` p95が前週同曜日比で2倍以上 |
| サーキットブレーカ作動 | `envoy_cluster_circuit_breakers_default_rq_open > 0` |
| クォータ枯渇直前 | `ratelimit_service_rate_limit_near_limit` 増加 |
| ExtProc Pod異常 | controller-runtime `workqueue_depth` が継続的に増加、`reconcile_total{result="error"}` 増加 |
| トークンコスト急増 | `sum(rate(gen_ai_client_token_usage_tokens_sum{gen_ai_token_type="output"}[1h]))` が予算閾値超過 |

---

## 7. ハマりポイント・運用上の注意

### 7.1 メトリクスが取れない場合のチェック

1. **AI/LLM固有メトリクスが空** → ExtProcサイドカーが対応エンドポイント（`/v1/chat/completions` 等）を実際に処理したか？ `AIGatewayRoute` でルーティングされていないリクエストは計測されない。
2. **`gen_ai.response.model` が空** → ストリーミング途中で切断されたケースや、エラーレスポンスでは欠落する。
3. **TTFT / TPOTがゼロ件** → non-streamingリクエストでは記録されない。ストリーミング有効化（`stream: true`）が必要。
4. **Prometheus exporter形式とOTLP push形式の名前差** → OTLP push経由でPrometheusに入れた場合、サフィックス（`_seconds`、`_total`）の付き方がexporterで異なる。クエリ作成時にメトリクス名を実環境で確認すること。

### 7.2 カーディナリティ管理

- `gen_ai.request.model` / `gen_ai.response.model` は **モデルのマイナーバージョンまで含まれる**（例: `gpt-4o-2024-08-06`）。新モデルが出るたびシリーズが増える。
- 上記4.1の通り `session.id` はデフォルトでメトリクスに付与されない仕様。**カスタムヘッダで誤ってhigh cardinality値を昇格させない**こと。
- マルチテナント環境で `tenant.id` をラベルにする場合、テナント数の上限を見積もる。

### 7.3 v0.5 → v0.7の変化に注意

- v0.5系の **`/docs/0.5/capabilities/metrics/`** に書かれていた構成（`AIGatewayRoute.spec.filterConfig`）は **v0.6で削除済**。ExtProc関連の設定は `GatewayConfig` CRDに移動。
- v0.6で **`LLMRequestCostType.ReasoningToken`** が追加され、`gen_ai.token.type` に `reasoning` が増えた。Claude Opus 4.7 / o1系などreasoningモデル運用時はダッシュボードの分解能を上げる。
- v0.6で**リクエスト/レスポンスbodyのredaction機能**が追加。これを有効にすると、アクセスログ／トレースに本文が出力されない。**メトリクス自体はredactionの影響を受けない**（カウントのみのため）。

---

## 8. 参考リンク

- [AI/LLM Metrics | Envoy AI Gateway 公式ドキュメント](https://aigateway.envoyproxy.io/docs/capabilities/observability/metrics/)
- [GenAI Distributed Tracing | Envoy AI Gateway](https://aigateway.envoyproxy.io/docs/capabilities/observability/tracing/)
- [OpenTelemetry GenAI Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)
- [Gateway Observability | Envoy Gateway](https://gateway.envoyproxy.io/docs/tasks/observability/gateway-observability/)
- [Gateway Exported Metrics | Envoy Gateway](https://gateway.envoyproxy.io/docs/tasks/observability/gateway-exported-metrics/)
- [RateLimit Observability | Envoy Gateway](https://gateway.envoyproxy.io/docs/tasks/observability/rate-limit-observability/)
- [Visualising metrics using Grafana | Envoy Gateway](https://gateway.envoyproxy.io/docs/tasks/observability/grafana-integration/)
- [Monitoring Envoy AI Gateway with Apache SkyWalking](https://skywalking.apache.org/blog/2026-04-02-envoy-ai-gateway-monitoring/)
- [envoyproxy/ai-gateway internal/metrics（実装ソース）](https://github.com/envoyproxy/ai-gateway/tree/main/internal/metrics)
