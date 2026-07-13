# SQLAlchemyの relationship() について

`relationship()` は、**Pythonのオブジェクト同士の関連（つながり）を定義する**ための宣言。
「`user.posts` でその人の投稿一覧を取れる」「`post.author` で投稿者を取れる」といった、**オブジェクトをたどる橋渡し**を作る。

> [!IMPORTANT]
> ### `ForeignKey` と `relationship()` は役割が違う
> - **`ForeignKey`** … **DB（テーブル）側**の制約。「この列は他テーブルの主キーを指す」という**カラム**の定義。DBに実際の外部キー制約を作る。
> - **`relationship()`** … **Python（ORM）側**の定義。実テーブルの列は作らない。**オブジェクトをたどるための属性**を提供するだけ。
>
> 多対1・1対多を作るには**両方**が要る。`ForeignKey` が「どの列で結びつくか」を決め、`relationship()` が「その結びつきをオブジェクトとしてどう見せるか」を決める。

## 前提モデル（1対多）
「1人のUserが複数のPostを持つ」という関係を例にする。

```python
from sqlalchemy import ForeignKey, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship

class Base(DeclarativeBase):
    pass

class User(Base):
    __tablename__ = "users"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(100))

    # 「1人のUser」から見た「複数のPost」 → list
    posts: Mapped[list["Post"]] = relationship(back_populates="author")

class Post(Base):
    __tablename__ = "posts"
    id: Mapped[int] = mapped_column(primary_key=True)
    title: Mapped[str] = mapped_column(String(200))

    # DB上の結びつき（外部キー列）
    author_id: Mapped[int] = mapped_column(ForeignKey("users.id"))

    # 「1つのPost」から見た「1人のUser」 → 単数
    author: Mapped["User"] = relationship(back_populates="posts")
```

使い方:

```python
user = User(name="太郎")
post = Post(title="はじめての投稿", author=user)  # authorにUserオブジェクトを渡すだけ
user.posts.append(post)                          # どちら側から関連づけてもよい

print(post.author.name)     # 太郎  （子→親をたどる）
print(user.posts[0].title)  # はじめての投稿  （親→子をたどる）
```

## 型ヒントで「多重度」を表す
SQLAlchemy 2.0 では、`Mapped[...]` の型ヒント自体が「1対多か・多対1か」を表現する。

| 型ヒント | 意味 | 例 |
| --- | --- | --- |
| `Mapped[list["Post"]]` | 複数持つ側（1対多の「1」側 / 多対多） | `User.posts` |
| `Mapped["User"]` | 単数を指す側（多対1・1対1） | `Post.author` |
| `Mapped["User \| None"]` | 単数だが**任意**（関連が無くてもよい） | 親が無くても成立するとき |

> [!NOTE]
> `"Post"` のように**文字列で書く**のは、まだ定義前のクラスを前方参照するため（forward reference）。相互に参照し合うモデルではこの書き方が必要になる。

## `back_populates` — 双方向の関連をつなぐ
- `User.posts` と `Post.author` は**同じ1つの関連を、両側から見たもの**
- `back_populates="相手側の属性名"` を**両側に**書くことで、SQLAlchemyに「この2つは対（つい）である」と教える
- これにより、片側を変更すると**もう片側にも自動で反映**される（同期される）

```python
user = User(name="花子")
post = Post(title="投稿A")

post.author = user          # 子側で親をセットすると…
print(post in user.posts)   # True  ← 親側のリストにも自動で入る（back_populatesのおかげ）
```

#### 「親側」とはどれのこと？

この例は「1人のUserが複数のPostを持つ」1対多。1対多では「1」側を**親**、「多」側を**子**と呼ぶ。

- **親 = `User`（`user` オブジェクト）** … 複数の子を持つ側
- **子 = `Post`（`post` オブジェクト）** … 親に属する側

コメントの「**親側のリスト**」とは **`user.posts`**（親Userが持つ投稿一覧）のこと。

```
        user （親 / User）
          │  user.posts  ← 「親側のリスト」。ここに自動で post が入る
          ▼
        post （子 / Post）
             post.author = user  ← ここだけ手動でセットした
```

やったのは `post.author = user`（**子側で親をセットしただけ**）。それでも `back_populates` が対に
なっているおかげで、反対側の `user.posts` にも自動で `post` が追加され、`post in user.posts` が `True` になる。
`back_populates` を書いていなければ `user.posts` は空のまま（`False`）になる。

> [!CAUTION]
> `back_populates` を書き忘れると、片側だけ更新してももう片側に反映されず、コミット後まで不整合に見えるバグの原因になる。**双方向にしたいなら両側に対で書く**のが原則。

### `backref`（古い書き方）との違い
- `backref` は**片側だけに書けば**、もう片側の属性を**自動生成**してくれる省略記法。

  ```python
  class User(Base):
      # これ1つで Post.author も自動生成される
      posts: Mapped[list["Post"]] = relationship(backref="author")
  # Post側には relationship を書かない
  ```
- 手軽だが、**もう片側の定義がコード上に見えない**ため可読性・型補完で不利。

