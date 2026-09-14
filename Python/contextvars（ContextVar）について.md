# 概念
- `ContextVar`は、標準ライブラリ`contextvars`が提供する「実行コンテキストに紐づいた変数」を作るためのクラス。
  - 通常のグローバル変数と違い、**どのコンテキスト（呼び出しの流れ）から見ているか**によって値が変わる
  - 同期コードの`threading.local`に近いが、`threading.local`はスレッド単位でしか値を分離できないのに対し、`ContextVar`は **`asyncio`のTask単位** でも値を分離できる
- 主な用途は「関数の引数として明示的にバケツリレーしたくないが、非同期処理・並行処理下でも値が混ざってはいけない」ような横断的な状態を持ち回ること
  - 例: リクエストID・トレースID、ログイン中のユーザー情報、DBセッション、ロケール設定など

# なぜ普通のグローバル変数や`threading.local`ではダメなのか
- **グローバル変数**: プロセス全体で1つの値しか持てない。非同期処理で複数のリクエストを並行処理していると、リクエストAの値をリクエストBが上書きしてしまう
- **`threading.local`**: スレッドごとに値を分離できるが、`asyncio`は基本的に**1スレッドで複数のTask（≒リクエスト）を並行実行する**ため、スレッド単位の分離では粒度が粗すぎる（全Taskで同じ値を共有してしまう）
- **`ContextVar`**: `asyncio.Task`ごと（正確にはコンテキストごと）に独立した値を持てるため、非同期処理でも安全にリクエストスコープの状態を持ち回せる

```
【threading.localの場合（asyncioでは機能しない）】
Thread-1 ─┬─ Task A (request_id=A) ┐
          └─ Task B (request_id=B) ┘→ 同じスレッドなので値が混ざる/上書きされる

【ContextVarの場合】
Thread-1 ─┬─ Task A → ContextVar = "A" （Task Aだけから見える）
          └─ Task B → ContextVar = "B" （Task Bだけから見える）
```

# どういう時に使うか
- **FastAPI/Starletteなどの非同期Webアプリで、リクエスト単位のトレースID・ユーザー情報をログや下位関数に伝播させたい時**
  - 各エンドポイント処理・ミドルウェアが並行して動いていても、それぞれのリクエストに紐づいた値だけを取得できる
- **ロギングでリクエストIDを自動的に全ログ行に埋め込みたい時**
  - `logging.Filter`や`LoggerAdapter`と組み合わせて、`ContextVar`から取得した値をログフォーマットに差し込む
- **DBセッションやトランザクションのコンテキストを、引数で渡さずに下位の関数から参照したい時**
- 逆に言うと、使い分けの軸は「並行処理があるかどうか」ではなく、**値を使う関数同士が直接の呼び出し関係（＝引数で渡せる関係）にあるかどうか**
  - 呼び出し元から数珠つなぎに直接呼ばれている関数だけで値が完結するなら、非同期処理下であっても素直に引数やローカル変数で渡す方がシンプル
  - ロガー・共通ライブラリ・ミドルウェアなど、値を使う場所が「直接の呼び出し関係にない/引数を通したくない場所」に散らばっている場合に`ContextVar`が効果を発揮する
  ```python
  # ケースA: 直接の呼び出し関係で完結している → 引数で渡せば済む（ContextVar不要）
  def handler(request):
      user_id = extract_user_id(request)
      do_something(user_id)      # 直接呼んでいるので引数で渡せる

  def do_something(user_id):
      ...

  # ケースB: 呼び出し関係が離れている/経由しない → ContextVarが有効
  def handler(request):
      request_id_var.set(str(uuid4()))
      do_something()              # request_idを引数で渡していない

  def do_something():
      call_external_library()     # さらに奥の、自分が書いていないライブラリのコード

  def call_external_library():
      ...
      logger.info("処理中")        # ここでログにrequest_idを埋め込みたいが、
                                   # handlerからここまで引数で繋ぐのは非現実的
  ```

# 基本的な使い方

## 1. 生成・set・get
```python
from contextvars import ContextVar

# 第2引数のdefaultは省略可能（省略した場合、setされていないと get() でLookupErrorになる）
request_id_var: ContextVar[str] = ContextVar("request_id", default="unknown")

def handler():
    print(request_id_var.get())  # "unknown"（まだsetされていない場合）

def middleware():
    request_id_var.set("abc-123")
    handler2()

def handler2():
    print(request_id_var.get())  # "abc-123"
```
- `ContextVar(name, default=...)`の`name`はデバッグ用の識別名で、変数自体の参照には使わない（モジュールレベルの変数として保持し、それを`import`して使う）
- `.set(value)`で値をセット、`.get()`で現在のコンテキストの値を取得
  - `.get(default)`のように呼び出し時に`default`を渡すこともでき、その場合は`ContextVar`生成時の`default`よりも優先される

> [!CAUTION]
> **`ContextVar`は必ずモジュールのトップレベルで生成し、関数やクロージャの中で生成しない。** これは公式ドキュメントが明示的に注意している作法。`Context`オブジェクトは自身が保持する`ContextVar`を強参照で持ち続けるため、関数呼び出しのたびにクロージャ内で使い捨てのつもりで生成すると、参照が残り続けてガベージコレクションされず、メモリリークの原因になる。
> ```python
> # ❌ NG: 呼び出すたびに新しいContextVarが生成され、GCされずリークしていく
> def handler():
>     request_id_var = ContextVar("request_id", default="-")  # 関数（クロージャ）内で生成
>     token = request_id_var.set("abc-123")
>     ...
>     request_id_var.reset(token)
>
> # ✅ OK: モジュールのトップレベルで1度だけ生成し、使い回す
> request_id_var: ContextVar[str] = ContextVar("request_id", default="-")
>
> def handler():
>     token = request_id_var.set("abc-123")
>     ...
>     request_id_var.reset(token)
> ```

