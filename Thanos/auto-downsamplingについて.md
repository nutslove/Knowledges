## DownSampledデータ取得のための設定
- Querier実行時`--query.auto-downsampling`フラグを指定
- Query Frontend実行時`--query-range.request-downsampled`フラグを指定  
  > Make additional query for downsampled data in case of empty or incomplete response to range request.  
- GrafanaのDataSource設定時に「Custom query parameters」に`max_source_resolution=auto`を追加する必要があるみたい  
  ![](./image/grafana_datasource_setting.jpg)
  ![](./image/auto_downsampling_1.jpg)
  ![](./image/auto_downsampling_2.jpg)
  - `auto`以外に`5m`、`1h`を設定することも可能

## `--query.auto-downsampling`設定時のQuerierの処理ロジックについて
- queryのAPIのエンドポイントを公開している部分
  - https://github.com/thanos-io/thanos/blob/main/cmd/thanos/query.go
  ```go
  import (
         ・
         ・
    apiv1 "github.com/thanos-io/thanos/pkg/api/query"
         ・
         ・
  )
         ・
         ・
  // runQuery starts a server that exposes PromQL Query API. It is responsible for querying configured
  // store nodes, merging and duplicating the data to satisfy user query.
  func runQuery(
         ・
         ・
  ) error {
         ・
         ・
		api := apiv1.NewQueryAPI(
			logger,
			endpoints.GetEndpointStatus,
			engineFactory,
			apiv1.PromqlEngineType(defaultEngine),
			lookbackDeltaCreator,
			queryableCreator,
			// NOTE: Will share the same replica label as the query for now.
			rules.NewGRPCClientWithDedup(rulesProxy, queryReplicaLabels),
			targets.NewGRPCClientWithDedup(targetsProxy, queryReplicaLabels),
			metadata.NewGRPCClient(metadataProxy),
			exemplars.NewGRPCClientWithDedup(exemplarsProxy, queryReplicaLabels),
			enableAutodownsampling,
			enableQueryPartialResponse,
			enableRulePartialResponse,
			enableTargetPartialResponse,
			enableMetricMetadataPartialResponse,
			enableExemplarPartialResponse,
			queryReplicaLabels,
			flagsMap,
			defaultRangeQueryStep,
			instantDefaultMaxSourceResolution,
			defaultMetadataTimeRange,
			disableCORS,
			gate.New(
				extprom.WrapRegistererWithPrefix("thanos_query_concurrent_", reg),
				maxConcurrentQueries,
				gate.Queries,
			),
			store.NewSeriesStatsAggregatorFactory(
				reg,
				queryTelemetryDurationQuantiles,
				queryTelemetrySamplesQuantiles,
				queryTelemetrySeriesQuantiles,
			),
			reg,
			tenantHeader,
			defaultTenant,
			tenantCertField,
			enforceTenancy,
			tenantLabel,
		)

		api.Register(router.WithPrefix("/api/v1"), tracer, logger, ins, logMiddleware)

		srv := httpserver.New(logger, reg, comp, httpProbe,
			httpserver.WithListen(httpBindAddr),
			httpserver.WithGracePeriod(httpGracePeriod),
			httpserver.WithTLSConfig(httpTLSConfig),
		)
		srv.Handle("/", router)
         ・
         ・
  }
         ・
         ・
  ```
  - `query.NewQueryAPI`は`QueryAPI`structを返す
    ```go
    // QueryAPI is an API used by Thanos Querier.
    type QueryAPI struct {
    	baseAPI         *api.BaseAPI
    	logger          log.Logger
    	gate            gate.Gate
    	queryableCreate query.QueryableCreator
    	// queryEngine returns appropriate promql.Engine for a query with a given step.
    	engineFactory       *QueryEngineFactory
    	defaultEngine       PromqlEngineType
    	lookbackDeltaCreate func(int64) time.Duration
    	ruleGroups          rules.UnaryClient
    	targets             targets.UnaryClient
    	metadatas           metadata.UnaryClient
    	exemplars           exemplars.UnaryClient

    	enableAutodownsampling              bool
    	enableQueryPartialResponse          bool
    	enableRulePartialResponse           bool
    	enableTargetPartialResponse         bool
    	enableMetricMetadataPartialResponse bool
    	enableExemplarPartialResponse       bool
    	disableCORS                         bool

    	replicaLabels  []string
    	endpointStatus func() []query.EndpointStatus

    	defaultRangeQueryStep                  time.Duration
    	defaultInstantQueryMaxSourceResolution time.Duration
    	defaultMetadataTimeRange               time.Duration

    	queryRangeHist prometheus.Histogram

    	seriesStatsAggregatorFactory store.SeriesQueryPerformanceMetricsAggregatorFactory

    	tenantHeader    string
    	defaultTenant   string
    	tenantCertField string
    	enforceTenancy  bool
    	tenantLabel     string
    }

    // NewQueryAPI returns an initialized QueryAPI type.
    func NewQueryAPI(
    	logger log.Logger,
    	endpointStatus func() []query.EndpointStatus,
    	engineFactory *QueryEngineFactory,
    	defaultEngine PromqlEngineType,
    	lookbackDeltaCreate func(int64) time.Duration,
    	c query.QueryableCreator,
    	ruleGroups rules.UnaryClient,
    	targets targets.UnaryClient,
    	metadatas metadata.UnaryClient,
    	exemplars exemplars.UnaryClient,
    	enableAutodownsampling bool,
    	enableQueryPartialResponse bool,
    	enableRulePartialResponse bool,
    	enableTargetPartialResponse bool,
    	enableMetricMetadataPartialResponse bool,
    	enableExemplarPartialResponse bool,
    	replicaLabels []string,
    	flagsMap map[string]string,
    	defaultRangeQueryStep time.Duration,
    	defaultInstantQueryMaxSourceResolution time.Duration,
    	defaultMetadataTimeRange time.Duration,
    	disableCORS bool,
    	gate gate.Gate,
    	statsAggregatorFactory store.SeriesQueryPerformanceMetricsAggregatorFactory,
    	reg *prometheus.Registry,
    	tenantHeader string,
    	defaultTenant string,
    	tenantCertField string,
    	enforceTenancy bool,
    	tenantLabel string,
    ) *QueryAPI {
    	if statsAggregatorFactory == nil {
    		statsAggregatorFactory = &store.NoopSeriesStatsAggregatorFactory{}
    	}
    	return &QueryAPI{
    		baseAPI:                                api.NewBaseAPI(logger, disableCORS, flagsMap),
    		logger:                                 logger,
    		engineFactory:                          engineFactory,
    		defaultEngine:                          defaultEngine,
    		lookbackDeltaCreate:                    lookbackDeltaCreate,
    		queryableCreate:                        c,
    		gate:                                   gate,
    		ruleGroups:                             ruleGroups,
    		targets:                                targets,
    		metadatas:                              metadatas,
    		exemplars:                              exemplars,
    		enableAutodownsampling:                 enableAutodownsampling,
    		enableQueryPartialResponse:             enableQueryPartialResponse,
    		enableRulePartialResponse:              enableRulePartialResponse,
    		enableTargetPartialResponse:            enableTargetPartialResponse,
    		enableMetricMetadataPartialResponse:    enableMetricMetadataPartialResponse,
    		enableExemplarPartialResponse:          enableExemplarPartialResponse,
    		replicaLabels:                          replicaLabels,
    		endpointStatus:                         endpointStatus,
    		defaultRangeQueryStep:                  defaultRangeQueryStep,
    		defaultInstantQueryMaxSourceResolution: defaultInstantQueryMaxSourceResolution,
    		defaultMetadataTimeRange:               defaultMetadataTimeRange,
    		disableCORS:                            disableCORS,
    		seriesStatsAggregatorFactory:           statsAggregatorFactory,
    		tenantHeader:                           tenantHeader,
    		defaultTenant:                          defaultTenant,
    		tenantCertField:                        tenantCertField,
    		enforceTenancy:                         enforceTenancy,
    		tenantLabel:                            tenantLabel,

    		queryRangeHist: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
    			Name:    "thanos_query_range_requested_timespan_duration_seconds",
    			Help:    "A histogram of the query range window in seconds",
    			Buckets: prometheus.ExponentialBuckets(15*60, 2, 12),
    		}),
    	}
    }
    ```

