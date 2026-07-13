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

### `DeclarativeBase` とは（クラスの階層）

`DeclarativeBase` は **SQLAlchemy 2.0 が提供する、ORMモデルの基底クラス**。
「Pythonクラス ↔ DBテーブル」を対応づける仕組み（宣言的マッピング）を提供する。
これを継承して**プロジェクトに1つだけ** `Base` を作り、さらにそれを継承して
テーブルごとのモデルクラスを定義する。

```
DeclarativeBase               ← SQLAlchemyが提供する基底クラス
      ↓ 継承
class Base(DeclarativeBase)    ← 自分で1つだけ作る土台（全テーブル定義の集約点）
      ↓ 継承
class User(Base)              ← テーブルに対応するモデルクラス（= "users"テーブルの定義）
      ↓ インスタンス化
user = User(email="a@b.com")   ← "users"テーブルの1行（レコード）
```

| 段階 | 対応するもの |
|---|---|
| `DeclarativeBase` | ORMの仕組みそのもの（SQLAlchemy提供） |
| `Base` | プロジェクトの土台。継承した全テーブル定義が `Base.metadata` に集約される |
| `User` クラス | **テーブルそのものの定義**（1クラス = 1テーブル） |
| `User()` インスタンス | **テーブルの1行（レコード）** |

- ポイントは「`Base` を継承したモデルの情報がすべて `Base.metadata` に集まる」こと。
  だから Alembic は `target_metadata = Base.metadata` の一行で全テーブルを把握できる。
- 1.x系の `Base = declarative_base()`（関数）はレガシー扱い。2.0では
  `class Base(DeclarativeBase):` のクラス継承が公式推奨で、mypy/Pyright等の型チェッカが
  モデルの属性を正しく認識できる。

> [!NOTE]
> `Base` から作るのは「テーブルに対応する**クラス**」で、そのクラスの**インスタンス**が
> 「テーブルの1行」になる。「テーブル用のオブジェクトを作る」というより
> 「テーブル定義（クラス）を作り、それをインスタンス化して行データを扱う」と捉えると正確。

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

#### Mixin とは何か（なぜ切り出すのか）

`created_at` / `updated_at` のようなカラムは、`users`・`posts`・`comments`… と
**ほぼ全テーブルに欲しくなる**。各モデルに同じ定義をコピペすると、
テーブルが増えるたびに書く手間がかかり、仕様変更時に**全モデルを直す**羽目になる（直し漏れも起きる）。

そこで共通カラムを**専用の部品クラス（Mixin）に1回だけ**書き、各モデルに継承させて使い回す。

```python
class TimestampMixin:            # ← Base を継承しない「部品クラス」
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        server_default=func.now(), onupdate=func.now()
    )

class User(TimestampMixin, Base):   # ← 継承するだけで created_at/updated_at が付く
    __tablename__ = "users"
    id: Mapped[int] = mapped_column(primary_key=True)

class Post(TimestampMixin, Base):   # ← 別テーブルにも足すだけ
    __tablename__ = "posts"
    id: Mapped[int] = mapped_column(primary_key=True)
```

- **Mixin** = 「それ単体では使わず、他のクラスに"混ぜ込む(mix in)"ための部品クラス」（Python一般の概念）。
- `class User(TimestampMixin, Base):` は「`User` は `Base`（テーブルの土台）であり、
  そこに `TimestampMixin`（タイムスタンプ機能）を混ぜ込む」という読み方。
- 仕様を変えたいときは **`TimestampMixin` を1箇所直せば全モデルに反映**される（＝重複＝DRY違反を避ける）。

| | 普通の基底クラス（`Base`） | Mixin（`TimestampMixin`） |
|---|---|---|
| 役割 | 「〜である」の中心的な親（`User is a Base`） | カラム・機能の部品を足すだけ |
| 単体で使う？ | 使う | 使わない |

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

