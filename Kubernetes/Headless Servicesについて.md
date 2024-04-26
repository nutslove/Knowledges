## Headless Serviceとは
- 特定のPodに直接アクセスするために使用されるService。主にStatefulSetで使われる。
- 通常のServiceはServiceにIPアドレスが割り当てられ、そのServiceに紐づいている複数のEndpointsにロードバランシングされるが、Headless ServiceはServiceにIPアドレスが割り当てられない。
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
- `StatefulSet`の`spec.serviceName`にHeadless Serviceの`metadata.name`を指定することで、`<pod名>.<service名>.<namespace名>.svc.cluster.local`での名前解決ができる  
  → この例だと`my-statefulset-0.my-headless-service.poc.svc.cluster.local`でPodのIPアドレスが得られる
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
