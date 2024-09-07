- Java Auto InstrumentationはPrometheus Exporterを通じてExemplarsを生成してくれる
  - **現状ExemplarsはHistogramsのメトリクス`http_server_duration_bucket`,`http_client_duration_bucket`にだけ生成される**
    > Currently the OTel Prometheus exporter only supports exemplars in histograms.
    - https://github.com/open-telemetry/opentelemetry-java-instrumentation/discussions/7564

## 注意事項
#### **ADOT Java Auto Instrumentation AgentもExemplarsを生成してくれるけど、メトリクス内のTraceIDがW3C形式(ここはX-Ray形式に変換してくれない)なのでX-Rayで検索する時にTraceIDを書き換える必要がある**
- ADOT Java Instrumentationページ
  - https://github.com/aws-observability/aws-otel-java-instrumentation
- 参考URL
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation/issues/4616
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation/discussions/4655
  - https://github.com/open-telemetry/opentelemetry-java-instrumentation/discussions/7564
  - https://github.com/open-telemetry/opentelemetry-java/issues/5141
##### **2023/10のアップデートにより、X-RayがW3C形式のTraceIDをサポートするようになった**
- つまり、**w3c形式のTraceをX-Rayにそのまま送れるようになった**
- ADOT Collector version 0.34.0以上のものを使う必要がある
  - ADOT Collectorがデフォルトでw3c形式でexportするか、別の設定が必要かは要確認
- 参考URL
  - https://aws.amazon.com/jp/about-aws/whats-new/2023/10/aws-x-ray-w3c-format-trace-ids-distributed-tracing/
  - https://aws.amazon.com/jp/about-aws/whats-new/2023/10/aws-x-ray-w3c-format-trace-ids-distributed-tracing/
  - https://aws.amazon.com/jp/about-aws/whats-new/2023/10/aws-x-ray-w3c-format-trace-ids-distributed-tracing/