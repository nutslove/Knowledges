# FastAPIのBackgroundTasksについて

**`BackgroundTasks`** は、**レスポンスをクライアントに返した後**に実行したい処理を登録する仕組み。「結果を待たせたくないが、やっておきたい処理」（メール送信、ログ書き込み、通知など）を、リクエストのレイテンシに影響させずに実行できる。

> 出典: [FastAPI公式 - Background Tasks](https://fastapi.tiangolo.com/tutorial/background-tasks/)

---

## 1. 基本形

```python
from fastapi import BackgroundTasks, FastAPI

app = FastAPI()


def write_log(message: str):
    with open("log.txt", "a") as f:
        f.write(message + "\n")


@app.post("/send-notification/{email}")
async def send_notification(email: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(write_log, f"notification sent to {email}")
    return {"message": "Notification sent in the background"}
```

- `background_tasks: BackgroundTasks` を引数に宣言すると、FastAPIが自動でインスタンスを注入する（`Depends`不要。特別扱いされる型）
- `add_task(関数, *args, **kwargs)` で登録した関数は、**レスポンスがクライアントに送出された後**に実行される
- 複数回 `add_task` を呼べば、**登録した順に**すべて実行される

---

## 2. `Depends` の中からも登録できる

依存関数（`Depends`）の中でも `BackgroundTasks` を受け取って `add_task` できる。同一リクエスト内であれば、エンドポイント側とタスクは**同じ`BackgroundTasks`インスタンスを共有**する。

```python
def get_notifier(background_tasks: BackgroundTasks) -> str:
    background_tasks.add_task(write_log, "dependency executed")
    return "notifier"


@app.post("/items/")
async def create_item(
    item: str,
    notifier: Annotated[str, Depends(get_notifier)],
    background_tasks: BackgroundTasks,
):
    background_tasks.add_task(write_log, f"item created: {item}")
    return {"item": item}
```

---

## 3. 同期関数・非同期関数どちらも登録できる

```python
def sync_task(x: int):
    time.sleep(1)          # ブロッキングでもOK。BackgroundTasksが別スレッドで実行する

async def async_task(x: int):
    await asyncio.sleep(1) # 非同期I/Oならこちらが自然

background_tasks.add_task(sync_task, 1)
background_tasks.add_task(async_task, 2)
```

- 内部的には、同期関数は[async def と def の使い分け](async%20def%20と%20def%20の使い分けについて.md)と同じ理屈で**スレッドプール**に逃がされ、非同期関数はイベントループ上で直接実行される

---

## 4. 適用範囲の限界（重要な注意点）

`BackgroundTasks` は**同一プロセス内**で、**レスポンス送出後にすぐ**実行される軽量な仕組み。以下のようなケースには**向かない**。

| 状況 | `BackgroundTasks` で良いか |
|---|---|
| 数秒〜数十秒で終わる軽い処理（ログ書き込み、通知メール送信、キャッシュ更新） | ◎ 十分 |
| 数分かかる重い処理（動画エンコード、大規模バッチ、機械学習の推論） | △ プロセスが落ちると失われる。専用ワーカー（Celery, RQ, Arq等）を検討 |
| 確実に実行完了を保証したい（リトライ、永続キュー） | ✗ 不向き。メッセージキュー（SQS, RabbitMQ等）+ ワーカーが必要 |
| Podの再起動・スケールインが起きても処理を失いたくない | ✗ 不向き。`BackgroundTasks`はPodのプロセスに紐づくため、Podが落ちればタスクも消える |

> [!WARNING]
> `BackgroundTasks`はアプリケーションプロセスの中で動く。**Uvicornワーカーが再起動・クラッシュ・スケールインされるとタスクは失われる**。信頼性が必要な非同期処理（決済処理の後続処理など）は、Celery/RQ/ArqやSQS+Lambdaのような**永続的なジョブキュー**に任せるべき。`BackgroundTasks`は「軽くて失敗しても致命的でない処理」向け。

---

## 5. `StreamingResponse` との違い

どちらも「レスポンスに関連して何かを遅延実行する」点が似ているが、目的が異なる。

| 項目 | `BackgroundTasks` | [StreamingResponse](%20StreamingResponseについて.md) |
|---|---|---|
| クライアントに結果を返すか | 返さない（レスポンス後の裏処理） | 返す（本体そのもの） |
| 実行タイミング | レスポンス送出**後** | レスポンス送出**中**（生成しながら送る） |
| 用途 | 通知・ログ・後片付け | LLM出力、大容量データの逐次送信 |

---

## ポイントまとめ

- `BackgroundTasks`は「レスポンスは即返し、付随処理は裏で実行したい」場合に使う軽量な仕組み
- 引数に `background_tasks: BackgroundTasks` と書くだけで注入される（`Depends`不要）。`add_task(fn, *args, **kwargs)`で登録
- 同期・非同期どちらの関数も登録可能。同期はスレッドプールで実行される
- **プロセス内蔵の仕組みなので、Podが落ちればタスクも消える**。重い処理・確実性が必要な処理には向かない → Celery/RQ/Arqやメッセージキューを検討
- `Depends`の中からも`BackgroundTasks`を受け取ってタスク登録できる（エンドポイントと同じインスタンスを共有）
