# 概念
- 非同期処理は、I/O待ち時間などのブロッキング操作が多いアプリケーションのパフォーマンスを向上させるための強力な機能。特にネットワーク通信やファイル操作などの場面で効果を発揮する。
  - CPUバウンドな処理には向いていない。CPUバウンドな処理には`multiprocessing`や`concurrent.futures`の方が適している。
- 非同期プログラミングでは、タスクが完了するのを待つ間に他のタスクを実行できる。Pythonでは`asyncio`パッケージと`async`と`await`キーワードを使用して実装する。
- 参考URL
  - https://zenn.dev/iharuoru/articles/45dedf1a1b8352

# 主要な要素
![](./image/coroutine_task_eventloop_1.jpg)


           ┌────────────┐
           │ async def  │
           │ コルーチン  │
           └────┬───────┘
                │
        asyncio.create_task()
                ↓
           ┌────────┐
           │ Task   │───▶ Event Loop が管理
           └────────┘


## コルーチン（Coroutine）
- `async def`で定義される関数は「コルーチン」と呼ばれ、非同期処理の基本単位
- 一時停止と再開が可能
- 通常の関数と異なり、処理の途中で他のタスクに実行を譲ることができる

## タスク（Task）
- タスクはコルーチンをスケジュールし、その実行を管理するためのオブジェクト
  - コルーチンをイベントループで実行するためのオブジェクト
  - コルーチンをタスクとして登録することで、イベントループによる管理が可能になる
- 基本的に1つのタスクは1つのコルーチンに対応
- 複数のコルーチンを並行して実行するには、タスクを作成する  
  ```python
  async def main():
      # 3つのタスクを作成して並行実行
      task1 = asyncio.create_task(fetch_data())
      task2 = asyncio.create_task(fetch_data())
      task3 = asyncio.create_task(fetch_data())
      
      # 全てのタスクが完了するまで待機
      results = await asyncio.gather(task1, task2, task3)
      print(results)
  ```

## イベントループ（EventLoop）
- コルーチンの実行を管理し、I/Oイベントの発生を監視する中心的な役割を担う
- タスクをスケジュールする

# コルーチンの実行方法
## `await`キーワードを使用
- 別の非同期関数（`async def`で定義された関数）の中からコルーチンを呼び出す時に使う
- **`await`キーワードは「この操作が完了するまで待機し、その間は他のタスクを実行できる」ことを示す**
  - 何か別の処理が完了するまで待つ時に`await`を使う  
  ```python
  async def fetch_data():
      print("データ取得開始")
      await asyncio.sleep(2)  # データ取得を模擬（2秒待機）
      print("データ取得完了")
      return {"data": "結果"}
  ```
- **`await`は`async`内でしか使えない**

## `asyncio.run()`を使用
- 最上位レベルからコルーチンを実行するために使う。これはプログラムのエントリーポイントで一度だけ呼び出すべき。
- **`asyncio.run`は逆に`async`内では使えない**

## `asyncio.create_task()`を使用
- コルーチンをタスクに変換して並行実行させる方法

## `asyncio.gather()`を使用
- 複数のコルーチンを管理するための方法
- `asyncio.create_task()`で作成した複数のタスクをまとめて管理（すべてのタスクが完了するまで待機）  
  ```python
  async def main():
      # 複数のタスクを作成
      task1 = asyncio.create_task(fetch_data('url1'))
      task2 = asyncio.create_task(fetch_data('url2')) 
      task3 = asyncio.create_task(fetch_data('url3'))
      
      # すべてのタスクを並行実行し、結果を待機
      results = await asyncio.gather(task1, task2, task3)
      
      # 結果の処理
      for result in results:
          print(result)
  ```
  またはリスト内包表記と組み合わせる方法  
  ```python
  async def main():
      urls = ['url1', 'url2', 'url3', 'url4']
      
      # タスクのリストを作成
      tasks = [asyncio.create_task(fetch_data(url)) for url in urls]
      
      # すべてのタスクを並行実行
      results = await asyncio.gather(*tasks)
  ```
