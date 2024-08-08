- ブラウザにて`<OpenSearchのIP>:9200/<index名>?pretty`で対象indexについての情報を確認できる
- ブラウザにて`<OpenSearchのIP>:9200/<index名>/_search?pretty`で対象indexのデータを確認できる
  - デフォルトでは10件しか表示されず、`<OpenSearchのIP>:9200/<index名>/_search?size=<表示件数>`のように`?size`で表示件数を指定できる

### document
- https://opensearch.org/docs/latest/getting-started/intro/#document
- RDBで言うと**レコード**
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

### index
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

### fields
- RDBで言うと**カラム**
- fieldのtypeを定義(指定)できる
  - **https://opensearch.org/docs/latest/field-types/**

### OpenSearchの本番運用に向けて
- https://zenn.dev/istyle/articles/9d8dcfcd16c1b9

### API
- https://opensearch.org/docs/latest/api-reference/search/

### adminのパスワード
- v2.11.1まではデフォルトのadminのパスワードとして`admin`で自動的に払い出されたけど、  
  v2.12.0からは環境変数`OPENSEARCH_INITIAL_ADMIN_PASSWORD`にデフォルトのadminパスワードを設定する必要がある。（相当複雑なPWじゃないとweakとエラーとなる）
  - https://opensearch.org/blog/replacing-default-admin-credentials/