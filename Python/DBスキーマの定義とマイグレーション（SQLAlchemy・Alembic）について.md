# PythonでのDBスキーマ定義とマイグレーションについて

PythonでDBのスキーマ（テーブル構造）を定義し、実際にDBへ作成・変更する際の主流とベストプラクティスをまとめる。

---

## 結論（主流の構成）

| 役割 | ツール |
|---|---|
| スキーマ定義（モデル） | **SQLAlchemy 2.0**（`DeclarativeBase` + `Mapped[]`）または **SQLModel** |
| マイグレーション（実際にDBへ反映） | **Alembic** |

**鉄則:** 本番環境では「モデルからDBを直接生成する」のではなく、
**モデルで定義 → Alembicでマイグレーション生成 → レビュー → DBに反映** の流れにする。
`Base.metadata.create_all()` は本番では使わない（後述）。

---

## 1. スキーマの定義（SQLAlchemy 2.0）

SQLAlchemy 2.0では型ヒントベースの `Mapped[]` / `mapped_column()` を使うのが現在のスタイル。
1.x系の `Column(...)` だけで書くのは古い書き方。

```python
from datetime import datetime
from sqlalchemy import String, ForeignKey, func, MetaData
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship

# 命名規則を最初に定義しておくのがベストプラクティス（詳細は後述）
convention = {
    "ix": "ix_%(column_0_label)s",
    "uq": "uq_%(table_name)s_%(column_0_name)s",
    "ck": "ck_%(table_name)s_%(constraint_name)s",
    "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
    "pk": "pk_%(table_name)s",
}

class Base(DeclarativeBase):
    metadata = MetaData(naming_convention=convention)

class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    name: Mapped[str | None] = mapped_column(String(100))
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    posts: Mapped[list["Post"]] = relationship(back_populates="author")

class Post(Base):
    __tablename__ = "posts"

    id: Mapped[int] = mapped_column(primary_key=True)
    title: Mapped[str] = mapped_column(String(200))
    author_id: Mapped[int] = mapped_column(ForeignKey("users.id"))
    author: Mapped["User"] = relationship(back_populates="posts")
```

### ポイント
- `Mapped[str]` は **NOT NULL**、`Mapped[str | None]` は **NULLABLE** を型で表現できる。
- `server_default=func.now()` はDB側のデフォルト（`DEFAULT now()`）。
  Python側で入れる `default=...` とは別物なので注意。
- 共通カラム（`created_at`/`updated_at` など）は **Mixin** に切り出して重複を避ける。

```python
class TimestampMixin:
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        server_default=func.now(), onupdate=func.now()
    )

class User(TimestampMixin, Base):
    __tablename__ = "users"
    ...
```

### SQLModel という選択肢

FastAPI主体のプロジェクトなら **SQLModel**（SQLAlchemy + Pydantic の薄いラッパー、FastAPI作者のSebastián Ramírez製）も有力。
ORMモデルとAPIの入出力スキーマを1つのクラスで共有でき、ボイラープレートと「フィールドの二重管理によるズレ（field drift）」を減らせる。

```python
from sqlmodel import SQLModel, Field

class User(SQLModel, table=True):
    id: int | None = Field(default=None, primary_key=True)
    email: str = Field(unique=True, index=True)
    name: str | None = None
```

| | SQLModel が向く | 素の SQLAlchemy が向く |
|---|---|---|
| ケース | greenfieldなFastAPIプロジェクト、API重視 | 複雑なドメインモデル、細かい制御、レガシー連携 |
| 特徴 | コヒーレンス（一貫性）最大化、記述量が少ない | 制御性最大化、情報量・成熟度が高い |

- SQLModelは内部がSQLAlchemyなので、必要になったら生のSQLAlchemyに「降りる」ことも可能。
- **マイグレーションはSQLModelでもAlembicを使う**（この点は変わらない）。

> **SQLModel + Alembic 固有の注意点**（2025〜2026時点でも残っている定番のハマりどころ）
> - **FK/制約に名前が付かない** → `SQLModel.metadata` に `naming_convention` を設定する（後述の命名規則と同じ理由）。
> - autogenerateがSQLModel独自の型を解決できるよう、`env.py` の `context.configure(...)` に
>   `user_module_prefix="sqlmodel.sql.sqltypes."` を渡す。
> - SQLiteを使う場合は `render_as_batch=True` も渡す（SQLiteは `ALTER` 制約が弱く、
>   バッチモードでテーブル再作成する必要があるため。PostgreSQL/MySQLでは不要）。

---

## 2. DBへの作成（Alembicマイグレーション）

### なぜ `create_all()` ではダメか

```python
Base.metadata.create_all(engine)   # ← テスト／プロトタイプ限定
```

- これは「今のモデル定義に合うテーブルを作る」だけ。
- **既存テーブルへの変更（カラム追加・型変更・削除）を扱えない**。
- スキーマの変更履歴（バージョン管理）が残らない。
- → 本番では必ず **Alembic** を使う。

### セットアップ

