## HTML rendering
- https://gin-gonic.com/docs/examples/html-rendering/
- goファイルがあるところに`templates`ディレクトリを作成し、その配下に`.tmpl`拡張子でHTMLテンプレートファイルを作成する
- 例  
  - goファイル
    ~~~go
    func main() {
    	router := gin.Default()
    	router.LoadHTMLGlob("templates/*")
    	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
    	router.GET("/index", func(c *gin.Context) {
    		c.HTML(http.StatusOK, "index.tmpl", gin.H{
    			"title": "Main website",
    		})
    	})
    	router.Run(":8080")
    }
    ~~~
  - templates/index.tmpl
    ~~~html
    <html>
    	<h1>
    		{{ .title }}
    	</h1>
    </html>
    ~~~

## 使い方
~~~go
import (
	github.com/gin-gonic/gin
)

func main() {
	router = gin.Default()
}
~~~
### `gin.Default()`
- default routerを初期化
  - このrouterがHTTPリクエストを処理し、適切なHandler関数にルーティングする役割
- 大元 (https://github.com/gin-gonic/gin/blob/master/gin.go)
	~~~go
  // Default returns an Engine instance with the Logger and Recovery middleware already attached.
  func Default() *Engine {
  		debugPrintWARNINGDefault()
  		engine := New()
  		engine.Use(Logger(), Recovery())
  		return engine
  }
	~~~
  - `debugPrintWARNINGDefault()` (https://github.com/gin-gonic/gin/blob/master/debug.go#L68)
  	~~~go
    func debugPrintWARNINGDefault() {
    	if v, e := getMinVer(runtime.Version()); e == nil && v < ginSupportMinGoVer {
    		debugPrint(`[WARNING] Now Gin requires Go 1.18+.

    `)
    	}
    	debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

    `)
    }
  	~~~
  - `New()` (https://github.com/gin-gonic/gin/blob/master/gin.go#L183C1-L213C2)
  	~~~go
    func New() *Engine {
    	debugPrintWARNINGNew()
    	engine := &Engine{
    		RouterGroup: RouterGroup{
    			Handlers: nil,
    			basePath: "/",
    			root:     true,
    		},
    		FuncMap:                template.FuncMap{},
    		RedirectTrailingSlash:  true,
    		RedirectFixedPath:      false,
    		HandleMethodNotAllowed: false,
    		ForwardedByClientIP:    true,
    		RemoteIPHeaders:        []string{"X-Forwarded-For", "X-Real-IP"},
    		TrustedPlatform:        defaultPlatform,
    		UseRawPath:             false,
    		RemoveExtraSlash:       false,
    		UnescapePathValues:     true,
    		MaxMultipartMemory:     defaultMultipartMemory,
    		trees:                  make(methodTrees, 0, 9),
    		delims:                 render.Delims{Left: "{{", Right: "}}"},
    		secureJSONPrefix:       "while(1);",
    		trustedProxies:         []string{"0.0.0.0/0", "::/0"},
    		trustedCIDRs:           defaultTrustedCIDRs,
    	}
    	engine.RouterGroup.engine = engine
    	engine.pool.New = func() any {
    		return engine.allocateContext(engine.maxParams)
    	}
    	return engine
    }
		~~~
### `*gin.Context`
- リクエスト情報の取得やレスポンス設定に使用される
- `gin.Context`は、GinフレームワークでHTTPリクエストとレスポンスを処理する際に中心となる概念。`gin.Context`は、リクエストの詳細情報を保持し、レスポンスを生成するためのメソッドを提供するオブジェクト。このコンテキストを通じて、リクエストのパラメータやヘッダ、ボディなどのデータにアクセスしたり、レスポンスのステータスコードやヘッダ、ボディを設定することができる。
- 以下のようなユースケースなどにも使える
  - **コンテキスト変数**: アプリケーションの実行中にハンドラ間でデータを共有するために、コンテキストに変数を設定・取得する機能。
  - **セッション管理**: ユーザーセッションの管理に関連するデータと機能。

  #### `gin.Context`が保持するデータ
  - `gin.Context`には、以下のようなリクエストに関連する多くの情報が含まれている：
    - **パラメータ**: URLパスやクエリパラメータからの値を取得することができる。
    - **ヘッダー**: HTTPリクエストヘッダーの値にアクセスすることが可能。
    - **ボディ**: POSTやPUTリクエストのボディからデータを読み取ることができる。
    - **クッキー**: HTTPクッキーの値にアクセスすることができる。

  #### `gin.Context`の使い方
  - `gin.Context`を使用することで、開発者はリクエストに対する応答を柔軟に制御することができる。以下に、その使い方の例をいくつか挙げる：
    - **パラメータの取得**: URLパスやクエリからパラメータを取得することができる。
      ```go
      func(c *gin.Context) {
          id := c.Param("id") // URLパスから:idに対応するパラメータを取得
          query := c.Query("query") // クエリパラメータからqueryの値を取得
      }
      ```
    - **レスポンスの送信**: ステータスコード、ヘッダ、ボディを含むHTTPレスポンスを送信することができる。
      ```go
      func(c *gin.Context) {
          c.JSON(200, gin.H{"message": "hello world"}) // JSONレスポンスを送信
      }
      ```
    - **リクエストボディの読み取り**: JSONやフォームデータなど、リクエストボディからのデータを解析する。
      ```go
      func(c *gin.Context) {
          var jsonBody SomeStruct
          if err := c.ShouldBindJSON(&jsonBody); err != nil {
              c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
              return
          }
      }
      ```
    - **ミドルウェアとの連携**: ミドルウェア内で`gin.Context`を操作し、リクエストの前処理や後処理を行うことができる。
      ```go
      func MyMiddleware(c *gin.Context) {
          // リクエストの前処理
          c.Set("example", "12345") // コンテキストにデータを設定

          c.Next() // 次のハンドラやミドルウェアを呼び出し

          // リクエストの後処理
      }
      ```

## `Use`メソッド
Ginフレームワークにおける`Use`メソッドは、グローバルまたはルートレベルのミドルウェアを登録するために使用されます。ミドルウェアは、HTTPリクエストの処理中に特定の機能（ログ記録、認証、エラーハンドリングなど）を実行するための関数です。

`Use`メソッドの動作は以下のステップで行われます：

1. **ミドルウェアの登録：** `Use`メソッドを呼び出すことで、指定されたミドルウェアがGinエンジンに登録されます。

2. **リクエストの処理：** HTTPリクエストがサーバーに到着すると、登録されたミドルウェアは順番に実行されます。各ミドルウェアは`gin.Context`オブジェクトを受け取り、必要に応じてリクエストデータの処理やレスポンスデータの変更を行います。

3. **次のミドルウェアへの移行：** ミドルウェアは通常、`c.Next()`を呼び出して、次のミドルウェアまたは最終的なハンドラへの実行を続けます。`c.Abort()`を呼び出すと、チェーンの残りのミドルウェアの実行が停止されます。

4. **レスポンスの送信：** 最終的なハンドラがレスポンスを生成し、クライアントに送信されます。この時点で、登録されたミドルウェアの`c.Next()`の後の部分が実行される場合があります（例えば、レスポンス後のログ記録など）。

`Use`メソッドを使用することで、Ginアプリケーションは柔軟にミドルウェアを活用し、リクエストの前後でさまざまな処理を行うことができます。

### `c.Next()`について
- ミドルウェア内で現在のリクエストに対する処理を一時停止し、次のミドルウェアまたはルートハンドラに制御を移行するために使う
- `c.Next()`の必要性
	- **複数のミドルウェアがある場合**： 複数のミドルウェアを使用する場合、各ミドルウェア内で`c.Next()`を呼び出すことで、次のミドルウェアへの処理が移行します。`c.Next()`を呼ばない場合、チェーン内の後続のミドルウェアやルートハンドラは実行されません。
	- **ミドルウェアが最後の場合**： チェーン内の最後のミドルウェアである場合、そのミドルウェア内で`c.Next()`を呼び出す必要はありません。なぜなら、それ以上実行するミドルウェアやハンドラがないためです。
	- **外部ライブラリ/自作ミドルウェアの場合**： 外部ライブラリによって提供されるミドルウェアの場合、通常は`c.Next()`が内部的に適切に呼び出されていることが多いです。そのため、これを直接呼び出す場合に自分で`c.Next()`を追加する必要はありません。自作のミドルウェアの場合は`c.Next()`を適切な位置に配置することが重要。
- チェーン内の最後のミドルウェアであっても、`c.Next()`を含めても問題はありません。実際、`c.Next()`を含めることは一般的な慣習であり、ミドルウェアの一貫性を保つために推奨されることも多いです。

#### `c.Abort()`について
- `c.Abort()`が呼ばれた場合は、それ以降の`c.Next()`は実行されず、チェーン内の後続のミドルウェアやリクエストハンドラは実行されない点に注意が必要です。

## ミドルウェア(middleware)について
- ミドルウェアはリクエストごとに順番に実行される
  - 同じページでのリロード（F5を押すなど）や、異なるページに移動するなど、サーバーに送信される各HTTPリクエストに対して、登録されたミドルウェアが順番に実行される

#### middlewareとHandlerの関係
- https://zenn.dev/villa_ak99/articles/88998d20f512bc

##### ミドルウェアの動作の概要
- **リクエスト毎の実行：** ユーザーがブラウザでページをリロードしたり、新しいページに移動したりすると、新しいHTTPリクエストがサーバーに送信されます。サーバーがこのリクエストを受け取ると、設定されたミドルウェアが実行されます。
- **処理の流れ：** ミドルウェアは、登録された順序に従って処理されます。各ミドルウェアはリクエストを受け取り、必要に応じてリクエストの内容を変更したり、レスポンスを生成したりすることができます。
- **用途の多様性：** ミドルウェアは様々な目的で使用されます。例えば、認証チェック、ログ記録、リクエストのバリデーション、セキュリティヘッダーの追加、エラーハンドリングなどです。