- Mockの方法は主に３つある

# 1. `unittest.mock`（標準ライブラリ）
- Python標準の`unittest.mock`モジュールで、pytestでも広く使われる

```python
from unittest.mock import Mock, MagicMock, patch

# Mock: 基本的なモックオブジェクト
mock_obj = Mock()
mock_obj.some_method.return_value = "mocked"

# MagicMock: マジックメソッド（__len__, __iter__など）も自動でモック
magic_mock = MagicMock()
magic_mock.__len__.return_value = 5

# patch: 特定のオブジェクトを一時的に置き換える
@patch('module.ClassName')
def test_something(mock_class):
    mock_class.return_value.method.return_value = "mocked"
```

> [!NOTE]  
> `patch`の第２引数に置き換える値を指定できる（省略すると自動的にMagicMockが作られる）  
> ```python
> # conftest.py
> def pytest_configure(config):
>    os.environ["SECRETMANAGER_SECRET_ID"] = "test-secret-id"
>
>    # インポート前にモックを設定
>    mock_secret = MagicMock(return_value={
>        "SLACK_BOT_TOKEN": "xoxb-test-token",
>        "SLACK_SIGNING_SECRET": "test-signing-secret",
>        "LANGFUSE_SECRET_KEY": "test-langfuse-secret",
>        "LANGFUSE_PUBLIC_KEY": "test-langfuse-public",
>    })
>    config._aws_secret_mock = patch("common.aws_secret_manager.get_secret", mock_secret)
>    config._aws_secret_mock.start()
>
>    # slack_bolt.Appのモック（トークン検証をスキップ）
>    config._slack_app_mock = patch("slack_bolt.App")
>    config._slack_app_mock.start()

## 1-1. 関数の戻り値を固定する

```python
from unittest.mock import Mock

def test_simple_return():
    mock_func = Mock(return_value=10)

    assert mock_func() == 10
```

## 1-2. 引数によって戻り値を変える（`side_effect`）

```python
from unittest.mock import Mock

def test_side_effect_by_args():
    def fake(x):
        return x * 2

    mock_func = Mock(side_effect=fake)

    assert mock_func(3) == 6
```

## 1-3. 例外を発生させる

```python
from unittest.mock import Mock

def test_raise_exception():
    mock_func = Mock(side_effect=ValueError("boom"))

    try:
        mock_func()
    except ValueError:
        assert True
```

👉 **APIエラー・DBエラーの再現**で超頻出

## 1-4. メソッドチェーンをモックする

```python
from unittest.mock import Mock

def test_method_chain():
    mock_obj = Mock()
    mock_obj.a().b().c.return_value = "ok"

    assert mock_obj.a().b().c() == "ok"
```

👉 `client.get().json()` みたいなケース

## 1-5. 呼び出し回数・引数の検証
```python
from unittest.mock import Mock

def test_call_assertions():
    mock_func = Mock()

    mock_func(1)
    mock_func(2)

    assert mock_func.call_count == 2
    mock_func.assert_any_call(1)
    mock_func.assert_called_with(2)  # 最後の呼び出しを検証
```

### 検証メソッド一覧

| メソッド | 検証内容 |
|---------|---------|
| `assert_called()` | 1回以上呼ばれたか |
| `assert_called_once()` | ちょうど1回呼ばれたか（引数は問わない） |
| `assert_called_with(...)` | **最後の呼び出し**が指定した引数か |
| `assert_called_once_with(...)` | 1回だけ＆指定した引数で呼ばれたか |
| `assert_any_call(...)` | 指定した引数での呼び出しが1回でもあるか |
| `assert_not_called()` | 一度も呼ばれていないか |

### `assert_called_once_with` の使い方
```python
def test_called_once_with():
    mock_func = Mock()
    mock_func("hello", key="value")

    # ✅ 1回だけ & 正しい引数で呼ばれた
    mock_func.assert_called_once_with("hello", key="value")
```

👉 **「この関数が正しい引数で1回だけ呼ばれたこと」を保証したいときに使う（最頻出）**

### ⚠️ タイポに注意
```python
mock_func.assert_called_once_With("hello")  # ❌ タイポ！エラーにならない
```

Mockは存在しないメソッド名でもエラーにならないため、テストが常にパスしてしまう。`autospec=True` を使うと防げる。

## 1-6. `MagicMock`（マジックメソッド）

```python
from unittest.mock import MagicMock

