## Grafana本体の設定
- 基本的な設定ファイルは`/etc/grafana/grafana.ini`
- `/etc/grafana/grafana.ini`に設定できるのはすべて環境変数で設定できる
  - フォーマット
    ~~~
    GF_<SectionName>_<KeyName>
    ~~~
  - 例
    - `/etc/grafana/grafana.ini`
      ~~~
      # default section
      instance_name = ${HOSTNAME}

      [security]
      admin_user = admin

      [auth.google]
      client_secret = 0ldS3cretKey

      [plugin.grafana-image-renderer]
      rendering_ignore_https_errors = true
      ~~~
    - 環境変数
      ~~~
      export GF_DEFAULT_INSTANCE_NAME=my-instance
      export GF_SECURITY_ADMIN_USER=owner
      export GF_AUTH_GOOGLE_CLIENT_SECRET=newS3cretKey
      export GF_PLUGIN_GRAFANA_IMAGE_RENDERER_RENDERING_IGNORE_HTTPS_ERRORS=true
      ~~~

## DashBoard / Panel設定
#### ■ Tooltipについて  
- グラフにマウスをかざした時に表示する対象についての設定  
![Tooltip](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/Tooltip.jpg)
  - Single  
    → 1つだけ表示
    ![Tooltip_Single](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/Tooltip_Single.jpg)
  - All  
    → すべて表示
    ![Tooltip_All](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/Tooltip_All.jpg)
  - Hidden  
    → 表示しない
    ![Tooltip_Hidden](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/Tooltip_Hidden.jpg)

### Graph(old) Panel
- Grafana9.0からGraph(old)がPanelから選択できないようになった
  - 既存のGraph(old)Panelはそのまま使い続けられる
- ただ一旦Timeseries Panelを作成後、jsonからtypeをtimeseries → graphに変更することでGraph(old) Panelを作成できる
  - https://stackoverflow.com/questions/73353757/grafana-get-graph-old-back 

## Alert設定
#### ■ AlertのGroupingについて
- 参考URL
  - https://grafana.com/docs/grafana/latest/alerting/notifications/  
- GrafanaのAlertもPrometheusと同様にラベルによるGroupingを行う  
- `Group by`に何も設定しない場合はすべてのAlertが1つのGroupとして扱う
- `Group by`に`...`を設定するとすべてのlabelによってGroupingされるので、すべてのアラートが個別のGroupとして扱われる  
  (1つのalertnameに属していてもpodやhostname等が違えば別のGroupとして扱われる)
  >Note: Grafana also has a special label named ... that you can use to group all alerts by all labels (effectively disabling grouping), 
therefore each alert will go into its own group. It is different from the default of group_by: null where all alerts go into a single group.  
- 複数のNotification policiesが存在する場合、各Policy側で`Group by`設定を`...`に上書きすること  
  ![Notification_policies](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/NotificationPolicies.jpg)

#### ■ GrafanaのAlertに関するコンポーネントについて
- 参考URL
  - https://grafana.com/docs/grafana/next/alerting/high-availability/
- Grafana Alerting systemは内部的に`Scheduler`と`Alertmanager`を持っている
  > The Grafana Alerting system has two main components: a Scheduler and an internal Alertmanager. The Scheduler evaluates your alert rules, while the internal Alertmanager manages routing and grouping.
- Scheduler  
  → alert rulesを評価する
- Alertmanager  
  → Alertのroutingとgroupingを行う

#### ■ Grafana Alerting(Webhook)から発行されるアラートデータ形式
- Alertmanager形式と同じ
  - https://prometheus.io/docs/alerting/latest/clients/
  - https://prometheus.io/docs/alerting/latest/notifications/

#### ■ GrafanaのAlert発行単位について
- 参考URL
  - https://grafana.com/docs/grafana/latest/alerting/fundamentals/alert-rules/alert-instances/
- Grafanaは1つのAlert Ruleで複数のインスタンすを作成できる
- つまり、1つのAlert Ruleから複数のアラートが発行される（上記URL参照）
- 1つのAlert Ruleから発行される複数のアラートは1回の処理で連携される。  
  例えば、Webhookに連携する場合、以下添付のように1つのAlert Ruleから同時に発行される62個のAlertは1回のWebhook(POST)で連携される  
![Alert](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/Grafana_MultipleAlerts.jpg)

#### ■ CloudWatch LogsのAlert設定
- CloudWatch Logsに対してアラートを設定するためにはCloudwatch Logs Insightsを使ってnumericデータが返ってくるようにクエリーを投げる必要がある
  > Alerting require queries that return numeric data, which CloudWatch Logs support. For example through the use of the stats command, alerts are supported.

  > **Warning**  
  > When trying to alert on a query, if an error like `input data must be a wide series but got ...` is received, make sure that your query returns valid numeric data that can be printed in a Time series panel.
- `fileds`で`@message`を指定し、`filter`で`like`または`not like`や`=~`で検知したい文字列を特定し、`stats count(*) by bin(1m)`で件数を取得する
- `bin`はLogQL`count_over_time`の`[]`で指定する時間と同じ感覚で、件数をまとめる間隔を指定
  - 例えばあるログが12:00に2件、12:03に5件、12:08に3件あったとした場合、`bin(10m)`は`12:00 10件`と表示されるけど、`bin(1m)`は`12:00 2件`,`12:03 5件`,`12:08 3件`と表示される
  - 1mが無難
- Cloudwatch Logs Insightsクエリー例
  - 条件が1つのみの場合
    ```
    fields @message
    | filter @message like /error/
    | stats count(*) by bin(1m)
    ```
  - 条件が複数の場合は`and`や`or`でつなげることができる
    ```
    fields @message
    | filter @message like /error/ and @message not like /exception/
    | stats count(*) by bin(1m)
    ```
- `fields`で指定できる項目はマネコンのCloudwatch Logs Insightsから確認できる
![CloudWatch_Logs_Insights_fields](https://github.com/nutslove/Knowledges/blob/main/Grafana/image/CloudWatch_Logs_Insights_fields.jpg)
- 参考URL
  - https://grafana.com/docs/grafana/latest/datasources/aws-cloudwatch/
  - https://docs.aws.amazon.com/ja_jp/AmazonCloudWatch/latest/logs/CWL_QuerySyntax.html
  - https://qiita.com/suuu/items/8387df88f134348f22c7 

## Plugin
#### ■ X-Ray
- 参考URL
  - https://github.com/grafana/x-ray-datasource
  - https://grafana.com/grafana/plugins/grafana-x-ray-datasource/
- X-Rayプラグイン利用に必要なIAMポリシーに`ec2:DescribeRegions`も含まれる
  - be used to list regions
  - https://github.com/grafana/x-ray-datasource/issues/135

#### ■ CloudWatch
- Grafanaから表示できるCloudWatch LogsのLog Groupsの数は50個までだが、これはGrafanaの制約ではなく、AWS側の制約である
  - AWSの`DescribeLogGroups`APIの制約
  - https://github.com/grafana/grafana/issues/50532

## その他
- GrafanaもGrafanaに関するメトリクスを開示している  
  - https://grafana.com/docs/grafana/latest/alerting/fundamentals/evaluate-grafana-alerts/#metrics-from-the-alerting-engine
  - https://grafana.com/docs/grafana/latest/alerting/images-in-notifications/#metrics