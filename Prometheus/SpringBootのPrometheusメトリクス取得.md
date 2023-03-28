#### 前提知識
- **Micrometer**
  - Javaのメトリクスのためのライブラリ
  - https://micrometer.io/
- **Actuator**
  - ActuatorはSpringBootアプリケーションの監視および管理に役立つ多くの追加機能を提供するもので、その中にMicrometerが含まれている。Actuatorを有効化すると自動的にMicrometerによるメトリクスが生成される。
  - https://spring.pleiades.io/spring-boot/docs/current/reference/html/actuator.html#actuator.metrics
    > Spring Boot Actuator は、以下を含む多数のモニタリングシステム (英語)  をサポートするアプリケーションメトリクスファサードである Micrometer (英語)  の依存関係管理と自動構成を提供します。
  - https://www.baeldung.com/micrometer#6-spring-integration
- 参考URL
  - https://engineering.linecorp.com/ja/blog/metrics-capture-spring
  - https://speakerdeck.com/hainet50b/spring-boot-3-dot-0-obuzababiriteituagaido
  - https://spring.pleiades.io/spring-boot/docs/current/reference/html/actuator.html
  - https://micrometer.io/

### Micrometerで取得できるメトリクス
- Grafana Dashboardが用意されている
  - https://grafana.com/grafana/dashboards/4701-jvm-micrometer/