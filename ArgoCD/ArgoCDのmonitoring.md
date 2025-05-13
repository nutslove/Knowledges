- https://argo-cd.readthedocs.io/en/stable/operator-manual/metrics/

## 設定関連
- 各ArgoCDコンポーネントごとにMetricsを開示するPort番号が異なる
- 以下はPrometheusのConfigのサンプル  
  ```yaml
  - job_name: 'argocd-application-controller'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: argocd
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      regex: argocd-application-controller.*
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_pod_ip]
      regex: (.+)
      target_label: __address__
      replacement: ${1}:8082

  - job_name: 'argocd-applicationset-controller'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: argocd
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      regex: argocd-applicationset-controller.*
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_pod_ip]
      regex: (.+)
      target_label: __address__
      replacement: ${1}:8080

  - job_name: 'argocd-server'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: argocd
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      regex: argocd-server.*
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_pod_ip]
      regex: (.+)
      target_label: __address__
      replacement: ${1}:8083

  - job_name: 'argocd-repo-server'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: argocd
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      regex: argocd-repo-server.*
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_pod_ip]
      regex: (.+)
      target_label: __address__
      replacement: ${1}:8084
  ```

## 各コンポーネントごとの主要メトリクス
|メトリクス名|概要|備考|
|---|---|---|
|``|||