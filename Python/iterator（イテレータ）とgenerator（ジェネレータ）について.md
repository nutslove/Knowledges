# イテレータ（iterator）とは
- 要素を順次取得できるオブジェクト
- イテレータは、`__iter__()`メソッドと`__next__()`メソッドを持つオブジェクト
- `__iter__()`メソッドは、イテレータ自身を返す
- `__next__()`メソッドは、次の要素を返し、要素がなくなった場合は`StopIteration`例外を発生させる
- `iter()`関数を使ってイテレータを作成できる

## 基本的な使い方
```python
# リストからイテレータを作成
numbers = [1, 2, 3, 4, 5]
iterator = iter(numbers)

# next()で要素を順次取得
print(next(iterator))  # 1
print(next(iterator))  # 2
print(next(iterator))  # 3

# for文での自動的な使用
for num in numbers:
    print(num)  # イテレータが内部的に使われる
```

## イテラブルとイテレータの違い
- **イテラブル（iterable）**: `__iter__()`メソッドを持つオブジェクト（リスト、文字列、辞書など）
- **イテレータ（iterator）**: `__iter__()`と`__next__()`の両方を持つオブジェクト。**iterableを`iter()`関数で変換して得られる。**
```python
# イテラブルからイテレータを作成
my_list = [1, 2, 3]  # イテラブル
my_iterator = iter(my_list)  # イテレータ

print(hasattr(my_list, '__iter__'))     # True
print(hasattr(my_list, '__next__'))     # False
print(hasattr(my_iterator, '__iter__')) # True  
print(hasattr(my_iterator, '__next__')) # True
```

---

# ジェネレータ（generator）とは
- ジェネレータは特殊なイテレータの一種で、`yield`キーワードを使って簡潔に作成できる
- イテレータを簡単に書ける仕組み
- ジェネレータはイテレータを返す関数  
  ```python
  gen = count_up_to(3)
  print(iter(gen) is gen)  # True → ジェネレータはイテレータ
  ```
- **ジェネレータは通常の関数と違い、`return`ではなく`yield`を使って値を返す**
- **`yield`を関数の中で使うと、その関数はジェネレータ関数になる**
- `yield` は「**値を返して関数を一時停止する**」という動作をする
- **再度呼び出されると、前回の続きから実行が再開されるのが特徴**

## 基本構文と動作
```python
def my_generator():
    print("Start")
    yield 1
    print("Middle")
    yield 2
    print("End")
    yield 3

gen = my_generator()
print(next(gen))  # Start → 1
print(next(gen))  # Middle → 2
print(next(gen))  # End → 3
# print(next(gen))  # StopIteration例外
```

## `for`文での使用
```python
def count_up_to(n):
    i = 1
    while i <= n:
        yield i
        i += 1

for num in count_up_to(3):
    print(num)
# 出力: 1, 2, 3
```

## ジェネレータの特徴
| 特徴          | 説明                      |
| ----------- | ----------------------- |
| 🧠 メモリ効率が良い | 全データを一度にメモリに載せない。1つずつ生成 |
| ⚙️ 状態を保持できる | `yield` によって実行状態が保存される  |
| ✨ 簡潔な記述     | イテレータクラスよりずっと短く書ける      |
