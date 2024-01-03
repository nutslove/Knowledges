## Context
- (一連の)Deadline(`WithDeadline`)やTimeout(`WithTimeout`)を設定してそれが過ぎたら親Goroutineとそこ派生したすべてのGoroutineをcancelしてリソースleakを防いだり、`WithCancel`でGoroutineをcancelするタイミングを制御したり、`WithValue`で1つのrequestで関連する複数のGoroutine間で値をpassしたりするために使われる
  > In Go servers, each incoming request is handled in its own goroutine. Request handlers often start additional goroutines to access backends such as databases and RPC services. The set of goroutines working on a request typically needs access to request-specific values such as the identity of the end user, authorization tokens, and the request’s deadline. When a request is canceled or times out, all the goroutines working on that request should exit quickly so the system can reclaim any resources they are using.
  > 
  > At Google, we developed a context package that makes it easy to **pass request-scoped values, cancellation signals, and deadlines across API boundaries to all the goroutines involved in handling a request.**
- Contextがcancelされると、そのGoroutineから派生したすべてのGoroutineがcancelされる
  > When a Context is canceled, all Contexts derived from it are also canceled.
- 以下4つのfunctionがある
  1. `WithCancel` -> [例](https://pkg.go.dev/context#example-WithCancel)
  2. `WithDeadline` -> [例](https://pkg.go.dev/context#example-WithDeadline)
  3. `WithTimeout` -> [例](https://pkg.go.dev/context#example-WithTimeout)
  4. `WithValue` -> [例](https://pkg.go.dev/context#example-WithValue)
- `Done()`は`select`文の中で使う必要がある
  ~~~go
  package main

  import (
      "context"
      "fmt"
      "runtime"
      "time"
  )

  func main() {
      ctx, cancel := context.WithCancel(context.Background())

      fmt.Println("error check 1:", ctx.Err())
      fmt.Println("num gortins 1:", runtime.NumGoroutine())

      go func() {
          n := 0
          for {
              select {
              case <-ctx.Done():
                  return // go func()を抜ける
              default:
                  n++
                  time.Sleep(time.Millisecond * 200)
                  fmt.Println("working", n)
              }
          }
      }()

      time.Sleep(time.Second * 2)
      fmt.Println("error check 2:", ctx.Err())
      fmt.Println("num gortins 2:", runtime.NumGoroutine())

      fmt.Println("about to cancel context")
      cancel()
      fmt.Println("cancelled context")

      time.Sleep(time.Second * 2)
      fmt.Println("error check 3:", ctx.Err())
      fmt.Println("num gortins 3:", runtime.NumGoroutine())
  }
  ~~~
- 参考URL
  - **https://go.dev/blog/context**
  - **https://peter.bourgon.org/blog/2016/07/11/context.html**
  - https://pkg.go.dev/context#pkg-overview
  - https://zenn.dev/hsaki/books/golang-context/viewer/definition
