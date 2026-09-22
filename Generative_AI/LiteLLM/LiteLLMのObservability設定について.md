# LiteLLMのObservability設定について

このリポジトリのdocker-compose構成が出しているテレメトリ(メトリクス・トレース・ログ)の仕組み、設定、注意点をまとめる。

## 全体構成

```
                                    ┌──────────────┐
                                    │  Prometheus  │ <--- scrape :4000/metrics (15s間隔)
                                    └──────────────┘
                                          ▲
┌──────────┐   traces/logs (OTLP/HTTP)   │
│ litellm  │ ───────────────┐            │
└──────────┘                ▼            │
                      ┌──────────────┐   │
                      │ otel-collector│   │
                      └──────────────┘   │
                        │           │    │
                traces  │           │ logs
                        ▼           ▼    │
                    ┌───────┐   ┌──────┐ │
                    │ Tempo │   │ Loki │ │
                    └───────┘   └──────┘ │
                        │           │    │
                        └─────┬─────┴────┘
                              ▼
                        ┌──────────┐
                        │ Grafana  │ (Prometheus / Loki / Tempo datasource 事前設定済み)
                        └──────────┘
```

- **メトリクス**: LiteLLM `prometheus` コールバック → LiteLLMが`/metrics`エンドポイントを公開 → Prometheusがpull(scrape)
- **トレース/ログ**: LiteLLM `otel` コールバック → OTLP/HTTPでotel-collectorにpush → collectorがtracesをTempoへ、logsをLokiへ振り分け(push)

## 関連ファイル

| ファイル | 役割 |
|---|---|
| `config.yaml` | LiteLLM本体の設定。`litellm_settings.callbacks`でテレメトリを有効化 |
| `docker-compose.yml` | 全サービス定義 |
| `prometheus/prometheus.yml` | Prometheusのscrape設定 |
| `otel-collector/otel-collector-config.yaml` | OTLP受信→Tempo/Lokiへの振り分け設定 |
| `tempo/tempo-config.yaml` | Tempoのトレース受信・保存設定 |
| `loki/loki-config.yaml` | Lokiのログ受信・保存設定 |
| `grafana/provisioning/datasources/datasources.yml` | Grafanaのデータソース事前登録 |

---

## 1. LiteLLM側の設定 (`config.yaml`)

```yaml
litellm_settings:
  require_auth_for_metrics_endpoint: false   # /metricsを認証なしで公開(Prometheusがscrapeできるように)
  enable_end_user_cost_tracking_prometheus_only: true  # end_userラベル付きのコスト系メトリクスを有効化
  callbacks:
    - prometheus   # メトリクス収集を有効化
    - otel         # トレース/ログ収集を有効化
```

`callbacks`に指定した文字列がそのままテレメトリの「入口」になる。この2つは独立したコールバックで、それぞれ別の環境変数群で制御される。

---

## 2. `prometheus` コールバック(メトリクス)

### 有効化方法
`callbacks: [prometheus]` を指定するだけ。追加の環境変数は不要。LiteLLMが `/metrics` エンドポイント(Prometheus text exposition format)をHTTPで公開し、Prometheus側がpull(scrape)しに行く方式。

> [!CAUTION]
> マルチワーカー構成(`--num_workers`等でuvicornワーカーを複数起動する場合)では `PROMETHEUS_MULTIPROC_DIR` 環境変数が必須。Prometheus Python clientはマルチプロセス環境ではワーカー間でメトリクスをファイル経由で共有する必要があり、このディレクトリが未設定/ワーカー間で共有されていないと `/metrics` の値が欠落・不整合になる。

### 収集できるデータ
`/metrics` から実際に確認できた主なメトリクス(v1.101.0時点、`docker exec litellm-litellm-1` またはホストから `curl http://localhost:4000/metrics` で全量確認可能。**総数は約88種類**):

#### トークン使用量 (Counter)
- `litellm_input_tokens_metric_total` — 入力トークン合計
- `litellm_output_tokens_metric_total` — 出力トークン合計
- `litellm_total_tokens_metric_total` — 入力+出力合計
- `litellm_cached_tokens_metric_total` — キャッシュから供給されたトークン(LiteLLM側キャッシュ)
- `litellm_input_cached_tokens_metric_total` — プロバイダ側プロンプトキャッシュ読み取り(OpenAI `prompt_tokens_details.cached_tokens`、Anthropic `cache_read_input_tokens` 等)
- `litellm_input_cache_creation_tokens_metric_total` — プロバイダ側プロンプトキャッシュ書き込み(Anthropic `cache_creation_input_tokens`)
- `litellm_input_audio_tokens_metric_total` / `litellm_output_audio_tokens_metric_total` — 音声入出力トークン
- `litellm_output_reasoning_tokens_metric_total` — reasoningトークン(`completion_tokens_details.reasoning_tokens`)

#### コスト (Counter)
- `litellm_spend_metric_total` — 総支出額。ラベルに `model`, `api_key_alias`, `team`, `user`, `end_user`, `user_agent` などが付与され、キー/チーム/ユーザー/エンドユーザー単位で集計可能
- `litellm_remaining_api_key_budget_metric` / `litellm_remaining_team_budget_metric` / `litellm_remaining_user_budget_metric` / `litellm_remaining_org_budget_metric` — 各予算の残額(Gauge)
- `litellm_api_key_max_budget_metric` 等 — 設定上の予算上限

