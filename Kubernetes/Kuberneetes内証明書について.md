## kubeconfig
- `clusters`ブロックの`cluster.certificate-authority-data`はkubernetesクラスターのCAの証明書(公開鍵)
  - api-serverのTLS証明書が信頼できることをPodに保証
- `users`ブロックの`user`配下
  - `client-certificate-data`
    - クライアントの公開鍵証明書。Kubernetes api-serverは、この証明書を使用してクライアントの身元を認証する。この証明書は、通常、証明機関（CA）によって署名され、クライアントの身元を証明する。この証明書にはクライアントの公開鍵が含まれているが、秘密鍵は含まれていない。
  - `client-key-data`
    - クライアントの秘密鍵。この秘密鍵は、クライアントが発行するリクエストの署名に使用され、Kubernetes api-serverは公開鍵証明書を使用してこれらの署名を検証する。
      - 上記のリクエストとは`kubectl get po`など、Kubernetes api-serverに送る各種APIリクエストのこと
- 以下Chat-GPTからの回答
> 1. **クライアント証明書の使用**: クライアントからKubernetes APIサーバーへのリクエストには、`client-certificate-data`（クライアントの公開鍵証明書）が含まれています。この証明書は、KubernetesクラスターのCA（Certificate Authority）によって署名されています。APIサーバーはこの証明書を使用して、リクエストが信頼できるクライアントから来たものであるかを確認します。
>
> 2. **リクエストの署名**: クライアントは自身の秘密鍵（`client-key-data`）を使用してリクエストにデジタル署名を行います。この署名は、リクエストがクライアントによって作成され、途中で改ざんされていないことを保証するためのものです。
>
> 3. **署名の検証**: Kubernetes APIサーバーは、クライアントの公開鍵証明書（`client-certificate-data`）を使用して、リクエストに添付されたデジタル署名を検証します。これにより、リクエストの真正性（クライアントから来たものであること）と、データの整合性（途中で改ざんされていないこと）が確認されます。
>
> 要するに、クライアント証明書（`client-certificate-data`）はクライアントの身元を証明し、クライアントの秘密鍵（`client-key-data`）はリクエストの信頼性を保証します。Kubernetes APIサーバーはこれらの要素を使ってクライアントの認証とリクエストの検証を行います。