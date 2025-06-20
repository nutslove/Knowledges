# Langfuseのアーキテクチャ
- https://langfuse.com/self-hosting  
![](./image/langfuse_arch_1.jpg)

## 各データの保存先
- Traceは先にS3に保存されて、その後ClickHouseにロードされる
  - https://github.com/orgs/langfuse/discussions/6133
- PromptsはPostgreSQL内に保存される

---

# Install（Self Hosting）
## Docker Compose
- https://langfuse.com/self-hosting/docker-compose

## Helm（k8s）
- **https://github.com/langfuse/langfuse-k8s**
- `langfuse.nextauth.secret.value`には`openssl rand -hex 32`で生成された値を入れる  
  ```shell
  helm install langfuse langfuse/langfuse -n langfuse -f values.yaml
  ```
- ClickHouse、RedisなどのHelmチャートはbitnamiのものを使っている
  - ClickHouseのHelmチャート
    - https://github.com/bitnami/charts/tree/main/bitnami/clickhouse
  - Redis（ValKey）のHelmチャート
    - https://github.com/bitnami/charts/tree/main/bitnami/redis
  - PostgreSQLのHelmチャート
    - https://github.com/bitnami/charts/tree/main/bitnami/postgresql
- common(Podのresource設定など)もbitnamiのHelmチャートを使っている
  - https://github.com/bitnami/charts/tree/main/bitnami/common
- PostgreSQLのPasswordに特殊文字が入っていると以下のようなエラーが出る。Encodingする必要がある。  
  ```shell
  P1013: The provided database string is invalid. invalid port number in database URL. Please refer to the documentation in https://www.prisma.io/docs/reference/database-reference/connection-urls for constructing a correct connection string. In some cases, certain characters must be escaped. Please check the string for any illegal characters.
  ```  
  - 参考URL
    - https://harusame.dev/blog/posts/supabase-prisma-p1013/
    - https://stackoverflow.com/questions/63684133/prisma-cant-connect-to-postgresql

> [!CAUTION]  
> - **ESOでPostgreSQLのPasswordをSecretにしている場合、ESOの`template`に以下を追加してencodingする必要がある**  
>    ```yaml
>    template:
>      type: Opaque
>      data:
>        password: "{{ .password | urlquery }}"
>    ```  
>    - 全体のyaml  
>      ```yaml
>      apiVersion: external-secrets.io/v1beta1
>      kind: ExternalSecret
>      metadata:
>        name: langfuse-rds-auth
>        namespace: monitoring
>      spec:
>        refreshInterval: 1h
>        secretStoreRef:
>          name: aws-secrets-manager
>          kind: ClusterSecretStore
>        target:
>          name: langfuse-rds-auth
>          creationPolicy: Owner
>          template:
>            type: Opaque
>            data:
>              password: "{{ .password | urlquery }}"
>        data:
>        - secretKey: password
>          remoteRef:
>            key: <AWSのSecretManager名>
>            property: password
>      ``` 

---

# S3
- S3のリージョン(Region)を指定しないと、langfuse-webから以下のようなエラーが出る場合がある  
  ```shell
  error  Failed to upload event to S3
  Failed to upload JSON to S3
  getaddrinfo ENOTFOUND s3.auto.amazonaws.com
  ```
  - Helmの場合、`values.yaml`で`s3.region`でリージョンを指定できる  
    ```yaml
    s3:
      deploy: false
      bucket: <S3バケット名>
      region: ap-northeast-1
    ```

---

# Redisについて
- Redisが再起動されてもeventが消失されることはないらしい（以下LangfuseのAsk AIの回答）
  > Redis/Valkey serves as a caching layer and queue in Langfuse's architecture [(1)](https://langfuse.com/self-hosting/infrastructure/cache). It has several key roles:
  >
  > 1. Event Processing:
  >     - Accepts new events quickly on the API [(1)](https://langfuse.com/self-hosting/infrastructure/cache)
  >     - Defers processing and insertion of these events [(1)](https://langfuse.com/self-hosting/infrastructure/cache)
  >     - Helps handle request peaks gracefully [(1)](https://langfuse.com/self-hosting/infrastructure/cache)
  >
  > 2. Performance Optimization:
  >     - Caches API keys in-memory to reduce database load [(2)](https://langfuse.com/self-hosting)
  >     - Provides caching for prompts, allowing fast retrieval without hitting the database [(2)](https://langfuse.com/self-hosting)
  >     - Acts as a read-through cache for prompts [(2)](https://langfuse.com/self-hosting)
  >
  > 3. Queue Management:
  >     - Used for queuing event metadata for worker processing [(3)](https://github.com/orgs/langfuse/discussions/5924)
  >
  > Regarding cache loss during restart:
Langfuse is designed to handle this scenario safely. The system uses S3/Blob Storage to persist all incoming events first, and only after successful processing are they written to the database [(2)](https://langfuse.com/self-hosting). This ensures that even if Redis becomes temporarily unavailable, events are not lost and can be processed later [(2)](https://langfuse.com/self-hosting).
  >
  > For optimal operation, Redis should be configured with `maxmemory-policy=noeviction` to ensure queue jobs are not evicted from the cache [(1)](https://langfuse.com/self-hosting/infrastructure/cache). 

---

# LangGraphとLangfuseの連携
- https://langfuse.com/docs/integrations/langchain/example-python-langgraph

---

# Trace IDの設定
- UUIDを使って利用者側でTraceIDを指定することもできる
  - これを使えば連続しない処理も同じTraceIDを指定することでトレースを連結させることができる
- https://langfuse.com/docs/tracing-features/trace-ids