#### レイテンシ (Histogram)
- `litellm_llm_api_time_to_first_token_metric` — **TTFT** (Time To First Token)。ストリーミング応答の最初のトークンが返るまでの時間
- `litellm_llm_api_latency_metric` — LLM API呼び出し自体のレイテンシ(プロバイダとの通信時間)
- `litellm_request_total_latency_metric` — プロキシ着信〜応答完了までのEnd-to-Endレイテンシ(認証・pre/post-callフック込み)
- `litellm_overhead_latency_metric` — LiteLLMが追加するオーバーヘッド(ミリ秒)
- `litellm_overhead_with_guardrails_latency_metric` — ガードレール処理込みのオーバーヘッド
- `litellm_deployment_latency_per_output_token` — **TPOT相当**(出力トークンあたりのレイテンシ)。実測確認: v1.101.0では **Histogram**(旧バージョン/一部ドキュメントではGauge表記だが、現行はHistogram)
- `litellm_guardrail_latency_seconds` — ガードレール実行のレイテンシ
- `litellm_request_queue_time_seconds` — リクエスト到着〜処理開始までのキュー待ち時間

#### リクエスト数・成功/失敗 (Counter)
- `litellm_proxy_total_requests_metric_total` — プロキシへの総リクエスト数
- `litellm_proxy_failed_requests_metric_total` — 失敗レスポンス数
- `litellm_deployment_success_responses_total` / `litellm_deployment_failure_responses_total` — デプロイメント(モデル)単位の成功/失敗数
- `litellm_deployment_total_requests_total` — デプロイメント単位の総リクエスト数
- `litellm_deployment_failed_fallbacks_total` / `litellm_deployment_successful_fallbacks_total` — フォールバック発生数
- `litellm_deployment_cooled_down_total` — ロードバランシングによるクールダウン発生回数
- `litellm_deployment_state` — デプロイメントの状態(0=healthy, 1=partial outage, 2=complete outage)

#### その他
- `litellm_cache_hits_metric_total` / `litellm_cache_misses_metric_total` — LiteLLMキャッシュのヒット/ミス
- `litellm_images_generated_metric_total` / `litellm_video_duration_seconds_metric_total` — 画像/動画生成量
- `litellm_mcp_tool_calls_total` / `litellm_mcp_tool_call_spend_metric_total` — MCPツール呼び出し回数・コスト
- `litellm_guardrail_requests_total` / `litellm_guardrail_errors_total` — ガードレール呼び出し数・エラー数
- `litellm_active_users` / `litellm_total_users` / `litellm_teams_count` — ユーザー/チーム数(Gauge)
- `litellm_in_flight_requests` — 現在処理中のリクエスト数

### 注意点

> [!NOTE]
> メトリクスは pull型。LiteLLM自体はPrometheusに何かをpushしない。`/metrics`が常時アクセス可能である必要がある(`prometheus/prometheus.yml`のscrape対象に含める)。
>
> `_created` サフィックスの系列(例: `litellm_spend_metric_created`)はPrometheus Python clientが自動生成するタイムスタンプ系列で、実データではなく無視してよい。
>
> TTFTはストリーミングリクエストでのみ意味のある値になる(非ストリーミングでは応答全体が返るまでの時間とほぼ同義になる)。

> [!CAUTION]
> `require_auth_for_metrics_endpoint: false` にしないと `/metrics` に認証が必要になり、Prometheusのscrapeが失敗する(トークン設定が別途必要)。
>
> ラベルに `hashed_api_key`, `api_key_alias`, `model_id` など高カーディナリティなものが多数含まれる。APIキー数・モデル数が多い環境ではメトリクスのカーディナリティ(時系列数)が急増するので注意。`attributes` (config側 `callback_settings.otel.attributes` 相当)で絞り込み可能な仕組みがある。

---

## 3. `otel` コールバック(トレース・ログ・メトリクス)

`otel`コールバックはOpenTelemetry Protocol (OTLP) を使って **トレース・ログ・メトリクスの3シグナルすべて** を出力できる。ただし **トレース以外はデフォルトOFF**。

### 3.1 トレース (Traces)

- **デフォルトで有効**(`callbacks: [otel]` を入れるだけで動く。追加フラグ不要)
- 1リクエストにつき複数スパンが生成される: `Received Proxy Server Request` (root) → `auth`, `postgres` (get_user_object等), `raw_gen_ai_request` (実際のLLM API呼び出し) など
- `trace_id` / `span_id` がスパンに付与され、Tempoで検索・可視化可能

### 3.2 ログ/イベント (Logs)

- **デフォルトOFF**。有効化するには環境変数が必要:
  ```
  LITELLM_OTEL_INTEGRATION_ENABLE_EVENTS=true
  ```
- 有効化すると `gen_ai.content.prompt` / `gen_ai.content.completion` といったOTel GenAI Semantic Conventionsのイベント(ログ)が、LiteLLM管理の `LoggerProvider` 経由でOTLP Logsとして出力される
- これらは**スパンイベントではなく独立したログレコード**として送信される(トレースのスパンに埋め込まれるわけではない)
- ログには `trace_id` / `span_id` が付与されるため、Grafana上でトレースとログを相関(Tempo↔Loki連携)させられる

### 3.3 メトリクス (Metrics, via OTel)

