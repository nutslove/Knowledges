## Observability
- Loki/promtailも自身に関するメトリクスを開示している
  - https://grafana.com/docs/loki/latest/operations/observability/
- Lokiのmetricsを収集するためのPrometheus設定
  - 以下はLokiと同じcluster内にPrometheusが存在する時の設定例
    ~~~yaml
    scrape_configs:
      - job_name: 'sos-loki'
        kubernetes_sd_configs:
        - role: endpoints
        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_name]
          regex: ^sos-loki-distributed-(distributor|ingester|quer).+
          action: keep
        - source_labels: [__meta_kubernetes_pod_name]
          target_label: pod
        - source_labels: [__meta_kubernetes_pod_ip]
          regex: (.+)
          target_label: __address__
          replacement: ${1}:3100
    ~~~
- 役に立つメトリクス[^1]
  [^1]: https://taisho6339.gitbook.io/grafana-loki-deep-dive/monitoring
  - __Distributor__
    - `loki_distributor_lines_received_total` (counter)  
      → Distributorが受け付けたログ数(per tanant)
    - `loki_distributor_bytes_received_total` (counter)  
      → Distributorが受信した圧縮前のlogのbytes  
      → **Lokiに連携されるlogのsizeを確認する時はこのmetric(すべてのdistributorのsum)から確認できそう**
      > The total number of uncompressed bytes received per both tenant and retention hours.  
      - **例えば1ヶ月(30日)分のログ(データ)量を確認したい場合は`increase(loki_distributor_bytes_received_total{}[30d])`**
    - `loki_distributor_ingester_append_failures_total` (counter)  
      → The total number of failed batch appends sent to ingesters.  
        > **Note**  
        > ingesterへのappendが失敗した場合再送されるのか、このメトリクスの影響を確認！  
        > replication_factorの中で一部失敗したけど過半数は成功したので問題なしとか？
  - __Ingester__
    - `loki_ingester_chunks_flushed_total` (counter)  
      → どの要因でflushされたか、以下の`reason`ごとにflushされた件数  
       ・`full` → `chunk_target_size`の条件を満たしてflushされたもの  
       ・`idle` → `chunk_idle_period`の条件を満たしてflushされたもの  
       ・`max_age` → `max_chunk_age`の条件を満たしてflushされたもの  
  - __promtail__
    - `promtail_sent_entries_total` (counter)  
      → promtailがingesterに送ったログ数
    - `promtail_dropped_entries_total` (counter)  
      → promtailが設定されているすべてのretry回数内にingesterへの送信に失敗した(dropされた)ログ数