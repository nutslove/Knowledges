# Opentelemetry Collectorでログを連携する方法
- https://newrelic.com/blog/how-to-relic/enrich-logs-with-opentelemetry-collector#:~:text=To%20collect%20logs%20from%20files,them%20to%20the%20OpenTelemetry%20Collector.
- 2023/10/28現在`filelog`Receiverと`fluent forward`Receiverの2つがあるっぽい
  - **Filelog Receiver**
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver
  - **Fluent Forward Receiver**
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver
    - **FluentdもしくはFluent Bitからログを受信できる**
  - その他にはLoki Receiverもある（まだAlpha）
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/lokireceiver

# otlpのhttpとgrpcによる設定の違い
- **httpの場合は`exporter`の種類が`otlp`ではなく、`otlphttp`！**
  - https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/otlphttpexporter/README.md
- 設定例  
  ```yaml
  exporters:
    debug:
    otlphttp/loki:
      endpoint: http://123.12.12.123:31100
      tls:
        insecure: true
    otlphttp/tempo:
      endpoint: http://123.12.12.123:31090
      tls:
        insecure: true
    prometheusremotewrite: # Thanos
      endpoint: http://123.12.12.123:31700/api/v1/receive
      tls:
        insecure: true
      external_labels:
        system: otel_demo
  ```

# `UseLocalHostAsDefaultHost`について
- Otel Collector v0.104.0から`receiver.otlp`で`endpoint`を指定してない場合、default値が`0.0.0.0`から`localhost`に変更された。  
  Otel CollectorがSidecar方式の場合は問題ないけど、Otel CollectorをGateway方式の場合は以下のように明示的に`endpoint`に`0.0.0.0:<port番号>`と記載する必要がある。
  ```yaml
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
  ```
- 参照URL
  - https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.104.0
  - https://zenn.dev/ryoyoshii/articles/e87bc69d616ee1

# 各種Receiver
## Filelog Receiver
- https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver
- Collectorがあるホスト/コンテナ上のログファイルを収集するReceiver

## Prometheus Receiver
- https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/prometheusreceiver/README.md
- Prometheusと同様にメトリクスをスクレイピングするReceiver
- **exemplars**の受信をサポートする。ただ、OTLPフォーマットに変換する  
  > This receiver accepts exemplars coming in Prometheus format and converts it to OTLP format.
  > 1. Value is expected to be received in `float64` format
  > 2. Timestamp is expected to be received in `ms`
  > 3. Labels with key `span_id` in prometheus exemplars are set as OTLP `span id` and labels with key `trace_id` are set as `trace id`
  > 4. Rest of the labels are copied as it is to OTLP format

# 各種Exporter
## PrometheusRemoteWrite Exporter
#### Amazon Prometheusにメトリクスを送る方法
- https://aws-otel.github.io/docs/getting-started/prometheus-remote-write-exporter
- `sigv4auth`の`extensions`を使う
  - `service`は`"aps"`固定
- 設定例  
  ```yaml
  extensions:
    sigv4auth:
      service: "aps"
      region: "user-region"

  exporters:
    prometheusremotewrite:
      endpoint: "https://aws-managed-prometheus-endpoint/v1/api/remote_write"
      auth:
        authenticator: sigv4auth
  ```

# Opentelemetry Collectorの`receivers`と`exporters`、`extensions`について
- デフォルトで使える`receivers`と`exporters`、`extensions`は以下から確認できる
  - https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol/manifest.yaml
- **ocb(OpenTelemetry Collector builder)を使ってデフォルトでは含まれてないreceiver、exporterなどを追加した独自のOpentelemetry Collectorをビルドできる**
  - https://zenn.dev/k6s4i53rx/articles/df59cb65b34ef8
  - https://opentelemetry.io/docs/collector/custom-collector/
  - **設定ファイル全量**
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/cmd/otelcontribcol/builder-config.yaml

# その他
- `<OpenTelemetry_CollectorのIP>:8888/metrics`でOtel-Collector自体のメトリクスを確認できる
  - **Backendに連携されない時は上記のOtel-Collectorのメトリクスを見ること！各データタイプごとの送信成功/失敗件数が公開されている**