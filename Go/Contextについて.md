## Context
- https://pkg.go.dev/context  
  > **Overview** ¶
Package context defines the Context type, which carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.
  >
  > Incoming requests to a server should create a Context, and outgoing calls to servers should accept a Context. The chain of function calls between them must propagate the Context, optionally replacing it with a derived Context created using WithCancel, WithDeadline, WithTimeout, or WithValue. When a Context is canceled, all Contexts derived from it are also canceled.
  >
  > The WithCancel, WithDeadline, and WithTimeout functions take a Context (the parent) and return a derived Context (the child) and a CancelFunc. Calling the CancelFunc cancels the child and its children, removes the parent's reference to the child, and stops any associated timers. Failing to call the CancelFunc leaks the child and its children until the parent is canceled or the timer fires. The go vet tool checks that CancelFuncs are used on all control-flow paths.
  >
  > The WithCancelCause function returns a CancelCauseFunc, which takes an error and records it as the cancellation cause. Calling Cause on the canceled context or any of its children retrieves the cause. If no cause is specified, Cause(ctx) returns the same value as ctx.Err().
  >
  > Programs that use Contexts should follow these rules to keep interfaces consistent across packages and enable static analysis tools to check context propagation:
  >
  > Do not store Contexts inside a struct type; instead, pass a Context explicitly to each function that needs it. The Context should be the first parameter, typically named ctx:
  >
  > ~~~
  > func DoSomething(ctx context.Context, arg Arg) error {
	>   // ... use ctx ...
  > }
  > ~~~
  > **Do not pass a nil Context, even if a function permits it. Pass `context.TODO` if you are unsure about which Context to use.**
  >
  > Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions.
  >
  > The same Context may be passed to functions running in different goroutines; Contexts are safe for simultaneous use by multiple goroutines.
  >
  > See https://blog.golang.org/context for example code for a server that uses Contexts.
- (一連の)Deadline(`WithDeadline`)やTimeout(`WithTimeout`)を設定してそれが過ぎたら親Goroutineとそこから派生したすべてのGoroutineをcancelしてリソースleakを防いだり、`WithCancel`でGoroutineをcancelするタイミングを制御したり、`WithValue`で1つのrequestで関連する複数のGoroutine間で値を連携したりするために使われる
  > In Go servers, each incoming request is handled in its own goroutine. Request handlers often start additional goroutines to access backends such as databases and RPC services. The set of goroutines working on a request typically needs access to request-specific values such as the identity of the end user, authorization tokens, and the request’s deadline. When a request is canceled or times out, all the goroutines working on that request should exit quickly so the system can reclaim any resources they are using.
  > 
  > At Google, we developed a context package that makes it easy to **pass request-scoped values, cancellation signals, and deadlines across API boundaries to all the goroutines involved in handling a request.**
  - 値の連携はContext interfaceの`Value`メソッドで行われる
- Context interface
  - https://pkg.go.dev/context#Context  
    ~~~go
    type Context interface {
    	// Deadline returns the time when work done on behalf of this context
    	// should be canceled. Deadline returns ok==false when no deadline is
    	// set. Successive calls to Deadline return the same results.
    	Deadline() (deadline time.Time, ok bool)

    	// Done returns a channel that's closed when work done on behalf of this
    	// context should be canceled. Done may return nil if this context can
    	// never be canceled. Successive calls to Done return the same value.
    	// The close of the Done channel may happen asynchronously,
    	// after the cancel function returns.
    	//
    	// WithCancel arranges for Done to be closed when cancel is called;
    	// WithDeadline arranges for Done to be closed when the deadline
    	// expires; WithTimeout arranges for Done to be closed when the timeout
    	// elapses.
    	//
    	// Done is provided for use in select statements:
    	//
    	//  // Stream generates values with DoSomething and sends them to out
    	//  // until DoSomething returns an error or ctx.Done is closed.
    	//  func Stream(ctx context.Context, out chan<- Value) error {
    	//  	for {
    	//  		v, err := DoSomething(ctx)
    	//  		if err != nil {
    	//  			return err
    	//  		}
    	//  		select {
    	//  		case <-ctx.Done():
    	//  			return ctx.Err()
    	//  		case out <- v:
    	//  		}
    	//  	}
    	//  }
    	//
    	// See https://blog.golang.org/pipelines for more examples of how to use
    	// a Done channel for cancellation.
    	Done() <-chan struct{}

    	// If Done is not yet closed, Err returns nil.
    	// If Done is closed, Err returns a non-nil error explaining why:
    	// Canceled if the context was canceled
    	// or DeadlineExceeded if the context's deadline passed.
    	// After Err returns a non-nil error, successive calls to Err return the same error.
    	Err() error

    	// Value returns the value associated with this context for key, or nil
    	// if no value is associated with key. Successive calls to Value with
    	// the same key returns the same result.
    	//
    	// Use context values only for request-scoped data that transits
    	// processes and API boundaries, not for passing optional parameters to
    	// functions.
    	//
    	// A key identifies a specific value in a Context. Functions that wish
    	// to store values in Context typically allocate a key in a global
    	// variable then use that key as the argument to context.WithValue and
    	// Context.Value. A key can be any type that supports equality;
    	// packages should define keys as an unexported type to avoid
    	// collisions.
    	//
    	// Packages that define a Context key should provide type-safe accessors
    	// for the values stored using that key:
    	//
    	// 	// Package user defines a User type that's stored in Contexts.
    	// 	package user
    	//
    	// 	import "context"
    	//
    	// 	// User is the type of value stored in the Contexts.
    	// 	type User struct {...}
    	//
    	// 	// key is an unexported type for keys defined in this package.
    	// 	// This prevents collisions with keys defined in other packages.
    	// 	type key int
    	//
    	// 	// userKey is the key for user.User values in Contexts. It is
    	// 	// unexported; clients use user.NewContext and user.FromContext
    	// 	// instead of using this key directly.
    	// 	var userKey key
    	//
    	// 	// NewContext returns a new Context that carries value u.
    	// 	func NewContext(ctx context.Context, u *User) context.Context {
    	// 		return context.WithValue(ctx, userKey, u)
    	// 	}
    	//
    	// 	// FromContext returns the User value stored in ctx, if any.
    	// 	func FromContext(ctx context.Context) (*User, bool) {
    	// 		u, ok := ctx.Value(userKey).(*User)
    	// 		return u, ok
    	// 	}
    	Value(key any) any
    }
    ~~~
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