> [!CAUTION]
> **SQLModel + Alembic 固有の注意点**（2025〜2026時点でも残っている定番のハマりどころ）
> - **FK/制約に名前が付かない** → `SQLModel.metadata` に `naming_convention` を設定する（後述の命名規則と同じ理由）。
>   設定は**テーブル（モデル）が定義される前**に行う必要がある点に注意。
> - **`NameError: name 'sqlmodel' is not defined`** → SQLModelは `str` を独自型 `sqlmodel.sql.sqltypes.AutoString` にマップし、
>   autogenerateは生成物にその型名を書くが `import` を足さないため、マイグレーション実行時に落ちる。
>   定番の対策は **`script.py.mako`（マイグレーションのひな形）に `import sqlmodel` を1行足す**こと
>   （以降の生成ファイルすべてに import が入る）。既存ファイルは手で import を追記する。
> - SQLiteを使う場合は `env.py` の `context.configure(..., render_as_batch=True)` も渡す（SQLiteは `ALTER` が弱く、
>   バッチモード＝テーブル再作成で対応するため）。全バックエンドで付けたままでも安全（SQLite以外は通常動作）。

---

## 2. DBへの作成（Alembicマイグレーション）

### なぜ `create_all()` ではダメか

```python
Base.metadata.create_all(engine)   # ← テスト／プロトタイプ限定
```

- これは「今のモデル定義に合うテーブルを作る」だけ。
- **既存テーブルへの変更（カラム追加・型変更・削除）を扱えない**。
- スキーマの変更履歴（バージョン管理）が残らない。

> [!CAUTION]
> 本番では必ず **Alembic** を使う。`create_all()` はテスト・プロトタイプ限定。

### Alembicとは（考え方）

Alembic は **DBスキーマの「バージョン管理システム」**。イメージとしては **schema版のGit**。

- モデル（`Base.metadata`）と実DBの状態を比べ、その差分を「**マイグレーションファイル**」という
  Pythonスクリプトとして1つずつ積み上げていく。
- 各ファイルには **`upgrade()`（進める）** と **`downgrade()`（戻す）** が書かれる。
- 適用済みのバージョンは、DB内の **`alembic_version` テーブル**に記録される。
  「今このDBはどのバージョンまで適用されたか」をAlembic自身が覚えている。

```
モデル定義 ──(差分検出: autogenerate)──▶ マイグレーションファイル
                                          （0001 → 0002 → 0003 …と連なる）
                                                    │
                                          (alembic upgrade head)
                                                    ▼
                                              実際のDB
```

用語：

| 用語 | 意味 |
|---|---|
| **revision** | 1つのマイグレーション（バージョン）。ファイルごとに固有のID（例: `a1b2c3d4`）を持つ |
| **down_revision** | そのrevisionの「1つ前」。これで各revisionが鎖状につながり順序が決まる |
| **head** | 連なりの**最新**。`upgrade head` = 最新まで全部適用する、の意味 |
| **base** | 連なりの**最初**（何もない状態） |
| `alembic_version` | 実DB側で「現在どのrevisionまで来たか」を保持するテーブル |

### セットアップ

```bash
pip install sqlalchemy alembic
alembic init migrations       # ← migrations/ ディレクトリ一式を生成する
```

#### `alembic init` はどこで実行するか

**プロジェクトのルート（トップ階層＝アプリコードや `pyproject.toml` を置いている場所）で実行する。**

```
プロジェクトルート/            ← ここで alembic init migrations を実行する
├── alembic.ini              ← カレントに生成される（設定ファイル）
├── migrations/              ← 引数で指定した名前のフォルダが生成される
│   ├── env.py
│   ├── script.py.mako
│   └── versions/
├── myapp/                   ← 既存の自分のアプリコード
│   └── models.py
└── pyproject.toml
```

- `alembic.ini` は**カレントディレクトリ**に置かれ、`alembic` コマンドは実行時にカレントの `alembic.ini` を読む。
  → 以降 `alembic upgrade head` などを叩く作業ルートに置くのが自然。
- `env.py` から `from myapp.models import Base` のように**自分のモデルを import**するため、
  そのパッケージを import できる場所（＝ルート）に `migrations/` があると素直に通る。
