## `@dataclass`について
- Python 3.7で導入された`dataclasses`モジュールの一部で、クラス定義を簡素化するためのデコレータ（`@dataclass`）
- `@dataclass`を使用すると、クラスの属性に基づいて自動的に初期化メソッド（`__init__`）、文字列表現メソッド（`__repr__`）、比較メソッド（`__eq__`など）などが生成される

### 基本的な使い方  
```python
from dataclasses import dataclass

@dataclass
class Person:
    name: str
    age: int
    email: str = None  # デフォルト値を設定可能
```

### 従来のクラス定義と比較（`@dataclass`を使わない場合）  
```python
class PersonTraditional:
    def __init__(self, name: str, age: int, email: str):
        self.name = name
        self.age = age
        self.email = email
    
    def __repr__(self):
        return f"PersonTraditional(name='{self.name}', age={self.age}, email='{self.email}')"
    
    def __eq__(self, other):
        if not isinstance(other, PersonTraditional):
            return False
        return (self.name, self.age, self.email) == (other.name, other.age, other.email)
```

### カスタムメソッドの定義
- `@dataclass`は`__init__`などを自動生成するだけで、**通常のクラスと同様に独自のメソッドを定義できる**
```python
from dataclasses import dataclass

@dataclass
class Point:
    x: int
    y: int

    def distance_from_origin(self) -> float:
        return (self.x ** 2 + self.y ** 2) ** 0.5

p = Point(3, 4)
print(p.distance_from_origin())  # 5.0
```

#### `__post_init__`
- `__init__`が自動生成されるため、初期化直後に追加処理を行いたい場合は`__post_init__`メソッドを定義する
- `__init__`の最後に自動的に呼び出される
```python
@dataclass
class Rectangle:
    width: float
    height: float
    area: float = field(init=False)  # initに含めない

    def __post_init__(self):
        self.area = self.width * self.height

r = Rectangle(3, 4)
print(r.area)  # 12
```

> [!NOTE]
> `frozen=True`のクラスでは`self.x = ...`のような属性への直接代入ができないため、`__post_init__`内で値を設定する場合は`object.__setattr__(self, "area", ...)`を使う必要がある。

### `dataclasses`モジュールのヘルパー関数
- **`fields(obj)`**: フィールド情報（`Field`オブジェクト）のタプルを取得
  ```python
  from dataclasses import dataclass, fields

  @dataclass
  class Point:
      x: int
      y: int

  for f in fields(Point):
      print(f.name, f.type)
  # x <class 'int'>
  # y <class 'int'>
  ```
- **`asdict(obj)`**: インスタンスを辞書に変換（ネストしたdataclassやコンテナも再帰的に変換される）
  ```python
  from dataclasses import asdict

  print(asdict(Point(1, 2)))  # {'x': 1, 'y': 2}
  ```
- **`astuple(obj)`**: インスタンスをタプルに変換（`asdict`同様、再帰的に変換される）
  ```python
  from dataclasses import astuple

  print(astuple(Point(1, 2)))  # (1, 2)
  ```
- **`replace(obj, **changes)`**: 指定したフィールドだけ変更した**新しいインスタンス**を作成する（`frozen=True`のクラスを"更新"する際によく使う。浅いコピーに相当）
  ```python
  from dataclasses import replace

  p1 = Point(1, 2)
  p2 = replace(p1, x=10)
  print(p1)  # Point(x=1, y=2)
  print(p2)  # Point(x=10, y=2)
  ```
- **`is_dataclass(obj)`**: dataclassかどうかを判定

### コピーと`deepcopy`
- 通常の代入(`p2 = p1`)は**同じオブジェクトを参照**するだけ(コピーではない)
- `copy.copy()`（浅いコピー）: 新しいインスタンスを作るが、**ネストしたミュータブルな属性は同じオブジェクトを参照したまま**
- `copy.deepcopy()`（深いコピー）: ネストした属性も再帰的に複製する。ミュータブルな属性（`list`, `dict`など）を含むdataclassを完全に独立させたい場合はこちらを使う