```bash
pip install sqlalchemy alembic
alembic init migrations
```

`migrations/env.py` でモデルの metadata を指定する（autogenerateがこれと実DBを比較する）：

```python
from myapp.models import Base
target_metadata = Base.metadata
```

### 日常の運用サイクル

```bash
# 1. モデルを編集した後、差分から migration を自動生成
alembic revision --autogenerate -m "add users table"

# 2. ★生成されたスクリプトを必ず目視レビュー★
#    autogenerate はインデックス削除・サーバデフォルト・型変更などを
#    取りこぼすことがある

# 3. DBへ適用
alembic upgrade head

# 戻す場合
alembic downgrade -1
```

生成されるマイグレーションスクリプトの例：

```python
def upgrade():
    op.create_table(
        "users",
        sa.Column("id", sa.Integer(), nullable=False),
        sa.Column("email", sa.String(length=255), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_users"),
        sa.UniqueConstraint("email", name="uq_users_email"),
    )

def downgrade():
    op.drop_table("users")
```

### マイグレーションのベストプラクティス

- **autogenerate の結果は必ずレビューしてから適用**（自動生成を鵜呑みにしない）。
- **命名規則（naming_convention）を最初に設定**しておく。
  設定しないとDBによって制約名が自動命名され、後で `downgrade` や制約変更ができなくなる。
- マイグレーション内でデータ操作をする場合は、モデルを import せず
  **その時点のテーブルを migration 内で明示定義**する
  （将来モデルが変わっても過去の migration が壊れないように）。
- **ゼロダウンタイム**では「旧コードと新コードの両方が動くマイグレーション」にする。
  → **expand-contract パターン**（例: カラム削除・リネームは複数リリースに分割する）。
- CI/CD やコンテナ起動時は **アプリ起動前にマイグレーションを完走**させる。
  Docker Compose なら `depends_on` + healthcheck で「DB起動 → migration → アプリ起動」の順序を保証。
- マイグレーションファイルは Git 管理する。
  ブランチマージで複数 head ができたら `alembic merge` で統合。

---

## 3. テスト時のDB戦略

「単体テスト・E2EテストではローカルでDBコンテナを立ててやるのか？」への回答。
**基本は Yes。ただしテストの階層によって使い分ける。**

### 方針の全体像

| テスト種別 | 推奨するDB | 理由 |
|---|---|---|
| 純粋なロジックの単体テスト | DBを使わない（モック/インメモリ） | DBに依存しない部分は速く回す |
| DB絡みの単体〜結合テスト | **本番と同じDBのコンテナ**（Testcontainers等） | SQLと本番の挙動を一致させる |
| E2Eテスト | **本番と同じDBのコンテナ** | 本番に最も近い環境で検証 |

### SQLite `:memory:` か、本番同一DB(コンテナ)か

よくある2択とトレードオフ：

- **SQLite `:memory:`**
  - ○ 起動が速く、依存が少ない。CIが軽い。
  - ✗ 本番がPostgreSQL/MySQLだと**方言（型・制約・SQL関数・JSONB・`ON CONFLICT`等）が違う**。
    SQLiteで通ったSQLが本番で落ちる／その逆が起こる。
  - → **「本番DBと同じものを使う」のが今のベストプラクティスの主流**。
    SQLiteは「DB方言に依存しない純ロジックの高速テスト」に限定するのが安全。

- **本番と同じDBをコンテナで立てる（Testcontainers-Python / docker-compose）**
  - ○ 本番と同じ方言で検証できる。信頼性が高い。
  - ✗ 起動が重い → **fixture のスコープ**と**トランザクションのロールバック**で高速化する（下記）。

> 実務では「**Testcontainersで本番同一DBを起動 + 各テストごとにトランザクション/SAVEPOINTでロールバック**」が定番構成。

### 高速化のパターン（pytest）

- **コンテナ/エンジンは使い回す、データはテストごとに巻き戻す**
  - エンジンやテーブル作成は `session`/`module` スコープの fixture で1回だけ。
  - 各テスト（`function` スコープ）は**トランザクションを張って、テスト後に必ずロールバック**する。
    → DBを毎回作り直すより圧倒的に速く、テスト間が確実に独立する。

```python
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from myapp.models import Base

# コンテナ/エンジンはテスト全体で1回だけ（重い処理を使い回す）
@pytest.fixture(scope="session")
def engine():
    # 例: Testcontainers で立てた Postgres の URL を使う
    engine = create_engine("postgresql+psycopg://user:pass@localhost:5432/test")
    Base.metadata.create_all(engine)   # ← テストDBの初期化はここでは create_all でOK
    yield engine
    engine.dispose()

# 各テストは「接続 → 外側トランザクション開始 → テスト → ロールバック」で独立させる
@pytest.fixture
def session(engine):
    connection = engine.connect()
    transaction = connection.begin()   # 外側のトランザクション
    # ★ join_transaction_mode="create_savepoint" が重要（下記の注意を参照）
    session = Session(bind=connection, join_transaction_mode="create_savepoint")
    yield session
    session.close()
    transaction.rollback()   # ← 外側を巻き戻す＝テストで書いた内容を全て破棄
    connection.close()
```

