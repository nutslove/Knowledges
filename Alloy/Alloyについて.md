- Grafana Labsが出しているOtel Collectorのディストリビューション
- AlloyからKubernetesのPodログを収集できる（Lokiへの連携ももちろん可能）
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.podlogs/

## インストール
- Kubernetes上にHelmでインストール
  - https://grafana.com/docs/alloy/latest/configure/kubernetes/

## AlloyのHealthCheck Endpoint
- `http://<alloy-host>:12345/-/healthy`

## Alloyのユーザ設定（非Root）
- https://grafana.com/docs/alloy/latest/configure/nonroot/
```yaml
alloy:
  securityContext:
    runAsUser: 473
    runAsGroup: 473

configReloader:
  securityContext:
    # this is the UID of the "nobody" user that the configReloader image runs as
    runAsUser: 65534
    runAsGroup: 65534
```