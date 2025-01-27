## WebHook設定
- Incident（アラート）が発生したときに外部にWebHookを設定することができる
### 手順
- 「Integrations」タブの「Generic Webhooks（v3）」をクリック  
  ![](./image/webhook_1.png)
- 「＋ New Webhook」をクリック  
  ![](./image/webhook_2.png)
- 「Webhook URL」にPOST先のURLを入力し、「Scope Type」(e.g. Service)と「SCOPE」を選択し、「Event Subscription」でどういうときにWebhookを投げるかを設定し、「Add Webhook」を押下  
  ![](./image/webhook_3.png)  
  ![](./image/webhook_4.png)
- その後、Webhookで関連付けたServiceの「Integrations」タブの「Webhooks」ブロックに設定したWebhookが表示されていることを確認  
  ![](./image/webhook_5.png)

#### 作成後のテスト
- 作成したWebhookは「Test」ブロックの「Send Test Event」でテストのWebhookを飛ばすことができる

## GrafanaのアラートのPagerDutyへの連携
- 対象Serviceの「Integrations」タブで「＋ Add an Integration」を押下  
  ![](./image/pagerduty_integration_grafana_1.png)
- Webhookを検索し、選択後、「Add」を押下  
  ![](./image/pagerduty_integration_grafana_2.png)
- 作成されたWebhookから「Integration Key」を押さえておく  
  ![](./image/pagerduty_integration_grafana_3.png)
- Grafanaの「Contact Points」でintegrationとしてPagerDutyを選択し、「Integration Key」に上で確認した値を入力して保存する  
  ![](./image/pagerduty_integration_grafana_4.png)