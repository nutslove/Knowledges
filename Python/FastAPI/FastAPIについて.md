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