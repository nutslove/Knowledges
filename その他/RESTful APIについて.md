## RESTful APIとは
- REST（Representational State Transfer）の原則に従って設計されたAPIのこと
- 参考URL
  - https://learn.microsoft.com/ja-jp/azure/architecture/best-practices/api-design
  - https://aws.amazon.com/jp/what-is/restful-api/
  - https://blog.postman.com/rest-api-examples/

## REST API成熟度モデル
- **0レベル**
  - 1つのURIを定義し、すべての操作がこのURIに対するPOST要求
- **1レベル**
  - リソースごとに個別のURIを作成
- **2レベル**
  - HTTPメソッド（e.g. `DELETE`、`GET`など）を使用して、リソースに対する操作を定義
- **３レベル**
  - ハイパーメディア（HATEOAS）を使用

#### HATEOAS（Hypermedia As The Engine Of Application State）とは
- RESTful APIにおける重要な概念の一つで、APIのレスポンスにハイパーメディア（リンクやアクション）を含めて、クライアントが次に取るべきアクションをガイドする仕組み
- HATEOASの主な特徴は、クライアントがサーバーとのインタラクションに必要な情報をすべてレスポンスから得られること。具体的には、レスポンスには次に呼び出すべきリソースやアクションへのリンクが含まれており、クライアントはそのリンクを辿ることでリソースを操作していくことができる
- たとえば、以下のようなユーザー情報を取得するAPIレスポンスがあるとする  
  ```json
  {
    "id": 123,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "_links": {
      "self": {
        "href": "/users/123"
      },
      "update": {
        "href": "/users/123/update",
        "method": "PUT"
      },
      "delete": {
        "href": "/users/123",
        "method": "DELETE"
      }
    }
  }
  ```  
  この例では、`_links`というハイパーメディアが含まれている。これにより、クライアントはこのユーザーの情報を取得した後に、以下のような次のアクションが可能であることが分かる。  
  
  `self`: 現在のリソースへのリンク（この場合はユーザー自身の情報を指す）。
  `update`: ユーザー情報を更新するためのリンクと、使用すべきHTTPメソッド（PUT）。
  `delete`: ユーザーを削除するためのリンクと、使用すべきHTTPメソッド（DELETE）。

## RESTの6つの原則
### 1. クライントとサーバーの分離
- クライアントとサーバーが互いに独立し、完全に分離していること
### 2. 統一されたインターフェイス
- 統一されたインターフェイスは、次の 4 つのアーキテクチャ上の制約を課す
  1. リクエストはリソースを特定する必要があり、これはURIを使用することによって行う
  2. 
### 3. Stateless
### 4. 階層型システム
- 
### 5. キャッシュ可能性
### 6. オンデマンドのコード

## 更新のためのHTTPメソッド
#### POST
- 新規作成やサーバー側の処理を伴うリソースの作成
#### PUT
- 既存リソースの完全な置き換え
- クライアントが全てのフィールド（リソースの完全な表現）を送信し、そのデータでリソース全体を更新
#### PATCH
- 部分的な更新
- リソース全体ではなく、変更したい部分のみを送信して更新

## URIの設計
- 

## RESTful APIの認証
### HTTP認証
#### Basic認証
#### Bearer 認証
### API Key
### JWT（JSON Web Token）
### OAuth