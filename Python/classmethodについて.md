## classmethodとは
- `classmethod`は、クラスに紐づくメソッドで、インスタンスではなく**クラス自体**を第一引数として受け取る
  - 通常のメソッドはインスタンス(`self`)を第一引数として受け取るのに対し、`classmethod`はクラス(**`cls`**)を受け取る  
    ```python
    class MyClass:
        class_var = "クラス変数"
        
        @classmethod
        def show_class(cls):
            print(f"cls = {cls}")
            print(f"cls.class_var = {cls.class_var}")
            print(f"clsはMyClassそのもの? {cls is MyClass}")

    MyClass.show_class()
    # cls = <class '__main__.MyClass'>
    # cls.class_var = クラス変数
    # clsはMyClassそのもの? True
    ```
- `@classmethod`デコレータを使用して定義される
- クラス名からでも、インスタンスからでも呼び出せる
  ```python
  class Dog:
      kind = "犬"

      def __init__(self, name):
          self.name = name

      @classmethod
      def what_kind(cls):
          return cls.kind

  # クラス名から呼び出し
  print(Dog.what_kind())  # 出力: 犬
  # インスタンスから呼び出し
  my_dog = Dog("ポチ")
  print(my_dog.what_kind())  # 出力: 犬
  ```

> [!CAUTION]  
> `classmethod`からインスタンス変数（`self.name`）にはアクセスできない

### 主な用途（よくある使い方）
#### 1. ファクトリメソッドの定義
- 別の方法でインスタンスを生成するためのメソッドを定義する際に使用される  
  ```python
  from datetime import datetime

  class Person:
      def __init__(self, name, age):
          self.name = name
          self.age = age
      
      @classmethod
      def from_birth_year(cls, name, birth_year):
          """生まれ年からインスタンスを作成"""
          age = datetime.now().year - birth_year
          return cls(name, age)
      
      @classmethod
      def from_string(cls, person_str):
          """文字列からインスタンスを作成"""
          name, age = person_str.split(',')
          return cls(name, int(age))
      
      def __repr__(self):
          return f"Person(name='{self.name}', age={self.age})"

  # 通常の初期化
  p1 = Person("太郎", 30)
  print(p1)  # Person(name='太郎', age=30)

  # 生まれ年から作成
  p2 = Person.from_birth_year("花子", 1995)
  print(p2)  # Person(name='花子', age=30)

  # 文字列から作成
  p3 = Person.from_string("次郎,25")
  print(p3)  # Person(name='次郎', age=25)
  ```

#### 2. クラスレベルの設定や状態管理
- クラス全体で共有される設定や状態を管理するために使用される  
  ```python
  class Configuration:
      settings = {}

      @classmethod
      def set_setting(cls, key, value):
          cls.settings[key] = value
      
      @classmethod
      def get_setting(cls, key):
          return cls.settings.get(key)

  # 設定の追加
  Configuration.set_setting("theme", "dark")
  Configuration.set_setting("language", "ja")

  # 設定の取得
  print(Configuration.get_setting("theme"))     # 出力: dark
  print(Configuration.get_setting("language"))  # 出力: ja
  ```

> [!IMPORTANT]  
> ### インスタンスメソッドとの違い
> ```python
> class Counter:
>     count = 0  # クラス変数
>     
>     def __init__(self):    
>         self.instance_count = 0  # インスタンス変数
>    
>     # インスタンスメソッド:selfを受け取る
>     def increment_instance(self):
>         self.instance_count += 1
>     
>     # クラスメソッド:clsを受け取る
>     @classmethod
>     def increment_class(cls):
>         cls.count += 1
>
>     c1 = Counter()
>     c2 = Counter()
>
>     c1.increment_instance()
>     print(c1.instance_count)  # 1
>     print(c2.instance_count)  # 0 (別のインスタンス)
>
>     Counter.increment_class()
>     Counter.increment_class()
>     print(Counter.count)  # 2 (すべてのインスタンスで共有)
> ```