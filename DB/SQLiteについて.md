## 各種コマンド
- テーブル一覧確認  
  ```shell
  SELECT name FROM sqlite_master WHERE type='table';
  ```
- テーブルのフィールド情報確認
  ```shell
  .schema <テーブル名>
  ```
  - テーブル名省略の場合、すべてのテーブルのフィールド情報を出力