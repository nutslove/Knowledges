# Cephとは
- オープンソースの分散ストレージシステム
- **オブジェクトストレージ**、**ブロックストレージ**、**ファイルシステム**の3つの主要なストレージインターフェイスを提供している
- **Ceph stores data as objects within logical storage pools. Using the CRUSH algorithm, Ceph calculates which placement group (PG) should contain the object, and which OSD should store the placement group. The CRUSH algorithm enables the Ceph Storage Cluster to scale, rebalance, and recover dynamically.  
  A Ceph Storage Cluster requires the following: at least one Ceph Monitor and at least one Ceph Manager, and at least as many Ceph Object Storage Daemons (OSDs) as there are copies of a given object stored in the Ceph cluster (for example, if three copies of a given object are stored in the Ceph cluster, then at least three OSDs must exist in that Ceph cluster).  
  The Ceph Metadata Server is necessary to run Ceph File System clients.**
  - https://docs.ceph.com/en/reef/start/intro/

# Cephのアーキテクチャ
- https://docs.ceph.com/en/latest/architecture/
  ![](./image/ceph_architecture.jpg)
  ![](./image/ceph_architecture2.jpg)

## RADOS（Reliable Autonomous Distributed Object Store）
- The core storage layer in Ceph
- Cephの基盤となるObject Storageシステム。データの分散、レプリケーション、回復を自動的に行う。
- 複数の`OSDs（object storage daemons）`で構成されている
  - **各`OSD`は独立していて、各`OSD`は一般的に１つのディスクと１対１でマッピングされている**

## Librados
- Ceph Storage Clusterと通信するためのライブラリ
- RADOS Gateway、RBDなどの上位レイヤーのサービスで使用される
- プログラミング言語でライブラリを使って直接RADOSと通信することもできる

## RADOS Gateway
- Object Storageのインターフェースを提供し、S3やOpenStack Swift互換のAPIを通じてアクセスできる

## RDB（RADOS Block Device）
- ブロックストレージインターフェースを提供
- 仮想マシン、コンテナ、ベアメタルサーバーなどにブロックデバイスを提供

## CephFS
- 分散ファイルシステムで、POSIX互換のインターフェースを提供
- ファイルやディレクトリの階層構造を持ち、複数のクライアントが同時にアクセスできる

# Ceph Storage Cluster
- https://docs.ceph.com/en/latest/glossary/#term-Ceph-Storage-Cluster
  > The collection of **Ceph Monitors**, **Ceph Managers**, **Ceph Metadata Servers**, and **OSDs** that work together to store and replicate data for use by applications, Ceph Users, and **Ceph Clients**. Ceph Storage Clusters receive data from Ceph Clients.
- Ceph Storage Clusterは以下のデーモンで構成されている

## `OSD（Object Storage Daemon）`
- データの保管と複製を担当するデーモン
- データの分散、レプリケーション、リカバリ、リバランスを自動的に処理
- Process(`ceph-osd`)とそれに紐づいてるStorageのセットで考えることもできる
> An Object Storage Daemon (Ceph OSD, ceph-osd) stores data, handles data replication, recovery, rebalancing, and provides some monitoring information to Ceph Monitors and Managers by checking other Ceph OSD Daemons for a heartbeat. At least three Ceph OSDs are normally required for redundancy and high availability.
- OSDはストレージデバイス (ハードディスクまたはその他のブロックデバイスなど) をCeph Storageクラスターに接続する。個々のストレージサーバーは複数のOSDデーモンを実行し、クラスターに複数のOSDを提供することができる。RADOS内にデータを保存する*BlueStore*という機能をサポートしていて、BlueStoreはローカルストレージデバイスを raw モードで使用し、高いパフォーマンスを実現するように設計されている。
OSD 操作の 1 つの設計目標は、計算能力をできる限り物理データに近付けて、クラスターが最高の効率で動作できるようにすること。Ceph クライアントと OSD デーモンはどちらも、中央ルックアップテーブルに依存するのではなく、Controlled Replication Under Scalable Hashing (CRUSH) アルゴリズムを使用して、オブジェクトの場所に関する情報を効率的に計算する。

