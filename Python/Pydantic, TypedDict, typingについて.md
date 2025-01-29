# `typing`
## `typing`ライブラリの型ヒント
- `typing` ライブラリは（Python3.5以降に導入された）Pythonの標準ライブラリの一部で、型チェックをサポートするための機能を提供
- 普通の型ヒントにはない`Optional`や`List`などの複雑な型ヒントを使える
- **型の強制力はない**（実行時に型チェックなどは行われない）
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

# `TypedDict`
