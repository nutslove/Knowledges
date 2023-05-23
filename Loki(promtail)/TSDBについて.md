- 参考URL
  - https://grafana.com/blog/2022/12/01/grafana-loki-2.7-release/?pg=blog&plcmt=body-txt
  - https://grafana.com/docs/loki/latest/operations/storage/tsdb/?pg=blog&plcmt=body-txt
  - https://lokidex.com/posts/tsdb/
  - https://grafana.com/blog/2023/04/06/grafana-loki-2.8-release-tsdb-ga-logql-enhancements-and-a-third-target-for-scalable-mode/

- TSDBによりindexの大きさが75%減少し、圧縮率も高くなったらしい。  
  また、クエリー時のリソース使用率も下がり、検索速度も以前より４倍は早くなったという。
  - https://grafana.com/blog/2022/12/01/grafana-loki-2.7-release/

- TSDBではIndex Cachingは不要
  - BoltDBスキーマのログの参照期間がすぎたらIndex Caching設定を削除すること
  - https://grafana.com/docs/loki/latest/operations/storage/tsdb/?pg=blog&plcmt=body-txt
    > **Index Caching not required**
TSDB is a compact and optimized format. Loki does not currently use an index cache for TSDB. If you are already using Loki with other index types, it is recommended to keep the index caching until all of your existing data falls out of retention or your configured max_query_lookback under limits_config. After that, we suggest running without an index cache (it isn’t used in TSDB).