# 概念
- 非同期処理は、I/O待ち時間などのブロッキング操作が多いアプリケーションのパフォーマンスを向上させるための強力な機能。特にネットワーク通信やファイル操作などの場面で効果を発揮する。
- 非同期プログラミングでは、タスクが完了するのを待つ間に他のタスクを実行できる。Pythonでは`asyncio`パッケージと`async`と`await`キーワードを使用して実装する。
- 参考URL
  - https://zenn.dev/iharuoru/articles/45dedf1a1b8352

# 主要な要素
![](./image/coroutine_task_eventloop_1.jpg)

## コルーチン（Coroutine）
- `async def`で定義される関数は「コルーチン」と呼ばれ、非同期処理の基本単位

## タスク（Task）
- タスクはコルーチンをスケジュールし、その実行を管理するためのオブジェクト
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
- タスクをスケジュールする

# コルーチンの実行方法
## `await`キーワードを使用
- 別の非同期関数（`async def`で定義された関数）の中からコルーチンを呼び出す時に使う
- `await`キーワードは「この操作が完了するまで待機し、その間は他のタスクを実行できる」ことを示す
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
> [!TIP]
> `*`はPythonのアンパック演算子（unpacking operator）と呼ばれるもので、リストやタプルなどのイテラブルをアンパック（展開）する  
>  ```python
>  def add(a, b, c):
>      return a + b + c
>
>  numbers = [1, 2, 3]
>  result = add(*numbers)  # add(1, 2, 3) と同等
>  ```

> [!TIP]
> `async with`は通常の`with`文の非同期版で、非同期コンテキストマネージャを扱うために使用する。これにより、リソースの初期化と解放を非同期的に行える。
> 例えば`async with session.get(url) as response`の部分では、
> 1. `session.get(url)`は非同期操作で、HTTPリクエストを開始
> 2. `async with`はこの操作の完了を待ち、結果を`response`に代入
> 3. ブロックが終了すると、`response`オブジェクトの`__aexit__`メソッドが非同期的に呼び出され、リソース（接続など）が適切に解放される

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