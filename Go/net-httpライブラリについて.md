- URL
  https://pkg.go.dev/net/http
- Goの基本的なhttp client/server用ライブラリ

### ■ `Handler`とは
- HTTP リクエストを処理するための関数または構造体のこと。`net/http`パッケージでは、`http.Handler`インターフェースを定義している。
- `ServeHTTP`メソッドはHTTPリクエストを受け取り、適切な処理を行い、HTTPレスポンスを返す役割を持つ
  - 第1引数`ResponseWriter`は、HTTP レスポンスを書き込むためのインターフェース
  - 第2引数`*Request`は、受信した HTTP リクエストを表す構造体へのポインタ
~~~go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
~~~
> A Handler responds to an HTTP request.
- https://pkg.go.dev/net/http#Handler

### ■ `Handle`関数と`HandleFunc`関数について
- `http.Handle`と`http.HandleFunc`はどちらもハンドラーを登録するための関数。  
  ただ、それぞれ異なる方法でハンドラーを登録する。
- `http.Handle`も`http.HandleFunc`も、第１引数にリクエストを待ち受けるURLパスを指定し、第２引数にリクエストの処理を指定するというのは基本的に同じ
#### ▲`Handle`関数
- `http.Handle`関数は、`http.Handler`インターフェースを実装した型を受け取る。
- > Handle registers the handler for the given pattern in the DefaultServeMux. The documentation for ServeMux explains how patterns are matched.
  → ここでいう**patternとはURLのこと(`/metrics`等)**

- Format（Signature）
  ~~~go
  func Handle(pattern string, handler Handler)
  ~~~
  - 第1引数`pattern`は、ハンドラーを登録するパターン（URL パス）を指定
  - 第2引数`handler`は、`http.Handler`インターフェースを実装した型を指定
- 例
  ~~~go
  http.Handle("/metrics", promhttp.Handler())
  ~~~
  - `promhttp.Handler()`は戻り値が`http.Handler`  
    ~~~go
    func Handler() http.Handler {
        return InstrumentMetricHandler(
            prometheus.DefaultRegisterer, HandlerFor(prometheus.DefaultGatherer, HandlerOpts{}),
        )
    }
    ~~~
#### ▲`HandleFunc`関数
- 通常の関数をハンドラーとして登録するために使用
- Format（Signature）
  ~~~go
  func HandleFunc(pattern string, handler func(ResponseWriter, *Request))
  ~~~
  - 第1引数`pattern`は、ハンドラーを登録するパターン（URL パス）を指定
  - 第2引数`handler`は、`func(ResponseWriter, *Request)`型の関数を指定
- 例
  ~~~go
  func myHandler(w http.ResponseWriter, r *http.Request) {
      // ハンドラーの処理を記述する
  }

  http.HandleFunc("/custom", myHandler)
  ~~~

### ■ `ServeMux`とは
~~~go
type ServeMux struct {
	// contains filtered or unexported fields
}
~~~
> ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL.

※*multiplexer*： 多重装置、多重化装置 -→ ふたつ以上の入力をひとつの信号として出力する機構
> DefaultServeMux is the default ServeMux used by Serve. 
- https://pkg.go.dev/net/http#ServeMux

### ■ `ListenAndServe`functionについて
> ListenAndServe listens on the TCP network address addr and then calls Serve with handler to handle requests on incoming connections. Accepted connections are configured to enable TCP keep-alives.
The handler is typically nil, in which case the DefaultServeMux is used.
ListenAndServe always returns a non-nil error.

> ListenAndServe starts an HTTP server with a given address and handler. The handler is usually nil, which means to use DefaultServeMux. Handle and HandleFunc add handlers to DefaultServeMux:
- Format
  ~~~go
  func ListenAndServe(addr string, handler Handler) error
  ~~~
- 例
  ~~~go
  log.Fatal(http.ListenAndServe(":8080", nil))
  ~~~
- 参考URL
  - https://pkg.go.dev/net/http#ListenAndServe
  - https://pkg.go.dev/net/http#pkg-overview
