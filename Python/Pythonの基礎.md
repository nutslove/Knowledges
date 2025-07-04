- PythonでもGoと同じように使わない変数を`_`にすることができる
  - e.g.
    ~~~python
    display = []
    for _ in range(len(chosen_word)):
      display.append("_")
    ~~~

# Pythonの位置引数とキーワード引数について
- 以下のような関数があるとする
  ~~~python
  def greet(name, greeting):
    print(f"{greeting}, {name}!")
  ~~~
- 位置引数の場合は関数を呼び出す時のArgumentの順番通り、ParameterにArgumentが入る
  ~~~python
  greet("Alice", "Hello")
  # 出力: Hello, Alice!
  ~~~
- キーワード引数の場合は、Parameter名で指定するので、関数呼び出し時のArgumentの順番に影響されない
  ~~~python
  greet(greeting="Hello",name="Alice")
  # 出力: Hello, Alice!
  ~~~

---

# `except`（例外処理）について
- ある関数内の`except`で`raise`で上げたエラー内容は呼び出し元の関数に伝播される
- 例
  - 呼び出し元  
    ```python
    import test1

    try:
      test1.test_func1("ng")
    except Exception as e:
      print(f"error: {e}")
    ```  
    **→ `"error: 呼び出し先でエラーが発生しました"`が出力される**
  - 呼び出し先（`test1.py`）  
    ```python
    def test_func1(arg1: str):
      try:
        if arg1 == "ok":
          pass
        else:
          raise ValueError("呼び出し先でエラーが発生しました")
      except Exception as e:
        raise e
    ```

---

# 組み込み関数
## `index()`メソッド
- **ある要素がListの中で何番目にあるか確認できる**
- 以下のリストがあるとしたら`index = fruits.index("apple")`(→0が返ってくる)のように`<List名>.index("<要素名>")`で確認できる
  - `fruits = ["apple", "banana", "cherry"]`
- Listの中に同じ要素が複数ある場合、`index()`メソッドは最初にヒットした要素のIndexを返す
- 要素がリストにない場合、IndexError例外が発生する

## `rindex()`メソッド
- **同じ要素がListの中に複数ある場合、最後のIndexを確認する方法**
```python
fruits = ["apple", "banana", "apple", "cherry"]
index = fruits.rindex("apple")
print(index) --> 2が出力
```
- 要素がリストにない場合、-1 が返される

## `sorted()`関数
- **Listをソートするための組み込み関数**
- 元のリストを変更せずに、新しいソートされたリストを返す
- 例  
  ```python
  numbers = [5, 2, 9, 1, 5, 6]
  sorted_numbers = sorted(numbers)
  print(numbers)  # 出力: [5, 2, 9, 1, 5, 6] (元のリストは変更されない)
  print(sorted_numbers)  # 出力: [1, 2, 5, 5, 6, 9]
  ```

## `set()`関数
- **Listの重複を排除して、ユニークな要素のみを含むセットを作成するための組み込み関数**
- 例  
  ```python
  numbers = [1, 2, 2, 3, 4, 4, 5]
  unique_numbers = set(numbers)
  print(unique_numbers)  # 出力: {1, 2, 3, 4, 5} (順序は保証されない)

  raw_data = [1, 3, 2, 1, 4, 2, 5, 3]

  # 重複排除 + ソート
  clean_data = sorted(set(raw_data))
  print(clean_data)  # [1, 2, 3, 4, 5]

  # 機械学習でのラベル一覧取得など
  labels = ['cat', 'dog', 'cat', 'bird', 'dog']
  unique_labels = sorted(set(labels))
  print(unique_labels)  # ['bird', 'cat', 'dog']
  ```

---

