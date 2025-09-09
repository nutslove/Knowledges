> [!NOTE]  
> これらのAPIは *Querier* または *Query Frontend*（9090ポート）に対して実行

> [!NOTE]  
> Multi Tenancyモードを使っている場合は、`THANOS-TENANT`ヘッダーを付けること  
> - 例 
>   ```shell
>   curl -s -H 'THANOS-TENANT: <テナント名>' http://<Querier/Query FrontendのIP>:9090/api/v1/label/__name__/values
>   ```

## メトリクス取得
> [!NOTE]  
> `start`と`end`にはUnix timestamp（秒またはミリ秒）or RFC3339形式（ISO8601）を使うことができる
> - 例1: Unix timestamp（秒単位）  
>   ```shell
>   /api/v1/query_range?query=up&start=1725798000&end=1725801600&step=30
>   ```
> - 例2: Unix timestamp（ミリ秒単位）  
>   ```shell
>   /api/v1/query_range?query=up&start=1725798000000&end=1725801600000&step=30
>   ```
> - 例3: RFC3339形式（ISO8601）  
>   ```shell
>   /api/v1/query_range?query=up&start=2025-09-09T00:00:00Z&end=2025-09-09T01:00:00Z&step=30
>   ```

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