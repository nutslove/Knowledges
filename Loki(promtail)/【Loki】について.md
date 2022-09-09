## 参考URL
- https://taisho6339.hatenablog.com/entry/2021/05/26/104449
- https://taisho6339.gitbook.io/grafana-loki-deep-dive/
- https://speakerdeck.com/line_developers/grafana-loki-deep-dive

## Architecture
![Loki_Architecture](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Loki_Architecture.jpg)
[^1]出典: https://grafana.com/blog/2021/08/11/a-guide-to-deploying-grafana-loki-and-grafana-tempo-without-kubernetes-on-aws-fargate/

## Components
- __Querier__
  - LogQLで
  - 参照URL
    - https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#querier
    - 

## Observability
- Loki/promtailも自身に関するメトリクスを開示している
  - https://grafana.com/docs/loki/latest/operations/observability/
- 役に立つメトリクス[^2]
  [^2]: 参考URL: https://taisho6339.gitbook.io/grafana-loki-deep-dive/monitoring
  - __Distributor__
    - `loki_distributor_lines_received_total` (counter)  
      → Distributorが受け付けたログ数(per tanant)
    - `loki_distributor_ingester_append_failures_total` (counter)  
      → The total number of failed batch appends sent to ingesters.  
        > **Note**  
        > ingesterへのappendが失敗した場合再送されるのか、このメトリクスの影響を確認！  
  - __Ingester__
    - `loki_ingester_chunks_flushed_total` (counter)  
      → どの要因でflushされたか、以下の`reason`ごとにflushされた件数  
       ・`full` → `chunk_target_size`の条件を満たしてflushされたもの  
       ・`idle` → `chunk_idle_period`の条件を満たしてflushされたもの  
       ・`max_age` → `max_chunk_age`の条件を満たしてflushされたもの  
  - __promtail__

## Configuration
#### ingester
- https://grafana.com/docs/loki/latest/configuration/#ingester
- 以下の3つがingesterからBackend(S3等)にflushされるタイミングに影響する設定  
  → 個の3つの値を大きくしすぎるとメモリ使用量も上がるので要注意
  - `chunk_target_size`
    - chunkがここに設定したsizeに達したらingesterがBackend(S3)にchunkをflushする
  - `max_chunk_age`
    - ここに指定した時間が経過したchunkをflushする
  - `chunk_idle_period`
    - ここに指定した時間の間、chunkに更新がない場合flushする