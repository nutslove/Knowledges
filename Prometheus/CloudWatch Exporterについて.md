https://github.com/prometheus/cloudwatch_exporter

## GetMetricData vs GetMetricStatistics
- CloudWatch Metricsを取得するAPIには`GetMetricData`と`GetMetricStatistics`の2種類がある
- ２つの違いについては以下URL参照  
  https://aws.amazon.com/jp/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/  
  https://github.com/prometheus/cloudwatch_exporter/pull/414
  - 基本的に`GetMetricData`が推奨されている
- CloudWatch Exporterは`GetMetricStatistics`を使っていたけど、`GetMetricData`へ移行されつつある。(2022/08/22時点ではオプションでどっちを使うか選択できる)

## メトリクス取得間隔
- CloudWatch ExporterはPrometheusの`scrape_interval`に指定した間隔で、
  CloudWatchからメトリクスを取得する。  
  ※CloudWatch Exporterの/metricsにアクセス(GET?)するたびにAPIを叩く？  
  https://github.com/prometheus/cloudwatch_exporter/issues/58
    ~~~yaml
    - job_name: 'cloudwatch'
    static_configs:
      - targets: ['cloudwatch:9106']
    scrape_interval: 3m
    scrape_timeout: 1m
    ~~~

## Proxy設定
- OS環境変数ではなく、Java実行時引数として渡す必要がある
  ~~~yaml
        command:
          - java
        args:
          - '-Dhttp.proxyHost=xxx.xxx.xxx.xxx'
          - '-Dhttp.proxyPort=3124'
          - '-Dhttps.proxyHost=xxx.xxx.xxx.xxx'
          - '-Dhttps.proxyPort=3124'
          - '-Dhttp.nonProxyHosts="localhost|127.0.0.1|169.254.169.254"'
          - '-jar'
          - '/cloudwatch_exporter.jar'
          - '9106'
          - '/config/config.yml'
  ~~~

## リソースのTagによる集計
- 直接メトリクスのラベルにリソースTagが反映されたり、
  収集するメトリクスが絞られるわけではない  
  ※下記の例だとUnHealthyHostCountメトリクスは「Project: PODB」Tagを持ってるものだけではなく、すべてのNLBのUnHealthyHostCountメトリクスを収集される
- `aws_tag_select.tag_selections`にて集計したいTagを指定する
  ~~~yaml
  ---
  region: ap-northeast-1
  role_arn: arn:aws:iam::xxxxxxxxx:role/xxxxxxxx
  metrics:
  - aws_namespace: AWS/NetworkELB
    aws_metric_name: UnHealthyHostCount
    aws_dimensions:
      - AvailabilityZone
      - LoadBalancer
      - TargetGroup
    aws_tag_select:
        tag_selections:
            Project: ["PODB"]
        resource_type_selection: "elasticloadbalancing:loadbalancer"
        resource_id_dimension: LoadBalancer
    aws_statistics: [Maximum,Minimum]
    range_seconds: 120
    use_get_metric_data: true
  ~~~
- すると対象Tagを持ってるリソースの`aws_resource_info`というメトリクスが生成される  
  この`aws_resource_info`メトリクスには`aws_tag_select.tag_selections`に指定した
  Tagだけではなく、すべてのTagを持っている。  
  `aws_resource_info`と集計したいメトリクスを`on`/`group_left`でつなげてTagによる集計を実現
  ~~~
  sum by(load_balancer,target_group)(aws_networkelb_un_healthy_host_count_maximum * on(load_balancer) group_left(tag_Project) aws_resource_info)
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