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
- 実行の例  
  ```shell
  export OTEL_EXPORTER_OTLP_INSECURE="true"
  java -javaagent:./opentelemetry-javaagent.jar -Dotel.exporter.otlp.endpoint=http://localhost:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf -Dotel.service.name=petclinic-demo -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=prometheus -jar target/*.jar
  ```