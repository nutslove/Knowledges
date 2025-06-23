## `recover()`とは
- Goの組み込み関数
- **`defer`で呼び出された関数内でのみ有効**
- panic状態からの回復を試みる
- **panicが発生したGoroutineと同じGoroutine内で呼び出す必要がある**
- 例:
  ```go
  func safeFunction() {
      defer func() {
          if r := recover(); r != nil {
              fmt.Println("Recovered from:", r)
          }
      }()
      // panicを引き起こすコード
      panic("Something went wrong")
  }

  func main() {
      safeFunction()
      fmt.Println("Program continues after recovery")
  }
  ```  
  - 上記の例では、`safeFunction`内でpanicが発生しても、`recover()`によって回復され、プログラムは続行される（以下が出力される）  
    ```shell
    Recovered from: Something went wrong
    Program continues after recovery
    ```