- **`asyncio.gather()`はコルーチンを内部的にタスクに変換して実行してくれるので、引数にタスクではなく、コルーチンを指定してもいい**  
  ```python
  async def hello():
      print('I say,')
      await asyncio.sleep(1) # 1秒かかる
      print('hello')

  async def goodbye():
      print('you say,')
      await asyncio.sleep(2) # 2秒かかる
      print('goodbye')

  async def main():
      await asyncio.gather(goodbye(), hello()) # タスクも同様
  ```

## `asyncio.as_completed()`
- `asyncio.gather()`はすべてのタスクが完了するまで待機するのに対し、`asyncio.as_completed()`はタスクが完了した順に結果を取得(処理)できる
```python
async def main():
    tasks = [fetch_data(url) for url in urls]
    for coro in asyncio.as_completed(tasks):
        result = await coro
        print(f"完了: {result}")  # 完了した順番で処理
```

> [!IMPORTANT]  
> - 複数のコルーチンを`asyncio.create_task()`でtaskに変換せずに直接`await`する場合、各コルーチンが逐次処理になるため、`asyncio.create_task()`を使って並行実行する方が効率的。
> - 直接`await`する場合(**6秒**かかる)  
>   ```python
>   import asyncio
>   import time
>
>   async def fetch_data(name, delay):
>       print(f"{name}: 開始")
>       await asyncio.sleep(delay)
>       print(f"{name}: 完了")
>       return f"{name}の結果"
>
>   async def sequential_example():
>       print("=== 逐次実行 ===")
>       start = time.time()
>
>       result1 = await fetch_data("タスク1", 2)  # 2秒待つ
>       result2 = await fetch_data("タスク2", 3)  # その後3秒待つ
>       result3 = await fetch_data("タスク3", 1)  # その後1秒待つ
>
>       end = time.time()
>       print(f"総実行時間: {end - start:.2f}秒")
>       return [result1, result2, result3]
>
>   def main():
>       asyncio.run(sequential_example())
>
>   if __name__ == "__main__":
>       main()
>
>   # 実行結果:
>   # === 逐次実行 ===
>   # タスク1: 開始
>   # タスク1: 完了
>   # タスク2: 開始
>   # タスク2: 完了
>   # タスク3: 開始
>   # タスク3: 完了
>   # 総実行時間: 6.00秒（2+3+1秒）
>   ```
> - `asyncio.create_task()`を使って並行実行する場合(**3秒**かかる)  
>   ```python
>   import asyncio
>   import time
>
>   async def fetch_data(name, delay):
>       print(f"{name}: 開始")
>       await asyncio.sleep(delay)
>       print(f"{name}: 完了")
>       return f"{name}の結果"
>
>   async def concurrent_example():
>       print("=== 並行実行 ===")
>       start = time.time()
>
>       # タスクを作成（この時点ですぐに実行開始）
>       task1 = asyncio.create_task(fetch_data("タスク1", 2))
>       task2 = asyncio.create_task(fetch_data("タスク2", 3))
>       task3 = asyncio.create_task(fetch_data("タスク3", 1))
>
>       # 結果を待つ
>       result1 = await task1
>       result2 = await task2
>       result3 = await task3
>
>       end = time.time()
>       print(f"総実行時間: {end - start:.2f}秒")
>       return [result1, result2, result3]
>
>   def main():
>       asyncio.run(concurrent_example())
>
>   if __name__ == "__main__":
>       main()
> 
>   # 実行結果:
>   # === 並行実行 ===
>   # タスク1: 開始
>   # タスク2: 開始
>   # タスク3: 開始
>   # タスク3: 完了
>   # タスク1: 完了
>   # タスク2: 完了
>   # 総実行時間: 3.00秒
>   ```

