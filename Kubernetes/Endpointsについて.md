# Endpointsについて
- `Service`を作成すると、Kubernetesは自動的にその`Service`に対応する`Endpoints`リソースを作成する
- `Endpoints`リソースは、`Service`がトラフィックをルーティングすべきバックエンドのPodのIPアドレスとポート番号のリストを保持する
  - これにより、`Service`は`Endpoints`を通じて、適切なPodにトラフィックを分散させる(ロードバランシングする)
- 例えば、以下の`Service`リソースがある場合、`app: my-app`を持つすべて(複数)のPodのIPアドレスとポート番号(以下の例では8080)をリストに保持し、`Service`へのトラフィックを分散する
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: my-app-service
  spec:
    selector:
      app: my-app
    ports:
      - port: 80
        targetPort: 8080
  ```

## `selector`なしService
- `selector`を指定しないと自動的に`Endpoints`リソースは作成されず、明示的に`Endpoints`リソースを作成する必要がある。  
  `Endpoints`リソースにk8s外にあるもののIPアドレスを指定すれば、k8s内からServiceのDNSで名前解決/アクセスできる
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  clusterIP: None
---
apiVersion: v1
kind: Endpoints
metadata:
  labels:
    app: nginx
  name: nginx
subsets:
- addresses:
  - ip: 202.232.2.180
```

```shell
/ # nslookup nginx.default.svc.cluster.local
Server:         172.30.0.10
Address:        172.30.0.10:53

Name:   nginx.default.svc.cluster.local
Address: 202.232.2.180
```

# EndpointSliceについて
- `Endpoints`リソースのスケーラビリティとパフォーマンスを向上させるために導入されたリソース
- `Endpoints`リソースは、`Service`に関連するすべてのPodのIPアドレスとポート番号を1つのリソースに保持するが、`EndpointSlice`は複数のスライスに分割して管理する
- これにより、大規模なクラスターでの`Endpoints`リソースの管理が効率化され、APIサーバーへの負荷が軽減される
- `EndpointSlice`は、`Endpoints`と同様に`Service`に関連するPodのIPアドレスとポート番号を保持するが、1つの`EndpointSlice`にはデフォルトで最大100個のエンドポイントしか含まれない（kube-controller-managerの `--max-endpoints-per-slice` フラグで変更可能）
- 例えば、`app: my-app`を持つ200個のPodがある場合、Kubernetesは2つの`EndpointSlice`リソースを作成し、それぞれに100個のエンドポイントを保持する
  - `EndpointSlice`の命名規則は `Service名-ランダム文字列`（例: `frontend-v8n2p`）
- `Endpoints`リソースは互換性のために`EndpointSlice`と一緒に作成されるが、Kubernetes v1.33で`Endpoints` APIは正式に非推奨（deprecated）となっており、将来的には`Endpoints`コントローラーの実行も不要になる方向で進んでいる