# FastAPIの `async def` と `def` の使い分けについて

FastAPIではPath Operation Function（エンドポイント）を **`async def`** でも **`def`** でも書ける。どちらでも動くが、**中で何を呼ぶか**によって正しい選択が変わる。誤ると「速いはずのFastAPIが詰まる」典型的なハマりどころになる。

> 出典: [FastAPI公式 - Concurrency and async / await](https://fastapi.tiangolo.com/async/)

---

## 1. 結論（ルール・オブ・サム）

| 中で呼ぶもの | 書き方 |
|---|---|
| `await` する非同期ライブラリ（`httpx.AsyncClient`、非同期DBドライバ等） | **`async def`** |
| ブロッキングなライブラリ（多くの同期DB・`requests`・重いCPU処理・`time.sleep`） | **`def`** |
| 判断がつかない | **`def`**（安全側） |

> 公式の「In a hurry?」も同じ結論:「`await` を使うなら `async def`、ブロッキングなら `def`、迷ったら `def`」。どの場合でもFastAPIは非同期に動作し十分速い。

---

## 2. なぜこうなるのか（内部の挙動）

FastAPI（実体はStarlette）は**単一のイベントループ**上で多数のリクエストを捌く。ここで鍵になるのが「`def` で書いたエンドポイントの扱い」。

- **`def`（同期）で宣言したPath Operation** → FastAPIが**外部スレッドプール**で実行し、その完了を `await` する。
  - つまりブロッキング処理をしても**イベントループはブロックされない**（別スレッドで動くため）。
- **`async def`（非同期）で宣言したPath Operation** → イベントループ上で**直接**実行される。
  - この中で**ブロッキング処理をすると、イベントループ全体が止まる**（他の全リクエストが待たされる）。

```python
import time
import asyncio

# ❌ 危険: async def の中でブロッキング（time.sleep）
@app.get("/bad")
async def bad():
    time.sleep(5)      # ← イベントループを5秒間止める。全リクエストが固まる
    return {"ok": True}

# ✅ OK: ブロッキング処理なら def にする（スレッドプールで動く）
@app.get("/good-sync")
def good_sync():
    time.sleep(5)      # ← 別スレッドなのでイベントループは止まらない
    return {"ok": True}

# ✅ OK: 非同期版があるなら async def + await
@app.get("/good-async")
async def good_async():
    await asyncio.sleep(5)   # ← ノンブロッキング。他のリクエストを捌ける
    return {"ok": True}
```

> [!WARNING]
> 一番やってはいけないのが **`async def` の中でブロッキング処理を呼ぶ**こと（`requests.get()`、同期DBクエリ、`time.sleep()`、重いCPUループなど）。イベントループが専有され、**サーバー全体のスループットが激減**する。
> - ブロッキングライブラリしか無い → エンドポイントを **`def`** にする（FastAPIがスレッドプールへ逃がす）
> - どうしても `async def` の中で同期処理を呼びたい → `await run_in_threadpool(func, ...)`（`from starlette.concurrency import run_in_threadpool`）でスレッドに逃がす

---

## 3. 依存関数（`Depends`）でも同じルール

Path Operationだけでなく、**依存関数（`Depends`）も `def` なら外部スレッドプールで実行**される。`async def` の依存はイベントループ上で直接実行される。

```python
# ブロッキングなDBセッションを使う依存 → def でよい（スレッドプールで動く）
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# 非同期クライアントを使う依存 → async def
async def get_http() -> httpx.AsyncClient:
    ...
```

`def` と `async def` の依存は**自由に混在**できる。FastAPIがそれぞれ適切に扱う（→[APIRouterについて](APIRouterについて.md) の `Depends`）。

---

## 4. スレッドプールには上限がある（見落としがち）

`def` エンドポイント／依存を捌くスレッドプールは**無制限ではない**。FastAPI/Starletteは並行処理基盤に **[AnyIO](https://anyio.readthedocs.io/)** を使っており、スレッドで動かせる同期処理の同時実行数には既定の上限がある（歴史的に既定は40スレッド）。

つまり:

- 大量の同時リクエストがすべて**遅いブロッキング `def`** だと、スレッドプールが埋まって**待ち行列**ができる。
- 高スループットが要る経路で外部I/Oを叩くなら、**非同期ライブラリ + `async def`** の方がスケールしやすい。

> [!NOTE]
> 「`def` にしておけば安全」は正しいが、**万能ではない**。ブロッキング経路が高トラフィックなら、非同期対応ライブラリへの移行（`requests`→`httpx`、同期DBドライバ→非同期ドライバ等）を検討する。ストリーミング（→[StreamingResponseについて](%20StreamingResponseについて.md)）でも、generator内のブロッキングI/Oは同様にイベントループを止める点に注意。

---

## 5. `async def` にしても速くならないケース

`await` する非同期I/Oが無いのに `async def` にしても、並行性のメリットは出ない（イベントループを手放す箇所が無いため）。純粋なCPU計算は `async def` にしても速くならず、むしろブロッキングとして働くので、**重いCPU処理は `def`**（スレッドプール）か、別プロセス／ワーカーへ逃がすのが基本。

---

## ポイントまとめ

- `await` する非同期ライブラリを使うなら `async def`、ブロッキングなら `def`、**迷ったら `def`**
- `def` のエンドポイント／依存はFastAPIが**外部スレッドプール**で実行 → イベントループを塞がない
- 最悪手は **`async def` の中でブロッキング処理**（イベントループが止まり全体が遅くなる）。逃がすなら `run_in_threadpool`
- スレッドプールには既定の上限（AnyIO、歴史的に40）があり、ブロッキング経路が高トラフィックなら非同期化を検討
- 重いCPU処理は `async def` では速くならない。`def` かワーカー分離で
