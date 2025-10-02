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
- `sqlalchemy`の`create_async_engine`（`from sqlalchemy.ext.asyncio import create_async_engine`）でDBコネクションを初期化し、`langchain_postgres`の`PGEngine`の`from_engine`メソッドでDB Engineインスタンスを初期化し、`langchain_postgres`の`PGVectorStore`の`create`メソッドでVectorStoreインスタンスを初期化し`aadd_documents`メソッドでEmbeddingする
- retrieveするときはVectorStoreインスタンスの`asimilarity_search`を使う
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

    await pg_engine.ainit_vectorstore_table(
      table_name=TABLE_NAME,
      vector_size=VECTOR_SIZE,
    )

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
- `langchain_postgres`の`PGEngine`の`from_connection_string`メソッドでDB Engineインスタンスを初期化し、`PGVectorStore`の`create_sync`メソッドでVectorStoreインスタンスを初期化し`add_documents`メソッドでEmbeddingする
- retrieveするときはVectorStoreインスタンスの`similarity_search`を使う
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

### metadataを使った保存・フィルタリング
#### 保存
```python
import uuid

from langchain_core.documents import Document

docs = [
    Document(
        id=str(uuid.uuid4()),
        page_content="Red Apple",
        metadata={"description": "red", "content": "1", "category": "fruit"},
    ),
    Document(
        id=str(uuid.uuid4()),
        page_content="Banana Cavendish",
        metadata={"description": "yellow", "content": "2", "category": "fruit"},
    ),
    Document(
        id=str(uuid.uuid4()),
        page_content="Orange Navel",
        metadata={"description": "orange", "content": "3", "category": "fruit"},
    ),
]

await store.aadd_documents(docs)
```
#### フィルタリング（Retriever）
- https://python.langchain.com/docs/integrations/vectorstores/pgvectorstore/#search-for-documents-with-metadata-filter
```python
import uuid

docs = [
    Document(
        id=str(uuid.uuid4()),
        page_content="Red Apple",
        metadata={"description": "red", "content": "1", "category": "fruit"},
    ),
    Document(
        id=str(uuid.uuid4()),
        page_content="Banana Cavendish",
        metadata={"description": "yellow", "content": "2", "category": "fruit"},
    ),
    Document(
        id=str(uuid.uuid4()),
        page_content="Orange Navel",
        metadata={"description": "orange", "content": "3", "category": "fruit"},
    ),
]

await custom_store.aadd_documents(docs)

# Use a dictionary filter on search
docs = await custom_store.asimilarity_search(query, filter={"content": {"$gte": 1}})

print(docs)
```

---

## Hybrid Search（ハイブリッド検索）
- https://python.langchain.com/docs/integrations/vectorstores/pgvectorstore/#hybrid-search-with-pgvectorstore
- ベクトル(類似度)検索とキーワード検索を組み合わせた検索方法
- `PGVectorStore`は、TSV（Text Search Vector）ベースのキーワード検索をサポートしている
  - TSV（Text Search Vector）とは、テキストを検索用に最適化した形式に変換したデータ型（PostgreSQLの全文検索機能で使用されるデータ型）
