# list・dictの参照渡しと安全なコピー方法

## 1. 基本概念

- Pythonの変数は**オブジェクトへの参照（メモリアドレス）**を保持している
- `list`、`dict`、`set`などのミュータブル（可変）オブジェクトは、同じ参照を持つ変数からの変更が**すべての参照元に影響**する
- `int`、`str`、`tuple`などのイミュータブル（不変）オブジェクトは、変更時に新しいオブジェクトが作成されるため、この問題は起きない

```python
# ミュータブルオブジェクトの参照共有
a = [1, 2, 3]
b = a          # bはaと同じリストオブジェクトを参照
b.append(4)
print(a)       # [1, 2, 3, 4] ← aも変更されている！

# イミュータブルオブジェクトの場合
x = "hello"
y = x
y = y + " world"
print(x)       # "hello" ← xは変更されない
```

## 2. よくある問題パターンと対処法

### 2.1 ループ中に辞書・リストを変更する

ループ中に元のコレクションから要素を削除・追加すると`RuntimeError`が発生する。**コピーを作成してからループ**することで安全に変更できる。

```python
# ❌ ループ中に辞書から項目を削除するとエラー
tracking_ids = {"id1": 100, "id2": 200, "id3": 300}
for key, value in tracking_ids.items():
    if value > 150:
        del tracking_ids[key]  # RuntimeError: dictionary changed size during iteration

# ✅ list()でコピーを作成してからループ
tracking_ids = {"id1": 100, "id2": 200, "id3": 300}
for key, value in list(tracking_ids.items()):  # コピーを作成
    if value > 150:
        del tracking_ids[key]  # 安全に削除できる
print(tracking_ids)  # {'id1': 100}

# ✅ リストの場合も同様
items = [1, 2, 3, 4, 5]
for item in list(items):  # コピーを作成
    if item % 2 == 0:
        items.remove(item)
print(items)  # [1, 3, 5]
```

### 2.2 関数に渡したオブジェクトが変更される

関数にミュータブルオブジェクトを渡すと、関数内での変更が**呼び出し元にも反映**される。元のオブジェクトを保護するには、**コピーを渡す**。

```python
# ❌ 関数内で辞書を変更すると、呼び出し元も変わる
def process(data):
    data["processed"] = True
    del data["raw"]

state = {"raw": "value", "id": 1}
process(state)
print(state)  # {'id': 1, 'processed': True} ← 元のstateが変更されている！

# ✅ dict()やcopy()でコピーを渡す
state = {"raw": "value", "id": 1}
process(dict(state))        # dict()でコピーを渡す
# または
process(state.copy())       # copy()でコピーを渡す
print(state)                # {'raw': 'value', 'id': 1} ← 元のstateは変更されない
```

### 2.3 デフォルト引数にミュータブルオブジェクトを使用する

関数のデフォルト引数は**定義時に1回だけ評価**され、呼び出しごとに再利用される。

```python
# ❌ デフォルト値にリストを直接書くと共有される
def add_item(item, lst=[]):
    lst.append(item)
    return lst

print(add_item("A"))  # ['A']
print(add_item("B"))  # ['A', 'B'] ← 前の呼び出しの結果が残っている！

# ✅ Noneをデフォルト値にして関数内で初期化する
def add_item(item, lst=None):
    if lst is None:
        lst = []
    lst.append(item)
    return lst

print(add_item("A"))  # ['A']
print(add_item("B"))  # ['B'] ← 独立したリスト
```

> [!NOTE]
> `@dataclass`の場合は`field(default_factory=list)`で同じ問題を解決できる。詳細は [dataclassについて.md](dataclassについて.md) を参照。

## 3. コピーの種類

### 3.1 シャローコピー（浅いコピー）

最上位のオブジェクトだけを複製し、内部のオブジェクトは**参照を共有**する。

```python
import copy

# シャローコピーの方法
original = [1, [2, 3], [4, 5]]

shallow1 = list(original)       # list()で作成
shallow2 = original.copy()      # copy()メソッド
shallow3 = original[:]          # スライスで作成
shallow4 = copy.copy(original)  # copyモジュール

# 辞書の場合
original_dict = {"a": 1, "b": [2, 3]}
shallow_dict1 = dict(original_dict)
shallow_dict2 = original_dict.copy()
shallow_dict3 = copy.copy(original_dict)

# シャローコピーの注意点：ネストしたオブジェクトは共有される
original = {"key": [1, 2, 3]}
shallow = original.copy()
shallow["key"].append(4)
print(original)  # {'key': [1, 2, 3, 4]} ← ネストしたリストは共有されている！
```

### 3.2 ディープコピー（深いコピー）

ネストされたオブジェクトも含め、**すべてを再帰的に複製**する。

```python
import copy

original = {"key": [1, 2, 3], "nested": {"a": [4, 5]}}
deep = copy.deepcopy(original)

deep["key"].append(4)
deep["nested"]["a"].append(6)
print(original)  # {'key': [1, 2, 3], 'nested': {'a': [4, 5]}} ← 元は変更されない
```

### 3.3 使い分け

| 方法 | ネストなし | ネストあり | 速度 |
|------|-----------|-----------|------|
| `dict()` / `list()` / `.copy()` / `[:]` | ✅ 安全 | ❌ 内部は共有 | 速い |
| `copy.deepcopy()` | ✅ 安全 | ✅ 安全 | 遅い |

- ネストしたミュータブルオブジェクトを含まない場合 → **シャローコピーで十分**
- ネストしたミュータブルオブジェクトを含む場合 → **ディープコピーが必要**

## 4. 実践例

### LangGraphのstateを関数に渡す場合

```python
# ❌ stateのtracking_idsが関数内で変更されてしまう
sub_agent.track_sub_agent_status(state["tracking_ids"])

# ✅ dict()でコピーを渡して元のstateを保護する
sub_agent.track_sub_agent_status(dict(state["tracking_ids"]))
```

### ループ中にtracking_idsから完了した項目を削除する場合

```python
# ✅ list()でコピーを作成してループ
for tracking_id, start_time in list(tracking_ids.items()):
    if time.time() - start_time <= 630:
        # 追跡処理...
        pass
    else:
        del tracking_ids[tracking_id]  # 安全に削除できる
```
