# FastAPIのRequest/Responseスキーマについて

- FastAPIでは`pydantic`の`BaseModel`を継承したクラスでRequest/Responseのスキーマ（データ構造）を定義する
- Requestスキーマ → Path Operation Functionの引数の型に指定する
- Responseスキーマ → デコレータの`response_model`引数（または戻り値の型アノテーション）に指定する
- これにより、FastAPIは自動的に以下を行ってくれる
  - **データのバリデーション**（型チェック・必須チェックなど）
  - **JSON ⇔ Pythonオブジェクトの相互変換**（シリアライズ／デシリアライズ）
  - **OpenAPI（Swagger UI / ReDoc）ドキュメントの自動生成**

---

## Requestスキーマ

- `BaseModel`を継承したクラスを定義し、Path Operation Functionの引数の型に指定する
- FastAPIはリクエストボディ（JSON）を読み取り、定義したモデルに自動でパースする
- データがモデル定義に合致しない場合は、自動的に`422 Unprocessable Entity`のエラーレスポンスを返す

```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class Item(BaseModel):
    name: str
    description: str | None = None  # オプショナル（デフォルト値あり）
    price: float
    tax: float | None = None

@app.post("/items/")
async def create_item(item: Item):
    return item
```

- 型のルール
  - デフォルト値を指定しないフィールドは**必須**（リクエストに含まれていないとエラー）
  - `| None = None`（または`Optional[str] = None`）を付けると**任意**になる

> [!NOTE]
> `create_item(item: Item)`の`item`は単なる**パラメータ名（変数名）なので任意**で、`body`でも`req_data`でも何でも構わない。
> 重要なのは **型アノテーション側（`: Item`）** で、FastAPIは「引数の型が`BaseModel`を継承したクラスかどうか」を見てリクエストボディとして解釈する。パラメータ名は判定に使われない。
>
> ```python
> # どれも同じ動作（リクエストボディとして解釈される）
> async def create_item(item: Item): ...
> async def create_item(body: Item): ...
> async def create_item(req_data: Item): ...
> ```
>
> ただし、以下のケースでは**パラメータ名がそのまま対応付けに使われる**ので注意。
>
> - **パスパラメータ** … パスの`{...}`内の名前と一致させる必要がある
>   ```python
>   @app.get("/items/{item_id}")
>   async def read_item(item_id: str):  # {item_id} と名前を合わせる必要あり
>       ...
>   ```
> - **クエリパラメータ** … パラメータ名がそのままクエリのキー名になる
>   ```python
>   @app.get("/items/")
>   async def read_items(skip: int = 0):  # ?skip=10 のキー名になる
>       ...
>   ```
>
> つまり「**ボディ（`BaseModel`）は型で判定 → 名前は自由**」「**パス/クエリは名前で判定 → 名前が意味を持つ**」という違いがある。

---

## Responseスキーマ（`response_model`）

- `response_model`にスキーマを指定すると、戻り値をそのモデルに従って**フィルタリング・変換**してからクライアントに返す
- Path Operation Functionの戻り値（dictやORMオブジェクトなど）が`response_model`で定義されていないフィールドを持っていても、**定義されたフィールドだけに絞られて返される**

```python
class UserIn(BaseModel):
    username: str
    password: str       # パスワードを受け取る
    email: str

class UserOut(BaseModel):
    username: str
    email: str          # passwordは含めない

@app.post("/user/", response_model=UserOut)
async def create_user(user: UserIn):
    # 戻り値にpasswordが含まれていても、response_model(UserOut)に
    # 従ってpasswordは除外されてレスポンスされる
    return user
```

- 上記のように、**パスワードなどの機密情報をレスポンスから除外する**といった用途に有効

### 戻り値の型アノテーションでも指定可能

- 最近のFastAPIでは、`response_model`引数の代わりに**戻り値の型アノテーション**でも指定できる（こちらが推奨）

```python
@app.post("/user/")
async def create_user(user: UserIn) -> UserOut:
    return user
```

- `response_model`引数と型アノテーションの両方を指定した場合は、`response_model`引数が優先される

#### FastAPIがやってくれること（効果）は両者で同じ

- **戻り値の型アノテーション**でも **`response_model`引数** でも、FastAPIが実行時に行うことは**まったく同じ**
  - データのバリデーション
  - 未定義フィールドのフィルタリング（除外）
  - JSONへのシリアライズ
  - OpenAPIドキュメントのスキーマ生成
- 違いは「**型チェッカー（mypy）／エディタから見たときの扱い**」だけ
  - 戻り値アノテーション（`-> UserOut`）… **戻り値の型として認識され、型チェックが効く**
  - `response_model=UserOut`引数 … FastAPI独自の引数なので**型チェッカーは認識しない**

##### 使い分け

