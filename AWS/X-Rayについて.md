### フィルタ式
- X-Ray上のトレースを検索するためのクエリー
- 参考URL
  - https://docs.aws.amazon.com/ja_jp/xray/latest/devguide/xray-console-filters.html
- **フィルタ式はルートセグメント(Parent TraceIdを持たないRoot Span)しかヒット(検索)しない**
  - 以下AWSサポートからの回答
    > X-Ray の GetTraceSummaries API によるトレース検索の機能では、http.url による検索はルートセグメント (parent_id を持たないセグメント) のみが検索対象となっており、ルートセグメントではないセグメントについては URL が一致していても http.url での検索にヒットしない動作であることを確認いたしました。 
- フィルタ式でのokやfaultなどの意味
  - ok
    - Response StatusCodeが *2XX* Success
  - error
    - Response StatusCodeが *4XX* Client Error
  - throttle
    - Response StatusCodeが *429* Too Many Requests
  - fault
    - Response StatusCodeが *5XX* Server Error