- 引数（`migrations`）は生成される**フォルダ名**。慣習的に `migrations` や `alembic`（公式デフォルト）がよく使われる。
- 置き場所を変えたい場合は `alembic.ini` の `script_location = migrations` で後から指定し直せるが、
  特別な理由がなければ**ルートでそのまま実行するのが一番シンプル**。

`alembic init` で作られる主なもの：

| 生成物 | 役割 |
|---|---|
| `alembic.ini` | 接続先DB（`sqlalchemy.url`）などの設定ファイル |
| `migrations/env.py` | マイグレーション実行時の起点スクリプト。**どのmetadataと比較するか**をここで指定 |
| `migrations/versions/` | 生成されたマイグレーションファイルが1つずつ入るフォルダ |

`migrations/env.py` でモデルの metadata を指定する。**これがautogenerateの心臓部**で、
Alembicは「ここで渡された `target_metadata`（＝あるべき姿）」と「実DBの現状」を突き合わせて差分を出す：

```python
# `migrations/env.py`
from myapp.models import Base
target_metadata = Base.metadata   # ← ここを指すモデルを import し忘れると差分が出ない（空マイグレーションになる）
```

> [!NOTE]
> #### よくあるハマり
> モデルクラスを定義しただけで `env.py` から辿れる場所に import されていないと、
> `Base.metadata` にそのテーブルが登録されず、autogenerateが「テーブルなし」と誤検出する。
> 全モデルが読み込まれる状態にしておくこと。

### 日常の運用サイクル

```bash
# 1. モデルを編集した後、差分から migration を自動生成
alembic revision --autogenerate -m "add users table"
#    → migrations/versions/ に新しい .py ファイルが1つ生成される
#    （--autogenerate を付けないと upgrade/downgrade が空のひな形だけ作られる＝手書き用）

# 2. ★生成されたスクリプトを必ず目視レビュー★
#    autogenerate はリネーム・サーバデフォルト・CHECK制約・匿名制約などを
#    取りこぼす／誤検出することがある

# 3. DBへ適用（未適用のrevisionを順に実行し、alembic_version を更新）
alembic upgrade head

# 戻す場合（直近の1つ分 downgrade() を実行）
alembic downgrade -1

# 現在の状態を確認するコマンド
alembic current    # 今DBがどのrevisionにいるか
alembic history    # revisionの連なりを一覧
```

各コマンドの流れを整理すると：

1. **モデルを編集**（例: `User` に新カラム追加）
2. `revision --autogenerate` → Alembic が `Base.metadata`（あるべき姿）と実DB（現状）を比較し、
   差分を埋める `upgrade()` / `downgrade()` を書いた**新ファイル**を生成
3. **人間がレビュー**（後述のとおり autogenerate は万能ではない）
4. `upgrade head` → 未適用ファイルを順に実行してDBを最新化。`alembic_version` に「ここまで来た」と記録

生成されるマイグレーションファイルの中身（例）。先頭の**ヘッダ部分でrevisionの鎖**が表現される：

```python
"""add users table

Revision ID: a1b2c3d4e5f6
Revises: 9f8e7d6c5b4a         # ← 1つ前のrevision（down_revision）。これで順序が決まる
"""
revision = "a1b2c3d4e5f6"      # このファイル自身のID
down_revision = "9f8e7d6c5b4a" # 直前のID（最初のファイルなら None）

def upgrade():                 # ← upgrade head で実行される「進める」処理
    op.create_table(
        "users",
        sa.Column("id", sa.Integer(), nullable=False),
        sa.Column("email", sa.String(length=255), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_users"),
        sa.UniqueConstraint("email", name="uq_users_email"),
    )

def downgrade():               # ← downgrade で実行される「戻す」処理（upgradeの逆操作を書く）
    op.drop_table("users")
```

- `op.create_table` / `op.add_column` / `op.drop_constraint` などの **`op` はAlembicが提供する操作API**。
  これがDBごとの適切なSQL（`CREATE TABLE` / `ALTER TABLE …`）に変換される。
