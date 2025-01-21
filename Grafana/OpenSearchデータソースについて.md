# DataSource登録
- https://grafana.com/grafana/plugins/grafana-opensearch-datasource/
## AWS Managed OpenSearch
- OpenSearchを内部データベースの「マスターユーザーの作成」で作成した場合、Grafana DataSourceで「SigV4 auth」ではなく、「**Basic auth**」を選択してユーザ名とパスワードを入力すること  
  ![](./image/opensearch_auth.jpg)  
  ![](./image/opensearch_auth_2.jpg)
- 対象indexを指定する「OpenSearch details
」ブロックの「Index name」フィールドには`*`や`security-*`などワイルドカードが使える  
  ![](./image/opensearch_datasource_for_log_2.jpg)

# GrafanaでOpenSearch上のログをログとして（メトリクスとかではなく）確認する方法
- DataSourceの設定で「Logs」ブロックの「Message field name」にログ本文が入っているフィールド名を入力。「Level field name」は必要に応じて入力(optional)  
  ![](./image/opensearch_datasource_for_log_1.jpg)  
  - ログ本文が入っているfield名はOpenSearch Dashboardなどで確認
- ExploreやDashboard設定でDataSourceとしてOpenSearchを選択後、Query Typeに「**Lucene**」を選択し、表示形式として**Logs**または**Raw Data**を選択後、クエリーに`*`や`__index:"<index名>"`など入力して実行  
  ![](./image/opensearch_datasource_for_log_3.jpg)