## 2. `Token`による値の巻き戻し（reset）
- `.set()`は`Token`オブジェクトを返す。これを`.reset(token)`に渡すと、`set`する前の状態に戻せる
- 特にミドルウェアなどで「このリクエスト処理が終わったら元に戻したい」場合に使う
```python
def middleware():
    token = request_id_var.set("abc-123")
    try:
        handler2()
    finally:
        request_id_var.reset(token)  # setする前の状態（未設定 or 前の値）に戻す
```
> [!CAUTION]
> 1つの`Token`は1回しか使えない。同じ`Token`を2回以上`reset()`に渡すと`RuntimeError`、別の`Context`で発行された`Token`を渡すと`ValueError`になる（実際に実行して確認済み）。
> ```python
> token = request_id_var.set("abc-123")
> request_id_var.reset(token)  # OK: "abc-123"をsetする前の状態に戻る
> request_id_var.reset(token)  # RuntimeError: ...has already been used once
> ```
> ```python
> ctx = contextvars.copy_context()
> token = ctx.run(lambda: request_id_var.set("in-ctx"))  # ctx側で発行されたtoken
> request_id_var.reset(token)  # ValueError: ...was created in a different Context
> ```

> [!NOTE]
> `reset()`を使わず`.set(default)`のような書き方で「戻したつもり」にすると、元々別の値がsetされていた場合に情報が失われる。正しく巻き戻すには必ず`set()`が返した`Token`を使う。

> [!TIP]
> Python 3.14以降、`ContextVar.set()`の戻り値はコンテキストマネージャとしても使えるようになった。
> ```python
> with request_id_var.set("abc-123"):
>     handler2()
> # ブロックを抜けると自動的にreset()される
> ```
> 上記の`try`/`finally`と等価だが、より簡潔に書ける。プロジェクトの対応Pythonバージョンが3.14以上の場合はこちらが推奨。

## 3. FastAPIのミドルウェアでの典型例
```python
import uuid
from contextvars import ContextVar
from starlette.middleware.base import BaseHTTPMiddleware

request_id_var: ContextVar[str] = ContextVar("request_id", default="-")

class RequestIdMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        token = request_id_var.set(str(uuid.uuid4()))
        try:
            response = await call_next(request)
        finally:
            request_id_var.reset(token)
        return response

# 下位の関数やロガーからは引数で渡さずに参照できる
def get_logger_extra():
    return {"request_id": request_id_var.get()}
```

# `asyncio.Task`との関係・伝播の仕組み
- `ContextVar`の値は「コンテキスト（`contextvars.Context`）」というスナップショットに保存されている
- `asyncio.create_task()`でTaskを作成すると、**作成時点のコンテキストがコピーされて新しいTaskに引き継がれる**
  - そのため、親のコルーチンで`set`した値は、そこから作られた子Taskにも見える（生成時点までの値がコピーされる）
  - 逆に、子Taskの中で`set`した値は、コピーされた別のコンテキストに書き込まれるため、**親や兄弟Taskには影響しない**
```python
async def child():
    request_id_var.set("child-value")  # このTask内でしか有効にならない
    print(request_id_var.get())  # "child-value"

async def parent():
    request_id_var.set("parent-value")
    task = asyncio.create_task(child())  # 作成時点の"parent-value"がコピーされる
    await task
    print(request_id_var.get())  # "parent-value"（childの変更は影響しない）
```
- 同期的な`await`で単に別のコルーチンを呼ぶ場合（Taskを作らない場合）は、同じコンテキストを共有するため、変更がそのまま見える点に注意（Taskの境界を越えるかどうかがポイント）
- `copy_context()`を使うと、現在のコンテキストを明示的にコピーして`Context.run()`で別の関数を実行できる（`concurrent.futures`のスレッドに処理を渡す際などに、コンテキストを引き継ぎたい場合に使う）
  ```python
  import contextvars

  ctx = contextvars.copy_context()
  ctx.run(some_function)  # some_function内でも現在のContextVarの値が見える
  ```
  - `asyncio.Task`と同様、`ctx.run()`の中で`set()`した変更は**コピーされた`ctx`側にのみ反映され**、`copy_context()`を呼び出した側（呼び出し元で今まさに実行中のコンテキスト）には影響しない

# 注意点・既知の問題
- `BaseHTTPMiddleware`は内部でタスクを分けて処理を橋渡しする実装になっており、**`ContextVar`の伝播が期待通りに行われないケースがある**（エンドポイント側でsetした値が外側のMiddlewareから見えない、またはその逆）
  - 詳細・回避策（pure ASGI middlewareへの書き換え）は [Middleware（CORSを含む）について.md](FastAPI/Middleware（CORSを含む）について.md) を参照
- `default`を指定しない`ContextVar`に対して、`set()`される前に`get()`を呼ぶと`LookupError`になる。ミドルウェアの実行順序やテストコードで意図せず未setのまま`get()`してしまわないよう注意
- グローバルなミュータブルオブジェクト（`list`や`dict`）を`ContextVar`にセットして使い回すと、コンテキストが分離されていても中身の書き換えは共有されてしまう（参照が同じであれば当然どのコンテキストからも同じオブジェクトを指す）。値そのものを都度新しく`set()`するか、イミュータブルな値を使うのが安全