- `downgrade()` は `upgrade()` の**逆操作を自分で書く**（作ったら消す、追加したカラムは削る、など）。
  autogenerate はここもある程度埋めてくれるが、内容が正しいかはレビューで確認する。

### upgrade / downgrade はどこまで実行されるか

基本の対応はシンプル：

- `alembic upgrade` → マイグレーションファイルの **`upgrade()` を実行**（スキーマを進める）
- `alembic downgrade` → マイグレーションファイルの **`downgrade()` を実行**（スキーマを戻す）

ただし**対象は1ファイルとは限らない**。revision は `down_revision` で鎖状につながっており
（`0001 → 0002 → 0003`）、Alembic は `alembic_version` に記録された**現在地**と**目的地**を比べ、
その間の該当ファイルを**順に**実行する。

```
# upgrade: 今 0001 で `alembic upgrade head`（最新=0003）を実行
今: 0001  ──▶  目的地: 0003（head）
実行: 0002.upgrade() → 0003.upgrade()   （未適用の分を古い順に。0001は適用済みなので実行されない）

# downgrade: 今 0003 で `alembic downgrade -1`（1つ戻す）
今: 0003  ──▶  0002 に戻る
実行: 0003.downgrade()                   （今いるrevisionのdowngrade()を新しい順に）
```

| コマンド | 実行される関数 | 対象範囲 | 実行順 |
|---|---|---|---|
| `alembic upgrade head` | `upgrade()` | 現在地〜最新の**未適用**ファイル全部 | 古い→新しい |
| `alembic upgrade +1` | `upgrade()` | 次の1つだけ | — |
| `alembic downgrade -1` | `downgrade()` | 直近の1つだけ | — |
| `alembic downgrade base` | `downgrade()` | 現在地〜最初まで全部 | 新しい→古い |

> [!NOTE]
> 「`upgrade` は `upgrade()` を、`downgrade` は `downgrade()` を実行する」で正しい。
> そこに「対象は1ファイルとは限らず、`alembic_version`（現在地）から目的地までの該当ファイルを順に実行する」を足すと正確。

### autogenerate の目視レビュー観点

**どのファイルを見るか**: `alembic revision --autogenerate` 直後に `migrations/versions/` に生成された
新規 `.py`（例: `a1b2c3d4e5f6_add_users_table.py`）の **`upgrade()` / `downgrade()` の中身**。

autogenerate は「スキーマの見た目の差分」しか出さず、取りこぼし・誤検出がある。以下を重点的に確認する
（**上ほど危険＝データ損失・適用失敗に直結**）。

**1. リネームが「削除＋追加」になっていないか【最重要・データ損失】**
Alembic はカラム/テーブルの**rename を検出できず** `drop_column` + `add_column` を生成する。そのまま適用すると**データが消える**。

```python
# ❌ 自動生成（rename のつもりが drop+add になっている）
op.drop_column("users", "name")
op.add_column("users", sa.Column("full_name", sa.String()))
# ✅ 手で直す
op.alter_column("users", "name", new_column_name="full_name")
```

**2. NOT NULL 追加時の既存データ対策があるか【適用失敗】**
既存テーブルに `nullable=False` のカラムを足すと既存行が埋められず適用時にエラーになる。「NULL許可で追加 → backfill → NOT NULL化」の3段に直す。

```python
op.add_column("users", sa.Column("status", sa.String(), nullable=True))
op.execute("UPDATE users SET status = 'active' WHERE status IS NULL")
op.alter_column("users", "status", nullable=False)
```

**3. 型変更（type change）が意図どおりか**
型変更は **Alembic 1.12.0 以降デフォルトで検出される**（`compare_type=True` が既定）。ただし**バックエンドによって精度が異なり**、
`String(100)→String(200)` の長さ変更などは検出されないこともある。またPostgreSQLで型を変えるとき `USING` 句が要るケースは手当てが必要。
「検出された変更が正しいか」「検出されるべき変更が漏れていないか」の両面を確認する。

