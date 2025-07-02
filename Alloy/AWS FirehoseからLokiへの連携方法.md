- 参考URL
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.awsfirehose/
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.process/#stagetenant

## 前提
- AWS firehoseのHTTP Endpointにhttpのエンドポイントは設定できず、必ずHTTPSのエンドポイントを設定する必要がある。
- 2025/07時点では、**firehoseからVPC内のリソース(e.g. Loki on EKS)への接続はできない。回避策としてInternet facing ALBを作成して、Firehose → ALB → Lokiの構成を取る必要がある。あと、FirehoseのSource IPは（ほぼ）固定されていないため、セキュリティグループで制御することはほぼ無理であり、ヘッダーなどで認証の仕組みを実現する必要がある**

## マルチテナントLokiに連携するための設定
### Firehose側の設定
- `Parameters`で、Keyに`lbl_<テナントID>`を、Valueに`<テナント名>`を設定する。
  - Alloy側で`lbl_`が除去されてLabelとして設定される。
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.awsfirehose/
    > You can use the `X-Amz-Firehose-Common-Attributes` header to set extra static labels. You can configure the header in the Parameters section of the Data Firehose delivery stream configuration. Label names must be prefixed with `lbl_`. The prefix is removed before the label is stored in the log entry. Label names and label values must be compatible with the [Prometheus data model](https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels) specification.
- HTTPエンドポイントは **`https://<ALBのDNS名>/awsfirehose/api/v1/push`**

### Alloy側の設定
- `loki.source.awsfirehose`を追加する  
  ```yaml
  loki.source.awsfirehose "loki_firehose_receiver" {
      http {
          listen_address = "0.0.0.0"
          listen_port = 9999
      }
      forward_to = [
          loki.process.set_tenant.receiver,
      ]
  }
  ```
  - helmの場合、`alloy.extraPorts`で追加したPortを追加する  
    ```yaml
    alloy:
      extraPorts:
        - name: fh-receiver
          port: 9999
          targetPort: 9999
          protocol: TCP
    ```
- `loki.process.set_tenant`を追加する  
  ```yaml
  loki.process "set_tenant" {
    forward_to = [loki.write.dynamic_tenant_loki.receiver]
    stage.tenant {
      label = "tenant_id"
    }
  }
  ```
- `loki.write.dynamic_tenant_loki`を追加する  
  ```yaml
  loki.write "dynamic_tenant_loki" {
    loki_url = "https://<Lokiのエンドポイント>/loki/api/v1/push"
  }
  ```
  - **`tenant_id`は自動的に設定されるので、明示的な設定は不要**