- Chat-GPTのContextに関する回答  
  > contextは、API境界やプロセス間でのデータのキャンセル信号、期限、その他のリクエストスコープの値を伝達するためのパッケージです。主に、次のような用途に使用されます：
  > 
  > 1. **キャンセル信号の伝達**: contextは、長時間実行される作業や、ユーザーのリクエストに基づいている作業（HTTPリクエストなど）をキャンセルするための仕組みを提供します。例えば、ユーザーがブラウザでページを閉じた場合、関連するバックエンドの処理をキャンセルすることができます。
  > 
  > 2. **期限の設定**: contextを使用すると、特定の操作に対してタイムアウトを設定できます。これにより、リソースの無駄遣いを防ぐことができます。
  > 
  > 3. **値の伝達**: contextを使用して、リクエスト全体のライフサイクルにわたって値を伝達することができます。これは、リクエストIDや認証トークンなど、複数の関数やゴルーチン間で共有したい情報に役立ちます。主に、リクエストのライフサイクルやゴルーチン間で共有する必要があるメタデータや制御情報を伝達するために使用されます。これには、リクエストID、認証情報、ロギングやトレーシングのための情報などが含まれます。
  >
  > contextパッケージは、標準のcontext.Contextインターフェースを提供しており、Goの多くの標準ライブラリやサードパーティのライブラリで広くサポートされています。これにより、異なるライブラリやアプリケーションコンポーネント間での連携が容易になります。

### ■ `Background()`と`TODO()`の違い
- ２つとも、Contextの初期化に使用される関数
- BackgroundとTODOは、どちらも実際の操作（キャンセルや期限設定）を行わない空のContextだが、その意図が異なる。BackgroundはデフォルトのContextとして、TODOは将来的なContextの設定を予定している場所で使う。
##### `Background()`
- 処理の最上位(e.g. main関数、最上位レベルのGoroutine)で使用されることが多く、キャンセル信号や値を伝達するためのデフォルトのContextとして機能。 空のContextが返される。
  -  https://go.dev/blog/context
      > Background is the root of any Context tree; it is never canceled
- https://pkg.go.dev/context#Background  
  > **Background returns a non-nil, empty Context. It is never canceled, has no values, and has no deadline. It is typically used by the main function, initialization, and tests, and as the top-level Context for incoming requests.**
- https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/context/context.go;l=211  
  > ~~~go
  > func Background() Context {
	>   return backgroundCtx{}
  > }
  > ~~~

##### `TODO()`
- 将来的に適切なContextを設定する必要があるが、現時点ではどのContextを使用すべきか不明な場合に使用される。 Backgroundと同様に空のContextが返される。
- https://pkg.go.dev/context#TODO  
  > **TODO returns a non-nil, empty Context. Code should use context.TODO when it's unclear which Context to use or it is not yet available (because the surrounding function has not yet been extended to accept a Context parameter).**
- https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/context/context.go;l=211   
  > ~~~go
  > func TODO() Context {
	>   return todoCtx{}
  > }
  > ~~~