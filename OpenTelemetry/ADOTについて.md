- ADOT collector v0.29.0からAODTでもtail samplingが可能になった
  - https://github.com/aws-observability/aws-otel-collector/issues/1135
  - https://aws-otel.github.io/docs/ReleaseBlogs/aws-distro-for-opentelemetry-collector-v0.29.0

## ADOT Collectorで使えるReceiver、Processor、Exporter、Extensions
- https://github.com/aws-observability/aws-otel-collector?tab=readme-ov-file#adot-collector-built-in-components

## ADOT Java Auto Instrumentation Agent
- https://aws-otel.github.io/docs/getting-started/java-sdk/auto-instr  
  > Out of the box, it propagates traces using all of W3C Trace Context, B3, and X-Amzn-Trace-Id.
- https://github.com/aws-observability/aws-otel-java-instrumentation/blob/main/awspropagator/src/main/java/software/amazon/opentelemetry/awspropagator/AwsCompositePropagator.java
- 標準のOpenTelemetry Java Agentは、デフォルトでW3C Trace Contextのみをサポートしているが、  
  ADOT Java Agentは、W3C Trace Context, B3, X-Amzn-Trace-Idの３つをサポートしている