# 実践的な例
## 複数のURLから同時にデータを取得
```python
import asyncio
import aiohttp
import time

async def fetch(session, url):
    async with session.get(url) as response:
        return await response.text()

async def fetch_all(urls):
    async with aiohttp.ClientSession() as session:
        tasks = [fetch(session, url) for url in urls]
        return await asyncio.gather(*tasks)

async def main():
    urls = [
        'https://example.com',
        'https://python.org',
        'https://docs.python.org'
    ]
    
    start = time.time()
    results = await fetch_all(urls)
    end = time.time()
    
    print(f"取得完了: {len(results)} URLs in {end - start:.2f} 秒")
    print(f"最初の結果の長さ: {len(results[0])} 文字")

# 実行
asyncio.run(main())
```
> [!NOTE]
> `*`はPythonのアンパック演算子（unpacking operator）と呼ばれるもので、リストやタプルなどのイテラブルをアンパック（展開）する  
>  ```python
>  def add(a, b, c):
>      return a + b + c
>
>  numbers = [1, 2, 3]
>  result = add(*numbers)  # add(1, 2, 3) と同等
>  ```

> [!NOTE]
> `async with`は通常の`with`文の非同期版で、非同期コンテキストマネージャを扱うために使用する。これにより、リソースの初期化と解放を非同期的に行える。
> 例えば`async with session.get(url) as response`の部分では、
> 1. `session.get(url)`は非同期操作で、HTTPリクエストを開始
>     - この操作が完了するまでの間、CPUをブロックせず、他のコルーチンや処理がCPUを使えるようにする 
> 2. `async with`はこの操作の完了を待ち、結果を`response`に代入
> 3. ブロック（`async with`の下のコード）が終了すると、`response`オブジェクトの`__aexit__`メソッドが非同期的に呼び出され、リソース（接続など）が適切に解放される

## 非同期でのファイル操作
```python
import asyncio
import aiofiles

async def read_file(filename):
    async with aiofiles.open(filename, 'r') as f:
        return await f.read()

async def write_file(filename, content):
    async with aiofiles.open(filename, 'w') as f:
        await f.write(content)

async def process_files():
    # 複数のファイルを並行して読み込み
    tasks = [
        read_file('file1.txt'),
        read_file('file2.txt'),
        read_file('file3.txt')
    ]
    
    contents = await asyncio.gather(*tasks)
    
    # 処理を行う（例：すべての内容を結合）
    combined = '\n'.join(contents)
    
    # 結果を書き込み
    await write_file('combined.txt', combined)

# 実行
asyncio.run(process_files())
```

# その他
## エラーハンドリング
- `asyncio.gather()`には`return_exceptions=True`オプションがあり、これを指定するといずれかのタスクが例外を発生させても他のタスクは続行され、例外オブジェクトが結果リストに含まれる。指定しない場合は最初の例外で全体が中断される。
```python
async def main():
    # 3つのタスクを作成して並行実行
    task1 = asyncio.create_task(fetch_data())
    task2 = asyncio.create_task(fetch_data())
    task3 = asyncio.create_task(fetch_data())
    
    # 全てのタスクが完了するまで待機
    results = await asyncio.gather(task1, task2, task3, return_exceptions=True)
    print(results)
    # 例外が発生した場合の結果例
    # results = [正常な結果, Exception('エラー'), 正常な結果]
```

## タイムアウト処理
- 長時間実行されるタスクに対して`asyncio.wait_for()`を使ってタイムアウトを設定できる
```python
try:
    result = await asyncio.wait_for(long_running_task(), timeout=5.0)  # 5秒でタイムアウト
except asyncio.TimeoutError:
    print("処理がタイムアウトしました")
```

## キャンセル処理
```python
task = asyncio.create_task(some_coroutine())
# 何らかの条件でキャンセルしたい場合
task.cancel()
try:
    await task  # キャンセルされたタスクを待機するとCancelledErrorが発生
except asyncio.CancelledError:
    print("タスクがキャンセルされました")
```

## セマフォを使った並行処理の制限
```python
# 最大5つのタスクを同時実行
semaphore = asyncio.Semaphore(5)

async def limited_task(n):
    async with semaphore:  # セマフォを獲得
        await some_heavy_task(n)  # リソースを消費する処理
```

## 非同期ジェネレータ
- `async for`を使った非同期イテレーション
```python
import asyncio

async def async_generator():
    for i in range(10):
        await asyncio.sleep(0.1)
        yield i

async def main():
    async for value in async_generator():
        print(value)

asyncio.run(main())
```