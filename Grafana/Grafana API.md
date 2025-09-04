#### ■ 特定のアラートルールの詳細を取得する
```shell
curl -H "Authorization: Bearer <YOUR_API_TOKEN>" \
     -H "Content-Type: application/json" \
     "http://<your-grafana-url>/api/v1/provisioning/alert-rules/<ALERT_RULE_UID>"
```
- Alert Rule UIDは、アラートルールの一意の識別子で、Grafanaからのアラート情報（`generatorURL`）に含まれている
  - 例（`fex0rwuqjy03kf`の部分がアラートのUID）: `"generatorURL": "https://grafana.example.com/alerting/grafana/fex0rwuqjy03kf/view?orgId=1",`

---