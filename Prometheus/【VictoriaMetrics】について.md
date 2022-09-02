## Trouble-Shooting
#### maxLabelsPerTimeseriesによるlabel dropについて
- vminsertではデフォルトで１メトリクスに当たり30個までのlabelを受け付けて、それを超えたらlabelをdropする
- 以下2つのうちどれかで対策
  - labelを減らす
  - vminsert実行時のフラグ`-maxLabelsPerTimeseries`をデフォルトの30から増やす  
    https://docs.victoriametrics.com/Cluster-VictoriaMetrics.html