- **デフォルトOFF**。有効化するには:
  ```
  LITELLM_OTEL_INTEGRATION_ENABLE_METRICS=true
  ```
- 本スタックでは**未使用**(メトリクスは`prometheus`コールバック経由でPull方式のみ使っている)。OTel経由のメトリクスはPush方式(`PeriodicExportingMetricReader`、デフォルト5秒間隔)になる点が`prometheus`コールバックと異なる

> [!NOTE]
> `prometheus`と`otel`のメトリクス出力を両方同時に有効化することも可能だが、二重計測・二重コストになるため通常はどちらか一方で十分。

### 3.4 プロンプト/レスポンス本文のキャプチャ制御

トレース・ログにLLMへの実際の入出力内容(プロンプト/コンプリーション本文)を含めるかどうかは別軸で制御される:

- `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT=true/false` — OTel GenAI標準の環境変数。本文をスパン属性/ログイベントに含めるかを制御
- `litellm.turn_off_message_logging=True` (config.yaml側の `litellm_settings`) — **キルスイッチ**。これがTrueだと上記設定に関わらず本文キャプチャが一切行われない

> [!NOTE]
> `otel`コールバックの実装には新旧2系統があり、`LITELLM_OTEL_V2=true` で切り替える。
> - **v1(デフォルト)**: `raw_gen_ai_request` スパンに**デフォルトで本文を含む**(Capture by default)
> - **v2**(`LITELLM_OTEL_V2=true`): **デフォルトで本文を含まない**(Safe by default)。含めたい場合は明示的にオプトインが必要

> [!CAUTION]
> 本構成(`config.yaml`)では`turn_off_message_logging`等を明示的に設定していない=デフォルト挙動(本文が記録される)。機密情報を含むプロンプトを扱う場合は要検討。

**config.yamlでの本文ログ無効化設定**(`litellm.turn_off_message_logging` の具体的な書き方):

```yaml
# config.yaml — 全コールバック共通で本文ログを止める(グローバル設定)
litellm_settings:
  callbacks:
    - otel
  turn_off_message_logging: true
```

`otel`コールバックだけに限定して本文ログを止めたい場合は、`callback_settings`側で個別指定も可能:

```yaml
# config.yaml — otelコールバックだけ本文ログを止める
litellm_settings:
  callbacks:
    - otel

callback_settings:
  otel:
    message_logging: false
```

いずれの設定でも、本文以外のメタデータ(モデル名、トークン数、コストなど)は引き続き記録される。

### 3.4.1 実データ検証: TraceとLokiログの内容重複、および「内部処理」spanの実態

「LiteLLMの出すtraceは、Lokiに転送されているログと同じ内容(ユーザ入力等)と、DB処理・認証処理などの内部処理が主な中身なのでは?」という疑問を、ソースコード解析(デプロイ済み v1.101.0 の実際のパッケージ)とTempo/Lokiの実データ突合の両面から検証した結果をまとめる。

**結論**: 前半(Lokiと同じ内容が重複している)は正しい。後半(内部処理が「主な中身」)は不正確 — DB/認証系spanは件数は多いが1つあたりの情報量は意図的にごく小さく、トレースの情報量の大半を占めるのは実際には (a) Lokiと重複するプロンプト/レスポンス本文と、(b) Lokiには一切出ないコスト/トークン/メタデータの2つ。

#### なぜ重複するのか(コード根拠)

本デプロイ (`LITELLM_OTEL_V2` 未設定 → レガシー実装 `litellm/integrations/opentelemetry.py` の `OpenTelemetry` クラスが稼働) では、コンテンツ取得モードは以下の優先順位で決まる:

```python
def _resolve_capture_mode(self) -> str:
    if litellm.turn_off_message_logging:
        return CAPTURE_MODE_NO_CONTENT          # 1. キルスイッチ(最優先)
    if self._capture_mode_cached is not None:
        return self._capture_mode_cached         # 2. OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT (明示設定時)
    return CAPTURE_MODE_SPAN_AND_EVENT if self.message_logging else CAPTURE_MODE_NO_CONTENT  # 3. レガシーフラグ(デフォルト True)
```

本構成の `config.yaml` は上記いずれも設定していないため、デフォルトの `message_logging=True` が効き、**常に `CAPTURE_MODE_SPAN_AND_EVENT`** になる。さらに `docker-compose.yml` で `LITELLM_OTEL_INTEGRATION_ENABLE_EVENTS: "true"` を明示しているため、span側とログ(event)側の**両方**にプロンプト/レスポンス本文が書き込まれる。

| 出力先 | 生成箇所 | 内容 |
|---|---|---|
| Tempo span属性 (`litellm_request`) | `set_attributes()` 内の `gen_ai.input.messages` / `gen_ai.output.messages` / `gen_ai.system_instructions` | プロンプト全文・レスポンス全文 |
| Tempo span属性 (`raw_gen_ai_request`, 子span) | `set_raw_request_attributes()` — `llm.{provider}.*` 属性としてプロバイダの生リクエスト/レスポンスをほぼそのままダンプ | プロバイダ固有の生JSON |
| Loki ログイベント | `_emit_semantic_logs()` が `gen_ai.content.prompt` / `gen_ai.content.completion` を**別イベント**として `trace_id`/`span_id` 付きで送信 | メッセージごと・choiceごとの本文 |

