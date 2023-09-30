- Prometheusを本番環境で利用する際に直面する可能性のある問題について  
  https://labs.gree.jp/blog/2017/10/16614/

## Metric Typeについて
- https://prometheus.io/docs/concepts/metric_types/

### Counter
- 値が増加するだけのType
- `rate`や`increase`関数と組み合わせて使うことが多い
- exporterなどmetricを持つものがrestartするとzeroにリセットされる
- `rate()`と`increase()`関数はmetric(値)のリセットを検出してうまく処理してくれるらしい

### Gauge
- 値がUp/DownするType（e.g. memory利用率や気温）
- snapshotのようなスクレイピングした瞬間の値

### Histogram
- 値（例えば、応答時間）の分布を観察するのに使用される。バケット(bucket)が使われ、各バケットは値の特定の範囲を表す。例えば、0.1,0.2,0.5というバケットを事前に作成しておいて、応答時間が0.1秒以下、0.2秒以下、0.5秒以下などのリクエストの数を計算することができる。
  > **Warning**
  > **事前にバケットの範囲(leの値)を定義する必要がある**
- 各metricごとに以下３つのmetricsが生成される
  1. `<metric名>_bucket{le=<事前に定義したleの値>}`
     - `le`以下の値を持つmetricのカウント
  2. `<metric名>_sum`
     - (すべてのバケットの)metric値の合計値
  3. `<metric名>_count`
     - (すべてのバケットの)metricの数
     - **`<metric名>_bucket{le="+Inf"}`の値と同じ**
       - `{+Inf}`は(Infinity)上限なしを意味し、すべての値のカウントが入る
- `le`は「less than or equal to」の略で以下という意味
- Histogramメトリクスの例
  ~~~
  http_request_duration_seconds_bucket{le="0.05"} 100
  http_request_duration_seconds_bucket{le="0.1"} 150
  http_request_duration_seconds_bucket{le="0.2"} 200
  http_request_duration_seconds_bucket{le="0.5"} 250
  http_request_duration_seconds_bucket{le="1"} 300
  http_request_duration_seconds_bucket{le="+Inf"} 350
  http_request_duration_seconds_count 350
  http_request_duration_seconds_sum 120
  ~~~

### Summary
- 値の分布のquantile(percentile)（e.g. 中央値、90percentileなど）を直接計算する
- 各metricごとに以下３つのmetricsが生成される(`_sum`と`_count`はHistogramと同じ)
  1. `<metric名>{quantile=<0~1の間>}`
     - 各quantileの値
  2. `<metric名>_sum`
     - すべてのmetric値の合計値
  3. `<metric名>_count`
     - すべてのmetricの数
- Summaryメトリクスの例
  ~~~
  http_request_duration_seconds{quantile="0.01"} 0.05
  http_request_duration_seconds{quantile="0.05"} 0.06
  http_request_duration_seconds{quantile="0.5"} 0.09
  http_request_duration_seconds{quantile="0.9"} 1.5
  http_request_duration_seconds{quantile="0.99"} 8.2
  http_request_duration_seconds_count 144320
  http_request_duration_seconds_sum 53423
  ~~~

##### HistogramとSummaryの違い
- Histogramはバケットを使用して観測値の分布を表し、Summaryはquantileを直接計算する。
- Histogramでは事前にバケットの範囲を設定する必要があるが、Summaryでは不要。
- Summaryは通常、サンプリングを使用してquantileを計算するが、Histogramではすべての観測値を使用する。

## 各function(関数)について
#### `rate`
- Time Rangeの間の増加値の1秒ごとの平均値
  > rate(v range-vector) calculates the per-second average rate of increase of the time series in the range vector. 
- Counterタイプのmetricsに対して使う 
#### `increase`
- https://prometheus.io/docs/prometheus/latest/querying/functions/#increase
  > increase(v range-vector) calculates the increase in the time series in the range vector.
- Time Rangeの間の増加値。例えば`increase(http_requests_total)[5m]`はHTTPリクエストの５分間の増加値。
- **increaseは推定値を返すため、実際の増加値が整数でも結果値は整数ではない時がある。**
  > The increase is extrapolated to cover the full time range as specified in the range vector selector, so that it is possible to get a non-integer result even if a counter increases only by integer increments.
  - 例えば10秒間で3が増加したメトリクスがあるとして、`increase(メトリクス[15s])`にした場合、  
      結果値は3ではなく、rangeの15秒を見て10秒間で3だっらから5秒間は1.5増加すると推定し、4.5になる
- **新しく作成されたメトリクスに対してはincreaseは動作しない**
  - 0 → 1に対してはできるけど、無 → 1のメトリクスはincreaseが適用されない
  - https://stackoverflow.com/questions/67985867/why-is-increase-showing-only-zero-values-when-i-can-see-the-metric-value-incre
