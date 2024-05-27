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