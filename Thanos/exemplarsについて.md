- https://thanos.io/tip/components/receive.md/
- OpenMetricsでフォーマットが決まったいる
  - https://github.com/prometheus/OpenMetrics/blob/v1.0.0/specification/OpenMetrics.md#exemplars
- Thanosではデフォルトではexemplarsの受信が無効になっている
- **有効にするためにはReceiver（ingesting-receiver）で`--tsdb.max-exemplars`パラメータを1以上に設定する必要がある**
- ` --tsdb.max-exemplars`は**テナント / Receiver ごと**に保存されるexemplarsの数。**exemplarsの数が` --tsdb.max-exemplars`に達した場合は、最も古いexemplarsが削除されて新しく入ってきたものが保存される。**
- **Sample（Data Point）ごとに１exemplarなので「exemplarsを持つSeries数 x (メトリクス保持期間 ÷ スクレイピング間隔)」の数の分（5m、1hのメトリクスも要考慮？）が必要**
- 現在保存されているexemplarsの数は`prometheus_tsdb_exemplar_exemplars_in_storage`メトリクスから確認できる
- ` --tsdb.max-exemplars=1000`で設定した時の`prometheus_tsdb_exemplar_exemplars_in_storage`メトリクスの例（1kで頭打ちしている）  
![](./image/max-exemplars.jpg)