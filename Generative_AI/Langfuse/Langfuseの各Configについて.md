# Langfuse on Kubernetes(Helm)
- https://github.com/langfuse/langfuse-k8s
- **`values.yaml`内の各項目の説明**
  - https://github.com/langfuse/langfuse-k8s/blob/main/charts/langfuse/README.md

- postgresとS3はAWSサービスを使って、RedisとClickHouseはPodとしてデプロイする`values.yml`の設定例  
  ```yaml
  langfuse: 
    encryptionKey: # `openssl rand -hex 32`で生成
      secretKeyRef:
        name: langfuse-auth
        key: encryptionKey
    salt: # `openssl rand -base64 32`で生成
      secretKeyRef:
        name: langfuse-auth
        key: salt
    nextauth: # `openssl rand -base64 32`で生成
      secret:
        secretKeyRef:
          name: langfuse-auth
          key: nextauth-secret
    serviceAccount:
      create: true
      name: langfuse-serviceaccount # Pod IdentityでこのServiceAccountに対してS3の権限を与えること

  postgresql:
    deploy: false
    host: manual-aurora-pg-serverless.cluster-xxxxxx.ap-northeast-1.rds.amazonaws.com
    auth:
      database: postgres_langfuse # default database name for langfuse（事前にRDSに入って作成しておく必要がある）
      username: postgres
      existingSecret: langfuse-auth
      secretKeys:
        userPasswordKey: rds-password

  clickhouse:
    deploy: true
    auth:
      existingSecret: langfuse-auth
      existingSecretKey: clickhouse-password

  redis:
    deploy: true
    auth:
      existingSecret: langfuse-auth
      existingSecretPasswordKey: redis-password

  s3:
    deploy: false
    bucket: sandbox-langfuse-bucket
  ```
  - External Secret  
    ```yaml
    apiVersion: external-secrets.io/v1beta1
    kind: ExternalSecret
    metadata:
      name: langfuse-auth
      namespace: monitoring
    spec:
      refreshInterval: 1h
      secretStoreRef:
        name: aws-secrets-manager
        kind: ClusterSecretStore
      target:
        name: langfuse-auth
        creationPolicy: Owner
      data:
      - secretKey: salt
        remoteRef:
          key: sandbox/langfuse
          property: salt
      - secretKey: nextauth-secret
        remoteRef:
          key: sandbox/langfuse
          property: nextauth-secret
      - secretKey: encryptionKey
        remoteRef:
          key: sandbox/langfuse
          property: encryptionKey
      - secretKey: rds-password
        remoteRef:
          key: sandbox/langfuse
          property: rds-password
      - secretKey: redis-password
        remoteRef:
          key: sandbox/langfuse
          property: redis-password
      - secretKey: clickhouse-password
        remoteRef:
          key: sandbox/langfuse
          property: clickhouse-password
    ```