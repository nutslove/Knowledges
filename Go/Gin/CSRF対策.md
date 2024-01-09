## CSRFとは
- https://www.lanscope.jp/blogs/cyber_attack_pfs_blog/20231020_15621/
- session管理の脆弱性を悪用するサイバー攻撃の1つ

## Go(Gin)のサーバ側での対策
- `github.com/utrack/gin-csrf`と`gin-contrib/sessions`パッケージを使用
- 実装例
  ~~~go
	// Cookieベースのセッションを設定
	secretKey := os.Getenv("SESSION_SECRET_KEY") // Sessionの暗号化キーは固定の値を使用することで、アプリの再起動時にセッションが維持されるようにする
	if secretKey == "" {
		fmt.Println("SESSION_SECRET_KEY環境変数が設定されていません")
		return
	}

	store := cookie.NewStore([]byte(secretKey))
	router.Use(sessions.Sessions("session", store)) // ブラウザのCookieにセッションIDを保存する

	// CSRFミドルウェアの設定
	// HTML内の_csrfの値を取得して、リクエストトークンと比較を行い、一致しない場合ErrorFuncを実行する（https://github.com/utrack/gin-csrf/blob/master/csrf.go）
	router.Use(csrf.Middleware(csrf.Options{
		Secret: secretKey, // 上のCookieベースのセッションと同じ値を指定
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))
  ~~~

## CSRF隠しフィールドトークンのメカニズム
1. **トークンの生成とセッションの紐付け:** セッションが開始される際（たとえばユーザーがログインページにアクセスした際）、サーバーは一意のCSRFトークンを生成し、ユーザーのセッションに保存します。このトークンは、ユーザーがフォームを通じてデータを送信する際にサーバーに送り返される必要があります。

2. **フォームへのトークンの埋め込み:** サーバーは生成したCSRFトークンをフォームの隠しフィールドに埋め込みます。これにより、ユーザーがフォームを送信する際に、トークンも一緒に送信されます。

3. **トークンの検証:** フォームがサーバーに送信された際、サーバーはフォームに含まれるCSRFトークンをセッションに保存されているトークンと照合します。トークンが一致しない場合、リクエストは拒否されます。

4. **偽装サイトからの攻撃の防止:** もし攻撃者がユーザーを偽装して不正なPOSTリクエストを送ろうとしても、攻撃者のサイトではユーザーのセッション情報を知ることができません。したがって、正しいCSRFトークンをフォームに含めることができず、サーバーはこのリクエストを拒否します。

このように、CSRFトークンはユーザーのセッションに固有のものであり、偽装サイトではこのトークンを知ることができないため、不正なリクエストを効果的に防ぐことができます。