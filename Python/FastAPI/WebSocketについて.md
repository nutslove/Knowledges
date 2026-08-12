# FastAPIのWebSocketについて

**WebSocket** は、クライアント⇔サーバー間で**双方向・持続的な接続**を維持できるプロトコル。[StreamingResponseについて](%20StreamingResponseについて.md)のSSEが「サーバー→クライアントの一方向」だったのに対し、WebSocketは**両方向にいつでも送受信できる**点が異なる。

> 出典: [FastAPI公式 - WebSockets](https://fastapi.tiangolo.com/advanced/websockets/)

---

## 1. SSE / HTTPポーリングとの違い

| 方式 | 通信の向き | 接続の持続性 | 向いている用途 |
|---|---|---|---|
| 通常のHTTP | クライアント→サーバー（都度リクエスト） | リクエストごとに完結 | 一般的なCRUD API |
| SSE（[StreamingResponse](%20StreamingResponseについて.md)） | サーバー→クライアントのみ | 持続（片方向） | LLMのトークンストリーミング、通知配信 |
| **WebSocket** | **双方向いつでも** | 持続 | チャット、リアルタイム共同編集、ゲーム、双方向の対話型エージェント |

- ブラウザからの利用のしやすさでは**SSEの方が単純**（`EventSource`で自動再接続もされる）。**双方向にメッセージを送りたい**（クライアントからも随時送信したい）ならWebSocketを選ぶ

---

## 2. 基本形

```python
from fastapi import FastAPI, WebSocket

app = FastAPI()


@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()          # ハンドシェイクを受け入れる（必須）
    while True:
        data = await websocket.receive_text()   # クライアントからのメッセージを待つ
        await websocket.send_text(f"echo: {data}")
```

- `@app.websocket(path)` はHTTPの`@app.get`等とは別のデコレータ
- `await websocket.accept()` を呼ばないとハンドシェイクが完了しない
- `receive_text()` / `receive_bytes()` / `receive_json()`、送信は `send_text()` / `send_bytes()` / `send_json()`
- クライアントが切断すると `receive_*()` が `WebSocketDisconnect` 例外を送出するので、`try/except`でループを抜ける

```python
from fastapi import WebSocketDisconnect


@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    try:
        while True:
            data = await websocket.receive_text()
            await websocket.send_text(f"echo: {data}")
    except WebSocketDisconnect:
        print("client disconnected")
```

---

## 3. パス・クエリパラメータ、`Depends`も使える

HTTPエンドポイントと同様に、パスパラメータやクエリパラメータ、[Depends](APIRouterについて.md)による依存性注入も使える。

```python
from fastapi import Depends, WebSocket, WebSocketException, Query, status


async def get_token(token: str = Query()) -> str:
    if token != "valid-token":
        # accept() より前に投げると、ハンドシェイク自体を指定コードで拒否できる
        raise WebSocketException(code=status.WS_1008_POLICY_VIOLATION, reason="Invalid token")
    return token


@app.websocket("/ws/{client_id}")
async def websocket_endpoint(
    websocket: WebSocket,
    client_id: int,
    token: str = Depends(get_token),
):
    await websocket.accept()
    await websocket.send_text(f"client_id={client_id}, token={token}")
```

> [!WARNING]
> 接続を**拒否**したい場合は `WebSocketDisconnect` ではなく **`WebSocketException`** を`raise`する（`WebSocketDisconnect`はクライアント側が切断したときにサーバー側が受け取る例外であり、サーバーから能動的に投げて拒否するためのものではない）。`WebSocketException`は`accept()`より前（依存関数の中など）で投げると、ハンドシェイクをその場で指定した`code`で終了できる。`Depends`内で`HTTPException`を投げても正しく処理されない点にも注意。
>
> ```python
> from fastapi import WebSocketException, status
> raise WebSocketException(code=status.WS_1008_POLICY_VIOLATION, reason="Invalid token")
> ```

> [!NOTE]
> WebSocketでの認証は「接続確立時（`accept()`前）にクエリパラメータやヘッダーでトークンを渡して検証する」のが一般的。ブラウザの`WebSocket` APIは任意のヘッダーを付けられないため、**トークンをクエリパラメータで渡す**実装がよく使われる（URLやログにトークンが残る点はセキュリティ上のトレードオフとして留意）。

---

## 4. 複数クライアントへブロードキャストする（接続管理）

チャットのように「複数の接続に同じメッセージを配信」したい場合、接続を保持する`ConnectionManager`を自前で用意するのが定番パターン。

```python
class ConnectionManager:
    def __init__(self):
        self.active_connections: list[WebSocket] = []

    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)

    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)

    async def broadcast(self, message: str):
        for connection in self.active_connections:
            await connection.send_text(message)


manager = ConnectionManager()


@app.websocket("/ws/{client_id}")
async def websocket_endpoint(websocket: WebSocket, client_id: int):
    await manager.connect(websocket)
    try:
        while True:
            data = await websocket.receive_text()
            await manager.broadcast(f"client#{client_id}: {data}")
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast(f"client#{client_id} left")
```

> [!WARNING]
> `ConnectionManager`をモジュールグローバルな変数（プロセス内メモリ）で持つ構成は、**Uvicornワーカーが複数（マルチプロセス）だと接続一覧がプロセスごとに分断される**。ワーカーAに繋いだクライアントへ、ワーカーBが受けたメッセージをブロードキャストできない。複数ワーカー/複数Podで動かす場合は、Redis Pub/Sub等の**外部のメッセージブローカーを介して配信**する構成が必要になる。

---

## 5. JSONでやり取りする

```python
@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    while True:
        data = await websocket.receive_json()      # {"type": "...", "payload": {...}} 等
        await websocket.send_json({"echo": data})
```

- Pydanticモデルでバリデーションしたい場合は、受け取ったdictを`SomeModel.model_validate(data)`のように自分で変換する（HTTPエンドポイントのような自動バリデーションはWebSocketには無い）

---

## 6. テストする

`TestClient`は`websocket_connect()`でWebSocketエンドポイントもテストできる（→[テスト（TestClient）について](テスト（TestClient）について.md)）。

```python
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)


def test_websocket():
    with client.websocket_connect("/ws") as websocket:
        websocket.send_text("hello")
        data = websocket.receive_text()
        assert data == "echo: hello"
```

---

## 7. LLMのストリーミング応答をSSEではなくWebSocketで行うべきか

[StreamingResponseについて](%20StreamingResponseについて.md)で扱ったLLMトークンストリーミングは、多くの場合SSEで十分（サーバー→クライアントの一方向で足りるため）。WebSocketを選ぶ判断基準は以下。

| 要件 | 適した方式 |
|---|---|
| サーバーからトークンを流すだけ、途中でクライアントから割り込まない | SSE（`StreamingResponse`） |
| ユーザーが生成中にメッセージを追加送信できる（会話の途中で割り込み・中断指示） | WebSocket |
| ブラウザの自動再接続や`Last-Event-ID`によるレジューム | SSE（`EventSource`が標準対応） |
| バイナリデータのやり取りが必要 | WebSocket（SSEはテキストのみ） |

---

## ポイントまとめ

- WebSocketは**双方向・持続的**な通信路。SSE（[StreamingResponse](%20StreamingResponseについて.md)）は片方向配信に強いが、クライアントからの随時送信にはWebSocketが必要
- `@app.websocket(path)` + `await websocket.accept()` が基本形。切断は`WebSocketDisconnect`を`try/except`で拾う
- パスパラメータ・クエリパラメータ・`Depends`はHTTPエンドポイントと同様に使える。認証はクエリパラメータでトークンを渡すのが一般的
- 複数クライアントへの配信には接続管理（`ConnectionManager`）が必要。**マルチワーカー/マルチPod構成ではプロセスをまたいだ配信にRedis等の外部ブローカーが必須**になる点に注意
- テストは`TestClient.websocket_connect()`で可能（→[テスト（TestClient）について](テスト（TestClient）について.md)）