# `if <変数名>:`の意味
> Pythonにおける if <変数名>: 構文は、変数の「真偽値（truthiness）」を評価します。これは、**変数が定義されているかどうかだけではなく、その値が「偽（False）」と評価されるか「真（True）」と評価されるかをチェック**します。
>
> 以下は、Pythonにおける「偽（False）」と評価される主な値のリストです：
>
>- None
>- False（ブーリアン型のFalse）
>- ゼロの数値：0, 0.0, 0j（整数、浮動小数点数、複素数のゼロ）
>- 空のコレクション：'', (), [], {}（空文字列、空タプル、空リスト、空辞書）
>- カスタムクラスのインスタンスで __bool__() や __len__() メソッドがゼロまたは偽を返すもの。
>
> 逆に、これら以外の値はすべて「真（True）」と評価されます。
>
> したがって、**`if <変数名>:`のコードは変数が定義されていて、かつその値が「真」であるかどうかをチェック**します。**値が「偽」である場合、ifブロックは実行されません**。もし変数が定義されていない場合、PythonはNameErrorを投げます。

---

# lambda関数(無名関数)
- 文法
  `lambda 引数: 戻り値`
  - 以下のように引数なしで実行することもできる  
    `lambda: random.rand()`
- 例
  ~~~python
  loaders = {
      "pdf": PyPDFLoader,
      "txt": lambda path: TextLoader(path, autodetect_encoding=True),
      "docx": Docx2txtLoader,
  }

  if file_type in loaders:
      loader = loaders[file_type](tmp_location)
  ~~~
  - TextLoaderクラスのインスタンスを生成する際に、`autodetect_encoding=True`を自動的に引数として渡します。このlambda関数自体が、`TextLoader`を呼び出す際に必要なすべての引数を内包しており、外部から直接`autodetect_encoding`に関する指定をする必要はありません。  
  `loader = loaders[file_type](tmp_location)`の行で、ファイルタイプに応じたローダーが呼び出される際には、そのローダーに対して`tmp_location`のみが引数として渡されます。しかし、"txt"のファイルタイプに対応するローダー（この場合はlambda関数）には、このlambda関数内で`TextLoader`のコンストラクタに`path`と`autodetect_encoding=True`の両方を渡すように定義されています。  
  つまり、lambda関数を介して`TextLoader`を呼び出す際には、lambda関数が受け取った`tmp_location`（`path`として受け取る）を`TextLoader`の第一引数として、そしてlambda関数の定義により`autodetect_encoding=True`が自動的に第二引数として`TextLoader`に渡されます。

---

# `None`とは
- 他の言語の`null`や`nil`に該当するもの。  
  Pythonにおける特殊な値で、"何もない"、"値が存在しない"を意味する。
- NoneはPythonの組み込み定数であり、変数が何も参照していないことを示すために使用される。

---

# 型ヒント
## 変数の型ヒント
- 書き方
  - `<変数名>: <型> = <値>`
- `:`の後ろにあるのはPython3.5以降で追加された型ヒント機能で、変数・関数の引数・戻り値の期待されるデータ型を指定するために使用される。
- `|`（パイプ）演算子はPython3.10で導入され、型ヒントの文脈で使用されると「和」または「ユニオン」型 (つまり **"OR"** )を意味する。つまり、変数が指定された型のいずれか一つであることを示す。例えば、`int | None`は、変数がint型またはNoneのいずれかであることを意味する。
- **型を強制する機能はない**
  - 主にコードの可読性を高め、開発者が変数や関数の期待するデータ型を明示するために使用される
- 例
  ~~~python
  from ragas.metrics.base import Metric
  def evaluate_with_chain(
    r_metrics: list[Metric] | None = None, ---> r_metricsには Metric型のリスト or None が入るという型ヒント
    r_is_async: bool = False, ---> r_is_asyncにはbool型が入るという型ヒント
    r_column_map: dict[str, str] | None = {}, ---> r_column_mapにはString型のKeyとString型のValueの辞書か、Noneが入るという型ヒント
  ):
    ・・・ある処理・・・
  ~~~
## 関数の型ヒント
- 書き方
  ~~~python
  def <関数名>(引数名: 型, ・・・) -> <戻り値の型>:

    ・・・ある処理・・・

    return <戻り値>
  ~~~