実際にTempoの1トレース(`/api/traces/{traceID}`)と、同一時間帯のLoki(`{service_name="unknown_service"}` で `/loki/api/v1/query_range`)を突き合わせたところ、**同一の `trace_id`/`span_id` を持つLokiログ**(`gen_ai.content.prompt` ×3, `gen_ai.content.completion` ×2)に、Tempo側 `litellm_request` のspan属性とほぼ同一のツール呼び出し内容・アシスタント応答(日本語テキスト含む)が確認できた。実データでの重複を確認済み。

#### 「内部処理」spanは軽量・非機密設計

`auth`, `postgres`(複数回), `proxy_pre_call`, `self`, `batch_write_to_db` などの内部サービスspan(`ServiceTypes` enum, `litellm/types/services.py`)は、`_start_service_span()` で生成されるが、意図的に情報量を絞ってある:

- **DB系span**: `db_span_attributes()` (`litellm/integrations/otel/model/db_endpoint.py`) は `DATABASE_URL` から得られるホスト/ポート/DB名等の非機密メタのみを付与。認証情報は明示的に除外。
- **`log_db_metrics` デコレータ** (`litellm/proxy/db/log_db_metrics.py`) 配下のDB関数呼び出しspanは `_safe_db_event_metadata()` により `table_name` 以外を意図的に除去。コード内コメント曰く「生のkwargs/argsにはPrismaクライアントやOTel spanなどの生きたオブジェクト、トークン等の機密情報が含まれるため、spanに乗せて良いのはテーブル名のみ」。
- **`auth` span**: `user_api_key_auth.py` の `_return_user_api_key_auth_obj` を見る限り `call_type=route` 程度で、APIキーやリクエストボディは記録されない。

これらのspanは件数は多いが1つあたり属性数は2〜8個程度で、情報量としての「主な中身」には該当しない。

#### Traceだけが持つ情報(Lokiには出ない)

`set_attributes()` はコンテンツキャプチャの可否に関わらず**常に**以下を設定する。これらはLokiのログイベントには含まれない、トレース固有の価値:

- トークン数(input/output/total)、コスト計算結果(`response_cost` 等)
- APIキー・チーム・エンドユーザー・モデルグループ等のメタデータ
- ガードレール実行結果(`_create_guardrail_span` が `guardrail_name`, `guardrail_mode`, `guardrail_response`(JSON), `masked_entity_count` を付与)
- span階層そのもの(DB呼び出し回数、認証にかかった時間などの構造的情報)

#### まとめ

正確な内訳は「Trace = Lokiと重複するプロンプト/レスポンス本文 + Lokiには無いコスト/メタデータ + (情報量としては小さい)内部処理spanの構造」。