def test_magic_len():
    m = MagicMock()
    m.__len__.return_value = 3

    assert len(m) == 3
```

```python
def test_iterable():
    m = MagicMock()
    m.__iter__.return_value = iter([1, 2, 3])

    assert list(m) == [1, 2, 3]
```

### MagicMockについて (Mockとの違い)
- `Mock`の拡張版で、`__len__`, `__iter__`, `__getitem__`などの**マジックメソッド**が最初から使える
```python
from unittest.mock import Mock, MagicMock

# Mockの場合
mock = Mock()
len(mock)  # ❌ TypeError: object of type 'Mock' has no len()

# MagicMockの場合
magic = MagicMock()
len(magic)  # ✅ 0（デフォルト値が返る）
```

#### マジックメソッドとは
- Pythonの特殊メソッド（`__xxx__` の形式）
```python
magic = MagicMock()

# これらが全部最初から動く
len(magic)           # __len__
iter(magic)          # __iter__
str(magic)           # __str__
magic[0]             # __getitem__
magic + 1            # __add__
bool(magic)          # __bool__
```

- Mockでマジックメソッドを使いたい場合  
    ```python
    mock = Mock()
    mock.__len__ = Mock(return_value=5)
    len(mock)  # ✅ 5
    ```

## 1-7. `patch`（超重要）

### ❗「**importした場所をpatchする**」

```python
# service.py
from utils import now

def get_time():
    return now()
```

```python
# test_service.py
from unittest.mock import patch

@patch("service.now")
def test_patch_function(mock_now):
    mock_now.return_value = "2025-01-01"

    from service import get_time
    assert get_time() == "2025-01-01"
```

## 1-8. `patch` を context manager で使う

```python
from unittest.mock import patch

def test_patch_with():
    with patch("module.func") as mock_func:
        mock_func.return_value = 1
        assert module.func() == 1
```

## 1-9. クラスのモック

```python
@patch("module.Client")
def test_mock_class(mock_client):
    instance = mock_client.return_value
    instance.fetch.return_value = "data"

    result = instance.fetch()
    assert result == "data"
```

> [!NOTE]  
> ### `patch`の使い方
> 通常、patchはデコレータやコンテキストマネージャとして使う
> #### デコレータとして
> ```python
> @patch("module.function", mock_value)
> def test_something():
>      ...
> ```
>
> #### コンテキストマネージャとして 
> ```python
> with patch("module.function", mock_value):
>      ...
> ```
>
> しかし`pytest_configure`ではどちらも使えないため、手動で`start()`と`stop()`を呼び出して開始・停止する必要がある
>
> ```python
> patcher = patch("module.function", mock_value)
> patcher.start()   # モックを開始
> # ... モックが有効な状態 ...
> patcher.stop()    # モックを停止


---

# 2. `pytest-mock`（プラグイン）
- **`mocker` fixture**を提供し、より使いやすいインターフェースを提供する
- `pip install pytest-mock`でインストールが必要
- **`pytest-mock`は`unittest.mock`のラッパーなので、基本的な機能は同じだが、fixtureとして使えるので書き方がシンプルになる**

👉 **実務ではこれが一番使われる**

```python
def test_something(mocker):
    # patchのラッパー
    mock_func = mocker.patch('<対象module名>.<対象function>') # Mockする対象の関数やメソッドを指定
    mock_func.return_value = "mocked"
    
    # spy: 実際の関数を呼びつつ、呼び出しを記録
    spy = mocker.spy(obj, 'method')
```

## 2-1. 関数をパッチする（最頻出）

```python
def test_patch_function(mocker):
    mock = mocker.patch("module.func")
    mock.return_value = 42

    # 以下のようにpatchの中でreturn_valueを設定することもできる
    mocker.patch("module.func", return_value=42)

    assert module.func() == 42
```

## 2-2. クラスをパッチする

```python
def test_patch_class(mocker):
    mock_cls = mocker.patch("module.Client")
    instance = mock_cls.return_value
    instance.get.return_value = "ok"

    assert instance.get() == "ok"
```

## 2-3. `side_effect`（例外）

```python
def test_side_effect_exception(mocker):
    mock = mocker.patch("module.func")
    mock.side_effect = RuntimeError

    # 以下のようにpatchの中でside_effectを設定することもできる
    mocker.patch("module.func", side_effect=RuntimeError)

    import pytest
    with pytest.raises(RuntimeError):
        module.func()
```

## 2-4. 複数回呼ばれたことを検証