### 戻り値が複数ある場合の戻り値の型ヒント
- 戻り値が複数ある場合、戻り値の型ヒントは`typing`の`Tuple`を使って１つのTupleの中に入れる必要がある
  - Pythonの型ヒントは戻り値が１つの型であることを前提にしていて、Tupleにすることで１つの型として扱えるようにする
- 例  
  ```python
  from typing import Tuple

  def get_user_info() -> Tuple[str, int, bool]:
      name = "Alice"
      age = 30
      is_active = True
      return name, age, is_active
  ```

## クラスの型ヒント
- 例(1)
  ~~~python
  class Curry:
      beef: int
      onion: int
      potato: int
      carrot: int
      roux: int
      rice: int

      def __init__(self, beef: int, onion: int, potato: int, carrot: int, roux: int) -> None:
          self.beef = beef
          self.onion = onion
          self.potato = potato
          self.carrot = carrot
          self.roux = roux

  curry: Curry = Curry(beef=250, onion=400, potato=230, carrot=100, roux=115)
  ~~~
- 例(2)
  ~~~python
  class Person:
      name: str
      age: int

      def __init__(self, name: str, age: int):
          self.name = name
          self.age = age

      def greet(self) -> str:
          return f"Hello, my name is {self.name} and I am {self.age} years old."
  ~~~

### `Config` クラスを通じて追加の制約を設定する
- **Pydanticは `Config` クラスを通じて追加の制約を設定することができる。例えば、`frozen=True` を設定すると、インスタンス化後の属性変更を完全に禁止することができる。**
  ```python
  class User(BaseModel):
      name: str
      age: int

      class Config:
          frozen = True

  user = User(name="Alice", age=30)
  user.age = 31  # これはエラーになる
  ```

---

# 別のPythonファイルをimportする方法
## 同じディレクトリ内の別ファイルをimportする場合
- ファイル名をそのまま使ってimport  
  ```python
  import <インポートするpythonファイル名(e.g. mymodule.pyの場合、mymodule)>
  from <インポートするpythonファイル名> import <関数名>
  ```
## 異なるディレクトリ内の別ファイルをimportする場合
- 相対import  
  ```python
  import module  # 同じディレクトリ
  from .. import module  # 親ディレクトリ
  from ..somedirectory import module  # 親ディレクトリの中のsomedirectoryからのインポート
  from ...somedirectory import module  # 2つ上の親ディレクトリの中のsomedirectoryからのインポート
  from ..somedirectory.somedirectory import module # 親ディレクトリの中のsomedirectoryの中のsomedirectoryからのインポート
  from .subdirectory import module  # 子ディレクトリ
  ```

---

# `*args`と`**kwargs`について
- **`*args`には *tuple* で、`**kwargs`には *dict()* で入る**
- 例１
  - Python  
    ```python
    def test(*args, **kwargs):
        print(args)
        print(kwargs)

    test(1, 2, 3, 4, 5, col=4, row=5)
    ```
  - Output  
    ```shell
    (1, 2, 3, 4, 5)
    {'col': 4, 'row': 5}
    ```
- 例２（該当する引数がない場合）
  - Python  
    ```python
    def test(*args, **kwargs):
        print(args)
        print(kwargs)

    test()
    ```
  - Output  
    ```shell
    ()
    {}
    ```
- 例３（`**kwargs`から１つずつkeyとvalueを受け取る）
  - Python  
    ```python
    def print_info(**kwargs):
        for key, value in kwargs.items():
            print(f"{key}: {value}")

    print_info(name="Alice", age=30, city="New York")
    ```
  - Output  
    ```shell
    name: Alice
    age: 30
    city: New York
    ```
- 例４（`**kwargs`から特定のKeyのValueを取り出す）
  - `kwargs.get('<対象Key名>')`と`kwargs['<対象Key名>']`の２通りのやり方がある
    - `kwargs.get`は`kwargs.get('<対象Key名>', '<対象Keyのものがない場合のdefault値>')`でdefault値を定義することもできる
    ```python
    def extract_key(**kwargs):
        value = kwargs.get('key_name', 'default_value')
        print(f"The value is: {value}")

    # 使用例
    extract_key(key_name='example', other_key='other_value')  # The value is: example
    extract_key(other_key='other_value')  # The value is: default_value
    ```
