# FastAPIでのDB連携（SQLAlchemy AsyncSession）について

このファイルは、[非同期処理(async, await)について](../非同期処理%28async,%20await%29について.md)（asyncioの基礎）と、Python/DBフォルダの[DBスキーマの定義とマイグレーション（SQLAlchemy・Alembic）について](../DB/DBスキーマの定義とマイグレーション（SQLAlchemy・Alembic）について.md)等（同期SQLAlchemyでのモデル設計）の**両方を前提に、FastAPIのエンドポイントの中で非同期SQLAlchemy（`AsyncSession`）をどう組み込むか**をまとめたもの。単体のasyncio知識・単体のSQLAlchemy知識では埋まらない、両者の接続部分に焦点を当てている。

> 出典: [SQLAlchemy 2.0公式 - Asynchronous I/O (asyncio)](https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html)

---

## 1. 同期版との違い（全体像）

DBフォルダのモデル定義（`DeclarativeBase`で書いたテーブル定義）自体は**同期・非同期で共通**。変わるのは「Engine/Sessionの作り方」と「クエリの実行の仕方」の部分だけ。

| レイヤー | 同期（DBフォルダで既出） | 非同期（このファイル） |
|---|---|---|
| モデル定義（`DeclarativeBase`） | 共通 | 共通 |
| Engine作成 | `create_engine(...)` | `create_async_engine(...)` |
| Session作成 | `sessionmaker(...)` / `Session(...)` | `async_sessionmaker(...)` / `AsyncSession(...)` |
| クエリ実行 | `session.execute(stmt)` | `await session.execute(stmt)` |
| FastAPIへの注入 | `def get_db()` + `Depends` | `async def get_db()` + `Depends` |
| Alembic | そのまま動く | ひと手間必要（→セクション7） |

---

## 2. 非同期Engine / Sessionのセットアップ

```python
# db/session.py
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

DATABASE_URL = "postgresql+psycopg://user:pass@localhost:5432/app"  # 例: psycopg(v3)

engine = create_async_engine(DATABASE_URL, pool_size=10, max_overflow=20, pool_pre_ping=True)

async_session_maker = async_sessionmaker(
    engine,
    expire_on_commit=False,  # 非同期では特に重要（後述）
    autoflush=False,
)
```

- `async_sessionmaker`（SQLAlchemy 2.0で追加）が現行の推奨API。`sessionmaker(class_=AsyncSession)`という旧式の書き方も動くが、型補完の面で`async_sessionmaker`の方が扱いやすい
- `expire_on_commit=False`を付けるのが定石。デフォルト（`True`）のままだと、`commit()`後にオブジェクトの属性へアクセスした時点で**暗黙のリフレッシュ（DBへの再アクセス）が発生**し、非同期コンテキストではこれが正しく扱えず、後述の`MissingGreenlet`系のエラーやリクエスト外でのアクセスで問題になりやすい

### ドライバの選択（DBフォルダとの接続点）

DBフォルダのマイグレーション解説では`postgresql+psycopg://`（psycopg3、同期）を使っていたが、**psycopg3は非同期にも対応しており、同じ`postgresql+psycopg://`のままドライバを変えずに`create_async_engine`へ切り替えるだけでよい**（SQLAlchemyが`create_engine`/`create_async_engine`のどちらで呼ばれたかを見て、psycopgの同期/非同期モードを自動的に切り替える）。

| ドライバ文字列 | 特徴 |
|---|---|
| `postgresql+psycopg://`（psycopg3） | 同期・非同期を同じドライバで両対応。DBフォルダの同期コードからの移行が楽 |
| `postgresql+asyncpg://`（asyncpg） | 非同期専用。ベンチマーク上psycopgより高速なことが多く、非同期前提の新規プロジェクトでよく選ばれる |

> [!NOTE]
> `psycopg2`（v2系）は非同期に対応していない。`create_async_engine`に`postgresql+psycopg2://`を渡すとエラーになる。既存プロジェクトが`psycopg2`のままなら、`psycopg`（v3）または`asyncpg`への切り替えが必要。

---

## 3. FastAPIの`Depends`との組み込み（session-per-request）

