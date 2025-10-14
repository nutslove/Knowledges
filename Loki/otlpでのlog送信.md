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

## Structured metadataの設定
- https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata
- `limits_config`ブロックで`allow_structured_metadata: true`でStructured metadataを有効にする必要がある  
  > If you are ingesting data in OpenTelemetry format, using Grafana Alloy or an OpenTelemetry Collector. Structured metadata was designed to support native ingestion of OpenTelemetry data.

  > Note structured metadata is required to support ingesting OTLP data.

## Lokiのotlphttpエンドポイントにログを送る際のLabelについて
- **otel sdkなどからLokiのotlphttpエンドポイントにログを送る際に、`resource.New`や`*slog.Logger`の各メソッド(e.g. `*slog.Logger.Warn`)で設定するAttributesは、Loki側でLabelではなく、Structured metadataとして認識される。**
- Structured metadataをLokiのLabelに変換するためにはLoki側で設定が必要
  - **distributorの`default_resource_attributes_as_index_labels`に変換するattributeを指定**
  - **https://grafana.com/docs/loki/latest/send-data/otel/**
  - https://community.grafana.com/t/add-additional-index-labels-in-loki-3-0-via-otlp/121225/11
  - https://grafana.com/docs/loki/latest/configure/#distributor
- **`distributor`ブロックの`default_resource_attributes_as_index_labels`ではなく、`limits_config`ブロックで`otlp_config.resource_attributes`でTenantごとに設定することもできるっぽい**
  - https://grafana.com/docs/loki/latest/send-data/otel/#changing-the-default-mapping-of-otlp-to-loki-format

## メモ
- 2025/03/14 `default_resource_attributes_as_index_labels`にラベルに変換してほしいAttributesを追加したら、ログ送信時以下のエラーが出るようになった・・・。  
  ```shell
  loki body: 4error at least one label pair is required per stream
  ```  
  `limits_config`ブロックで`otlp_config.resource_attributes`で変換しようとしてもラベルに変換されず・・・。

---

# loki exporter *VS* otlp exporter
- https://grafana.com/docs/loki/latest/send-data/otel/native_otlp_vs_loki_exporter/



# Attributes
- Attributesは、ログイベントに関する追加的なコンテキスト情報を提供するものであり、Key-Valueのペアで構成される。

## Attributeの種類
- Resource AttributesとLogRecord Attributesの2種類がある。

### Resource Attributes
- ログイベントが生成されたエンティティに関する情報を提供する。
- 例: `service.name`, `service.version`, `host.name`, `cloud.provider`

### LogRecord Attributes 
- 個々のログレコードで異なるKey-Valueペアを持つことができる。
- 例: `severity`, `timestamp`, `trace_id`, `span_id`