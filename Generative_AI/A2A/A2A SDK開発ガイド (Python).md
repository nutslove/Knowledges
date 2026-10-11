# はじめに

- 対象: **Python 公式 SDK `a2a-sdk` 1.2.2** (2026-10-05 リリース、Python 3.10 以上) で A2A のサーバ/クライアントを作る人。
- 概念 (Agent Card / Task / Message / 通信パターン / 認可の考え方など) は [A2Aについて.md](A2Aについて.md) を参照。**このノートは SDK の使い方だけ**を書く。
- 本ノートのコードは、`a2a-sdk` 1.2.2 を実際にインストールし、サーバを起動して動作を確認したもの。確認していないものは「未確認」と明記している。
- v0.3 向けの古い記事 (`A2AStarletteApplication` / `A2AClient` / `MessageSendParams` など) は 1.x では動かない。v0.3 互換は `a2a.compat.v0_3` と `enable_v0_3_compat=True` を使う。

---

# 1. インストール

```bash
pip install "a2a-sdk[http-server]"   # サーバ (Starlette + SSE) とクライアント
```

| extra | 入るもの | いつ必要か |
|---|---|---|
| (なし) | クライアント、型、ヘルパー、`opentelemetry-api` など | クライアントだけ作る |
| `http-server` | `starlette`, `sse-starlette` | **サーバを作る (JSON-RPC / REST)** |
| `fastapi` | `fastapi` ほか | 既存の FastAPI アプリに A2A のルートを足す |
| `grpc` | `grpcio`, `grpcio-tools`, `grpcio-reflection` | gRPC で公開/接続する |
| `sql` / `sqlite` / `postgresql` / `mysql` | `sqlalchemy` と各ドライバ | `DatabaseTaskStore` でタスクを永続化する |
| `encryption` | `cryptography` | DB に保存する Push 通知設定を暗号化する |
| `signing` | `pyjwt` | Agent Card の署名・検証 |
| `telemetry` | `opentelemetry-sdk` | トレースを出す |
| `db-cli` | `alembic` | `a2a-db` コマンドで DB マイグレーション |
| `all` | 上記すべて | 試すとき |

- サーバを動かすには別途 ASGI サーバが必要 (`pip install uvicorn`)。
- 本ノートの動作確認: `[http-server]` だけの環境で、クライアント・サーバ・ヘルパー・型が import できることを確認済み。`DatabaseTaskStore` は `sqlalchemy` が無いと使えない (`[sqlite]` などを足す)。

---

# 2. 全体像 (どのクラスが何をするか)

```
[クライアント]                         [サーバ]
 create_client() ──HTTP/gRPC──▶  Starlette(routes)
   Client                           ├ create_agent_card_routes   … /.well-known/agent-card.json
   ClientConfig                     └ create_jsonrpc_routes / create_rest_routes
   Interceptor                              │ context_builder で ServerCallContext を作る
                                            ▼
                                   DefaultRequestHandler(agent_executor, task_store, agent_card, ...)
                                            │ RequestContext + EventQueue を渡す
                                            ▼
                                   AgentExecutor.execute(context, event_queue)   ← 自分で書く
                                            │ event_queue.enqueue_event(...) / TaskUpdater
                                            ▼
                                   TaskStore (InMemory / Database)
```

**自分で書くのは基本的に `AgentExecutor` (中身の処理) と `AgentCard` (名刺) だけ。** それ以外は SDK が用意している。

| 役割 | クラス / 関数 | 場所 |
|---|---|---|
| 処理を実装する | `AgentExecutor` (`execute` / `cancel`) | `a2a.server.agent_execution` |
| 処理に渡される入力 | `RequestContext` | 同上 |
| 結果を返す出口 | `EventQueue` (`enqueue_event`) | `a2a.server.events` |
| 状態更新の簡易API | `TaskUpdater` | `a2a.server.tasks` |
| リクエスト処理の本体 | `DefaultRequestHandler` | `a2a.server.request_handlers` |
| タスクの保存先 | `InMemoryTaskStore` / `DatabaseTaskStore` | `a2a.server.tasks` |
| ルート生成 | `create_agent_card_routes` / `create_jsonrpc_routes` / `create_rest_routes` / `add_a2a_routes_to_fastapi` | `a2a.server.routes` |
| 型 (protobuf) | `AgentCard` / `Message` / `Part` / `Task` ... | `a2a.types` |
| 生成ヘルパー | `new_task_from_user_message` / `new_text_message` ... | `a2a.helpers` |
| クライアント | `create_client` / `ClientConfig` | `a2a.client` |

> [!NOTE]
> 型 (`AgentCard` など) は **protobuf のメッセージ**。`dict` ではなくキーワード引数で作り、`HasField()` / `WhichOneof()` / `CopyFrom()` で扱う。Enum の値は数値で返るので、名前にしたいときは `TaskState.Name(state)` を使う。

---

# 3. 最小のサーバ (コピペで動く)

```python
# server.py
from starlette.applications import Starlette

from a2a.helpers import new_task_from_user_message
from a2a.server.agent_execution import AgentExecutor, RequestContext
from a2a.server.events import EventQueue
from a2a.server.request_handlers import DefaultRequestHandler
from a2a.server.routes import create_agent_card_routes, create_jsonrpc_routes
from a2a.server.tasks import InMemoryTaskStore, TaskUpdater
from a2a.types import AgentCapabilities, AgentCard, AgentInterface, AgentSkill, Part


class EchoExecutor(AgentExecutor):
    async def execute(self, context: RequestContext, event_queue: EventQueue) -> None:
        # 初回は Task がまだ無いので、先に Task を積む (4章参照)
        task = context.current_task or new_task_from_user_message(context.message)
        if not context.current_task:
            await event_queue.enqueue_event(task)

        updater = TaskUpdater(event_queue, task.id, task.context_id)
        text = context.get_user_input()
        await updater.complete(updater.new_agent_message([Part(text=f"echo: {text}")]))

    async def cancel(self, context: RequestContext, event_queue: EventQueue) -> None:
        raise NotImplementedError  # 実装例は 6章


agent_card = AgentCard(
    name="Echo Agent",
    description="受け取ったテキストをそのまま返す",
    version="1.0.0",
    supported_interfaces=[
        AgentInterface(url="http://localhost:9999/", protocol_binding="JSONRPC", protocol_version="1.0"),
    ],
    capabilities=AgentCapabilities(streaming=True),
    default_input_modes=["text/plain"],
    default_output_modes=["text/plain"],
    skills=[AgentSkill(id="echo", name="Echo", description="テキストを返す", tags=["demo"], examples=["hello"])],
)

handler = DefaultRequestHandler(
    agent_executor=EchoExecutor(),
    task_store=InMemoryTaskStore(),
    agent_card=agent_card,
)

app = Starlette(
    routes=create_agent_card_routes(agent_card=agent_card)
    + create_jsonrpc_routes(request_handler=handler, rpc_url="/")
)
```