[APIRouterについて](APIRouterについて.md)で説明した「`yield`を使う依存」のパターンがそのまま当てはまる。リクエストごとに1つの`AsyncSession`を払い出し、リクエスト終了時にクローズする。

```python
# db/session.py（続き）
from collections.abc import AsyncGenerator


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session_maker() as session:
        try:
            yield session
            await session.commit()   # 例外なく処理が終わったらコミット
        except Exception:
            await session.rollback()
            raise
```

```python
# routers/items.py
from typing import Annotated
from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from db.session import get_db

router = APIRouter(prefix="/items")
DbSession = Annotated[AsyncSession, Depends(get_db)]


@router.get("/{item_id}")
async def read_item(item_id: int, db: DbSession):
    item = await db.get(Item, item_id)
    return item
```

- `async with async_session_maker() as session:` を使うと、`yield`後（＝レスポンス生成後）に**自動でセッションがクローズ**される（同期版の`finally: db.close()`に相当する処理をコンテキストマネージャが肩代わりする）
- コミットのタイミングは大きく2通りある

| 方針 | 説明 |
|---|---|
| 依存関数側で自動コミット（上記の例） | エンドポイントは`db.add()`等だけ書けばよく、シンプル。ただし「このエンドポイントではコミットしたくない」という細かい制御はしにくい |
| エンドポイント側で明示的に`await db.commit()` | 制御は柔軟だが、書き忘れるとデータが保存されない |

チームやプロジェクトの規模に応じてどちらかに統一するのが無難。

---

## 4. `lifespan`とEngineのライフサイクル（[lifespanについて](lifespanについて.md)との接続点）

`create_async_engine`自体は呼び出した時点で実際に接続を開くわけではなく、**内部にコネクションプールを持つオブジェクトを作るだけ**なので、モジュールレベルで作成しても問題ない。ただし、**アプリ終了時に確実にプールを破棄する**ために、`dispose()`は`lifespan`のシャットダウン処理で呼ぶのが定石。

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI
from db.session import engine


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield
    await engine.dispose()   # 終了時にコネクションプールを破棄


app = FastAPI(lifespan=lifespan)
```

- [lifespanについて](lifespanについて.md)で説明した「起動時に1回だけ用意し、終了時に1回だけ後始末する」パターンそのもの
- 起動時に接続確認をしておきたい場合は、`yield`の前で`async with engine.begin(): pass`のように軽いクエリを投げてヘルスチェックすることもある

---

## 5. クエリの書き方（同期版との違い）

SQLAlchemy 2.0の`select()`スタイルのクエリ構文自体は同期・非同期で共通。**実行部分にすべて`await`が付く**のが違い。

```python
from sqlalchemy import select

# 主キーで1件取得
item = await db.get(Item, item_id)

# 条件検索（複数件）
stmt = select(Item).where(Item.price > 100)
result = await db.execute(stmt)
items = result.scalars().all()

# 条件検索（1件、無ければNone）
stmt = select(Item).where(Item.name == "foo")
result = await db.execute(stmt)
item = result.scalar_one_or_none()

# 新規作成
new_item = Item(name="bar", price=200)
db.add(new_item)
await db.flush()     # IDを確定させたい場合（コミット前でも可）
```

- `session.query(...)`という旧スタイル（1.x系のAPI）は非同期版には無い。**必ず`select()` + `execute()`のスタイル**（2.0スタイル）を使う
- `db.get(Model, pk)`は主キー取得専用の非同期対応ショートカット

---

## 6. 非同期での`lazy loading`の罠（[SQLAlchemyのローディング戦略](../DB/SQLAlchemyのローディング戦略（eager%20loading・N+1問題）について.md)との接続点）

DBフォルダの「ローディング戦略」ファイルで説明されている`lazy loading`（属性へのアクセス時に暗黙で追加クエリを発行する仕組み）は、**非同期では素直に動かない**。

```python
# ❌ 非同期でこれをやるとエラーになりやすい
item = await db.get(Item, 1)
# item.owner はリレーション。lazy="select"（デフォルト）だと
# セッションのスコープ外や、await不可能な文脈でアクセスした時点で例外
print(item.owner.name)
```

これを実行すると多くの場合、以下のようなエラーになる。

```
MissingGreenlet: greenlet_spawn has not been called; can't call await_only() here.
```

**原因**: `AsyncSession`は「暗黙のブロッキングI/O」を一切許可しない設計になっている。同期版なら属性アクセス時に裏でこっそりSQLを発行できたが、非同期ではその「こっそり」ができない（`await`できる場所でしか新しいクエリを発行できない）。

### 対処法

DBフォルダの[SQLAlchemyのローディング戦略](../DB/SQLAlchemyのローディング戦略（eager%20loading・N+1問題）について.md)で説明した`selectinload`等の**明示的なeager loading**が、非同期では同期版以上に重要になる。

```python
from sqlalchemy.orm import selectinload

