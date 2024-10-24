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
  - **Cluster UUIDは最初にクラスタリング時にCluster manager（master）eligibleに選出されたmasterノードによってアサインされて、各ノードのディスクの保存される。**  
    Cluster manager（master）eligibleが選出されるまではUUIDは払い出されない。
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
- https://kazuhira-r.hatenablog.com/entry/2019/11/17/234315
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

## クラスターの状態確認
- クラスターに参加したコンポーネントの一覧が確認できる
  - `<OpenSearch(またはClientノード)のエンドポイント>/_cat/nodes?v`
- clusterの状態確認
  - `<OpenSearch(またはClientノード)のエンドポイント>/_cluster/health?pretty`
- Disk使用率
  - `<OpenSearch(またはClientノード)のエンドポイント>/_cat/allocation?v`
- indexごとの情報
  - `<OpenSearch(またはClientノード)のエンドポイント>/_cat/indices?v&s=index:asc`
    - `s=xx:asc` はsort指定。`s=<sortしたい項目名>:<昇順/降順>`
      - 昇順：`asc`
      - 降順：`desc`
- shardごとの情報
  - `<OpenSearch(またはClientノード)のエンドポイント>/_cat/shards?v`
    - `state=STARTED` が正常。`UNASSIGNED`があれば、data-Nodeへの割り当てがされていない状態
    - `prirep` = `p`:primary `r`:replica(secondary)
    - `r`が`UNASSIGNED`の状態でPrimary ShardがLostすると該当indexのデータは復旧できなくなる
- UNASSIGNEDの理由を表示
  - `<OpenSearch(またはClientノード)のエンドポイント>/_cat/shards?h=index,shard,state,prirep,unassigned.reason`
    - `h=`でヘッダー(出力する項目)を指定
  - `unassigned.reason`
    | reason | |
    | --- | --- |
    |INDEX_CREATED | インデックスを作成するためのAPIで問題が発生 |
    |CLUSTER_RECOVERED | クラスターに対して完全なデータ復元が実行された |
    |INDEX_REOPENED | インデックスが有効または無効になっている |
    |DANGLING_INDEX_IMPORTED | dangling indexの結果がインポートされていない |
    |NEW_INDEX_RESTORED | データがスナップショットから新しいインデックスに復元されている |
    |EXISTING_INDEX_RESTORED | データがスナップショットから無効なインデックスに復元された |
    |REPLICA_ADDED | レプリカ シャードが明示的に追加された |
    |ALLOCATION_FAILED | シャードの割り当てが失敗した |
    |NODE_LEFT | シャードが割り当てられていたノードがクラスターから除外された |
    |REINITIALIZED | シャードの移動からシャードの初期化までのプロセスに誤った操作 (シャドウ レプリカ シャードの使用など) が存在する |
    |REROUTE_CANCELLED | ルーティングが明示的にキャンセルされたために割り当てがキャンセルされた |
    |REALLOCATED_REPLICA | より適切なレプリカの場所が使用され、既存のレプリカの割り当てがキャンセルされる。その結果、シャードは割り当てられない |

## Network関連設定
- https://opensearch.org/docs/latest/install-and-configure/configuring-opensearch/network-settings/
- OpenSearchのネットワークには、**OpenSearch内部でノード間の通信に使われる`transport`（デフォルトのPort番号:`9300`）** と **外部(クライアント)からの通信のための`http`（REST Layer、デフォルトのPort番号:`9200`）** の２つがある
#### **`transport.publish_host`**
- クラスタリングに使うIPアドレスを明示的に指定（他のノードに私に接続するときこのIPアドレスでアクセスしてねって広報し、他のノードはこのIPアドレスでアクセスしてくる）  
  > Specifies an address or addresses that an OpenSearch node publishes to other nodes for transport communication.
#### **`transport.port`**
- クラスタリングに使うPort番号を明示的に指定（デフォルトは`9300`ポート）

## Helmチャート
- https://github.com/opensearch-project/helm-charts
- Helm chartのvaluesに`roles`があって、`master`（`cluster_manager`）,`ingest`,`data`,`remote_cluster_client`が指定できる
  - `roles`に複数のロール(e.g. `ingest`と`data`)を指定することもできる
  - https://github.com/opensearch-project/helm-charts/blob/0c910008bbdaa66f200307864d3b9466c2864319/charts/opensearch/values.yaml#L16
  - `ingest`はStatelessで、`master`と`data`がStateful