- Counterタイプのmetricsに対して使う
- 例えばAというCounterタイプのmetricsが1m(60s)間30増えたとする
  - `rate(A[60s])` → 0.5
  - `increase(A[60s])` → 30

  ※ここでいうvectorは1次元リストのこと

#### `offset`
- 過去時間のデータポイントを取得することができる
- `<PromQL> offset <遡る時間単位>`
  - 例）現在のGETメソッドリクエストの合計と**1時間前**のGETメソッドリクエストの合計の差分
    `sum(http_requests_total{method="GET"}) - sum(http_requests_total{method="GET"} offset 1h)`
- https://prometheus.io/docs/prometheus/latest/querying/basics/#offset-modifier

#### `unless`
- PromQLでは同じLabelを複数使うことはできない
- 例えば、nameラベルにある値が必ず設定されてるけど、`_`が入っているものは除外したい場合、以下のような書き方はできない
  - `time() - container_last_seen{pod!~".+",name=~".+",name!~".*_.*"}`
- そこで`unless`を使えば同じLabelに対して特定のデータを除外することができる
  - `(time() - container_last_seen{pod!~".+",name=~".+"} unless {name=~".*_.*"})`

#### `label_replace`によるリラベル
- Prometheus側の設定`relabel_configs`による永続的なリラベルではなく、PromQL`label_replace`で一時的(そのクエリーに限る)にリラベルするとこができる  
  https://stackoverflow.com/questions/71794543/promql-join-on-different-label-names  
  https://prometheus.io/docs/prometheus/latest/querying/functions/#label_replace  
  - 書式
    ~~~
    label_replace(v instant-vector, dst_label string, replacement string, src_label string, regex string)
    ~~~
  - あるラベルの値をそのまま別のラベルとしてリラベルする例
    ~~~
    label_replace(metric, "new_label", "$1", "old_label", "(.*)"
    ~~~

- `on`/`ignoring`と`group_left`/`group_right`を使ってメトリクスの結合ができる  
  https://prometheus.io/docs/prometheus/latest/querying/operators/#vector-matching  
  - 1対1の結合と1対多/多対1の結合がある
  - 1対1の場合は`on`または`ignoring`だけを使う
  - 1対多/多対1の結合は`group_left`または`group_right`も一緒に使う
  - 2つのメトリクスのあるラベルを元に結合する時は`on`を、あるラベルを除外して結合する場合は`ignoring`を使う
    - Example input
      ~~~
      method_code:http_errors:rate5m{method="get", code="500"}  24
      method_code:http_errors:rate5m{method="get", code="404"}  30
      method_code:http_errors:rate5m{method="put", code="501"}  3
      method_code:http_errors:rate5m{method="post", code="500"} 6
      method_code:http_errors:rate5m{method="post", code="404"} 21

      method:http_requests:rate5m{method="get"}  600
      method:http_requests:rate5m{method="del"}  34
      method:http_requests:rate5m{method="post"} 120
      ~~~
    - Example query
      ~~~
      method_code:http_errors:rate5m{code="500"} / ignoring(code) method:http_requests:rate5m
      ~~~
    - Result
      ~~~
      {method="get"}  0.04            //  24 / 600
      {method="post"} 0.05            //   6 / 120
      ~~~
  - 2つのメトリクスのうち、左のメトリクスのcardinalityが高い場合は`group_left`で、右のメトリクスのcardinalityが高い場合は`group_right`で結合
    - Example input
      ~~~
      method_code:http_errors:rate5m{method="get", code="500"}  24
      method_code:http_errors:rate5m{method="get", code="404"}  30
      method_code:http_errors:rate5m{method="put", code="501"}  3
      method_code:http_errors:rate5m{method="post", code="500"} 6
      method_code:http_errors:rate5m{method="post", code="404"} 21

      method:http_requests:rate5m{method="get"}  600
      method:http_requests:rate5m{method="del"}  34
      method:http_requests:rate5m{method="post"} 120
      ~~~
    - Example query
      ~~~
      method_code:http_errors:rate5m / ignoring(code) group_left method:http_requests:rate5m
      ~~~
    - Result
      ~~~
      {method="get", code="500"}  0.04            //  24 / 600
      {method="get", code="404"}  0.05            //  30 / 600
      {method="post", code="500"} 0.05            //   6 / 120
      {method="post", code="404"} 0.175           //  21 / 120
      ~~~

## Prometheus Podを差起動せずConfigをReloadする方法
> A configuration reload is triggered by sending a SIGHUP to the Prometheus process or sending a HTTP POST request to the /-/reload endpoint
- **Prometheus起動時`--web.enable-lifecycle`flagを付ける必要がある**
- 参考URL
  - https://prometheus.io/docs/prometheus/latest/configuration/configuration/
- 修正後のConfigMapをapplyしてから少し間を空けて(1分くらい?)、Reloadすること