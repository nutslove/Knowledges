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

- **アプリが送るプロトコルとcollectorが処理するプロトコルが一致しないときに発生するエラー**
  ```shell
  [otel.javaagent 2025-05-26 06:05:47:451 +0000] [OkHttp http://localhost:4318/...] WARN io.opentelemetry.exporter.internal.grpc.GrpcExporter - Failed to export spans. Server responded with gRPC status code 2. Error message: closed
  ```
  - 原因
    - OpenTelemetry Java SDKは、エンドポイントのプロトコルを自動判別しようとして、ポート4318でも、明示的にプロトコルを指定しないとgRPCとして扱われる場合があるらしい。ADOT Collectorは4318でHTTPを期待しているため、プロトコルの不一致が発生。
  - 解決方法
    - `JAVA_TOOL_OPTIONS`に`-Dotel.exporter.otlp.protocol=http/protobuf`を追加するか、環境変数に`OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`を設定
      - `OTEL_EXPORTER_OTLP_PROTOCOL`には`grpc` (to use OTLP/gRPC)、`http/protobuf` (to use OTLP/HTTP + protobuf)、`http/json` (to use OTLP/HTTP + JSON)を指定できる