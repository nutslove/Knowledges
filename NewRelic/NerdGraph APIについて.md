- https://zenn.dev/aeonpeople/articles/ea5b0363cdf989#nerdgraph-api-%E3%81%A8%E3%81%AF
- https://docs.newrelic.com/jp/docs/apis/nerdgraph/examples/nerdgraph-entities-api-tutorial/
- https://docs.newrelic.com/jp/docs/apis/nerdgraph/examples/nerdgraph-nrql-tutorial/
- https://docs.newrelic.com/jp/docs/apis/nerdgraph/get-started/nerdgraph-explorer/

# NerdGraph APIとは
- データのクエリや New Relic 機能の設定変更に使用できる GraphQL API
- defaultでは10件しか取得しない（defaultで`LIMIT 10`が適用されているみたい）
  - `LIMIT`の最大値は1000っぽい

## NRQL
- https://docs.newrelic.com/jp/docs/nrql/nrql-syntax-clauses-functions/

# NerdGraph APIの例
## ログ
- **`WHERE`句で条件を絞る時（以下例のERROR）、シングルクォーテーション('')で囲む必要があるけど、curlで叩くとこは「`'\''<値>'\''`」のフォーマットで指定する必要がある**

1.  
```
curl -X POST https://api.newrelic.com/graphql -H 'Content-Type: application/json' \
-H 'API-Key: <APIキー>' \
-d '{ "query": "{ actor { account(id: <アカウントID>) { nrql(query: \"SELECT `level`, `message` FROM Log WHERE `level` = '\''ERROR'\'' OR `level` = '\''WARN'\'' SINCE 10 hours ago LIMIT 1000\") { results } } } }" }' | jq
```

2.   
```
curl -X POST https://api.newrelic.com/graphql -H 'Content-Type: application/json' \
-H 'API-Key: <APIキー>' -d '{ "query": "{ actor { account(id: <アカウントID>) { nrql(query: \"SELECT message, level FROM Log WHERE level = '\''ERROR'\'' SINCE 12 hours ago UNTIL now LIMIT 100\") { results }
 } } }" }' | jq
```

## メトリクス
- 基本的にはnrqlを使ってやる場合ログと一緒

1. `SINCE`と`UNTIL`でメトリクスの範囲を指定  
```
curl -X POST https://api.newrelic.com/graphql \
     -H 'Content-Type: application/json' \
     -H 'API-Key: <APIキー>' \
     -d '{ "query": "{ actor { account(id: <アカウントID>) { nrql(query: \"SELECT filter(count(*), WHERE request.uri NOT LIKE '"'"'%dealer/new-applications%'"'"') AS WEB商流, filter(count(*), WHERE request.uri LIKE '"'"'%dealer/new-applications%'"'"') AS 販売店商流 FROM Transaction WHERE appName = '"'"'prod-goku-core'"'"' AND request.method = '"'"'POST'"'"' AND request.uri LIKE '"'"'%new-applications%'"'"' AND request.uri NOT LIKE '"'"'%/cancel%'"'"' FACET http.statusCode TIMESERIES 1 hour SINCE 6 hours ago UNTIL now LIMIT MAX EXTRAPOLATE\") { results } } } }" }' | jq
```

## アラート
1. アラート一覧確認（状態とpriorityも含めて）  
```
curl -X POST https://api.newrelic.com/graphql -H 'Content-Type: application/json' -H 'API-Key: <APIキー>' -d '{"query": "{ actor { account(id: <アカウントID>) { aiIssues { issues { issues { issueId priority state title } } } } } }"}' | jq
```

2. `filter`で条件を絞ってアラート一覧確認  
```
curl -X POST https://api.newrelic.com/graphql -H 'Content-Type: application/json' -H 'API-Key: <APIキー>' -d '{"query": "{ actor { account(id: <アカウントID>) { aiIssues { issues(filter: { priority: \"CRITICAL\", states: ACTIVATED }) { issues { issueId priority state title } } } } } }"}' | jq
```