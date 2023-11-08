## OpenTelemetryを使ってOCI APMにトレース/メトリクスを送る方法
- APMのAdministratorでまずドメインを作成する必要がある
  ![](image/APM_1.jpg)
  ![](image/APM_2.jpg)
- 作成したドメインに入ってData Keysタブで自動生成されたPublic Keyを押さえておく(ShowやCopyで確認可能)
  ![](image/APM_3.jpg)
- 同じくデータを連携するエンドポイントも押さえておく
  ![](image/APM_4.jpg)
- 環境変数を設定する
  ~~~
  export OTEL_TRACES_EXPORTER=otlp
  export OTEL_SERVICE_NAME=<任意のMicroService名>
  export OTEL_EXPORTER_OTLP_TRACES_PROTOCOL=http/protobuf
  export OTEL_EXPORTER_OTLP_TRACES_HEADERS="authorization=dataKey <上の手順で押さえたPublic Data Key>"
  export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=<上の手順で押さえたエンドポイント>/20200101/opentelemetry/private/v1/traces/
  export OTEL_METRICS_EXPORTER=otlp
  export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=<上の手順で押さえたエンドポイント>/20200101/opentelemetry/v1/metrics
  export OTEL_EXPORTER_OTLP_METRICS_HEADERS="authorization=dataKey <上の手順で押さえたPublic Data Key>"
  export OTEL_EXPORTER_OTLP_METRICS_PROTOCOL=http/protobuf
  ~~~
  - https://docs.oracle.com/en-us/iaas/application-performance-monitoring/doc/configure-open-source-tracing-systems.html
- Opentelemetry java auto instrumentationをjavaagentとして指定してJavaアプリを起動する
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation

#### 参考ページ
- https://docs.oracle.com/ja-jp/iaas/application-performance-monitoring/doc/configure-open-source-tracing-systems.html#GUID-4D941163-F357-4839-8B06-688876D4C61F__GUID-6E301FC6-5CED-4AE1-9308-D6A342DE1339
- https://guides.micronaut.io/latest/micronaut-cloud-trace-oci-maven-kotlin.html
- https://blogs.oracle.com/cloud-infrastructure/post/opentelemetry-instrumentation-oci-apm
- https://docs.oracle.com/en-us/iaas/releasenotes/changes/abe28c55-1d52-4368-88e5-601b83dbca18/
- https://docs.oracle.com/ja-jp/iaas/application-performance-monitoring/doc/use-trace-explorer.html