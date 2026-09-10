## ABC（Abstract Base Classes）とは
- Pythonの標準ライブラリ`abc`モジュールで提供される、抽象的な（共通の）インターフェースを定義するための仕組み
- サブクラスで必ず実装しなければならないメソッドを指定できる（サブクラスに「必ず実装しなければならないメソッド」を強制できる）。抽象メソッドが実装されていないサブクラスはインスタンス化できない。
- 抽象基底クラスで`ABC`を継承し、`@abstractmethod`デコレータを使用して抽象メソッドを定義する
  - サブクラスは抽象基底クラスを継承し、`@abstractmethod`で定義されたすべての抽象メソッドを実装する必要がある
- 例  
  ```python
  from abc import ABC, abstractmethod
  # ABCの代わりにABCMetaになっているところもあるが、Python 3.4以降はABCを使うのが一般的
  from typing import List

  ## 抽象基底クラス
  class DataProcessor(ABC):
      @abstractmethod
      def load_data(self, source: str) -> List:
          pass
      
      @abstractmethod
      def process_data(self, data: List) -> List:
          pass
      
      @abstractmethod
      def save_data(self, data: List, destination: str) -> None:
          pass
      
      # テンプレートメソッド（通常のメソッドも定義可能）
      def execute(self, source: str, destination: str):
          data = self.load_data(source)
          processed_data = self.process_data(data)
          self.save_data(processed_data, destination)

  ## サブクラス
  class CSVProcessor(DataProcessor):
      def load_data(self, source: str) -> List:
          # CSV読み込み処理
          return ["csv", "data"]
      
      def process_data(self, data: List) -> List:
          # CSV用の処理
          return [item.upper() for item in data]
      
      def save_data(self, data: List, destination: str) -> None:
          # CSV保存処理
          print(f"Saving {data} to {destination}")
  ```

---

## `collections.abc`モジュール
- Pythonが用意してくれた「よく使う型の抽象基底クラス集」

### 主なもの
| 抽象基底クラス | 意味 | 該当する型の例 |
| --- | --- | --- |
| `Iterable` | イテラブル（forで回せる）なオブジェクト | `list`, `tuple`, `set`, `dict`, `str` |
| `Iterator` | イテレータ（`__next__`メソッドを持つ）なオブジェクト | `file`オブジェクト、ジェネレータ、`iter()`の戻り値 |
| `Sequence` | シーケンス型（順序があり、インデックスでアクセスできる）なオブジェクト | `list`, `tuple`, `str` |
| `Mapping` | マッピング型（キーと値のペアでデータを保持する）なオブジェクト | `dict` |
| `Set` | 集合型（重複しない要素の集まり）なオブジェクト | `set`, `frozenset` |
| `MutableSequence` | 変更可能なシーケンス型 | `list` |
| `MutableMapping` | 変更可能なマッピング型 | `dict` |
| `Callable` | 呼び出し可能なオブジェクト | 関数、メソッド、lambda |
| `Container` | `in`演算子が使えるオブジェクト | `list`, `tuple`, `set`, `dict`, `str` |
| `Awaitable` | `await`可能なオブジェクト | コルーチン、`asyncio.Future`, `asyncio.Task` |

> [!NOTE]  
> #### Iteratorの例
> ```python
> my_list = [1, 2, 3]
> my_iterator = iter(my_list)  # これがIterator
> # next(my_iterator)
> # 1
> # next(my_iterator)
> # 2
> # next(my_iterator)
> # 3
> # next(my_iterator)
> # StopIteration例外が発生
> ```

> [!NOTE]  
> #### Callableの例
> ```python
> from collections.abc import Callable
>
> # Callable[[引数1の型, 引数2の型, ...], 戻り値の型]
> # → Callable[[int, int], int] は「int型の引数を2つ受け取り、int型を返す」呼び出し可能オブジェクト
> def apply_func(func: Callable[[int, int], int], a: int, b: int) -> int:
>     return func(a, b)
>
> def add(a: int, b: int) -> int:
>     return a + b
>
> print(apply_func(add, 3, 5))              # 8
> print(apply_func(lambda x, y: x * y, 3, 5))  # 15
>
> # isinstanceでのチェックも可能
> print(isinstance(add, Callable))  # True
> print(isinstance(3, Callable))    # False
> ```

> [!NOTE]  
> #### Awaitableの例
> ```python
> import asyncio
> from collections.abc import Awaitable
>
> async def fetch_data() -> str:
>     await asyncio.sleep(1)
>     return "data"
>
> coro = fetch_data()
> print(isinstance(coro, Awaitable))  # True（コルーチンはAwaitable）
>
> async def main():
>     result = await coro  # Awaitableなオブジェクトはawaitできる
>     print(result)
>
> asyncio.run(main())
> ```

### 主な用途
1. **型チェック**
   - `isinstance`関数を使って、オブジェクトが特定の抽象基底クラスを実装しているかどうかを確認できる  
     ```python
     from collections.abc import Iterable

     def process_items(items):
         if not isinstance(items, Iterable):
             raise TypeError("items must be an iterable")
         for item in items:
             print(item)
     ```
2. **型ヒント**
   - 関数の引数や戻り値の型ヒントとして使用できる  
     ```python
     from collections.abc import Mapping, Sequence

     def merge_dicts(dict1: Mapping, dict2: Mapping) -> Mapping:
         result = dict(dict1)
         result.update(dict2)
         return result

     # 型ヒント：「listでもtupleでもSequenceならOK」
     def process(items: Sequence[int]):
         for item in items:
             print(item)

     process([1, 2, 3])    # OK
     process((1, 2, 3))    # OK
     ```