## Grafana 本体の設定

- 基本的な設定ファイルは`/etc/grafana/grafana.ini`
- `/etc/grafana/grafana.ini`に設定できるのはすべて環境変数で設定できる

  - フォーマット
    ```
    GF_<SectionName>_<KeyName>
    ```
  - 例

    - `/etc/grafana/grafana.ini`

      ```
      # default section
      instance_name = ${HOSTNAME}

      [security]
      admin_user = admin

      [auth.google]
      client_secret = 0ldS3cretKey

      [plugin.grafana-image-renderer]
      rendering_ignore_https_errors = true
      ```

    - 環境変数
      ```
      export GF_DEFAULT_INSTANCE_NAME=my-instance
      export GF_SECURITY_ADMIN_USER=owner
      export GF_AUTH_GOOGLE_CLIENT_SECRET=newS3cretKey
      export GF_PLUGIN_GRAFANA_IMAGE_RENDERER_RENDERING_IGNORE_HTTPS_ERRORS=true
      ```

- **デフォルトでは Viewer 権限ユーザは Explore を操作できない(表示されない)が、`users`セクションの`viewers_can_edit`を`true`(default は false)にすれば Viewer 権限ユーザでも Explore を操作できるようになる**
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#viewers_can_edit

> [!CAUTION]
>
> - 2026/01 このオプションは非推奨(deprecated)になった
> - https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#viewers_can_edit
>   > This option is deprecated - assign your viewers as editors, if you are using RBAC assign the data sources explorer role to your users.

- `server`セクションの`serve_from_sub_path`を`true`したら`root_url`に指定した URL の subpath で Grafana にアクセスすることもできる
  - 例えば`root_url`を`http://<GrafanaのIP>/unk`にしたら`<GrafanaのIP>/unk`でアクセスできるようになる
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#serve_from_sub_path

## Grafana built-in variables

- https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#global-variables

#### `$__interval`

- Dashboard の TimeRange と panel のグラフの横幅に合わせて自動的に最適な interval を設定してくれる
  - > The $\_\_interval is calculated using the time range and the width of the graph (the number of pixels).

#### `$__rate_interval`

- DataSource が Prometheus の時のみ使える変数
- rate 関数の interval を TimeRange と panel のグラフの横幅に合わせて自動的に設定してくれる

#### `$__range`

- DataSource が Prometheus と Loki の時のみ使える変数
- 現在の Dashboard の Range(恐らく Time Range?)
  - > It is calculated by `to - from`

## DashBoard / Panel 設定

### ■ Variables について

- Grafana ダッシュボードの中の値(Query、Min interval、Panel タイトルなどなど)をダッシュボードで選択/記入した値に動的に変更できる仕組み
- 事前にダッシュボードの「Settings」の「Variables」で定義しておく必要がある
  - variables にはいくつかの Type があり、Variables の値に動的に Prometheus などのクエリーの結果を使ったり(`Query`type)、Variables の値を静的に設定したり(`Custom`type)することができる
- 参考 URL
  - https://grafana.com/docs/grafana/latest/dashboards/variables/
  - https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/
  - https://grafana.com/docs/grafana/latest/dashboards/variables/inspect-variable/
  - https://grafana.com/docs/grafana/latest/dashboards/variables/variable-syntax/

### ■ Tooltip について

- グラフにマウスをかざした時に表示する対象についての設定  
  ![Tooltip](image/Tooltip.jpg)
  - Single  
    → 1 つだけ表示
    ![Tooltip_Single](image/Tooltip_Single.jpg)
  - All  
    → すべて表示
    ![Tooltip_All](image/Tooltip_All.jpg)
  - Hidden  
    → 表示しない
    ![Tooltip_Hidden](image/Tooltip_Hidden.jpg)

### ■ Graph styles の Stack series

- Bar/グラフ/Points で各値を重複して見せるか、重ねて見せるかの設定
  - Off  
    → 小さい値は大きい値の Bar/グラフ中に表示される  
    ![OFF](image/stack_series_off.jpg)
  - Normal  
    → 各値が別の値の上に(から)表示される  
    ![Normal](image/stack_series_normal.jpg)
  - 100%  
    → 全体を 100%にして各値の割合で表示される  
    ![100per](image/stack_series_100per.jpg)
- Grafana 関連ページ
  - https://grafana.com/docs/grafana/latest/panels-visualizations/visualizations/time-series/#stack-series

#### `$__interval`、`$__rate_interval`、`$__range`について

- `$__interval`
  > Grafana automatically calculates an interval that can be used to group by time in queries.
- `$__rate_interval`
  > We recommend using \$**rate_interval in the rate and increase functions instead of \$**interval or a fixed interval value. Because $\_\_rate_interval is always at least four times the value of the Scrape interval, it avoid problems specific to Prometheus.
