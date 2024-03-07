- Prometheus Go client library
https://pkg.go.dev/github.com/prometheus/client_golang

### `promhttp.Handler()`
- https://github.com/prometheus/client_golang/blob/main/prometheus/promhttp/http.go
- Prometheusのメトリクスをエクスポートするための HTTP ハンドラーを提供する関数

### **`promhttp`Package**
  - *Sub-packages allow to expose the registered metrics via HTTP*
  - [promhttpドキュメント](https://pkg.go.dev/github.com/prometheus/client_golang@v1.14.0/prometheus/promhttp)

### **`MustRegister`関数**
- Metrics have to be registered to be exposed by `MustRegister` func
> While MustRegister is the by far most common way of registering a Collector, sometimes you might want to handle the errors the registration might cause. As suggested by the name, MustRegister panics if an error occurs. With the Register function, the error is returned and can be handled.
- 一般的に`func init()`で登録しておく? 
- Format
  ~~~go
  func MustRegister(cs ...Collector)
  ~~~
  > MustRegister registers the provided Collectors with the DefaultRegisterer and panics if any error occurs.
  - Type `Collector`
    - https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Collector
- 参考URL
  - https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#MustRegister
  - https://pkg.go.dev/github.com/prometheus/client_golang/prometheus@v1.13.0#hdr-Advanced_Uses_of_the_Registry

### **`NewGaugeVec`関数**
> NewGaugeVec creates a new GaugeVec based on the provided GaugeOpts and partitioned by the given label names.
- Format
  ~~~go
  func NewGaugeVec(opts GaugeOpts, labelNames []string) *GaugeVec
  ~~~

### **`GaugeVec`Type**
> GaugeVec is a Collector that bundles a set of Gauges that all share the same Desc, but have different values for their variable labels. This is used if you want to count the same thing partitioned by various dimensions (e.g. number of operations queued, partitioned by user and operation type). Create instances with NewGaugeVec.
- Format
  ~~~go
  type GaugeVec struct {
   *MetricVec
  }
  ~~~
- https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#GaugeVec

### **`GaugeVec.With`**
- `GaugeVec`にラベルと一緒にメトリクスをセットする
> With works as GetMetricWith, but panics where GetMetricWithLabels would have returned an error. Not returning an error allows shortcuts like
- Format
  ~~~go
  func (v *GaugeVec) With(labels Labels) Gauge
  ~~~
- 実際(?)のGaugeタイプのMetricを追加するFormat
  ~~~go
  *GaugeVec.With(prometheus.Labels{"<ラベル名>": <ラベル>,"<ラベル名>": <ラベル>,[・・・]}).Set(メトリクス)
  ~~~
- https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#GaugeVec.With

### **`GaugeOpts`Type**
> GaugeOpts is an alias for Opts. See there for doc comments.
- https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#GaugeOpts

### **`Opts`Type**
> Opts bundles the options for creating most Metric types. Each metric implementation XXX has its own XXXOpts type, but in most cases, it is just an alias of this type (which might change when the requirement arises.)
>
>It is mandatory to set Name to a non-empty string. All other fields are optional and can safely be left at their zero value, although it is strongly encouraged to set a Help string.
- Format
  ~~~go
  type Opts struct {
    // Namespace, Subsystem, and Name are components of the fully-qualified
    // name of the Metric (created by joining these components with
    // "_"). Only Name is mandatory, the others merely help structuring the
    // name. Note that the fully-qualified name of the metric must be a
    // valid Prometheus metric name.
    Namespace string
    Subsystem string
    Name      string

    // Help provides information about this metric.
    //
    // Metrics with the same fully-qualified name must have the same Help
    // string.
    Help string

    // ConstLabels are used to attach fixed labels to this metric. Metrics
    // with the same fully-qualified name must have the same label names in
    // their ConstLabels.
    //
    // ConstLabels are only used rarely. In particular, do not use them to
    // attach the same labels to all your metrics. Those use cases are
    // better covered by target labels set by the scraping Prometheus
    // server, or by one specific metric (e.g. a build_info or a
    // machine_role metric). See also
    // https://prometheus.io/docs/instrumenting/writing_exporters/#target-labels-not-static-scraped-labels
    ConstLabels Labels
  }
  ~~~
- https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Opts

### **`Labels`関数**
- metricにlabelを追加する際に利用
- 事前にmetricが入る変数にsliceとしてkeyだけ作成しといて`<変数名>.With(prometheus.Labels{"<Key>": "<Value>"}).<演算メソッド>`でlabelを追加する
- 例(1)
  ~~~go
  var (
      awsCost = prometheus.NewGaugeVec(prometheus.GaugeOpts{
          Name: "my_aws_cost_for_this_month",
          Help: "My AWS cost for this month.",
      },
          []string{
              "key",
          },
      )
  )

  func init() {
      // Metrics have to be registered to be exposed:
      prometheus.MustRegister(awsCost)
  }

  func main() {
      awsCost.With(prometheus.Labels{"key": "Tax"}).Set(0.05)

      http.Handle("/metrics", promhttp.Handler())
      log.Fatal(http.ListenAndServe(":8080", nil))
  }
  ~~~
- 例(2)
  ~~~go
  var (
      cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
          Name: "cpu_temperature_celsius",
          Help: "Current temperature of the CPU.",
      })
      hdFailures = prometheus.NewCounterVec(
          prometheus.CounterOpts{
              Name: "hd_errors_total",
              Help: "Number of hard-disk errors.",
          },
          []string{"device"},
      )
  )

  func init() {
      // Metrics have to be registered to be exposed:
      prometheus.MustRegister(cpuTemp)
      prometheus.MustRegister(hdFailures)
  }

  func main() {
      cpuTemp.Set(65.3)
      hdFailures.With(prometheus.Labels{"device":"/dev/sda"}).Inc()

      // The Handler function provides a default handler to expose metrics
      // via an HTTP server. "/metrics" is the usual endpoint for that.
      http.Handle("/metrics", promhttp.Handler())
      log.Fatal(http.ListenAndServe(":8080", nil))
  }
  ~~~ 

### **`NewMetricWithTimestamp`関数**
- https://pkg.go.dev/github.com/prometheus/client_golang@v1.13.1/prometheus#NewMetricWithTimestamp
- 一般的にはmetricsにtimestampは定義せず、PrometheusがExporterからscrapeした時刻をtimestampにする(なる)けどすでに外部のmetric sourceから作成されているmetricをPrometheusに取り込む場合(e.g. AWS CloudWatch MetricsをYACE等でPrometheusに取り込む)は元のmetric sourceのtimestampを明示的に使いたい場合は`NewMetricWithTimestamp`関数でメトリクスを生成する
- コード例は[YACEのコード](https://github.com/nerdswords/yet-another-cloudwatch-exporter/blob/master/pkg/prometheus.go)を参照