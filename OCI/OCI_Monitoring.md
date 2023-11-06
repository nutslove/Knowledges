## サーバ上のログをLoggingに連携する方法
- まず`動的グループ`を～～～
- `エージェント構成`～～～
- ログの保存期間はdefaultで30日で180日まで延長できる
- Fluentdのconfigファイルは`/etc/unified-monitoring-agent/unified-monitoring-agent.conf`
  - 実際の定義は上記ファイルからincludeされる`/etc/unified-monitoring-agent/conf.d/fluentd_config/fluentd.conf`ファイル
  - **Loggingの`エージェント構成`の設定で反映されるものであり、Configファイルの手動変更は(検証結果)反映されないっぽい**
- **`エージェント構成`で指定するファイル名に正規表現は使えないっぽい**
  - 例えば`/var/log/*`はできるけど`/var/log/*.log`はできなさそう
- Oracle Linuxの場合は`Oracle Cloudエージェント`タブの`カスタム・ログのモニタリング`を有効化するだけで使える

## ログに対してアラートを設定する方法
- ログに対するアラートは`Connector Hub`で`Monitoring`にメトリクスとして連携して`Monitoring`で、そのメトリクスに対してアラートを設定する必要がある
  - https://docs.oracle.com/ja-jp/iaas/Content/connector-hub/alarmlogs.htm
  - **`Connector Hub`でメトリクスに連携する`ディメンション`に`message`は入れないこと！**
    - `message`が異なるとすべて別のメトリクスとして扱われ、`message`がTimestampなどで少しでも異なるとSumで合算されない
- フィルタの`=`と`!=`は部分一致
- フィルタで正規表現は使えないっぽい
  - `*`などワイルドカードは使えるみたい
    - https://docs.oracle.com/ja-jp/iaas/Content/connector-hub/queryreference.htm
    - https://docs.oracle.com/ja-jp/iaas/Content/Logging/Reference/query_language_specification.htm
- フィルタで大文字・小文字は区別されない
- ログの中身が入るのは`data.message`フィールド
- ログファイルパスが入るのは`data.tailed_path`フィールド
- **ログがLoggingに連携(表示)されるまで数分(3分程度？)のラグがある**
  - なので`Connector Hub`で`Monitoring`に送られるのも数分後になる
  - これはFluentdの`flush_interval`によるもののように見える（180sに設定されている）
    - `flush_interval`の値を手動で変更し、`unified-monitoring-agent`を再起動しても180sのままに見える

## Management Agent(管理エージェント)でExporter(Prometheus形式メトリクス)をOCI Monitoringに送る方法
- https://oracle-japan.github.io/ocitutorials/management/monitoring_prometheus/
- 複数のExporterの収集の設定もできる！
  - https://qiita.com/NICAREGI/items/f7070cb398ccac84bf2a
- Management Agentプロセスは`systemctl status oracle-cloud-agent`で確認できる
  - うまく設定が反映されない時は`systemctl restart oracle-cloud-agent`でプロセスを再起動してみること
- debugは`/var/lib/oracle-cloud-agent/plugins/oci-managementagent/polaris/agent_inst/log/mgmt_agent.log`ログから可能(`PrometheusEmitter`で検索)
- Management Agent(管理エージェント)が動いてるインスタンスが含まれるように動的グループを作成し、その動的グループに対してポリシーの作成が必要
  ~~~
  allow service loganalytics to read loganalytics-features-family in tenancy
  allow dynamic-group ＜動的グループ名＞ to use metrics in tenancy / compartment <Management Agentが属してるコンパートメント名>
  ~~~

## その他
- **`検索`で表示できるのはログ件数は500件がMax**
  - それ以上ロードされず、表示時間間隔を短くするかフィルタで対象のログを絞って500件以下にするしかなさそう
- __MQL(Monitoring Query Language)__
  - OCI Monitoring Metrics用のクエリー言語
  - https://docs.oracle.com/ja-jp/iaas/Content/Monitoring/Reference/mql.htm