- `$__range`
  - Dashboard の TimeRange の時間範囲が適用される
  - 例えば Dashboard の TimeRange を"Last 30 minutes"にしてる場合、`[$__range]`は`[30m]`になる
- 参考 URL
  - https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#global-variables
  - https://grafana.com/docs/grafana/latest/datasources/prometheus/template-variables/#use-__rate_interval

### ■ Min interval について

- データの集計間隔(rollup)
- 例えば Min interval を 1h に設定したらデータは 12h、13h、14h、・・・と 1 時間間隔で表示される
  ![min_interval_1h](image/min_interval_1h.jpg)
  ![min_interval_2h](image/min_interval_2h.jpg)

### ■ Relative time について

- Panel ごとの TimeRange（Panel ごとに設定できる）
- Relative time が設定されている Panel は Dashboard の TimeRange に影響されなくなる
  ![relative_time1](image/Relative_time1.jpg)
  ![relative_time2](image/Relative_time2.jpg)

### ■ Graph(old) Panel

- Grafana9.0 から Graph(old)が Panel から選択できないようになった
  - 既存の Graph(old)Panel はそのまま使い続けられる
- ただ一旦 Timeseries Panel を作成後、json から type を timeseries → graph に変更することで Graph(old) Panel を作成できる
  - https://stackoverflow.com/questions/73353757/grafana-get-graph-old-back

### ■ No Data を 0 に変える方法

- Grafana の機能で No Value を他の値に変える
  ![NoData1](image/NoData_To_0.jpg)
- (Prometheus)クエリーの最後に`or vector(0)`をつけて No Data の時は 0 を表示させる
  - e.g. `http_requests_total{method="GET", code="400"} or vector(0)`
- 参考 URL
  - https://community.grafana.com/t/how-to-get-no-data-to-0-when-there-is-no-data-value/64845/6

## Alert 設定

#### ■ Alert の Grouping について

- 参考 URL
  - https://grafana.com/docs/grafana/latest/alerting/notifications/
- Grafana の Alert も Prometheus と同様にラベルによる Grouping を行う
- `Group by`に何も設定しない場合はすべての Alert が 1 つの Group として扱う
- `Group by`に`...`を設定するとすべての label によって Grouping されるので、すべてのアラートが個別の Group として扱われる  
   (1 つの alertname に属していても pod や hostname 等が違えば別の Group として扱われる)
  > Note: Grafana also has a special label named ... that you can use to group all alerts by all labels (effectively disabling grouping),
  > therefore each alert will go into its own group. It is different from the default of group_by: null where all alerts go into a single group.
- 複数の Notification policies が存在する場合、各 Policy 側で`Group by`設定を`...`に上書きすること  
  ![Notification_policies](image/NotificationPolicies.jpg)

#### ■ Grafana の Alert に関するコンポーネントについて

- 参考 URL
  - https://grafana.com/docs/grafana/next/alerting/high-availability/
- Grafana Alerting system は内部的に`Scheduler`と`Alertmanager`を持っている
  > The Grafana Alerting system has two main components: a Scheduler and an internal Alertmanager. The Scheduler evaluates your alert rules, while the internal Alertmanager manages routing and grouping.
- Scheduler  
  → alert rules を評価する
- Alertmanager  
  → Alert の routing と grouping を行う

#### ■ Grafana Alerting(Webhook)から発行されるアラートデータ形式

- Alertmanager 形式と同じ
  - https://prometheus.io/docs/alerting/latest/clients/
  - https://prometheus.io/docs/alerting/latest/notifications/

#### ■ Alert のメッセージ文の中にクエリーの Value を埋め込む方法

- `Summary and annotaions`部分の本文にクエリーの Value(e.g. CPU 使用率)を埋め込むことができる
- `Classic condition`では使えない
  - `Threshold`や`Math`でアラートを設定すること
- 設定例
  - `DX Ping mean success rate is {{ $values.B.Value }}%`

#### ■ Grafana の Alert 発行単位について

- 参考 URL
  - https://grafana.com/docs/grafana/latest/alerting/fundamentals/alert-rules/alert-instances/
- Grafana は 1 つの Alert Rule で複数のインスタンすを作成できる
- つまり、1 つの Alert Rule から複数のアラートが発行される（上記 URL 参照）
- 1 つの Alert Rule から発行される複数のアラートは 1 回の処理で連携される。  
   例えば、Webhook に連携する場合、以下添付のように 1 つの Alert Rule から同時に発行される 62 個の Alert は 1 回の Webhook(POST)で連携される  
  ![Alert](image/Grafana_MultipleAlerts.jpg)

#### ■ Notification policies について

- どの nested policy にも match しなかったら大元の Default policy が適用される
- default では 1 つの nested policy に match すると次の nested policy は評価されないけど、`Continue matching subsequent sibling nodes`の enable にすればすべての nested policy が評価される
- `Mute Timings`タブでアラートを mute する時間帯を設定し、特定の nested policy と紐づけることができる