**4. `server_default` の変更が反映されているか**
DB側デフォルトの追加・変更・削除は **`compare_server_default` が既定 `False` のため比較対象外**。必要なら `env.py` の `context.configure(..., compare_server_default=True)` を有効化し、生成物に反映されているか確認。

**5. インデックス / 制約の検出範囲**
autogenerateが**検出するもの**: テーブル/カラムの増減、NULL可否の変更、インデックス、外部キー、**名前付き**の一意制約。
**検出しないもの**: **CHECK制約**、**匿名（無名）の制約**、`Enum` など一部の特殊型、シーケンス。
→ 一意制約やCHECK制約は**必ず名前を付ける**（＝`naming_convention` 設定が効く）。名前がないと検出対象から外れる。削除系の取りこぼしも確認する。

**6. `downgrade()` が正しい逆操作か**
`upgrade()` の逆を**正しい順序で**行うか（FKを張ってからテーブルを消すと失敗）。空・非対称になっていないか。

**7. 意図しない差分（余計なDROPなど）が混ざっていないか**
`target_metadata` に載っていない管理外テーブルを Alembic が「不要」と判断し `drop_table` を生成することがある。身に覚えのない削除がないか。→ `include_object` で除外設定。

**8. データ移行（DML）は自動生成されない**
autogenerate は**DDLだけ**を見る。値の変換・移送などデータ操作は生成されないので、必要なら `op.execute(...)` で追記（その際はモデルを import せずマイグレーション内でテーブルを明示定義）。

**9. 操作の順序・依存関係**
FKが絡むテーブルは作成・削除の順序が重要（参照される側を先に作る／参照する側を先に消す）。

| # | 観点 | 危険度 |
|---|---|---|
| 1 | rename が drop+add になっていないか | ★★★ データ損失 |
| 2 | NOT NULL 追加に backfill があるか | ★★★ 適用失敗 |
| 3 | 型変更が意図どおりか（既定で検出されるが精度はバックエンド依存） | ★★ |
| 4 | `server_default` の変更が反映されているか（既定で検出**しない**） | ★★ 検出漏れ |
| 5 | CHECK制約・匿名制約の取りこぼし（そもそも検出されない） | ★★ |
| 6 | `downgrade()` が正しい逆操作か | ★★ |
| 7 | 意図しない `drop_table` 等が混ざっていないか | ★★ |
| 8 | 必要なデータ移行(DML)を足したか | ★★ |
| 9 | FK等の操作順序 | ★ |

> [!NOTE]
> 一言でいうと「autogenerate は万能ではない。**rename・NOT NULL追加のbackfill・server_default・CHECK/匿名制約・データ移行**が典型的な穴」。ここを重点的に見て手で補うのがレビューの勘所。（型変更は既定で検出されるが、精度はバックエンド依存なので中身は要確認。）

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

### 命名規則（naming_convention）の詳細

`naming_convention` は、**DBの制約やインデックスに付く「名前」を、自分で決めたルールで自動的に統一する**仕組み。

#### 前提：制約・インデックスには「名前」がある

```sql
CREATE TABLE users (
    id INTEGER NOT NULL,
    email VARCHAR(255) NOT NULL,
    CONSTRAINT pk_users PRIMARY KEY (id),        -- ← "pk_users" が制約名
    CONSTRAINT uq_users_email UNIQUE (email)     -- ← "uq_users_email" が制約名
);
```

#### なぜ設定が必要か

`unique=True` と書くだけでは制約名を指定していないため、**名前はDBが勝手に付ける**（しかもDBごとにルールが違う）。

- PostgreSQL → `users_email_key`
- MySQL → `email`
- 別環境 → `uq_users_email_1` のような連番

→ **同じモデルでも環境によって制約名がバラバラ**になる。

制約を後から変更・削除するときSQLは**名前で対象を指定**する（`ALTER TABLE users DROP CONSTRAINT uq_users_email;`）。
Alembicの `downgrade()` も `op.drop_constraint("uq_users_email", "users")` のように名前を使うため、
**名前がDB任せだと実DBと食い違って `downgrade` が失敗する**。これが「設定しないと後で制約変更ができなくなる」の意味。

