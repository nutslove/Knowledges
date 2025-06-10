- Prometheus Go client library
https://pkg.go.dev/github.com/prometheus/client_golang
- **実装サンプルコード**
  - https://github.com/prometheus/client_golang/tree/main/examples

### **`promhttp`Package**
  - *Sub-packages allow to expose the registered metrics via HTTP*
  - [promhttpドキュメント](https://pkg.go.dev/github.com/prometheus/client_golang@v1.14.0/prometheus/promhttp)

#### `promhttp.Handler()`
- https://github.com/prometheus/client_golang/blob/main/prometheus/promhttp/http.go
- Prometheusのメトリクスをエクスポートするための HTTP ハンドラーを提供する関数

### `promauto`と`prometheus`の違い
#### `promauto`
- メトリクスの自動登録を提供
- 例えば`promauto.NewCounterVec`関数を使用すると、定義したメトリクスは自動的にPrometheusのデフォルトレジストリに登録される。
  - `MustRegister`や`Register`でメトリクスをレジストリに登録する必要がなくなる
- コード内で明示的にメトリクスを登録する必要がないため、便利で簡潔なコードを書くことができる。  
  ただし、メトリクスの登録解除や、カスタムレジストリへの登録はできない。
#### `prometheus`
- `MustRegister`や`Register`メソッドを使って明示的にメトリクスをレジストリに登録する必要がある
- 例  
  ~~~go
  var (
      errorCounter = prometheus.NewCounterVec(
          prometheus.CounterOpts{
              Name: "error_count",
              Help: "The total number of errors with reasons.",
          },
          []string{"reasons"},
      )
  )

  func init() {
      // メトリクスをデフォルトレジストリに登録
      prometheus.MustRegister(errorCounter)
  }
  ~~~
##### レジストリ(Registry)とは
- メトリクスの集合を管理するための構造体(struct)  
  ~~~go
  // Registry registers Prometheus collectors, collects their metrics, and gathers
  // them into MetricFamilies for exposition. It implements Registerer, Gatherer,
  // and Collector. The zero value is not usable. Create instances with
  // NewRegistry or NewPedanticRegistry.
  //
  // Registry implements Collector to allow it to be used for creating groups of
  // metrics. See the Grouping example for how this can be done.
  type Registry struct {
  	mtx                   sync.RWMutex
  	collectorsByID        map[uint64]Collector // ID is a hash of the descIDs.
  	descIDs               map[uint64]struct{}
  	dimHashesByName       map[string]uint64
  	uncheckedCollectors   []Collector
  	pedanticChecksEnabled bool
  }
  ~~~
- Prometheusのクライアントライブラリを使用する際、メトリクスはレジストリに登録されている必要がありる。登録されたメトリクスは、Prometheusサーバーによって定期的に収集され、監視やアラートに使用される。
- レジストリは、メトリクスの登録、登録解除、収集などの機能を提供
  - メトリクスの登録解除は、特定のメトリクスを収集対象から外したい場合に使用。ただし、一般的には、メトリクスの登録解除を明示的に行う必要はない。
- 以下２種類のレジストリが提供されている
  1. **デフォルトレジストリ（Default Registry）**
     - Prometheusクライアントライブラリには、グローバルなデフォルトレジストリが用意されている。
     - `promauto`パッケージや`prometheus.MustRegister`関数を使用してメトリクスを登録する場合、デフォルトレジストリが使用される。
     - デフォルトレジストリは、`prometheus.DefaultRegisterer`変数や`prometheus.DefaultGatherer`変数を通じてアクセスできる。
     - ほとんどの場合、デフォルトレジストリで十分。
  2. **カスタムレジストリ（Custom Registry）**
     - 特別な要件がある場合、カスタムレジストリを作成して使用することができる。
     - カスタムレジストリは、`prometheus.NewRegistry`関数を使用して作成する。
     - カスタムレジストリを使用すると、メトリクスの名前空間を分離したり、特定のメトリクスセットを個別に管理したりすることができる。
     - カスタムレジストリにメトリクスを登録するには、`prometheus.Register`関数や`prometheus.MustRegister`関数を使用する。

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