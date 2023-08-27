- index gatewayはLoki v2.3から追加されたコンポーネント
  - https://grafana.com/docs/loki/latest/release-notes/v2-3/  
    > We created an index gateway which takes on the task of downloading the boltdb-shipper index files allowing you to run your queriers without any local disk requirements, this is really helpful in Kubernetes environments where you can return your queriers from Statefulsets back to Deployments and save a lot of PVC costs and operational headaches.
- index gatewayはTSDBでも使われる
  - https://grafana.com/docs/loki/latest/operations/storage/tsdb/

## Index Gatewayを使うメリット
- querierが必要とするindexをquerierの代わりにダウンロード/保持することで、querierをPVを持つStatefulSetではなく、PVを待たないDeploymentsとして動かすことができる
- querierをDeploymentsで動かすことによってよりquerierの拡張性が高くなる
- https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#index-gateway

## Index Gatewayを使う際の注意点
- **index gatewayはPV(EBS)付きでStatefulSetとして動かすこと**
  - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#index-gateway  
  > When using the Index Gateway within Kubernetes, we recommend using a StatefulSet with persistent storage for downloading and querying index files. This can obtain **better read performance**, avoids noisy neighbor problems by not using the node disk, and **avoids the time consuming index downloading step on startup after rescheduling to a new node**. 
- `Queriers`と`Rulers`がIndex Gatewayからindexを取得できるように、`Queriers`と`Rulers`にIndex Gatewayのアドレスを設定する必要がある(gRPC port(9095)で)  
  > To connect Queriers and Rulers to the Index Gateway, set the address (with gRPC port) of the Index Gateway with the `-[tsdb]boltdb.shipper.index-gateway-client.server-address` CLI flag or its equivalent YAML value under StorageConfig.
  - Helmでdistributed chartでデプロイしている場合は`<Helmリソース名>-loki-distributed-index-gateway-headless`でServiceが作成されるので、`tsdb_shipper.index_gateway_client.server_address`に`<Helmリソース名>-loki-distributed-index-gateway-headless.<NameSpace>.svc.cluster.local:9095`を設定
    ~~~yaml
    tsdb_shipper:
      active_index_directory: /var/loki/tsdb-index
      shared_store: s3
      cache_location: /var/loki/tsdb-cache
      index_gateway_client:
        server_address: tsdb-loki-distributed-index-gateway-headless.monitoring.svc.cluster.local:9095 ## initial char(in this case "tsdb") must match helm resource name
    ~~~
    > **Warning**
    > `loki-distributed-index-gateway-headless`(Headless)と`loki-distributed-index-gateway`(ClusterIP)の２つがあるのでHeadlessであってるか確認！(一応検証ではHeadlessで特に問題は見られなかった)