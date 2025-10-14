# otlpでのログ送信
- Otel CollectorからotlpでLokiにデータを送る時は、`otlphttp` exporterで、`http://<loki-addr>:3100/otlp` endpointに送ること
  - https://grafana.com/docs/loki/latest/send-data/otel/
  - 設定例  
    ```yaml
    exporters:
      debug:
      otlphttp/loki:
        endpoint: http://<loki-addr>:<port>/otlp
        tls:
          insecure: true
        headers:
          X-Scope-OrgID: otel
    ```

---

# Loki側の設定

## Attributes
- Attributesは、ログイベントに関する追加的なコンテキスト情報を提供するものであり、Key-Valueのペアで構成される。

### Attributeの種類
- Resource AttributesとLogRecord Attributesの2種類がある。

#### １．Resource Attributes
- ログイベントが生成されたエンティティに関する情報を提供する。
- 例: `service.name`, `service.version`, `host.name`, `cloud.provider`

##### １－１．Resource AttributesをLokiのLabelに変換する方法
- https://grafana.com/docs/loki/latest/configure/
- **distributorの`distributor.otlp_config.default_resource_attributes_as_index_labels`に変換するattributeを指定**
- (上記の項目を設定しなかった場合) defaultで以下のResource AttributesがLokiのLabelに変換される
  - [`service.name`, `service.namespace`, `service.instance.id`, `deployment.environment`, `deployment.environment.name`, `cloud.region`, `cloud.availability_zone`, `k8s.cluster.name`, `k8s.namespace.name`, `k8s.pod.name`, `k8s.container.name`, `container.name`, `k8s.replicaset.name`, `k8s.deployment.name`, `k8s.statefulset.name`, `k8s.daemonset.name`, `k8s.cronjob.name`, `k8s.job.name`]
- 上記のdefaultのResource Attributesを無視して、自分で指定したResource AttributesのみをLokiのLabelに変換したい場合は、`limits_config.otlp_config.resource_attributes`に変換するattributeを指定

#### ２．LogRecord Attributes
- 個々のログレコードで異なるKey-Valueペアを持つことができる。
- 例: `severity`, `timestamp`, `trace_id`, `span_id`

##### ２－１．LogRecord AttributesをLokiのLabelに変換する方法
- https://grafana.com/docs/loki/latest/configure/
- **Lokiの3.4.xまではLogRecord AttributesをLabelに変換することはできなかったが、Lokiの3.5.xからはLogRecord AttributesをLabelに変換できるようになった**
  - https://github.com/grafana/loki/pull/16673/
- `limits_config.otlp_config.log_attributes`に変換するattributeを指定
  - 設定例  
    ```yaml
    limits_config:
      otlp_config:
        log_attributes:
          - action: index_label
            attributes:
              - severity
              - trace_id
              - span_id
    ```

## Structured metadataの設定
- https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata
- `limits_config`ブロックで`allow_structured_metadata: true`でStructured metadataを有効にする必要がある  
  > If you are ingesting data in OpenTelemetry format, using Grafana Alloy or an OpenTelemetry Collector. Structured metadata was designed to support native ingestion of OpenTelemetry data.

  > Note structured metadata is required to support ingesting OTLP data.

## Lokiのotlphttpエンドポイントにログを送る際のLabelについて
- **otel sdkなどからLokiのotlphttpエンドポイントにログを送る際に、`resource.New`や`*slog.Logger`の各メソッド(e.g. `*slog.Logger.Warn`)で設定するAttributesは、Loki側でLabelではなく、Structured metadataとして認識される。**

---

# loki exporter *VS* otlp exporter
- https://grafana.com/docs/loki/latest/send-data/otel/native_otlp_vs_loki_exporter/
