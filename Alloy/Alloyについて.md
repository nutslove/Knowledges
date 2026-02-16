- Grafana Labsが出しているOtel Collectorのディストリビューション
- AlloyからKubernetesのPodログを収集できる（Lokiへの連携ももちろん可能）
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.podlogs/

# インストール
- Kubernetes上にHelmでインストール
  - https://grafana.com/docs/alloy/latest/configure/kubernetes/

# AlloyのHealthCheck Endpoint
- `http://<alloy-host>:12345/-/healthy`

# Alloyのユーザ設定（非Root）
- https://grafana.com/docs/alloy/latest/configure/nonroot/
```yaml
alloy:
  securityContext:
    runAsUser: 473
    runAsGroup: 473

configReloader:
  securityContext:
    # this is the UID of the "nobody" user that the configReloader image runs as
    runAsUser: 65534
    runAsGroup: 65534
```

# パイプラインについて
## `loki.process`
- https://grafana.com/docs/alloy/latest/reference/components/loki/loki.process/
- 他のLokiコンポーネント（例：`loki.source.podlogs`）からログを受け取り、stageを使ってログの加工やフィルタリングを行うためのコンポーネント。

## 各種stage
### `stage.timestamp`
- https://grafana.com/docs/alloy/latest/reference/components/loki/loki.process/#stagetimestamp
- ログのタイムスタンプを指定したフィールドから抽出して、ログエントリのタイムスタンプとして使用するためのstage。
- `time`フィールド（RFC3339形式）の値を抽出して、ログエントリのタイムスタンプとして使用する例  
  ```
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
          selector = "{logtype=\"lambda\"}" # `logtype`が`lambda`のログに対してのみtimestampをセットする

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
  ```