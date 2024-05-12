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
- 実際にクエリーの処理が記述されているのは`pkg/api/query/v1.go`
  - https://github.com/thanos-io/thanos/blob/main/pkg/api/query/v1.go
  - 例えばquery_rangeに関する処理は`queryRange`メソッドに定義されている  
    ```go
    func (qapi *QueryAPI) queryRange(r *http.Request) (interface{}, []error, *api.ApiError, func()) {
    	start, err := parseTime(r.FormValue("start"))
    	if err != nil {
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}
    	end, err := parseTime(r.FormValue("end"))
    	if err != nil {
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}
    	if end.Before(start) {
    		err := errors.New("end timestamp must not be before start time")
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}
    	step, apiErr := qapi.parseStep(r, qapi.defaultRangeQueryStep, int64(end.Sub(start)/time.Second))

    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	if step <= 0 {
    		err := errors.New("zero or negative query resolution step widths are not accepted. Try a positive integer")
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}

    	// For safety, limit the number of returned points per timeseries.
    	// This is sufficient for 60s resolution for a week or 1h resolution for a year.
    	if end.Sub(start)/step > 11000 {
    		err := errors.New("exceeded maximum resolution of 11,000 points per timeseries. Try decreasing the query resolution (?step=XX)")
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}

    	ctx := r.Context()
    	if to := r.FormValue("timeout"); to != "" {
    		var cancel context.CancelFunc
    		timeout, err := parseDuration(to)
    		if err != nil {
    			return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    		}

    		ctx, cancel = context.WithTimeout(ctx, timeout)
    		defer cancel()
    	}

    	enableDedup, apiErr := qapi.parseEnableDedupParam(r)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	replicaLabels, apiErr := qapi.parseReplicaLabelsParam(r)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	storeDebugMatchers, apiErr := qapi.parseStoreDebugMatchersParam(r)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	// If no max_source_resolution is specified fit at least 5 samples between steps.
    	maxSourceResolution, apiErr := qapi.parseDownsamplingParamMillis(r, step/5)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	enablePartialResponse, apiErr := qapi.parsePartialResponseParam(r, qapi.enableQueryPartialResponse)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	shardInfo, apiErr := qapi.parseShardInfo(r)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	engine, _, apiErr := qapi.parseEngineParam(r)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	lookbackDelta := qapi.lookbackDeltaCreate(maxSourceResolution)
    	// Get custom lookback delta from request.
    	lookbackDeltaFromReq, apiErr := qapi.parseLookbackDeltaParam(r)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}
    	if lookbackDeltaFromReq > 0 {
    		lookbackDelta = lookbackDeltaFromReq
    	}

    	queryStr, tenant, ctx, err := tenancy.RewritePromQL(ctx, r, qapi.tenantHeader, qapi.defaultTenant, qapi.tenantCertField, qapi.enforceTenancy, qapi.tenantLabel, r.FormValue("query"))
    	if err != nil {
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}

    	// Record the query range requested.
    	qapi.queryRangeHist.Observe(end.Sub(start).Seconds())

    	// We are starting promQL tracing span here, because we have no control over promQL code.
    	span, ctx := tracing.StartSpan(ctx, "promql_range_query")
    	defer span.Finish()

    	var seriesStats []storepb.SeriesStatsCounter
    	qry, err := engine.NewRangeQuery(
    		ctx,
    		qapi.queryableCreate(
    			enableDedup,
    			replicaLabels,
    			storeDebugMatchers,
    			maxSourceResolution,
    			enablePartialResponse,
    			false,
    			shardInfo,
    			query.NewAggregateStatsReporter(&seriesStats),
    		),
    		promql.NewPrometheusQueryOpts(false, lookbackDelta),
    		queryStr,
    		start,
    		end,
    		step,
    	)
    	if err != nil {
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}

    	res := qry.Exec(ctx)

    	analysis, err := qapi.parseQueryAnalyzeParam(r, qry)
    	if err != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	tracing.DoInSpan(ctx, "query_gate_ismyturn", func(ctx context.Context) {
    		err = qapi.gate.Start(ctx)
    	})
    	if err != nil {
    		return nil, nil, &api.ApiError{Typ: api.ErrorExec, Err: err}, qry.Close
    	}
    	defer qapi.gate.Done()

    	beforeRange := time.Now()
    	if res.Err != nil {
    		switch res.Err.(type) {
    		case promql.ErrQueryCanceled:
    			return nil, nil, &api.ApiError{Typ: api.ErrorCanceled, Err: res.Err}, qry.Close
    		case promql.ErrQueryTimeout:
    			return nil, nil, &api.ApiError{Typ: api.ErrorTimeout, Err: res.Err}, qry.Close
    		}
    		return nil, nil, &api.ApiError{Typ: api.ErrorExec, Err: res.Err}, qry.Close
    	}
    	aggregator := qapi.seriesStatsAggregatorFactory.NewAggregator(tenant)
    	for i := range seriesStats {
    		aggregator.Aggregate(seriesStats[i])
    	}
    	aggregator.Observe(time.Since(beforeRange).Seconds())

    	// Optional stats field in response if parameter "stats" is not empty.
    	var qs stats.QueryStats
    	if r.FormValue(Stats) != "" {
    		qs = stats.NewQueryStats(qry.Stats())
    	}
    	return &queryData{
    		ResultType:    res.Value.Type(),
    		Result:        res.Value,
    		Stats:         qs,
    		QueryAnalysis: analysis,
    	}, res.Warnings.AsErrors(), nil, qry.Close
    }
    ```
    - auto-downsamplingと関係するのは以下の部分  
      ```go
                            ・
                            ・
    	step, apiErr := qapi.parseStep(r, qapi.defaultRangeQueryStep, int64(end.Sub(start)/time.Second))

    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}

    	if step <= 0 {
    		err := errors.New("zero or negative query resolution step widths are not accepted. Try a positive integer")
    		return nil, nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}, func() {}
    	}

    	// If no max_source_resolution is specified fit at least 5 samples between steps.
    	maxSourceResolution, apiErr := qapi.parseDownsamplingParamMillis(r, step/5)
    	if apiErr != nil {
    		return nil, nil, apiErr, func() {}
    	}
                            ・
                            ・
      ```
    - `parseDownsamplingParamMillis`メソッドの中身  
      `--query.auto-downsampling`フラグが指定されている or `max_source_resolution`パラメータの値が`auto`(`max_source_resolution=auto`)の場合は`maxSourceResolution`は`step/5`になり、`--query.auto-downsampling`フラグが指定されてない場合は`maxSourceResolution`は`0`(=rawデータ)になる  
      ```go
      func (qapi *QueryAPI) parseDownsamplingParamMillis(r *http.Request, defaultVal time.Duration) (maxResolutionMillis int64, _ *api.ApiError) {
      	maxSourceResolution := 0 * time.Second

      	val := r.FormValue(MaxSourceResolutionParam)
      	if qapi.enableAutodownsampling || (val == "auto") {
      		maxSourceResolution = defaultVal
      	}
      	if val != "" && val != "auto" {
      		var err error
      		maxSourceResolution, err = parseDuration(val)
      		if err != nil {
      			return 0, &api.ApiError{Typ: api.ErrorBadData, Err: errors.Wrapf(err, "'%s' parameter", MaxSourceResolutionParam)}
      		}
      	}

      	if maxSourceResolution < 0 {
      		return 0, &api.ApiError{Typ: api.ErrorBadData, Err: errors.Errorf("negative '%s' is not accepted. Try a positive integer", MaxSourceResolutionParam)}
      	}

      	return int64(maxSourceResolution / time.Millisecond), nil
      }
      ```
      - **stepとは、グラフ上の各データポイント間の時間間隔を指す。例えば、00:00から01:00までの1時間の範囲で取得していて、step=15sの場合はグラフ上の各データポイントが15秒間隔で表示されることを意味する。**
      - `step`は`parseStep`メソッド(`step, apiErr := qapi.parseStep(r, qapi.defaultRangeQueryStep, int64(end.Sub(start)/time.Second))`)で求めている  
        ```go
        func (qapi *QueryAPI) parseStep(r *http.Request, defaultRangeQueryStep time.Duration, rangeSeconds int64) (time.Duration, *api.ApiError) {
        	// Overwrite the cli flag when provided as a query parameter.
        	if val := r.FormValue(Step); val != "" {
        		var err error
        		defaultRangeQueryStep, err = parseDuration(val)
        		if err != nil {
        			return 0, &api.ApiError{Typ: api.ErrorBadData, Err: errors.Wrapf(err, "'%s' parameter", Step)}
        		}
        		return defaultRangeQueryStep, nil
        	}
        	// Default step is used this way to make it consistent with UI.
        	d := time.Duration(math.Max(float64(rangeSeconds/250), float64(defaultRangeQueryStep/time.Second))) * time.Second
        	return d, nil
        }
        ```