- https://langfuse.com/self-hosting/kubernetes-helm

## Helm Chart
- https://github.com/langfuse/langfuse-k8s/tree/main/charts/langfuse

### `values.yaml`の設定例
```yaml
langfuse:
  nodeSelector:
    karpenter.sh/nodepool: arm64-nodepool
    karpenter.sh/capacity-type: on-demand
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
    url: https://langfuse.nutslove.com
  serviceAccount:
    create: true
    name: langfuse-serviceaccount # Pod IdentityでこのServiceAccountに対してS3の権限を与えること
  web:
    service:
      port: 3100 # default portは3000
  additionalEnv:
    - name: AUTH_AZURE_AD_CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: langfuse-auth
          key: azure-client-id
    - name: AUTH_AZURE_AD_CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: langfuse-auth
          key: azure-client-secret
    - name: AUTH_AZURE_AD_TENANT_ID
      valueFrom:
        secretKeyRef:
          name: langfuse-auth
          key: azure-tenant-id
    - name: AUTH_AZURE_ALLOW_ACCOUNT_LINKING
      value: "true"

postgresql:
  deploy: false
  host: lee.cluster-abcdefg1234.ap-northeast-1.rds.amazonaws.com
  auth:
    database: postgres_langfuse # default database name for langfuse（事前にRDSに入って作成しておく必要がある）
    username: postgres
    existingSecret: langfuse-rds-auth
    secretKeys:
      userPasswordKey: password

clickhouse:
  deploy: true
  auth:
    existingSecret: langfuse-auth
    existingSecretKey: clickhouse-password
  persistence:
    enabled: true
    storageClass: "efs-sc"
  extraEnvVars:
  zookeeper:
    persistence:
      enabled: true
      storageClass: "efs-sc"

redis:
  deploy: true
  auth:
    existingSecret: langfuse-auth
    existingSecretPasswordKey: redis-password
  primary:
    persistence:
      storageClass: "auto-ebs-sc"
  replica:
    persistence:
      storageClass: "auto-ebs-sc"

s3:
  deploy: false
  bucket: langfuse-bucket
  region: ap-northeast-1
```

#### PVの設定例
- **EKSバージョンアップ（Blue/Greenデプロイメント）に備えて、ClickHouseとZookeeperのPVとしてEFSを使う場合、PVを事前にLangfuse（厳密にはClickHouse、Zookeeperの） Helmチャート（のStatefulSet）が作成するPVCの名前に合わせて作成しておく必要がある**
- PVのマニフェストファイル例  
  ```yaml
  ## ClickHouse 0
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: clickhouse-efs-pv-0
    labels:
      app: clickhouse
  spec:
    capacity:
      storage: 10Gi
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
    storageClassName: efs-sc
    claimRef:
      namespace: monitoring
      name: data-langfuse-clickhouse-shard0-0
    csi:
      driver: efs.csi.aws.com
      volumeHandle: fs-aaaaaaa::fsap-bbbbb001  # FileSystem::AccessPoint

  ## ClickHouse 1
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: clickhouse-efs-pv-1
    labels:
      app: clickhouse
  spec:
    capacity:
      storage: 10Gi
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
    storageClassName: efs-sc
    claimRef:
      namespace: monitoring
      name: data-langfuse-clickhouse-shard0-1
    csi:
      driver: efs.csi.aws.com
      volumeHandle: fs-aaaaaaa::fsap-bbbbb002  # FileSystem::AccessPoint

  ## ClickHouse 2
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: clickhouse-efs-pv-2
    labels:
      app: clickhouse
  spec:
    capacity:
      storage: 10Gi
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
    storageClassName: efs-sc
    claimRef:
      namespace: monitoring
      name: data-langfuse-clickhouse-shard0-2
    csi:
      driver: efs.csi.aws.com
      volumeHandle: fs-aaaaaaa::fsap-bbbbb003  # FileSystem::AccessPoint

  ## Zookeeper 0
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: zookeeper-efs-pv-0
  spec:
    capacity:
      storage: 10Gi
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
    storageClassName: efs-sc
    csi:
      driver: efs.csi.aws.com
      volumeHandle: fs-aaaaaaa::fsap-bbbbb004 # FileSystem::AccessPoint


  ## Zookeeper 1
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: zookeeper-efs-pv-1
  spec:
    capacity:
      storage: 10Gi
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
    storageClassName: efs-sc
    csi:
      driver: efs.csi.aws.com
      volumeHandle: fs-aaaaaaa::fsap-bbbbb005 # FileSystem::AccessPoint

  ## Zookeeper 2
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: zookeeper-efs-pv-2
  spec:
    capacity:
      storage: 10Gi
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
    storageClassName: efs-sc
    csi:
      driver: efs.csi.aws.com
      volumeHandle: fs-aaaaaaa::fsap-bbbbb006 # FileSystem::AccessPoint
  ```

## EKSバージョンアップ時のEKSクラスター間の移行
### 前提
- ClickHouse、ZookeeperのPVはEFSを使っている
  - 事前にEFSのPVをPVCの名前に合わせて作成しておく
### 手順
> [!NOTE]  
> 旧EKSクラスターからPVとPVCの削除は不要

1. 旧EKSクラスターからLangfuseを削除
2. 新EKSクラスターにClickHouse, Zookeeperの（EFS）PVをデプロイ
3. 新EKSクラスターにLangfuseをデプロイ
4. Langfuse用のALB TargetGroupの`eks:eks-cluster-name`タグの値を新EKSクラスター名に変更
5. 新EKSクラスターに`TargetGroupBinding`をデプロイ