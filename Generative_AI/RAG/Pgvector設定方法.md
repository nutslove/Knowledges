- https://github.com/pgvector/pgvector
- https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/AuroraUserGuide/AuroraPostgreSQL.VectorDB.html （参考）

## Docker単体のコンテナとしてPGVector対応のコンテナ作成方法
- container用のVolumeを作成  
  ```shell
  docker volume create pg_data
  ```
- PGVector対応済みのpgvectorプロジェクトが公式でPostgresと統合されたDockerイメージでコンテナ実行  
  ```shell
  docker run --name pgvector \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  -v pg_data:/var/lib/postgresql/data \
  -d pgvector/pgvector:pg17
  ```

## 手順
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
- vector用テーブル作成  
  ```shell
  CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(1536) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
  - `content`
    - ベクトル化される前の生テキスト（原文）
  - `embedding`
    - contentをembeddingモデルでベクトル化した数値配列
  - `metadata`
    - メタデータによるフィルタリングに使える
  - `created_at`, `updated_at`
    - 必須ではないが、ドキュメントの挿入順序を把握できたり、古いデータの削除やアーカイブ処理、データの更新履歴管理などができる等のメリットがあるので追加しておく

> [!IMPORTANT]  
> `vector(n)`の `n`の部分の数字はEmbeddingモデルの **ディメンション** サイズに合わせる

- 作成したテーブルのカラム情報を確認  
  ```shell
  \d documents
  ```

- インデックス作成（検索性能向上）  
  ```shell
  CREATE INDEX ON documents USING hnsw (embedding vector_cosine_ops) WITH (ef_construction=256);
  ```

> [!CAUTION]  
> Embeddingモデルのディメンションサイズ（次元）（`vector(n)`）が2000より大きい場合、HNSWインデックスは使えない。  
> その場合は次元数を下げるか、IVFFlatインデックスを使用する。  
> ```shell
> CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
> ```