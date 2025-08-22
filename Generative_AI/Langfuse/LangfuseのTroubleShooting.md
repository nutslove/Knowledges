- https://langfuse.com/self-hosting/troubleshooting#clickhouse-handling-failed-migrations

# Langfuse Webで "no migration found for version xx" が出て、Langfuse Web（Pod）が起動しない	
- エラーログ全文  
  ```
  error: no migration found for version 23: read down for version 23 .: file does not existerror: no migration found for version 23: read down for version 23 .: file does not exist"
  ```
### 原因
- 不明・・・。

### 対処方法
- 不明・・・。一旦Langfuse一式を削除（RDSとS3は残したまま）して、再度デプロイしたら問題なく起動した。おそらく最初にHelm CLIでデプロイした後にArgoCDから更新したせいで不整合が発生したかも・・・？

---

# Langfuse Webで "error: Dirty database version xx. Fix and force version." が出て、Langfuse Web（Pod）が起動しない

### 原因
- https://github.com/langfuse/langfuse/issues/6679

### 対処方法
- RDSのLangfuse用のDBを削除し、ClickHouseとZookeeper用のEFS（PV）もAccessPointを作り直して、すべてクリアした状態でデプロイしたら問題なく起動した。(既存のデータはすべて消えた)