- **実際の戻り値の型 = レスポンスの型** の場合 → 戻り値アノテーション（`-> UserOut`）が推奨。型チェックも効いて分かりやすい。
- **実際の戻り値の型 ≠ レスポンスの型** の場合 → `response_model`引数を使う（または戻り値を`-> Any`にする）。
  - 典型例が上の`UserIn`→`UserOut`。実際には`password`を持つ`UserIn`を返しているので、`-> UserOut`と書くと「違う型を返している」と**型チェッカーが警告する**。
  - そのため戻り値を`Any`にしてチェッカーを黙らせつつ、実際の変換は`response_model`でFastAPIに任せる。

```python
from typing import Any

@app.post("/user/", response_model=UserOut)
async def create_user(user: UserIn) -> Any:  # 実際にはUserInを返すのでAnyにする
    # FastAPIはresponse_model(UserOut)に従ってpasswordを除外してレスポンスする
    return user
```

> [!CAUTION]
> `StreamingResponse`（や`Response`、`JSONResponse`などの`Response`系オブジェクトを直接返す場合）では、**`response_model`は効かない**。
> `response_model`は「Path Operation Functionの戻り値をFastAPIが`BaseModel`に従ってシリアライズ・フィルタリングする」仕組みだが、`StreamingResponse`は**FastAPIを介さずに生のバイト列／チャンクをそのままクライアントへ流す**ため、バリデーションも変換もフィルタリングも行われない。
>
> ```python
> from fastapi.responses import StreamingResponse
>
> # たとえresponse_modelを書いても、ストリーム内容には一切適用されない
> @app.get("/stream", response_model=Item)  # ← Itemとして検証・変換はされない
> async def stream():
>     def gen():
>         yield b"chunk1"
>         yield b"chunk2"
>     return StreamingResponse(gen(), media_type="text/plain")
> ```
>
> - FastAPIは、戻り値が`Response`のサブクラス（`StreamingResponse`など）の場合は**自動的に`response_model`の処理をバイパス**する。
> - そのため、ストリーミングするデータの形式（JSON行、SSEなど）やスキーマの保証は**自分でコントロールする必要がある**。
> - 同様に、OpenAPIドキュメント上もレスポンススキーマが自動生成されないので、必要なら`responses`引数などで手動で記述する。
> - 戻り値の型アノテーションが`-> dict`などモデル型なのに実際は`Response`を返すケースでは、FastAPIが矛盾を検知してエラーになることがある。その場合は`response_model=None`を指定してモデル生成を明示的に無効化する。
>   ```python
>   @app.get("/stream", response_model=None)
>   async def stream() -> StreamingResponse | dict:
>       ...
>   ```
> - 参考：[StreamingResponseについて](StreamingResponseについて.md)

---

## `response_model`の主なパラメータ

- `response_model_exclude_unset=True`
  - **デフォルト値のまま（クライアントが明示的にセットしなかった）フィールドをレスポンスから除外**する
  - 実際に値がセットされたフィールドだけを返したい場合に使う

```python
class Item(BaseModel):
    name: str
    description: str | None = None
    price: float
    tax: float = 10.5

items = {
    "foo": {"name": "Foo", "price": 50.2},
}

@app.get("/items/{item_id}", response_model=Item, response_model_exclude_unset=True)
async def read_item(item_id: str):
    # description, taxはセットしていないので、レスポンスには
    # name, priceだけが含まれる
    return items[item_id]
```

- `response_model_include` / `response_model_exclude`
  - レスポンスに含める／除外するフィールドを個別に指定する
  - ただし、用途が限定的なら専用のモデルクラス（`UserOut`のような）を分けて定義する方が分かりやすい

```python
@app.get("/items/{item_id}", response_model=Item, response_model_exclude={"tax"})
async def read_item(item_id: str):
    return items[item_id]
```

---

## リストを返す場合

- `response_model`に`list[...]`を指定すると、複数件のレスポンスもバリデーション・変換される

```python
@app.get("/items/", response_model=list[Item])
async def read_items():
    return [
        {"name": "Foo", "price": 50.2},
        {"name": "Bar", "price": 62.0},
    ]
```

---

## モデルの継承で共通化する

- Request/Responseで共通するフィールドは、基底クラスにまとめて継承させると重複を減らせる

```python
class UserBase(BaseModel):
    username: str
    email: str

class UserIn(UserBase):
    password: str        # 入力時のみpasswordを追加

class UserOut(UserBase):
    pass                 # 出力はUserBaseのフィールドだけ

class UserInDB(UserBase):
    hashed_password: str # DB保存用
```

---

## ポイントまとめ

- **Request用とResponse用のモデルは分けて定義する**のが基本（入力と出力で必要なフィールドが異なるため）
- `response_model`を指定することで、**意図しないフィールドの漏洩を防ぎ、レスポンスのスキーマを保証**できる
- バリデーション・ドキュメント生成・型変換がすべて自動で行われるのがFastAPIの大きな利点
- 関連：[PydanticのBaseModelを使用してRequest Body内のJSONパラメータを受け取る方法](PydanticのBaseModelを使用してRequest%20Body内のJSONパラメータを受け取る方法.md)