- 例５（`**kwargs`で受け取ったdictにkey,valueペアを追加）
  - Python  
    ```python
    def add_to_kwargs(**kwargs):
        # 新しいkey-valueペアを追加
        kwargs['new_key'] = 'new_value'
        
        # 別の方法で追加
        kwargs.update({'another_key': 'another_value'})
        
        # 結果を表示
        for key, value in kwargs.items():
            print(f"{key}: {value}")

    # 関数を呼び出し
    add_to_kwargs(existing_key='existing_value', foo='bar')
    ```
  - Output  
    ```shell
    existing_key: existing_value
    foo: bar
    new_key: new_value
    another_key: another_value
    ```

---

# `isinstance()`関数によるobjectの型判定
- https://docs.python.org/ja/3/library/functions.html#isinstance
- `isinstance(object, classinfo)`の形で、第１引数で指定した`object`が、第２引数で指定した`classinfo`型(またはそのサブクラスのインスタンス)である場合`True`を返す。`object`が`classinfo`型のオブジェクトでない場合`False`を返す。

## `isinstance()`関数と`type()`関数の違い
- https://qiita.com/Ryo-0131/items/c5c650359ab8ce10b507
- isinstance()は継承関係を考慮して型をチェックするのに対し、type()はオブジェクトの型そのものを返す関数なので、サブクラスまで考慮したい場合はisinstance()を使うこと。
- 比較例  
  ```python
  class Fruit:
      pass

  class Apple(Fruit):
      pass

  obj_fruit = Fruit()
  obj_apple = Apple()

  print(isinstance(obj_fruit, Apple))  # False
  print(type(obj_fruit) == Apple)      # False

  print(isinstance(obj_apple, Fruit))  # True
  print(type(obj_apple) == Fruit)      # False
  ```

---

# 関数で、デフォルト値を持つ引数とデフォルト値を持たない引数の順番
- Pythonの文法上の制約で、関数にてデフォルト値を持つ引数の後にデフォルト値を持たない引数を置くことはできない
- https://stackoverflow.com/questions/24719368/syntaxerror-non-default-argument-follows-default-argument
- NG例  
  ```python
  def a(len1, hgt=len1, til, col=0): # デフォルト値を持つhgtの後のtilがデフォルト値を持たないためNG
  ```
- OK例  
  ```python
  def example(a, b, c=None, r="w", d=[], *ae,  **ab):
  ```

---

# 特殊メソッド
## `__dict__`メソッド
- Objectが持つ**属性**と**その値**を格納する**辞書**
- **Class**や**Instance**の属性を動的に確認・操作するために使用される
- 例 (インスタンス)  
  ```python
  class MyClass:
      def __init__(self, x, y):
          self.x = x
          self.y = y

  obj = MyClass(1, 2)
  print(obj.__dict__)
  # {'x': 1, 'y': 2}

  ## 追加
  obj.__dict__['z'] = 3
  print(obj.__dict__)
  # {'x': 1, 'y': 2, 'z': 3}

  ## 更新
  obj.__dict__['x'] = 4
  print(obj.__dict__)
  # {'x': 4, 'y': 2, 'z': 3}

  ## 削除h
  del obj.__dict__['x']
  print(obj.__dict__)
  # {'y': 2, 'z': 3}
  ```

- 例 (クラス)  
  ```python
  class MyClass:
      class_var = "クラス変数"

      def __init__(self, instance_var):
          self.instance_var = instance_var

  ## クラスの__dict__を表示
  print(MyClass.__dict__)
  # {'__module__': '__main__', 'class_var': 'クラス変数', '__init__': <function MyClass2.__init__ at 0x7ff57b163790>, '__dict__': <attribute '__dict__' of 'MyClass2' objects>, '__weakref__': <attribute '__weakref__' of 'MyClass2' objects>, '__doc__': None}
  ```