> [!NOTE]
> Web調査でも一致する記述を確認: [LiteLLM公式ドキュメント](https://docs.litellm.ai/docs/observability/opentelemetry_integration) は `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT` 未設定時にレガシー `message_logging`(デフォルト `True` → `SPAN_AND_EVENT`)へフォールバックすると明記。[OpenTelemetry v2](https://docs.litellm.ai/docs/observability/opentelemetry_v2) ではv1.81.0以降、非標準の `raw_gen_ai_request` 子spanを作らず親spanに直接属性を乗せる設計に変わりつつある(本環境はレガシーV1のため非該当)。実運用の既知issue([PR #40562](https://github.com/BerriAI/litellm/pull/40562))では `SPAN_ONLY` モードで長い会話になるとメッセージ属性が肥大化しコスト等の重要属性を溢れさせる問題も報告されており、本構成のように `SPAN_AND_EVENT`(デフォルト)で本文全文をspanとログの両方に流す設計は、機密情報を扱う場合は重複コスト・肥大化の観点からも見直す価値がある。

### 3.4.2 ユーザごとのToken利用量・料金確認だけが目的なら otel は不要

「ユーザごとのトークン利用量や料金を確認したいだけ」という目的に対しては、`otel`コールバック(トレース・ログ、およびそのバックエンドであるTempo/Loki)は**不要**。`prometheus`コールバックのメトリクスだけで完結する。

実際に `/metrics` を確認すると、`litellm_spend_metric_total`(コスト)や `litellm_total_tokens_metric_total`(トークン数)などのカウンタには、最初から以下のラベルが付与されている:

```
litellm_spend_metric_total{
  end_user="...", user="...", user_email="...",
  api_key_alias="...", hashed_api_key="...",
  team="...", team_alias="...", org_id="...", org_alias="...",
  model="...", model_id="...", requested_model="...", ...
} <値>
```

つまりユーザ単位・APIキー単位・チーム単位・モデル単位の集計は`prometheus`コールバックのみで正確に取得できる。`otel`が追加で提供する価値は「個々のリクエストの中身」(プロンプト全文、DB/認証処理の内訳、レイテンシの内部構造)であって、集計値としてのトークン数・コストではない。Grafana上でも `sum by (end_user) (increase(litellm_total_tokens_metric_total[1h]))` のようなPromQL(VictoriaMetricsでも同じクエリ言語)で完結する。

**目的をユーザ別利用量・料金確認だけに絞る場合、止めても支障がない要素**:

| 要素 | 停止可否 |
|---|---|
| `config.yaml` の `callbacks: [otel]` | 削除可(`prometheus`のみ残す) |
| OTel Collectorの `traces`/`logs` パイプライン、`otlp` receiver | 削除可(`metrics`パイプラインのみ残す) |
| Tempo, Loki コンテナ | 削除可 |
| Grafanaの Tempo/Loki データソース | 削除可 |
| `LITELLM_OTEL_INTEGRATION_ENABLE_EVENTS` 環境変数 | 不要に |

> [!NOTE]
> トレース/ログを止めると、リクエスト単位のデバッグ(認証失敗、DB遅延、ガードレール挙動、個別リクエストのレイテンシ内訳)ができなくなる。障害調査・パフォーマンス分析の用途が将来出てくる可能性があるなら、otel自体は残しつつ3.4節の設定で本文キャプチャだけ止める、という折衷案もある。用途がユーザ別利用量・料金の可視化に限定されているなら、上記の縮小構成(`prometheus`コールバック + OTel Collectorの`metrics`パイプラインのみ + VictoriaMetrics + Grafana)で十分。

### 3.5 重要な注意点: エンドポイント環境変数の非互換性

> [!CAUTION]
> これが本スタック構築時に最もハマった箇所。
>
> OpenTelemetry公式SDKは、シグナルごとに別々のエンドポイントを指定できる標準環境変数を持っている:
> - `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`
> - `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`
> - `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`
> - (各シグナル別の `_PROTOCOL` も同様に存在)
>
> **しかしLiteLLMの `opentelemetry.py` 実装(v1.101.0時点)はこれらのper-signal環境変数を一切読まない。** 実際にコードが参照しているのは:
>
> ```python
> exporter  = os.getenv("OTEL_EXPORTER_OTLP_PROTOCOL", os.getenv("OTEL_EXPORTER", "console"))
> endpoint  = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", os.getenv("OTEL_ENDPOINT"))
> headers   = os.getenv("OTEL_EXPORTER_OTLP_HEADERS", os.getenv("OTEL_HEADERS"))
> ```
>
> つまり**単一の `OTEL_EXPORTER_OTLP_ENDPOINT` を traces/logs/metrics 全シグナルで共用**する設計になっている。`_normalize_otel_endpoint()`が末尾に`/v1/traces`, `/v1/logs`, `/v1/metrics`を自動付与して同一ホスト上の別パスに送るようにはなっているが、**別ホスト・別ポートの複数バックエンド(例: Tempo用とLoki用で別サービス)に振り分けることはできない**。
>
> `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` / `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` を個別に設定しても**黙って無視され**、`exporter`が未設定なら `console`(標準出力にダンプするだけで外部送信されない)にフォールバックする。エラーも出ないため気づきにくい。

**対処法(本構成で採用)**: LiteLLMとバックエンドの間に **OpenTelemetry Collector** を1台挟み、LiteLLMは単一の `OTEL_EXPORTER_OTLP_ENDPOINT` でCollectorにのみ送信する。Collector側でtraces→Tempo、logs→Lokiにファンアウトする。

```yaml
# docker-compose.yml (litellm環境変数)
OTEL_EXPORTER_OTLP_ENDPOINT: http://otel-collector:4318
OTEL_EXPORTER_OTLP_PROTOCOL: http/protobuf
LITELLM_OTEL_INTEGRATION_ENABLE_EVENTS: "true"
```

### 3.6 `OTEL_EXPORTER_OTLP_PROTOCOL` の指定値

- `http/protobuf` — OTLP/HTTP + Protobuf(本構成で採用。Collector, Tempo, Lokiいずれも対応)
- `http/json` — OTLP/HTTP + JSON
- `grpc` — OTLP/gRPC(別途 `grpcio` のインストールが必要、`litellm[grpc]`)
- 未指定/`console` — 標準出力のみ

> [!CAUTION]
> プロトコル未指定/`console`の場合、外部に一切送信されないため、設定ミスに気づかないままログを眺めて「動いていない」と誤解しやすい最大の落とし穴。

---

## 4. OpenTelemetry Collector (`otel-collector/otel-collector-config.yaml`)

LiteLLMからのOTLP(traces + logs)を一箇所で受け、シグナルごとに異なるバックエンドへ転送するハブ。

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

exporters:
  otlphttp/tempo:
    endpoint: http://tempo:4318
    tls:
      insecure: true
  otlphttp/loki:
    endpoint: http://loki:3100/otlp   # Collectorが末尾に /v1/logs を自動付与 -> http://loki:3100/otlp/v1/logs
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlphttp/tempo]
    logs:
      receivers: [otlp]
      exporters: [otlphttp/loki]
```

### 注意点

> [!NOTE]
> `otlphttp` エクスポーターは `endpoint` に指定したベースURLへ、シグナルごとに `/v1/traces` `/v1/logs` `/v1/metrics` を**自動付与**する。Lokiの場合はOTLP受信パスが `/otlp/v1/logs` なので、`endpoint: http://loki:3100/otlp` と指定することで最終的に正しいパスになる。
>
> `otlphttp` は非推奨エイリアスで、起動ログに `"otlphttp" alias is deprecated; use "otlp_http" instead` という warning が出る(動作に影響はないが将来のCollectorバージョンで削除される可能性あり)。
>
> メトリクス用パイプラインは定義していない(本構成ではPrometheus pull方式のみ使用のため)。

---

## 5. Tempo (`tempo/tempo-config.yaml`)

