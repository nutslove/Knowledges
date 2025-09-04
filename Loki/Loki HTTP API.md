- **https://grafana.com/docs/loki/latest/reference/loki-http-api/**

> [!NOTE]  
> Loki Microservice modeでは、これらのAPIは *gateway(nginx)*（8080ポート）に対して実行

> [!NOTE]  
> Multi Tenancyモードを使っている場合は、`X-Scope-OrgID`ヘッダーを付けること  
> - 例 
>   ```shell
>   curl -s -H 'X-Scope-OrgID: <テナント名>' http://<gatewayのIP>:8080/loki/api/v1/labels
>   ```

## ラベル一覧取得
```shell
curl -s http://<gatewayのIP>:8080/loki/api/v1/labels
```
- output例  
  ```json
  {
    "status": "success",
    "data": [
      "app",
      "cluster",
      "container",
      "container_runtime",
      "instance",
      "job",
      "namespace",
      "pod",
      "service_name"
    ]
  }
  ```

## 特定のStreamが持っているラベル一覧（ラベルと値のすべての組み合わせ）取得
> returns the list of streams (unique set of labels) that match a certain given selector.

```shell
curl -s "http://localhost:3100/loki/api/v1/series" \
  --data-urlencode 'match[]={container_name=~"prometheus.*", component="server"}' \
  --data-urlencode 'match[]={app="loki"}' | jq '.'
```
- output  
  ```json
  {
    "status": "success",
    "data": [
      {
        "container_name": "loki",
        "app": "loki",
        "stream": "stderr",
        "filename": "/var/log/pods/default_loki-stack-0_50835643-1df0-11ea-ba79-025000000001/loki/0.log",
        "name": "loki",
        "job": "default/loki",
        "controller_revision_hash": "loki-stack-757479754d",
        "statefulset_kubernetes_io_pod_name": "loki-stack-0",
        "release": "loki-stack",
        "namespace": "default",
        "instance": "loki-stack-0"
      },
      {
        "chart": "prometheus-9.3.3",
        "container_name": "prometheus-server-configmap-reload",
        "filename": "/var/log/pods/default_loki-stack-prometheus-server-696cc9ddff-87lmq_507b1db4-1df0-11ea-ba79-025000000001/prometheus-server-configmap-reload/0.log",
        "instance": "loki-stack-prometheus-server-696cc9ddff-87lmq",
        "pod_template_hash": "696cc9ddff",
        "app": "prometheus",
        "component": "server",
        "heritage": "Tiller",
        "job": "default/prometheus",
        "namespace": "default",
        "release": "loki-stack",
        "stream": "stderr"
      },
      {
        "app": "prometheus",
        "component": "server",
        "filename": "/var/log/pods/default_loki-stack-prometheus-server-696cc9ddff-87lmq_507b1db4-1df0-11ea-ba79-025000000001/prometheus-server/0.log",
        "release": "loki-stack",
        "namespace": "default",
        "pod_template_hash": "696cc9ddff",
        "stream": "stderr",
        "chart": "prometheus-9.3.3",
        "container_name": "prometheus-server",
        "heritage": "Tiller",
        "instance": "loki-stack-prometheus-server-696cc9ddff-87lmq",
        "job": "default/prometheus"
      }
    ]
  }
  ```

## 特定ラベルの値一覧取得
```shell
curl -s http://<gatewayのIP>:8080/loki/api/v1/label/<ラベル名>/values
```
- output例  
  ```json
  {
    "status": "success",
    "data": [
      "alloy",
      "argocd-application-controller",
      "argocd-applicationset-controller",
      "argocd-notifications-controller",
      "argocd-repo-server",
      "argocd-server",
      "cert-controller",
      "clickhouse",
      "codebase-analysis",
      "compactor",
      "controller",
      "csi-provisioner",
      "distributor",
      "external-secrets",
      "grafana",
      "index-gateway",
      "ingester",
      "langfuse-web",
      "langfuse-worker",
      "manager",
      "nginx",
      "prometheus",
      "querier",
      "query-frontend",
      "split-brain-fix",
      "tf-runner",
      "thanos-compactor",
      "thanos-ingesting-receiver",
      "thanos-routing-receiver",
      "thanos-store",
      "tofu-controller",
      "webhook"
    ]
  }
  ```