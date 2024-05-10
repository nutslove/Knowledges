## CSRFとは
- https://www.lanscope.jp/blogs/cyber_attack_pfs_blog/20231020_15621/
- session管理の脆弱性を悪用するサイバー攻撃の1つ

## Go(Gin)のサーバ側での対策
- `github.com/utrack/gin-csrf`と`gin-contrib/sessions`パッケージを使用
- 実装例
  ~~~go
  import (
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    csrf "github.com/utrack/gin-csrf"
  )

  router = gin.Default()
  // Cookieベースのセッションを設定
  secretKey := os.Getenv("SESSION_SECRET_KEY") // Sessionの暗号化キーは固定の値を使用することで、アプリの再起動時にセッションが維持されるようにする
  if secretKey == "" {
  fmt.Println("SESSION_SECRET_KEY環境変数が設定されていません")
  return
  }

  store := cookie.NewStore([]byte(secretKey))
  router.Use(sessions.Sessions("session", store)) // ブラウザのCookieにセッションIDを保存する

  secretKey_for_csrf := "csrfSecretKey"

  // CSRFミドルウェアの設定
  // HTML内の_csrfの値を取得して、リクエストトークンと比較を行い、一致しない場合ErrorFuncを実行する（https://github.com/utrack/gin-csrf/blob/master/csrf.go）
  router.Use(csrf.Middleware(csrf.Options{
    Secret: secretKey_for_csrf, // CSRFトークンの生成に使用される秘密鍵
    ErrorFunc: func(c *gin.Context) {
    	c.String(400, "CSRF token mismatch")
    	c.Abort()
    },
  }))

	router.GET("/", func(c *gin.Context) {
		token := csrf.GetToken(c)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"head_title": "Golang-Gin",
			"title":      "Main website",
			"content":    "OpenTelemetry with Golang-Gin",
			"csrfToken":  token,
		})
	})
  ~~~

- CSRFトークンの生成に使用される秘密鍵(上記の場合`secretKey_for_csrf`)は固定値でも、saltで毎回生成されるトークンは異なる

## CSRFトークンの生成と検証の流れ
1. ユーザーがGETメソッドでフォームを表示するページにアクセスする。
2. サーバー側では、**CSRFミドルウェアがリクエストごとに一意のCSRFトークンを生成する。**
3. 生成されたCSRFトークンは、レスポンスのHTMLに隠しフィールド（例えば`<input type="hidden" name="_csrf" value="生成されたトークン">`）として埋め込まれて、クライアントに返される。
4. ユーザーがフォームを送信すると、POSTメソッドでサーバーにリクエストが送信される。このとき、隠しフィールド(`name="_csrf"`)に埋め込まれたCSRFトークンも一緒にサーバーに送信される。
5. サーバー側では、CSRFミドルウェアがPOSTリクエストを受信し、送信されたCSRFトークンを検証する。
6. 送信されたCSRFトークンが、サーバーが発行したトークンと一致していれば、リクエストは有効とみなされ、処理が続行される。
7. トークンが一致しない場合は、CSRFの可能性があるため、エラーハンドラが呼び出され、適切なエラーレスポンスが返される。