## サーバ上のログをLoggingに連携する方法
- まず`動的グループ`を～～～
- `エージェント構成`～～～
- ログの保存期間はdefaultで30日で180日まで延長できる
- Fluentdのconfigファイルは`/etc/unified-monitoring-agent/unified-monitoring-agent.conf`
  - 実際の定義は上記ファイルからincludeされる`/etc/unified-monitoring-agent/conf.d/fluentd_config/fluentd.conf`ファイル
  - **Loggingの`エージェント構成`の設定で反映されるものであり、Configファイルの手動変更は(検証結果)反映されないっぽい**
- Oracle Linuxの場合はPluginに

## ログに対してアラートを設定する方法
- ログに対するアラートは`Connector Hub`で`Monitoring`にメトリクスとして連携して`Monitoring`で、そのメトリクスに対してアラートを設定する必要がある
  - https://docs.oracle.com/ja-jp/iaas/Content/connector-hub/alarmlogs.htm
  - **`Connector Hub`でメトリクスに連携する`ディメンション`に`message`は入れないこと！**
    - `message`が異なるとすべて別のメトリクスとして扱われ、`message`がTimestampなどで少しでも異なるとSumで合算されない
- フィルタの`=`と`!=`は部分一致
- フィルタで正規表現は使えないっぽい（**もう少しやって(調べて)みること**）
- フィルタで大文字・小文字は区別されない
- ログの中身が入るのは`data.message`フィールド
- ログファイルパスが入るのは`data.tailed_path`フィールド
- **ログがLoggingに連携(表示)されるまで数分(3分程度？)のラグがある**
  - なので`Connector Hub`で`Monitoring`に送られるのも数分後になる
  - これはFluentdの`flush_interval`によるもののように見える（180sに設定されている）
    - `flush_interval`の値を手動で変更し、`unified-monitoring-agent`を再起動しても180sのままに見える

## その他
- **`検索`で表示できるのはログ件数は500件がMax**
  - それ以上ロードされず、表示時間間隔を短くするかフィルタで対象のログを絞って500件以下にするしかなさそう
