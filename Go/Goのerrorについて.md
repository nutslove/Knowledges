## `error`型の戻り値を返す方法
1. `errors.New()`を使う  
   ```go
   import "errors"

   func someFunction() error {
       // エラーが発生した場合
       return errors.New("エラーメッセージ")
   }
   ```
2. `fmt.Errorf()`を使う  
   ```go
   import "fmt"

   func someFunction() error {
       // エラーが発生した場合
       return fmt.Errorf("エラーメッセージ: %v", someValue)
   }

## `error`の型確認
- `error`の型を確認するためにはgolang標準ライブラリ`errors`の`errors.Is`関数を使う
- `errors.Is(err, target error)`の形で使い、`err`が `target`と同じエラーか、あるいは`err`がラップしているエラーが`target`と同じかどうかを確認
- 例  
  ```go
  package main

  import (
      "errors"
      "fmt"
      "gorm.io/gorm"
  )

  func main() {
      err := gorm.ErrRecordNotFound

      if errors.Is(err, gorm.ErrRecordNotFound) {
          fmt.Println("Record not found error detected")
      } else {
          fmt.Println("Some other error")
      }
  }
  ```

## `chan error`について
- `error`型のzero valueは`nil`  
  なのでcloseされた`error`型のchannelから値を取り出そうとすると`nil`が返ってくる  
  ~~~go
  func main() {
    ch := make(chan error)
    close(ch)

    val, ok := <-ch
    fmt.Println(val, ok) // 出力: <nil> false
  }
  ~~~

