# Langfuse Webで "no migration found for version xx" が出て、Langfuse Web（Pod）が起動しない	
- エラーログ全文  
  ```
  error: no migration found for version 23: read down for version 23 .: file does not existerror: no migration found for version 23: read down for version 23 .: file does not exist"
  ```
### 原因
- 不明・・・。

### 対処方法
- 不明・・・。一旦Langfuse一式を削除（RDSとS3は残したまま）して、再度デプロイしたら問題なく起動した。おそらく最初にHelm CLIでデプロイした後にArgoCDから更新したせいで不整合が発生したかも・・・？