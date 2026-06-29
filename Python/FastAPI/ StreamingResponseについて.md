## StreamingResponse とは

**`StreamingResponse`** は、FastAPI（実体はStarlette）が提供するレスポンスクラスの一つで、
**レスポンスボディを一度に全部メモリに乗せず、少しずつ（チャンク単位で）クライアントへ送る**ための仕組み。

- ボディの中身を **iterator / generator**（同期・非同期どちらも可）として渡す
- サーバーは生成された分から順に送出するため、**全データが揃う前にレスポンスを開始**できる
- HTTP/1.1では **chunked transfer encoding**（`Transfer-Encoding: chunked`）として送られる
  - HTTP/2・HTTP/3には`Transfer-Encoding`は存在せず、プロトコル固有のフレーム（DATAフレーム）で逐次送られる（chunkedは使わない）

```python
from fastapi import FastAPI
from fastapi.responses import StreamingResponse

app = FastAPI()

def number_generator():
    for i in range(10):
        yield f"{i}\n"

@app.get("/stream")
def stream():
    return StreamingResponse(number_generator(), media_type="text/plain")
```

## なぜ必要か（通常のレスポンスとの違い）

通常の `return {...}` や `JSONResponse` は、**レスポンスボディ全体をメモリ上に組み立ててから**一気に送る。
そのため次のようなケースで困る。

| 課題 | 通常レスポンス | StreamingResponse |
|------|----------------|-------------------|
| 巨大ファイル/大量データ | 全部メモリに載せる → OOMの危険 | チャンクごとに送る → メモリ一定 |
| 生成に時間がかかる（LLM等） | 全部生成し終わるまで待たせる | 生成できた分から即送る |
| 体感速度（TTFB） | 遅い（完成まで何も返らない） | 速い（最初のチャンクですぐ返り始める） |

> [!NOTE]
> #### chunked transfer encoding とは
>
> ボディ全体のサイズ（`Content-Length`）が事前に分からない場合に、
> データを「チャンク（塊）」に区切って順次送るHTTPの仕組み。
> `StreamingResponse` は基本的に `Content-Length` を付けず、`Transfer-Encoding: chunked` で送る。

## 引数

```python
StreamingResponse(
    content,                     # iterator / async generator / 同期generator など
    status_code=200,
    headers=None,
    media_type=None,             # 例: "text/plain", "application/json", "text/event-stream"
    background=None,             # BackgroundTask
)
```

- `content`: **iterable**（`bytes` または `str` をyieldする）。`str`はUTF-8でエンコードされる
  - 同期iterableを渡すと、Starletteが内部で `iterate_in_threadpool()` を使い**スレッドプール上で**回す（イベントループを塞がないため）
- `media_type`: `Content-Type` ヘッダーになる。**指定しない（`None`）と`Content-Type`が付かない**ので明示が無難。text系のmedia_typeを指定すると `charset=utf-8` が自動付与される

## 同期 generator と 非同期 generator

両方サポートされる。**重い同期I/Oをそのまま回すとイベントループをブロックする**点に注意。

```python
# 同期 generator（FastAPIが別スレッドで回す）
def sync_gen():
    for chunk in read_big_file():
        yield chunk

# 非同期 generator（async I/Oと相性が良い）
async def async_gen():
    async for chunk in fetch_from_upstream():
        yield chunk

@app.get("/a")
def a():
    return StreamingResponse(async_gen())
```

> [!TIP]
> generator内で `time.sleep()` のようなブロッキング処理を使うと、
> 非同期generatorの場合はイベントループ全体が止まる。`await asyncio.sleep()` を使うこと。

## 主なユースケース

### 1. 大きなファイルのダウンロード

```python
@app.get("/download")
def download():
    def iterfile():
        with open("large.zip", "rb") as f:
            while chunk := f.read(1024 * 1024):  # 1MBずつ
                yield chunk
    headers = {"Content-Disposition": 'attachment; filename="large.zip"'}
    return StreamingResponse(
        iterfile(), media_type="application/octet-stream", headers=headers
    )
```

> [!NOTE]
> 単にローカルファイルを返すだけなら **`FileResponse`** の方が簡潔（`Content-Length`等も自動設定）。
> `StreamingResponse` は「動的に生成しながら流す」用途で本領を発揮する。

### 2. LLM / ChatのトークンをリアルタイムにストリーミングするSSE

LLMの出力を1トークンずつ流すのは典型的なユースケース。
**SSE (Server-Sent Events)** の形式 (`media_type="text/event-stream"`) で送ることが多い。

```python
import json

async def sse_gen():
    async for token in llm_stream():
        # SSEは "data: ...\n\n" の形式
        yield f"data: {json.dumps({'token': token})}\n\n"
    yield "data: [DONE]\n\n"

@app.get("/chat")
async def chat():
    return StreamingResponse(
        sse_gen(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "X-Accel-Buffering": "no",  # Nginxのバッファリング無効化
        },
    )
```

> [!WARNING]
> #### `Connection: keep-alive` ヘッダーは手動で付けない
>
> 多くのSSEチュートリアルは `"Connection": "keep-alive"` を付けているが、
> `Connection` は **hop-by-hop ヘッダー**で **HTTP/2・HTTP/3では送信禁止**。
> 手動で付けると **Safari はHTTP/2でレスポンスを拒否**する（Chrome/Firefoxは無視）。
> HTTP/1.1ではkeep-aliveがそもそもデフォルトなので、基本的に**付けないのが正解**。
> 接続維持はASGIサーバー（Uvicorn等）が管理する。

> [!NOTE]
> #### SSE形式のルール
> - 1イベントは `data: <内容>\n\n`（末尾は**空行＝改行2つ**）で区切る
> - イベント名は `event: <name>`、ID は `id: <id>` で付与可能
> - 1メッセージを複数行に分けたい場合は `data:` 行を複数並べる（受信側で改行連結される）
> - 接続維持のため、無通信時に**コメント行 `: ping\n\n`** を定期送信することが多い（heartbeat）
> - ブラウザ側は `EventSource` API で受け取れる（ただし `EventSource` は **GETのみ**・カスタムヘッダー不可）
> - 双方向通信が必要なら SSE ではなく WebSocket を使う

## クライアント側の受け取り（参考）

```python
import httpx

with httpx.stream("GET", "http://localhost:8000/stream") as r:
    for chunk in r.iter_text():
        print(chunk, end="")
```

## 注意点・ハマりどころ

| 注意点 | 内容 |
|--------|------|
| バッファリング | Nginx等のリバースプロキシがバッファすると即時性が消える。`X-Accel-Buffering: no` 等で抑制 |
| エラー処理 | レスポンス送信開始後（ヘッダー送出後）にgenerator内で例外が出ても、ステータスコードは変えられない |
| `Content-Length` | 基本付かない。事前にサイズが分かるなら headers で明示することも可能 |
| ブロッキング | async generator内でのブロッキングI/OはイベントループをStallさせる |
| 接続切断 | クライアント切断後もgeneratorが回り続けるとリソースを無駄に消費する場合がある |

## 関連レスポンスクラスとの比較

| クラス | 用途 |
|--------|------|
| `JSONResponse` | 通常のJSON（全体を一括送信） |
| `FileResponse` | 静的ファイルの送信（`Content-Length`等を自動設定） |
| `StreamingResponse` | 動的に生成しながらチャンク送信・SSE・巨大データ |
| `Response` | 任意の生バイト/文字列を返す汎用クラス |
