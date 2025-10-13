- PrometheusはすべてのJobに対して、Remote Write設定にて動的にヘッダー(e.g. `THANOS-TENANT`などのテナント識別用ヘッダー)を付与することができない
- **cortex-tenant**というラツールを使えば、PrometheusのJobごとに異なるヘッダーを付与できる

## Cortex-Tenant
- https://github.com/blind-oracle/cortex-tenant
- 特定のLabelの値（デフォルトは `__tenant__` Label）に基づいて、特定のHTTPヘッダー（デフォルトは `X-Scope-OrgID` ヘッダー）を追加するプロキシサーバー
- **MimirとLokiのために作られたツールだが、ヘッダー名を（`THANOS-TENANT`ヘッダー名に）変更出来て、ThanosのReceiverにも使える**（動作確認済み）
- `/push`エンドポイントでRemote Writeを受け付ける

### 設定例
#### `cortex-tenant`側の設定例（k8s manifest file）  
- https://github.com/blind-oracle/cortex-tenant/tree/main/deploy/k8s から確認可能  
  ```yaml
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      release: cortex-tenant
      app.kubernetes.io/name: cortex-tenant
    name: cortex-tenant
    namespace: monitoring
  spec:
    replicas: 1
    selector:
      matchLabels:
        release: cortex-tenant
    template:
      metadata:
        labels:
          release: cortex-tenant
        namespace: monitoring
      spec:
        containers:
          - image: ghcr.io/blind-oracle/cortex-tenant:latest
            imagePullPolicy: IfNotPresent
            name: cortex-tenant
            ports:
              - containerPort: 8080
                name: cortex-tenant
                protocol: TCP
            volumeMounts:
              - mountPath: /data/
                name: config-file
        volumes:
          - configMap:
              name: cortex-tenant-configmap
            name: config-file
  ---
  apiVersion: v1
  kind: ConfigMap
  metadata:
    labels:
      app.kubernetes.io/name: cortex-tenant
    name: cortex-tenant-configmap
    namespace: monitoring
  data:
    cortex-tenant.yml: |
      # Where to listen for incoming write requests from Prometheus
      listen: 0.0.0.0:8080
      # Profiling API, remove to disable
      listen_pprof: 0.0.0.0:7008
      # Where to send the modified requests (Cortex)
      target: http://thanos-routing-receiver.monitoring.svc:19291/api/v1/receive
      # Log level
      log_level: warn
      # HTTP request timeout
      timeout: 10s
      # HTTP request idle timeout
      idle_timeout: 60s
      # Timeout to wait on shutdown to allow load balancers detect that we're going away.
      # During this period after the shutdown command the /alive endpoint will reply with HTTP 503.
      # Set to 0s to disable.
      timeout_shutdown: 10s
      # Max number of parallel incoming HTTP requests to handle
      concurrency: 1000
      # Whether to forward metrics metadata from Prometheus to Cortex
      # Since metadata requests have no timeseries in them - we cannot divide them into tenants
      # So the metadata requests will be sent to the default tenant only, if one is not defined - they will be dropped
      metadata: false
      # Address where metrics are available
      # env: CT_LISTEN_METRICS_ADDRESS
      listen_metrics_address: 0.0.0.0:9090

      # If true, then a label with the tenant’s name will be added to the metrics
      # env: CT_METRICS_INCLUDE_TENANT
      metrics_include_tenant: true

      tenant:
        # Which label to look for the tenant information
        label: tenant ★→ ヘッダーに変換するラベル名を指定

        # Optional hard-coded prefix with delimeter for all tenant values.
        # Delimeters allowed for use:
        # https://grafana.com/docs/mimir/latest/configure/about-tenant-ids/
        prefix: ""
        # If true will use the tenant ID of the inbound request as the prefix of the new tenant id.
        # Will be automatically suffixed with a `-` character.
        # Example:
        #   Prometheus forwards metrics with `X-Scope-OrgID: Prom-A` set in the inbound request.
        #   This would result in the tenant prefix being set to `Prom-A-`.
        # https://grafana.com/docs/mimir/latest/configure/about-tenant-ids/
        prefix_prefer_source: false
        # Whether to remove the tenant label from the request
        label_remove: false
        # To which header to add the tenant ID
        header: THANOS-TENANT ★→ 変換後のヘッダー名を指定（デフォルトはX-Scope-OrgID）
        # Which tenant ID to use if the label is missing in any of the timeseries
        # If this is not set or empty then the write request with missing tenant label
        # will be rejected with HTTP code 400
        default: cortex-tenant-default
        # Enable if you want all metrics from Prometheus to be accepted with a 204 HTTP code
        # regardless of the response from Cortex. This can lose metrics if Cortex is
        # throwing rejections.
        accept_all: false
  ---
  apiVersion: v1
  kind: Service
  metadata:
    labels:
      app.kubernetes.io/name: cortex-tenant
    name: cortex-tenant
    namespace: monitoring
  spec:
    ports:
      - name: cortex-tenant
        port: 8080
        protocol: TCP
        targetPort: cortex-tenant
      - name: http-metrics
        port: 9090
        targetPort: http-metrics
        protocol: TCP
    selector:
      release: cortex-tenant
  ```

#### Prometheus側の設定例
- configMapの例  
  ```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: prometheus-thanos-config
    namespace: monitoring
  data:
    prometheus.yml: |
      global:
        scrape_interval: 30s

      remote_write:
        - url: http://cortex-tenant.monitoring.svc:8080/push ★→ `cortex-tenant`をRemote Write先に指定

      scrape_configs:
      - job_name: 'kube-state-metrics'
        static_configs:
        - targets: ['kube-state-metrics.monitoring.svc.cluster.local:8080']
          labels:
            tenant: kube-state-metrics ★→ ここで指定したラベル名が`cortex-tenant`の設定で指定したラベル名と一致している必要があり、、ここで指定した値が`THANOS-TENANT`ヘッダーの値になる

      - job_name: 'grafana'
        static_configs:
        - targets: ['grafana.monitoring.svc.cluster.local']
          labels:
            tenant: grafana ★→ ここで指定したラベル名が`cortex-tenant`の設定で指定したラベル名と一致している必要があり、、ここで指定した値が`THANOS-TENANT`ヘッダーの値になる
  ```