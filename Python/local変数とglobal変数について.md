- 関数の中にあるものだけがlocal変数として扱われる
  - if文やfor/while文の中の変数はlocal変数として扱われない
- **global変数の値を関数の中で修正しないこと！(可読性のため)**
  - **常数**(値が固定で変えないもの)は**大文字**で定義する！
    - e.g. `URL = ***`

- global変数は基本、関数の中で使えないけど、以下では使える
  1. `global`でglobal変数を読み込む
       - 例
         ~~~python
         enemies = 1

         def increase_enemies():
           global enemies
           enemies = 2 --> この時点で最初に宣言した`enemies`変数の値が2に変わる
           print(f"enemies inside function: {enemies}")

         increase_enemies()
         print(f"enemies outside function: {enemies}")
         ~~~  
         → 両方とも2が出力される
  2. `return`で指定
       - 例
         ~~~python
         enemies = 1

         def increase_enemies():
           print(f"enemies inside function: {enemies}")
           return enemies + 1

         new_enemies = increase_enemies()
         print(f"enemies outside function: {enemies}")
         print(new_enemies)
         ~~~  
         → 1, 1, 2が出力される
