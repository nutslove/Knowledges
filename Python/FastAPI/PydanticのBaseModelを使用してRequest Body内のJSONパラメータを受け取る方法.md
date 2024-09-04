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

- `BaseModel`を継承しているクラス型の引数（上記だと`item: Item`）を初期化するときは、`BaseModel`を継承しているクラス型で初期化する
  - 例（`reqData: logaas.LOGaaSClientBase = logaas.LOGaaSClientBase()`の部分）  
    ```python
    from typing import Optional

    class LOGaaSClientBase(BaseModel):
    cluster_type: Optional[str] = default_cluster_type
    opensearch_version: Optional[str] = default_opensearch_version
    opensearch_dashboard_version: Optional[str] = default_opensearch_dashboard_version
    scale_size: Optional[int] = default_scale_size
    ```

    ```python
    import logaas

    @logaas_router.post("/", response_class=JSONResponse)
    @keystone.token_validation_check
    async def logaas_create(request: Request, token: str = None, project_id: str = None, is_admin: bool = False, reqData: logaas.LOGaaSClientBase = logaas.LOGaaSClientBase()):
      cluster_id = utilities.get_uuid()
      client = logaas.LOGaaSClient(
        project_id,
        cluster_id,
        reqData.cluster_type,
        reqData.opensearch_version,
        reqData.opensearch_dashboard_version,
        reqData.scale_size
      )
    ```
