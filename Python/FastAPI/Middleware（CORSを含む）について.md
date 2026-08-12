# FastAPIのMiddlewareについて

**Middleware（ミドルウェア）** は、**すべてのリクエスト/レスポンスに共通して割り込む処理**を書く仕組み。ルーティングが決まる前後で動作するため、「全エンドポイント共通の処理」をエンドポイントごとに書かずに済む。

> 出典: [FastAPI公式 - Middleware](https://fastapi.tiangolo.com/tutorial/middleware/) / [CORS](https://fastapi.tiangolo.com/tutorial/cors/)

---

## 1. `Depends` との違い

[APIRouterについて](APIRouterについて.md)の`Depends`と役割が似ているが、動作する層が異なる。

| 項目 | `Depends`（依存性注入） | Middleware |
|---|---|---|
| 適用単位 | エンドポイント／ルーター単位で選べる | **アプリ全体**（基本的に全リクエスト） |
| リクエストの内容を見て処理を分岐 | 得意（型注釈で必要な値だけ取り出せる） | 生の`Request`を自分でパースする必要がある |
| レスポンスの**後**に処理を挟む | `yield`を使えば可能（依存の後処理） | 得意（`call_next`の戻り値を触れる） |
| パスパラメータ・ボディの中身 | 自動で解決される | 基本的に扱いにくい（まだルーティング前） |
| 用途 | 認証、DBセッション、共通パラメータ | ロギング、CORS、圧縮、処理時間計測、リクエストID付与 |

> [!NOTE]
> 「特定のエンドポイントだけに認証をかけたい」なら `Depends`、「**全リクエスト**に対してヘッダーを付ける・ログを取る・圧縮する」なら Middleware、と使い分ける。

---

## 2. 自作Middleware（`@app.middleware("http")`）

```python
import time
from fastapi import FastAPI, Request

app = FastAPI()


@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    start = time.perf_counter()
    response = await call_next(request)          # ← ここで実際のPath Operationが実行される
    duration = time.perf_counter() - start
    response.headers["X-Process-Time"] = str(duration)
    return response
```

- `call_next(request)` の**前**に書いた処理はリクエスト受信直後に、**後**に書いた処理はレスポンス生成後に実行される
- `call_next` を呼ばない・例外を投げっぱなしにすると、後続のエンドポイントが実行されない／レスポンスが返らないので注意
- 複数登録した場合は**後から登録したものが外側**になる（玉ねぎのような入れ子構造）

```
リクエスト → [Middleware2] → [Middleware1] → Path Operation
                                                    ↓
レスポンス ← [Middleware2] ← [Middleware1] ← ------
```

### リクエストIDを付与する例（ロギングでよく使う）

```python
import uuid
from fastapi import FastAPI, Request

app = FastAPI()


@app.middleware("http")
async def add_request_id(request: Request, call_next):
    request_id = str(uuid.uuid4())
    request.state.request_id = request_id        # request.state経由でエンドポイント側からも参照可能
    response = await call_next(request)
    response.headers["X-Request-ID"] = request_id
    return response
```

- `request.state` に詰めた値は、そのリクエスト処理中どこからでも `request.state.xxx` で参照できる（[lifespanについて](lifespanについて.md)の`app.state`と対になる、**リクエスト単位**のスコープ）

> [!WARNING]
> `@app.middleware("http")` 内で例外が発生すると、後述の`@app.exception_handler`（→[エラーハンドリングについて](エラーハンドリングについて.md)）が効かないことがある（Middlewareはルーティングより外側で動くため）。Middleware内の処理は極力シンプルに保ち、例外を出さないようにするか、`try/except`で自前処理する。

> [!CAUTION]
> #### `@app.middleware("http")` / `BaseHTTPMiddleware` には既知の制限がある
>
> `@app.middleware("http")`は内部的に**`BaseHTTPMiddleware`**（`app.add_middleware(BaseHTTPMiddleware, dispatch=関数)`の糖衣構文）として動く。このクラスにはStarletteメンテナも認めている既知の問題があり、本番の複雑な用途では注意が必要。
>
> - **`contextvars`が正しく伝播しない**: `BaseHTTPMiddleware`はASGIとRequest/Responseベースの独自APIを橋渡しするために内部でタスクを分けて実行するため、エンドポイント側で`ContextVar`にセットした値が、外側のMiddlewareから見えない（またはその逆）ことがある
> - クライアントが接続を切断した際の挙動や、バックグラウンドタスクとの相互作用にも既知の不具合報告がある
>
> Starlette公式の結論は「**複雑な要件や本番のクリティカルパスでは、`BaseHTTPMiddleware`ではなく`ASGIApp`を直接実装する「pure ASGI middleware」を使うべき**」というもの。単純な処理時間計測やヘッダー付与程度なら`@app.middleware("http")`で十分実用的だが、`request.state`やコンテキスト変数に依存する高度な処理を挟む場合は、pure ASGI middleware（`__call__(self, scope, receive, send)`を実装する形）への書き換えを検討する。
> - 参考: [Starlette Discussion #1729 - Remaining bugs/limitations of BaseHTTPMiddleware](https://github.com/Kludex/starlette/discussions/1729)

---

## 3. クラスベースMiddleware（`add_middleware`）

Starlette由来の `BaseHTTPMiddleware` を使うクラス形式や、既製のMiddlewareクラスを登録する場合は `app.add_middleware()` を使う。

```python
from starlette.middleware.base import BaseHTTPMiddleware


class ProcessTimeMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        start = time.perf_counter()
        response = await call_next(request)
        response.headers["X-Process-Time"] = str(time.perf_counter() - start)
        return response


app.add_middleware(ProcessTimeMiddleware)
```

- `@app.middleware("http")` は内部的に `add_middleware(BaseHTTPMiddleware, dispatch=...)` の糖衣構文
- 既製Middleware（`CORSMiddleware`、`GZipMiddleware`など）は基本的に `add_middleware()` で登録する

---

## 4. CORS Middleware（最頻出）

**CORS (Cross-Origin Resource Sharing)** は、ブラウザが「別オリジン（ドメイン/ポート/プロトコルが異なる）へのリクエストを許可するか」を制御する仕組み。フロントエンド（例: `https://app.example.com`）とバックエンドAPI（例: `https://api.example.com`）のドメインが分かれている構成では**ほぼ必須**の設定。

```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://app.example.com"],   # 許可するオリジン一覧
    allow_credentials=True,                        # Cookie等の資格情報を許可
    allow_methods=["*"],                            # 許可するHTTPメソッド
    allow_headers=["*"],                            # 許可するリクエストヘッダー
)
```

| パラメータ | 意味 |
|---|---|
| `allow_origins` | 許可するオリジンのリスト（`["*"]`は全許可だが`allow_credentials=True`とは併用不可） |
| `allow_origin_regex` | 正規表現でオリジンを許可（サブドメインが可変な場合など） |
| `allow_credentials` | Cookie・Authorizationヘッダーなどを含むリクエストを許可するか |
| `allow_methods` | 許可するHTTPメソッド（`GET`, `POST`など） |
| `allow_headers` | 許可するリクエストヘッダー |
| `expose_headers` | JS側から読めるようにするレスポンスヘッダー（デフォルトは一部のみ） |

> [!WARNING]
> `allow_origins=["*"]`（全オリジン許可）と `allow_credentials=True` は**ブラウザの仕様上、同時に使えない**（セキュリティ上意味をなさなくなるため）。Cookie/認証情報を使うAPIでは、**具体的なオリジンを列挙**するか `allow_origin_regex` を使う。

> [!NOTE]
> CORSはブラウザ側が強制するルールであり、**サーバー間通信（curl、サーバーサイドのHTTPクライアント等）には無関係**。「Postmanでは通るのにブラウザからだけ失敗する」という症状はCORS設定漏れがほぼ確定原因。

---

## 5. その他の既製Middleware

```python
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.middleware.httpsredirect import HTTPSRedirectMiddleware

app.add_middleware(GZipMiddleware, minimum_size=1000)          # 一定サイズ以上のレスポンスをgzip圧縮
app.add_middleware(TrustedHostMiddleware, allowed_hosts=["example.com", "*.example.com"])  # Hostヘッダー検証
app.add_middleware(HTTPSRedirectMiddleware)                     # HTTP→HTTPSへ強制リダイレクト
```

- `TrustedHostMiddleware` はHostヘッダー偽装によるキャッシュポイズニング等の対策になる（ALB配下では別途ALB側の設定と役割分担を検討）
- ALB/Nginxなど手前のインフラ側で圧縮やTLS終端をすでにやっている場合、アプリ側のMiddlewareと**重複しないか確認**する（→[ASGIについて](ASGIについて.md)の本番構成）

---

## 6. Middlewareの登録順序に注意

`add_middleware()` は**後に登録したものほど外側（リクエストを最初に受け、レスポンスを最後に返す層）** になる（セクション2の「後から登録したものが外側」と同じルール）。CORSは**できるだけ外側**に置くのが定石なので、通常は**最後に**`add_middleware(CORSMiddleware, ...)`を呼ぶ。

```python
app.add_middleware(GZipMiddleware)
app.add_middleware(CORSMiddleware, allow_origins=[...])
# 実行順序（リクエスト時）: CORS → GZip → Path Operation
# 実行順序（レスポンス時）: Path Operation → GZip → CORS
```

---

## ポイントまとめ

- Middlewareは**全リクエストに共通する横断的処理**（ロギング、CORS、圧縮、処理時間計測など）を1箇所に書く仕組み。個別のエンドポイント制御は`Depends`が向いている
- 簡単なものは `@app.middleware("http")`、既製クラスや複雑なものは `add_middleware()`
- `call_next(request)` の前後でリクエスト前処理・レスポンス後処理を挟める。複数登録すると後登録ほど外側になる
- **CORS設定はフロント/バックのオリジンが分かれる構成でほぼ必須**。`allow_origins=["*"]`と`allow_credentials=True`は併用不可な点に注意
- CORSエラーはブラウザだけで発生する（サーバー間通信では起きない）ので切り分けの参考にする
- `@app.middleware("http")`（`BaseHTTPMiddleware`）は`contextvars`の伝播などに既知の制限がある。単純な用途では実用上問題ないが、コンテキスト変数に依存する複雑な処理を挟むなら pure ASGI middleware への書き換えを検討する
