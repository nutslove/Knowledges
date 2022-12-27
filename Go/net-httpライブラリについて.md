- URL
  https://pkg.go.dev/net/http
- Goの基本的なhttp client/server用ライブラリ

### ■ `Handler`とは
~~~go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
~~~
> A Handler responds to an HTTP request.
- https://pkg.go.dev/net/http#Handler

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

### ■ `Handle`functionについて
>Handle registers the handler for the given pattern in the DefaultServeMux. The documentation for ServeMux explains how patterns are matched.

→ ここでいうpatternとはURLのこと("/metrics"等)

- Format
  ~~~go
  func Handle(pattern string, handler Handler)
  ~~~
- 例
  ~~~go
  http.Handle("/metrics", promhttp.Handler())
  ~~~