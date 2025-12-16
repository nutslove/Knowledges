- Mockの方法は主に３つある

## 1. `unittest.mock`（標準ライブラリ）
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

## 2. `pytest-mock`（プラグイン）
- **`mocker` fixture**を提供し、より使いやすいインターフェースを提供する
- `pip install pytest-mock`でインストールが必要
- **`pytest-mock`は`unittest.mock`のラッパーなので、基本的な機能は同じだが、fixtureとして使えるので書き方がシンプルになる**

```python
def test_something(mocker):
    # patchのラッパー
    mock_func = mocker.patch('<対象module名>.<対象function>') # Mockする対象の関数やメソッドを指定
    mock_func.return_value = "mocked"
    
    # spy: 実際の関数を呼びつつ、呼び出しを記録
    spy = mocker.spy(obj, 'method')
```

## 3. `monkeypatch`（pytest組み込みfixture）
- pytestに標準で付属するfixtureで、属性や環境変数の一時的な変更に便利

```python
def test_something(monkeypatch):
    # 属性の置き換え
    monkeypatch.setattr('module.CONSTANT', 'new_value')
    
    # 環境変数の設定
    monkeypatch.setenv('API_KEY', 'test_key')
    
    # dictのアイテム設定
    monkeypatch.setitem(some_dict, 'key', 'value')
```