#### 設定内容と、Key / Value の関係

```python
convention = {
    "ix":  "ix_%(column_0_label)s",                                        # index
    "uq":  "uq_%(table_name)s_%(column_0_name)s",                          # unique制約
    "ck":  "ck_%(table_name)s_%(constraint_name)s",                        # check制約
    "fk":  "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",  # 外部キー
    "pk":  "pk_%(table_name)s",                                            # 主キー
}

class Base(DeclarativeBase):
    metadata = MetaData(naming_convention=convention)
```

**Key 側（`ix`/`uq`/`ck`/`fk`/`pk`）は決まっている** — SQLAlchemyが認識する制約種別のキー。勝手な名前は使えない。

| キー | 制約の種類 | 対応クラス |
|---|---|---|
| `ix` | インデックス | `Index` |
| `uq` | 一意制約 | `UniqueConstraint` |
| `ck` | チェック制約 | `CheckConstraint` |
| `fk` | 外部キー | `ForeignKeyConstraint` |
| `pk` | 主キー | `PrimaryKeyConstraint` |

> [!NOTE]
> 厳密にはキーには文字列の略称だけでなく、`UniqueConstraint` のような**制約クラスそのもの**も指定できる（細かく制御したいとき）。通常は略称で十分。

**Value 側は「テンプレートの形は自由・差し込む部品は決まっている」** の2層に分かれる。

- **リテラル文字**（`uq_` の接頭辞、区切りの `_` など）→ **完全に自由**。
- **`%(...)s` の差し込み変数** → SQLAlchemyが用意したものしか使えない（独自トークンは不可）。ただし並び順・使う/使わないは自由に組める。

| トークン | 中身 | 例 |
|---|---|---|
| `%(table_name)s` | テーブル名 | `users` |
| `%(column_0_name)s` | 対象カラム名 | `email` |
| `%(column_0_label)s` | テーブル+カラムのラベル | `users_email` |
| `%(column_0_key)s` | 対象カラムのkey | `email` |
| `%(referred_table_name)s` | 参照先テーブル名（FK用） | `users` |
| `%(referred_column_0_name)s` | 参照先カラム名（FK用） | `id` |
| `%(constraint_name)s` | 制約に付いた明示名（CK用） | — |

複数カラム版（`%(column_0N_name)s` / `%(column_0_N_name)s` など）もある。独自トークンは
辞書にコールバックを登録すれば追加もできる。

> [!CAUTION]
> `ck` の `%(constraint_name)s` は**制約に明示的な名前が付いていることを前提**とする。
> `CheckConstraint("age > 0")` のように `name=` なしで書くと、埋める名前がなくエラーになる。
> → CHECK制約には `CheckConstraint("age > 0", name="age_positive")` のように**必ず名前を付ける**
> （名前を付けたくない場合は `ck` 値を `%(column_0_name)s` ベースにする代替もあるが、明示名が確実）。

#### 実際に生成される名前

- `users` の主キー → **`pk_users`**
- `users.email` のunique → **`uq_users_email`**
- `posts.author_id` → `users.id` の外部キー → **`fk_posts_author_id_users`**

これらが**DB・環境に関係なく常に同じ名前**で作られる。

> [!CAUTION]
> **最初に1回だけ設定するのが鉄則**。途中から入れると、既にDBにある「旧ルールで付いた名前」と「新ルールの名前」が食い違い、制約を貼り直すマイグレーションが必要になる。

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

> [!NOTE]
> 実務では「**Testcontainersで本番同一DBを起動 + 各テストごとにトランザクション/SAVEPOINTでロールバック**」が定番構成。

### 高速化のパターン（pytest）

- **コンテナ/エンジンは使い回す、データはテストごとに巻き戻す**
  - エンジンやテーブル作成は `session`/`module` スコープの fixture で1回だけ。
  - 各テスト（`function` スコープ）は**トランザクションを張って、テスト後に必ずロールバック**する。
    → DBを毎回作り直すより圧倒的に速く、テスト間が確実に独立する。

```python
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
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

> [!CAUTION]
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
