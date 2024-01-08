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
