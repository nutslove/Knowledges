## `isinstance`関数とは
- Python の組み込み関数で、オブジェクトが特定のクラスやデータ型のインスタンスかどうかを確認するために使われる

### 基本構文
```python
isinstance(object, classinfo)
```
- `object`: 型を確認したいオブジェクト
- `classinfo`: データ型、クラス、またはデータ型/クラスのタプル
- オブジェクトが指定されたクラスまたはサブクラスのインスタンスである場合は `True` を返し、そうでない場合は `False` を返す
- 例  
  ```python
  x = 5
  print(isinstance(x, int))  # True (xは整数)
  print(isinstance(x, str))  # False (xは文字列ではない)

  name = "Python"
  print(isinstance(name, str))  # True (nameは文字列)

  my_list = [1, 2, 3]
  print(isinstance(my_list, list))  # True (my_listはリスト)

  # 複数の型を確認する場合
  print(isinstance(x, (int, float)))  # True (xは整数またはfloat)
  ```