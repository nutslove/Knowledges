- 参考URL
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.awsfirehose/
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.process/#stagetenant

## 前提
- AWS firehoseのHTTP Endpointにhttpのエンドポイントは設定できず、必ずHTTPSのエンドポイントを設定する必要がある。
- 2025/07時点では、**firehoseからVPC内のリソース(e.g. Loki on EKS)への接続はできない。回避策としてInternet facing ALBを作成して、Firehose → ALB → Lokiの構成を取る必要がある。あと、FirehoseのSource IPは（ほぼ）固定されていないため、セキュリティグループで制御することはほぼ無理であり、ヘッダーなどで認証の仕組みを実現する必要がある**

## マルチテナントLokiに連携するための設定
### Firehose側の設定
- `Parameters`で、Keyに`lbl_<テナントID>`を、Valueに`<テナント名>`を設定する。