- ブラウザにて`<OpenSearchのIP>:9200/<index名>?pretty`で対象indexについての情報を確認できる
- ブラウザにて`<OpenSearchのIP>:9200/<index名>/_search?pretty`で対象indexのデータを確認できる
  - デフォルトでは10件しか表示されず、`<OpenSearchのIP>:9200/<index名>/_search?size=<表示件数>`のように`?size`で表示件数を指定できる

### index
- OpenSearchはデータをindexに格納する
  - データをindexに格納することで検索が可能になる
- 一般的に特定のタイプの文書やデータの**集合**に対して作成

### OpenSearchの本番運用に向けて
- https://zenn.dev/istyle/articles/9d8dcfcd16c1b9

### API
- https://opensearch.org/docs/latest/api-reference/search/

### adminのパスワード
- v2.11.1まではデフォルトのadminのパスワードとして`admin`で自動的に払い出されたけど、  
  v2.12.0からは環境変数`OPENSEARCH_INITIAL_ADMIN_PASSWORD`にデフォルトのadminパスワードを設定する必要がある。（相当複雑なPWじゃないとweakとエラーとなる）
  - https://opensearch.org/blog/replacing-default-admin-credentials/