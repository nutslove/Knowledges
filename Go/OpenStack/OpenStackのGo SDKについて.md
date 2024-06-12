- OpenStackのGo SDK
  - https://github.com/gophercloud/gophercloud

## KeyStone認証
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