> [!TIP]
> 現在は **`back_populates` で両側を明示するのが推奨**。`backref` はレガシーコードで見かける旧スタイル、という理解でよい。

## 関連の種類ごとの定義パターン

### 多対1（Many-to-One）
- 上の `Post.author`（多くのPostが1人のUserを指す）がこれ。
- **外部キー（`ForeignKey`）は「多」側のテーブルに持たせる**のが鉄則。

### 1対1（One-to-One）
- 1対多の「多」側を単数に制限したもの。片側に `uselist=False` を付ける。

  ```python
  class User(Base):
      __tablename__ = "users"
      id: Mapped[int] = mapped_column(primary_key=True)
      # listではなく単数。uselist=False で1対1にする
      profile: Mapped["Profile"] = relationship(back_populates="user", uselist=False)

  class Profile(Base):
      __tablename__ = "profiles"
      id: Mapped[int] = mapped_column(primary_key=True)
      user_id: Mapped[int] = mapped_column(ForeignKey("users.id"), unique=True)
      user: Mapped["User"] = relationship(back_populates="profile")
  ```

### 多対多（Many-to-Many）
- **中間テーブル（association table）** を用意し、`relationship(secondary=...)` で指定する。
- 中間テーブルだけなら `Table` で定義するのが簡単（両側の外部キーだけを持つ）。

  ```python
  from sqlalchemy import Table, Column, ForeignKey

  # 中間テーブル（student と course をつなぐ）
  enrollment = Table(
      "enrollment",
      Base.metadata,
      Column("student_id", ForeignKey("students.id"), primary_key=True),
      Column("course_id", ForeignKey("courses.id"), primary_key=True),
  )

  class Student(Base):
      __tablename__ = "students"
      id: Mapped[int] = mapped_column(primary_key=True)
      courses: Mapped[list["Course"]] = relationship(
          secondary=enrollment, back_populates="students"
      )

  class Course(Base):
      __tablename__ = "courses"
      id: Mapped[int] = mapped_column(primary_key=True)
      students: Mapped[list["Student"]] = relationship(
          secondary=enrollment, back_populates="courses"
      )
  ```

> [!NOTE]
> 中間テーブル自体に**追加のカラム**（履修日・成績など）を持たせたい場合は、`secondary` ではなく**中間テーブルもモデル化**し、1対多を2本つなぐ「association object」パターンにする。

## `foreign_keys` — 関連が2本以上あるとき
同じテーブルへの外部キーが複数あると、SQLAlchemyが「どの列で結ぶか」を判断できずエラーになる。そのときは `foreign_keys=` で明示する。

```python
class Message(Base):
    __tablename__ = "messages"
    id: Mapped[int] = mapped_column(primary_key=True)
    sender_id: Mapped[int] = mapped_column(ForeignKey("users.id"))
    receiver_id: Mapped[int] = mapped_column(ForeignKey("users.id"))

    # usersへの外部キーが2本あるので、どちらを使うか明示する
    sender: Mapped["User"] = relationship(foreign_keys=[sender_id])
    receiver: Mapped["User"] = relationship(foreign_keys=[receiver_id])
```

## よく使う主なオプション
| オプション | 役割 |
| --- | --- |
| `back_populates` | 双方向の対になる属性名を指定（両側に書く） |
| `backref` | 片側だけで双方向を自動生成（旧スタイル） |
| `secondary` | 多対多の中間テーブルを指定 |
| `uselist=False` | 1対1にする（コレクションでなく単数に） |
| `foreign_keys` | 外部キーが複数あるとき、使う列を明示 |
| `cascade` | 親の削除・保存を子へ波及させる（例 `"all, delete-orphan"`） |
| `lazy` | 関連の**読み込み方の既定**を指定 → 詳細は [[SQLAlchemyのローディング戦略（eager loading・N+1問題）について]] |

> [!TIP]
> `cascade="all, delete-orphan"` は「親を消したら子も消す」「親のコレクションから外した子（orphan）も消す」挙動。親子が生存を共にする関係（User⇔Profileなど）でよく使う。

## まとめ
- `ForeignKey` は**DB側の列制約**、`relationship()` は**Python側のオブジェクト橋渡し**。多対1・1対多は両方セットで作る
- 型ヒント `Mapped[list["X"]]` / `Mapped["X"]` で「複数持つ / 単数を指す」を表す
- 双方向は **`back_populates` を両側に**書くのが現在の推奨（`backref` は旧スタイル）
- 多対多は `secondary=中間テーブル`、外部キーが複数なら `foreign_keys=` で曖昧さを解消
- 「定義した関連を**どう読み込むか**」は本ノートの範囲外 → [[SQLAlchemyのローディング戦略（eager loading・N+1問題）について]] を参照

## 関連
- [[DBスキーマの定義とマイグレーション（SQLAlchemy・Alembic）について]]
- [[SQLAlchemyのローディング戦略（eager loading・N+1問題）について]]

## 参考リンク
- SQLAlchemy公式: Relationship Configuration — https://docs.sqlalchemy.org/en/20/orm/relationships.html
- SQLAlchemy公式: Basic Relationship Patterns — https://docs.sqlalchemy.org/en/20/orm/basic_relationships.html
