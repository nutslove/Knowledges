- Prometheusを本番環境で利用する際に直面する可能性のある問題について  
  https://labs.gree.jp/blog/2017/10/16614/

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
