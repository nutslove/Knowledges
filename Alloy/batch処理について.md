# Batch処理について
- `otelcol.processor.batch`は、OpenTelemetry Collectorのprocessorの一種で、受信したログを一定のバッチサイズや一定の時間間隔でまとめて処理するためのコンポーネント。
- defaultではbatch processorは有効になってないので、receiverとexporterの間に明示的にbatch processorを追加する必要がある。
- https://grafana.com/docs/alloy/latest/reference/components/otelcol/otelcol.processor.batch/
- 例  
  ```
  // loki.source.awsfirehose receives logs from AWS Firehose( to alb - alloy ).
  loki.source.awsfirehose "loki_firehose_receiver" {
      http {
          listen_address = "0.0.0.0"
          listen_port = 9999
      }
      forward_to = [
          loki.process.set_timestamp.receiver,
      ]
  }

  loki.process "set_timestamp" {
      stage.match {
          selector = "{logtype=\"lambda\"}"

          stage.json {
              expressions = {
                  time_value = "time",
              }
          }

          stage.timestamp {
              source = "time_value"
              format = "RFC3339"
          }
      }

      forward_to = [otelcol.receiver.loki.otellogs.receiver]
  }

  otelcol.receiver.loki "otellogs" {
    output {
      logs = [otelcol.processor.batch.newrelic.input]
    }
  }

  otelcol.processor.batch "newrelic" {
    timeout         = "5s" // default: 200ms
    send_batch_size = 2000 // default: 2000
    output {
      logs = [otelcol.exporter.otlphttp.newrelic.input]
    }
  }

  otelcol.exporter.otlphttp "newrelic" {
    client {
      endpoint = "https://otlp.nr-data.net:4318"
      headers = {
        "api-key" = sys.env("NEWRELIC_API_KEY"),
      }
    }
    retry_on_failure {
      enabled = true
      max_interval = "60s"
      max_elapsed_time = "10m"
    }
    sending_queue {
      enabled = true
      num_consumers = 40
      queue_size = 4000
    }
  }
  ```


## `loki.write`
- `loki.write`は、Lokiにログを送信するためのコンポーネントであり、デフォルトでbatch処理が組み込まれている。
- https://grafana.com/docs/alloy/latest/reference/components/loki/loki.write/

| Name | Type | Description | Default | Required |
|------|------|-------------|---------|----------|
| `batch_size` | `string` | Maximum batch size of logs to accumulate before sending. | `"1MiB"` | No |
| `batch_wait` | `duration` | Maximum amount of time to wait before sending a batch. | `"1s"` | No |