## `Ceph Monitor（MON）`
- クラスタの状態を監視し、管理する役割を担う
- Cluster Mapの保持/管理、状態変更の検知、クライアントへの認証を行う
  - Cluster Mapは、クラスターの状態と設定に関する情報を含む5つのMapのコレクション
  - Cephは各クラスターイベントに対応し、適切なMapを更新し、更新されたMapを各Monitorデーモンに複製する必要がある
  - **更新を適用するには、MONでクラスターの状態に関する合意が確立される必要がある。設定されているMONの過半数が利用可能で、Mapの更新に同意する必要がある。奇数のMONで Cephクラスターを設定し、MONがクラスターの状態に投票するときにクォーラムを確立できるようにする。Ceph Storageクラスターを動作させ、アクセスできるようにするには、設定されたMONの半分以上が機能している必要がある。**
- 通常は奇数個（3、5、7など）で構成し、高可用性を確保
- MON is a daemon process(`ceph-mon`) that communicates with peer MONs, OSDs, and users, maintaing and distributing various information vital to cluster operations.
- 以下のデータを管理する
  - maps of OSDs
  - other MONs
  - PGs
  - CRUSH map（which describes where data should be placed and found）
> A Ceph Monitor (`ceph-mon`) maintains maps of the cluster state, including the monitor map, manager map, the OSD map, the MDS map, and the CRUSH map. These maps are critical cluster state required for Ceph daemons to coordinate with each other. Monitors are also responsible for managing authentication between daemons and clients. At least three monitors are normally required for redundancy and high availability.

## `Ceph Manager（ceph-mgr）`
- クラスタの監視とレポート生成を行う
- ランタイムメトリクスを追跡し、ブラウザーベースのダッシュボードとREST API/CLIを介してクラスター情報を公開する
> A Ceph Manager daemon (ceph-mgr) is responsible for keeping track of runtime metrics and the current state of the Ceph cluster, including storage utilization, current performance metrics, and system load. The Ceph Manager daemons also host python-based modules to manage and expose Ceph cluster information, including a web-based Ceph Dashboard and REST API. At least two managers are normally required for high availability.

## `Ceph Metadata Server`
- CephFSを使う時のみ必要なコンポーネント
- ファイルシステムのメタデータ（例: ファイル名、ディレクトリ構造、アクセス許可等）を管理
- メタデータを高速にアクセスできるようにメモリ内にキャッシュし、必要に応じてディスクに保存
> A Ceph Metadata Server (MDS, ceph-mds) stores metadata for the Ceph File System. Ceph Metadata Servers allow CephFS users to run basic commands (like ls, find, etc.) without placing a burden on the Ceph Storage Cluster.

## CRUSH（Controlled Replication Under Scalable Hashing）
- Cephの分散ストレージアーキテクチャにおいて、データの配置を決定するアルゴリズム
- Cephクラスタ内でデータの複製とストライピングを効率的に管理し、スケーラビリティ、パフォーマンス、および耐障害性を提供
- 従来のデータセンターのストレージソリューションでは、データの配置を中央の管理ノードが決定していたが、CRUSHは分散型のアプローチを採用している

## CRUSH map
- https://docs.ceph.com/en/reef/rados/operations/crush-map/
> The CRUSH algorithm computes storage locations in order to determine how to store and retrieve data. CRUSH allows Ceph clients to communicate with OSDs directly rather than through a centralized server or broker. By using an algorithmically-determined method of storing and retrieving data, Ceph avoids a single point of failure, a performance bottleneck, and a physical limit to its scalability.
>
> CRUSH uses a map of the cluster (the CRUSH map) to map data to OSDs, distributing the data across the cluster in accordance with configured replication policy and failure domains. For a detailed discussion of CRUSH, see CRUSH - Controlled, Scalable, Decentralized Placement of Replicated Data
>
> CRUSH maps contain a list of OSDs and a hierarchy of “buckets” (hosts, racks) and rules that govern how CRUSH replicates data within the cluster’s pools. By reflecting the underlying physical organization of the installation, CRUSH can model (and thereby address) the potential for correlated device failures. Some factors relevant to the CRUSH hierarchy include chassis, racks, physical proximity, a shared power source, shared networking, and failure domains. By encoding this information into the CRUSH map, CRUSH placement policies distribute object replicas across failure domains while maintaining the desired distribution. For example, to address the possibility of concurrent failures, it might be desirable to ensure that data replicas are on devices that reside in or rely upon different shelves, racks, power supplies, controllers, or physical locations.
>
> When OSDs are deployed, they are automatically added to the CRUSH map under a host bucket that is named for the node on which the OSDs run. This behavior, combined with the configured CRUSH failure domain, ensures that replicas or erasure-code shards are distributed across hosts and that the failure of a single host or other kinds of failures will not affect availability. For larger clusters, administrators must carefully consider their choice of failure domain. For example, distributing replicas across racks is typical for mid- to large-sized clusters.

