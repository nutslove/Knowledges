# 前提
- Exemplarsを持つメトリクスは以下２つ両方から生成される
  1. OpenTelemetry Java Auto Instrumentation agentをから生成されるメトリクス
  2. Spring Boot Actuatorを使っている場合にMicrometerから生成されるメトリクス
- ExemplarsはOpenMetricsの仕様の一部として定義されている

# Java側の設定
## OpenTelemetryの設定
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

> [!NOTE]  
> `OTEL_METRICS_EXPORTER=otlp`に設定して（すると自動的にOTLP ENDPOINTにメトリクスも送信する）、Otel Collectorなどにメトリクスを送信して、そこからPrometheus Remote WriteでPrometheusに送信することもできる

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

## Actuatorの設定
### `pom.xml`（または`build.gradle`）の設定
- `pom.xml`（または`build.gradle`）はビルドツール設定（依存関係・ビルド設定）
- `pom.xml`はMaven、`build.gradle`はGradleの設定ファイル（どちらもJavaのビルドツール）
```xml
・・中略・・
<dependencies>
  <dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-actuator</artifactId>
  </dependency>

  <!-- Micrometer Prometheus for exposing metrics -->
  <dependency>
      <groupId>io.micrometer</groupId>
      <artifactId>micrometer-registry-prometheus</artifactId>
  </dependency>

  <!-- Micrometer Tracing with OpenTelemetry Bridge for exemplars -->
  <dependency>
      <groupId>io.micrometer</groupId>
      <artifactId>micrometer-tracing-bridge-otel</artifactId>
  </dependency>

  <!-- Logback MDC integration -->
  <dependency>
    <groupId>io.opentelemetry.instrumentation</groupId>
    <artifactId>opentelemetry-logback-mdc-1.0</artifactId>
    <version>2.20.1-alpha</version>
    <!-- 「https://mvnrepository.com/artifact/io.opentelemetry.instrumentation/opentelemetry-logback-mdc-1.0」でバージョン確認可能 -->
  </dependency>
</dependencies>
```

> [!NOTE]  
> `io.opentelemetry.opentelemetry-api`は手動計装で使うOpenTelemetry APIライブラリなので、Auto Instrumentationだけで良い場合は不要  
> `io.opentelemetry.opentelemetry-context`も`opentelemetry-api`に含まれているので不要

### `application.properties`（または`application.yml`）の設定
- `application.properties`（または`application.yml`）はアプリ設定（ポート・DB・ログ設定など）
```properties
# OpenTelemetry configuration
otel.service.name=<App(システム)名（e.g. java-spring-boot-service）>
otel.traces.exporter=otlp
otel.metrics.exporter=prometheus
otel.logs.exporter=otlp
otel.exporter.otlp.endpoint=http://<traceのためのバックエンドのIP/Host名>:4317
## 上までの設定は環境変数で設定する場合は不要
otel.propagators=tracecontext,baggage

# Actuator endpoints
management.endpoints.web.exposure.include=health,prometheus # または '*'
management.metrics.export.prometheus.enabled=true
management.metrics.distribution.percentiles-histogram.http.server.requests=true
management.metrics.distribution.percentiles.http.server.requests=0.5, 0.75, 0.9, 0.95, 0.99
management.metrics.tags.<tag名>=<tag値> 
## 必要に応じてタグを追加（複数追加可）
```

### `logback.xml`（または`logback-spring.xml`）の設定（loggingライブラリとしてLogbackを使っている場合）
- LogにTraceIDやSpanID、Flagを含めるための設定
- MDC (Mapped Diagnostic Context) は、ログ出力にコンテキスト情報（TraceIDやSpanIDなど）を追加するための仕組み
- 直接ログをOpenTelemetry(otlp)のバックエンドに送る設定例  
  ```xml
  <?xml version="1.0" encoding="UTF-8"?>
  <configuration>
    <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
      <encoder>
        <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{36} - trace_id=%X{trace_id} span_id=%X{span_id} trace_flags=%X{trace_flags} - %msg%n</pattern>
      </encoder>
    </appender>

    <!-- Just wrap your logging appender, for example ConsoleAppender, with OpenTelemetryAppender -->
    <appender name="OpenTelemetry" class="io.opentelemetry.instrumentation.logback.mdc.v1_0.OpenTelemetryAppender">
      <appender-ref ref="CONSOLE"/>
    </appender>

    <!-- Use the wrapped "OpenTelemetry" appender instead of the original "CONSOLE" one -->
    <root level="INFO">
      <appender-ref ref="OpenTelemetry"/>
    </root>

  </configuration>
  ```
  - `<appender>`の`name`は任意の名前に変更可能
  - OpenTelemetry appenderが`<appender-ref>`でCONSOLEを指定してCONSOLEのアペンダーをラップし、`<root>`で`<appender-ref ref="OpenTelemetry"/>`を指定することで、CONSOLEアペンダーをOpenTelemetryアペンダーでラップして使用するようになる

