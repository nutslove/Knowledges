> [!NOTE]  
> これらのAPIは *Querier* または *Query Frontend*（9090ポート）に対して実行

> [!NOTE]  
> Multi Tenancyモードを使っている場合は、`THANOS-TENANT`ヘッダーを付けること  
> - 例 
>   ```shell
>   curl -s -H 'THANOS-TENANT: <テナント名>' http://<Querier/Query FrontendのIP>:9090/api/v1/label/__name__/values
>   ```

## メトリクス取得
```shell
curl -s "http://<Querier/Query FrontendのIP>:9090/api/v1/query_range?query=<PromQL>&start=<UNIX TIMESTAMP>&end=<UNIX TIMESTAMP>&step=<DataPointの間隔(秒)>"
```
- 例  
  ```shell
  curl -s "http://10.1.3.88:9090/api/v1/query_range?query=up\{\"job\"=\"kube-state-metrics\"\}&start=1756695030&end=1756698630&step=600"
  ```

## メトリクス一覧取得
```shell
curl -s http://<Querier/Query FrontendのIP>:9090/api/v1/label/__name__/values
```

## ラベル一覧
```shell
curl -s http://<Querier/Query FrontendのIP>:9090/api/v1/labels
```

## 特定のメトリクスが持っているラベル一覧（ラベルと値のすべての組み合わせ）
```shell
curl -s http://<Querier/Query FrontendのIP>:9090/api/v1/series?match[]=<対象メトリクス名>
```

## 特定ラベルの値一覧
```shell
curl -s http://<Querier/Query FrontendのIP>:9090/api/v1/label/<ラベル名>/values
```