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