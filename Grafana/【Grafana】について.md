## DashBoard / Panel設定
#### ・Tooltipについて  
グラフにマウスをかざした時に表示する対象についての設定  
![Tooltip](https://github.com/nutslove/all_I_need/blob/master/Knowledges/Grafana/image/Tooltip.jpg)
- Single  
  → 1つだけ表示
![Tooltip_Single](https://github.com/nutslove/all_I_need/blob/master/Knowledges/Grafana/image/Tooltip_Single.jpg)
- All  
  → すべて表示
![Tooltip_All](https://github.com/nutslove/all_I_need/blob/master/Knowledges/Grafana/image/Tooltip_All.jpg)
- Hidden  
  → 表示しない
![Tooltip_Hidden](https://github.com/nutslove/all_I_need/blob/master/Knowledges/Grafana/image/Tooltip_Hidden.jpg)

## Alert設定
#### ・AlertのGroupingについて
https://grafana.com/docs/grafana/latest/alerting/notifications/  
GrafanaのAlertもPrometheusと同様にラベルによるGroupingを行う  
- `Group by`に何も設定しない場合はすべてのAlertが1つのGroupとして扱う
- `Group by`に`...`を設定するとすべてのlabelによってGroupingされるので、すべてのアラートが個別のGroupとして扱われる  
  (1つのalertnameに属していてもpodやhostname等が違えば別のGroupとして扱われる)
  >Note: Grafana also has a special label named ... that you can use to group all alerts by all labels (effectively disabling grouping), 
therefore each alert will go into its own group. It is different from the default of group_by: null where all alerts go into a single group.  
- 複数のNotification policiesが存在する場合、各Policy側で`Group by`設定を`...`に上書きすること  
  ![Notification_policies](https://github.com/nutslove/all_I_need/blob/master/Knowledges/Grafana/image/NotificationPolicies.jpg)

#### ・GrafanaのAlertに関するコンポーネントについて
https://grafana.com/docs/grafana/next/alerting/high-availability/
- Grafana Alerting systemは内部的に`Scheduler`と`Alertmanager`を持っている
> The Grafana Alerting system has two main components: a Scheduler and an internal Alertmanager. The Scheduler evaluates your alert rules, while the internal Alertmanager manages routing and grouping.
- Scheduler  
  → alert rulesを評価する
- Alertmanager  
  → Alertのroutingとgroupingを行う

#### ・GrafanaのAlert発行単位について
https://grafana.com/docs/grafana/latest/alerting/fundamentals/alert-rules/alert-instances/
- Grafanaは1つのAlert Ruleで複数のインスタンすを作成できる
- つまり、1つのAlert Ruleから複数のアラートが発行される（上記URL参照）
- 1つのAlert Ruleから発行される複数のアラートは1回の処理で連携される  
  例えばWebhookに連携する場合、以下添付のように1つのAlert Ruleから同時に発行される複数のAlertは1回のWebhook(POST)で連携される  
  (jsonの中に含まれている)
![Alert](https://github.com/nutslove/all_I_need/blob/master/Knowledges/Grafana/image/Grafana_MultipleAlerts.jpg)