- 例 (`__repr__`メソッドとの組み合わせ)  
  ```python
  class ConfigObject:
      def __repr__(self):
          return str(self.__dict__)

  fluentd = ConfigObject()
  print("fluentd:", fluentd)
  # fluentd: {}

  fluentd.agent_config = ConfigObject()
  print("fluentd:", fluentd)
  # fluentd: {'agent_config': {}}
  print("fluentd.agent_config:", fluentd.agent_config)
  # fluentd.agent_config: {}

  fluentd.agent_config.url = "http://someurl.com"
  fluentd.agent_config.someparam = "somevalue"
  print("fluentd:", fluentd)
  # fluentd: {'agent_config': {'url': 'http://someurl.com', 'someparam': 'somevalue'}}
  print("fluentd.agent_config:", fluentd.agent_config)
  # fluentd.agent_config: {'url': 'http://someurl.com', 'someparam': 'somevalue'}

  conf = fluentd.agent_config
  print(conf.url)
  print(fluentd.agent_config.url)
  # http://someurl.com
  print(conf.someparam)
  print(fluentd.agent_config.someparam)
  # somevalue
  ```

## `__repr__`メソッド
- https://docs.python.org/ja/3.11/library/functions.html#repr
- `__repr__`は`object`クラスで定義されており、すべてのクラスは暗黙的に`object`クラスを継承する。そのため、すべてのクラスに`__repr__`メソッドは定義（継承）されている。
- **オブジェクトの内部状態（文字列）を返す特殊な文字列**
  - 通常デバックや開発者がオブジェクトの状態を確認するために使われるみたい
- クラスは、 **`__repr__()`メソッドを定義することで、この関数によりそのクラスのインスタンスが返すものを制御することができる**
- `__repr__`の戻り値は**文字列**でなければならない
- 例  
  ```python
  class Person:
    def __init__(self, name, age):
        self.name = name
        self.age = age

    def __repr__(self):
        return f"Person('{self.name}', {self.age})"

  person = Person("John", 52)
  print(person) 
  print(repr(person)) ## print(person)と同じ 
  # Person('John', 52)
  print(person.name)
  # John
  print(person.age)
  # 52
  ```

## `__call__`メソッド
- オブジェクトを関数のように呼び出し可能にする特殊メソッド
- このメソッドを定義することで、初期化したインスタンスに対して括弧`()`を使った呼び出し構文を使用できるようになる
- 基本的な使い方  
  ```python
  class Calculator:
      def __call__(self, a, b):
          return a + b

  calc = Calculator()
  result = calc(5, 3)  # calc.__call__(5, 3)と同じ
  print(result)  # 8
  ```

### `__call__`メソッドと`__init__`メソッドの違い
- `__init__`メソッドはクラスのインスタンス化時に呼び出され、オブジェクトの初期化を行う。
- `__call__`メソッドはインスタンスが呼び出されたときに実行される。
- **`__init__`メソッドはオブジェクトが作成されたときに1度だけ呼び出される。一方、`__call__`メソッドはインスタンスが呼び出されるたびに複数回呼び出される。**
- 例  
  ```python
  class Counter:
      def __init__(self):
          self.count = 0
          print("Counter initialized")
      def __call__(self):
          self.count += 1
          print(f"Counter called: {self.count}")
          return self.count
   
  C = Counter() # 出力: Counter initialized
  C()  # 出力: Counter called: 1
  C()  # 出力: Counter called: 2
  ```

---

# 三項演算子
![](./image/ternary_operator.jpg)
- 参照URL
  - https://atmarkit.itmedia.co.jp/ait/articles/2104/02/news016.html
- if文を1行で記述
- 例  
  ```python
  cinder_id = cinder.create_cinder_volume(
      f"{self.cluster_id}-master-opensearch-pv-{i}",
      self.data_disk_size if self.cluster_type == "standard" else 2,
      "az-a",
      self.disk_type if self.cluster_type == "standard" else "economy-medium"
  )
  ```