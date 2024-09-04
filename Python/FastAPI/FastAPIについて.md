## Requestオブジェクトについて
- Requestオブジェクトは、以下のようなHTTPリクエストに関する情報を含んでいる
  - クエリーパラメータ
  - Header
  - Cookie
  - Body（JSON、Formデータなど）
  - クライアントに関するデータ（ホスト名やIPアドレスなど）
  - HTTPメソッド（GET、POSTなど）
  - URLパス
- Path Operation Functionの引数に`Request`パラメータを追加することで使用できる  
  ```python
    from fastapi import FastAPI, Request

    app = FastAPI()

    @app.get("/info")
    async def get_request_info(request: Request):
        print(await request.json())  # JSON形式のボディを取得（request.json()はコルーチン（非同期関数）であるため、awaitキーワードが必要）
        print(request.headers.get("User-Agent"))  # ヘッダーから特定の値を取得

        return {
            "method": request.method,
            "url": str(request.url),
            "headers": dict(request.headers),
            "query_params": dict(request.query_params),
            "client": request.client.host,
        }
  ```

## `JSONResponse`について
- JSONでレスポンスを返す時、Path Operation Functionに`response_class=JSONResponse`を指定して`return`では普通のjson形式で返す方法と、Path Operation Functionには`response_class=JSONResponse`を指定せず、`return JSONResponse()`で返す方法がある
- Path Operation Functionに`response_class=JSONResponse`を指定して、`return`では普通のjson形式で返す例
  - `status_code`の指定はできない  
  ```python
  from fastapi.responses import HTMLResponse, JSONResponse

  @logaas_router.get("/", response_class=JSONResponse)
  @keystone.token_validation_check
  async def logaas_status(request: Request, token: str = None, project_id: str = None, is_admin: bool = False):
    client = logaas.LOGaaSClient(project_id=project_id)
    try:
      response = client.get_logaas_status()
    except Exception as e:
      return {
        "messages": f"Getting OpenSearch Cluster information failed: {e}"
      }
    else:
      return response
  ```
- Path Operation Functionには`response_class=JSONResponse`を指定せず、`return JSONResponse()`で返す例
  - `status_code`の指定ができる
  - **`content`の指定は必須**  
  ```python
  from fastapi.responses import HTMLResponse, JSONResponse

  @logaas_router.get("/", response_class=JSONResponse)
  @keystone.token_validation_check
  async def logaas_status(request: Request, token: str = None, project_id: str = None, is_admin: bool = False):
    client = logaas.LOGaaSClient(project_id=project_id)
    try:
      response = client.get_logaas_status()
    except Exception as e:
      return JSONResponse(status_code=500, content={
        "messages": f"Getting OpenSearch Cluster information failed: {e}"
      })
    else:
      return JSONResponse(status_code=200, content=response)
  ```