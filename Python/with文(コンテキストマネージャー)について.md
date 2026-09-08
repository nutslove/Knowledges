# `with`文について
- `with`文は、コンテキストマネージャーを使用するための構文で、リソースの管理（ファイル操作、ネットワーク接続、データベース接続、ロック管理など）を簡潔かつ安全に行うための仕組みであり、「使い終わったら必ず後処理が必要」な場面で活躍する
- `with`文を使うと、 **リソースの取得と解放** を自動的に行うことができる。これにより、 **リソースの解放忘れや例外発生時のリソースリークを防ぐ** ことができる。
- `with`文は内部で2つの特殊メソッドを呼び出す
  - `__enter__()`：`with`ブロックに入るときに呼び出され、リソースの取得や初期化を行う
  - `__exit__()`：`with`ブロックを抜けるときに呼び出され、リソースの解放や後処理を行う
- 基本的な使い方  
  ```python
  with expression as variable:
    # expression as variableで__enter__が呼ばれ、variableにその戻り値が代入される
    # ブロック内の処理
    ...
  # ブロックを抜けると自動的に__exit__が呼ばれる
  ```

## なぜwith文が必要か
- with文を使わない場合の問題  
  ```python
  # 悪い例：例外発生時にファイルが閉じられない可能性
  f = open('data.txt', 'r')
  content = f.read()  # ここで例外が発生したら...
  f.close()           # この行は実行されない！

  # try-finallyで対処できるが冗長
  f = open('data.txt', 'r')
  try:
      content = f.read()
  finally:
      f.close()
  ```
- with文を使う場合  
  ```python
  # シンプルかつ安全
  with open('data.txt', 'r') as f:
      content = f.read()
  # 例外が発生しても必ずファイルは閉じられる
  ```

## `__exit__`のシグネチャ
```python
def __exit__(self, exc_type, exc_val, exc_tb):
    # exc_type: 例外の型（例外がなければNone）
    # exc_val:  例外のインスタンス
    # exc_tb:   トレースバックオブジェクト
    
    # Trueを返すと例外を抑制
    # False/Noneを返すと例外を再送出
    return False
```

### `__exit__`の戻り値と例外の抑制/再送出
- 「withブロック内で例外が発生したとき、その例外をどう扱うか」というの制御の話
- 具体例  
  ```python
  class SuppressError:
      """例外を抑制する（握りつぶす）"""
      def __enter__(self):
          return self
      
      def __exit__(self, exc_type, exc_val, exc_tb):
          print(f"例外をキャッチ: {exc_val}")
          return True  # 例外を抑制 → プログラムは続行

  class PropagateError:
      """例外を再送出する（そのまま投げる）"""
      def __enter__(self):
          return self
      
      def __exit__(self, exc_type, exc_val, exc_tb):
          print(f"例外をキャッチ: {exc_val}")
          return False  # 例外を再送出 → 呼び出し元に伝播
  ```
  - 動作の違い  
    ```python
    # Trueを返す場合（例外を抑制）
    print("=== 抑制パターン ===")
    with SuppressError():
        raise ValueError("エラーだよ")
    print("withの後も実行される！")

    # 出力:
    # === 抑制パターン ===
    # 例外をキャッチ: エラーだよ
    # withの後も実行される！
    ```

    ```python
    # Falseを返す場合（例外を再送出）
    print("=== 再送出パターン ===")
    with PropagateError():
        raise ValueError("エラーだよ")
    print("ここは実行されない")

    # 出力:
    # === 再送出パターン ===
    # 例外をキャッチ: エラーだよ
    # Traceback (most recent call last):
    #   ...
    # ValueError: エラーだよ
    ```

- 図解  
  ```
  withブロック内で例外発生
          ↓
      __exit__が呼ばれる
          ↓
    ┌─────┴─────────┐
    ↓               ↓
  True返す     False/None返す
    ↓               ↓
  例外を握りつぶす  例外をそのまま投げる
    ↓               ↓
  with後の処理へ  プログラムがクラッシュ
                （またはtry-exceptでキャッチ）
  ```

- 実用的な使い分け  
  ```python
  # 例外を抑制したい場面の例
  from contextlib import suppress

  # ファイルがなくても気にしない
  with suppress(FileNotFoundError):
      os.remove('maybe_exists.txt')
  # ↑ 内部的に__exit__でTrueを返している

  # 例外を再送出する場面（大多数のケース）
  with open('file.txt') as f:
      data = f.read()
  # ↑ ファイル読み込みエラーは呼び出し元に知らせるべき
  ```

## 動作フロー例
```python
class MyContext:
    def __enter__(self):
        print("1. __enter__が呼ばれた")
        return self  # asで受け取る値
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        print(f"3. __exit__が呼ばれた (例外: {exc_type})")
        return False  # 例外を抑制しない

with MyContext() as ctx:
    print("2. withブロック内")

# 出力:
# 1. __enter__が呼ばれた
# 2. withブロック内
# 3. __exit__が呼ばれた (例外: None)
```

```python
import time

class Timer:
    """処理時間を計測するコンテキストマネージャー"""
    def __enter__(self):
        self.start = time.perf_counter()
        return self
    
    def __exit__(self, *args):
        self.elapsed = time.perf_counter() - self.start
        print(f"処理時間: {self.elapsed:.4f}秒")
        return False

with Timer() as t:
    # 何らかの処理
    sum(range(1000000))
# → 処理時間: 0.0234秒
```

## `contextlib`を使った方法
 - Pythonの標準ライブラリである`contextlib`の`contextlib.contextmanager`デコレータを使うと、ジェネレータ関数でコンテキストマネージャーを簡潔に定義できる  
  ```python
  from contextlib import contextmanager

  @contextmanager
  def managed_resource(name):
      print(f"リソース '{name}' を確保")
      try:
          yield name  # yieldの値がasで受け取れる
      finally:
          print(f"リソース '{name}' を解放")

  with managed_resource("database") as r:
      print(f"リソース {r} を使用中")

  # 出力:
  # リソース 'database' を確保
  # リソース database を使用中
  # リソース 'database' を解放
  ```

> [!IMPORTANT]
> ##### ポイント
> **`yield`より前が`__enter__`、`finally`ブロックが`__exit__`に相当**

## 非同期コンテキストマネージャー（async with）
- Python 3.5以降では、非同期版のコンテキストマネージャーも使える
- `async with`文、`__aenter__`と`__aexit__`を使うと、非同期リソースの管理が可能になる
- 例  
  ```python
  class AsyncResource:
      async def __aenter__(self):
          await self.connect()
          return self
      
      async def __aexit__(self, exc_type, exc_val, exc_tb):
          await self.disconnect()
          return False

  # 使用例
  async def main():
      async with AsyncResource() as resource:
          await resource.do_something()

  # asyncio.run(main()) で実行
  ```

## 複数のコンテキストマネージャー
```python
# 複数同時に使用（Python 3.1+）
with open('input.txt') as fin, open('output.txt', 'w') as fout:
    fout.write(fin.read())

# Python 3.10+では括弧で複数行に分割可能
with (
    open('input.txt') as fin,
    open('output.txt', 'w') as fout,
):
    fout.write(fin.read())
```
