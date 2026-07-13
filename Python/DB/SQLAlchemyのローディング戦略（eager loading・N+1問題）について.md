# SQLAlchemyのローディング戦略（eager loading・N+1問題）について

`relationship()` で定義した関連オブジェクト（例: `User.posts`）を、**いつ・どうやってDBから読み込むか**を決めるのがローディング戦略。
選択を誤ると **N+1問題**（クエリの大量発行）やメモリの無駄読みが発生する。

> [!NOTE]
> `relationship()` そのもの（`back_populates`・1対多／多対1／多対多の定義など）は [[SQLAlchemyのrelationship()について]] を参照。本ノートは「そこで定義した関連を、クエリ実行時にどう読み込むか」に絞る。
> スキーマ定義・Alembicマイグレーション側は [[DBスキーマの定義とマイグレーション（SQLAlchemy・Alembic）について]] を参照。

## 前提となるモデル
以下、この2モデルを例に説明する。

```python
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy import ForeignKey

class Base(DeclarativeBase):
    pass

class User(Base):
    __tablename__ = "users"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str]
    posts: Mapped[list["Post"]] = relationship(back_populates="author")

class Post(Base):
    __tablename__ = "posts"
    id: Mapped[int] = mapped_column(primary_key=True)
    title: Mapped[str]
    author_id: Mapped[int] = mapped_column(ForeignKey("users.id"))
    author: Mapped["User"] = relationship(back_populates="posts")
```

## lazy loading（遅延ロード）とN+1問題
- **デフォルトは lazy loading**（`lazy="select"`）。関連属性に**アクセスした瞬間**に、その都度SELECTが飛ぶ
- ループ内でアクセスすると、親1回 + 子N回のクエリが発行される ＝ **N+1問題**

```python
users = session.execute(select(User)).scalars().all()  # ① SELECT * FROM users

for user in users:
    print(user.name, len(user.posts))
    # ↑ user.posts に触れるたびに
    #   SELECT * FROM posts WHERE author_id = ?  が1回ずつ飛ぶ（N回）
# 合計 1 + N 回のクエリ
```

> [!NOTE]
> ### 「N+1問題」とは何か（名前の由来）
> 関連データを取得するときに、クエリが **「1 + N回」** 発行されてしまうパフォーマンス問題のこと。
> - **「1」** … 親（一覧）を取るための最初の1クエリ（例: `SELECT * FROM users` で10人取得）
> - **「N」** … その一覧の各行について、関連を取るために追加で飛ぶN回のクエリ（例: 各userの投稿を取る `SELECT ... WHERE author_id = ?` が10回）
>
> → 上の例なら合計 **1 + 10 = 11回**。userが1000人なら**1001回**になり、**件数に比例してクエリが増える**のが問題の本質。
> 本来なら1回のJOINや `IN` 句でまとめて**2回**で済むはずのものが、大量のDB往復になり、通信オーバーヘッドで劇的に遅くなる。
>
> **なぜ気づきにくいか**: ORMでは `user.posts` という「ただの属性アクセス」の裏で暗黙にSQLが飛ぶため、コード上はループしているだけに見え、クエリが発行されていること自体が見えない。

> [!CAUTION]
> N+1問題は、件数が少ないローカル開発では気づきにくく、本番でデータが増えて初めて顕在化することが多い。ログにSQLを出す（`echo=True` や `SQLALCHEMY_ECHO`）と発見しやすい。
> また `relationship(lazy="raise")` にしておくと、遅延ロードが起きた瞬間に例外になるため、N+1の温床を**実行時に強制検出**できる（後述）。

## eager loading（先読み）の3つの手段
関連を**あらかじめまとめて読み込む**ことでN+1を防ぐ。クエリ実行時に `.options(...)` で指定する。

### 1. `selectinload` — ★多くの場合の第一候補
- 親を取得した後、**別クエリで `IN (...)` を使って子をまとめて取得**する（クエリは合計2回で済む）
- one-to-many（`User.posts` のような1対多）で特に有利。行の重複が起きない

```python
from sqlalchemy.orm import selectinload

users = session.execute(
    select(User).options(selectinload(User.posts))
).scalars().all()
# ① SELECT * FROM users
# ② SELECT * FROM posts WHERE author_id IN (1, 2, 3, ...)
# → 以降 user.posts へのアクセスで追加クエリは飛ばない
```

### 2. `joinedload` — 1対1・多対1向き
- **JOINして1クエリ**で親子を同時取得する
- many-to-one（`Post.author` のような多対1）や1対1で有利
- one-to-many に使うと親の行が子の数だけ重複して返る（SQLAlchemyが内部で重複排除するが、LIMITと相性が悪い）

```python
from sqlalchemy.orm import joinedload

posts = session.execute(
    select(Post).options(joinedload(Post.author))
).scalars().all()
# SELECT ... FROM posts LEFT OUTER JOIN users ON ...  （1クエリ）
```

> [!WARNING]
> `joinedload` + `LIMIT` は要注意。1対多でJOINすると行が増えるため、`LIMIT 10` が「親10件」ではなく「JOIN後の10行」を制限してしまう。件数制限したいときは `selectinload` を使う。

### 3. `subqueryload`
- サブクエリで子を読み込む古い方式。現在は `selectinload` が上位互換なことが多く、積極採用の場面は少ない。

## `contains_eager` — 自分でJOINした結果を関連に載せる
`joinedload` / `selectinload` は「SQLAlchemyにJOINやIN句を**自動生成させる**」もの。
これに対し **`contains_eager` は、自分で書いた `join()` の結果を、関連属性に手動で紐づける**ためのオプション。