## Pool
- Cephクラスタ内のデータを論理的に区分けするための領域（データのグループ化）
  - 例えば、異なるアプリケーションのデータや異なるパフォーマンス要件を持つデータを別々のプールに格納できる
- 各poolごとに、独自のCRUSH Rule（データがどのOSD（Object Storage Daemon）に配置されるかを制御）やレプリケーション数、PGの数などの設定が割り当てられる

## PG (placement group)
- Objectを保存する論理的な単位
- PoolはPGの集合
- １つのPGは複数のOSDにマッピングされる
  - 通常、PGの数はOSDの数の100倍程度が推奨されている
- **PoolとPG、OSDのマッピングの例**
  - Poolの設定とOSD数
    - Pool A: レプリケーション数3、PG数128
    - OSD数: 10
  - 上記の場合、Pool Aの各PGは3つの異なるOSDにマッピングされる。具体的には、次のようなマッピングが行われる
    - PG 0: OSD 1, OSD 4, OSD 7
    - PG 1: OSD 2, OSD 5, OSD 8
    - PG 2: OSD 3, OSD 6, OSD 9
    - （以下略）
- 主な役割/機能
  - **データの分散**
    - PGはオブジェクトを複数のOSDに分散して保存する。これによりデータの負荷分散と並列処理が可能になる。
  - **レプリケーション**
    - 各PGは複数のOSDにレプリケートされ、障害時のデータ可用性を確保する。レプリカの数はプールごとに設定できる。
    - defaultではCephは３つの異なるOSDに各PGの複製を配置して管理する
      - １つがPrimaryとして指定され、残りはsecondariesとなる。  
        **すべてのClientがRead/WriteするのはPrimaryのOSD**
  - **スケーラビリティ**
    - PGの数を増やすことで、クラスタ全体のI/Oパフォーマンスとストレージ容量を向上できる。
  - **障害からの復旧**
    - OSDが故障した際、PGはデータの再同期とリバランスを行い、データの整合性を維持する。
- https://access.redhat.com/documentation/en-us/red_hat_ceph_storage/5/html/storage_strategies_guide/placement_groups_pgs
  ![](./image/PG.jpg)
- https://docs.ceph.com/en/reef/dev/placement-group/
> When placing data in the cluster, objects are mapped into PGs, and those PGs are mapped onto OSDs. We use the indirection so that we can group objects, which reduces the amount of per-object metadata we need to keep track of and processes we need to run (it would be prohibitively expensive to track eg the placement history on a per-object basis). Increasing the number of PGs can reduce the variance in per-OSD load across your cluster, but each PG requires a bit more CPU and memory on the OSDs that are storing it. We try and ballpark it at 100 PGs/OSD, although it can vary widely without ill effects depending on your cluster. You hit a bug in how we calculate the initial PG number from a cluster description.
>
> There are a couple of different categories of PGs; the 6 that exist (in the original emailer’s ceph -s output) are “local” PGs which are tied to a specific OSD. However, those aren’t actually used in a standard Ceph configuration.

## Cephのインストール
- https://docs.ceph.com/en/latest/install/
- 普通のサーバにデプロイする時は`Cephadm`で、Kubernetes上にデプロイする時は`Rook`を使うのが一般的？

## CephのリリースとEOL
- https://docs.ceph.com/en/latest/releases/index.html