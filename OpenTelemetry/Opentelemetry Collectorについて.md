## Opentelemetry Collectorでログを連携する方法
- https://newrelic.com/blog/how-to-relic/enrich-logs-with-opentelemetry-collector#:~:text=To%20collect%20logs%20from%20files,them%20to%20the%20OpenTelemetry%20Collector.
- 2023/10/28現在`filelog`Receiverと`fluent forward`Receiverの2つがあるっぽい
  - **Filelog Receiver**
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver
  - **Fluent Forward Receiver**
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver
    - **FluentdもしくはFluent Bitからログを受信できる**
  - その他にはLoki Receiverもある（まだAlpha）
    - https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/lokireceiver

## Opentelemetry Collectorの`receivers`と`exporters`、`extensions`について
- デフォルトで使える`receivers`と`exporters`、`extensions`は以下から確認できる
  - https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol/manifest.yaml
- **ocb(OpenTelemetry Collector builder)を使ってデフォルトでは含まれてないreceiver、exporterなどを追加した独自のOpentelemetry Collectorをビルドできる**
  - https://zenn.dev/k6s4i53rx/articles/df59cb65b34ef8
  - https://opentelemetry.io/docs/collector/custom-collector/

## その他
- `<OpenTelemetry_CollectorのIP>:8888/metrics`でOtel-Collector自体のメトリクスを確認できる
  - **Backendに連携されない時は上記のOtel-Collectorのメトリクスを見ること！各データタイプごとの送信成功/失敗件数が公開されている**