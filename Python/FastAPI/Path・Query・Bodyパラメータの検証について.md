# FastAPIのPath・Query・Bodyパラメータの検証について

Path Operation Functionの引数は、型アノテーションだけでも自動的に「パス/クエリ/ボディのどこから来るか」が判定されるが（→[request, responseのスキーマについて](request,%20responseのスキーマについて.md)）、**追加のバリデーション・メタデータ**（最大長、範囲、正規表現、説明文など）を付けたい場合は `Path()` / `Query()` / `Body()` を使う。

> 出典: [FastAPI公式 - Path Parameters and Numeric Validations](https://fastapi.tiangolo.com/tutorial/path-params-numeric-validations/) / [Query Parameters and String Validations](https://fastapi.tiangolo.com/tutorial/query-params-str-validations/)

---

## 0. 前提：パス/クエリ/ボディの判定ルール

- パスの `{item_id}` と**名前が一致** → **パスパラメータ**
- パスに無い単純型（`int`/`str`など） → **クエリパラメータ**
- `BaseModel` を継承したクラス → **リクエストボディ**

詳細は→[request, responseのスキーマについて](request,%20responseのスキーマについて.md)

```python
from pydantic import BaseModel


class Item(BaseModel):
    name: str
    price: float


# パスパラメータ：パスの {item_id} と名前が一致
@app.get("/items/{item_id}")
async def read_item(item_id: int):
    return {"item_id": item_id}


# クエリパラメータ：パスに無い単純型 → /items/?q=xxx
@app.get("/items/")
async def read_items(q: str | None = None):
    return {"q": q}


# リクエストボディ：BaseModel を継承したクラス
@app.post("/items/")
async def create_item(item: Item):
    return item
```

---

## 1. なぜ必要か

型アノテーションだけでは「型が合っているか」しか検証できない。「1以上100以下」「3文字以上50文字以内」「特定のパターンに一致」といった**値の制約**を付けたいときに `Path()` / `Query()` / `Body()` を使う。

```python
from typing import Annotated
from fastapi import FastAPI, Body, Path, Query

app = FastAPI()


@app.put("/items/{item_id}")
async def read_item(
    item_id: Annotated[int, Path(ge=1, le=1000, description="商品ID")],
    q: Annotated[str | None, Query(min_length=3, max_length=50)] = None,
    importance: Annotated[int, Body(ge=1, le=5, description="重要度")] = 1,
):
    return {"item_id": item_id, "q": q, "importance": importance}
```

- `item_id` はパスの `{item_id}` と名前が一致しているので**パスパラメータ**、`q` はパスに無いので**クエリパラメータ**、`importance` は単純な型だが `Body()` が付いているので**ボディパラメータ**と判定される（この判定ルール自体は `Path()`/`Query()`/`Body()` の有無に関係ない。`Body()` については[5.](#5-body-で単純な値をボディに含める)で詳述）
- `Path()` / `Query()` / `Body()` は「どこから来るか」を変えるものではなく、**そこに来た値へ制約・メタデータを追加する**ためのもの

---

## 2. `Annotated` を使う書き方（推奨）

FastAPI公式は `Annotated[型, Path(...)/Query(...)]` の形を推奨している。デフォルト引数に `Path()`/`Query()` を直接書く旧方式は、`Depends`（→[APIRouterについて](APIRouterについて.md)）と同様に**デフォルト値との衝突**が起きやすく非推奨気味。

```python
# ✅ 推奨（Annotated）
async def read_item(item_id: Annotated[int, Path(ge=1)]):
    ...

# ⚠️ 動くが非推奨（デフォルト引数方式。デフォルト値が要る場合に書き方が崩れやすい）
async def read_item(item_id: int = Path(ge=1)):
    ...
```

---

## 3. `Query()` でよく使う制約

| パラメータ | 意味 | 対象型 |
|---|---|---|
| `min_length` / `max_length` | 文字列の最小/最大長 | `str` |
| `pattern` | 正規表現（旧: `regex`。新しいFastAPIでは`pattern`） | `str` |
| `ge` / `gt` / `le` / `lt` | 以上/より大きい/以下/より小さい | `int` / `float` |
| `default` | デフォルト値（`= None`と同義だが明示できる） | 任意 |
| `alias` | 実際のクエリキー名を別名にする | 任意 |
| `deprecated` | OpenAPI上で非推奨マーク | 任意 |
| `include_in_schema` | OpenAPIドキュメントに載せない | 任意 |
| `description` / `title` | OpenAPIの説明文 | 任意 |

```python
@app.get("/items/")
async def read_items(
    q: Annotated[
        str | None,
        Query(
            min_length=3,
            max_length=50,
            pattern="^fixedquery$",
            alias="item-query",   # ?item-query=... で受け取る
            description="検索クエリ文字列",
        ),
    ] = None,
):
    return {"q": q}
```

### 必須のクエリパラメータ

デフォルト値を付けなければ必須になる（ボディと同じルール）。`...`（`Ellipsis`）で明示的に必須と書くこともできる。

```python
q: Annotated[str, Query(min_length=3)]              # デフォルト値なし → 必須
q: Annotated[str, Query(min_length=3)] = ...         # 明示的に必須（意味は同じ）
```

### 複数値を受け取るクエリパラメータ（`list`）

```python
# /items/?tags=a&tags=b&tags=c
@app.get("/items/")
async def read_items(tags: Annotated[list[str], Query()] = []):
    return {"tags": tags}
```

---

## 4. `Path()` でよく使う制約

数値の範囲制約はクエリと同じ（`ge`/`gt`/`le`/`lt`）。パスパラメータは**URLの一部なので常に必須**（デフォルト値を持てない）。

```python
@app.get("/items/{item_id}")
async def read_item(
    item_id: Annotated[int, Path(title="商品のID", ge=1)],
):
    return {"item_id": item_id}
```

> [!NOTE]
> パスパラメータとクエリパラメータの**宣言順は自由**（Python的には「デフォルト値なし引数を先に」というルールがあるが、`Annotated`を使えばこの制約を気にしなくてよくなる。従来の `item_id: int = Path(...)` 方式だと引数順序で詰まることがあった）。

---

## 5. `Body()` で単純な値をボディに含める

通常、ボディはPydanticモデル1つで受けるが（→[PydanticのBaseModelを使用してRequest Body内のJSONパラメータを受け取る方法](PydanticのBaseModelを使用してRequest%20Body内のJSONパラメータを受け取る方法.md)）、`int` や `str` のような単純な型をボディの一部として受け取りたい場合は `Body()` で明示する。

```python
from fastapi import Body


@app.put("/items/{item_id}")
async def update_item(
    item_id: int,
    item: Item,
    importance: Annotated[int, Body()],   # 単純な型だが Body() を付けるとボディ扱いになる
):
    return {"item_id": item_id, "item": item, "importance": importance}
```

- `importance` は本来ならパスの `{item_id}` にも一致せず `BaseModel` でもないため**クエリパラメータ**と解釈されるが、`Body()` を付けることで**ボディの一部**として扱われる
- 複数のPydanticモデルを同時にボディで受けると、FastAPIは自動的にモデル名をキーにしたネスト構造を期待する

```python
# item: Item, user: User の場合、期待されるボディ:
{
  "item": {...},
  "user": {...}
}
```

- `Body(embed=True)` を単一モデルに付けると、`item` 単体でも `{"item": {...}}` の形でネストさせられる

---

## 6. `Field()` との違い（Pydanticモデルのフィールドに制約を付ける）

`Path()` / `Query()` / `Body()` は**エンドポイントの引数**に付けるもの。Pydanticモデルの**フィールド**に同様の制約を付けたいときは `pydantic.Field()` を使う。

```python
from pydantic import BaseModel, Field


class Item(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    price: float = Field(gt=0, description="0より大きい価格")
    tax: float | None = Field(default=None, le=100)
```

| 使う場所 | 対象 |
|---|---|
| `Path()` / `Query()` / `Body()` | Path Operation Functionの**引数** |
| `Field()` | Pydanticモデルの**フィールド** |

考え方は同じ（`ge`/`le`/`min_length`/`pattern`など多くのキーワードが共通）で、「どこに制約を書くか」が違うだけ。

---

## 7. `Enum` で選択肢を制限する

決まった値しか受け付けたくない場合は `Enum` を使う。範囲外の値は自動的に `422` になる。

```python
from enum import Enum


class ModelName(str, Enum):
    alexnet = "alexnet"
    resnet = "resnet"
    lenet = "lenet"


@app.get("/models/{model_name}")
async def get_model(model_name: ModelName):
    if model_name is ModelName.alexnet:
        return {"model_name": model_name, "message": "Deep Learning FTW!"}
    return {"model_name": model_name}
```

- `str` を多重継承させると（`class ModelName(str, Enum)`）、値がJSONにそのまま文字列としてシリアライズされる
- OpenAPI（Swagger UI）上では**選択式のドロップダウン**として表示される

---

## 8. パス末尾に `/` を含むパスパラメータ

ルートのパス定義で `{param}` の代わりに **`{param:path}`**（Starletteのパスコンバータ構文）を使うと、`/` を含む値もまとめて1つのパラメータとして受け取れる（デフォルトの `{param}` は `/` の手前までしかマッチしない）。これは `fastapi.Path()` 関数の引数ではなく、**ルートのパス文字列側**に書く指定である点に注意。

```python
@app.get("/files/{file_path:path}")
async def read_file(file_path: str):
    # /files/home/user/data.txt → file_path = "home/user/data.txt"
    return {"file_path": file_path}
```

---

## ポイントまとめ

- **どこから値が来るか**（パス/クエリ/ボディ）は型アノテーションの種類で決まる。`Path()`/`Query()`/`Body()` はそこに**制約・メタデータを追加**するためのもの
- 書き方は `Annotated[型, Path(...)/Query(...)/Body(...)]` が現行の推奨スタイル
- 数値は `ge`/`gt`/`le`/`lt`、文字列は `min_length`/`max_length`/`pattern` が基本
- パスパラメータは常に必須。クエリはデフォルト値の有無で必須/任意が決まる
- Pydanticモデルの**フィールド**に制約を付けるときは `Field()` を使う（エンドポイント引数の `Path()`/`Query()`/`Body()` とは別物）
- 決まった選択肢しか受け付けないなら `Enum`（`str, Enum` の多重継承でJSONにも自然に乗る）
- `{param:path}` でパス区切り `/` を含む値も1パラメータとして受け取れる
