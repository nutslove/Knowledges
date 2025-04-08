# 概要
- https://www.traceloop.com/docs/openllmetry/introduction
  > OpenLLMetry is an open source project that allows you to easily start monitoring and debugging the execution of your LLM app. Tracing is done in a non-intrusive way, built on top of OpenTelemetry. You can choose to export the traces to Traceloop, or to your existing observability stack.
- LLMアプリケーションのmonitoringとdebuggingのためのOpenTelemetryを拡張したLLMアプリ用のOpentelemetryのような位置づけ
  - **otlpでデータを連携するため、otlpに対応しているものならどこにでも送れる**

## サポートされているModelやFramework
- https://www.traceloop.com/docs/openllmetry/tracing/supported

---

# 設定
- 参考URL
  - https://www.traceloop.com/docs/openllmetry/getting-started-python
  - https://www.traceloop.com/docs/openllmetry/tracing/annotations
  - https://www.traceloop.com/docs/openllmetry/integrations/otel-collector
- `traceloop-sdk`をインストール  
  ```
  pip install traceloop-sdk
  ```
- `TRACELOOP_BASE_URL`環境変数にotlp http（4318ポート）のエンドポイントを設定する  
  ```shell
  TRACELOOP_BASE_URL=http://<opentelemetry-collector-hostname>:4318
  ```
- `Traceloop`クラスをimportし、`init`メソッドで初期化  
  ```python
  from traceloop.sdk import Traceloop
  from traceloop.sdk.decorators import workflow as openllmetry_workflow, task as openllmetry_task, agent as openllmetry_agent, tool as openllmetry_tool
  from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
  from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter

  log_exporter = OTLPLogExporter(
      endpoint="https://multi-tenant-loki-gateway.monitoring.svc.cluster.local:8080/otlp"
  )

  metric_exporter = OTLPMetricExporter(
      endpoint="http://10.0.0.10:9090/api/v1/otlp/v1/metrics"
  )

  Traceloop.init(
      disable_batch=True,
      app_name="root_cause_analysis",
      headers={"X-Scope-OrgID": "rca"},
      metrics_exporter=metric_exporter,
      metrics_headers={"THANOS-TENANT": "platform-team"},
      logging_exporter=log_exporter,
      logging_headers={"X-Scope-OrgID": "rca"}
  )
  ```
- **`init`メソッドの引数でTraceのServceNameやHttp Header、otlpのLog, Metricのエンドポイント、それぞれのHeaderなどの設定ができる**
  - `Traceloop`クラスの`init`メソッドのソースコード
    - https://github.com/traceloop/openllmetry/blob/main/packages/traceloop-sdk/traceloop/sdk/__init__.py

## Metric
- `gen_ai_client_` prefixのメトリクス名でメトリクスが生成される
  - 使用Token数(`gen_ai_client_token_usage_*`)、処理時間(`gen_ai_client_operation_duration_seconds_*`)の2つのメトリクスが生成される
- 関連ソースコード
  - https://github.com/traceloop/openllmetry/blob/main/packages/opentelemetry-instrumentation-langchain/opentelemetry/instrumentation/langchain/__init__.py

---

# 注意点
## `ERROR:root:Error initializing LangChain instrumentor: No module named 'langchain_openai'`エラー
- OpenLLMetry内部でデフォルトで`langchain_openai`モジュールを使っているらしく、LLMアプリ側で使ってなくてもpipでインストールしておく必要がある（ないと上記のエラーが出る）

## `ERROR:opentelemetry.exporter.otlp.proto.http.metric_exporter:Failed to export batch code: 404, reason: 404 page not found` エラー
- `TRACELOOP_BASE_URL`環境変数のエンドポイントがMetricに対応していない場合（e.g. Tempo Distributorにダイレクトに送っている場合など）上記のエラーが出る。
- `init`メソッドの`metrics_exporter`引数にMetric用のエンドポイントを別途指定する

# 要確認
- `init`メソッドの`logging_exporter`引数にLokiなどotlpに対応しているエンドポイントを指定して、コード実行時特にエラーも出なかったけどLokiにログは連携されなかった。ログは明示的に連携のための設定が必要？？要確認