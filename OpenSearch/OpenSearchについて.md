- ブラウザにて`<OpenSearchのIP>:9200/<index名>?pretty`で対象indexについての情報を確認できる
- ブラウザにて`<OpenSearchのIP>:9200/<index名>/_search?pretty`で対象indexのデータを確認できる
  - デフォルトでは10件しか表示されず、`<OpenSearchのIP>:9200/<index名>/_search?size=<表示件数>`のように`?size`で表示件数を指定できる

# index
- https://opensearch.org/docs/latest/getting-started/intro/#index
- RDBで言うと**Table**
> **An index is a collection of documents.**
>
> You can think of an index in several ways:
>
> - In a database of students, an index represents all students in the database.
> - When you search for information, you query data contained in an index.
> - **An index represents a database table in a traditional database.**
>
>For example, in a school database, an index might contain all students in the school.
> |ID |	Name | GPA | Graduation year |
> | --- | --- | --- | --- |
> | 1	| John | Doe | 3.89 | 2022 |
> | 2 | Jonathan | Powers | 3.85 | 2025 |
> | 3	| Jane Doe | 3.52 | 2024 |
- OpenSearchはデータをindexに格納する
  - データをindexに格納することで検索が可能になる
- 一般的に特定のタイプの文書やデータの**集合**に対して作成

### indexの作成
- OpenSearch DashBoardのDev ToolにてPUTメソッドで追加できる
- 例  
  ```shell
  PUT /{新しいindex名}
  {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 1
    },
    "mappings": {
      "properties": {
        "my_field": {
          "type": "text"
        },
        "another_field": {
          "type": "keyword"
        }
      }
    }
  }
  ```

# document
- https://opensearch.org/docs/latest/getting-started/intro/#document
- RDBで言うと**レコード**
- OpenSearchで、データの最小単位
- **Documentごとに一意のIDを持たせる必要がある。すでに存在するIDでDocumentを登録すると既存のIDの内容が上書きされる。**
  - 登録時IDを指定しなかったら、自動で一意のIDが割り当てられる
- 複数のfieldsから構成される
> A document is a unit that stores information (text or structured data). In OpenSearch, documents are stored in JSON format.
> 
> You can think of a document in several ways:
> - In a database of students, a document might represent one student.
> - When you search for information, OpenSearch returns documents related to your search.
> - **A document represents a row in a traditional database.**
>
> For example, in a school database, a document might represent one student and contain the following data.
> |ID |	Name | GPA | Graduation year |
> | --- | --- | --- | --- |
> | 1	| John | Doe | 3.89 | 2022 |

# fields
- RDBで言うと**カラム**
- keyとvalueの組
- 転置インデックスはフィールドごとに作成/管理される。  
  なので、クエリー実行時基本的にフィールド単位で検索される。
- fieldのtypeを定義(指定)できる
  - **https://opensearch.org/docs/latest/field-types/**
### 主なfieldsのタイプ(型)
#### `text`
- analyzerで
- **部分一致**

#### `keyword`
- **完全一致**

### fieldsのタイプ(型)の確認方法
- OpenSearch DashBoardのDev Toolにて以下の通り打てば確認できる  
  ```shell
  GET /{index名}/_mapping
  {
    "query": {
      "match_all": {}
    }
  }
  ```

# OpenSearchの本番運用に向けて
- https://zenn.dev/istyle/articles/9d8dcfcd16c1b9

# API
- https://opensearch.org/docs/latest/api-reference/search/

### あるIndex内のあるIDのDocument検索
- 例  
  ```shell
  curl -k -u <ユーザ名>:<パスワード> http(s)://<APIのエンドポイント>/<対象index名>/_doc/<対象ID>?pretty
  ```

### あるIndex内の全Document検索
- 例  
  ```shell
  curl -k -u <ユーザ名>:<パスワード> http(s)://<APIのエンドポイント>/<対象index名>/_search?pretty
  ```
- キーワード検索の例  
  ```shell
  curl -k -u <ユーザ名>:<パスワード> http(s)://<APIのエンドポイント>/<対象index名>/_search?pretty -H 'Content-Type: application/json' -d'
  {
    "query": {
      "term": {
        "<Key>": "<Value>"
      }
    }
  }'
  ```

# adminのパスワード
- v2.11.1まではデフォルトのadminのパスワードとして`admin`で自動的に払い出されたけど、  
  v2.12.0からは環境変数`OPENSEARCH_INITIAL_ADMIN_PASSWORD`にデフォルトのadminパスワードを設定する必要がある。（相当複雑なPWじゃないとweakとエラーとなる）
  - https://opensearch.org/blog/replacing-default-admin-credentials/

## Searchの取得件数
- デフォルトではTopの10件のみ取得されるけど、`"size"`パラメータで10より大きい数字を指定することで10件以上取得することができる
  - https://medium.com/@musabdogan/elasticsearch-query-to-return-all-records-more-than-10-000-documents-84fe1bfee661
- 例  
  ```go
  content := strings.NewReader(fmt.Sprintf(`{
    "size": 100,
    "query": {
        "bool": {
            "should": [
                { "match_phrase": { "title": "%s" }},
                { "match_phrase": { "post": "%s" }}
            ]
        }
    }
  }`, searchKeyword, searchKeyword))
  ```