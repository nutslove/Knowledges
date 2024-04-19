- https://prometheus-operator.dev/docs/user-guides/getting-started/
- https://github.com/prometheus-operator/prometheus-operator/tree/main

## 構成
![](./image/Prometheus%20Operator.jpg)

## CRD
- https://github.com/prometheus-operator/prometheus-operator/tree/main?tab=readme-ov-file#customresourcedefinitions
- Prometheus Operatorが持っているCRDは以下の通り
  - `Prometheus`
    - defines a desired Prometheus deployment.
  - `PrometheusAgent`
    - defines a desired Prometheus deployment, but running in Agent mode.
  - `Alertmanager`
  - `ThanosRuler`
  - `ServiceMonitor`
    - declaratively specifies how groups of Kubernetes services should be monitored. The Operator automatically generates Prometheus scrape configuration based on the current state of the objects in the API server.
  - `PodMonitor`
    - declaratively specifies how group of pods should be monitored. The Operator automatically generates Prometheus scrape configuration based on the current state of the objects in the API server.
  - `Probe`
    - declaratively specifies how groups of ingresses or static targets should be monitored. The Operator automatically generates Prometheus scrape configuration based on the definition.
  - `ScrapeConfig`
    - declaratively specifies scrape configurations to be added to Prometheus. This CustomResourceDefinition helps with scraping resources outside the Kubernetes cluster.
  - `PrometheusRule`
    - defines a desired set of Prometheus alerting and/or recording rules. The Operator generates a rule file, which can be used by Prometheus instances.
  - `AlertmanagerConfig`
    - declaratively specifies subsections of the Alertmanager configuration, allowing routing of alerts to custom receivers, and setting inhibit rules.

### ServiceMonitor
- https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.ServiceMonitor
- Prometheus Operatorは`ServiceMonitor`リソースを監視し、`selector`,`namespaceSelector`を満たす`Service`リソースの作成/更新/削除を検知し、設定に基づいてPrometheusのスクレイプ設定を自動的に更新する
  - **`Service`に紐づいている`Endpoints`に対してPrometheusの`kubernetes_sd_config`の`endpoints`タイプのconfigを生成する**
    - https://prometheus.io/docs/prometheus/latest/configuration/configuration/#endpoints
- `ServiceMonitor`の例
  ```yaml
  ---
  apiVersion: v1
  kind: Service
  metadata:
    name: example-app
    labels:
      app: example-app
  spec:
    selector:
      app: example-app
    ports:
    - name: web
      port: 80
      targetPort: 8080  
  ---
  apiVersion: monitoring.coreos.com/v1
  kind: ServiceMonitor
  metadata:
    name: example-app
    labels:
      app: example-app
  spec:
    selector:
      matchLabels:
        app: example-app
    endpoints:
    - port: web
      path: /metrics
      interval: 30s
    namespaceSelector:
      matchNames:
      - default
  ```
  - `spec.selector`
    - Label selector to select the Kubernetes `Endpoints` objects to monitor.
  - `spec.endpoints`
    - `port`
      - Name of the Service port which this endpoint refers to.
      - `Service`リソースの対象の`spec.ports[].name`を指定
    - `path`
      - Metricsをスクレイピングするエンドポイントパスを指定（省略した場合は`/metrics`になる）
    - `interval`
      - Metricsのスクレイピング間隔（省略した場合は`30s`）
  - `spec.namespaceSelector`
    - 監視対象の`Service`が存在するnamespaceを指定

### PodMonitor
- https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.PodMonitor
- Prometheus Operatorは`PodMonitor`リソースを監視し、`selector`,`namespaceSelector`を満たす`Pod`リソースの作成/更新/削除を検知し、**Prometheusの`kubernetes_sd_config`の`pod`タイプのconfigを生成する**
  - https://prometheus.io/docs/prometheus/latest/configuration/configuration/#pod
- `PodMonitor`の例
  ```yaml
  apiVersion: monitoring.coreos.com/v1
  kind: PodMonitor
  metadata:
    name: example-app
    labels:
      app: example-app
  spec:
    podMetricsEndpoints:
    - port: metrics
      path: /metrics
      interval: 30s
    selector: ------> Label selector to select the Kubernetes Pod objects.
      matchLabels:
        app: example-app
    namespaceSelector:
      matchNames:
      - default
  ```

## Thanosとの統合
- https://prometheus-operator.dev/docs/operator/thanos/
- Prometheus OperatorがサポートするThanosコンポーネントは`Thanos Ruler`と`Thanos Sidecar`の２つ。
  他のThanosコンポーネント(e.g. Querier, Store Gateway, Compactor)は別途デプロイする必要がある。
- Prometheus Operatorの`Prometheus` CRDはThanos Sidecarもサポートしており、以下のように`spec`フィールドの下に`thanos`を追加することでThanos Sidecarコンテナを入れることができる。  
  ```yaml
  ...
  spec:
    ...
    thanos:
      image: quay.io/thanos/thanos:v0.28.1
  ...
  ```
  > The `Prometheus` CRD has support for adding a Thanos sidecar to the Prometheus Pod. To enable the sidecar, the `thanos` section must be set to a non empty value. For example, the simplest configuration is to just set a valid thanos container image url.

  ただ、`image`だけだとオブジェクトストレージのアップロードが行われず、同じ pod に存在する prometheus コンテナが収集したメトリクスを query API で取得する動作となる  
  なのでオブジェクトストレージに送信するようにするためには、以下のように`objectStorageConfig`セクションも追加する必要がある  

#### ■ 手順(OpenShift環境)
- 参考ページ
  - https://prometheus-operator.dev/docs/operator/thanos/
  - https://github.com/thanos-io/thanos/blob/main/docs/storage.md
  - https://prometheus-operator.dev/docs/user-guides/getting-started/#deploying-prometheus
- まず、`s3-config.yml`ファイルを作成する
  ```yaml
  type: s3
  config:
    bucket: <S3バケット名>
    endpoint: s3.ap-northeast-1.amazonaws.com
    access_key: <アクセスキー>
    secret_key: <シークレットキー>
  ```
- `s3-config.yml`から`Secret`リソースを作成する
  ```shell
  oc create secret generic thanos-objstore-config --from-file=thanos.yaml=./s3-config.yml
  ```

- 上で作成した`Secret`を利用して`Prometheus`をデプロイする
  ```yaml
  apiVersion: monitoring.coreos.com/v1
  kind: Prometheus
  metadata:
    name: prometheus
  spec:
    serviceAccountName: prometheus
    serviceMonitorSelector: {}
    podMonitorSelector: {}
    resources:
      requests:
        memory: 400Mi
    enableAdminAPI: false
    thanos:
      image: quay.io/thanos/thanos:v0.34.1
      objectStorageConfig:
        key: thanos.yaml
        name: thanos-objstore-config
  ```
  - 以下のように`additionalArgs`に追加の設定をすることもできる
    ```yaml
    apiVersion: monitoring.coreos.com/v1
    kind: Prometheus
    metadata:
      name: prometheus
    spec:
      thanos:
        image: quay.io/thanos/thanos:v0.34.1
        objectStorageConfig:
          key: thanos.yaml
          name: thanos-objstore-config
        additionalArgs:
          - "--tsdb.path=/prometheus"
          - "--prometheus.url=http://localhost:9090"
          - "--tsdb.min-block-duration=2h"
          - "--tsdb.max-block-duration=2h"
    ```
- 以下コマンドでPrometheusの`Route`を作成する  
  ※`Prometheus`リソースを作成すると`prometheus-operated`Serviceが作成されるのでそれを利用する
  ```shell
  oc expose svc prometheus-operated
  ```