```bash
uvicorn server:app --port 9999
curl http://localhost:9999/.well-known/agent-card.json
```

ポイント:
- `supported_interfaces[].url` は**クライアントが実際にアクセスする公開URL**にする (ポートやパスを変えたら合わせる)。
- `rpc_url="/"` は JSON-RPC を受けるパス。`AgentInterface.url` のパス部分と一致させる。
- `AgentInterface.protocol_version="1.0"` を明示しておく (v0.3 と 1.0 を併記したい場合は要素を並べる)。

---

# 4. `AgentExecutor` を書く

## 4.1 `execute()` が受け取るもの

### `RequestContext`

| 属性 / メソッド | 内容 |
|---|---|
| `message` | クライアントが送った `Message` |
| `get_user_input(delimiter="\n")` | `message` の text Part を連結した文字列 |
| `current_task` | 既存の `Task`。**初回は `None`**、`input-required` からの再開などでは入っている |
| `task_id` / `context_id` | リクエストの ID (初回はサーバが採番) |
| `configuration` | `SendMessageConfiguration` (`history_length` など) |
| `metadata` | リクエストの `metadata` (dict) |
| `tenant` | マルチテナントのキー |
| `requested_extensions` | クライアントが `A2A-Extensions` で指定した拡張URIの集合 |
| `call_context` | `ServerCallContext` (認証済みユーザー等。7章) |
| `related_tasks` | 関連 Task のリスト |

### `EventQueue` と `enqueue_event`

- `EventQueue` は、Executor が処理結果をフレームワークへ送る**出口**。SDK が作って渡してくる (自分では作らない)。
- Executor から使えるのは `enqueue_event()` だけ。積まれたイベントはフレームワークが取り出し、`SendMessage` なら最終結果として、`SendStreamingMessage` なら SSE としてクライアントに返す。
- 積めるのは4種類で、クライアントが受け取る `StreamResponse` の4フィールドに対応する。

| 積むイベント | クライアント側 (`StreamResponse`) |
|---|---|
| `Message` | `message` |
| `Task` | `task` |
| `TaskStatusUpdateEvent` | `status_update` |
| `TaskArtifactUpdateEvent` | `artifact_update` |

### 積む順序のルール (実機確認済み)

| Executor の動き | 結果 |
|---|---|
| `Task` を積まずに状態更新 (`start_work()` など) を積む | `InvalidAgentResponseError: Agent should enqueue Task before TaskStatusUpdateEvent event` |
| `Message` だけを積む | OK。クライアントには `message` が返り、Task は作られない |
| `Task` を積んだあとに `Message` を積む | `InvalidAgentResponseError: Received Message object in task mode` |
| `Task` → 状態更新 → ... と積む | OK |
| 終端状態にしたあとに `TaskUpdater` で更新する | `RuntimeError: Task ... is already in a terminal state.` |

**「`Message` だけで返す」か「`Task` で状態管理する」かのどちらか一方を選ぶ。** Task を使うなら、初回 (`context.current_task is None`) に最初の1回だけ `Task` を積む。

```python
task = context.current_task or new_task_from_user_message(context.message)
if not context.current_task:
    await event_queue.enqueue_event(task)
```

## 4.2 `TaskUpdater`

`enqueue_event()` を呼ぶ薄いラッパー。コンストラクタは `TaskUpdater(event_queue, task_id, context_id)`。

| メソッド | 状態 | 備考 |
|---|---|---|
| `submit(message=None)` | `submitted` | |
| `start_work(message=None)` | `working` | |
| `requires_input(message=None)` | `input-required` | 5章 |
| `requires_auth(message=None)` | `auth-required` | |
| `complete(message=None)` | `completed` | 終端 |
| `failed(message=None)` | `failed` | 終端。実行中のエラー |
| `reject(message=None)` | `rejected` | 終端。実行しないと判断して断る |
| `cancel(message=None)` | `canceled` | 終端 |
| `update_status(state, message=None, timestamp=None, metadata=None)` | 任意 | 上記の元になる汎用版 |
| `add_artifact(parts, artifact_id=None, name=None, metadata=None, append=None, last_chunk=None, extensions=None)` | - | 成果物 (`TaskArtifactUpdateEvent`) |
| `new_agent_message(parts)` | - | `role=ROLE_AGENT` の `Message` を作る |

- 終端状態にしたあとにさらに更新すると `RuntimeError` になる (`TaskUpdater` 側で拒否される)。

## 4.3 エラーの扱い (`failed` と `reject`)

```python
async def execute(self, context, event_queue):
    task = context.current_task or new_task_from_user_message(context.message)
    if not context.current_task:
        await event_queue.enqueue_event(task)
    u = TaskUpdater(event_queue, task.id, task.context_id)

    if not self.is_supported(context.get_user_input()):
        await u.reject(u.new_agent_message([Part(text="対応範囲外です")]))   # 実行しない
        return

    await u.start_work()
    try:
        result = await self.do_work(context)
    except Exception as e:
        await u.failed(u.new_agent_message([Part(text=f"失敗: {e}")]))      # 実行したが失敗
        return
    await u.complete(u.new_agent_message([Part(text=result)]))
```

- `reject` / `failed` のどちらも、クライアントには `status_update` (または最終 `task`) として届く。確認済み: `TASK_STATE_REJECTED` / `TASK_STATE_FAILED` と、渡した `Message` の本文が届く。
- `execute()` 内で捕捉されない例外が出た場合の挙動は、本ノートでは未確認。必ず自分で `failed()` に変換する書き方にしておく。

---

# 5. 状態管理のパターン

## 5.1 ストリーミングで成果物を少しずつ返す

```python
await u.start_work()
aid = "answer-1"
chunks = ["Hel", "lo ", "world"]
for i, chunk in enumerate(chunks):
    await u.add_artifact(
        [Part(text=chunk)],
        artifact_id=aid,            # 同じ artifact_id で連結する
        name="answer",
        append=(i > 0),             # 2個目以降は追記
        last_chunk=(i == len(chunks) - 1),
    )
await u.complete()
```

