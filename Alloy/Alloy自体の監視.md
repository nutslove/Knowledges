- https://grafana.com/docs/alloy/latest/troubleshoot/controller_metrics/
- Alloy自体は`12345`ポート番号、`/metrics`エンドポイントでメトリクスを開示する
- 以下はPrometheusのConfigのサンプル  
  ```yaml
  - job_name: 'alloy'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: monitoring
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      regex: .*alloy-.*
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_pod_ip]
      regex: (.+)
      target_label: __address__
      replacement: ${1}:12345
  ```