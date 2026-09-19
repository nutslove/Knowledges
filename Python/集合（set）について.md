## 集合（set）
- **重複しない**要素の集まりを表すデータ型
- ミュータブル（変更可能）な`set`型とイミュータブル（変更不可能）な`frozenset`型がある
- `{}`（※空の場合は dict になる点に注意）または `set()` で作成できる
  ```python
  d = {}  # 空の場合は dict になる点に注意
  print(type(d))  # <class 'dict'>

  s = set()
  print(type(s))  # <class 'set'>

  # 中身がある set() の例(イテラブルなら何でも渡せる)
  s2 = set([1, 2, 2, 3])       # リスト
  print(s2)  # {1, 2, 3}

  s3 = set((1, 2, 2, 3))       # タプル
  print(s3)  # {1, 2, 3}

  s4 = set("hello")            # 文字列(1文字ずつ要素になる)
  print(s4)  # {'h', 'e', 'l', 'o'}

  s5 = set(range(5))           # range
  print(s5)  # {0, 1, 2, 3, 4}

  s6 = set({"a": 1, "b": 2})   # 辞書(キーのみが対象になる)
  print(s6)  # {'a', 'b'}

  s7 = set({1, 2, 3})          # 他の set(コピーになる)
  print(s7)  # {1, 2, 3}

  s8 = set(x for x in range(3))  # ジェネレータ
  print(s8)  # {0, 1, 2}
  ```
- 主な操作
  - 要素の追加: `add()`, `update()`
  - 要素の削除: `remove()`, `discard()`, `pop()`
  - 集合演算: 和集合（`|`）、積集合（`&`）、差集合（`-`）、対称差集合（`^`）
- 例  
  ```python
  numbers = {1, 2, 3, 4, 5}
  print(numbers)  # {1, 2, 3, 4, 5}
  numbers.add(6)
  print(numbers)  # {1, 2, 3, 4, 5, 6}
  numbers.remove(3)
  print(numbers)  # {1, 2, 4, 5, 6}
  evens = {2, 4, 6, 8}
  intersection = numbers & evens
  print(intersection)  # {2, 4, 6}

  a = {1, 2, 3}
  b = {3, 4, 5}
  a | b  # 和集合 {1, 2, 3, 4, 5}
  a & b  # 積集合 {3}
  a - b  # 差集合 {1, 2}
  a ^ b  # 対称差集合 {1, 2, 4, 5}
  ```

### `remove()`と`discard()`の違い
- `remove(elem)`: 指定した要素が存在しない場合、`KeyError`例外を発生させる
- `discard(elem)`: 指定した要素が存在しなくても例外を発生させない

```python
s = {1, 2, 3}

s.remove(4)   # KeyError
s.discard(4)  # 何も起きない
```

## listとの違い
| 特徴 | set | list |
| --- | --- | --- |
| 重複要素 | 保持**不可** | 保持**可** |
| 順序 | 無し（順序保証されない） | 有り（挿入順に保持） |
| 主な用途 | 要素の存在確認、集合演算 | 順序付きデータの保持、インデックスアクセス |
| パフォーマンス | 高速な要素存在確認（平均O(1)）※`in` での存在チェックが速い | 線形探索（O(n)） |
| ミュータブル/イミュータブル | `set`（ミュータブル）、`frozenset`（イミュータブル） | `list`（ミュータブル）、`tuple`（イミュータブル） |

### 存在チェック
```python
lst = list(range(1000000))
st = set(lst)

10 in lst  # 遅い
10 in st   # 速い
```

## 型ヒントとしての `set[]`
- `set[Ingredient]`のように書くと、「`Ingredient`型の要素を持つ`set`」という意味の型ヒント（type hint）になる
- Python 3.9以降、`list[int]`や`dict[str, int]`のようにビルトインの型に`[]`で要素の型を指定できる（PEP 585）
- 実行時には普通の`set`と同じで、型チェッカー（mypyなど）や可読性のための注釈にすぎない
- `set`の要素はハッシュ可能である必要があるため、`@dataclass(frozen=True)`でイミュータブルにしたクラスなどが対象になる

```python
from dataclasses import dataclass
from enum import Enum, auto
import datetime


class TablespoonMeasure(Enum):
    OSAJI = auto()
    KOSAJI = auto()
    CUP = auto()

# メイン具材
class Broth(Enum):
    FU = auto()
    TOFU = auto()
    WAKAME = auto()

# 追加具材
@dataclass(frozen=True)  # frozen=True でイミュータブル（≒ハッシュ可能）にする
class Ingredient:
    name: str
    amount: float = 1
    units: TablespoonMeasure = TablespoonMeasure.CUP

# レシピ
@dataclass
class Recipe:
    aromatics: set[Ingredient]  # 香辛料
    broth: Broth
    vegetables: set[Ingredient]  # 野菜
    meats: set[Ingredient]
    time_to_cook: datetime.timedelta


recipe = Recipe(
    aromatics={Ingredient("生姜"), Ingredient("にんにく")},
    broth=Broth.TOFU,
    vegetables={Ingredient("大根")},
    meats=set(),
    time_to_cook=datetime.timedelta(minutes=30),
)
```