## Connector
- `receivers`と`exporters`の両方の役割を持ち、ある`receivers`から受け取ったデータを`exporters`から別の`receivers`にデータを流すことができる
- Spanを受け取って特定のSpanに対してMetricsにしたり（[Span Metrics Connector](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/connector/spanmetricsconnector/README.md)）、一つのパイプラインに流れてきたデータをAttributeによって分岐させたり（[Routing Connector](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/connector/routingconnector/README.md)）、特定のログの件数をMetricsにしたり（[Count Connector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/countconnector)）することができる

### Span Metrics Connector
- 参考URL
  - https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/connector/spanmetricsconnector/README.md
  - https://grafana.com/docs/tempo/latest/metrics-generator/span_metrics/
  - https://zenn.dev/k6s4i53rx/articles/2023-advent-calender-otel
- Tempoでも`metrics_generator.processor.span_metrics`を使って特定のdimensionのSpanに対してメトリクスを生成することができる
  - 例  
    ```yaml
    server:
      http_listen_port: 3200

    distributor:
      receivers:
          otlp:
            protocols:
              http:
              grpc:

    compactor:
      compaction:
        block_retention: 744h                # configure total trace retention here

    storage:
      trace:
        backend: s3
        s3:
          endpoint: s3.ap-northeast-1.amazonaws.com
          bucket: {{ s3_bucket_name }}
          forcepathstyle: true
          #set to true if endpoint is https
          insecure: true
        wal:
          path: /tmp/tempo/wal         # where to store the the wal locally
        local:
          path: /tmp/tempo/blocks

    overrides:
      # max_search_bytes_per_trace: 0 // v2.1.1からこの設定はなくなった
      metrics_generator_processors:
        - span-metrics

    metrics_generator:
      ring:
        kvstore:
          store: memberlist
      processor:
        service_graphs:
        span_metrics:
          intrinsic_dimensions:
          dimensions:
            - "db.statement"
      registry:
        external_labels:
          source: tempo
        collection_interval: 15s
        stale_duration: 15m
        max_label_name_length: 1024
        max_label_value_length: 2048
      storage:
        path: /opt/tempo/wal
        wal:
        remote_write:
          - url: {{ remote_write_url }}

    usage_report:
      reporting_enabled: false
    ```

### Count Connector
- 参考URL
  - https://zenn.dev/katzchang/articles/8ef357a35f0496
  - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/countconnector

### Routing Connector
- 参考URL
  - https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/connector/routingconnector/README.md