```python
def test_call_count(mocker):
    mock = mocker.patch("module.func")

    module.func()
    module.func()

    assert mock.call_count == 2
```

## 2-5. 引数チェック

```python
def test_called_with(mocker):
    mock = mocker.patch("module.func")

    module.func(1, 2)
    mock.assert_called_with(1, 2)
```

## 2-6. `spy`（実装を動かしつつ監視）

```python
def test_spy(mocker):
    class Foo:
        def add(self, x, y):
            return x + y

    foo = Foo()
    spy = mocker.spy(foo, "add")

    result = foo.add(1, 2)

    assert result == 3
    spy.assert_called_once_with(1, 2)
```

👉 **副作用は残したいが、呼び出しは確認したい**とき

## 2-7. `reset_mock`

```python
def test_reset_mock(mocker):
    mock = mocker.patch("module.func")
    module.func()

    mock.reset_mock()
    assert mock.call_count == 0
```

---

# 3. `monkeypatch`（pytest組み込みfixture）
- pytestに標準で付属するfixtureで、属性や環境変数の一時的な変更に便利

👉 **「値の差し替え」に特化**

```python
def test_something(monkeypatch):
    # 属性の置き換え
    monkeypatch.setattr('module.CONSTANT', 'new_value')
    
    # 環境変数の設定
    monkeypatch.setenv('API_KEY', 'test_key')
    
    # dictのアイテム設定
    monkeypatch.setitem(some_dict, 'key', 'value')
```

## 3-1. 定数の差し替え

```python
# config.py
TIMEOUT = 30
```

```python
def test_constant(monkeypatch):
    monkeypatch.setattr("config.TIMEOUT", 1)

    import config
    assert config.TIMEOUT == 1
```

## 3-2. 関数を差し替える（簡易Mock）

```python
def fake_func():
    return "fake"

def test_replace_func(monkeypatch):
    monkeypatch.setattr("module.func", fake_func)

    assert module.func() == "fake"
```

## 3-3. 環境変数

```python
def test_env(monkeypatch):
    monkeypatch.setenv("ENV", "test")

    import os
    assert os.getenv("ENV") == "test"
```

## 3-4. 環境変数削除

```python
def test_unset_env(monkeypatch):
    monkeypatch.delenv("API_KEY", raising=False)
```

## 3-5. dict の書き換え

```python
def test_dict(monkeypatch):
    data = {"a": 1}
    monkeypatch.setitem(data, "a", 2)

    assert data["a"] == 2
```

## 3-6. クラスの属性差し替え

```python
class Foo:
    value = 10

def test_class_attr(monkeypatch):
    monkeypatch.setattr(Foo, "value", 99)

    assert Foo.value == 99
```

---

# シナリオごとの例

## シナリオ1: 外部APIを呼び出す関数のテスト

### テスト対象のコード（`myapp/user_service.py`）
```python
import requests

def get_user(user_id: int) -> dict:
    """外部APIからユーザー情報を取得する"""
    response = requests.get(f"https://api.example.com/users/{user_id}")
    if response.status_code == 200:
        return response.json()
    elif response.status_code == 404:
        raise ValueError(f"User {user_id} not found")
    else:
        raise ConnectionError("API error")
```

### テストコード（`tests/test_user_service.py`）
```python
from myapp.user_service import get_user
import pytest

# ============================================
# pytest-mock を使う場合（推奨）
# ============================================

def test_get_user_success(mocker):
    """正常にユーザーが取得できるケース"""
    # requests.getをモックする
    # ※ "myapp.user_service.requests.get" とする（importしている場所を指定）
    mock_get = mocker.patch("myapp.user_service.requests.get")
    
    # モックの戻り値を設定
    mock_get.return_value.status_code = 200
    mock_get.return_value.json.return_value = {"id": 1, "name": "Alice"}
    
    # 実行
    result = get_user(1)
    
    # 検証
    assert result == {"id": 1, "name": "Alice"}
    mock_get.assert_called_once_with("https://api.example.com/users/1")


def test_get_user_not_found(mocker):
    """ユーザーが見つからない場合、ValueErrorが発生する"""
    mock_get = mocker.patch("myapp.user_service.requests.get")
    mock_get.return_value.status_code = 404
    
    with pytest.raises(ValueError, match="User 1 not found"):
        get_user(1)


def test_get_user_api_error(mocker):
    """APIがエラーを返す場合、ConnectionErrorが発生する"""
    mock_get = mocker.patch("myapp.user_service.requests.get")
    mock_get.return_value.status_code = 500
    
    with pytest.raises(ConnectionError, match="API error"):
        get_user(1)


def test_get_user_network_error(mocker):
    """ネットワークエラーが発生する場合"""
    mock_get = mocker.patch("myapp.user_service.requests.get")
    # side_effectで例外を発生させる
    mock_get.side_effect = requests.exceptions.Timeout("Connection timeout")
    
    with pytest.raises(requests.exceptions.Timeout):
        get_user(1)
```

