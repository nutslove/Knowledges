- vminsert, vmselect, vmstorageそれぞれmetricsを開示している
- Clusterバージョン専用のGrafana Dashboardが用意されている
  - https://grafana.com/grafana/dashboards/11176-victoriametrics-cluster/

##### `vminsert`
- `http://<vminsertのIP>:8480/metrics`から取得できる
##### `vmselect`
- `http://<vmselectのIP>:8481/metrics`から取得できる
##### `vmstorage`
- `http://<vmstorageのIP>:8482/metrics`から取得できる

### 各コンポーネントのメトリクスを収集するためのPrometheus側の設定
~~~yaml
- job_name: 'vmstorage'
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_name]
    regex: ^vmstorage-.+
    action: keep
  - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
  - source_labels: [__meta_kubernetes_pod_ip]
    regex: (.+)
    target_label: __address__
    replacement: ${1}:8482
- job_name: 'vminsert'
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_name]
    regex: ^vminsert-.+
    action: keep
  - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
  - source_labels: [__meta_kubernetes_pod_ip]
    regex: (.+)
    target_label: __address__
    replacement: ${1}:8480
- job_name: 'vmselect'
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_name]
    regex: ^vmselect-.+
    action: keep
  - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
  - source_labels: [__meta_kubernetes_pod_ip]
    regex: (.+)
    target_label: __address__
    replacement: ${1}:8481
~~~
- ストレージとしてEFSを使っている場合はCloudWatchでEFSも監視すること
  - https://grafana.com/grafana/dashboards/653-aws-efs/
