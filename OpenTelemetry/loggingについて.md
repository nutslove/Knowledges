- loggingはlokiなど外部ログ保存ツールにログを送るためのものじゃなくてCollectorやOTLPのdebugging用としてCollector(Pod)の標準出力に表示される用途
  - https://aws-otel.github.io/docs/components/misc-exporters#logging-exporter
    > The logging exporter and file exporter are local exporters intended for debugging the Collector or OTLP formatted data without setting up a backend.
  - https://grafana.com/docs/opentelemetry/collector/send-logs-to-loki/