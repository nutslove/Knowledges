- https://fluxcd.io/flux/monitoring/metrics/
- `8080`ポートの`/metrics`エンドポイントで開示される
- 以下はPrometheusのConfigのサンプル  
  ```yaml
  - job_name: 'flux'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: flux-system
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_pod_ip]
      regex: (.+)
      target_label: __address__
      replacement: ${1}:8080
  ```