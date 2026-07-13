## staticmethodとは
- `staticmethod`は、**クラスに属してはいるが、インスタンス（`self`）にもクラス（`cls`）にもアクセスしない**メソッド
  - 通常のインスタンスメソッドは`self`を、`classmethod`は`cls`を第一引数として自動で受け取るのに対し、`staticmethod`は**暗黙の第一引数を一切受け取らない**（＝ただの関数に近い）
    ```python
    class MyClass:
        class_var = "クラス変数"

        @staticmethod
        def show():
            print("staticmethodはselfもclsも受け取らない")
            # print(self.xxx)  → NameError: selfは定義されていない
            # print(cls.xxx)   → NameError: clsは定義されていない

    MyClass.show()
    # staticmethodはselfもclsも受け取らない
    ```
- `@staticmethod`デコレータを使用して定義される
- クラス名からでも、インスタンスからでも呼び出せる
  ```python
  class MathUtils:
      @staticmethod
      def add(a, b):
          return a + b

  # クラス名から呼び出し
  print(MathUtils.add(2, 3))  # 出力: 5

  # インスタンスから呼び出し
  m = MathUtils()
  print(m.add(10, 20))        # 出力: 30
  ```

> [!CAUTION]
> `staticmethod`からはインスタンス変数（`self.xxx`）にもクラス変数（`cls.xxx`）にもアクセスできない。
> クラス変数を参照したい場合はクラス名を直接書く（`MyClass.class_var`）必要があり、その場合は`classmethod`の方が適していることが多い。

## なぜ「ただの関数」ではなくクラス内に置くのか
- 機能的には「モジュールレベルの関数」と同じことができるが、**論理的にそのクラスに関連する処理**をクラス内にまとめることで、コードの凝集度と可読性が上がる
  ```python
  class Temperature:
      def __init__(self, celsius):
          self.celsius = celsius

      # 「温度」に関連するユーティリティなのでクラス内に置く
      @staticmethod
      def celsius_to_fahrenheit(celsius):
          return celsius * 9 / 5 + 32

      @staticmethod
      def is_valid_celsius(celsius):
          return -273.15 <= celsius

  print(Temperature.celsius_to_fahrenheit(100))  # 212.0
  print(Temperature.is_valid_celsius(-300))      # False
  ```
- 名前空間を汚さず、`Temperature.celsius_to_fahrenheit(...)` のように「どのクラスに属する処理か」が明確になる

### 主な用途（よくある使い方）
#### 1. ユーティリティ・ヘルパー関数
- クラスに関連するが、インスタンスやクラスの状態に依存しない補助的な処理
  ```python
  class StringUtils:
      @staticmethod
      def is_palindrome(s):
          return s == s[::-1]

      @staticmethod
      def reverse(s):
          return s[::-1]

  print(StringUtils.is_palindrome("たけやぶやけた"))  # True
  print(StringUtils.reverse("hello"))                 # olleh
  ```

#### 2. バリデーション（入力チェック）
- インスタンスを生成する前の入力値検証など
  ```python
  class User:
      def __init__(self, email):
          if not self.is_valid_email(email):
              raise ValueError(f"不正なメールアドレス: {email}")
          self.email = email

      @staticmethod
      def is_valid_email(email):
          return "@" in email and "." in email

  u = User("test@example.com")
  print(u.email)  # test@example.com
  # User("invalid")  → ValueError
  ```

#### 3. 内部的な計算ロジックの切り出し
- インスタンスメソッドの中から呼ばれる、状態に依存しない純粋な計算処理を分離する
  ```python
  class Order:
      TAX_RATE = 0.1

      def __init__(self, price):
          self.price = price

      def total(self):
          # 状態(self.price)を渡して純粋計算に委譲
          return self.calc_with_tax(self.price, self.TAX_RATE)

      @staticmethod
      def calc_with_tax(price, tax_rate):
          return int(price * (1 + tax_rate))

  o = Order(1000)
  print(o.total())  # 1100
  ```

## classmethod・インスタンスメソッドとの違い
| 種類 | デコレータ | 第一引数 | インスタンス変数 | クラス変数 | 主な用途 |
| --- | --- | --- | --- | --- | --- |
| インスタンスメソッド | なし | `self` | ○ | ○ | インスタンスの状態を使う処理 |
| クラスメソッド | `@classmethod` | `cls` | ✕ | ○ | ファクトリメソッド、クラス状態の管理 |
| スタティックメソッド | `@staticmethod` | なし | ✕ | ✕ | クラスに関連するが状態に依存しない処理 |

> [!IMPORTANT]
> ### 3種類のメソッドの比較（コードで確認）
> ```python
> class Sample:
>     class_var = "クラス変数"
>
>     def __init__(self):
>         self.instance_var = "インスタンス変数"
>
>     # インスタンスメソッド: self を受け取る
>     def instance_method(self):
>         return f"self経由: {self.instance_var}, {self.class_var}"
>
>     # クラスメソッド: cls を受け取る
>     @classmethod
>     def class_method(cls):
>         return f"cls経由: {cls.class_var}"  # インスタンス変数は不可
>
>     # スタティックメソッド: 何も受け取らない
>     @staticmethod
>     def static_method():
>         return "self も cls も受け取らない（引数は自分で定義したものだけ）"
>
> s = Sample()
> print(s.instance_method())  # self経由: インスタンス変数, クラス変数
> print(Sample.class_method())  # cls経由: クラス変数
> print(Sample.static_method())  # self も cls も受け取らない...
> ```

## 継承時の挙動
- `classmethod`は`cls`がサブクラスに置き換わるため、サブクラスからの呼び出しでサブクラス自身を参照できる
- 一方`staticmethod`は`cls`を受け取らないため、**呼び出し元のクラスを知る手段がない**
  ```python
  class Base:
      name = "Base"

      @classmethod
      def who_class(cls):
          return cls.name  # 呼び出したクラスに追従する

      @staticmethod
      def who_static():
          return Base.name  # 常にBaseを直接参照してしまう

  class Child(Base):
      name = "Child"

  print(Child.who_class())   # Child （clsがChildになる）
  print(Child.who_static())  # Base  （Baseを直書きしているため追従しない）
  ```

> [!TIP]
> サブクラスで挙動を変えたい（多態性が必要）なら`classmethod`、
> 純粋にクラスの状態と無関係な処理なら`staticmethod`を選ぶ。

## いつstaticmethodを使うべきか（判断基準）
- **`self`を使わない** → インスタンスメソッドである必要はない
- **`cls`（クラス変数・クラス状態）も使わない** → `classmethod`である必要もない
- **そのクラスに論理的に属する処理** → モジュール関数ではなくクラス内に置く価値がある

  上記3つを満たすときが`staticmethod`の出番。

> [!NOTE]
> `self`も`cls`も使わないインスタンスメソッドを書いていると、linter（例: [[ruff（linter）について]]）やIDEが「`staticmethod`にできる」と警告することがある（Ruffの`PLR6301`など）。

## 関連
- [[classmethodについて]]
- [[propertyについて]]
- [[Decoratorsについて]]
- [[Classについて]]
- [[_から始まる関数について]]