```python
import copy
from dataclasses import dataclass, field

@dataclass
class Team:
    members: list[str] = field(default_factory=list)

t1 = Team(members=["Alice"])

t2 = copy.copy(t1)       # 浅いコピー
t2.members.append("Bob")
print(t1.members)  # ['Alice', 'Bob'] ← t1にも影響してしまう！

t3 = copy.deepcopy(t1)   # 深いコピー
t3.members.append("Carol")
print(t1.members)  # ['Alice', 'Bob'] ← t1は影響を受けない
print(t3.members)  # ['Alice', 'Bob', 'Carol']
```

> [!NOTE]
> `dataclasses.replace()`は浅いコピーに相当する。指定しなかったフィールドは元のオブジェクトへの参照がそのままコピーされるため、ミュータブルなフィールドが絡む場合は`copy.deepcopy()`との使い分けが必要。

### 主なな機能とオプション
#### 1. 比較機能
- `__eq__` が自動で定義されるので、値比較が可能
```python
@dataclass
class Point:
    x: int
    y: int

p1 = Point(1, 2)
p2 = Point(1, 2)

print(p1 == p2)  # True
print(p1 is p2)  # False（別オブジェクト）
```
#### 1.1 `order=True`で大小比較（`<`, `<=`, `>`, `>=`）を有効化
- `order=True`を指定すると、フィールド定義順にタプル比較する形で`__lt__`, `__le__`, `__gt__`, `__ge__`が自動生成される（`__eq__`は`eq`の設定に従う）
```python
from dataclasses import dataclass

@dataclass(order=True)
class Point:
    x: int
    y: int

p1 = Point(1, 2)
p2 = Point(1, 3)
p3 = Point(2, 0)

print(p1 < p2)   # True  ← (1,2) < (1,3)
print(p1 < p3)   # True  ← (1,2) < (2,0) xが優先される
print(p2 <= p2)  # True
print(p3 > p1)   # True
```
- 比較対象が同じクラスのインスタンスでない場合は`NotImplemented`が返る（結果的に`TypeError`）
- `field(compare=False)`を指定したフィールドは比較（`__eq__`・順序比較の両方）から除外される
```python
from dataclasses import dataclass, field

@dataclass(order=True)
class Item:
    priority: int
    name: str = field(compare=False)  # 比較には使わない

print(Item(1, "a") < Item(2, "z"))  # True（nameは無視）
```

> [!WARNING]
> `eq=False`と`order=True`は同時指定できない（`eq`はデフォルト`True`なので通常は意識不要だが、明示的に`eq=False`にすると`ValueError: eq must be true if order is true`になる）。

> [!WARNING]
> `order=True`を指定した状態で`__lt__`などを自分で定義（オーバーライド）すると、`@dataclass`側の自動生成と衝突して **クラス定義時に`TypeError`** になる。
> ```python
> @dataclass(order=True)
> class Point:
>     x: int
>     y: int
>
>     def __lt__(self, other):
>         return self.x < other.x
> # ❌ TypeError: Cannot overwrite attribute __lt__ in class Point.
> #    Consider using functools.total_ordering
> ```
> 独自の比較ロジックを定義したい場合は`order=True`を外し（デフォルトの`order=False`のまま）、自分で`__lt__`等を実装するか、`functools.total_ordering`を使う。

#### 2. 不変（イミュータブル）なデータクラス
- `frozen=True` を指定すると **不変（immutable）** なクラスにできる
```python
@dataclass(frozen=True)
class Point:
    x: int
    y: int

p = Point(1, 2)
# p.x = 10  # ❌ dataclasses.FrozenInstanceError が発生
```

#### 2.1 `frozen=True`と`set`の要素（ハッシュ可能性）
- `set`の要素はハッシュ可能（hashable）である必要がある
- 通常の`@dataclass`はミュータブルなので`__hash__`が`None`になり、`set`の要素にはできない
- `frozen=True`にすると自動的にハッシュ可能になり、`set`の要素として使える

```python
from dataclasses import dataclass

@dataclass
class MutablePoint:
    x: int
    y: int

# {MutablePoint(1, 2)}  # ❌ TypeError: unhashable type

@dataclass(frozen=True)
class Point:
    x: int
    y: int

points = {Point(1, 2), Point(3, 4)}  # ✅ OK

points_list = [Point(1, 2), Point(3, 4), Point(1, 2)]  # Point(1, 2)が重複
points_set = set(points_list)
print(points_set)  # {Point(x=1, y=2), Point(x=3, y=4)} ← 重複が排除される
```