- `PGVectorStore`の`create`や`create_sync`メソッドの`hybrid_search_config`パラメータに [HybridSearchConfig](https://python.langchain.com/api_reference/postgres/v2/langchain_postgres.v2.hybrid_search_config.HybridSearchConfig.html) オブジェクトを渡すことで、ハイブリッド検索を有効化できる
  - `tsv_lang`
    - tokenizationに使用する言語を指定（デフォルトは`'pg_catalog.english'`で、**japaneseはない**）
    - `english`の場合、半角スペースなどで単語が分割されるため、日本語はスペースがないので、うまく分割されないため、精度が悪くなる可能性がある

### 設定例（使い方の例）
- **metadataを使ったフィルタリングも同時に行う例**  
  ```python
  from langchain_postgres.v2.hybrid_search_config import (
      HybridSearchConfig,
      reciprocal_rank_fusion,
  )
  from langchain_postgres import PGVectorStore, PGEngine
  from langchain_postgres.v2.engine import Column
  from langchain_aws import BedrockEmbeddings

  hybrid_search_config = HybridSearchConfig(
      tsv_column="hybrid_description", ★ # TSV（Text Search Vector）カラム。単語単位で分割されたテキストを格納されるカラム
      tsv_lang="pg_catalog.english",
      fusion_function=reciprocal_rank_fusion,
      fusion_function_parameters={
          "rrf_k": 60,
          "fetch_top_k": 10,
      },
  )

  connection_string = (
      f"postgresql+psycopg://{postgre_user}:{postgre_password}@{db_host}"
      f":5432/{db_name}"
  )
  pg_engine = PGEngine.from_connection_string(connection_string)

  pg_engine.init_vectorstore_table(
      table_name="vectorstore", # このTable名で作成される
      schema_name="rag_schema", # 事前にschemaを作成しておく必要がある
      vector_size=3072, # embeddingの次元数に合わせる
      id_column="id", # primary key（uuid）
      content_column="content", # Embeddingされる前の元のドキュメントが格納されるカラム
      embedding_column="embedded_content", # Embeddingベクトルを格納するカラム
      metadata_columns=[ # metadataを格納するカラム（以下の１行１行がカラムとして作成される。以下の例だと`system_id`、`incident_id`、`created_at`カラムが作成される）
          Column(name="system_id", data_type="text"), # `data_type`は`text`、`integer`、`timestamp`などPostgreSQLのデータ型を指定可能
          Column(name="incident_id", data_type="text"),
          Column(name="created_at", data_type="timestamp"),
      ],
      metadata_json_column="metadata", # metadata全体をJSON形式で格納するカラム
      hybrid_search_config=hybrid_search_config, # `HybridSearchConfig`オブジェクトを渡す
      store_metadata=True,
  )

  embeddings = BedrockEmbeddings(
      model_id="cohere.embed-multilingual-v3",
      region_name="us-west-2",
  )

  hybrid_store = PGVectorStore.create_sync( # Storeインスタンスを初期化
      engine=pg_engine,
      schema_name="rag_schema",
      table_name="vectorstore",
      embedding_service=embeddings,
      content_column="content",
      id_column="id",
      embedding_column="embedded_content",
      metadata_json_column="metadata",
      metadata_columns=["system_id", "incident_id", "created_at"],
      hybrid_search_config=hybrid_search_config,
  )

  # VectorStoreにドキュメントを格納
  docs = [
      Document(
          page_content="Alert Name: ECS CPU usage alert, Priority: warning, summary: ECS CPU usage high",
          metadata={"system_id": "apple", "incident_id": "1", "created_at": "2025-10-01T12:00:00Z"},
      ),
      Document(
          page_content="Alert Name: rds-error log alert, Priority: warning, summary: some rds-error occurred",
          metadata={"system_id": "apple", "incident_id": "2", "created_at": "2024-12-01T12:00:00Z"},
      ),
      Document(
          page_content="Alert Name: ALB 5xx error alert, Priority: critical, summary: ALB 5xx error occurred",
          metadata={"system_id": "samsung", "incident_id": "3", "created_at": "2004-12-01T12:00:00Z"},
      ),
      Document(
          page_content="Alert Name: slow query log alert, Priority: critical, summary:  slow query log detected",
          metadata={"system_id": "samsung", "incident_id": "4", "created_at": "2024-04-03T12:00:00Z"},
      ),
  ]

  hybrid_store.add_documents(docs)

  # ハイブリッド検索
  query = "CPU usage high"
  results = hybrid_store.similarity_search_with_score(
              query=query,
              k=5, # 取得する上位k件
              filter={"system_id": {"$eq": "apple"}}, ★ # `metadata_columns`で指定したカラムを使ったフィルタリング
          )
  print(results)
  ```

> [!CAUTION]
> `metadata_json_column`で指定したカラムを使ったフィルタリングはうまく機能しなかった。要確認。