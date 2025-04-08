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

## Log
- defaultではLogの連携は無効化されている
  - TraceとMetricはdefaultで有効になっている
- そのため、`TRACELOOP_LOGGING_ENABLED`環境変数を`"true"`に設定する必要がある
  - 関連ソースコード
    - https://github.com/traceloop/openllmetry/blob/main/packages/traceloop-sdk/traceloop/sdk/__init__.py#L167
    - https://github.com/traceloop/openllmetry/blob/main/packages/traceloop-sdk/traceloop/sdk/config/__init__.py
- そのうえで以下のように`Traceloop.init`時に`logging_exporter`を設定/指定する必要がある  
  ```python
  from traceloop.sdk import Traceloop
  from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter

  os.environ["TRACELOOP_LOGGING_ENABLED"] = "true"
  log_exporter = OTLPLogExporter(
      endpoint="https://10.0.0.0:3100/otlp" # Loki Distributorのotlp endpoint
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
- **環境変数`TRACELOOP_LOGGING_ENABLED="true"`と`Traceloop.init`メソッドで`logging_exporter`を設定したとしてもMetricとTraceのように自動でログを生成してexpoterしてくれるわけではない。`Logger`だけ設定してくれるので、その`Logger`を使って自分でログを生成・送信しばければならない**  
  - 例  
    ```python
    import logging
    from opentelemetry import trace
    from opentelemetry.logs import get_logger_provider, LogRecord, SeverityNumber
    from opentelemetry.trace import get_current_span

    # --- Traceloop.init() 이 이미 호출되어 로깅이 설정되었다고 가정 ---
    # 예:
    # from traceloop.sdk import Traceloop
    # from opentelemetry.sdk._logs.export import ConsoleLogExporter
    # Traceloop.init(app_name="my_langchain_app", logging_exporter=ConsoleLogExporter())
    # -------------------------------------------------------------

    # 1. Logger 객체 가져오기
    # __name__ 대신 로거를 식별할 수 있는 고유한 이름을 사용하는 것이 좋습니다.
    logger = get_logger_provider().get_logger(__name__)

    # 현재 Trace 컨텍스트 가져오기 (선택 사항)
    current_span = get_current_span()
    span_context = current_span.get_span_context()
    trace_id = span_context.trace_id
    span_id = span_context.span_id

    # 2. LogRecord 생성
    log_record = LogRecord(
        timestamp=None, # None으로 설정 시 현재 시간 사용
        observed_timestamp=None,
        trace_id=trace_id, # 현재 트레이스와 연결
        span_id=span_id,   # 현재 스팬과 연결
        trace_flags=span_context.trace_flags,
        severity_text=logging.getLevelName(logging.INFO), # 예: INFO, WARN, ERROR
        severity_number=SeverityNumber.INFO,
        body="사용자 정의 로그 메시지: 특정 작업 완료.",
        attributes={
            "app.specific.key": "some_value",
            "user.id": "user123"
        }
    )

    # 3. 로그 전송 (Emit)
    print(f"로그 전송 시도: {log_record.body}")
    logger.emit(log_record)
    print("로그 전송 완료.")

    # INFO 심각도로 간단하게 로그 생성 (OpenTelemetry 표준 로깅 핸들러가 설정된 경우)
    # 참고: 이 방식은 LogRecord를 직접 만드는 것보다 덜 유연할 수 있습니다.
    #      또한 OpenTelemetry 표준 Python 로깅 통합이 설정되어 있어야 합니다.
    # import logging
    # logging.getLogger(__name__).info("간단한 로그 메시지", extra={
    #     "otel.trace_id": trace.format_trace_id(trace_id),
    #     "otel.span_id": trace.format_span_id(span_id),
    #     "custom.attribute": "value"
    # })
    ```
  - `LoggerProvider`設定関連コード
    - https://github.com/traceloop/openllmetry/blob/main/packages/traceloop-sdk/traceloop/sdk/logging/logging.py
    - https://github.com/traceloop/openllmetry/blob/main/packages/traceloop-sdk/traceloop/sdk/__init__.py
  - MetricとTraceは自動でデータを取得するようになっていることが確認できる
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