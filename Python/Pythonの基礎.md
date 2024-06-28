- PythonでもGoと同じように使わない変数を`_`にすることができる
  - e.g.
    ~~~python
    display = []
    for _ in range(len(chosen_word)):
      display.append("_")
    ~~~

## Pythonの位置引数とキーワード引数について
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

## `index()`メソッドで、ある要素がListの中で何番目にあるか確認する方法
- 以下のリストがあるとしたら`index = fruits.index("apple")`(→0が返ってくる)のように`<List名>.index("<要素名>")`で確認できる
  - `fruits = ["apple", "banana", "cherry"]`
- Listの中に同じ要素が複数ある場合、`index()`メソッドは最初にヒットした要素のIndexを返す
- 要素がリストにない場合、IndexError例外が発生する

## `rindex()`メソッドで、同じ要素がListの中に複数ある場合、最後のIndexを確認する方法
~~~python
fruits = ["apple", "banana", "apple", "cherry"]
index = fruits.rindex("apple")
print(index) --> 2が出力
~~~
- 要素がリストにない場合、-1 が返される

## `if <変数名>:`の意味
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

## lambda関数(無名関数)
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

## `None`とは
- 他の言語の`null`や`nil`に該当するもの。  
  Pythonにおける特殊な値で、"何もない"、"値が存在しない"を意味する。
- NoneはPythonの組み込み定数であり、変数が何も参照していないことを示すために使用される。

## 型ヒント
### 変数の型ヒント
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
### 関数の型ヒント
- 書き方
  ~~~python
  def <関数名>(引数名: 型, ・・・) -> <戻り値の型>:

    ・・・ある処理・・・

    return <戻り値>
  ~~~
### クラスの型ヒント
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

## 別のPythonファイルをimportする方法
### 同じディレクトリ内の別ファイルをimportする場合
- ファイル名をそのまま使ってimport  
  ```python
  import <インポートするpythonファイル名(e.g. mymodule.pyの場合、mymodule)>
  from <インポートするpythonファイル名> import <関数名>
  ```
### 異なるディレクトリ内の別ファイルをimportする場合
- 相対import  
  ```python
  import module  # 同じディレクトリ
  from .. import module  # 親ディレクトリ
  from ..somedirectory import module  # 親ディレクトリの中のsomedirectoryからのインポート
  from ...somedirectory import module  # 2つ上の親ディレクトリの中のsomedirectoryからのインポート
  from ..somedirectory.somedirectory import module # 親ディレクトリの中のsomedirectoryの中のsomedirectoryからのインポート
  from .subdirectory import module  # 子ディレクトリ
  ```

## `*args`と`**kwargs`について
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