```yaml
distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
```

### 注意点(ハマりどころ)

> [!CAUTION]
> Tempoのデフォルト設定では、`protocols.otlp.grpc:` / `http:` を値なしで書くと `127.0.0.1` にのみバインドされる。これだとTempoコンテナ自身からしかアクセスできず、他のコンテナ(otel-collector)から接続すると `connection refused` になる。**`endpoint: 0.0.0.0:4317` / `0.0.0.0:4318` を明示的に指定する必要がある。**

- ポート`3200`: TempoのHTTP API(検索・クエリ用、Grafanaのdatasourceが使う)
- ポート`4317`: OTLP gRPC受信
- ポート`4318`: OTLP HTTP受信
- `storage.trace.backend: local` — ローカルファイルシステムに保存(本番ではS3/GCS等を推奨)

---

## 6. Loki (`loki/loki-config.yaml`)

```yaml
limits_config:
  allow_structured_metadata: true   # OTLP ingestionに必須
```

### 注意点

> [!CAUTION]
> `allow_structured_metadata: true` が無いとOTLP経由のログ取り込みができない(OTLPのリソース属性/ログ属性はLokiのStructured Metadataとして格納される)。
>
> `schema: v13` + `store: tsdb` の組み合わせがOTLP + structured metadataに対応した推奨スキーマ(v11/v12等の旧スキーマではstructured metadata非対応)。

> [!NOTE]
> OTLP受信パスは `/otlp/v1/logs`(SDK/Collectorから直接送る場合のパス)。Loki標準のpush API(`/loki/api/v1/push`)とは別物。
>
> Lokiはデフォルトで `[::]` (全アドレス)にバインドするため、Tempoのような明示的なbindアドレス指定は不要だった。

---

## 7. Prometheus (`prometheus/prometheus.yml`)

```yaml
scrape_configs:
  - job_name: litellm
    static_configs:
      - targets: ["litellm:4000"]
```

- 15秒間隔でLiteLLMの `/metrics` をscrape
- 認証なし(`require_auth_for_metrics_endpoint: false` と対応)

---

## 8. Grafana (`grafana/provisioning/datasources/datasources.yml`)

Prometheus / Loki / Tempo の3データソースを起動時に自動プロビジョニング。

```yaml
datasources:
  - name: Prometheus
    uid: prometheus
    type: prometheus
    url: http://prometheus:9090
    isDefault: true
  - name: Loki
    uid: loki
    type: loki
    url: http://loki:3100
  - name: Tempo
    uid: tempo
    type: tempo
    url: http://tempo:3200
```

- Grafana v12+ には **Logs Drilldown** / **Traces Drilldown** アプリが標準搭載されており、Loki/Tempoを使う本構成ではプラグイン追加なしでそのまま使える(VictoriaLogs/VictoriaTracesを使う場合は非公式プラグインや回避策が必要になるため、本構成ではLoki/Tempoを採用した)
- Loki⇔Tempoの相互リンク(ログからトレースへジャンプ等、`derivedFields` / `tracesToLogsV2`)は本構成では未設定(シンプルさを優先)。必要なら `jsonData` に追加可能

> [!CAUTION]
> `GF_AUTH_ANONYMOUS_ENABLED: "true"` + `GF_AUTH_ANONYMOUS_ORG_ROLE: Admin` によりログイン不要でAdmin権限アクセス可能な設定にしている。**ローカル開発専用設定**であり、外部公開する場合は必ず認証を有効化すること。
>
> **既知の罠**: `grafana_data` ボリュームに古いデータソース定義の内部DB状態が残っていると、`datasources.yml`を書き換えても起動時に `Datasource provisioning error: data source not found` で起動failするケースがある。データソース構成を大きく変更した場合は `docker volume rm <project>_grafana_data` してボリュームごと作り直すのが確実(ダッシュボード等の手動設定は失われる点に注意)。

### 8.1 ダッシュボードのプロビジョニング (`grafana/provisioning/dashboards/`)

データソースと同様に、ダッシュボードJSONも起動時に自動プロビジョニングできる。

```
grafana/provisioning/dashboards/
├── dashboards.yml              # プロバイダ定義(このディレクトリ配下のJSONを自動読み込み)
└── litellm-user-usage.json     # ユーザ別利用状況ダッシュボード本体
```

```yaml
# dashboards.yml
apiVersion: 1

providers:
  - name: litellm
    orgId: 1
    folder: LiteLLM          # Grafana上でこのフォルダ配下に表示される
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30  # このJSONファイルの変更を定期的に再読込
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
```

`grafana/provisioning` はdocker-compose.ymlで既にまるごとボリュームマウント済みなので、`dashboards/`ディレクトリを追加しても`docker-compose.yml`側の変更は不要。反映させるには他のprovisioning変更と同様に `docker-compose up -d --force-recreate grafana` でコンテナを再作成する(ボリュームマウントだけなので `docker-compose restart grafana` でも通常は反映されるが、確実性は`force-recreate`の方が高い)。

**`litellm-user-usage.json` (「LiteLLM - ユーザ別利用状況」ダッシュボード) の内容**:

