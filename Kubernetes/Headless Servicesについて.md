## Headless Serviceとは
- 特定のPodに直接アクセスするために使用されるService。主にStatefulSetで使われる。
- **通常のServiceはServiceにIPアドレスが割り当てられ、そのServiceの`selector`に指定した(複数の)PodのIPが含まれた`Endpoints`が自動で作成され (その後Podの追加/削除時に自動で`Endpoints`も更新される)、複数のPodにロードバランシングされるが、Headless ServiceはServiceにIPアドレスが割り当てられない。**
  - `selector`の条件に一致するPodのIPアドレスで`Endpoints`が作成されて、Serviceとマッピングされるのは一緒
- `ClusterIP: None`にするとHeadless Serviceになる  
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: my-headless-service
    namespace: poc
  spec:
    clusterIP: None
    selector:
      app: my-app
    ports:
    - port: 80
      targetPort: 8080
  ```

- **`StatefulSet`の`spec.serviceName`にHeadless Serviceの`metadata.name`を指定することで、`<pod名>.<service名>.<namespace名>.svc.cluster.local`での名前解決ができる**  
  → 下の例だと`my-statefulset-0.my-headless-service.poc.svc.cluster.local`でPodのIPアドレスが得られる  

- `<pod名>.<service名>.<namespace名>.svc.cluster.local`で個別のPodのIPアドレスだけ取得することもできるし、  
  普通のServiceと同様に`<service名>.<namespace名>.svc.cluster.local`で紐づいているすべてのPodのIPを取得することもできる

> [!IMPORTANT]  
> **普通のServiceは`<service名>.<namespace名>.svc.cluster.local`でServiceに付いている`Cluster IP`が返ってくるが、  
> Headless Serviceは`<service名>.<namespace名>.svc.cluster.local`で`selector`を満たすすべてのPodのIPアドレスが返ってくる**  
> また、普通のServiceはCluster IPにアクセスするとPodのIPにロードバランシングされるが、  
> Headless ServiceはCluster IPがないため、DNSの名前解決機能だけを使うもの

  ```yaml
  apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: my-statefulset
    namespace: poc
  spec:
    serviceName: my-headless-service
    replicas: 1
    selector:
      matchLabels:
        app: my-app
    template:
      metadata:
        labels:
          app: my-app
      spec:
        containers:
        - name: my-container
          image: my-image:latest
          ports:
          - containerPort: 8080
    volumeClaimTemplates:
    - metadata:
        name: my-volume
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
  ```
