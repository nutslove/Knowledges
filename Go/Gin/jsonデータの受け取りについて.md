- Ginでjsonデータを受け取って処理するためにはまず構造体(Struct)で型を定義し、`gin.Context.BindJSON(*<jsonとマッピングするためのStructの変数>)`で関連付ける
  - `gin.Context.BindJSON(*<jsonとマッピングするためのStructの変数>)`で関連付けた時点で`jsonとマッピングするためのStructの変数`にjsonフォーマットのリクエストデータが入る
  - 以下の例では`req`にリクエストで送られたjsonデータが格納されて、`req.Db_User`,`req.Iam_User`,`req.Os_User`にスライスとして各データが入る
  - リクエストを送る側は以下のようにstructに定義したKey名で送る必要がある
    - `curl -X POST -H "Content-Type: application/json" -d '{"db_user":["dbuser1","dbuser2"], "iam_user":["iam_user1","iam_user2","iam_user3"], "os_user":["os_user1"]}' http://<サーバURL/IP>:8080/post`
  - リクエストを送る側はstructの中のすべての項目のデータを送る必要はない。  
    以下のように一部の項目だけ送った場合、送られてない項目(以下の例では`Db_User`)Go側で型のゼロ値 (e.g. 空のスライス) になる
      - ``curl -X POST -H "Content-Type: application/json" -d '{"iam_user":["iam_user1","iam_user3"], "os_user":["os_user1"]}' http://<サーバURL/IP>:8080/post``
    ~~~go
    // 受け取るJsonデータのフォーマット
    type UserExistCheck struct {
    	Db_User       []string `json:"db_user"`
    	Iam_User      []string `json:"iam_user"`
    	Os_User       []string `json:"os_user"`
    }

    func main() {
        r := gin.Default()

        r.POST("/post", func(c *gin.Context) {
            var req ExampleRequest
            if err := c.BindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
            }

            // ここでreqを使用して処理を行う。
            fmt.Println(req.Db_User)
            fmt.Println(req.Iam_User)
            fmt.Println(req.Os_User)
            c.JSON(http.StatusOK, gin.H{"status": "OK"}) // JSON形式で返す
        })

        r.Run() // デフォルトでは ":8080" でサーバーを起動。
    }
    ~~~
    - 以下Chat-GPTからの回答
      > **この例では、POSTリクエストからJSONデータを読み込んで ExampleRequest 型の req 変数に格納しています。** その後、この変数を使用してさまざまな操作を行うことができます。例えば、データベースへの保存、ロジックの処理、またはそのデータをレスポンスとして返すことなどです。
      > **c.JSON はレスポンスをクライアントに送り返すときに使用し、c.BindJSON はリクエストからデータを読み込むときに使用します。**