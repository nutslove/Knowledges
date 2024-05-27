- ServiceのTypeがデフォルトの`ClusterIP`の場合、デフォルトでは外部からアクセスできないけど、  
  `externalIPs`フィールドでIPアドレスを追加すると、指定したIPアドレスで外部からServiceに紐づいているPodにアクセスできる
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: external-ip
  spec:
    type: ClusterIP
    externalIPs:
      - 10.20.30.20
    ports:
      - name: cluster-port
        protocol: TCP
        port: 8080
        targetPort: 80
    selector:
      app: nginx-dep
  ```

> [!NOTE]
> もちろん`externalIPs`に指定したIPはk8sクラスターへのアクセス性があるIP(e.g. Worker NodeのIP)を指定する必要がある

- `NodePort`ではすべてのWorker NodeのIPからアクセス可能だけど、`externalIPs`にWorker NodeのIPを指定した場合は当然ながら`externalIPs`に指定したWorker NodeのIPからしかアクセスできない
- 参考URL
  - https://qiita.com/dingtianhongjie/items/8f3c320c4eb5cf25d9de