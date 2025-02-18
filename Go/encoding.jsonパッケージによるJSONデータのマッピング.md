## 使い方 (流れ)
- jsonデータのスキーマを`struct`で予め構造体(型)を作っておいて、`encoding/json`の`Marshal`メソッドでJSON形式にエンコードする。
- `struct`の構造体の例  
  ~~~go
  type UserExistCheck struct {
          Db_User       []string `json:"db_user"`
          Iam_User      []string `json:"iam_user"`
          Os_User       []string `json:"os_user"`
  }
  ~~~
  - __\`json:<任意の名前>\`__ の部分はtagで、構造体のフィールドにメタデータを付加するために使用される。  
    **<任意の名前>** の部分はJSONのキー名を意味する。
  - 上記の例だと`db_user`、`iam_user`、`os_user`がJSONのキー名
  - このtagは単なる型ヒントとかではなく、**`encoding/json`パッケージが、構造体とJSONの間でデータをマーシャリング（エンコード）およびアンマーシャリング（デコード）する際に、これらのtagを使用する。**  
    例えば、`json.Marshal`関数を使ってこの構造体をJSONにエンコードする場合、`json` tagで指定されたキー名が実際のJSONのキー名として使用される。  
    同様に、`json.Unmarshal`関数を使ってJSONをこの構造体にデコードする場合、JSONのキー名と `json` tagで指定されたキー名が一致するフィールドにデータが格納される。
- コード例  
  - 以下を実行すると **`{"db_user":["user1","user2"],"iam_user":["user3","user4"],"os_user":["user5","user6"]}`** と出力される
  ~~~go
  package main

  import (
          "fmt"
          "encoding/json"
  )

  func main() {
          type UserExistCheck struct {
                  Db_User       []string `json:"db_user"`
                  Iam_User      []string `json:"iam_user"`
                  Os_User       []string `json:"os_user"`
          }

          user := UserExistCheck{
                  Db_User: []string{"user1", "user2"},
                  Iam_User: []string{"user3", "user4"},
                  Os_User: []string{"user5", "user6"},
          }

          jsonData, _ := json.Marshal(user)

          fmt.Println(string(jsonData))
  }
  ~~~

## 深い階層のJSONデータのマッピング
- 例えば、以下の`jsonData`のデータの場合、`results`の部分を取得するためのStructは以下のようになる  
  ```go
  package main

  import (
  	"encoding/json"
  	"fmt"
  )

  // JSONの構造をGoの構造体にマッピング
  type NrqlResponse struct {
  	Data struct {
  		Actor struct {
  			Account struct {
  				Nrql struct {
  					Results []map[string]interface{} `json:"results"`
  				} `json:"nrql"`
  			} `json:"account"`
  		} `json:"actor"`
  	} `json:"data"`
  }

  func main() {
  	// NewRelicのJSONデータ
  	jsonData := `{"data":{"actor":{"account":{"nrql":{"results":[{"level":"WARN","message":"Business exception occurred.","timestamp":1739861118060}]}}}}}`

  	// JSONをパース
  	var response NrqlResponse
  	err := json.Unmarshal([]byte(jsonData), &response)
  	if err != nil {
  		fmt.Println("JSONデコードエラー:", err)
  		return
  	}

  	// 結果を出力
  	fmt.Println(response.Data.Actor.Account.Nrql.Results)
  }
  ```