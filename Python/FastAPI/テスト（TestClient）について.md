# FastAPIのテスト（TestClient）について

FastAPIアプリは **`TestClient`** を使うことで、実際にサーバーを起動せずにHTTPリクエストを模擬してテストできる。内部的には`httpx`ベースで、ASGIアプリ（→[ASGIについて](ASGIについて.md)）に直接リクエストを渡すため高速。

> 出典: [FastAPI公式 - Testing](https://fastapi.tiangolo.com/tutorial/testing/)

---

## 1. 前提

```bash
pip install pytest httpx
```

- FastAPI 0.100系以降、`TestClient` は `fastapi.testclient` からインポートするのが基本（内部は`starlette.testclient`のラップ）
- `TestClient` 自体が `httpx.Client` をラップしているため、`httpx`は必須

---

## 2. 基本形

```python
# main.py
from fastapi import FastAPI

app = FastAPI()


@app.get("/items/{item_id}")
async def read_item(item_id: int, q: str | None = None):
    return {"item_id": item_id, "q": q}
```

```python
# test_main.py
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)


def test_read_item():
    response = client.get("/items/5?q=hello")
    assert response.status_code == 200
    assert response.json() == {"item_id": 5, "q": "hello"}
```

- `TestClient` は`requests`ライクなAPI（`.get()` / `.post()` / `.put()` / `.delete()`等）を持つ
- **サーバーを別プロセスで立てる必要がない**（ASGIアプリの呼び出しをインメモリで行う）ので高速かつCI向き

### JSONボディを送る例

```python
def test_create_item():
    response = client.post("/items/", json={"name": "Foo", "price": 10.5})
    assert response.status_code == 200
    body = response.json()
    assert body["name"] == "Foo"
```

### ヘッダー・認証トークンを付ける例

```python
def test_protected():
    response = client.get("/protected", headers={"X-API-Key": "secret-key-123"})
    assert response.status_code == 200
```

---

## 3. `pytest` フィクスチャとして使う（推奨）

テストが増えたら`client`をフィクスチャ化し、テスト間で使い回す。

```python
# conftest.py
import pytest
from fastapi.testclient import TestClient
from main import app


@pytest.fixture
def client():
    with TestClient(app) as c:
        yield c
```

```python
# test_items.py
def test_read_item(client):
    response = client.get("/items/5")
    assert response.status_code == 200
```

> [!NOTE]
> `with TestClient(app) as c:` のように **コンテキストマネージャとして使う**と、[lifespanについて](lifespanについて.md)で説明した起動・終了処理（`lifespan`）も実行される。**単に`client = TestClient(app)`とインスタンス化しただけでは`lifespan`は実行されない**（`with`ブロックに入った時点で初めてstartupが走る仕様）。lifespanで用意したリソース（DBプール、`app.state`の値など）をテストで使うなら`with`文は必須。

---

## 4. 依存性の差し替え（`app.dependency_overrides`）

[APIRouterについて](APIRouterについて.md)の`Depends`で説明した依存関係は、テスト時に**別の実装へ丸ごと差し替え**られる。DB接続や認証を本物にせずモックする、実務で最頻出のテストパターン。

```python
# main.py
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


@app.get("/users")
async def list_users(db=Depends(get_db)):
    return db.query(User).all()
```

```python
# test_main.py
from main import app, get_db


def override_get_db():
    db = TestingSessionLocal()   # テスト用（SQLiteのin-memory等）のセッション
    try:
        yield db
    finally:
        db.close()


app.dependency_overrides[get_db] = override_get_db


def test_list_users(client):
    response = client.get("/users")
    assert response.status_code == 200
```

- キーは**依存関数そのもの**（`get_db`）、値は**差し替え後の関数**
- テスト終了後は `app.dependency_overrides.clear()` で元に戻すのを忘れない（他のテストに影響しないよう、フィクスチャの`yield`後片付けで行うのが定石）

```python
@pytest.fixture
def client():
    app.dependency_overrides[get_db] = override_get_db
    with TestClient(app) as c:
        yield c
    app.dependency_overrides.clear()
```

### 認証をモックする例

```python
def fake_current_user():
    return {"username": "test-user", "role": "admin"}


app.dependency_overrides[get_current_user] = fake_current_user
```

→ [認証・認可（Security）について](認証・認可（Security）について.md)のJWT検証を毎回本物で通すのではなく、テストでは固定ユーザーを返す関数に差し替えるのが一般的。

---

## 5. 非同期テスト（`httpx.AsyncClient` + `ASGITransport`）

`async def`のテスト関数（DBの非同期セッション等、テスト内でも`await`したい場合）を書きたいときは、`TestClient`ではなく`httpx.AsyncClient`を直接使う。

```bash
pip install pytest-asyncio
```

```python
import pytest
from httpx import ASGITransport, AsyncClient
from main import app


@pytest.mark.asyncio
async def test_read_item_async():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        response = await ac.get("/items/5")
    assert response.status_code == 200
```

- `TestClient`は内部で同期的にイベントループを回すため、多くの場合は`TestClient`で十分。非同期のテストフィクスチャ（非同期DBセッションなど）と混ぜたい場合に`AsyncClient`を使う

---

## 6. StreamingResponse / WebSocketのテスト

```python
def test_stream(client):
    with client.stream("GET", "/stream") as response:
        chunks = list(response.iter_text())
    assert "".join(chunks) != ""
```

- [StreamingResponseについて](%20StreamingResponseについて.md)のエンドポイントは`client.stream()`でチャンクを受け取れる
- WebSocketは`client.websocket_connect("/ws")`をコンテキストマネージャとして使う（詳細は[WebSocketについて](WebSocketについて.md)を参照）

---

## 7. `pytest` の便利な構成例

```
project/
├── main.py
├── conftest.py          # client フィクスチャ、DB初期化など
├── tests/
│   ├── test_users.py
│   ├── test_items.py
│   └── test_auth.py
```

```python
# conftest.py
import pytest
from fastapi.testclient import TestClient
from main import app, get_db
from database import Base, engine, TestingSessionLocal


@pytest.fixture(scope="session", autouse=True)
def setup_db():
    Base.metadata.create_all(bind=engine)   # テスト用DBのテーブル作成
    yield
    Base.metadata.drop_all(bind=engine)


@pytest.fixture
def client():
    def override_get_db():
        db = TestingSessionLocal()
        try:
            yield db
        finally:
            db.close()

    app.dependency_overrides[get_db] = override_get_db
    with TestClient(app) as c:
        yield c
    app.dependency_overrides.clear()
```

---

## ポイントまとめ

- `TestClient`（`fastapi.testclient`）はASGIアプリに直接リクエストを渡すため、実サーバー起動なしで高速にテストできる
- `lifespan`（→[lifespanについて](lifespanについて.md)）を検証したい／使っているアプリは`with TestClient(app) as c:`で使う
- **`app.dependency_overrides`が最重要**: DB・認証などの依存を丸ごとモックへ差し替えてテストする。テスト後は`clear()`で戻す
- `async def`のテストで内部から`await`したい場合は`httpx.AsyncClient` + `ASGITransport`を使う（`pytest-asyncio`必須）
- ストリーミング応答は`client.stream()`でチャンク単位に検証できる
