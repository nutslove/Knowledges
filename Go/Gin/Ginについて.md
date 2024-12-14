# Ginについて
- [httprouter](https://github.com/julienschmidt/httprouter)という軽量で高速なHTTPリクエストルーティングライブラリを基盤として使用している
  - `httprouter`はGo標準の`net/http`をベースにしている

# 使い方
~~~go
import (
	github.com/gin-gonic/gin
)

func main() {
	router = gin.Default()
}
~~~

## `gin.Default()`
- default routerを初期化
  - このrouterがHTTPリクエストを処理し、適切なHandler関数にルーティングする役割
  - Handlerについては「net-httpライブラリについて.md」ファイルを参照
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

## `gin.New()`
- `gin.New()`は、Ginフレームワークの新しいインスタンスを作成するための関数。この関数は、カスタマイズ可能な空のGinルーターを返す。
- `gin.New()`関数は以下のようにシグネチャを持つ：  
  ```go
  func New() *Engine
  ```  
  この関数は、`*Engine`型のポインタを返す。**`Engine`はGinフレームワークのメインのルーター構造体で、HTTPリクエストのルーティングやミドルウェアの管理などを行う。**
- `gin.New()`で作成されたルーターには、デフォルトのミドルウェアは含まれていません。必要なミドルウェアを手動で追加する必要がある。  
  一方、`gin.Default()`関数を使用すると、よく使われるミドルウェア（ロギング、パニックリカバリー）があらかじめ設定された`*Engine`インスタンスを取得できる。
- 以下は、`gin.New()`を使用してカスタムルーターを作成し、ミドルウェアを追加する例：  
  ```go
  package main

  import (
      "github.com/gin-gonic/gin"
  )

  func main() {
      // 新しいGinルーターを作成
      router := gin.New()

      // ミドルウェアを追加
      router.Use(gin.Logger())
      router.Use(gin.Recovery())

      // ルートを定義
      router.GET("/", func(c *gin.Context) {
          c.JSON(200, gin.H{"message": "Hello, World!"})
      })

      // サーバーを起動
      router.Run(":8080")
  }
  ```  
  この例では、`gin.New()`で新しいルーターを作成し、`router.Use()`を使用して`Logger`と`Recovery`ミドルウェアを追加している。その後、ルートを定義し、サーバーを起動している。

- `gin.New()`を使用することで、必要なミドルウェアを柔軟に選択し、アプリケーションの要件に合わせてルーターをカスタマイズすることができる。ただし、ほとんどの場合、`gin.Default()`で提供されるデフォルトのミドルウェアで十分であり、特別な理由がない限り`gin.Default()`を使用することが推奨されている。

## ミドルウェア(middleware)とは
- リクエストの処理を行う前後に実行される関数。主にリクエストの前処理（認証、ロギング、トークン検証など）や後処理（レスポンスの加工など）を行う。ミドルウェアは通常、ハンドラーの前に実行され、リクエストを加工したり、特定の条件に応じてリクエストを拒否したりする。
- ミドルウェアはリクエストごとに順番に実行される
  - 同じページでのリロード（F5を押すなど）や、異なるページに移動するなど、サーバーに送信される各HTTPリクエストに対して、登録されたミドルウェアが順番に実行される

### middlewareとHandlerの関係
- https://zenn.dev/villa_ak99/articles/88998d20f512bc

### ミドルウェアの動作の概要
- **リクエスト毎の実行：** ユーザーがブラウザでページをリロードしたり、新しいページに移動したりすると、新しいHTTPリクエストがサーバーに送信され、サーバーがこのリクエストを受け取ると、設定されたミドルウェアが実行される。
- **処理の流れ：** ミドルウェアは、登録された順序に従って処理される。各ミドルウェアはリクエストを受け取り、必要に応じてリクエストの内容を変更したり、レスポンスを生成したりすることができる。
- **用途の多様性：** ミドルウェアは様々な目的で使用される。例えば、認証チェック、ログ記録、リクエストのバリデーション、セキュリティヘッダーの追加、エラーハンドリングなど。

### `Logger()`と`Recovery()`ミドルウェアについて
`Logger()`と`Recovery()`は、Ginフレームワークがデフォルトで提供している組み込みのミドルウェアです。

1. `Logger()`ミドルウェア：
   - 受信したHTTPリクエストのログを出力します。
   - リクエスト開始時間、レスポンスステータスコード、レイテンシ、クライアントIPアドレス、HTTPメソッド、パス、プロトコルバージョンなどの情報を含みます。
   - デバッグやパフォーマンスの監視に役立ちます。

2. `Recovery()`ミドルウェア：
   - Panic（ランタイムエラー）からの復帰を処理します。
   - アプリケーションがクラッシュすることを防ぎ、500 Internal Server Errorをクライアントに返します。
   - パニックが発生した場合、エラーメッセージとスタックトレースをログに出力します。

これらのミドルウェアは、`gin.Default()`関数を使用して作成されたルーターには自動的に含まれています。以下は、`gin.Default()`の定義です：

```go
func Default() *Engine {
    debugPrintWARNINGDefault()
    engine := New()
    engine.Use(Logger(), Recovery())
    return engine
}
```

この関数は、新しい`*Engine`インスタンスを作成し、`Logger()`と`Recovery()`ミドルウェアを追加してから、そのインスタンスを返します。

通常、これらのミドルウェアは、ほとんどのアプリケーションで必要とされるため、`gin.Default()`を使用することが推奨されています。ただし、特別な理由でこれらのミドルウェアを使用したくない場合は、`gin.New()`を使用して新しいルーターを作成し、必要なミドルウェアを手動で追加することもできます。

以上のように、`Logger()`と`Recovery()`は、Ginフレームワークが提供する便利な組み込みミドルウェアであり、アプリケーションの開発をシンプルかつ安全にするために役立ちます。

## `*gin.Context`
- リクエスト情報の取得やレスポンス設定に使用される
- `gin.Context`は、GinフレームワークでHTTPリクエストとレスポンスを処理する際に中心となる概念。`gin.Context`は、リクエストの詳細情報を保持し、レスポンスを生成するためのメソッドを提供するオブジェクト。このコンテキストを通じて、リクエストのパラメータやヘッダ、ボディなどのデータにアクセスしたり、レスポンスのステータスコードやヘッダ、ボディを設定することができる。
- 以下のようなユースケースなどにも使える
  - **コンテキスト変数**: アプリケーションの実行中にハンドラ間でデータを共有するために、コンテキストに変数を設定・取得する機能。
  - **セッション管理**: ユーザーセッションの管理に関連するデータと機能。

### `gin.Context`が保持するデータ（一部）
- `gin.Context`には、以下のようなリクエストに関連する多くの情報が含まれている：
  - **パラメータ**: URLパスやクエリパラメータからの値を取得することができる。
  - **ヘッダー**: HTTPリクエストヘッダーの値にアクセスすることが可能。
  - **ボディ**: POSTやPUTリクエストのボディからデータを読み取ることができる。
  - **クッキー**: HTTPクッキーの値にアクセスすることができる。

### `gin.Context`の使い方
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

### `*gin.Context.AbortWithStatusJSON`と`*gin.Context.JSON`の違い
- `AbortWithStatusJSON`と`JSON`両方とも、第１引数にHTTPステータスを第２引数に戻りのJSONを指定するのは一緒だけど、**`AbortWithStatusJSON`はリクエストの処理を中断し、その後のハンドラーやミドルウェアの処理実行されない**
- `AbortWithStatusJSON`の例  
  ```go
  // 認証ミドルウェア
  func authMiddleware() gin.HandlerFunc {
      return func(c *gin.Context) {
          token := c.GetHeader("X-Auth-Token")
          if token == "" {  // http.StatusUnauthorizedは401
              c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
              return
          }
          c.Next() // 次のハンドラーまたはミドルウェアを呼び出す
      }
  }
  ```

### `c.Next()`について
- ミドルウェア内で現在のリクエストに対する処理を一時停止し、次のミドルウェアまたはルートハンドラに制御を移行するために使う
- **以下のような後処理や条件に応じて処理を中断する必要がない場合、`c.Next()`は不要**
  - `c.Next()`を明示的に定義しなくても、次のミドルウェアやハンドラに処理が続く
#### `c.Next()`の用途
1. 後処理の実装
   - 例  
     ```go
     func Middleware1() gin.HandlerFunc {
         return func(c *gin.Context) {
             fmt.Println("M1 開始")  // 前処理1
             c.Next()
             fmt.Println("M1 終了")  // 後処理1
         }
     }

     func Middleware2() gin.HandlerFunc {
         return func(c *gin.Context) {
             fmt.Println("M2 開始")  // 前処理2
             c.Next()
             fmt.Println("M2 終了")  // 後処理2
         }
     }

     func main() {
         r := gin.New()
         r.Use(Middleware1(), Middleware2())
         
         r.GET("/", func(c *gin.Context) {
             fmt.Println("ルートハンドラ実行")
         })
     }
     ```
    - 上記の例の実行順序は以下の通り。  
      ~~~
       "M1 開始"
       "M2 開始"
       "ルートハンドラ実行"
       "M2 終了"  （Middleware2の後処理）
       "M1 終了"  （Middleware1の後処理）
      ~~~  
      つまり、**前処理は外側のミドルウェアから内側に向かって実行（M1 → M2）→ ルートハンドラの実行 → 後処理は内側のミドルウェアから外側に向かって実行（M2 → M1）という「スタック」のような動作をする。**
2. 条件に応じて後続の処理を中断するなど、ミドルウェアチェーンの実行フローを制御する手段
   - 例  
     ```go
     func AuthMiddleware() gin.HandlerFunc {
         return func(c *gin.Context) {
             if !isAuthenticated(c) {
                 c.AbortWithStatus(401)
                 return  // c.Next()を呼ばないことで後続の処理を実行しない
             }
             c.Next()
         }
     }
     ```

#### `c.Abort()`について
- `c.Abort()`が呼ばれた場合は、それ以降の`c.Next()`は実行されず、チェーン内の後続のミドルウェアやリクエストハンドラは実行されない点に注意が必要です。

## `Use`メソッド
Ginフレームワークにおける`Use`メソッドは、グローバルまたはルートレベルのミドルウェアを登録するために使用される。
ミドルウェアは、HTTPリクエストの処理中に特定の機能（ログ記録、認証、エラーハンドリングなど）を実行するための関数。

`Use`メソッドの動作は以下のステップで行われる：

1. **ミドルウェアの登録** 
   - `Use`メソッドを呼び出すことで、指定されたミドルウェアがGinエンジンに登録される。

2. **リクエストの処理** 
   - HTTPリクエストがサーバーに到着すると、登録されたミドルウェアは順番に実行される。各ミドルウェアは`gin.Context`オブジェクトを受け取り、必要に応じてリクエストデータの処理やレスポンスデータの変更を行う。

3. **次のミドルウェアへの移行**
   - ミドルウェアは通常、`c.Next()`を呼び出して、次のミドルウェアまたは最終的なハンドラへの実行を続ける。
   - `c.Abort()`を呼び出すと、チェーンの残りのミドルウェアの実行が停止される。

4. **レスポンスの送信**
   - 最終的なハンドラがレスポンスを生成し、クライアントに送信される。  
     **この時点で、登録されたミドルウェアの`c.Next()`の後の部分が実行される場合がある（例えば、レスポンス後のログ記録など）。**
     - 例  
       ```go
       func AfterMiddleware() gin.HandlerFunc {
       	return func(c *gin.Context) {
       		// リクエスト処理の前に実行
       		startTime := time.Now()

       		// 次の処理へ進む
       		c.Next()

       		// リクエスト処理の後に実行
       		latency := time.Since(startTime)
       		fmt.Printf("Request processed in %v\n", latency)
       	}
       }

       func main() {
       	r := gin.Default()

       	// ミドルウェアを全ルートに適用
       	r.Use(AfterMiddleware())

       	r.GET("/hello", func(c *gin.Context) {
       		c.JSON(200, gin.H{"message": "Hello, World!"})
       	})

       	r.Run(":8080")
       }
       ```

`Use`メソッドを使用することで、Ginアプリケーションは柔軟にミドルウェアを活用し、リクエストの前後でさまざまな処理を行うことができる。

## Directory構造
- 決まったディレクトリ名/構造はないっぽい。下記は一例。
```
myproject/
├── main.go
├── go.mod
├── go.sum
├── controllers/
│   └── ...
├── models/
│   └── ...
├── middlewares/
│   └── ...
├── router/
│   └── ...
├── services/
│   └── ...
├── static/
│   ├── css/
│   ├── js/
│   └── images/
├── templates/
│   └── ...
└── config/
    └── ...
```
- 各ディレクトリの役割
  - `controllers/`: リクエストハンドラ（ルーティングロジック）を含むパッケージ
  - `models/`: データベーススキーマやモデル構造体を定義するパッケージ
  - `middlewares/`: 認証、ロギング、エラーハンドリングなどのミドルウェアを含むパッケージ
  - `router/`: ルーティング設定を行うパッケージ
  - `services/`: ビジネスロジックを実装するパッケージ
  - `static/`: 静的ファイル（CSS、JavaScript、画像など）を格納するディレクトリ
  - `templates/`: HTMLテンプレートファイルを格納するディレクトリ
  - `config/`: 設定ファイル（データベース接続情報など）を格納するパッケージ
- `main.go`の例  
  ```go
  package main

  import (
      "myproject/config"
      "myproject/router"
      "myproject/middlewares"
      "github.com/gin-gonic/gin"
  )

  func main() {
      // 設定ファイルの読み込み
      config.LoadConfig()

      // Ginルーターの初期化
      router := gin.Default()

      // ミドルウェアの登録
      router.Use(middlewares.Logger())
      router.Use(middlewares.ErrorHandler())

      // 静的ファイルの提供
      router.Static("/static", "./static")

      // HTMLテンプレートの設定
      router.LoadHTMLGlob("templates/*")

      // ルーティングの設定
      routes.SetupRoutes(router)

      // サーバーの起動
      router.Run(":8080")
  }
  ```
- `router/`ディレクトリ内のコード (e.g. `router/router.go`) の例  
  ```go
  package router

  import (
      "myproject/controllers"
      "github.com/gin-gonic/gin"
  )

  func SetupRoutes(router *gin.Engine) {
      // ルートグループの作成
      v1 := router.Group("/api/v1")
      {
          // ユーザー関連のルート
          users := v1.Group("/users")
          {
              users.POST("/", controllers.CreateUser)
              users.GET("/", controllers.GetUsers)
              users.GET("/:id", controllers.GetUser)
              users.PUT("/:id", controllers.UpdateUser)
              users.DELETE("/:id", controllers.DeleteUser)
          }

          // 記事関連のルート
          articles := v1.Group("/articles")
          {
              articles.POST("/", controllers.CreateArticle)
              articles.GET("/", controllers.GetArticles)
              articles.GET("/:id", controllers.GetArticle)
              articles.PUT("/:id", controllers.UpdateArticle)
              articles.DELETE("/:id", controllers.DeleteArticle)
          }
      }

      // インデックスページのルート
      router.GET("/", controllers.Index)
  }
  ```

#### `services/`ディレクトリと`controllers/`ディレクトリの違いについて
- `controllers/`と`services/`は、アプリケーションの異なる層を表現している。それぞれの役割は以下の通り：

1. `controllers/`（コントローラー層）：
   - HTTPリクエストを受け取り、必要なデータを`services/`から取得する。
   - 取得したデータを適切な形式（JSON、HTML）でレスポンスとして返す。
   - リクエストのバリデーションやエラーハンドリングを行う。
   - ビジネスロジックは含まず、主にリクエストとレスポンスの処理に専念。

2. `services/`（サービス層）：
   - ビジネスロジックを実装します。
   - データベースやAPIなど、外部のリソースとのやり取りを行う。
   - 複雑な処理を行い、データを加工して`controllers/`に返す。
   - `controllers/`から受け取ったデータを処理し、結果を返す。
   - サービス層は、アプリケーションの核となるビジネスルールを定義する。

つまり、`controllers/`はHTTPリクエストとレスポンスを処理し、`services/`はビジネスロジックを実装する。

例えば、ユーザー登録の処理を考えてみよう：

1. `controllers/`では、リクエストからユーザーの情報を取得し、バリデーションを行う。
2. バリデーションが成功した場合、`services/`の`CreateUser`関数を呼び出す。
3. `services/`では、受け取ったユーザー情報をデータベースに保存する処理を実装する。
4. 保存が成功した場合、`controllers/`に成功レスポンスを返す。
5. `controllers/`では、受け取った結果をHTTPレスポンスとしてクライアントに返す。

このように、`controllers/`と`services/`を分離することで、関心事を分離し、コードの可読性とメンテナンス性を向上させることができる。また、`services/`は`controllers/`だけでなく、他の部分からも呼び出すことができるため、コードの再利用性も高くなる。

ただし、プロジェクトの規模や要件によっては、`controllers/`と`services/`を明確に分離しない場合もある。小規模なプロジェクトでは、`controllers/`にビジネスロジックを含めることもあるが、アプリケーションが成長するにつれて、徐々に`services/`を導入していくことが望ましい。

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
