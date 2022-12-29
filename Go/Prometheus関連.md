- Prometheus Go client library
https://pkg.go.dev/github.com/prometheus/client_golang

- **`promhttp`Package**
  - *Sub-packages allow to expose the registered metrics via HTTP*
  - [promhttpドキュメント](https://pkg.go.dev/github.com/prometheus/client_golang@v1.14.0/prometheus/promhttp)

- **`MustRegister`関数**
  - Metrics have to be registered to be exposed by`MustRegister`func
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

- **`MustRegister`関数**

- **`Labels`関数**
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

- **`NewMetricWithTimestamp`関数**
  - https://pkg.go.dev/github.com/prometheus/client_golang@v1.13.1/prometheus#NewMetricWithTimestamp
  - 一般的にはmetricsにtimestampは定義せず、PrometheusがExporterからscrapeした時刻をtimestampにする(なる)けどすでに外部のmetric sourceから作成されているmetricをPrometheusに取り込む場合(e.g. AWS CloudWatch MetricsをYACE等でPrometheusに取り込む)は元のmetric sourceのtimestampを明示的に使いたい場合は`NewMetricWithTimestamp`関数でメトリクスを生成する
  - コード例は[YACEのコード](https://github.com/nerdswords/yet-another-cloudwatch-exporter/blob/master/pkg/prometheus.go)を参照