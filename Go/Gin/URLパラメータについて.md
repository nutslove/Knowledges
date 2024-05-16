- URLで`:<some_id>`の部分はURLパラメータを表すプレースホルダー。これは、動的なセグメントを定義するために使用される。  
  Ginでは`*gin.Context.Param("<some_id>")`の形式でURLパラメータを取得できる。  
  例えば以下の例だと、`/api/v1/logaas/12`でアクセスすると、`:logaas_id`に`12`が入り、`c.Param("logaas_id")`で`12`が取得される。
  ```go
    func SetupRouter(r *gin.Engine) {
        v1 := r.Group("/api/v1")

        {
            logaas := v1.Group("/logaas")
            {
                logaas.POST("/:logaas_id", services.CreateLogaas)
                logaas.GET("/:logaas_id", services.GetLogaas)
                logaas.DELETE("/:logaas_id", services.DeleteLogaas)
                logaas.GET("/", services.GetLogaases)
            }
        }
    }
  ```

  ```go
    func GetLogaas(c *gin.Context) {
        LogaasID := c.Param("logaas_id")
        // LogaasIDを使用してロジックを実装
        // ...
    }
  ```