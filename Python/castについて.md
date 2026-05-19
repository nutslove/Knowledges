# castとは
`cast` は Python 標準ライブラリ `typing` に含まれる関数で、**静的型チェッカー（mypy / pyright など）に「この値はこの型だとみなして」と伝えるためのツール**。

```python
from typing import cast
```

> [!NOTE]
> `cast` は**実行時には何もしない**（引数をそのまま返すだけ）。型変換は行われない。あくまで型チェッカー向けのヒント。

## 基本構文

```python
cast(typ, val)
```

| パラメータ | 説明 |
|-----------|------|
| `typ` | この値が持つとみなしたい型 |
| `val` | 対象の値 |

**戻り値**: `val` をそのまま返す（実行時は no-op）

---

## 基本的な使い方

### 例1: 型チェッカーが推論できない型を明示

```python
from typing import cast

def get_data() -> object:
    return {"name": "田中", "age": 30}

data = get_data()
# data は object 型と推論されるが、実際は dict であることが分かっている
user = cast(dict, data)
user["name"]  # 型チェッカーがエラーを出さない
```

### 例2: JSON パース結果に型を付ける

```python
import json
from typing import cast, TypedDict

class User(TypedDict):
    name: str
    age: int

raw = json.loads('{"name": "田中", "age": 30}')
# json.loads の戻り値は Any なので、TypedDict で型を明示
user = cast(User, raw)
user["name"]  # str として扱われる
```

### 例3: ダウンキャスト（基底クラス → 派生クラス）

```python
from typing import cast

class Animal:
    pass

class Dog(Animal):
    def bark(self) -> str:
        return "Woof!"

def get_animal() -> Animal:
    return Dog()

animal = get_animal()
dog = cast(Dog, animal)
dog.bark()  # 型チェッカーがエラーを出さない
```

---

## 実践的なユースケース

### ユースケース1: Mock オブジェクトの型ヒント

```python
from typing import cast
from unittest.mock import Mock, MagicMock

class UserService:
    def get_user(self, user_id: int) -> dict:
        ...

# Mock を UserService 型として扱う（IDE 補完が効く）
service = cast(UserService, Mock(spec=UserService))
service.get_user(1)
```

### ユースケース2: 環境変数の型変換後

```python
import os
from typing import cast

# os.environ.get の戻り値は str | None
port_str = os.environ.get("PORT", "8000")
port = int(port_str)

# 後で None でないことが保証された変数として扱いたい場合
api_key = cast(str, os.environ.get("API_KEY"))  # None ではないと宣言
```

### ユースケース3: 辞書からの値取り出し

```python
from typing import cast, Any

config: dict[str, Any] = {"timeout": 30, "retries": 3}

# config["timeout"] は Any 型なので int として扱いたい
timeout = cast(int, config["timeout"])
```

---

## `cast` vs `isinstance` vs `# type: ignore` vs `assert`

| 手段 | 実行時チェック | 静的型チェック | 用途 |
|------|---------------|---------------|------|
| `cast(T, val)` | なし（no-op） | 型を `T` として扱う | 型チェッカーへのヒント |
| `isinstance(val, T)` | あり（True/False） | 型ガードとして機能 | 実際に型を確認したい |
| `# type: ignore` | なし | 行ごと型チェックを無効化 | 型エラーを抑制（最終手段） |
| `assert isinstance(val, T)` | あり（失敗時例外） | 型ガードとして機能 | 確認+型絞り込み |

### 使い分けの例

```python
from typing import cast

def process(value: object) -> None:
    # ❌ cast: 実際にチェックしないので、間違っていてもバグになる
    s1 = cast(str, value)

    # ✅ isinstance: 実行時にも確認できる
    if isinstance(value, str):
        s2 = value  # 型チェッカーが str として扱う

    # ✅ assert: 「ここは絶対 str」と保証したい場合
    assert isinstance(value, str)
    s3 = value  # 以降は str として扱われる
```

**原則**: 実行時にチェックできるなら `isinstance` / `assert` を使う。型チェッカーが推論できないだけで実際は安全な場合のみ `cast` を使う。

---

## 注意点

### 実行時には型変換されない

```python
from typing import cast

x = cast(int, "123")
print(type(x))  # → <class 'str'>  ← 文字列のまま！
x + 1            # ❌ TypeError: can only concatenate str (not "int") to str
```

`cast(int, "123")` は `int("123")` ではない。実際に変換したい場合は `int(...)` を使う。

### 間違った cast はバグを隠す

```python
from typing import cast

data: object = 42
s = cast(str, data)  # 型チェッカーはパスするが、実際は int
s.upper()             # ❌ 実行時に AttributeError
```

`cast` は「型チェッカーへの嘘」が書ける。安易に使うと型システムの恩恵を失う。

### Generic 型のキャスト

```python
from typing import cast

items: list = [1, 2, 3]
typed_items = cast(list[int], items)  # OK
```

---

## まとめ

| 項目 | 内容 |
|------|------|
| 目的 | 静的型チェッカーに値の型を明示する |
| 実行時の動作 | 何もしない（値をそのまま返す） |
| 主な用途 | `Any` / `object` からの絞り込み、Mock の型付け、JSON パース結果 |
| 代替手段 | `isinstance`（実行時チェック付き）、`assert isinstance`、`# type: ignore` |
| 関連 | [[isinstance関数について]], [[Pydantic, TypedDict, typingについて]] |