##### `Continue matching subsequent sibling nodes`について

> If the Continue matching subsequent sibling nodes option is enabled for a nested policy, then evaluation continues even after one or more matches. A parent policy’s configuration settings and contact point information govern the behavior of an alert that does not match any of the nested policies. A default policy governs any alert that does not match a nested policy.
>
> You can configure Grafana-managed notification policies as well as notification policies for an external Alertmanager data source.

- `Continue matching subsequent sibling nodes`を enable にすると 1 つのアラートに対して複数の通知を受け取ることができる
  > enable Continue matching subsequent sibling nodes to continue matching sibling policies even after the alert matched the current policy. When this option is enabled, you can get more than one notification for one alert.
- nested policy の順番で想定とは違う動きをすることもあり得るので注意！
  - 例えば、`Mute Timings`と紐づいている nested policy があるとして、その nested policy が下の方にあると上の policy で評価されて発砲されてしまい、mute が効かない
  - そういう場合は、既存の nested policy に mute の為のラベル追加と`Mute Timings`の紐づけを行う
- 参考 URL
  - https://grafana.com/docs/grafana/latest/alerting/alerting-rules/create-notification-policy/
  - https://grafana.com/docs/grafana/latest/alerting/fundamentals/notification-policies/

#### ■ Alert の Custom Labels にメトリクスなどのラベルを変数として追加する方法

- https://community.grafana.com/t/path-different-label-s-value-to-my-custom-label-at-custom-labels-section/77670
- フォーマット
  - `{{ index $labels "<対象ラベル名>" }}`

#### ■ CloudWatch Logs の Alert 設定

- CloudWatch Logs に対してアラートを設定するためには Cloudwatch Logs Insights を使って numeric データが返ってくるようにクエリーを投げる必要がある

  > Alerting require queries that return numeric data, which CloudWatch Logs support. For example through the use of the stats command, alerts are supported.

  > **Warning**  
  > When trying to alert on a query, if an error like `input data must be a wide series but got ...` is received, make sure that your query returns valid numeric data that can be printed in a Time series panel.

- `fileds`で`@message`を指定し、`filter`で`like`または`not like`や`=~`で検知したい文字列を特定し、`stats count(*) by bin(1m)`で件数を取得する
- `bin`は LogQL`count_over_time`の`[]`で指定する時間と同じ感覚で、件数をまとめる間隔を指定
  - 例えばあるログが 12:00 に 2 件、12:03 に 5 件、12:08 に 3 件あったとした場合、`bin(10m)`は`12:00 10件`と表示されるけど、`bin(1m)`は`12:00 2件`,`12:03 5件`,`12:08 3件`と表示される
  - 1m が無難
- Cloudwatch Logs Insights クエリー例
  - 条件が 1 つのみの場合
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
- `fields`で指定できる項目はマネコンの Cloudwatch Logs Insights から確認できる
  ![CloudWatch_Logs_Insights_fields](image/CloudWatch_Logs_Insights_fields.jpg)
- 参考 URL
  - https://grafana.com/docs/grafana/latest/datasources/aws-cloudwatch/
  - https://docs.aws.amazon.com/ja_jp/AmazonCloudWatch/latest/logs/CWL_QuerySyntax.html
  - https://qiita.com/suuu/items/8387df88f134348f22c7

#### ■ Alert の Evaluation(評価)で時々`fetching of query results exceeded max number of attempts`エラーが出る件

- Grafana の Alert で時々以下のようなエラーが出た
  ```
  Failed to evaluate queries and expressions: [plugin.downstreamError] failed to query data:
  fetching of query results exceeded max number of attempts
  ```
- `[unified_alerting]`ブロックで`evaluation_timeout`(default: `30s`)と`max_attempts`(default: `3`)で評価時の Timeout と評価を何回まで試すかを設定できる
  - **これで再発しないか確認しあとで Update する**
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#unified_alerting

## Plugin

#### ■ X-Ray

- 参考 URL
  - https://github.com/grafana/x-ray-datasource
  - https://grafana.com/grafana/plugins/grafana-x-ray-datasource/
- X-Ray プラグイン利用に必要な IAM ポリシーに`ec2:DescribeRegions`も含まれる
  - be used to list regions
  - https://github.com/grafana/x-ray-datasource/issues/135

#### ■ CloudWatch

- Grafana から表示できる CloudWatch Logs の Log Groups の数は 50 個までだが、これは Grafana の制約ではなく、AWS 側の制約である
  - AWS の`DescribeLogGroups`API の制約
  - https://github.com/grafana/grafana/issues/50532

## その他

- Grafana も Grafana に関するメトリクスを開示している
  - https://grafana.com/docs/grafana/latest/alerting/fundamentals/evaluate-grafana-alerts/#metrics-from-the-alerting-engine
  - https://grafana.com/docs/grafana/latest/alerting/images-in-notifications/#metrics
