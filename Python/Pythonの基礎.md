- PythonでもGoと同じように使わない変数を`_`にすることができる
  - e.g.
    ~~~python
    display = []
    for _ in range(len(chosen_word)):
      display.append("_")
    ~~~

### Pythonの位置引数とキーワード引数について
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

### `index()`メソッドで、ある要素がListの中で何番目にあるか確認する方法
- 以下のリストがあるとしたら`index = fruits.index("apple")`(→0が返ってくる)のように`<List名>.index("<要素名>")`で確認できる
  - `fruits = ["apple", "banana", "cherry"]`
- Listの中に同じ要素が複数ある場合、`index()`メソッドは最初にヒットした要素のIndexを返す
- 要素がリストにない場合、IndexError例外が発生する
### `rindex()`メソッドで、同じ要素がListの中に複数ある場合、最後のIndexを確認する方法
~~~python
fruits = ["apple", "banana", "apple", "cherry"]
index = fruits.rindex("apple")
print(index) --> 2が出力
~~~
- 要素がリストにない場合、-1 が返される