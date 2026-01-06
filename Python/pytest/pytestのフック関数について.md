## フック関数とは
- **pytestが決まったタイミングで呼び出す関数**
- テストのライフサイクルの各ポイントでカスタム処理を差し込む仕組み
- https://docs.pytest.org/en/stable/reference/reference.html#hooks
- **`conftest.py`ファイルに定義することで利用可能**

## 主なフック関数の例
| フック関数名 | 説明 |
|--------------|------|
| `pytest_configure` | テストセッション開始前（収集より前）に呼び出される |
| `pytest_collection_modifyitems` | テスト収集後、実行前に呼び出される |
| `pytest_runtest_setup` | 各テスト関数の実行前に呼び出される |
| `pytest_sessionstart` | テストセッション開始時に呼び出される |
| `pytest_sessionfinish` | テストセッション終了時に呼び出される |

### 例 `conftest.py`
#### 1. 設定・初期化系
```python
def pytest_configure(config):
    """pytest起動時に1回だけ実行される"""
    # カスタムマーカーの登録など
    config.addinivalue_line(
        "markers", "slow: 時間のかかるテスト"
    )

def pytest_unconfigure(config):
    """pytest終了時に1回だけ実行される"""
    # クリーンアップ処理
    print("テスト終了")
```

##### `config`引数について
- pytestの設定オブジェクト（`pytest.Config`オブジェクト）
- pytestの設定情報やコマンドラインオプションにアクセス可能

```python
  def pytest_configure(config):
    # _aws_secret_mockに保存しているのは、後でpytest_unconfigureでstop()するため
      config._aws_secret_mock = patch(...)
      config._aws_secret_mock.start()

  def pytest_unconfigure(config):
      config._aws_secret_mock.stop()  # クリーンアップ
    # ただし、テスト終了時に自動でクリーンアップされるので、
    # pytest_unconfigureは省略しても動く。
```

#### 2. テスト収集系
```python
def pytest_collection_modifyitems(config, items):
    """収集されたテストを加工できる（並び替え、スキップなど）"""
    
    # 例: "slow" マーカー付きのテストを最後に実行
    slow_tests = []
    other_tests = []
    
    for item in items:
        if item.get_closest_marker("slow"):
            slow_tests.append(item)
        else:
            other_tests.append(item)
    
    items[:] = other_tests + slow_tests

def pytest_generate_tests(metafunc):
    """テストのパラメータ化を動的に行う"""
    
    if "db_type" in metafunc.fixturenames:
        metafunc.parametrize("db_type", ["postgres", "mysql", "sqlite"])
```

#### 3. テスト実行系
```python
def pytest_runtest_setup(item):
    """各テストのsetup前に実行"""
    print(f"Setting up: {item.name}")

def pytest_runtest_call(item):
    """各テストの実行時に呼ばれる"""
    pass

def pytest_runtest_teardown(item):
    """各テストのteardown後に実行"""
    print(f"Tearing down: {item.name}")
```

#### 4. レポート・出力系
```python
def pytest_report_header(config):
    """テスト実行前のヘッダーに情報を追加"""
    return "環境: ステージング"

def pytest_terminal_summary(terminalreporter, exitstatus, config):
    """テスト終了後のサマリーに情報を追加"""
    terminalreporter.write_line("カスタムサマリー: 全テスト完了")
```

#### 5. 結果処理系
```python
def pytest_runtest_makereport(item, call):
    """テスト結果をカスタマイズ"""
    # call.when: "setup", "call", "teardown"
    # call.excinfo: 例外情報（あれば）
    pass

import pytest

@pytest.hookimpl(hookwrapper=True)
def pytest_runtest_makereport(item, call):
    """失敗時にスクリーンショットを保存する例"""
    outcome = yield
    report = outcome.get_result()
    
    if report.when == "call" and report.failed:
        # 失敗時の処理
        print(f"テスト失敗: {item.name}")
```