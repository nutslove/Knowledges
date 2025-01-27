## GrafanaのアラートのPagerDutyへの連携
- 対象Serviceの「Integrations」タブで「＋ Add an Integration」を押下  
  ![](../PagerDuty/image/pagerduty_integration_grafana_1.png)
- Webhookを検索し、選択後、「Add」を押下  
  ![](../PagerDuty/image/pagerduty_integration_grafana_2.png)
- 作成されたWebhookから「Integration Key」を押さえておく  
  ![](../PagerDuty/image/pagerduty_integration_grafana_3.png)
- Grafanaの「Contact Points」でintegrationとしてPagerDutyを選択し、「Integration Key」に上で確認した値を入力して保存する  
  ![](../PagerDuty/image/pagerduty_integration_grafana_4.png)