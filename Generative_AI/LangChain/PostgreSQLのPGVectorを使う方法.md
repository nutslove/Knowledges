- https://python.langchain.com/docs/integrations/vectorstores/pgvector/
- データ投入時metadataを与えて、Retriever時にmetadataでフィルタリングすることができる

## 設定手順
- まず、PostgreSQLに接続する
  - postgresql Pod/Dockerの場合、Pod/Dockerコンテナに入って `psql -U postgres -d postgres`でアクセス
- 以下のコマンドでPGVector拡張機能を作成  
  ```shell
  CREATE EXTENSION IF NOT EXISTS vector;
  ```
- PGVectorのバージョンを確認  
  ```shell
  \dx
  ```
  もしくは  
  ```shell
  SELECT extversion FROM pg_extension WHERE extname='vector';
  ```
  - PostgreSQL自体のバージョンは `select * From version();`で確認可能
- PostgreSQL内の設定はこれで完了

## 利用方法
- ``

> [!CAUTION]  
> https://github.com/langchain-ai/langchain-postgres
> `langchain-postgres` v0.0.14から `PGVector`は非推奨になった。代わりに **`PGVectorStore`** を使うこと。
> - `PGVectorStore`
>   - https://python.langchain.com/docs/integrations/vectorstores/pgvectorstore/