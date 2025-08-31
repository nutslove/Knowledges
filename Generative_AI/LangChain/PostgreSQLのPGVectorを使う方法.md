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
> [!WARNING]  
> https://github.com/langchain-ai/langchain-postgres  
> `langchain-postgres` v0.0.14から `PGVector`は非推奨になった。代わりに **`PGVectorStore`** を使うこと。
> - `PGVectorStore`
>   - https://python.langchain.com/docs/integrations/vectorstores/pgvectorstore/

> [!CAUTION]  
> `ainit_vectorstore_table`、`init_vectorstore_table`メソッドはDB Tableを新規作成するので、最初の１回だけ実行すること
> - 既存のTableを使う設定も可能
>   - https://python.langchain.com/docs/integrations/vectorstores/pgvectorstore/#create-a-vector-store-using-existing-table

> [!NOTE]  
> `ainit_vectorstore_table`、`init_vectorstore_table`メソッドで作成されるデフォルトのテーブル構造は以下の通り  
> ※`vector`の次元はメソッドの`vector_size`パラメータで指定可能  
> ```
>                  Table "public.documents"
>       Column       |     Type     | Collation | Nullable | Default
> --------------------+--------------+-----------+----------+---------
>  langchain_id       | uuid         |           | not null |
>  content            | text         |           | not null |
>  embedding          | vector(3072) |           | not null |
>  langchain_metadata | json         |           |          |
> Indexes:
>     "documents_pkey" PRIMARY KEY, btree (langchain_id)
> ```
> - 対象コード  
>   - https://github.com/langchain-ai/langchain-postgres/blob/main/langchain_postgres/v2/engine.py#L148


- 同期・非同期によって若干設定が異なる
### 非同期
- `sqlalchemy`、`langchain-postgres`ライブラリーをインストール
- `sqlalchemy`の`create_async_engine`（`from sqlalchemy.ext.asyncio import create_async_engine`）でDBコネクションを初期化し、`langchain_postgres`の`PGEngine`の`from_engine`メソッドでDB Engineインスタンスを初期化し、`langchain_postgres`の`PGVectorStore`の`create`メソッドでVectorStoreを初期化し`aadd_documents`メソッドでEmbeddingする
- コード例  
  ```python
  from langchain_google_genai import GoogleGenerativeAIEmbeddings
  from langchain_postgres import PGVectorStore,PGEngine
  from sqlalchemy.ext.asyncio import create_async_engine
  from langchain_core.documents import Document
  import os
  import uuid
  import asyncio

  os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/home/nutslove/GCP_VertexAI/service-account-key.json"

  connection = "postgresql+psycopg://postgres:postgres@localhost:5432/postgres" # postgresql+psycopg://ユーザー名:パスワード@ホスト:ポート/データベース名

  async def main():
    engine = create_async_engine(
      connection,
    )
    pg_engine = PGEngine.from_engine(engine=engine)

    TABLE_NAME = "documents"
    VECTOR_SIZE = 3072

    embedding = GoogleGenerativeAIEmbeddings(model="gemini-embedding-001")

    store = await PGVectorStore.create(
      engine=pg_engine,
      table_name=TABLE_NAME,
      embedding_service=embedding,
    )

    docs = [
      Document(page_content="Apples and oranges"),
      Document(page_content="Cars and airplanes"),
      Document(page_content="Train")
    ]

    await store.aadd_documents(docs)

    query = "I'd like a fruit."
    results = await store.asimilarity_search(
      query,
      k=1
    )
    print(results)

  if __name__ == "__main__":
    asyncio.run(main())
  ```

### 同期
- `langchain-postgres`ライブラリーをインストール
- コード例  
  ```python
  from langchain_core.documents import Document
  from langchain_core.embeddings import DeterministicFakeEmbedding
  from langchain_postgres import PGEngine, PGVectorStore

  # Replace the connection string with your own Postgres connection string
  CONNECTION_STRING = "postgresql+psycopg3://langchain:langchain@localhost:6024/langchain"
  engine = PGEngine.from_connection_string(url=CONNECTION_STRING)

  # Replace the vector size with your own vector size
  VECTOR_SIZE = 768
  embedding = DeterministicFakeEmbedding(size=VECTOR_SIZE)

  TABLE_NAME = "my_doc_collection"

  engine.init_vectorstore_table(
      table_name=TABLE_NAME,
      vector_size=VECTOR_SIZE,
  )

  store = PGVectorStore.create_sync(
      engine=engine,
      table_name=TABLE_NAME,
      embedding_service=embedding,
  )

  docs = [
      Document(page_content="Apples and oranges"),
      Document(page_content="Cars and airplanes"),
      Document(page_content="Train")
  ]

  store.add_documents(docs)

  query = "I'd like a fruit."
  docs = store.similarity_search(query)
  print(docs)
  ```