### 環境変数
- DockerfileやKubernetesのマニフェスト、ECSのTask Definition fileなどで以下の環境変数を設定
```shell
OTEL_SERVICE_NAME=your-service-name
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://<traceのためのバックエンドのIP/Host名>:4317（gRPC）(もしくはhttp://<traceのためのバックエンドのIP/Host名>:4318（HTTP）)
OTEL_EXPORTER_OTLP_LOGS_ENDPOINT=http://<logのためのバックエンドのIP/Host名>:4317（gRPC）(もしくはhttp://<logのためのバックエンドのIP/Host名>:4318（HTTP）)
OTEL_TRACES_EXPORTER=otlp
OTEL_METRICS_EXPORTER=prometheus
OTEL_LOGS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_TRACES_PROTOCOL=<プロトコル>
OTEL_EXPORTER_OTLP_LOGS_PROTOCOL=<プロトコル>
```

> [!TIP]  
> 設定できる環境変数は以下の公式URLで確認可能  
> - https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/
> ### `OTEL_EXPORTER_OTLP_ENDPOINT`
> - metric, trace, logすべてのエクスポート先エンドポイントを指定する
> - 以下のように個別に指定することも可能
> - Default value:
>   - gRPC: "http://localhost:4317"
>   - HTTP: "http://localhost:4318"
>   - Example:
>     - gRPC: export OTEL_EXPORTER_OTLP_ENDPOINT="https://my-api-endpoint:443"
>     - HTTP: export OTEL_EXPORTER_OTLP_ENDPOINT="http://my-api-endpoint/"
> ### `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`
> - traceエクスポート先エンドポイントを指定する
> - Default value:
>   - gRPC: "http://localhost:4317"
>   - HTTP: "http://localhost:4318/v1/traces"
>   - Example:
>     - gRPC: export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="https://my-api-endpoint:443"
>     - HTTP: export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://my-api-endpoint/v1/traces"
> ### `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`
> - metricエクスポート先エンドポイントを指定する
> - Default value:
>   - gRPC: "http://localhost:4317"
>   - HTTP: "http://localhost:4318/v1/metrics"
>   - Example:
>     - gRPC: export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT="https://my-api-endpoint:443"
>     - HTTP: export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT="http://my-api-endpoint/v1/metrics"
> ### `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`
> - logエクスポート先エンドポイントを指定する
> - Default value:
>   - gRPC: "http://localhost:4317"
>   - HTTP: "http://localhost:4318/v1/logs"
>   - Example:
>     - gRPC: export OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="https://my-api-endpoint:443"
>     - HTTP: export OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="http://my-api-endpoint/v1/logs"
> ### `OTEL_EXPORTER_OTLP_PROTOCOL`
> - metric, trace, logすべてのエクスポートに使用するotlpプロトコルを指定する
> - 以下のように個別に指定することも可能
> - Valid values are:
>   - `grpc`: to use OTLP/gRPC
>   - `http/protobuf`: to use OTLP/HTTP + protobuf
>   - `http/json`: to use OTLP/HTTP + JSON
> ### `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`
> - traceエクスポートに使用するotlpプロトコルを指定する
> - Valid values are:
>   - `grpc`: to use OTLP/gRPC
>   - `http/protobuf`: to use OTLP/HTTP + protobuf
>   - `http/json`: to use OTLP/HTTP + JSON
> ### `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL`
> - metricエクスポートに使用するotlpプロトコルを指定する
> - Valid values are:
>   - `grpc`: to use OTLP/gRPC
>   - `http/protobuf`: to use OTLP/HTTP + protobuf
>   - `http/json`: to use OTLP/HTTP + JSON
> ### `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL`
> - logエクスポートに使用するotlpプロトコルを指定する
> - Valid values are:
>   - `grpc`: to use OTLP/gRPC
>   - `http/protobuf`: to use OTLP/HTTP + protobuf
>   - `http/json`: to use OTLP/HTTP + JSON

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

> [!NOTE]  
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