### いつ使うか
- **JOIN先のテーブルで絞り込み（WHERE）やソートをしたい**とき
- そのJOINを**関連のロードにも再利用したい**（＝二重にJOIN/クエリさせたくない）とき

`joinedload` は「関連を読み込むためだけ」のJOINなので、そのJOINに自分でWHERE条件を足すことはできない。そこで、自分で `join()` を書いて条件を付け、その同じJOINを `contains_eager` でロードにも使う。

```python
from sqlalchemy.orm import contains_eager

# 「タイトルにpythonを含む投稿だけ」を、userに紐づけてロードしたい
stmt = (
    select(User)
    .join(User.posts)                      # ← 自分でJOINを書く
    .where(Post.title.contains("python"))  # ← JOIN先を条件で絞る
    .options(contains_eager(User.posts))   # ← このJOIN結果を User.posts に載せる
)
users = session.execute(stmt).unique().scalars().all()

for user in users:
    for post in user.posts:
        # ここには「pythonを含む投稿」だけが入っている（追加クエリなし）
        print(user.name, post.title)
```

> [!IMPORTANT]
> ### `contains_eager` と `joinedload` の決定的な違い
> - `joinedload(User.posts)` … SQLAlchemyが**別名の追加JOIN**を勝手に作る。`.where(Post.title...)` を書いても、そのJOINには影響しない（＝ロード結果は絞られない）。
> - `contains_eager(User.posts)` … **自分が書いたJOIN**をそのままロードに使う。だから `.where` の絞り込みがロード結果に反映される。
>
> ```python
> # ✕ 意図通りにならない例（joinedloadは絞り込みに使えない）
> select(User).join(User.posts).where(Post.title.contains("python")) \
>             .options(joinedload(User.posts))
> # → user.posts には「全部の投稿」が入ってしまう（joinedloadが別JOINを張るため）
> ```

> [!CAUTION]
> `contains_eager` は「JOINした行を**そのまま**関連に載せる」ため、条件で絞った結果が入る＝**関連の全件が入るとは限らない**。
> 「pythonを含む投稿だけ」を意図的にロードするなら正しいが、「絞り込みは一覧用、関連は全件欲しい」という場合は誤り。その場合は絞り込みと関連ロードを分ける（別クエリにする / `selectinload` を使う）。
>
> また one-to-many の JOIN では親行が重複するため、`.unique()` を付けてから `.scalars()` する必要がある。

## `relationship()` 側で既定を指定する `lazy` パラメータ
クエリごとの `.options()` ではなく、**モデル定義でデフォルトの読み込み方**を決めることもできる。

```python
class User(Base):
    # このrelationshipは常に selectin で読む、というデフォルト
    posts: Mapped[list["Post"]] = relationship(lazy="selectin")
```

| `lazy` の値 | 挙動 | 対応する `.options()` |
| --- | --- | --- |
| `"select"` | 既定。アクセス時に都度SELECT（遅延） | （なし＝デフォルト） |
| `"selectin"` | 別クエリの `IN` でまとめて先読み | `selectinload` |
| `"joined"` | JOINで同時に先読み | `joinedload` |
| `"subquery"` | サブクエリで先読み | `subqueryload` |
| `"raise"` | 遅延ロードが起きたら**例外**を投げる | ― |
| `"noload"` | 常に空でロードしない | `noload` |

> [!TIP]
> `lazy="raise"` は「意図しない遅延ロード（＝N+1の温床）をコードレビュー/実行時に検出したい」ときに有効。うっかり lazy loading した箇所で例外になるので、必ず `.options()` で明示ロードする設計を強制できる。

## 使い分けの早見表
| 状況 | 推奨 |
| --- | --- |
| 1対多（親→子の一覧）をまとめて読む | `selectinload` |
| 多対1・1対1（子→親）を読む | `joinedload` |
| JOIN先を条件で絞りつつ、そのJOINを関連ロードにも使う | `contains_eager` |
| LIMIT/ページングと併用したい | `selectinload`（`joinedload`は避ける） |
| 意図しない遅延ロードを検出したい | `relationship(lazy="raise")` |

> [!NOTE]
> ネストした関連（例: `User → posts → comments`）は、`selectinload(User.posts).selectinload(Post.comments)` のようにチェーンで指定できる。

## まとめ
- デフォルトは lazy loading。ループ内アクセスで **N+1問題**が起きる
- 基本の解決策は **`selectinload`（1対多）** と **`joinedload`（多対1）**
- **`contains_eager`** は「自分で書いたJOIN＋WHEREの結果を、そのまま関連に載せたい」ときの専用オプション。`joinedload` では絞り込みをロードに反映できないのが使い分けの分岐点
- モデル定義側で既定を決めたいなら `relationship(lazy=...)`、N+1検出には `lazy="raise"`

## 関連
- [[DBスキーマの定義とマイグレーション（SQLAlchemy・Alembic）について]]
- [[Pydantic, TypedDict, typingについて]]
- [[FastAPI]]

## 参考リンク
- SQLAlchemy公式: Relationship Loading Techniques — https://docs.sqlalchemy.org/en/20/orm/queryguide/relationships.html
- SQLAlchemy公式: `contains_eager` — https://docs.sqlalchemy.org/en/20/orm/queryguide/relationships.html#routing-explicit-joins-statements-into-eagerly-loaded-collections
