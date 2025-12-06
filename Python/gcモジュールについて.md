## `gc`モジュールについて
- `gc`モジュールは、Pythonのガベージコレクション（メモリ管理）を制御するための標準ライブラリ

## 主な用途
### ガベージコレクションの制御
- メモリの自動解放を手動で制御できる
  - `gc.enable()`: 自動ガベージコレクションを有効にする
  - `gc.disable()`: 自動ガベージコレクションを無効にする
  - `gc.collect()`: ガベージコレクションを手動で実行する
  - `gc.isenabled()`: 自動ガベージコレクションが有効かどうかを確認する
- 例  
  ```python
  import gc
  gc.disable()  # ガベージコレクションを無効化
  # メモリ集中的な処理
  gc.enable()  # ガベージコレクションを再度有効化
  gc.collect()  # 手動でガベージコレクションを実行
  ```
### メモリリークのデバッグ
- 循環参照などでメモリリークが発生している場合に、オブジェクトの追跡やデバッグが可能
  - `gc.get_objects()`: GC管理下の現在のすべてのオブジェクトを取得する
  - `gc.get_referrers(obj)`: 指定したオブジェクト**を**参照しているオブジェクトを取得する
  - `gc.get_referents(obj)`: 指定したオブジェクト**が**参照しているオブジェクトを取得する
- 例  
  ```python
  import gc

  gc.set_debug(gc.DEBUG_LEAK)
  # ... 問題のあるコード ...
  gc.collect()
  print(gc.garbage)  # 回収できなかったオブジェクト
  ```

  ```python
  import gc

  # GC管理下の全オブジェクトを取得
  all_objects = gc.get_objects()
  print(f"Total objects: {len(all_objects)}")

  # 特定オブジェクトを参照しているオブジェクトを取得
  my_list = [1, 2, 3]
  referrers = gc.get_referrers(my_list)

  # 特定オブジェクトが参照しているオブジェクトを取得
  referents = gc.get_referents(my_list)
  ```

## 注意点
- 通常、Pythonのメモリ管理は自動で行われるため、`gc`モジュールを直接使用する必要はほとんどない
- 実行中にプログラムが一時停止するので、パフォーマンスに影響を与える可能性がある