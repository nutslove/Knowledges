- ブラウザにて`<OpenSearchのIP>:9200/<index名>?pretty`で対象indexについての情報を確認できる
- ブラウザにて`<OpenSearchのIP>:9200/<index名>/_search?pretty`で対象indexのデータを確認できる
  - デフォルトでは10件しか表示されず、`<OpenSearchのIP>:9200/<index名>/_search?size=<表示件数>`のように`?size`で表示件数を指定できる

### OpenSearchの本番運用に向けて
- https://zenn.dev/istyle/articles/9d8dcfcd16c1b9

## API
- https://opensearch.org/docs/latest/api-reference/search/