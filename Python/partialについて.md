# partialとは
`partial` は Python 標準ライブラリ `functools` に含まれる関数で、**既存の関数の引数を一部固定（部分適用）した新しい関数を作成する**ためのツール。

```python
from functools import partial
```

## 基本構文

```python
partial(func, *args, **kwargs)
```

| パラメータ | 説明 |
|-----------|------|
| `func` | 元となる関数 |
| `*args` | 固定したい位置引数 |
| `**kwargs` | 固定したいキーワード引数 |

**戻り値**: 引数が部分適用された新しい callable オブジェクト

---

## 基本的な使い方

### 例1: キーワード引数の固定

```python
from functools import partial

def greet(greeting, name):
    return f"{greeting}, {name}!"

# greeting を "Hello" に固定
say_hello = partial(greet, greeting="Hello")

say_hello(name="田中")  # → "Hello, 田中!"
say_hello(name="山田")  # → "Hello, 山田!"
```

### 例2: 位置引数の固定

```python
from functools import partial

def power(base, exponent):
    return base ** exponent

# base を 2 に固定（2のn乗を計算する関数）
power_of_two = partial(power, 2)

power_of_two(3)   # → 8  (2^3)
power_of_two(10)  # → 1024  (2^10)
```

### 例3: 複数の引数を固定

```python
from functools import partial

def send_email(to, subject, body, cc=None, priority="normal"):
    return f"To: {to}, Subject: {subject}, Priority: {priority}"

# 緊急メール用の関数を作成
urgent_email = partial(send_email, priority="high", cc="manager@example.com")

urgent_email("user@example.com", "緊急", "サーバーダウン")
```

---

## 実践的なユースケース

### ユースケース1: ロガーの作成

```python
from functools import partial

def log_message(level, message):
    print(f"[{level}] {message}")

log_info = partial(log_message, "INFO")
log_error = partial(log_message, "ERROR")

log_info("アプリケーション開始")  # → [INFO] アプリケーション開始
log_error("接続エラー")          # → [ERROR] 接続エラー
```

### ユースケース2: map との組み合わせ

```python
from functools import partial

def multiply(x, y):
    return x * y

triple = partial(multiply, 3)
numbers = [1, 2, 3, 4, 5]

list(map(triple, numbers))  # → [3, 6, 9, 12, 15]
```

### ユースケース3: LangChain でのトークンカウンター設定

```python
from functools import partial
from langchain_core.messages.utils import count_tokens_approximately

# chars_per_token を 2.0 に固定
custom_counter = partial(count_tokens_approximately, chars_per_token=2.0)

middleware = SummarizationMiddleware(
    model=model,
    trigger=("tokens", 50000),
    token_counter=custom_counter,
)
```

---

## `partial` vs `lambda`

```python
from functools import partial

def add(a, b, c):
    return a + b + c

# partial
add_10 = partial(add, 10)

# lambda
add_10_lambda = lambda b, c: add(10, b, c)

# 結果は同じ
add_10(20, 30)         # → 60
add_10_lambda(20, 30)  # → 60
```

| 観点 | `partial` | `lambda` |
|------|-----------|----------|
| 可読性 | 意図が明確 | 複雑になると読みにくい |
| イントロスペクション | `func`, `args`, `keywords` 属性でアクセス可能 | 不可 |
| pickle 対応 | 可能 | 不可（ローカルスコープの場合） |
| 柔軟性 | 引数の固定のみ | 任意のロジックを記述可能 |

### イントロスペクション

```python
from functools import partial

def greet(greeting, name, punctuation="!"):
    return f"{greeting}, {name}{punctuation}"

hello = partial(greet, "Hello", punctuation="!!!")

hello.func      # → <function greet at 0x...>
hello.args      # → ('Hello',)
hello.keywords  # → {'punctuation': '!!!'}
```

---

## 注意点

### 可変オブジェクトに注意

```python
from functools import partial

def append_to_list(lst, item):
    lst.append(item)
    return lst

my_list = [1, 2, 3]
append_func = partial(append_to_list, my_list)

append_func(4)  # → [1, 2, 3, 4]
append_func(5)  # → [1, 2, 3, 4, 5]
my_list         # → [1, 2, 3, 4, 5]  ← 元のリストも変更される
```

### キーワード引数は上書き可能

```python
say_hello = partial(greet, greeting="Hello")

say_hello(name="田中")                   # → "Hello, 田中!"
say_hello(name="田中", greeting="Hi")    # → "Hi, 田中!"
```

---

## まとめ

| 項目 | 内容 |
|------|------|
| 目的 | 関数の引数を一部固定して新しい関数を作成 |
| メリット | コードの再利用性向上、可読性向上 |
| 主な用途 | 設定値の固定、コールバック、高階関数との組み合わせ |
| 代替手段 | `lambda`（制限あり） |
