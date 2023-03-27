## Rate Limit

## Position

## Retry
- Lokiへのログ送信に失敗した際、リトライを行う
- デフォルトでは下記の通り10回リトライを行い、すべて失敗したらログがDropされる
  > Default backoff schedule:
  > 0.5s, 1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s(4.267m)
  > For a total time of 511.5s(8.5m) before logs are lost
- `clients.backoff_config`blockにて最大リトライ数(`max_retries`)等を設定(変更)できる
- 参考URL
  - https://grafana.com/docs/loki/latest/clients/promtail/configuration/#clients
  - https://grafana.com/docs/loki/latest/clients/promtail/troubleshooting/#loki-is-unavailable

## Pipeline
- データの前処理
  - 特定データをDropしたり、ラベルを付与/除外したり、ログに対するメトリクスを生成したりすることができる
- 複数のStageがある
  - https://grafana.com/docs/loki/latest/clients/promtail/stages/

### Stages
##### **metrics**
- ログに対するメトリクスをPrometheus形式で生成し、Promtailから開示する
- `source`で対象のLabelを絞り込まず、すべてのログLine追加に対してMetricを発生させる場合は`config.match_all`を`true`に設定する
- defaultでは5分間Updateされなかったmetricは削除される
  - `max_idle_duration`で変更できる
- 例
  - **以下のように`metrics`ステージの前で`labels`ステージでMetricsにLabelを付与して、最後に`labeldrop`でLabelをdropさせるとMetricsにだけLabelが付与されて、LogにはLabelが付与されない**
  ~~~yaml
  server:
    http_listen_port: 9080
    grpc_listen_port: 0

  positions:
    filename: /tmp/promtail-sos-metrics-positions.yaml

  clients:
    - url: http://loki:3100/loki/api/v1/push
          tenant_id: sos
          backoff_config:
            max_retries: 25
          external_labels:
            env: stg

    scrape_configs:
    - job_name: s3_logs
      loki_push_api:
        server:
          http_listen_port: 3500
          grpc_listen_port: 3600
      pipeline_stages:
      - match:
          selector: '{source="cloudfront"} |~"\t/www/sos/index.html\t"'
          stages:
          - regex:
              expression: "^.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t(?P<http_referer>.+)\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+\t.+"
          - labels:
              http_referer:
          - metrics:
              cloudfront_lines_total:
                type: Counter
                max_idle_duration: 24h
                source: http_referer
                config:
                  action: inc
          - labeldrop:
              - http_referer
      - match:
          selector: '{source="cloudfront"} |~"\t/www/selfplus/assets/images/img_login01.svg|\t/www/selfplus/assets/images/common/logo-povo.svg"'
          stages:
          - metrics:
              cloudfront_selfplus_login:
                type: Counter
                max_idle_duration: 24h
                config:
                  match_all: true
                  action: inc
  ~~~