クライアント (`ClientConfig(streaming=True)` が既定) には次の順で届く (確認済み):

```
task → status_update(WORKING) → artifact_update('Hel', append=False, last_chunk=False)
     → artifact_update('lo ', append=True) → artifact_update('world', append=True, last_chunk=True)
     → status_update(COMPLETED)
```

- `ClientConfig(streaming=False)` にすると、最終的な `Task` が1件だけ届き、`task.artifacts[0].parts` には 'Hel' 'lo ' 'world' が連結されて入る (確認済み)。
- サーバ側は `AgentCapabilities(streaming=True)` を宣言しておく。`False` / 未設定のサーバでは、クライアントが既定 (`streaming=True`) のままでも、**自動で非ストリーミングに切り替わって成功する** (確認済み: `task` が1件届く)。ただし仕様上、`streaming` 非対応のエージェントに `SendStreamingMessage` / `SubscribeToTask` を直接送ると `UnsupportedOperationError` を返すべき (MUST、仕様書 §3.3.4)。

## 5.2 `input-required` で追加入力をもらう (往復)

```python
async def execute(self, context, event_queue):
    if context.current_task is None:
        # 1回目: Task を作って、質問を返して中断
        task = new_task_from_user_message(context.message)
        await event_queue.enqueue_event(task)
        u = TaskUpdater(event_queue, task.id, task.context_id)
        await u.requires_input(u.new_agent_message([Part(text="どの都市ですか？")]))
    else:
        # 2回目: 同じ Task に回答が届いた
        t = context.current_task
        u = TaskUpdater(event_queue, t.id, t.context_id)
        await u.start_work()
        city = context.get_user_input()
        await u.complete(u.new_agent_message([Part(text=f"{city} の天気は晴れ")]))
```

クライアントは、**2回目の `Message` に `task_id` と `context_id` を付けて**送る:

```python
t1 = ...  # 1回目の結果の Task (state=INPUT_REQUIRED)
msg2 = Message(role=Role.ROLE_USER, parts=[Part(text="東京")], message_id=uuid4().hex,
               task_id=t1.id, context_id=t1.context_id)
```

確認済みの結果: 1回目 `INPUT_REQUIRED` + 質問文、2回目 `COMPLETED` + 回答、`history` は3件 (ユーザー → エージェント質問 → ユーザー回答 ...)。

- **終端状態の Task に送る**と `UnsupportedOperationError: Task ... is in terminal state: TASK_STATE_COMPLETED` になる。

## 5.3 長時間タスクを「すぐ返す」

クライアントが `SendMessageConfiguration(return_immediately=True)` を付けると、Task 作成直後に `submitted` で返り、処理は裏で続く。あとから `get_task` / `subscribe` で追う (6章)。

### ポーリングの実装 (クライアント)

```python
import asyncio
from a2a.types import GetTaskRequest, SendMessageConfiguration, SendMessageRequest, TaskState

TERMINAL = {TaskState.TASK_STATE_COMPLETED, TaskState.TASK_STATE_FAILED,
            TaskState.TASK_STATE_CANCELED, TaskState.TASK_STATE_REJECTED}
INTERRUPTED = {TaskState.TASK_STATE_INPUT_REQUIRED, TaskState.TASK_STATE_AUTH_REQUIRED}


async def start_task(client, msg):
    req = SendMessageRequest(message=msg, configuration=SendMessageConfiguration(return_immediately=True))
    async for ev in client.send_message(req):
        if ev.WhichOneof("payload") == "task":
            return ev.task                      # state は SUBMITTED


async def wait_for_task(client, task_id, *, timeout=300.0):
    delay, waited = 0.5, 0.0
    while True:
        task = await client.get_task(GetTaskRequest(id=task_id))
        if task.status.state in TERMINAL | INTERRUPTED:     # 終端 + 中断 の両方で抜ける
            return task
        if waited >= timeout:
            raise TimeoutError(f"task {task_id} は {timeout} 秒で終わらなかった")
        await asyncio.sleep(delay)
        waited += delay
        delay = min(delay * 2, 30.0)            # バックオフ (上限30秒)
```

確認済み: `return_immediately=True` で開始した直後は `SUBMITTED`、0.5秒間隔の `get_task` を6回繰り返して `COMPLETED` になり、成果物 (`chunk0 ... chunk5` を連結したもの) を取得できた。

- `return_immediately=False` (既定) のまま `send_message` を呼ぶと、終端/中断状態になるまで待ってから返るので、ポーリングは不要 (その代わり、その間 HTTP 接続が保たれる)。
- 概念と使い分けは [A2Aについて.md](A2Aについて.md) の「3つの比較表」を参照。

---

# 6. キャンセルと購読

## 6.1 `cancel()` の実装

```python
async def cancel(self, context: RequestContext, event_queue: EventQueue) -> None:
    u = TaskUpdater(event_queue, context.task_id, context.context_id)
    await u.cancel()
```

- 実行中の処理 (asyncio タスク等) を止める処理は自分で書く。上の例は**状態を `canceled` にするだけ**。
- 確認済み: `cancel_task` の応答は `TASK_STATE_CANCELED`、直後の `get_task` も `CANCELED`。終端状態の Task を再度キャンセルすると `TaskNotCancelableError: Task cannot be canceled`。

## 6.2 実行中の Task を購読する

```python
from a2a.types import SubscribeToTaskRequest

async for ev in client.subscribe(SubscribeToTaskRequest(id=task.id)):
    kind = ev.WhichOneof("payload")
    ...
```

- 購読を始めると、まず現在の `task` が届き、その後の `status_update` / `artifact_update` が続く (確認済み: `task(WORKING)` → `status_update(CANCELED)`)。
- 別の接続や別のクライアントから、同じ Task を追跡できる。
- 参考: `opentelemetry` が入っていると、実行中のストリームを閉じたときに `Failed to detach context` というログが出ることがあった (動作には影響なし)。

## 6.3 ストリームが切れたときの復旧 (再購読)

