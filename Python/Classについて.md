## 概念
- Class
  - 鋳型(주형)
- Object
  - 鋳型(주형)から作られたもの
- Attribute
  - Class内の変数
- Method
  - Class内の関数
- `__init__` (생성자)
  - Objectを作るときに実行される関数
- Instance
  - メモリ内のObject
  - Objectの概念の中にInstanceがあるイメージ  
    ![](image/object&instancejpg.jpg)

## `self`について
- Object自分自身を指すもの  
  ~~~python
  class Dog:
    def __init__(self, name):
      self.name = name
    
  my_dog = Dog("dasomi")
  print(my_dog.name)
  ~~~
  - 例えば上記の場合、`my_dog`が`self`に、`dasomi`が`name`に入る
    ![](image/self.jpg)