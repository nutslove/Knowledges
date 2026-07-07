# FastAPIのlifespan（起動・終了処理）について

**lifespan** は、アプリの**起動時に一度だけ実行する初期化**と、**終了時に一度だけ実行するクリーンアップ**を定義する仕組み。DBコネクションプール、LLMクライアント、MLモデルのロードなど「アプリの生存期間中ずっと使い回すリソース」の準備・後始末に使う。

> 出典: [FastAPI公式 - Lifespan Events](https://fastapi.tiangolo.com/advanced/events/)

---

## 1. 基本形（`@asynccontextmanager` + `yield`）

`contextlib.asynccontextmanager` でデコレートした関数を書き、`FastAPI(lifespan=...)` に渡す。`yield` の**前**が起動処理、**後**が終了処理。

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI

ml_models = {}


@asynccontextmanager
async def lifespan(app: FastAPI):
    # === 起動時（リクエスト受付を始める前に1回）===
    ml_models["answer"] = load_model()   # 重いモデルのロードなど
    yield
    # === 終了時（全リクエスト処理が終わった後に1回）===
    ml_models.clear()                    # 後始末


app = FastAPI(lifespan=lifespan)


@app.get("/predict")
async def predict(x: float):
    return {"result": ml_models["answer"](x)}
```

- `yield` の前 … サーバーが**リクエストを受け付け始める前**に実行される
- `yield` の後 … サーバーが**停止して全リクエストを処理し終えた後**に実行される
- `yield` 自体は値を返さなくてもよい（上の例のように `yield` だけでもOK）

> [!NOTE]
> `yield` を境に前処理／後処理を書くのは、依存性注入の `yield`（→[APIRouterについて](APIRouterについて.md) の「`yield` による後処理」）とよく似た発想。違いは、**依存の `yield` はリクエストごと**に対して、**lifespan の `yield` はアプリ全体で1回きり**という点。

---

## 2. なぜ使うのか（典型ユースケース）

「リクエストのたびに作ると重い／使い回したいオブジェクト」を起動時に1回だけ用意する。

| リソース | 起動時 | 終了時 |
|---|---|---|
| DBコネクションプール | プール生成 | プールをクローズ |
| HTTPクライアント（`httpx.AsyncClient`） | クライアント生成 | `await client.aclose()` |
| LLM / 外部APIクライアント | 認証・セッション確立 | セッション破棄 |
| MLモデル | メモリへロード | 解放 |
| キャッシュ（Redis等）接続 | 接続確立 | 切断 |

```python
import httpx
from contextlib import asynccontextmanager
from fastapi import FastAPI


@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.http = httpx.AsyncClient(timeout=30)   # 起動時に1回だけ生成
    yield
    await app.state.http.aclose()                    # 終了時にクローズ


app = FastAPI(lifespan=lifespan)
```

---

## 3. 作ったリソースを各エンドポイントから使う

起動時に生成したオブジェクトを共有する方法は主に3つ。

### (a) `app.state` に持たせる → `request.app.state` で参照

```python
@app.get("/proxy")
async def proxy(request: Request):
    client = request.app.state.http     # lifespanで作ったクライアント
    r = await client.get("https://example.com")
    return r.json()
```

### (b) `lifespan` で state を `yield` する → `request.state` で参照

`yield` に**dict**を渡すと、その中身が各リクエストの `request.state` から読めるようになる（Starletteのlifespan state機能）。

```python
@asynccontextmanager
async def lifespan(app: FastAPI):
    async with httpx.AsyncClient() as client:
        yield {"http": client}          # ← dictをyield


@app.get("/proxy")
async def proxy(request: Request):
    r = await request.state.http.get("https://example.com")
    return r.json()
```

### (c) モジュールグローバルに持たせる

```python
ml_models = {}   # モジュールトップレベル

# lifespan内で ml_models["x"] = ... して、各所から ml_models["x"] を参照
```

> [!TIP]
> テスト容易性・型の明確さでは `app.state` / yield state が扱いやすい。グローバル変数方式は手軽だが、依存の差し替え（テスト時のモック）がしにくくなりがち。

---

## 4. 旧方式 `@app.on_event` は非推奨

以前は起動・終了処理を次のように書いていたが、**現在は非推奨**。

```python
# ⚠️ 非推奨（deprecated）
@app.on_event("startup")
async def startup():
    ...

@app.on_event("shutdown")
async def shutdown():
    ...
```

> [!WARNING]
> `lifespan` を指定すると、`@app.on_event` のハンドラは**呼ばれなくなる**。公式には「**全部 `lifespan` か、全部 `on_event` のどちらか**（混在させない）」と明記されている。新規コードは `lifespan` に統一する。

移行は単純で、`startup` の処理を `yield` の前、`shutdown` の処理を `yield` の後にまとめるだけ。関連する準備と後始末を**1つの関数に近接して書ける**のが `lifespan` の利点。

---

## 5. 注意点

- **lifespanはメインアプリでのみ実行される**。`app.mount()` したサブアプリ（Sub Application / Mount）の lifespan は実行されない。
- 起動処理で例外が出るとアプリは起動しない（そのままクラッシュ）。外部依存の初期化失敗はここで早期に検知できる、とも言える。
- Uvicornなどの**ASGIサーバーがlifespanイベントを駆動**する（→[ASGIについて](ASGIについて.md)。ASGI仕様にlifespanプロトコルが含まれる）。

---

## ポイントまとめ

- 起動・終了処理は `@asynccontextmanager` で書いた関数を `FastAPI(lifespan=...)` に渡すのが現行の標準
- `yield` の**前＝起動時**、**後＝終了時**。アプリ全体で1回ずつ
- DBプール・HTTP/LLMクライアント・MLモデルなど「使い回すリソース」の準備／後始末に使う
- 共有は `app.state`（`request.app.state`）／`yield` した state（`request.state`）／モジュールグローバルのいずれか
- `@app.on_event("startup"/"shutdown")` は**非推奨**。`lifespan` に統一する（混在不可）
