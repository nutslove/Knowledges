# 前提
- Exemplarsを持つメトリクスは以下２つ両方から生成される
  1. OpenTelemetry Java Auto Instrumentation agentをから生成されるメトリクス
  2. Spring Boot Actuatorを使っている場合にMicrometerから生成されるメトリクス

# Java側の設定
- exemplarsが含まれているOpentelemetryが公開するメトリクスは環境変数`OTEL_METRICS_EXPORTER="prometheus"`もしくはjava実行時のフラグ`-Dotel.metrics.exporter=prometheus`を設定する必要がある
- otelのprometheusメトリクスは **9464** ポートで公開される  
  - otel-collectorでスクレイピングする設定例  
    ```yaml
    receivers:
        otlp:
        protocols:
            grpc:
            endpoint: 0.0.0.0:4317
            http:
            endpoint: 0.0.0.0:4318

        prometheus:
        config:
            scrape_configs:
            - job_name: 'spring-petclinic'
            scrape_interval: 30s
            metrics_path: /actuator/prometheus
            static_configs:
                - targets: ['192.168.1.7:8080']
            - job_name: 'otel'
            scrape_interval: 30s
            metrics_path: /metrics
            static_configs:
                - targets: ['192.168.1.7:9464']
    ```
- その他のフラグについては以下のページで確認
  - **https://opentelemetry.io/docs/languages/java/configuration/**
- otel-collectorへの通信にTLSを使わない場合、環境変数`OTEL_EXPORTER_OTLP_INSECURE="true"`を設定する
  - 要らない気がする（逆にこれの設定をしたらトレースが連携されなくなった）ので、要確認
- 実行の例  
  ```shell
  export OTEL_EXPORTER_OTLP_INSECURE="true"
  java -javaagent:./opentelemetry-javaagent.jar -Dotel.exporter.otlp.endpoint=http://localhost:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf -Dotel.service.name=petclinic-demo -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=prometheus -jar target/*.jar
  ```

## アプリが送るプロトコルとcollectorが処理するプロトコルが一致しないときに発生するエラー
```shell
[otel.javaagent 2025-05-26 06:05:47:451 +0000] [OkHttp http://localhost:4318/...] WARN io.opentelemetry.exporter.internal.grpc.GrpcExporter - Failed to export spans. Server responded with gRPC status code 2. Error message: closed
```
- 原因
  - OpenTelemetry Java SDKは、エンドポイントのプロトコルを自動判別しようとして、ポート4318でも、明示的にプロトコルを指定しないとgRPCとして扱われる場合があるらしい。ADOT Collectorは4318でHTTPを期待しているため、プロトコルの不一致が発生。
- 解決方法
  - `JAVA_TOOL_OPTIONS`に`-Dotel.exporter.otlp.protocol=http/protobuf`を追加するか、環境変数に`OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`を設定
    - `OTEL_EXPORTER_OTLP_PROTOCOL`には`grpc` (to use OTLP/gRPC)、`http/protobuf` (to use OTLP/HTTP + protobuf)、`http/json` (to use OTLP/HTTP + JSON)を指定できる

---

# OpenTelemetry Java Auto Instrumentation agentが生成してくれるexemplarsを持つメトリクス
- Java Auto InstrumentationはPrometheus Exporterを通じてExemplarsを生成してくれる
  - **現状ExemplarsはHistogramsのメトリクスのみサポートしている？**  
    > Currently the OTel Prometheus exporter only supports exemplars in histograms.
    - https://github.com/open-telemetry/opentelemetry-java-instrumentation/discussions/7564
- JavaアプリのIPから`9464`ポートの`/metrics`エンドポイントからスクレーピングできる

## メトリクス一覧（実際に確認済みのメトリクス）
- `http_server_requests_seconds_bucket`
- `http_server_duration_milliseconds_bucket`
- `spring_data_repository_invocations_seconds_bucket`
- `db_client_connections_use_time_milliseconds_bucket`
- `db_client_connections_wait_time_milliseconds_bucket`
- `hikaricp_connections_usage_seconds_bucket`
- `hikaricp_connections_acquire_seconds_bucket`

---

# Java Spring ActuatorのMicrometerが生成してくれるexemplarsを持つメトリクス
- Spring Boot Actuatorを使っている場合、Micrometerが生成するメトリクスにもexemplarsが含まれる
- Javaアプリが起動しているポートの`/actuator/prometheus`からスクレーピングできる

## メトリクス一覧（実際に確認済みのメトリクス）
- `http_server_requests_seconds_bucket`
- `http_server_requests_seconds_count`
- `hikaricp_connections_usage_seconds_count`
- `hikaricp_connections_acquire_seconds_count`
- `logback_events_total`

> [!INFO]  
> ### Spring Boot Actuator
> - Spring Boot Actuatorは、Spring Bootアプリケーションに「運用・監視のための機能」を追加するためのモジュール
> - `/actuator/health`や`/actuator/metrics`などのエンドポイントを提供し、アプリケーションの状態やパフォーマンスに関する情報を取得できる。
> ### Micrometer
> - Micrometerは、Javaアプリケーションのメトリクス収集のためのライブラリ。
> - Spring Boot ActuatorはデフォルトでMicrometerを使用してメトリクスを収集し、Prometheusなどの監視システムにエクスポートできる。

# 注意事項
### **ADOT Java Auto Instrumentation AgentもExemplarsを生成してくれるけど、メトリクス内のTraceIDがW3C形式(ここはX-Ray形式に変換してくれない)なのでX-Rayで検索する時にTraceIDを書き換える必要がある**
- ADOT Java Instrumentationページ
  - https://github.com/aws-observability/aws-otel-java-instrumentation
- 参考URL
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/4616
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation/discussions/4655
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation/discussions/7564
  - https://github.com/open-telemetry/opentelemetry-java/issues/5141

#### **2023/10のアップデートにより、X-RayがW3C形式のTraceIDをサポートするようになった**
- つまり、**w3c形式のTraceをX-Rayにそのまま送れるようになった**
- ADOT Collector version 0.34.0以上のものを使う必要がある
  - ADOT Collectorがデフォルトでw3c形式でexportするか、別の設定が必要かは要確認
- 参考URL
  - https://aws.amazon.com/jp/about-aws/whats-new/2023/10/aws-x-ray-w3c-format-trace-ids-distributed-tracing/
  - https://aws.amazon.com/jp/about-aws/whats-new/2023/10/aws-x-ray-w3c-format-trace-ids-distributed-tracing/
  - https://aws.amazon.com/jp/about-aws/whats-new/2023/10/aws-x-ray-w3c-format-trace-ids-distributed-tracing/
