- OpenStackのGo SDK
  - https://github.com/gophercloud/gophercloud

## KeyStone認証
### パスワード認証
- `gophercloud.AuthOptions`構造体に認証に必要な情報を入れて、`AuthenticatedClient()`メソッドに渡して認証する
- その後、各サービスごとに用意されている`openstack`のメソッド(e.g. Cinderの場合`NewBlockStorageV3`)でクライアントを初期化して、そのクライアントを使って各種サービスを操作する

```go
package main

import (
    "fmt"

    "github.com/gophercloud/gophercloud"
    "github.com/gophercloud/gophercloud/openstack"
    _ "github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
    "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
)

func main() {
    opts := gophercloud.AuthOptions{
            IdentityEndpoint: "http(s)://<KeyStoneエンドポイント>:5000/v3",
            Username:         "<ユーザ名>",
            Password:         "<パスワード>",
            DomainName:       "Default",
            TenantName:       "<テナントID(Projectを使っている場合はProjectID)>",
    }

    // プロバイダーを作成
    provider, err := openstack.AuthenticatedClient(opts)
    if err != nil {
            fmt.Println("認証中にエラーが発生しました:", err)
            return
    }

    // ここでclientを使用してKeystone APIを操作できます
    fmt.Println("Keystone認証に成功しました！")

    // Keystoneサービスクライアントを作成
    keystoneclient, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{
            Region: "<region名>",
    })
    if err != nil {
            fmt.Println("Keystoneサービスクライアントの作成中にエラーが発生しました:", err)
            return
    }

    // Cinderサービスクライアントを作成
    cinderclient, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
            Region: "<region名>",
    })
    if err != nil {
            fmt.Println("Cinderサービスクライアントの作成中にエラーが発生しました:", err)
            return
    }

	// Cinderボリュームを作成
	createOpts := volumes.CreateOpts{
		Size: 1, // ボリュームサイズをギガバイトで指定
		Name: "<ボリューム名>",
	}
	vol, err := volumes.Create(client, createOpts).Extract()
	if err != nil {
		fmt.Println("ボリュームの作成中にエラーが発生しました:", err)
		return
	}
	fmt.Printf("ボリュームが作成されました: %+v\n", vol)

    // ボリュームリストを確認
    listOpts := volumes.ListOpts{}
    allPages, err := volumes.List(cinderclient, listOpts).AllPages()
    if err != nil {
            fmt.Println("ボリュームリストの取得中にエラーが発生しました:", err)
            return
    }
    allVolumes, err := volumes.ExtractVolumes(allPages)
    if err != nil {
            fmt.Println("ボリュームデータの抽出中にエラーが発生しました:", err)
            return
    }
    for _, volume := range allVolumes {
            fmt.Printf("ボリューム: %+v\n", volume)
    }
}
```

### Token認証
> [!TIP]
> tokenは`openstack token issue -f value -c id`で発行できる  
> https://docs.openstack.org/python-openstackclient/latest/cli/authentication.html

- `tokens.Validate()`メソッドでTokenの正当性を確認  
  Tokenが有効な場合は第1戻り値として`true`が、無効な場合は`false`が返ってくる
- `tokens.Get()`メソッドでTokenに関する詳細な情報(e.g. Project ID、RoleName)を取得できる
  - `tokens.Get()`メソッドは`Body`、`StatusCode`、`Header`、`Err`が含まれている`GetResult`型を返す
    - https://github.com/gophercloud/gophercloud/blob/master/results.go#L28
    - https://github.com/gophercloud/gophercloud/blob/master/openstack/identity/v3/tokens/requests.go
- 例
  ```go
  import (
        "github.com/gophercloud/gophercloud"
        "github.com/gophercloud/gophercloud/openstack"
        "github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
  )
  
  func main() {
        opts := gophercloud.AuthOptions{
                IdentityEndpoint: "http://<KeyStone認証エンドポイント>:5000/v3",
                Username:         "<ユーザ名>",
                Password:         "<パスワード>",
                DomainName:       "Default",
                TenantName:       "<Project名>",
        }

        // プロバイダーを作成
        provider, err := openstack.AuthenticatedClient(opts)
        if err != nil {
                fmt.Println("認証中にエラーが発生しました:", err)
                return
        }

        // Keystoneサービスクライアントを作成
        keystoneclient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{
                Region: "<Region名>",
        })
        if err != nil {
                fmt.Println("Keystoneサービスクライアントの作成中にエラーが発生しました:", err)
                return
        }

        token := "<発行したToken>"
        tokenValidateResult, err := tokens.Validate(keystoneclient, token)
        if err != nil {
                fmt.Println("トークンの検証中にエラーが発生しました:", err)
                return
        }
        fmt.Println("トークン検証結果:", tokenValidateResult)
        if !tokenValidateResult {
                fmt.Println("トークンは無効です")
                return
        }

        // トークンの詳細情報を取得
        tokenDetails := tokens.Get(keystoneclient, token)

        fmt.Println("Token詳細情報:", tokenDetails)
        fmt.Println("-------------------------------------------------------------\n")
        fmt.Println("Body:", tokenDetails.Body)
        fmt.Println("-------------------------------------------------------------\n")
        fmt.Println("StatusCode:", tokenDetails.StatusCode)
        fmt.Println("-------------------------------------------------------------\n")
        fmt.Println("Header:", tokenDetails.Header)
        fmt.Println("-------------------------------------------------------------\n")

        tokeninfo, ok := tokenDetails.Body.(map[string]interface{})["token"].(map[string]interface{})
        if !ok {
                fmt.Println("tokenDetails.Bodyの型変換に失敗しました")
                return
        }
        projectID := tokeninfo["project"].(map[string]interface{})["id"]
        fmt.Println("Project ID:", projectID)

        roles := tokeninfo["roles"].([]interface{})
        var isAdmin bool
        var roleName string
        for _, role := range roles {
                roleName = role.(map[string]interface{})["name"].(string)
                fmt.Println("Role Name:", roleName)
                if strings.Contains(roleName,"admin") {
                        isAdmin = true
                }
        }
        fmt.Println("isAdmin:", isAdmin)
  }
  ```