- 型ヒントとして`set[Point]`のように書くこともできる（[[集合（set）について.md]]参照）
```python
def get_unique_x_values(points: set[Point]) -> set[int]:
    return {p.x for p in points}

print(get_unique_x_values(points_set))  # {1, 3}
```

#### 3. `field()`関数で属性の細かい制御
- `@dataclass`のフィールドごとの挙動を細かく制御するための仕組み
- 設定可能なオプション
  - **`default_factory`**: リスト・辞書など「可変オブジェクト」をデフォルトにしたいときに使用  
    ```python
    from dataclasses import dataclass, field

    @dataclass
    class Team:
        members: list[str] = field(default_factory=list)

    t1 = Team()
    t2 = Team()
    t1.members.append("Alice")
    print(t1.members)  # ['Alice']
    print(t2.members)  # [] ← 共有されない！
    ```

> [!IMPORTANT]
> `default_factory`は、Pythonの「引数デフォルト値が共有される問題」を回避するために使用される。
> ### 引数デフォルト値が共有される問題とは
> - Python では、 **関数定義時にデフォルト引数が1回だけ評価され、関数が呼ばれるたびに再利用される** という仕様がある。つまり、リストや辞書などの可変オブジェクトをデフォルト引数にすると、全ての呼び出しで同じオブジェクトが共有される。
> - Pythonの変数はオブジェクトへの参照（メモリアドレス）を保持しており、可変オブジェクト（ミュータブル）は同じアドレス上で内容が変更されるため、意図しない共有が発生する。
> #### 問題のあるコード例
> ```python
> def add_item(item, lst=[]):  # ❌ デフォルト値にリストを直接書いてしまう
>     lst.append(item)
>     return lst
>
> print(add_item("A"))  # ['A']
> print(add_item("B"))  # ['A', 'B'] ← Aのリストが使い回されている！
> print(add_item("C"))  # ['A', 'B', 'C']
> ```
> - 本来なら`"B"`や`"C"`の呼び出しは`['B']`, `['C']`になることを期待するが、実際は同じリストが共有されてしまう
> - これは`@dataclass`を使っても同様の問題が発生する  
>   ```python
>   from dataclasses import dataclass
>
>   @dataclass
>   class Team:
>       members: list[str] = []  # ❌ 危険！
> 
>   t1 = Team()
>   t2 = Team()
>
>   t1.members.append("Alice")
>   print(t1.members)  # ['Alice']
>   print(t2.members)  # ['Alice'] ← t1と同じリストを参照してしまう！
>   ```
> 
> #### `default_factory`で解決
> - `field(default_factory=...)`を使うと、インスタンス生成ごとに新しいオブジェクトを作ってくれるので安全

  - **`init`**: デフォルトは`True`。`False` にすると、`__init__` メソッドの引数に含めない（初期化時に値を渡せなくなる）  
    ```python
    @dataclass
    class Example:
        x: int
        y: int = field(init=False, default=0)

    e = Example(10)
    print(e)  # Example(x=10, y=0)
    # e = Example(10, 5)  # ❌ yはinitに含まれない
    ```
  - **`repr`**: デフォルトは`True`。`False` にすると、`__repr__`の出力に含まれない  
    ```python
    @dataclass
    class Example:
        x: int
        secret: str = field(repr=False)

    e = Example(10, "hidden")
    print(e)  # Example(x=10)
    ```

### よくある組み合わせ例
- 辞書やリストのデフォルト値を設定したい場合  
  ```python
  @dataclass
  class Config:
      settings: dict = field(default_factory=dict)
      options: list = field(default_factory=list)
  ```
- DBの自動生成フィールド（idはユーザーに入力させない）  
  ```python
  @dataclass
  class User:
      id: int = field(init=False)  # DBが自動生成するのでinitに含めない
      name: str
      email: str
  ```
- パスワードなどの機密情報を`__repr__`に含めたくない場合  
  ```python
  @dataclass
  class User:
      id: int = field(init=False)  # DBが自動生成するのでinitに含めない
      name: str
      email: str
      password: str = field(repr=False)
  ```