> **重要な注意（よくある落とし穴）**
> `Session(bind=connection)` だけの素朴な実装だと、
> **テスト対象コードが内部で `session.commit()` を呼んだ瞬間に外側トランザクションまでコミットされ、
> ロールバックが効かなくなる**（テスト間の分離が壊れる）。
> SQLAlchemy 2.0公式が推奨するのは、Sessionに `join_transaction_mode="create_savepoint"` を渡す方法。
> こうするとSession側の `commit()` はSAVEPOINTのリリースに置き換わり、
> 外側トランザクションは触られないため、最後の `transaction.rollback()` で確実に全て巻き戻せる。
> （公式ドキュメント: "Joining a Session into an External Transaction (such as for test suites)"）

- Testcontainers を使う場合の起動例：

```python
from testcontainers.postgres import PostgresContainer

@pytest.fixture(scope="session")
def engine():
    with PostgresContainer("postgres:16") as pg:
        engine = create_engine(pg.get_connection_url())
        Base.metadata.create_all(engine)
        yield engine
```

### 注意点
- **fixtureスコープ**: 分離重視なら `function`、共有セットアップは `module`/`session`。
  重いエンジン生成を `module`/`session` に上げるだけで時間を大きく削減できる。
- **外部キーのCASCADE**はネストしたトランザクションだと期待通りロールバックされないことがある。
  → `DEFERRABLE INITIALLY DEFERRED` な制約にする等で対処。
- **並列実行（pytest-xdist）**では接続プールが競合するので、**ワーカーごとにエンジン/DBを分ける**。
- 非同期（`AsyncSession`）でも考え方は同じ。ネストトランザクション（SAVEPOINT）でロールバックする。
- **テスト用DBにも本番と同じマイグレーション（Alembic）を流して検証する**と、
  マイグレーション自体のテストにもなる（`create_all` ではなく `alembic upgrade head` を使う構成）。

### E2E / ローカル開発
- ローカル開発や E2E は **docker-compose で本番同等スタック（DBコンテナ含む）を立てる**のが一般的。
- 起動順は「DB(healthcheck) → migration → アプリ」を `depends_on` で保証する。

---

## まとめ

1. **定義** → SQLAlchemy 2.0（`DeclarativeBase` + `Mapped[]`）。FastAPI中心なら SQLModel。
2. **命名規則を最初に設定**（`naming_convention`）。
3. **作成・変更** → Alembic で `revision --autogenerate` → **レビュー** → `upgrade head`。
4. `create_all()` は本番で使わない（テスト初期化用途に留める）。
5. **テストは本番と同じDBをコンテナで立てる**のが主流。
   高速化は「コンテナ/エンジンは使い回し + テストごとにトランザクションでロールバック」。

---

## 参考リンク
- [SQLAlchemy 2.0 公式: Declarative Tables / mapped_column](https://docs.sqlalchemy.org/en/20/orm/declarative_tables.html)
- [SQLAlchemy 2.0 公式: Constraints（naming_convention）](https://docs.sqlalchemy.org/en/20/core/constraints.html)
- [SQLAlchemy 2.0 公式: Joining a Session into an External Transaction（テスト用ロールバック）](https://docs.sqlalchemy.org/en/20/orm/session_transaction.html#joining-a-session-into-an-external-transaction-such-as-for-test-suites)
- [Testcontainers for Python 公式](https://testcontainers.com/guides/getting-started-with-testcontainers-for-python/)
- [SQLModel & Alembic Tutorial（naming_convention / user_module_prefix の注意）](https://dev.to/mchawa/sqlmodel-alembic-tutorial-gc8)
- [Best Practices for Alembic and SQLAlchemy (DEV)](https://dev.to/welel/best-practices-for-alembic-and-sqlalchemy-3b34)
- [Alembic + SQLAlchemy: Migration Best Practices That Won't Break Production](https://medium.com/@ygsh0816/alembic-sqlalchemy-migration-best-practices-that-wont-break-production-09cc2f417715)
- [How to Handle Database Migrations with Alembic (OneUptime)](https://oneuptime.com/blog/post/2025-07-02-python-alembic-migrations/view)
- [SQLAlchemy vs SQLModel: Which Should You Choose?](https://tapanbasuli.medium.com/sqlalchemy-vs-sqlmodel-which-should-you-choose-for-your-python-project-7ea0b040af14)
- [FastAPI - SQL (Relational) Databases](https://fastapi.tiangolo.com/tutorial/sql-databases/)
- [How To Test Database Transactions With Pytest And SQLModel (Pytest with Eric)](https://pytest-with-eric.com/database-testing/pytest-sql-database-testing/)
- [A Guide To Database Unit Testing with Pytest and SQLAlchemy (CoderPad)](https://coderpad.io/blog/development/a-guide-to-database-unit-testing-with-pytest-and-sqlalchemy/)
