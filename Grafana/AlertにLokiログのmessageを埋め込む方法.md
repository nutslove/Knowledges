- 参考URL
  - https://github.com/grafana/loki/issues/5844
  - https://github.com/grafana/loki/issues/3507
  - https://github.com/grafana/loki/issues/3045#issuecomment-1178904555

- **`pattern`を使うためLokiしかできない (CloudWatch Logs等はできない) っぽい**
#### 設定例
- 以下の場合、`message`というラベルが追加される
  - `(count_over_time({pod_name="grafana-c5768ff6d-ppfx7"} |="error" | pattern `<message>` [1m]))`
- 上記の設定で、`{{ $values.B.Labels.message }}`でSummaryなどに記述できて、logの中身が連携される
- **ただ、messageごとに異なるアラートとして扱われるので、messageの内容が１文字でも違うと、LogQLにヒットするログの数の分アラートが発砲される**