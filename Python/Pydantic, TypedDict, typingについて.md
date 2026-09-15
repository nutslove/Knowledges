# `typing`
## `typing`ライブラリの型ヒント
- `typing` ライブラリは（Python3.5以降に導入された）Pythonの標準ライブラリの一部で、型チェックをサポートするための機能を提供
- 普通の型ヒントにはない`Optional`や`List`などの複雑な型ヒントを使える
- **型の強制力はない**（実行時にruntimeによる型チェックなどは行われない）
  - ただ、`mypy`などのツールを使って、型の一致をチェックすることができる
### `mypy`
- install
  - `python3 -m pip install -U mypy`
- usage
  - `mypy sommpython.py`

## `typing`ライブラリで使える型ヒント（一部）
- `typing`ライブラリで使えるすべての型ヒントは[typingドキュメント](https://docs.python.org/ja/3/library/typing.html)から確認
1. `Optional`
    - 変数が指定された型の値か、`None`であることを示す。  
      たとえば、`Optional[int]`は、その変数がint型の値またはNoneのいずれかを持つことを意味する。  
      `Optional[X]`は`X | None` (や `Union[X, None]`) と同等  
      ```python
      from typing import Optional

      def greet(name: Optional[str]) -> str:
          if name is None:
              return "Hello, Guest!"
          return f"Hello, {name}!"
      ```
2. `Dict`
    - 任意のKeyとValueの型を持つ辞書。  
      たとえば、`Dict[str, int]`は、文字列をKeyとし、整数をValueとする辞書を意味する。  
      ```python
      from typing import Dict

      def get_value(data: Dict[str, int], key: str) -> int:
          return data[key]
      ```
      - python 3.9からは`from typing import Dict`も不要となり、`dict[srt, int]`("d"が小文字)のように使える
3. `List`
    - 任意の型の要素を持つリストを示す。  
      たとえば、`List[int]`は整数のリストを意味する。  
      ```python
      from typing import List

      def sum_numbers(numbers: List[int]) -> int:
          return sum(numbers)
      ```
    - python 3.9からは`from typing import List`も不要となり、`list[int]`("l"が小文字)のように使える
4. `Tuple`
    - 固定長の不変なタプルを表す。各要素の型を個別に指定できる。  
      例えば、以下は文字列、整数、浮動小数点数の3要素のタプルを表す。  
      ```python
      from typing import Tuple

      Tuple[str, int, float]
      ```
    - python 3.9からは`from typing import Tuple`も不要となり、`tuple[str, int, float]`("t"が小文字)のように使える
5. `Union`
    - 複数の型のいずれかであることを示す。  
      たとえば、`Union[int, str]`は、整数または文字列のいずれかを意味する。  
      `Union[X, Y]`は`X | Y `と等価で X または Y を表す。  
      ```python
      from typing import Union

      def to_string(value: Union[int, str]) -> str:
          return str(value)
      ```
6. `Literal`
    - 特定の値のみを許可する  
      ```python
      from typing import Literal

      def set_mode(mode: Literal["auto", "manual"]) -> None:
          print(f"Mode is set to: {mode}")

      set_mode("auto")   # OK
      set_mode("manual") # OK
      set_mode("other")  # エラー（mypyなど型チェックツールが警告を出す）

      def set_priority(priority: Literal[1, 2, 3]) -> None:
          print(f"Priority set to: {priority}")

      set_priority(1)  # OK
      set_priority(4)  # エラー

      def enable_feature(flag: Literal[True, False]) -> None:
          print(f"Feature enabled: {flag}")

      enable_feature(True)  # OK
      enable_feature(False) # OK
      enable_feature(1)     # エラー（True/False しか受け付けない）
      ```
    - `Union`と組み合わせて使うことも可能  
      ```python
      def configure(option: Union[Literal["low", "medium", "high"], int]) -> None:
          print(f"Configuration set to: {option}")

      configure("low")  # OK
      configure(10)     # OK
      configure("other") # エラー
      ```
7. `Annotated`
   - 型情報に追加のメタデータを付与する
   - 基本の型情報を保持しながら、追加情報を付与することができる
   - 基本構文  
     ```python
     from typing import Annotated

     T = Annotated[型, メタデータ]
     ```
     - `型`: 基本となる型（例えば `int`, `str`, `list[str]` など）
     - `メタデータ`: 追加の情報を指定（例えば `max_length=10` などの制約）
   - **単体で使う時はあくまで型ヒントで制約の強制はできないけど、Pydanticと組み合わせて使うことで制約の強制ができる**  
     ```python
     from typing import Annotated
     from pydantic import BaseModel, Field

     class User(BaseModel):
         name: Annotated[str, Field(max_length=10)]
         age: Annotated[int, Field(ge=18)]  # 18歳以上であることを保証

     user = User(name="Alice", age=20)  # OK
     user = User(name="A very long name", age=15)  # 例外発生
     ```
8. `Self`
    - Python 3.11 以降に追加された型ヒントであり、「そのクラス自身のインスタンス」を示す型  
    ```python
    from typing import Self

    class MyClass:
        def __init__(self, value: int):
            self.value = value

        def validate(self) -> Self:
            print("Validation complete")
            return self

    # インスタンス作成
    obj = MyClass(10)

    # メソッドチェーンが可能になる
    new_obj = obj.validate()

    print(obj is new_obj)  # True （同じインスタンスを返している）
    ```
    - `self`を返すことで、メソッドチェーンを実現することができる  
      ```python
      obj.validate().validate().validate()
      ```
9. `NewType`
    - 新しい型を定義するための関数
    - 基本構文  
      ```python
      from typing import NewType

      NewTypeName = NewType('NewTypeName', ExistingType)
      ```
      - `NewTypeName`: 新しく定義する型の名前
      - `ExistingType`: 既存の型（例えば `int`, `str`, `list` など）
    - 例  
      ```python
      from typing import NewType

      UserId = NewType('UserId', int)
      ProductId = NewType('ProductId', int)

      def get_user(user_id: UserId):
          print(f"Getting user with ID: {user_id}")

      def get_product(product_id: ProductId):
          print(f"Getting product with ID: {product_id}")

      user_id = UserId(123)
      product_id = ProductId(456)

      get_user(user_id)        # OK
      get_product(product_id)  # OK

      get_user(product_id)     # 型チェックツールでエラーになる（実行時は問題ない）
      ```
10. `ClassVar`
    - その属性が**インスタンス変数ではなくクラス変数**であることを示す型ヒント（クラス変数・インスタンス変数の違いは[[Classについて]]を参照）
    - 実行時の強制力はなく、あくまで型チェッカー（mypyなど）や`dataclass`・`pydantic`に対して「この属性は個々のインスタンスが持つものではない」と伝えるためのもの
    - 例（通常のクラス）  
      ```python
      from typing import ClassVar

      class Counter:
          count: ClassVar[int] = 0  # 全インスタンスで共有されるクラス変数
          name: str                 # インスタンス変数

          def __init__(self, name: str):
              self.name = name
              Counter.count += 1

      c1 = Counter("a")
      c2 = Counter("b")
      print(Counter.count)  # 2
      ```
    - `dataclass`との組み合わせが代表的な用途  
      ```python
      from dataclasses import dataclass
      from typing import ClassVar

      @dataclass
      class Config:
          max_size: ClassVar[int] = 100  # __init__の引数にならない（dataclassのフィールドとして扱われない）
          name: str

      c = Config(name="app1")
      print(Config.max_size)  # 100
      ```
      - `ClassVar`を付けないと`dataclass`はその属性を通常のインスタンスフィールドとみなし、`__init__`の引数に含めてしまう
11. `TypeVar`
    - **ジェネリック**（総称型）な関数やクラスで使う「型変数」を定義するための関数
    - 呼び出し時に渡された実際の型に応じて、戻り値などの型が決まる
    - 基本構文  
      ```python
      from typing import TypeVar

      T = TypeVar("T")  # "T"は慣例的な名前。変数名と文字列は一致させるのが慣例（他の名前でも動作は同じ）

      def first(items: list[T]) -> T:
          return items[0]

      first([1, 2, 3])   # T は int と推論される
      first(["a", "b"])  # T は str と推論される
      ```
    - `bound`で「この型（のサブクラス）に限定する」という制約を付けられる  
      ```python
      from typing import TypeVar

      Number = TypeVar("Number", bound=int | float)

      def double(x: Number) -> Number:
          return x * 2

      double(3)     # OK（int）
      double(3.5)   # OK（float）
      double("a")   # 型チェッカーがエラーを検出
      ```
    - 候補となる型を列挙して制約することも可能（`bound`との違いはサブクラスを許容しない点）  
      ```python
      StrOrInt = TypeVar("StrOrInt", str, int)

      def show(value: StrOrInt) -> None:
          print(value)
      ```
    > [!NOTE]
    > Python 3.12以降は`def first[T](items: list[T]) -> T:`のように、`TypeVar`を明示的に定義せず書ける新しいジェネリック構文（PEP 695）も使える
    >
    > `first[T]`の`[T]`は引数でも戻り値でもなく、**その関数（やクラス）専用の型変数`T`をここで宣言している**部分。旧構文でいう以下の2行の役割をまとめて担っている。
    > ```python
    > # 旧構文（Python 3.11以前も使える）
    > T = TypeVar("T")                 # ← ①ここで型変数Tを定義
    > def first(items: list[T]) -> T:  # ← ②定義したTを使う
    >     return items[0]
    >
    > # 新構文（Python 3.12以降）
    > def first[T](items: list[T]) -> T:  # ← [T]で定義と使用が一体化
    >     return items[0]
    > ```
    > - `T`のスコープはその関数（やクラス）の中だけに閉じるため、モジュールレベルで`T = TypeVar("T")`を使い回すよりも名前の衝突を避けやすい
    > - クラスの場合も同様で、旧構文の`class Stack(Generic[T]):`は新構文で`class Stack[T]:`と書ける
12. `Generic`
    - `TypeVar`を使って**クラス自体をジェネリック化**するための基底クラス
    - `Generic[T]`を継承することで、インスタンス生成時に`Stack[int]`のように型パラメータを指定できるようになる
    - **なぜ`Generic`を使うのか**: 「型の対応関係を保ったまま、汎用的な（どんな型にも使える）クラスを書けるようにする」ため
      - `Generic`を使わずに汎用クラスを書こうとすると、`Any`に頼らざるを得ない  
        ```python
        from typing import Any

        class Stack:
            def __init__(self) -> None:
                self._items: list[Any] = []

            def push(self, item: Any) -> None:
                self._items.append(item)

            def pop(self) -> Any:
                return self._items.pop()

        s = Stack()
        s.push(1)
        s.push("a")     # int用のつもりが文字列も入ってしまう…型チェッカーは検出できない
        value = s.pop()
        value.upper()   # valueの実際の型はintかもしれないが、Anyなので型チェッカーはエラーを出さない
        ```
        - `Any`は「どんな型でもOK」という意味なので、型チェッカー（mypyなど）がそもそもチェックを放棄してしまい、バグを実行時まで発見できない
      - 型安全にしたいなら`IntStack`, `StrStack`のように型ごとにクラスを複製する手もあるが、中身は同じロジックを何度も書くことになり`DRY`原則に反する
      - `Generic`を使うとこの両方が解決する（下記の基本構文の例を参照）
        - クラスは1つだけ書けばよく、`Stack[int]`、`Stack[str]`のようにどんな型でも使い回せる（コードの重複がない）
        - それでいて、インスタンスごとに中身の型を型チェッカーに伝えられる（`Stack[int]`に文字列を`push`しようとするとエラーになる）
        - IDEの補完も効くようになる（`int_stack.pop()`の戻り値がちゃんと`int`として認識される）
      - 身近な例として、普段使っている`list[int]`や`dict[str, int]`も、Python標準の`list`や`dict`が内部的にジェネリック対応しているからこそ書ける。`Generic`を継承すれば自分が定義したクラスにも同じ仕組みを持たせられる（例: `Repository[User]`、`Response[T]`のような独自の汎用ラッパークラス）
    - 基本構文  
      ```python
      from typing import Generic, TypeVar

      T = TypeVar("T")

      class Stack(Generic[T]):
          def __init__(self) -> None:
              self._items: list[T] = []

          def push(self, item: T) -> None:
              self._items.append(item)

          def pop(self) -> T:
              return self._items.pop()

      int_stack: Stack[int] = Stack()
      int_stack.push(1)
      int_stack.push("a")  # 型チェッカーがエラーを検出（int_stackはStack[int]のため）
      ```
    - 型パラメータを複数持たせることも可能  
      ```python
      K = TypeVar("K")
      V = TypeVar("V")

      class Pair(Generic[K, V]):
          def __init__(self, key: K, value: V) -> None:
              self.key = key
              self.value = value

      p = Pair[str, int]("age", 30)
      ```
    - Python 3.12以降は`class Stack[T]:`のように`Generic`の明示的な継承なしでジェネリッククラスを書ける新しい構文（PEP 695）も使える

# `Enum`
- Python標準ライブラリ`enum`モジュールで提供される、**列挙型**を定義するためのクラス
- `typing`の`Literal`と似た用途で使えるが、**ランタイムで強制力を持つ**点が大きく異なる
  - `Literal`は型チェッカー（mypyなど）を通さないと不正値を検出できない
  - `Enum`は実行時に未定義のメンバーへアクセスすると例外が発生する
- 基本構文
```python
  from enum import Enum

  class Status(Enum):
      PENDING = "pending"
      RUNNING = "running"
      DONE = "done"

  s = Status.RUNNING
  print(s)        # Status.RUNNING
  print(s.name)   # "RUNNING"
  print(s.value)  # "running"
```
- **存在しないメンバーへのアクセスはランタイムで弾かれる**
```python
  Status.UNKNOWN       # AttributeError
  Status("invalid")    # ValueError: 'invalid' is not a valid Status
  Status["UNKNOWN"]    # KeyError
```
- メンバーのイテレートが可能
```python
  for s in Status:
      print(s.name, s.value)
```
- `str`や`int`を継承させると、生の値としても扱える
```python
  class Status(str, Enum):
      PENDING = "pending"
      RUNNING = "running"

  Status.RUNNING == "running"  # True
```
- `IntEnum`、`StrEnum`（Python 3.11以降）など派生クラスもある

## `Enum` と `Literal` の使い分け
| 観点 | `Enum` | `Literal` |
|---|---|---|
| ランタイムでの強制力 | あり（例外発生） | なし（型チェッカーのみ） |
| 値の実体 | 専用のオブジェクト | 生の文字列/整数など |
| メンバーのイテレート | 可能 | 不可 |
| メソッドの追加 | 可能 | 不可 |
| 軽量さ | クラス定義が必要 | 型エイリアスのみで済む |

- **外部入力（APIレスポンス、webhookペイロード、環境変数など）を扱う境界**では、`Enum`にして`Status(payload["status"])`のように変換時点で弾くほうが安全
- **内部的な選択肢を型ヒントで示すだけ**であれば、`Literal`のほうが軽量で十分
- なお、`pydantic`の`BaseModel`のフィールドに使う場合は、`Enum`でも`Literal`でもどちらもランタイム検証されるので安全

# `TypedDict`
- Python 3.8で公式に`typing`モジュールに追加されたので
  - Python 3.8以降の場合は`from typing import TypedDict`
  - Python 3.7までは`pip install typing_extensions`でインストール後、`from typing_extensions import TypedDict`
- `typing`の`Dict`の中で`Literal`と`Union`を組み合わせて使う場合、KeyとValueの順序まではチェックされない
  - 例  
    ```python
    from typing import Dict, Literal, Union

    Animal = Dict[Literal["name", "age"], Union[str, int]]

    def show_animal(a: Animal):
        print(a["name"])
        print(a["age"])

    show_animal({"name": "taro", "age": 5}) # OK
    show_animal({"name": "taro", "color": "black"}) # NG
    show_animal({"name": 5, "age": "taro"}) # OK ★→本当はNGになってほしいところ
    ```
  - `TypedDict`を使えば厳格な型チェックが可能
    ```python
    from typing import TypedDict

    class Animal(TypedDict):
        name: str
        age: int

    def show_animal(a: Animal):
        print(a["name"])
        print(a["age"])

    show_animal({"name": "taro", "age": 5}) # OK
    show_animal({"name": "taro", "color": "black"}) # NG
    show_animal({"name": 5, "age": "taro"}) # NG
    ```

# `pydantic`
## `pydantic`の`BaseModel`を使った型の強制
- **`pydantic`モジュール`BaseModel`を使って定義した型ヒントはある程度強制力を持つ**
- `pydantic`モジュールの型ヒントの特徴
  - **データ検証**： Pydanticは定義された型ヒントに基づいてデータを検証する。不適切な型のデータが渡されると、ValidationErrorを発生させる。
  - **型変換**: 可能な場合、Pydanticは入力データを指定された型に変換しようとする。
  - **実行時チェック**： Pydanticのモデルインスタンスを作成する際に型チェックが行われる。
- **ただ、この強制力はクラスのインスタンス化時にのみ適用される**
- 例  
  ```python
  from pydantic import BaseModel

  class User(BaseModel):
      name: str
      age: int

  # これは正常に動作します
  user1 = User(name="Alice", age=30)

  # これは age が文字列なので、整数に変換されます
  user2 = User(name="Bob", age="25")

  # これはエラーになります（age に文字列を変換できない）
  try:
      user3 = User(name="Charlie", age="twenty")
  except ValueError as e:
      print(f"エラー: {e}")

  # これもエラーになります（必須フィールドの name が欠けている）
  try:
      user4 = User(age=35)
  except ValueError as e:
      print(f"エラー: {e}")
  ```

## 強制力が適用される場合
1. モデルのインスタンス化時：
   ```python
   from pydantic import BaseModel

   class User(BaseModel):
       name: str
       age: int

   # この時点で型チェックと検証が行われる
   user = User(name="Alice", age=30)
   ```

2. モデルの `parse_obj` メソッドを使用する時：
   ```python
   data = {"name": "Bob", "age": 25}
   user = User.parse_obj(data)
   ```

3. モデルの `dict()` メソッドを使ってデータを取り出す時：
   ```python
   user_dict = user.dict()
   ```

## 強制力が適用されない場合
1. インスタンス化後の属性の直接変更：
   ```python
   user = User(name="Alice", age=30)
   user.age = "Not an integer"  # これは型チェックされない
   ```

2. `__dict__` を通じた直接アクセス：
   ```python
   user.__dict__['age'] = "Not an integer"  # これも型チェックされない
   ```

3. クラス定義時：
   ```python
   class User(BaseModel):
       name: str
       age: int = "Not an integer"  # この時点ではエラーにならない
   ```
   エラーはこのクラスのインスタンスを作成しようとした時に発生する。

4. 継承したサブクラスでの属性の上書き：
   ```python
   class AdminUser(User):
       age: str  # 型が変更されても、この時点ではエラーにならない
   ```

5. `setattr()` 関数を使用した場合：
   ```python
   setattr(user, 'age', "Not an integer")  # これも型チェックされない
   ```

6. モデルのメソッド内での属性変更：
   ```python
   class User(BaseModel):
       name: str
       age: int

       def update_age(self, new_age):
           self.age = new_age  # メソッド内での変更は型チェックされない

   user = User(name="Alice", age=30)
   user.update_age("Not an integer")  # これはエラーにならない
   ```

## その他`pydantic`の機能
### `Field`
- モデルのフィールド（属性）に対してメタデータや制約（バリデーション条件）を付与するための関数
- **`BaseModel`の属性と組み合わせて使うことで、データ構造を明示的に定義し、入力データの自動検証やドキュメント生成などが可能になる**
- 基本構文  
  ```python
  from pydantic import BaseModel, Field

  class User(BaseModel):
      id: int = Field(..., description="ユーザーID", ge=1)
      name: str = Field(..., min_length=1, max_length=50, description="ユーザー名")
      age: int = Field(default=18, ge=0, le=150, description="年齢")
  ```
  - **`...`は必須フィールド**を示す
    - 値がないと`ValidationError`になる
  - `description`: フィールドの説明
  - `ge`, `le`: 数値の範囲制約（greater than or equal, less than or equal）
  - `min_length`, `max_length`: 文字列やリストの長さ制約
  - `default`: デフォルト値
  - `default_factory`: デフォルト値を生成する関数を指定（動的デフォルト値）  
    ```python
    from datetime import datetime

    class Log(BaseModel):
        timestamp: datetime = Field(default_factory=datetime.utcnow)
    ``` 
  - `alias`: 別名を指定（Key名を変更する）  
    ```python
    class Item(BaseModel):
        item_name: str = Field(..., alias="name")

    data = {"name": "Book"}
    item = Item(**data)
    print(item.item_name)  # Book
    ```
- バリデーション例  
  ```python
  class Product(BaseModel):
      price: float = Field(..., gt=0, description="0より大きい価格")
      tags: list[str] = Field(default_factory=list, max_length=5)

  Product(price=-10)
  # ValidationError: 1 validation error for Product
  # price
  #   ensure this value is greater than 0 (type=value_error.number.not_gt; limit_value=0)
  ```

### `discriminator`（判別共用体 / Tagged Union）
- `Union`型のフィールドに対して、**どのメンバー型でバリデーションすべきかを判定するための「目印」となるフィールド名**を`Field(discriminator=...)`で指定する仕組み
  - Pydantic v2で本格導入(v1にも簡易版はあったが、v2でより柔軟・高性能になった)
  - 英語では**Discriminated Union**（判別共用体）や**Tagged Union**（タグ付き共用体）と呼ばれる
- **通常の`Union`の問題点**
  - discriminatorを指定しない場合、Pydanticは **デフォルトの"smart mode"** でUnion内の各メンバー型を検証し、「最も良くマッチした型」を採用する（[公式ドキュメント](https://docs.pydantic.dev/latest/concepts/unions/#smart-mode)）
    - `BaseModel`などでは、**入力データから実際に値がセットされたフィールド数が多い方**が優先される（ネストしたモデルのフィールド数も考慮される）
    - フィールド数が同点の場合は、型変換の少なさ（**exactness**: 完全一致 > strictモードで通る > laxモードで通る）が僅差の判定基準として使われる
    - （`BaseModel`などではない）単純な型同士の場合は、**型が完全一致するメンバーが見つかった時点でそれが即採用**され、以降のメンバーは試されない
  - このロジックは「両方とも検証には成功するが、どちらが本来の意図か」を**フィールド名の値ではなく構造的なスコアだけ**で決めてしまうため、型の構造が似ていると、意図しない型にマッチしてしまったり、エラーメッセージが「どの型にも合わなかった」という分かりにくいものになったりする
    ```python
    from pydantic import BaseModel

    class Cat(BaseModel):
        pet_type: str
        meows: int

    class Dog(BaseModel):
        pet_type: str
        barks: float

    class Owner(BaseModel):
        pet: Cat | Dog  # discriminatorなし

    # 構造が似ていると、総当たりの結果どちらにもマッチしうるため
    # 意図した型と違う型として解釈されたり、失敗時のエラーが分かりにくくなる
    ```
  - **具体例①: 必須項目が抜けているのに、エラーにならず意図と違う型として解釈されてしまう**
    - `meows`にデフォルト値を持たせておくと、本来`Dog`のつもりで`barks`（必須）を書き忘れたデータを渡しても、`ValidationError`にならず**黙って`Cat`として解釈されてしまう**
      ```python
      from pydantic import BaseModel
      from typing import Union

      class Cat(BaseModel):
          pet_type: str
          meows: int = 0  # デフォルト値があるので必須ではない

      class Dog(BaseModel):
          pet_type: str
          barks: float  # 必須

      class Owner(BaseModel):
          pet: Union[Cat, Dog]  # discriminatorなし

      # barks を書き忘れた（本来は Dog のつもり）
      owner = Owner(pet={"pet_type": "dog"})
      print(owner.pet)
      # pet_type='dog' meows=0  ← barks不足のエラーにならず、Catとして解釈されてしまう！
      print(type(owner.pet))  # <class '__main__.Cat'>
      ```
      - `Dog`は`barks`が必須のため検証に失敗するが、`Cat`は`meows`にデフォルト値があるため`pet_type`だけで検証に成功してしまう
      - smart modeはUnionメンバーの中から**「検証に成功した中で最もスコアの良いもの」**を選ぶ仕組みだが、そもそも`Dog`は検証自体に失敗して候補から外れてしまうため、`pet_type="dog"`という値と矛盾したまま`Cat`が採用され、`barks`不足に気づかれない
      - `discriminator="pet_type"`を指定していれば、`pet_type`の値だけを見て検証対象が`Dog`一択に絞られるため、`Dog`の`barks`不足がきちんと`ValidationError`として検出される
  - **具体例②: 両方失敗した場合、エラーメッセージが分かりにくい**
    - 一方、（具体例①とは異なり）`meows`にデフォルト値がなく`Cat`・`Dog`どちらも検証に失敗するケースでは、両方の検証エラーがまとめて出てしまい、どちらが「本命」だったのか分かりにくいエラーになる
      ```python
      from pydantic import BaseModel
      from typing import Union

      class Cat(BaseModel):
          pet_type: str
          meows: int  # 必須（デフォルトなし）

      class Dog(BaseModel):
          pet_type: str
          barks: float  # 必須

      class Owner(BaseModel):
          pet: Union[Cat, Dog]  # discriminatorなし

      Owner(pet={"pet_type": "dog"})  # barks（必須）が抜けている
      # ValidationError: 2 validation errors for Owner
      # pet.Cat.meows
      #   Field required [type=missing, ...]
      # pet.Dog.barks
      #   Field required [type=missing, ...]
      # → CatとDogの両方の検証エラーがまとめて出てしまい、
      #   「本当はDogのbarksが足りない」ということが一見して分かりにくい
      ```
    - `discriminator`を使えば、`pet_type="dog"`の時点で検証対象は`Dog`だけに絞られるため、`pet.dog.barks: Field required`という一意で分かりやすいエラーになる（実際に上記②のクラスに`discriminator="pet_type"`を付けて実行すると、`Dog`の`barks`不足だけを指すエラー1件になる）
- **`discriminator`を使うと、共通のLiteralフィールドの値を見て一意に型を決定できる**
  - 各メンバー型に**共通のフィールド名**を持たせ、その値を`Literal`で固定する
    ```python
    from typing import Literal, Union
    from pydantic import BaseModel, Field

    class Cat(BaseModel):
        pet_type: Literal["cat"]
        meows: int

    class Dog(BaseModel):
        pet_type: Literal["dog"]
        barks: float

    class Owner(BaseModel):
        pet: Union[Cat, Dog] = Field(discriminator="pet_type")
        # Python 3.10以降なら Union[Cat, Dog] は Cat | Dog と書ける

    Owner(pet={"pet_type": "cat", "meows": 3})    # Catとして解釈される
    Owner(pet={"pet_type": "dog", "barks": 2.5})  # Dogとして解釈される

    Owner(pet={"pet_type": "bird", "meows": 3})
    # ValidationError: pet_type が "cat" でも "dog" でもないため即座にエラー
    ```
- **`Annotated`と組み合わせて書くのが一般的**
  ```python
  from typing import Annotated, Literal, Union
  from pydantic import BaseModel, Field

  class Owner(BaseModel):
      pet: Annotated[Union[Cat, Dog], Field(discriminator="pet_type")]
      # Python 3.10以降なら Union[Cat, Dog] は Cat | Dog と書ける
  ```
- **メリット**
  1. **バリデーションが高速**: 各型を総当たりで試す必要がなく、discriminatorフィールドの値を見て対象の型だけを検証すればよい
  2. **エラーメッセージが分かりやすい**: 「pet_typeが不正な値」という具体的なエラーになる（通常のUnionだと「どの型にもマッチしなかった」という曖昧なエラーになりがち）
  3. **JSON Schemaにも反映される**: OpenAPIなどのドキュメントで`discriminator`付きの`oneOf`として表現され、フロントエンドのコード生成ツールなどとも相性が良い
- **discriminatorフィールドの制約**
  - 各メンバー型で**同じフィールド名**を使う必要がある
  - そのフィールドの型は **`Literal`（単一の値、またはLiteralのUnion）** である必要がある（`str`のような広い型は不可）
  - **`str`を継承した`Enum`を使うことも可能**
    ```python
    from enum import Enum
    from typing import Literal

    class PetType(str, Enum):
        CAT = "cat"
        DOG = "dog"

    class Cat(BaseModel):
        pet_type: Literal[PetType.CAT]
        meows: int
    ```
- **ネスト（入れ子）した判別共用体もサポート**
  - **1つのフィールドに設定できるdiscriminatorは1つだけ**だが、`Union`の中にさらに別のdiscriminatorを持つ`Union`を含めることで、複数のdiscriminatorを組み合わせられる（例: `pet_type`で種類を判別 → `color`でさらにサブ種類を判別）
    ```python
    from typing import Annotated, Literal, Union
    from pydantic import BaseModel, Field

    class BlackCat(BaseModel):
        pet_type: Literal["cat"]
        color: Literal["black"]
        black_name: str

    class WhiteCat(BaseModel):
        pet_type: Literal["cat"]
        color: Literal["white"]
        white_name: str

    Cat = Annotated[Union[BlackCat, WhiteCat], Field(discriminator="color")]

    class Dog(BaseModel):
        pet_type: Literal["dog"]
        name: str

    Pet = Annotated[Union[Cat, Dog], Field(discriminator="pet_type")]

    class Owner(BaseModel):
        pet: Pet

    Owner(pet={"pet_type": "cat", "color": "black", "black_name": "tama"})
    # pet=BlackCat(pet_type='cat', color='black', black_name='tama')
    Owner(pet={"pet_type": "dog", "name": "pochi"})
    # pet=Dog(pet_type='dog', name='pochi')
    ```
- **callable版：`Discriminator`（Pydantic v2で追加）**
  - 単純な1フィールドの値だけでは判別できない場合（フィールドの有無や複数フィールドの組み合わせで判別したいなど）に、**判別ロジックを関数として書ける**
    ```python
    from typing import Annotated, Union
    from pydantic import BaseModel, Discriminator, Tag

    class Cat(BaseModel):
        meows: int

    class Dog(BaseModel):
        barks: float

    def get_discriminator_value(v) -> str:
        # dictでもモデルインスタンスでも対応できるように書く
        if isinstance(v, dict):
            return "cat" if "meows" in v else "dog"
        return "cat" if hasattr(v, "meows") else "dog"

    class Owner(BaseModel):
        pet: Annotated[
            Union[Annotated[Cat, Tag("cat")], Annotated[Dog, Tag("dog")]],
            Discriminator(get_discriminator_value),
        ]
    ```
    - 各メンバー型に`Tag(...)`でタグ名を付け、`Discriminator`にそのタグ名を返す関数を渡す
    - **`get_discriminator_value(v)`の`v`には、型が確定する前の「生の入力データ」がそのまま渡ってくる**（渡し方によって型が変わる）
      - `dict`を渡した場合（`Owner(pet={"meows": 3})`や、JSON文字列からの`model_validate_json`経由の場合も含む）→ `v`は **`dict`**
      - 既にモデルインスタンスを渡した場合（`Owner(pet=Cat(meows=5))`）→ `v`は**そのモデルインスタンス自身**
      - どちらのパターンで呼ばれるか分からないため、`isinstance(v, dict)`と`hasattr(v, ...)`の両方に対応できるように書く必要がある
- **通常の`Union`との比較まとめ**

| 観点 | 通常の`Union`（discriminatorなし） | `discriminator`付き`Union` |
|---|---|---|
| 型の決定方法 | 総当たり（smart mode） | 目印フィールドの値で一意に決定 |
| バリデーション速度 | 型の数に応じて遅くなりうる | 高速 |
| エラーメッセージ | 曖昧（どれにも合わなかった） | 明確（目印フィールドの値が不正、など） |
| 各メンバー型の要件 | 特になし | 共通の`Literal`（または`Discriminator`で判別可能な）フィールドが必要 |

### `field_validator`
- **`Field`の宣言的な制約（`ge`、`max_length`など）では表現しきれない、カスタムなバリデーションや変換ロジックを書くためのデコレータ**
  - Pydantic v2で導入（v1の`@validator`の後継）。v1を使っている場合は`@validator`を使う
- 基本構文  
  ```python
  from pydantic import BaseModel, field_validator

  class User(BaseModel):
      name: str
      age: int

      @field_validator("name")
      @classmethod
      def name_must_not_be_empty(cls, v: str) -> str:
          if not v.strip():
              raise ValueError("name は空にできません")
          return v  # ★検証後の値を必ず return する（return した値がフィールドにセットされる）
  ```
  - **必ず`@classmethod`と一緒に使う**（第一引数は`self`ではなく`cls`）
  - **検証した値を`return`する必要がある**。`return`した値がそのままフィールドに格納されるため、変換（正規化）も同時に行える  
    ```python
    @field_validator("name")
    @classmethod
    def normalize_name(cls, v: str) -> str:
        return v.strip().title()  # 前後の空白を除去し、先頭を大文字に変換してセット
    ```
  - **エラーは`ValueError`（または`AssertionError`）を`raise`する**と、Pydanticがそれを補足して`ValidationError`に変換してくれる

- **`mode`引数による実行タイミングの制御**
  - ここでいう「実行される」の主語は **`@field_validator`が付いたバリデータ関数自身** （例では`check_price`）。Pydanticはまず自身の型変換・型チェックを行い、そのタイミングの**前**か**後**のどちらでバリデータ関数を呼び出すかを`mode`で制御する
  - `mode="after"`（**デフォルト**）: Pydanticの**型変換・型チェックが終わった後**に（バリデータ関数が）実行される。引数`v`は型変換済みの値  
    ```python
    class Product(BaseModel):
        price: int

        @field_validator("price", mode="after")
        @classmethod
        def check_price(cls, v: int) -> int:
            # ここに来る時点で v は int に変換済み
            if v <= 0:
                raise ValueError("price は正の数である必要があります")
            return v
    ```
  - `mode="before"`: Pydanticの**型変換が行われる前**に実行される。引数`v`はユーザーが渡した**生の値**（型は不定）。型変換前に値を整形したい場合に使う  
    ```python
    class Item(BaseModel):
        tags: list[str]

        @field_validator("tags", mode="before")
        @classmethod
        def split_str_to_list(cls, v):
            # "a,b,c" のような文字列を受け取ってリストに変換してから型チェックさせる
            if isinstance(v, str):
                return v.split(",")
            return v

    Item(tags="a,b,c")  # tags=['a', 'b', 'c']
    ```
  - `mode="wrap"`: **型変換・検証処理そのものを包み込む**モード。第二引数に`handler`（次の検証ステップを実行する関数）を受け取り、**自分で`handler(v)`を呼び出すタイミングを制御できる**。`before`と`after`を1つの関数で両方行いたい場合や、内部の検証エラーを捕捉して代替値にフォールバックしたい場合に使う  
    ```python
    from pydantic import BaseModel, field_validator
    from pydantic_core.core_schema import ValidatorFunctionWrapHandler

    class Item(BaseModel):
        price: int

        @field_validator("price", mode="wrap")
        @classmethod
        def lenient_price(cls, v, handler: ValidatorFunctionWrapHandler) -> int:
            try:
                return handler(v)  # 通常の型変換・検証を実行
            except Exception:
                return 0  # 変換に失敗したら0にフォールバック

    Item(price="abc")  # price=0（本来ならValidationErrorになるところをフォールバック）
    ```
  - `mode="plain"`: **`handler`を受け取らず、後続の型変換・検証を一切行わない**モード。バリデータの戻り値がそのままフィールドの値として確定する。Pydantic標準の型変換ロジックを完全にバイパスして独自ロジックだけで値を決定したい場合に使う  
    ```python
    class Item(BaseModel):
        price: int

        @field_validator("price", mode="plain")
        @classmethod
        def custom_only(cls, v) -> int:
            # ここで自前の変換・検証を完結させる（以降 Pydantic の型チェックは走らない）
            return int(str(v).replace(",", ""))

    Item(price="1,000")  # price=1000
    ```

  | `mode` | 実行タイミング | 引数`v`の型 | `handler` | 主な用途 |
  |---|---|---|---|---|
  | `"before"` | 型変換の**前** | 生の入力値（不定） | なし | 型変換前の前処理・整形 |
  | `"after"`（デフォルト） | 型変換の**後** | 変換済みの型 | なし | 変換後の値に対する検証 |
  | `"wrap"` | 型変換・検証を**包み込む** | 生の入力値（不定） | あり（呼び出しタイミングを制御） | 前後処理の統合、例外のフォールバック |
  | `"plain"` | 型変換・検証を**置き換える** | 生の入力値（不定） | なし | Pydantic標準の検証を完全にバイパス |

- **複数フィールドへの適用**
  - デコレータに複数のフィールド名を渡すと、同じバリデータをまとめて適用できる  
    ```python
    class Account(BaseModel):
        username: str
        password: str

        @field_validator("username", "password")
        @classmethod
        def not_blank(cls, v: str) -> str:
            if not v.strip():
                raise ValueError("空にできません")
            return v
    ```
  - 全フィールドに適用したい場合は`"*"`を指定できる  
    ```python
    @field_validator("*")
    @classmethod
    def no_none(cls, v):
        ...
    ```

- **他フィールドの値を参照する（`ValidationInfo`）**
  - 第二引数に`ValidationInfo`を受け取ると、`info.data`から**それまでに検証済みのフィールド**の値を参照できる  
    ```python
    from pydantic import BaseModel, field_validator, ValidationInfo

    class Event(BaseModel):
        start: int
        end: int

        @field_validator("end")
        @classmethod
        def end_after_start(cls, v: int, info: ValidationInfo) -> int:
            # info.data には end より前に定義・検証済みのフィールド（start）が入っている
            if "start" in info.data and v <= info.data["start"]:
                raise ValueError("end は start より大きい必要があります")
            return v
    ```
  - **注意**: `info.data`には**自分より前に定義され、かつ検証に成功したフィールドだけ**が入る。後に定義されたフィールドは参照できない → 複数フィールドにまたがる検証は次の`model_validator`の方が適している

### `field_validator` と `model_validator` の使い分け
- `model_validator`は**モデル全体（複数フィールドの組み合わせ）**を検証するためのデコレータ
- フィールドの定義順に依存せず、全フィールドが揃った状態で相互関係をチェックできる  
  ```python
  from pydantic import BaseModel, model_validator

  class Event(BaseModel):
      start: int
      end: int

      @model_validator(mode="after")
      def check_range(self):
          # mode="after" では self（モデルインスタンス）が渡され、全フィールドにアクセスできる
          if self.end <= self.start:
              raise ValueError("end は start より大きい必要があります")
          return self
  ```

| 観点 | `field_validator` | `model_validator` |
|---|---|---|
| 検証対象 | 個別のフィールド | モデル全体（複数フィールドの関係） |
| 受け取る値 | そのフィールドの値`v` | `mode="after"`では`self`、`mode="before"`では生のdict |
| 他フィールド参照 | `info.data`で**定義済みのものだけ**参照可 | すべてのフィールドを参照可 |
| 主な用途 | 単一フィールドの検証・正規化 | 「AとBのどちらか必須」など項目間の整合性チェック |

- **単一フィールドの値そのものを検証・変換**したい → `field_validator`
- **複数フィールドの組み合わせ（相関）を検証**したい → `model_validator`

### `computed_field`
- **他のフィールドの値から計算される「派生値」を、モデルのフィールドであるかのように`model_dump()`や`model_dump_json()`の出力・JSON Schemaに含めるためのデコレータ**
  - Pydantic v2で導入
  - **`@property`（または`@cached_property`）と組み合わせて使う必要がある**（`@computed_field`は`@property`の**上**に付ける）
    - `@property`（や`@cached_property`）を伴わない普通のメソッドに`@computed_field`を付けると、Pydanticが`PydanticUserError`を送出する
- 基本構文  
  ```python
  from pydantic import BaseModel, computed_field

  class Rectangle(BaseModel):
      width: float
      height: float

      @computed_field
      @property
      def area(self) -> float:
          return self.width * self.height

  r = Rectangle(width=3, height=4)
  print(r.area)  # 12.0 （通常のプロパティとしてアクセス可能）

  print(r.model_dump())
  # {'width': 3.0, 'height': 4.0, 'area': 12.0}  ← area も出力に含まれる

  print(r.model_dump_json())
  # {"width":3.0,"height":4.0,"area":12.0}
  ```
- **通常の`@property`だけでは`model_dump()`や`model_dump_json()`の出力に含まれない**が、`@computed_field`を付けることで出力対象になる
  ```python
  class Rectangle(BaseModel):
      width: float
      height: float

      @property  # computed_field なし
      def area(self) -> float:
          return self.width * self.height

  r = Rectangle(width=3, height=4)
  r.area           # 12.0（アクセスは可能）
  r.model_dump()   # {'width': 3.0, 'height': 4.0} ← area は含まれない
  ```
- **読み取り専用**（`r.area = 10`のように直接代入しようとするとエラーになる。値は必ず元のフィールドから計算される）
- **戻り値の型ヒントは必須**（`-> float`のように明示する。型ヒントがないと`computed_field`はエラーを出す）
- 主なオプション
  - `alias`: シリアライズ時のKey名を変更  
    ```python
    @computed_field(alias="totalArea")
    @property
    def area(self) -> float:
        return self.width * self.height
    ```
  - `return_type`: 戻り値の型を明示的に指定（型推論できない場合などに使用）
  - `repr`: `__repr__`にこのフィールドを含めるかどうか（デフォルトは`True`）
- **JSON Schemaにも反映される**（`readOnly: true`として出力される）ため、APIレスポンスのドキュメント化にも有用
- `@cached_property`との組み合わせ
  - `@property`は**アクセスするたびに毎回再計算**されるが、`@cached_property`は**初回アクセス時の計算結果をキャッシュ**し、以降は再計算しない
  - 値が変わらない・計算コストが高い派生値には`@cached_property`の方が適している  
    ```python
    from functools import cached_property
    from pydantic import BaseModel, computed_field

    class Rectangle(BaseModel):
        width: float
        height: float

        @computed_field
        @cached_property
        def area(self) -> float:
            print("計算中...")
            return self.width * self.height

    r = Rectangle(width=3, height=4)
    r.area  # "計算中..." が出力され、12.0が計算・キャッシュされる
    r.area  # キャッシュ済みのため "計算中..." は出力されない
    ```
  - **注意**: `@cached_property`はキャッシュした値をインスタンス自身に保持するため、モデルの`model_config`で`frozen=True`（イミュータブル化）を指定している場合は使えない（属性への書き込みが発生するため）
- `field_validator`・`model_validator`が**入力データの検証・変換**を行うのに対し、`computed_field`は**出力時に他フィールドから値を導出して追加する**という役割の違いがある

### `TypeAdapter`
- **`BaseModel`を継承したクラスを作らずに、任意の型（`list[int]`、`dict[str, int]`、`TypedDict`、`dataclass`、単一の`int`/`str`など）に対してPydanticのバリデーション・シリアライズ機能を使うための仕組み**
  - Pydantic v2で導入
  - 「この型のためだけにモデルクラスを定義するのは大げさ」というケースで使う
    - 例: `list[int]`をバリデーションしたいだけなのに、わざわざモデルクラスを作るのは冗長
      ```python
      # TypeAdapterを使わない場合 → この検証のためだけにクラスが必要になる
      from pydantic import BaseModel

      class IntList(BaseModel):
          values: list[int]

      IntList(values=["1", "2", "3"]).values  # [1, 2, 3]

      # TypeAdapterを使う場合 → クラス定義が不要
      from pydantic import TypeAdapter

      TypeAdapter(list[int]).validate_python(["1", "2", "3"])  # [1, 2, 3]
      ```
- 基本構文
  ```python
  from pydantic import TypeAdapter

  adapter = TypeAdapter(list[int])
  result = adapter.validate_python(["1", "2", "3"])
  print(result)  # [1, 2, 3] （文字列 "1" などがintに変換される）
  ```
- 主なメソッド
  - `validate_python(data)`: Pythonオブジェクト（dict、list、文字列など）を検証・型変換する
    ```python
    from pydantic import TypeAdapter

    class User(TypedDict):
        id: int
        name: str

    adapter = TypeAdapter(User)
    user = adapter.validate_python({"id": "1", "name": "Alice"})
    print(user)  # {'id': 1, 'name': 'Alice'}
    ```
  - `validate_json(json_data)`: JSON文字列（`bytes`/`str`）を直接検証・パースする（`validate_python(json.loads(...))`より高速）
  - `dump_python(obj)` / `dump_json(obj)`: 値をPythonオブジェクト／JSONにシリアライズする（`model_dump()`/`model_dump_json()`のTypeAdapter版）
  - `json_schema()`: JSON Schemaを生成する
- `validate_python`実行時に検証エラーがあると、`BaseModel`と同様に`ValidationError`が送出される
  ```python
  from pydantic import TypeAdapter, ValidationError

  adapter = TypeAdapter(list[int])
  try:
      adapter.validate_python(["1", "abc"])
  except ValidationError as e:
      print(e)  # "abc" は int に変換できないためエラー
  ```
- **`TypeAdapter`のインスタンス化はコストがかかる**ため、関数呼び出しのたびに生成せず、モジュールレベルなどで一度だけ生成して使い回すのが推奨される
  ```python
  from pydantic import TypeAdapter

  # 悪い例: 関数が呼ばれるたびに TypeAdapter を生成している（毎回コストがかかる）
  def parse_ids_bad(data: list) -> list[int]:
      return TypeAdapter(list[int]).validate_python(data)

  # 良い例: モジュールレベルで一度だけ生成し、使い回す
  _int_list_adapter = TypeAdapter(list[int])

  def parse_ids_good(data: list) -> list[int]:
      return _int_list_adapter.validate_python(data)
  ```
- 用途の例
  - 関数の引数・戻り値のバリデーション（`BaseModel`化するまでもない単純な型）
  - 外部APIから受け取った生の`list`/`dict`データの検証
  - `TypedDict`や`dataclass`をバリデーション付きで扱いたい場合
- `BaseModel`との使い分け
  - モデルとして再利用・メソッドを持たせたい、ネストした構造を型として明示したい → `BaseModel`
    ```python
    from pydantic import BaseModel

    class User(BaseModel):
        id: int
        name: str

        def greet(self) -> str:  # モデルにメソッドを持たせられる
            return f"Hello, {self.name}"

    user = User(id="1", name="Alice")
    user.greet()  # "Hello, Alice"
    ```
  - 単発の型（プリミティブ型、`list`/`dict`のコンテナ、既存の`TypedDict`/`dataclass`）を検証したいだけ → `TypeAdapter`
    ```python
    from pydantic import TypeAdapter

    # 関数の引数として受け取った dict[str, int] を検証したいだけで、
    # 再利用するモデルクラスもメソッドも不要なケース
    scores_adapter = TypeAdapter(dict[str, int])
    scores = scores_adapter.validate_python({"Alice": "90", "Bob": "85"})
    print(scores)  # {'Alice': 90, 'Bob': 85}
    ```