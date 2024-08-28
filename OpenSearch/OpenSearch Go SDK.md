- SDKのバージョンはv4まであって、これから実装する場合は基本的にv4を使うこと
- https://github.com/opensearch-project/opensearch-go/blob/main/USER_GUIDE.md

## 新規ドキュメント登録
- `opensearchapi.Client`の **`Index`メソッド** を使う
  - https://github.com/opensearch-project/opensearch-go/blob/main/opensearchapi/api_index.go#L19
- 例  
  ```go
  func OpensearchNewClient() (*opensearchapi.Client, error) {
    client, err := opensearchapi.NewClient(
      opensearchapi.Config{
        Client: opensearch.Config{
          Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
          },
          Addresses: []string{"https://localhost:9200"},
          Username:  "admin",
          Password:  os.Getenv("OPENSEARCH_ADMIN_PASSWORD"),
        },
      },
    )

    return client, err
  }

  type Post struct {
    Title  string `json:"title"`
    Author string `json:"author"`
    Post   string `json:"post"`
  }

  post := Post{
    Title:  requestData.Title,
    Author: username.(string),
    Post:   requestData.Content,
  }

  client, err := OpensearchNewClient()
  if err != nil {
    log.Fatal("cannot initialize", err)
  }

  // ドキュメントの挿入
  insertResp, err := client.Index(
    context.Background(),
    opensearchapi.IndexReq{
      Index:      "career",
      DocumentID: strconv.Itoa(addedPost.Number),
      Body:       opensearchutil.NewJSONReader(&post),
      Params: opensearchapi.IndexParams{
        Refresh: "true",
      },
    })

  fmt.Printf("Created document in %s\n  ID: %s\n", insertResp.Index, insertResp.ID)
  ```

## 既存ドキュメントの更新
- `opensearchapi.Client`の **`Update`メソッド** を使う
  - https://github.com/opensearch-project/opensearch-go/blob/main/opensearchapi/api_update.go#L19
- **更新の場合は直接ドキュメントのフィールドを指定するのではなく、`doc`フィールド内にUpdateしたいフィールドを指定する必要がある**
- OpenSearchに送信されるJSONは以下のような形式になる  
  ```json
  {
    "doc": {
      "title": "更新されたタイトル",
      "author": "ユーザー名",
      "post": "更新された内容"
    }
  }
  ```
- 例  
  ```go
  type Post struct {
    Title  string `json:"title"`
    Author string `json:"author"`
    Post   string `json:"post"`
  }

  type UpdateRequest struct {
    Doc Post `json:"doc"`
  }

  postForOpensearch := Post{
    Title:  requestData.Title,
    Author: username.(string),
    Post:   requestData.Content,
  }

  updatedPost := UpdateRequest{
    Doc: postForOpensearch,
  }

  client, err := config.OpensearchNewClient()
  if err != nil {
    log.Fatal("cannot initialize", err)
  }

  // ドキュメントの更新
  updateResp, err := client.Update(
    context.Background(),
    opensearchapi.UpdateReq{
      Index:      "career",
      DocumentID: strconv.Itoa(postForDB.Number),
      Body:       opensearchutil.NewJSONReader(&updatedPost),
      Params: opensearchapi.UpdateParams{
        Refresh: "true",
      },
    })
  
  fmt.Printf("Updated document in %s\n  ID: %s\n", updateResp.Index, updateResp.ID)
  ```