stmt = select(Item).options(selectinload(Item.owner)).where(Item.id == 1)
result = await db.execute(stmt)
item = result.scalar_one()

print(item.owner.name)   # 追加クエリなしでアクセスできる（既にロード済み）
```

> [!TIP]
> 開発中に「アクセスされたら困るリレーション」を早期発見したいなら、`relationship(..., lazy="raise")`を設定しておくと、eager loadingし忘れたアクセスがその場で例外になり、本番でのタイムアウトや原因不明のエラーより前に気づける（DBフォルダの「ローディング戦略」ファイルの`lazy`パラメータ解説を参照）。

---

## 7. Alembicとの関係

DBフォルダの[マイグレーションについて](../DB/DBスキーマの定義とマイグレーション（SQLAlchemy・Alembic）について.md)で説明したAlembicの運用は、アプリが非同期Engineを使っていても**基本的にそのまま使える**。Alembic自体はマイグレーション処理を同期的に実行する設計だが、`alembic init -t async`で生成される`env.py`が、非同期Engineから同期的なコネクションを取り出すブリッジ処理（`connection.run_sync(...)`）を自動的に用意してくれる。

```python
# migrations/env.py（async テンプレートの要点）
async def run_migrations_online():
    connectable = engine  # create_async_engine で作ったEngine
    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)  # ここで同期の世界へブリッジ
    await connectable.dispose()
```

- 新規に非同期構成でプロジェクトを始めるなら、`alembic init -t async migrations`のように**asyncテンプレートで初期化**すると、この部分を自分で書かずに済む
- 既にDBフォルダの手順で同期テンプレートのAlembicを組んでいる場合は、Alembic用に別途**同期Engine**（`postgresql+psycopg://`のまま`create_engine`で作る等）を用意し、マイグレーションだけ同期経路で回す構成でも問題ない（アプリ本体の非同期Engineとは別物として共存できる）

---

## ポイントまとめ

- モデル定義（DBフォルダの`DeclarativeBase`）は同期・非同期で共通。変わるのはEngine/Session作成とクエリ実行部分だけ
- `create_async_engine` + `async_sessionmaker(expire_on_commit=False)`が基本形。ドライバは`postgresql+psycopg://`（既存の同期コードと共存しやすい）か`postgresql+asyncpg://`（非同期専用・高速）を選ぶ。`psycopg2`は非同期非対応
- FastAPIへは[APIRouterについて](APIRouterについて.md)の`yield`依存パターンで注入する。`async with async_session_maker() as session: yield session`でリクエスト単位のセッションを払い出す
- Engineの`dispose()`は[lifespanについて](lifespanについて.md)のシャットダウン処理で呼ぶのが定石
- クエリは2.0スタイル（`select()` + `await session.execute()`）のみ。`session.query()`の旧スタイルは非同期には無い
- **非同期での`lazy loading`は`MissingGreenlet`エラーの原因になりやすい**。[ローディング戦略](../DB/SQLAlchemyのローディング戦略（eager%20loading・N+1問題）について.md)の`selectinload`等による明示的eager loadingが同期版以上に重要
- Alembicは`alembic init -t async`で非同期Engine向けのブリッジ処理を自動生成できる。同期Engineを別途用意してマイグレーション専用に使う構成でもよい