- `end_user` ラベルを選択するテンプレート変数(複数選択・全選択対応)でユーザを絞り込み可能
- サマリ: 選択期間内の総コスト(USD)・総トークン数・総リクエスト数・対象ユーザ数(stat パネル)
- **プロンプトキャッシュ効率**: キャッシュヒット率(全体、Input Tokens基準)、ユーザ別のキャッシュヒット率推移
- 時系列: ユーザ別のコスト推移・トークン使用量推移(積み上げエリア)
- **モデル別コスト内訳**: モデル別コスト割合(円グラフ)、モデル別コスト推移(積み上げ時系列)
- テーブル: ユーザ別コストランキング、ユーザ別トークン内訳(input/output/total)、チーム別コストランキング
- **ガードレール実行状況**: ガードレール呼び出し数・エラー数(選択期間内の合計)、ガードレール別内訳(呼び出し数/エラー数)テーブル

いずれも `prometheus`コールバックのメトリクスのみを参照しており、Tempo/Lokiには依存しない(→ [3.4.2](#342-ユーザごとのtoken利用量料金確認だけが目的なら-otel-は不要)で述べた縮小構成でもそのまま使える)。

**テンプレート変数 `end_user`**:

```
label_values(litellm_spend_metric_total, end_user)
```

全パネルの各クエリは、この変数で選んだユーザに `end_user=~"$end_user"` で絞り込んでいる(全選択時は `.*` にマッチするため実質フィルタなし)。

**実際に設定しているPromQL一覧**(パネルタイトル順):

| パネル | PromQL |
|---|---|
| 総コスト (USD) | `sum(increase(litellm_spend_metric_total{end_user=~"$end_user"}[$__range]))` |
| 総トークン数 | `sum(increase(litellm_tokens_metric_total{end_user=~"$end_user"}[$__range]))` |
| 総リクエスト数 | `sum(increase(litellm_proxy_requests_metric_total{end_user=~"$end_user"}[$__range]))` |
| 対象ユーザ数 | `count(count by (end_user) (litellm_spend_metric_total{end_user=~"$end_user"}))` |
| キャッシュヒット率 (全体) | `100 * sum(increase(litellm_provider_cache_read_input_tokens_metric_total{end_user=~"$end_user"}[$__range])) / sum(increase(litellm_input_tokens_metric_total{end_user=~"$end_user"}[$__range]))` |
| キャッシュヒット率推移 (ユーザ別) | `100 * sum by (end_user) (increase(litellm_provider_cache_read_input_tokens_metric_total{end_user=~"$end_user"}[$__interval])) / sum by (end_user) (increase(litellm_input_tokens_metric_total{end_user=~"$end_user"}[$__interval]))` |
| コスト推移 (ユーザ別) | `sum by (end_user) (increase(litellm_spend_metric_total{end_user=~"$end_user"}[$__interval]))` |
| トークン使用量推移 (ユーザ別) | `sum by (end_user) (increase(litellm_tokens_metric_total{end_user=~"$end_user"}[$__interval]))` |
| モデル別コスト割合 (円グラフ) | `sum by (model) (increase(litellm_spend_metric_total{end_user=~"$end_user"}[$__range]))` |
| モデル別コスト推移 | `sum by (model) (increase(litellm_spend_metric_total{end_user=~"$end_user"}[$__interval]))` |
| ユーザ別コストランキング | `sum by (end_user, team, model) (increase(litellm_spend_metric_total{end_user=~"$end_user"}[$__range]))` |
| ユーザ別トークン内訳 - Input | `sum by (end_user) (increase(litellm_input_tokens_metric_total{end_user=~"$end_user"}[$__range]))` |
| ユーザ別トークン内訳 - Output | `sum by (end_user) (increase(litellm_output_tokens_metric_total{end_user=~"$end_user"}[$__range]))` |
| ユーザ別トークン内訳 - Total | `sum by (end_user) (increase(litellm_tokens_metric_total{end_user=~"$end_user"}[$__range]))` |
| チーム別コストランキング | `sum by (team, team_alias) (increase(litellm_spend_metric_total{end_user=~"$end_user"}[$__range])) > 0` |
| ガードレール呼び出し数 | `sum(increase(litellm_guardrail_requests_total[$__range]))` |
| ガードレールエラー数 | `sum(increase(litellm_guardrail_errors_total[$__range]))` |
| ガードレール別内訳 - Requests | `sum by (guardrail_name, hook_type) (increase(litellm_guardrail_requests_total[$__range]))` |
| ガードレール別内訳 - Errors | `sum by (guardrail_name, hook_type) (increase(litellm_guardrail_errors_total[$__range]))` |

`$__range` は選択中の時間範囲全体(サマリ/ランキング系、`instant`クエリ)、`$__interval` はパネルの描画解像度に応じた自動間隔(時系列の各点の増加量)。ガードレール系のみ `end_user` ラベルを持たないため変数フィルタを適用していない。

**すべての時系列/円グラフパネルに `fieldConfig.defaults.color.mode: "palette-classic"` を明示**しており、`end_user`・`model`・`guardrail_name`などクエリの `by()` に指定したラベルの値ごとに、Grafana標準のパレットから自動的に異なる色が割り当てられる(同じラベル値は複数パネルをまたいでも概ね同じ色になる)。

**キャッシュヒット率について**: LiteLLM自体のレスポンスキャッシュ機構(`litellm_cache_hits_metric` / `litellm_cache_misses_metric_total`)は本構成ではキャッシュ未設定のため常に0(またはmisses一定)。ダッシュボードでは代わりに**プロバイダ側プロンプトキャッシュ**(Anthropicの`cache_read_input_tokens`等、`litellm_provider_cache_read_input_tokens_metric_total`)からの読み取りトークンを、総Inputトークン(`litellm_input_tokens_metric_total`)で割った値を「キャッシュヒット率」として算出している。こちらは実際のコスト削減効果に直結するため、この構成での目的に合う。

**ガードレールメトリクスについて**: `litellm_guardrail_requests_total` / `litellm_guardrail_errors_total` はLiteLLM側のソースコード(`litellm/proxy/utils.py`)を確認する限り、`config.yaml`で設定したガードレール(本構成では`headroom-compression`)がpre_callフックで実行されるたびに`guardrail_name`/`status`/`hook_type`ラベル付きで記録される設計だが、`should_run_guardrail()`の判定でガードレール自体がスキップされた場合は記録されない。本環境では執筆時点でこれらのメトリクスにまだデータが無く(パネルは空で表示される)、ガードレールが実際に実行され次第データが乗ってくる。

> [!CAUTION]
> クエリで参照しているメトリクス名は、LiteLLMが `/metrics` で公開する生の名前(例: `litellm_total_tokens_metric_total`, `litellm_proxy_total_requests_metric_total`)とは**一部異なる**。OTel Collectorの `prometheus` receiverがスクレイプ時にPrometheusのcounter命名規則に合わせて名前を正規化する際、末尾が `_total_total` のように重複するケースを1つの `_total` に短縮するため、VictoriaMetrics上では以下のように名前が変わる:
>
> | LiteLLMが`/metrics`で公開する名前 | VictoriaMetrics(Remote Write後)に格納される名前 |
> |---|---|
> | `litellm_total_tokens_metric_total` | `litellm_tokens_metric_total` |
> | `litellm_proxy_total_requests_metric_total` | `litellm_proxy_requests_metric_total` |
>
> `litellm_spend_metric_total` や `litellm_input_tokens_metric_total` / `litellm_output_tokens_metric_total` のように元々末尾が単一の `_total` であるものは変化しない。ダッシュボードやアラートルールを自作する際は、`curl http://localhost:4000/metrics` (LiteLLM生の名前)ではなく、VictoriaMetrics側 (`curl http://localhost:8428/api/v1/label/__name__/values`) で実際に格納されている名前を確認してから使うこと。

---

## 9. docker-compose運用上の注意点

### `pull_policy: missing` と `latest` タグ

全イメージを**バージョン固定タグ**にしている(`litellm:v1.101.0`, `prometheus:v3.13.3`, `loki:3.7.8`, `tempo:3.0.3`, `grafana:13.2.2`, `opentelemetry-collector-contrib:0.161.0`)。バージョンは定期的に手動で見直し、意図的に上げること。

> [!CAUTION]
> `pull_policy: missing` はローカルにイメージが存在すればpullをスキップする設定だが、**`latest`タグの場合はこの設定があっても常にレジストリに問い合わせに行く**(Docker Composeの既知の仕様)。バージョン固定 + `pull_policy: missing` の組み合わせで初めて「ローカルキャッシュを使い回す」意図通りの挙動になる。

### `docker-compose` vs `docker compose`

> [!NOTE]
> 環境によっては `docker compose`(v2 plugin構文)が使えず、standalone版の `docker-compose` コマンドしか無い場合がある。うまくいかない場合はどちらのコマンド体系が使えるか確認する。

### 設定ファイル変更後の反映

> [!CAUTION]
> `volumes:` でマウントしている設定ファイル(`tempo-config.yaml`等)を書き換えても、`docker-compose up -d <service>` だけではコンテナが再作成されず変更が反映されないことがある。確実に反映させるには対象サービスを `stop` → `rm -f` → `up -d` するか、`docker-compose up -d --force-recreate <service>` を使う。

---

## 10. トラブルシューティングチェックリスト

トレース/ログが届かない場合の確認手順:

1. **LiteLLM側でexporterが`console`にフォールバックしていないか**
   - `docker exec litellm-litellm-1 env | grep OTEL` で `OTEL_EXPORTER_OTLP_ENDPOINT` / `OTEL_EXPORTER_OTLP_PROTOCOL` が想定通り設定されているか確認
2. **Collectorがreceiverでデータを受けているか**
   - `docker logs litellm-otel-collector-1` でエラー・warningを確認
3. **Collector→Tempo/Lokiの接続確認**
   - `connection refused` が出ていればバインドアドレス(`0.0.0.0` vs `127.0.0.1`)を疑う
4. **バックエンド側で実際にデータを受信しているか、API直叩きで確認**
   ```bash
   # Tempo: 直近のトレース検索
   curl -s "http://localhost:3200/api/search?limit=5"
   # Tempoの受信カウンタ(0なら未受信)
   curl -s http://localhost:3200/metrics | grep tempo_distributor_push_duration_seconds_count

   # Loki: ラベル一覧(データがあれば service_name 等が返る)
   curl -s http://localhost:3100/loki/api/v1/labels

   # LiteLLM: Prometheusメトリクスの生存確認
   curl -s http://localhost:4000/metrics | head
   ```
5. **Grafana側でデータソースが正しく登録されているか**
   ```bash
   curl -s http://localhost:3000/api/datasources
   ```
