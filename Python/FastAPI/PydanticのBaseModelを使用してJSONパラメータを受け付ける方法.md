- `pydantic`の`BaseModel`を使って受け付けるパラメータの属性をclassで定義し、FastAPIのPath Operation Functionの引数の型に定義したclassを指定するだけ
- 例
  - FastAPI側
    ```python
    from fastapi import FastAPI
    from pydantic import BaseModel

    app = FastAPI()

    class Item(BaseModel):
        name: str
        description: str | None = None
        price: float
        tax: float | None = None

    @app.post("/items/")
    async def create_item(item: Item):
        return item

    @app.get("/")
    async def root():
        return {"message": "Hello World"}
    ```
  - クライアントの例
    ```shell
    curl -X 'POST' \
    'http://localhost:8000/items/' \
    -H 'accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
    "name": "Foo",
    "description": "An optional description",
    "price": 45.2,
    "tax": 3.5
    }'
    ```
- JSONデータは自動的に`Item`モデルにパースされる。  
  また、データが`Item`モデルの定義に合致しない場合、FastAPIは自動的に適切なエラーレスポンスを返す。
- FastAPIは自動的に以下を行う
  - リクエストボディを読み取る
  - JSONをPythonデータに変換する
  - データを検証する（型チエックなど）
  - Path Operation Functionのパラメータ(上記の例では`item`)にデータを渡す