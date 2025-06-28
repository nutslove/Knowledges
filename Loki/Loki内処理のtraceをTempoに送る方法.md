# 前提
- Loki内の処理に関するtraceをTempo(Otel Collector)に送信することができる
- Apache Thriftプロトコル（14268ポート）で送信する必要がある

# Loki側の設定（Helm）
- https://grafana.com/docs/loki/latest/setup/install/helm/monitor-and-alert/with-grafana-cloud/#enable-loki-tracing
- `loki.tracing.enabled`を`true`にする
- `loki.<各コンポーネント>.extraEnv`に以下の環境変数を追加する
  - `JAEGER_ENDPOINT`: TempoまたはOtel Collectorのエンドポイント（`<エンドポイント>:14268/api/traces`）
  - `JAEGER_SAMPLER_TYPE`: サンプリングタイプ（例: `const`）
  - `JAEGER_SAMPLER_PARAM`: サンプリングパラメータ（例: `1.0`）
- Helmの`values.yaml`の全体  
  ```yaml
  loki:
    auth_enabled: true # default is true
    images:
      tag: "3.5.1"
    schemaConfig:
      configs:
        - from: "2024-04-01"
          store: tsdb
          object_store: s3
          schema: v13
          index:
            prefix: loki_index_
            period: 24h
    tracing:
      enabled: true
    ingester:
      chunk_encoding: snappy
    querier:
      # Default is 4, if you have enough memory and CPU you can increase, reduce if OOMing
      max_concurrent: 4
    pattern_ingester:
      enabled: true
    limits_config:
      allow_structured_metadata: true
      volume_enabled: true
      query_timeout: 3m # Timeout when querying backends (ingesters or storage) during the execution of a query request. Default is 1m. 
    storage:
      type: s3
      s3:
        endpoint: http://minio-dev.minio-dev.svc:9000
        s3: s3://loki-data-bucket
        s3ForcePathStyle: true
        insecure: true
        accessKeyId: minioadmin
        secretAccessKey: minioadmin
      bucketNames:
        chunks: loki-data-bucket
        ruler: loki-ruler-bucket
        admin: loki-admin-bucket
    commonConfig:
      replication_factor: 3

  deploymentMode: Distributed

  gateway:
    replicas: 2
    affinity: # defaultのanti-affinityを無効化
    service:
      type: NodePort
      nodePort: 31100
  ingester:
    replicas: 3 # To ensure data durability with replication
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
    zoneAwareReplication:
      enabled: false
    affinity: # defaultのanti-affinityを無効化（insgesterの場合、zoneAwareReplication.enabled: falseも必要）
    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: ScheduleAnyway
        labelSelector:
          matchLabels:
            app.kubernetes.io/component: ingester
  querier:
    replicas: 3 # Improve query performance via parallelism
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
    maxUnavailable: 2
    affinity: # defaultのanti-affinityを無効化
    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: ScheduleAnyway
        labelSelector:
          matchLabels:
            app.kubernetes.io/component: querier
  queryFrontend:
    replicas: 2
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
    maxUnavailable: 1
  queryScheduler:
    replicas: 2
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
  distributor:
    replicas: 3 
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
    maxUnavailable: 2
    affinity: # defaultのanti-affinityを無効化
    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: ScheduleAnyway
        labelSelector:
          matchLabels:
            app.kubernetes.io/component: distributor
  compactor:
    replicas: 1
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
  indexGateway:
    replicas: 2
    extraEnv:
      - name: JAEGER_ENDPOINT
        value: "http://multi-tenant-tempo-distributor.monitoring.svc.cluster.local:14268/api/traces"
      - name: JAEGER_SAMPLER_TYPE
        value: "const"
      - name: JAEGER_SAMPLER_PARAM
        value: "1.0"
    maxUnavailable: 1

  # Disable (set 0) components that are not needed in Microservices Mode
  test:
    enabled: false
  lokiCanary:
    enabled: false
  bloomPlanner:
    replicas: 0
  bloomBuilder:
    replicas: 0
  bloomGateway:
    replicas: 0

  backend:
    replicas: 0
  read:
    replicas: 0
  write:
    replicas: 0

  singleBinary:
    replicas: 0
  ```

- Tempoに直接送る場合はTempo側も`traces.jaeger.thriftHttp.enabled: true`にする必要がある  
  ```yaml
  # multitenancyEnabled: true
  multitenancyEnabled: false
  storage:
    trace:
      backend: s3
      s3:
        endpoint: minio-dev.minio-dev.svc.cluster.local:9000
        bucket: tempo-bucket
        access_key: minioadmin
        secret_key: minioadmin
        insecure: true
  distributor:
    replicas: 1
  ingester:
    replicas: 1
    config:
      replication_factor: 1
  compactor:
    replicas: 1
    config:
      compaction:
        block_retention: 720h  # データを30日間（720時間）保存
  querier:
    replicas: 1
  queryFrontend:
    replicas: 1
  traces:
    otlp:
      grpc:
        enabled: true
      http:
        enabled: true
    jaeger:
      thriftHttp:
        enabled: true # 14268 port
  ```