```python
from a2a.types import GetTaskRequest, SubscribeToTaskRequest, TaskState

TERMINAL = {TaskState.TASK_STATE_COMPLETED, TaskState.TASK_STATE_FAILED,
            TaskState.TASK_STATE_CANCELED, TaskState.TASK_STATE_REJECTED}


async def follow_task(client, task_id, handle_event):
    """切れたら GetTask → SubscribeToTask で追いかけ直す。task_id は最初の `task` イベントで保存しておく。"""
    for attempt in range(5):
        task = await client.get_task(GetTaskRequest(id=task_id))
        if task.status.state in TERMINAL:                 # 終端なら購読できない (UnsupportedOperationError)
            return task
        try:
            async for ev in client.subscribe(SubscribeToTaskRequest(id=task_id)):
                handle_event(ev)                          # 最初のイベントは必ず現在の `task`
            return await client.get_task(GetTaskRequest(id=task_id))   # 正常終了後も最終状態を確認
        except Exception:                                 # 接続エラーなど
            await asyncio.sleep(min(2 ** attempt, 30))
    raise RuntimeError("再接続に失敗した")
```

確認済みの動作 (成果物を6個に分けて送るタスク):

| 手順 | 結果 |
|---|---|
| 1本目のストリームで `task`、`artifact_update` ×2 を受信して**読むのをやめる(切断)** | 処理はサーバ側で進み続ける |
| 1秒後に `subscribe` | **最初のイベントは `task` (state=`WORKING`)**。その `artifacts` に `chunk0 chunk1 chunk2 chunk3` までが入っていた |
| 続き | `chunk4`、`chunk5`、`COMPLETED` が届き、ストリームが閉じる |
| 完了後にもう一度 `subscribe` | `UnsupportedOperationError: Task ... is in terminal state: TASK_STATE_COMPLETED` |
| `input-required` で止まっている Task を `subscribe` | 最初に `task` (`INPUT_REQUIRED`) が届き、**ストリームは開いたまま** (3秒待っても閉じなかった) |

- 再購読時の `task` に、それまでの成果物が含まれている点は **a2a-sdk 1.2.2 の挙動**。他の実装でも同じとは限らないので、`task` のスナップショットを必ず「正」として扱い、`artifact_update` を足し込む前に `task.artifacts` で置き換える書き方にしておく。
- `artifact_update` を連結して成果物を組み立てているクライアントは、再購読で届く `task` の `artifacts` と二重に数えないよう、**`task` を受け取った時点で手元の組み立て状態を作り直す**。

---

# 7. 認証・認可 (サーバ側)

SDK は **認証そのものは実装しない**。HTTP リクエストから `User` を取り出す部分を、`context_builder` で差し込む。

```python
from starlette.requests import Request
from a2a.auth.user import User
from a2a.server.routes import DefaultServerCallContextBuilder, create_jsonrpc_routes


class TokenUser(User):
    def __init__(self, name: str):
        self._name = name

    @property
    def is_authenticated(self) -> bool:
        return True

    @property
    def user_name(self) -> str:
        return self._name


class BearerContextBuilder(DefaultServerCallContextBuilder):
    def build_user(self, request: Request) -> User:
        token = request.headers.get("authorization", "")
        if token == "Bearer good-token":          # 実際は JWT 検証などに置き換える
            return TokenUser("alice")
        return super().build_user(request)         # 未認証ユーザー

routes = create_jsonrpc_routes(request_handler=handler, rpc_url="/", context_builder=BearerContextBuilder())
```

Executor 側では `context.call_context.user` で使える:

```python
user = context.call_context.user
if not user.is_authenticated:
    await u.requires_auth(...)      # あるいは reject()
```

> [!WARNING]
> **`context_builder` には `ServerCallContextBuilder` を直接継承せず、`DefaultServerCallContextBuilder` を継承して `build_user()` だけ上書きする。**
> `ServerCallContextBuilder` を直接継承して `build()` を自作し、`state["headers"]` を入れ忘れると、サーバが `A2A-Version` ヘッダを読めず、**0.3 とみなして `VersionNotSupportedError` を返す** (実機で再現)。`context_builder` に関数を渡しても `'function' object has no attribute 'build'` で失敗する。

- Agent Card の `securitySchemes` / `securityRequirements` は**宣言**にすぎず、SDK が自動で検証するわけではない。検証は `context_builder` と `execute()` で自分で行う。宣言の書き方は [A2Aについて.md](A2Aについて.md) の「認可」の節を参照。

---

# 8. タスクの永続化 (`DatabaseTaskStore`)

`InMemoryTaskStore` はプロセスが落ちると消える。本番では DB に置く。

```bash
pip install "a2a-sdk[http-server,sqlite]"      # 本番は postgresql / mysql
```

```python
from sqlalchemy.ext.asyncio import create_async_engine
from a2a.server.tasks import DatabaseTaskStore

engine = create_async_engine("sqlite+aiosqlite:///./a2a.db")   # PostgreSQL なら postgresql+asyncpg://...
task_store = DatabaseTaskStore(engine)
await task_store.initialize()          # テーブルを作る (起動時に1回)

handler = DefaultRequestHandler(agent_executor=..., task_store=task_store, agent_card=agent_card)
```

- 確認済み: `initialize()` で `tasks` テーブルが作られる。エンジンを作り直して開き直しても、保存済みの Task (`COMPLETED`) を取得できた。
- `DatabaseTaskStore` の引数は `engine` のほか `create_table` / `table_name` / `owner_resolver` / 変換関数 (`core_to_model_conversion` / `model_to_core_conversion`)。
- `initialize()` は async なので、ASGI の `lifespan` 内などで呼ぶ。
- スキーマ変更への追従は、`pip install "a2a-sdk[db-cli]"` で入る **`a2a-db`** コマンドを使う (`a2a-db upgrade` / `downgrade` / `current`、`--database-url` で接続先を指定)。**実際のマイグレーションの実行は未確認**。
- Lambda など複数レプリカで動かす場合は、1.2.0 の cluster mode と合わせて検討する。この組み合わせは本ノートでは未確認。

---

# 9. Push 通知 (サーバ側)

クライアントが webhook URL を登録し、Task の更新時にサーバがそこへ POST する仕組み。

```python
import httpx
from a2a.server.tasks import BasePushNotificationSender, InMemoryPushNotificationConfigStore
from a2a.utils.push_url_validator import validate_push_notification_url

push_store = InMemoryPushNotificationConfigStore()           # 永続化するなら DatabasePushNotificationConfigStore
push_sender = BasePushNotificationSender(
    httpx.AsyncClient(),
    push_store,
    push_url_validator=validate_push_notification_url,        # 本番では必須級 (下記)
)

agent_card = AgentCard(..., capabilities=AgentCapabilities(streaming=True, push_notifications=True))

handler = DefaultRequestHandler(
    agent_executor=..., task_store=..., agent_card=agent_card,
    push_config_store=push_store,
    push_sender=push_sender,
    push_url_validator=validate_push_notification_url,
)
```