**ポイント**: 実際にAPIを呼ばずに、様々なレスポンスパターンをテストできる

---

## シナリオ2: データベースを使う関数のテスト

### テスト対象のコード（`myapp/order_service.py`）
```python
from myapp.database import get_db_connection

def get_order_total(order_id: int) -> int:
    """注文の合計金額を取得する"""
    conn = get_db_connection()
    cursor = conn.cursor()
    cursor.execute("SELECT total FROM orders WHERE id = ?", (order_id,))
    row = cursor.fetchone()
    if row is None:
        raise ValueError(f"Order {order_id} not found")
    return row[0]
```

### テストコード
```python
from myapp.order_service import get_order_total
import pytest

def test_get_order_total(mocker):
    """正常に合計金額が取得できる"""
    # get_db_connectionをモック
    mock_get_conn = mocker.patch("myapp.order_service.get_db_connection")
    
    # モックのチェーンを設定（conn.cursor().fetchone()）
    mock_cursor = mocker.MagicMock()
    mock_cursor.fetchone.return_value = (1500,)  # 合計1500円
    mock_get_conn.return_value.cursor.return_value = mock_cursor
    
    result = get_order_total(123)
    
    assert result == 1500
    mock_cursor.execute.assert_called_once_with(
        "SELECT total FROM orders WHERE id = ?", 
        (123,)
    )


def test_get_order_total_not_found(mocker):
    """注文が存在しない場合"""
    mock_get_conn = mocker.patch("myapp.order_service.get_db_connection")
    mock_cursor = mocker.MagicMock()
    mock_cursor.fetchone.return_value = None  # 該当なし
    mock_get_conn.return_value.cursor.return_value = mock_cursor
    
    with pytest.raises(ValueError, match="Order 999 not found"):
        get_order_total(999)
```

---

## シナリオ3: 環境変数を使う関数のテスト

### テスト対象のコード（`myapp/config.py`）
```python
import os

def get_database_url() -> str:
    """環境変数からDB接続URLを取得"""
    url = os.environ.get("DATABASE_URL")
    if url is None:
        raise RuntimeError("DATABASE_URL is not set")
    return url

def is_debug_mode() -> bool:
    """デバッグモードかどうか"""
    return os.environ.get("DEBUG", "false").lower() == "true"
```

### テストコード（monkeypatchを使う）
```python
from myapp.config import get_database_url, is_debug_mode
import pytest

def test_get_database_url(monkeypatch):
    """環境変数が設定されている場合"""
    monkeypatch.setenv("DATABASE_URL", "postgres://localhost/testdb")
    
    result = get_database_url()
    
    assert result == "postgres://localhost/testdb"


def test_get_database_url_not_set(monkeypatch):
    """環境変数が設定されていない場合"""
    monkeypatch.delenv("DATABASE_URL", raising=False)
    
    with pytest.raises(RuntimeError, match="DATABASE_URL is not set"):
        get_database_url()


def test_is_debug_mode_true(monkeypatch):
    monkeypatch.setenv("DEBUG", "true")
    assert is_debug_mode() is True


def test_is_debug_mode_false(monkeypatch):
    monkeypatch.setenv("DEBUG", "false")
    assert is_debug_mode() is False


def test_is_debug_mode_not_set(monkeypatch):
    """環境変数がない場合はデフォルトでfalse"""
    monkeypatch.delenv("DEBUG", raising=False)
    assert is_debug_mode() is False
```

**ポイント**: 環境変数の変更は `monkeypatch` が一番シンプル

---

## シナリオ4: 現在時刻に依存する関数のテスト

### テスト対象のコード（`myapp/greeting.py`）
```python
from datetime import datetime

def get_greeting() -> str:
    """時間帯に応じた挨拶を返す"""
    hour = datetime.now().hour
    if 5 <= hour < 12:
        return "おはようございます"
    elif 12 <= hour < 18:
        return "こんにちは"
    else:
        return "こんばんは"
```

