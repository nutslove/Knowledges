- 参考URL
  - https://docs.victoriametrics.com/Cluster-VictoriaMetrics.html#deduplication
  - https://docs.victoriametrics.com/#deduplication

### 概要
- HAのために複数のPrometheusから同じメトリクスがVictoriaMetricsに送信されるケースがあるので、Deduplicationを行う必要がある

### 設定
- `vmselect`と`vmstorage`2ヶ所に設定する必要がある
  - `-dedup.minScrapeInterval=<dedupの基準interval>`flagを付けて起動するだけ
- `-dedup.minScrapeInterval`の値はPrometheus(またはvmagentなどPrometheusの代替)の`scrape_interval`と合わせることが推奨されている