クライアントは `SendMessageConfiguration(task_push_notification_config=...)` で登録する (10章)。

確認済みの動作 (`push_url_validator` を付けない構成で webhook を立てて確認):
- 送信直後の応答は `TASK_STATE_SUBMITTED` (`return_immediately=True`)。
- webhook に **3回** 届いた (`task` → `statusUpdate` → `statusUpdate`)。
- 届いたヘッダ: `Authorization: Bearer abc123` (登録した `authentication` の `scheme` + `credentials`)、`X-A2A-Notification-Token: secret-token` (登録した `token`)。

`push_url_validator` (SSRF 対策):
- クライアントが登録する URL にサーバがリクエストを送るため、**内部アドレスへの送信を許すと SSRF になる**。
- `validate_push_notification_url` は、`http` / `https` 以外のスキーム、ループバック・プライベート・リンクローカル・予約済みアドレス (`169.254.169.254` を含む) に解決されるホストを拒否する。
- 確認済み: `localhost` → 拒否、`169.254.169.254` → 拒否、`file://` → 拒否、`https://example.com` → 許可。
- **既定は `None` (検証なし)**。公開サーバでは必ず渡す。
- 検証を付けたサーバに `localhost` の webhook を登録すると、**登録時に `InvalidParamsError: Invalid push notification URL`** になる (確認済み。webhook には届かない)。ローカルで試すときは、検証を付けないか、自作の関数を渡す。

## 9.1 受信側 (webhook サーバ) の実装例

通知を受けるのは**クライアント側のアプリ**。認証の検証 → 重複の除外 → すぐ 2xx を返す → 最終状態は `GetTask` で取り直す、の順に書く。

```python
import json
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette.routing import Route

EXPECT_AUTH = "Bearer abc123"           # 登録時に渡した authentication (scheme + credentials)
EXPECT_TOKEN = "secret-token"           # 登録時に渡した token
seen: set[tuple[str, str]] = set()      # 本番は Redis や DB などに置く (重複判定の保存先)
TERMINAL = {"TASK_STATE_COMPLETED", "TASK_STATE_FAILED", "TASK_STATE_CANCELED", "TASK_STATE_REJECTED"}


async def receive(request: Request) -> JSONResponse:
    # 1) 認証 (MUST): 一致しなければ 401
    if request.headers.get("authorization") != EXPECT_AUTH:
        return JSONResponse({"error": "unauthorized"}, status_code=401)
    if request.headers.get("x-a2a-notification-token") != EXPECT_TOKEN:    # SDK 独自のヘッダ名
        return JSONResponse({"error": "unauthorized"}, status_code=401)

    body = await request.json()
    kind = next(iter(body))             # "task" | "message" | "statusUpdate" | "artifactUpdate" のどれか1つ
    payload = body[kind]
    task_id = payload.get("id") if kind == "task" else payload.get("taskId")

    # 2) 冪等 (SHOULD): 同じ内容の重複配信は無視する
    key = (kind, json.dumps(payload, sort_keys=True))
    if key in seen:
        return JSONResponse({"ok": True, "duplicate": True})
    seen.add(key)

    # 3) 最終状態になったら、実データは GetTask で取り直す (通知の内容を鵜呑みにしない)
    state = (payload.get("status") or {}).get("state")
    if state in TERMINAL:
        enqueue_fetch_result(task_id)   # 自前の関数: バックグラウンドで client.get_task(...) を呼ぶ

    # 4) 2xx を即返す (MUST)。重い処理は上のようにバックグラウンドに回す
    return JSONResponse({"ok": True})


app = Starlette(routes=[Route("/hook", receive, methods=["POST"])])
```

確認済みの動作 (a2a-sdk 1.2.2。上のコードを抜き出して `TestClient` で実行: 認証なし → 401、token 違い → 401、同じ通知の再送 → `duplicate`、`COMPLETED` で取り直し対象に入る。さらに、実際のサーバと組み合わせた構成でも確認):
- 通知は `task` (SUBMITTED) → `statusUpdate` (WORKING) → `artifactUpdate` → `statusUpdate` (COMPLETED) の順に届いた。
- COMPLETED の通知を受けて `get_task` で取り直すと、成果物 (`result`) を取得できた。
- 認証が違う POST は 401 を返した。
- **Task ID を応答から記録して照合する方式だと、最初の通知が `send_message` の応答より先に届き、「未知のタスク」として弾いてしまうことがあった**。Task ID だけで照合せず、`contextId` や登録時に自分で決めた相関 ID で照合するか、未知でも一旦受け付けて後で突き合わせる。

> [!NOTE]
> `Content-Type` は、SDK 1.2.2 が送る通知では `application/json` だった (仕様書の例は `application/a2a+json`)。受信側はどちらも受け付けること。`X-A2A-Notification-Token` は SDK 独自のヘッダ名で、仕様書にはない。

---

# 10. クライアント

## 10.1 基本

```python
import asyncio
from uuid import uuid4

from a2a.client import ClientConfig, create_client
from a2a.types import Message, Part, Role, SendMessageRequest, TaskState


async def main():
    client = await create_client("http://localhost:9999")      # ベースURL (Agent Card を自動取得)
    msg = Message(role=Role.ROLE_USER, parts=[Part(text="hello")], message_id=uuid4().hex)

    async for ev in client.send_message(SendMessageRequest(message=msg)):
        kind = ev.WhichOneof("payload")          # "task" | "message" | "status_update" | "artifact_update"
        if kind == "message":                     # Message だけで返すエージェント
            print([p.text for p in ev.message.parts])
        elif kind == "status_update":             # Task で返すエージェント (3章の EchoExecutor はこちら)
            st = ev.status_update.status
            if st.HasField("message"):
                print(TaskState.Name(st.state), [p.text for p in st.message.parts])
    await client.close()

asyncio.run(main())
```

