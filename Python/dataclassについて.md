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