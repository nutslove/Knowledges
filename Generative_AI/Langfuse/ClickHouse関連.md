## DB操作
- (ClickHouse Pod内で) `clickhouse-client`コマンドを使ってDBに接続する
- DB一覧確認  
  ```sql
  SHOW DATABASES;
  ```
- Tables一覧確認  
  ```sql
  SHOW TABLES;
  ```

## Langfuse トレース取り込み処理の改善による ClickHouse の負荷低減
- https://tech.layerx.co.jp/entry/langfuse-clickhouse-read-skip
  - 最初からLangfuseのv3を使用している場合、`LANGFUSE_SKIP_INGESTION_CLICKHOUSE_READ_MIN_PROJECT_CREATE_DATE`環境変数に過去の年月日を`YYYY-MM-DD`形式で設定することで、ClickHouseへの負荷を低減できるという内容