- `send_message` は**常に `StreamResponse` の async iterator** を返す。`ClientConfig(streaming=False)` でも `async for` で受ける (1件だけ流れる)。
- `ev.WhichOneof("payload")` でどれが入っているかを判定する。`getattr(ev, "parts")` のような書き方は動かない。
- **何が届くかはサーバ次第**: Executor が `Message` だけを積むなら `message`、`Task` で管理するなら `task` → `status_update` → ... が届く。3章の `EchoExecutor` に上のコードを実行すると、`task` のあと `status_update` (`TASK_STATE_COMPLETED` + `['echo: hello']`) が届く (確認済み)。`message` が来る前提で書かない。
- 最終結果だけ欲しい場合は、`streaming=False` にして最後の `task` を使うか、`a2a.helpers.get_stream_response_text` を使う (後者は存在のみ確認)。

## 10.2 `ClientConfig`

| 引数 | 内容 |
|---|---|
| `streaming` | `True` (既定) で `SendStreamingMessage` を使う。Agent が `streaming` 非対応なら自動で非ストリーミングになる |
| `polling` | ポーリング設定 (未確認) |
| `httpx_client` | 自前の `httpx.AsyncClient` (タイムアウトや `ASGITransport` を差し込める。13章) |
| `supported_protocol_bindings` | 使うバインディングの優先順。`["HTTP+JSON"]` で REST を指定 (確認済み) |
| `accepted_output_modes` | 受け取れる出力メディアタイプ |
| `push_notification_config` | 既定の Push 通知設定 |

`create_client` の引数: `agent` (URL または `AgentCard`)、`client_config`、`interceptors`、`relative_card_path`、`resolver_http_kwargs`、`signature_verifier`。

> [!NOTE]
> `ClientConfig` が持つ `httpx_client` には、SDK が `A2A-Version: 1.0` ヘッダを既定で設定する。自前の `httpx.AsyncClient` を渡す場合も、同じ設定が追加される。

## 10.3 リクエストオプション

```python
from a2a.types import SendMessageConfiguration

req = SendMessageRequest(
    message=msg,
    configuration=SendMessageConfiguration(
        return_immediately=True,                    # すぐ返す (既定 False = 完了/中断まで待つ)
        history_length=5,                           # 返す履歴の最大件数
        accepted_output_modes=["text/plain"],
        task_push_notification_config=cfg,          # Push 通知の登録 (taskId は空にする)
    ),
    metadata=...,                                   # 任意の追加情報 (google.protobuf.Struct)
)
```

各オプションの意味は [A2Aについて.md](A2Aについて.md) の「リクエストオプション一覧」を参照。

`metadata` の渡し方 (確認済み):

```python
from google.protobuf.struct_pb2 import Struct
md = Struct(); md.update({"trace_id": "abc", "user_lang": "ja"})
SendMessageRequest(message=msg, metadata=md)
# サーバ側: context.metadata == {'trace_id': 'abc', 'user_lang': 'ja'}
```

Push 通知の登録:

```python
from a2a.types import AuthenticationInfo, TaskPushNotificationConfig

cfg = TaskPushNotificationConfig(
    url="https://my-app.example.com/hook",
    token="secret-token",
    authentication=AuthenticationInfo(scheme="Bearer", credentials="abc123"),
)
```

## 10.4 Task を扱うメソッド

| メソッド | リクエスト型 | 内容 |
|---|---|---|
| `get_task(...)` | `GetTaskRequest(id=..., history_length=...)` | Task の取得 (確認済み) |
| `list_tasks(...)` | `ListTasksRequest(page_size=..., page_token=..., context_id=..., status=...)` | 一覧 (確認済み。`next_page_token` は空文字なら終わり) |
| `cancel_task(...)` | `CancelTaskRequest(id=...)` | キャンセル (確認済み) |
| `subscribe(...)` | `SubscribeToTaskRequest(id=...)` | 実行中 Task の購読 (確認済み) |
| `get_extended_agent_card(...)` | `GetExtendedAgentCardRequest()` | 認証後の拡張カード (11章) |
| `create_task_push_notification_config(...)` ほか | Push 通知設定の CRUD | 既存 Task への登録/取得/一覧/削除 (**未確認**) |
| `add_interceptor(...)` / `close()` | | |

- `state` は数値で返る。名前にするには `TaskState.Name(task.status.state)` (例: `TASK_STATE_COMPLETED`)。

## 10.5 認証トークンを付ける

### (a) `AuthInterceptor` + `CredentialService` (Agent Card の宣言に従って自動付与)

```python
from a2a.client.auth.credentials import CredentialService
from a2a.client.auth.interceptor import AuthInterceptor


class MyCreds(CredentialService):
    async def get_credentials(self, security_scheme_name, context):
        return "good-token"          # スキーム名ごとにトークンを返す (None なら付けない)

client = await create_client("http://localhost:9999", interceptors=[AuthInterceptor(MyCreds())])
```

- Agent Card の `securitySchemes` を見て、`http (bearer)` / `oauth2` / `openIdConnect` は `Authorization: Bearer <token>` を付ける (ソースで確認)。
- 確認済み: Agent Card に `http_auth_security_scheme(scheme="Bearer")` と `security_requirements` を宣言したサーバに対して、トークンが付き、サーバ側で `alice` として認識された。
- `InMemoryContextCredentialStore` (セッション ID ごとのトークン保存) も同じ場所にある。

### (b) 自作 Interceptor

```python
from a2a.client.client import ClientCallContext
from a2a.client.interceptors import AfterArgs, BeforeArgs, ClientCallInterceptor


class StaticHeader(ClientCallInterceptor):
    def __init__(self, token: str):
        self.token = token

    async def before(self, args: BeforeArgs) -> None:
        if args.context is None:
            args.context = ClientCallContext()
        if args.context.service_parameters is None:
            args.context.service_parameters = {}
        args.context.service_parameters["Authorization"] = f"Bearer {self.token}"

    async def after(self, args: AfterArgs) -> None:
        pass

client = await create_client(url, interceptors=[StaticHeader("good-token")])
```

確認済み: JSON-RPC・REST のどちらでもヘッダが届き、`send_message` と `get_extended_agent_card` の両方が成功した。

## 10.6 エラー処理

```python
from a2a.client.errors import A2AClientError, AgentCardResolutionError, A2AClientTimeoutError
from a2a.utils.errors import A2AError, TaskNotFoundError, UnsupportedOperationError, TaskNotCancelableError

try:
    await client.get_task(GetTaskRequest(id="no-such-task"))
except TaskNotFoundError:
    ...
except A2AError as e:         # サーバ由来のエラーはすべて A2AError の子
    ...
```

