## otlpでのログ送信
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

## Lokiのotlphttpエンドポイントにログを送る際のLabelについて
- **otel sdkなどからLokiのotlphttpエンドポイントにログを送る際に、`resource.New`や`*slog.Logger`の各メソッド(e.g. `*slog.Logger.Warn`)で設定するAttributesは、Loki側でLabelではなく、Structured metadataとして認識される。**
- Structured metadataをLokiのLabelに変換するためにはLoki側で設定が必要
  - **distributorの`default_resource_attributes_as_index_labels`に変換するattributeを指定**
  - **https://grafana.com/docs/loki/latest/send-data/otel/**
  - https://community.grafana.com/t/add-additional-index-labels-in-loki-3-0-via-otlp/121225/11
  - https://grafana.com/docs/loki/latest/configure/#distributor
- **`distributor`ブロックの`default_resource_attributes_as_index_labels`ではなく、`limits_config`ブロックで`otlp_config.resource_attributes`でTenantごとに設定することもできるっぽい**
  - https://grafana.com/docs/loki/latest/send-data/otel/#changing-the-default-mapping-of-otlp-to-loki-format

### メモ
- 2025/03/14 `default_resource_attributes_as_index_labels`にラベルに変換してほしいAttributesを追加したら、ログ送信時以下のエラーが出るようになった・・・。  
  ```shell
  loki body: 4error at least one label pair is required per stream
  ```  
  `limits_config`ブロックで`otlp_config.resource_attributes`で変換しようとしてもラベルに変換されず・・・。

### Structed metadata
- https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata/

## loki exporter *VS* otlp exporter
- https://grafana.com/docs/loki/latest/send-data/otel/native_otlp_vs_loki_exporter/