- URL
  https://pkg.go.dev/net/http
- Goの基本的なhttp client/server用ライブラリ

### `ListenAndServe`
 > ListenAndServe starts an HTTP server with a given address and handler. The handler is usually nil, which means to use DefaultServeMux. Handle and HandleFunc add handlers to DefaultServeMux:
 ~~~go
 log.Fatal(http.ListenAndServe(":8080", nil))
 ~~~