確認済み:
- 存在しない Task → `TaskNotFoundError`。
- 終端状態の Task へ送信 → `UnsupportedOperationError`。
- 終端状態の Task をキャンセル → `TaskNotCancelableError`。
- Agent Card を取得できない (接続不可) → `AgentCardResolutionError` (`A2AClientError` の子)。

`a2a.utils.errors` にある例外: `TaskNotFoundError` / `TaskNotCancelableError` / `PushNotificationNotSupportedError` / `UnsupportedOperationError` / `ContentTypeNotSupportedError` / `InvalidAgentResponseError` / `ExtendedAgentCardNotConfiguredError` / `ExtensionSupportRequiredError` / `VersionNotSupportedError` / `InternalError` / `InvalidParamsError` / `InvalidRequestError` / `MethodNotFoundError` / `JSONParseError`。基底は `A2AError`。

---

## 10.7 Agent Card の取得と利用

### `create_client` に任せる(最も簡単)

```python
client = await create_client("http://localhost:9999")      # ベースURLだけ渡す
```

内部で `<ベースURL>/.well-known/agent-card.json` を取得して Agent Card を解決し、`supported_interfaces` から接続先と通信方式を選ぶ。存在しないパスを指定すると `AgentCardResolutionError` (確認済み)。

- Card のパスが標準と違う: `create_client(url, relative_card_path="/custom/card.json")`
- Card の取得に認証や追加ヘッダが要る: `resolver_http_kwargs={"headers": {...}}` (`httpx` の引数)
- Card の署名を検証したい: `signature_verifier=...` (未確認)

### 自分で取得して中身を見てから使う

```python
import httpx
from a2a.client import create_client
from a2a.client.card_resolver import A2ACardResolver

async with httpx.AsyncClient() as http:
    resolver = A2ACardResolver(httpx_client=http, base_url="http://localhost:9999")
    card = await resolver.get_agent_card()                 # 既定のパスは /.well-known/agent-card.json

print(card.name, card.capabilities.streaming, card.capabilities.push_notifications)
for i in card.supported_interfaces:
    print(i.protocol_binding, i.protocol_version, i.url)
for sk in card.skills:
    print(sk.id, sk.description)

client = await create_client(card)                         # 取得済みの AgentCard を渡す(再取得しない)
```

確認済み: `A2ACardResolver(httpx_client, base_url, agent_card_path)` と `get_agent_card(relative_card_path=None, http_kwargs=None, signature_verifier=None)`。取得した Card を `create_client` に渡せる。

### Card の内容で処理を分ける

```python
if not card.capabilities.streaming:
    config = ClientConfig(streaming=False)                  # SDK は自動でも非ストリーミングに切り替えるが、明示してもよい
if card.capabilities.push_notifications:
    ...                                                     # Push通知を登録できる
if card.capabilities.extended_agent_card:
    card = await client.get_extended_agent_card(GetExtendedAgentCardRequest())   # 認証後の詳細版 (11章)
```

- Card の `security_schemes` / `security_requirements` を見て、必要なトークンを用意する (10.5)。
- 複数の `supported_interfaces` があるとき、使いたい通信方式を `ClientConfig(supported_protocol_bindings=["HTTP+JSON"])` で優先指定できる (12.1)。
- 旧形式 (v0.3) の Agent Card (`url` / `preferredTransport` / `security` など) は、SDK が取得時に自動で v1.0 の形式に変換する (ソースで確認)。

### キャッシュ (サーバが付けるヘッダ)

```bash
curl -si http://localhost:9999/.well-known/agent-card.json | grep -i -E "etag|cache-control"
curl -si -H 'If-None-Match: <上のETag>' http://localhost:9999/.well-known/agent-card.json | head -1   # 304
```

- 確認済み: SDK 1.2.2 のサーバは `ETag` を付け、`If-None-Match` つきの再取得で `304` を返す。`Cache-Control` は既定では付かない。
- `Cache-Control` を付けるには `create_agent_card_routes(agent_card=..., cache_control="public, max-age=300")` のように指定する (確認済み: 応答に `Cache-Control: public, max-age=300` と `ETag` が付き、`If-None-Match` で `304`)。
- `A2ACardResolver` / `create_client` が条件付きリクエストでキャッシュを使うかどうかは、**未確認**。

---

# 11. 拡張 Agent Card (認証後だけ見せる情報)

Agent Card の公開版にはスキルの一部だけを載せ、認証済みユーザーには追加スキルを見せる。

```python
agent_card = AgentCard(..., capabilities=AgentCapabilities(streaming=False, extended_agent_card=True))

async def modifier(card: AgentCard, ctx: ServerCallContext) -> AgentCard:
    if not ctx.user.is_authenticated:
        return card
    out = AgentCard()
    out.CopyFrom(card)
    out.skills.add(id="admin", name="admin", description="認証済みユーザー向け", tags=["t"], examples=["x"])
    return out

handler = DefaultRequestHandler(
    agent_executor=..., task_store=..., agent_card=agent_card,
    extended_agent_card=agent_card,          # ベースになるカード
    extended_card_modifier=modifier,         # リクエストごとに加工
)
```

クライアント: `await client.get_extended_agent_card(GetExtendedAgentCardRequest())`

確認済み: トークンなしだと `['pub']`、`Bearer good` だと `['pub', 'admin']` が返った。

- `AgentCapabilities(extended_agent_card=True)` を宣言していないクライアントの `get_extended_agent_card` は、**サーバへ問い合わせず手元のカードを返す**。
- 7章の `DefaultServerCallContextBuilder` の落とし穴は、この機能でも同じ。

---

# 12. 公開方法 (REST / FastAPI / gRPC)

## 12.1 REST (HTTP+JSON) で公開

```python
from a2a.server.routes import create_rest_routes

agent_card = AgentCard(..., supported_interfaces=[
    AgentInterface(url="http://localhost:9999/", protocol_binding="JSONRPC", protocol_version="1.0"),
    AgentInterface(url="http://localhost:9999/rest", protocol_binding="HTTP+JSON", protocol_version="1.0"),
])

routes = (
    create_agent_card_routes(agent_card=agent_card)
    + create_jsonrpc_routes(request_handler=handler, rpc_url="/")
    + create_rest_routes(request_handler=handler, path_prefix="/rest")
)
```

