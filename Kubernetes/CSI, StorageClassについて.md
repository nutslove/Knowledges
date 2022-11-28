## CSI (Container Storage Interface)
- Container Orchestration向け (k8sだけではなくMesosなど別のToolも使える) の標準化されたストレージインターフェース
- CSIが登場する以前は、Kubernetesのストレージ関連の実装はk8s自身のソースコードに直接書かれていたのでストレージ機能を実装するストレージベンダ等は、Kubernetesのソースコードへアップストリームする必要があったけど、CSIが登場してからはInterfaceに沿って実装すれば誰でもk8sのストレージを提供できるようになった
- CSIに沿って実装された外部ストレージを使うためにはStorage ClassとCSI Driverをデプロイする必要がある
  - CSI Driverは通常コンテナイメージ(Pod)として提供される
  ![CSI Driver](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/CSIDriver.jpg)
- 参考URL
  - https://thinkit.co.jp/article/17635
  - https://access.redhat.com/documentation/ja-jp/openshift_container_platform/4.2/html/storage/persistent-storage-using-csi
  - https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/install-csi-driver?hl=ja
- CSIのGithub Document
  - https://kubernetes-csi.github.io/docs/

#### EBS-CSI
- Dynamic Volume ProvisioningとしてEBSを使うためにはebs-csi-driverのデプロイが必要
- helmでデプロイすることも可能
  - https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/install.md
  - デプロイ後`kube-system`namespaceで`ebs-csi-controller`と`ebs-csi-node`が動いていることを確認 

## StorageClass
- StorageClassとは
  > 스토리지클래스는 관리자가 제공하는 스토리지의 "classes"를 설명할 수 있는 방법을 제공한다. 다른 클래스는 서비스의 품질 수준 또는 백업 정책, 클러스터 관리자가 정한 임의의 정책에 매핑될 수 있다. 쿠버네티스 자체는 클래스가 무엇을 나타내는지에 대해 상관하지 않는다.
  - StorageClassはストレージの種類を示すオブジェクト
  - Dynamic Volume ProvisioningにStorageClassが必要
    - https://kubernetes.io/ja/docs/concepts/storage/dynamic-provisioning/
  - AWS EFSなど、特定のProviderが提供するStorageをVolumeとして使うために必要なもの
  - `provisioner`は必須でどこから提供されるどのようなストレージなのか(提供元)を指定する
  - `kind: StorageClass`で指定した`metadata.name`名と`PersistentVolume`の`storageClassName`を合せる必要がある
    - 例
      ~~~yaml
      ---
      kind: StorageClass
      apiVersion: storage.k8s.io/v1
      metadata:
        name: efs-sc -------------------------→ ここの名前
      provisioner: efs.csi.aws.com
      ---
      apiVersion: v1
      kind: PersistentVolume
      metadata:
        name: efs-pv1
        namespace: monitoring
      spec:
        capacity:
          storage: 5Gi
        volumeMode: Filesystem
        accessModes:
          - ReadWriteOnce
        persistentVolumeReclaimPolicy: Retain
        storageClassName: efs-sc -------------→ ここの名前
        csi:
          driver: efs.csi.aws.com
          volumeHandle: fs-0b0725443cb825a46:/vmstorage-1      
      ~~~
- 参考URL
  - https://kubernetes.io/ko/docs/concepts/storage/storage-classes/
  - https://cstoku.dev/posts/2018/k8sdojo-12/
