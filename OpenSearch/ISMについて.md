## ISMとは
- ISMはOpensearch独自の機能で、インデックスの状態とその遷移をポリシーという概念で定義し、それらをインデックスおよびインデックスパターンに適用してインデックスの状態を自動で管理するもの。**一定期間が経過した古いインデックスを定期的に削除したり、read_only状態に移行してリソースを削減することができる。**

## ISM Policyによるインデックス管理の注意点
- 作成されたインデックスが複数のポリシーの`index_pattern`に一致する場合は`priority`(優先度)が高いポリシーが適用される
- カスタムISMポリシーを新規作成する場合、既存のインデックスには自動で適用されない。  
  (`index_pattern`に当てはまる場合でも、既存のインデックスは手動で適用させる必要がある)

## ISM Policyを反映するスクリプト例
```shell
#!/bin/bash
# Add preset ISM
ES_URL=http://localhost:9200
POLS_URL=$ES_URL/_plugins/_ism/policies
while [[ "$(curl -u ism:$ISM_PASSWORD -s -o /dev/null -w '%{http_code}\n' $POLS_URL)" != "200" ]]; do sleep 1; done
curl -u ism:$ISM_PASSWORD -XPUT "$POLS_URL/delete_90d" -H 'Content-Type: application/json' -d'{"policy":{"description":"Default policy","default_state":"hot","ism_template":{"index_patterns":["*"],"priority":0},"states":[{"name":"hot","actions":[],"transitions": [{"state_name": "delete","conditions": {"min_index_age": "90d"}}]},{"name": "delete","actions": [{"delete": {}}],"transitions": []}]}}'
curl -u ism:$ISM_PASSWORD -XPUT "$POLS_URL/delete_14d" -H 'Content-Type: application/json' -d'{"policy":{"description":"Created by someone","default_state":"hot","ism_template":{"index_patterns":["*ism14d*"],"priority":1},"states":[{"name":"hot","actions":[],"transitions": [{"state_name": "delete","conditions": {"min_index_age": "14d"}}]},{"name": "delete","actions": [{"delete": {}}],"transitions": []}]}}'
curl -u ism:$ISM_PASSWORD -XPUT "$POLS_URL/delete_30d" -H 'Content-Type: application/json' -d'{"policy":{"description":"Created by someone","default_state":"hot","ism_template":{"index_patterns":["*ism30d*"],"priority":2},"states":[{"name":"hot","actions":[],"transitions": [{"state_name": "delete","conditions": {"min_index_age": "30d"}}]},{"name": "delete","actions": [{"delete": {}}],"transitions": []}]}}'
curl -u ism:$ISM_PASSWORD -XPUT "$POLS_URL/delete_180d" -H 'Content-Type: application/json' -d'{"policy":{"description":"Created by someone","default_state":"hot","ism_template":{"index_patterns":["*ism180d*"],"priority":3},"states":[{"name":"hot","actions":[],"transitions": [{"state_name": "delete","conditions": {"min_index_age": "180d"}}]},{"name": "delete","actions": [{"delete": {}}],"transitions": []}]}}'
curl -u ism:$ISM_PASSWORD -XPUT "$POLS_URL/delete_365d" -H 'Content-Type: application/json' -d'{"policy":{"description":"Created by someone","default_state":"hot","ism_template":{"index_patterns":["*ism365d*"],"priority":4},"states":[{"name":"hot","actions":[],"transitions": [{"state_name": "delete","conditions": {"min_index_age": "365d"}}]},{"name": "delete","actions": [{"delete": {}}],"transitions": []}]}}'
```