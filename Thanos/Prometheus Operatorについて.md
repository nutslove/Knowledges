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
  ```yaml
  piVersion: monitoring.coreos.com/v1
  kind: Prometheus
  ...
  spec:
    ...
    thanos:
      image: quay.io/thanos/thanos:v0.32.4
      objectStorageConfig:
        key: minio.yml --> オブジェクトストレージに関するconfigファイル
        name: minio-secret
  ```  
  ・`minio.yml`  
  ```yaml
  type: S3
  config:
    bucket: test1
    endpoint: minio.centre.com:9000
    access_key: admin
    insecure: true
    signature_version2: false
    secret_key: [password]
  prefix: ""
  ```
  **https://zenn.dev/zenogawa/articles/k8s_cluster_metrics**