- `AgentInterface.protocol_binding` の値は **`"JSONRPC"` / `"HTTP+JSON"` / `"GRPC"`**。
- クライアントは `ClientConfig(supported_protocol_bindings=["HTTP+JSON"])` で REST を選べる (確認済み)。
- `create_rest_routes` / `create_jsonrpc_routes` にも `enable_v0_3_compat` と `shutdown_grace_period` (SSE 終了時の猶予秒) がある。

## 12.2 既存の FastAPI アプリに足す (`[fastapi]`)

```python
from fastapi import FastAPI
from a2a.server.routes import add_a2a_routes_to_fastapi

app = FastAPI()
add_a2a_routes_to_fastapi(
    app,
    agent_card_routes=create_agent_card_routes(agent_card=agent_card),
    jsonrpc_routes=create_jsonrpc_routes(request_handler=handler, rpc_url="/"),
)

@app.get("/health")
def health():
    return {"ok": True}
```

確認済み: A2A のルートと、自前のルート (`/health`) が同居して動く。

## 12.3 gRPC

- `pip install "a2a-sdk[grpc]"` で `grpcio` 系が入り、`a2a.types.a2a_pb2_grpc` と `a2a.client.transports.grpc` がある。
- **サーバ・クライアントとも、本ノートでは動作確認していない**。使う場合は公式サンプル ([a2a-samples](https://github.com/a2aproject/a2a-samples)) を参照する。

---

# 13. テスト

HTTP サーバを立てずに、`httpx.ASGITransport` でアプリを直接呼べる。

```python
# test_agent.py
import httpx
import pytest
from uuid import uuid4

from a2a.client import ClientConfig, create_client
from a2a.types import Message, Part, Role, SendMessageRequest, TaskState

from server import app      # 3章の server.py (Starlette アプリ)


@pytest.fixture
async def client():
    http = httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://test")
    c = await create_client("http://test", client_config=ClientConfig(streaming=False, httpx_client=http))
    yield c
    await c.close()
    await http.aclose()


async def test_echo_completes(client):
    msg = Message(role=Role.ROLE_USER, parts=[Part(text="ping")], message_id=uuid4().hex)
    events = [ev async for ev in client.send_message(SendMessageRequest(message=msg))]
    task = next(e.task for e in events if e.WhichOneof("payload") == "task")
    assert task.status.state == TaskState.TASK_STATE_COMPLETED
    assert task.status.message.parts[0].text == "echo: ping"
```

```ini
# pytest.ini
[pytest]
asyncio_mode = auto
```

```bash
pip install pytest pytest-asyncio httpx
pytest -q
```

- 確認済み: 上のコードを 3章の `server.py` に対してそのまま実行して `1 passed`。
- 通信は `httpx_client` の `ASGITransport` でアプリに直接渡される。`AgentInterface.url` (`http://localhost:9999/`) と `base_url` (`http://test`) が食い違っていても通った (確認済み)。ネットワークは使わない。
- テスト用に Agent Card や Executor を差し替えたいときは、`build_app()` のようにアプリの組み立てを関数にしておくと楽。

---

# 14. よくあるハマりどころ

| 症状 | 原因 | 対処 |
|---|---|---|
| `InvalidAgentResponseError: Agent should enqueue Task before TaskStatusUpdateEvent event` | 初回に `Task` を積まず `start_work()` などを呼んだ | `new_task_from_user_message()` で作った Task を先に `enqueue_event` (4.1) |
| `InvalidAgentResponseError: Received Message object in task mode` | Task を積んだあとに `Message` を積んだ | 返信は `TaskUpdater.complete(message)` で行う (`Message` を直接積まない) |
| `RuntimeError: Task ... is already in a terminal state` | 終端状態にしたあとに更新した | `complete()` / `failed()` などは最後に1回だけ |
| `UnsupportedOperationError ... is in terminal state` | 終端状態の Task に追加メッセージを送った | 新しい依頼は `task_id` なしで送る |
| `VersionNotSupportedError: A2A version '0.3' is not supported` | `context_builder` を自作して `state["headers"]` を入れていない / `A2A-Version` ヘッダが消えている | `DefaultServerCallContextBuilder` を継承する (7章)。HTTP クライアントを自作する場合は `A2A-Version: 1.0` を付ける |
| `'function' object has no attribute 'build'` | `context_builder` に関数を渡した | `ServerCallContextBuilder` のサブクラスのインスタンスを渡す |
| `getattr(event, "parts")` が常に空 | `send_message` が返すのは `StreamResponse` | `ev.WhichOneof("payload")` で判定し、`ev.message.parts` などを読む (10.1) |
| ストリーミングで届かず、`task` が1件だけ届く | Agent Card で `streaming` が `False` / 未設定 (クライアントが非ストリーミングに切り替えた) | 段階的に返したいなら `AgentCapabilities(streaming=True)` を宣言 (5.1) |
| 登録時に `InvalidParamsError: Invalid push notification URL` | `push_url_validator` が `localhost` などの内部アドレスを拒否した | ローカル検証では検証関数を外す。本番では外さない |
| 古い記事のコードが動かない | v0.3 API (`A2AStarletteApplication`, `A2AClient`, `{"kind": "text"}` など) | 本ノートの 1.x 書き方に置き換える。互換が必要なら `enable_v0_3_compat=True` |

---

# 15. 未確認・今後の確認事項

- gRPC のサーバ/クライアント。
- `a2a-db` による DB マイグレーションの実行。
- `DatabasePushNotificationConfigStore` と `encryption` extra (暗号化) の動作。
- 署名付き Agent Card の作成・検証 (`signing` extra, `signature_verifier`)。
- `ClientConfig.polling`、Push 通知設定の CRUD メソッド (`create_task_push_notification_config` ほか)。
- 複数レプリカ (cluster mode) での `TaskStore` 共有。
- `execute()` 内で捕捉されない例外が出たときの挙動。
- OpenTelemetry (`telemetry` extra) によるトレース出力。
- `owner_resolver` を使ったマルチユーザーでの Task の分離。
- 拡張機構 (`A2A-Extensions` / `AgentExtension`) の実装。

---

# 参考

- 概念: [A2Aについて.md](A2Aについて.md)
- SDK: https://github.com/a2aproject/a2a-python / https://pypi.org/project/a2a-sdk/
- サンプル: https://github.com/a2aproject/a2a-samples
- 仕様: https://a2a-protocol.org/latest/specification/
