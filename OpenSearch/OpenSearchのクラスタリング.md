## 前提
- OpenSearchノードには (1)`cluster_manager` (旧名称`master`)、(2)`ingest`、(3)`data`のノードType(role)がある
  - それ以外にも`search`、`ml`、`remote_cluster_client`、`coordinating_only`もある（用途は要確認）
  - https://opensearch.org/docs/latest/install-and-configure/configuring-opensearch/configuration-system/
- 各ノードType(role)
  - **https://opensearch.org/docs/latest/tuning-your-cluster/**
  - **Cluster manager**  
    > Manages the overall operation of a cluster and keeps track of the cluster state. This includes creating and deleting indexes, keeping track of the nodes that join and leave the cluster, checking the health of each node in the cluster (by running ping requests), and allocating shards to nodes.
    - **Cluster manager eligible**  
      - cluster_manager（master）ノードのうち、1つをCluster manager eligibleノードとして選出する
      > Elects one node among them as the cluster manager node through a voting process.
  - **Ingest**  
    > Pre-processes data before storing it in the cluster. Runs an ingest pipeline that transforms your data before adding it to an index.
  - **Data**
    > Stores and searches data. Performs all data-related operations (indexing, searching, aggregating) on local shards. These are the worker nodes of your cluster and need more disk space than any other node type.
- OpenSearchはデフォルトではmulti-node clusterとして起動(構築)される。
  - single-node mode(1つのノードがすべてのroleを担う)として動かす時は`discovery.type`パラメータに`single-node`を指定する必要がある
  - https://opensearch.org/docs/latest/install-and-configure/configuring-opensearch/discovery-gateway-settings/
- **クラスタリングには「_Cluster UUID_」が識別子として使われ、異なるUUIDを持つノードはクラスターに参加できない。**
  - **Cluster UUIDは最初にクラスタリング時にCluster manager（master）eligibleに選出されたmasterノードによってアサインされて、各ノードのディスクの保存される**  
    > The cluster UUID is assigned by the elected master node when the cluster first forms, and is stored on disk on each node.

    https://discuss.elastic.co/t/data-node-s-cluster-uuid-diffrent-from-master-nodes-cluster-uuid/196737


### 用語の変更
- https://opensearch.org/blog/Adopting-inclusive-language-across-OpenSearch/
- 変更された用語
  - `master` -> `cluster manager`
  - `blacklist` -> `deny list`
  - `whitelist` -> `allow list`
- 上記に伴い、パラメータ名やendpointなども変更された。詳細は上記URLを参照。

## クラスタリングのためのparameter / config
#### **`discovery.seed_hosts`**
- OpenSearchクラスターに参加するために`Cluster manager（master） eligible`ノードを「IPアドレス」or「IPアドレス:ポート番号」or「FQDN」の形式でリストとして指定する
- `cluster_manager`（`master`）だけではなく、`ingest`や`data`ノードでも指定する必要がある
- https://opensearch.org/docs/latest/tuning-your-cluster/#step-4-configure-discovery-hosts-and-initial-cluster-manager-nodes-for-a-cluster
- https://www.elastic.co/guide/en/elasticsearch/reference/current/important-settings.html#unicast.hosts

#### **`cluster.initial_cluster_manager_nodes`** (旧:`cluster.initial_master_nodes`)
- `Cluster manager eligible`ノード候補となる`cluster_manager`（`master`）のリストを指定
- **IPアドレス/hostname/FQDNではなく、node nameを指定する必要がある**  
  > you need to configure the discovery hosts and specify the cluster manager nodes for the initial cluster election. **Note that this is the node name and not the IP Address, hostname, or fully-qualified hostname.**
- *ノードType(role)が`cluster_manager`（`master`）の時のみ必要*
- https://opensearch.org/docs/latest/tuning-your-cluster/#step-4-configure-discovery-hosts-and-initial-cluster-manager-nodes-for-a-cluster
- https://www.elastic.co/guide/en/elasticsearch/reference/current/important-settings.html#initial_master_nodes

#### **``**
-

## Helmチャート
- https://github.com/opensearch-project/helm-charts
- Helm chartのvaluesに`roles`があって、`master`（`cluster_manager`）,`ingest`,`data`,`remote_cluster_client`が指定できる
  - `roles`に複数のロール(e.g. `ingest`と`data`)を指定することもできる
  - https://github.com/opensearch-project/helm-charts/blob/0c910008bbdaa66f200307864d3b9466c2864319/charts/opensearch/values.yaml#L16
  - `ingest`はStatelessで、`master`と`data`がStateful