### テストコード
```python
from myapp.greeting import get_greeting
from datetime import datetime

def test_greeting_morning(mocker):
    """朝の挨拶"""
    # datetime.nowをモック
    mock_datetime = mocker.patch("myapp.greeting.datetime")
    mock_datetime.now.return_value = datetime(2024, 1, 1, 8, 0, 0)  # 8:00
    
    assert get_greeting() == "おはようございます"


def test_greeting_afternoon(mocker):
    """昼の挨拶"""
    mock_datetime = mocker.patch("myapp.greeting.datetime")
    mock_datetime.now.return_value = datetime(2024, 1, 1, 14, 0, 0)  # 14:00
    
    assert get_greeting() == "こんにちは"


def test_greeting_evening(mocker):
    """夜の挨拶"""
    mock_datetime = mocker.patch("myapp.greeting.datetime")
    mock_datetime.now.return_value = datetime(2024, 1, 1, 20, 0, 0)  # 20:00
    
    assert get_greeting() == "こんばんは"
```

---

## シナリオ5: クラスのメソッドをテスト（一部だけモック）

### テスト対象のコード（`myapp/notification.py`）
```python
class NotificationService:
    def __init__(self, api_key: str):
        self.api_key = api_key
    
    def _send_email(self, to: str, message: str) -> bool:
        """実際にメールを送る（外部APIを呼ぶ）"""
        # 実際の送信処理...
        pass
    
    def notify_user(self, user_email: str, event: str) -> str:
        """ユーザーに通知を送る"""
        message = f"イベント発生: {event}"
        success = self._send_email(user_email, message)
        if success:
            return "通知を送信しました"
        else:
            return "通知の送信に失敗しました"
```

### テストコード
```python
from myapp.notification import NotificationService

def test_notify_user_success(mocker):
    """通知成功のケース"""
    service = NotificationService(api_key="test-key")
    
    # _send_emailだけをモック（他のメソッドは本物）
    mocker.patch.object(service, "_send_email", return_value=True)
    
    result = service.notify_user("user@example.com", "新しいメッセージ")
    
    assert result == "通知を送信しました"
    service._send_email.assert_called_once_with(
        "user@example.com", 
        "イベント発生: 新しいメッセージ"
    )


def test_notify_user_failure(mocker):
    """通知失敗のケース"""
    service = NotificationService(api_key="test-key")
    mocker.patch.object(service, "_send_email", return_value=False)
    
    result = service.notify_user("user@example.com", "エラー")
    
    assert result == "通知の送信に失敗しました"
```

---

## シナリオ6: spy を使って実際の処理も実行する

### テスト対象のコード（`myapp/calculator.py`）
```python
def calculate_tax(price: int, rate: float = 0.1) -> int:
    """税込価格を計算"""
    return int(price * (1 + rate))

def get_total_with_tax(items: list[int]) -> int:
    """商品リストの税込合計を計算"""
    return sum(calculate_tax(price) for price in items)
```

### テストコード
```python
from myapp import calculator

def test_get_total_with_tax(mocker):
    """calculate_taxが各商品に対して呼ばれることを確認"""
    # spyは実際の処理を実行しつつ、呼び出しを記録する
    spy = mocker.spy(calculator, "calculate_tax")
    
    result = get_total_with_tax([100, 200, 300])
    
    # 実際の計算結果
    assert result == 110 + 220 + 330  # = 660
    
    # 呼び出し回数の確認
    assert spy.call_count == 3
    
    # 各呼び出しの引数を確認
    spy.assert_any_call(100)
    spy.assert_any_call(200)
    spy.assert_any_call(300)
```

**ポイント**: `spy`は「実際に動かしつつ、ちゃんと呼ばれてるか確認したい」時に使う

---

## まとめ：いつ何を使うか

| やりたいこと | 使うもの |
|------------|---------|
| 外部API・DBへの呼び出しを偽物に置き換えたい | `mocker.patch()` |
| 環境変数を一時的に変えたい | `monkeypatch.setenv()` |
| 設定値・定数を一時的に変えたい | `monkeypatch.setattr()` |
| クラスの特定メソッドだけモックしたい | `mocker.patch.object()` |
| 実際に動かしつつ呼び出しを記録したい | `mocker.spy()` |
| 例外を発生させたい | `mock.side_effect = Exception()` |
| 呼び出しごとに違う値を返したい | `mock.side_effect = [値1, 値2, ...]` |
