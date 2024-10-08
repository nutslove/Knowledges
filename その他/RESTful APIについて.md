## RESTful API（REST API）とは
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
  
  `self`: 現在のリソースへのリンク（この場合はユーザー自身の情報を指す）  
  `update`: ユーザー情報を更新するためのリンクと、使用すべきHTTPメソッド（PUT）  
  `delete`: ユーザーを削除するためのリンクと、使用すべきHTTPメソッド（DELETE）

## RESTの6つの原則
### 1. クライントとサーバーの分離
- クライアントとサーバーが互いに独立し、完全に分離していること
### 2. 統一されたインターフェイス
- 統一されたインターフェイスは、サーバーが標準形式で情報を転送することを示す。フォーマットされたリソース（のある断面）は、REST では「表現」と呼ばれる。この形式は、サーバーアプリケーション上のリソースの内部表現とは異なる場合がある。例えば、サーバーはデータをテキストとして保存されるが、クライアントにはJSON表現形式で送信する。  
  統一されたインターフェイスは、次の4つのアーキテクチャ上の制約を課す。
  1. **リソースの識別**
     - 各リソースはURIで一意に識別され、クライアントはリソースのURIを通じてそれにアクセスし、操作する
     - URIには動作(e.g. delete)を含まない
  2. **リソースの表現**
     - サーバー上のリソースは、クライアントに送信されるときに適切な表現（e.g. JSONやXML）に変換され、クライアントはこの表現を用いてリソースを操作する
  3. **自己記述型メッセージ**
     - 各メッセージには、リクエストやレスポンスに必要な情報（メソッド、ステータスコード、ヘッダーなど）がすべて含まれているべき。これにより、クライアントとサーバーがやり取りするメッセージが一貫性を保つ
       - e.g. ヘッダーにレスポンスデータタイプ（`Content-Type`）(e.g. json)が含まれている
  4. **アプリケーション状態のエンジンとしてのハイパーメディア（HATEOAS）**
     - クライアントが次にどのリソースにアクセスすべきかを判断できるよう、サーバーからのレスポンスにハイパーメディアリンクが含まれている
### 3. Stateless
- すべてのリクエストは１回の独立したリクエストで、リクエストのコンテキスト(文脈)が理解できて処理できる必要がある
### 4. 階層型システム（Layered System）
- クライアントとサーバーの間に複数の中間層（e.g. プロキシ、ゲートウェイ、キャッシュ、ロードバランサーなど）を導入しても、システム全体が機能し続けること
### 5. キャッシュ可能性
- サーバーの応答時間を改善するために、レスポンスを明示的または暗黙的にキャッシュ可能にする
### 6. オンデマンドのコード
- クライアントが必要なときにサーバーからコードを取得し、実行する
  - e.g. JavaScriptコードをサーバーから送ってブラウザ上でコードを実行する

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

## RESTful APIの認証
### HTTP認証
#### Basic認証
#### Bearer 認証
### API Key
### JWT（JSON Web Token）
### OAuth