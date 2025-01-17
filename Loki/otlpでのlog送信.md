## otlpでのログ送信
- Otel CollectorからotlpでLokiにデータを送る時は、`otlphttp` exporterで、`http://<loki-addr>:3100/otlp` endpointに送ること
  - https://grafana.com/docs/loki/latest/send-data/otel/
  - 設定例  
    ```yaml
    exporters:
      debug:
      otlphttp/loki:
        endpoint: http://<loki-addr>:<port>/otlp
        tls:
          insecure: true
        headers:
          X-Scope-OrgID: otel
    ```

## loki exporter *VS* otlp exporter
- https://grafana.com/docs/loki/latest/send-data/otel/native_otlp_vs_loki_exporter/