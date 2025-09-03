- https://langfuse.com/self-hosting/troubleshooting#clickhouse-handling-failed-migrations

# ■ Langfuse Webで "no migration found for version xx" が出て、Langfuse Web（Pod）が起動しない	
- エラーログ全文  
  ```
  error: no migration found for version 23: read down for version 23 .: file does not existerror: no migration found for version 23: read down for version 23 .: file does not exist"
  ```
### 原因
- 不明・・・。

### 対処方法
- 不明・・・。一旦Langfuse一式を削除（RDSとS3は残したまま）して、再度デプロイしたら問題なく起動した。おそらく最初にHelm CLIでデプロイした後にArgoCDから更新したせいで不整合が発生したかも・・・？

---

# ■ Langfuse Webで "error: Dirty database version xx. Fix and force version." が出て、Langfuse Web（Pod）が起動しない

- https://langfuse.com/faq/all/self-hosting-clickhouse-handling-failed-migrations

### 原因
- https://github.com/langfuse/langfuse/issues/6679

### 対処方法
- ClickHouse Podに入り、`clickhouse-client`コマンドで（defaultユーザのPW入力して）DBに接続する
- 以下のコマンドでDB一覧、Table一覧、`schema_migrations`テーブル内のDirtyなレコードを確認する  
  ```sql
  SHOW DATABASES;
  USE default;
  SHOW TABLES;
  SELECT * FROM schema_migrations;
  ```
- 以下のSQLでDirtyなレコードを削除する（`version`は自分の環境に合わせて変更すること）  
  ```sql
  DELETE FROM schema_migrations WHERE version = 23;
  ```

- 私の場合、上記でも解決せず、追加で以下の対応を行った
  - すべてのdirty（dirtyが1の）行を削除  
    ```
    DELETE FROM schema_migrations WHERE dirty = 1;
    ```
  - その後、langfuse-web podを削除（再起動）したら、今回は以下のようなエラーが出た  
    ```shell
    Script executed successfully.
    Prisma schema loaded from packages/shared/prisma/schema.prisma
    Datasource "db": PostgreSQL database "postgres_langfuse", schema "public" at "prod-rcallm-cluster-ap-northeast-1.cluster-cguwy7lk6qzo.ap-northeast-1.rds.amazonaws.com"
    338 migrations found in prisma/migrations
    No pending migrations to apply.
    error: migration failed in line 0: CREATE TABLE dataset_run_items_rmt ON CLUSTER default (
        -- primary identifiers
        `id` String,
        `project_id` String,
        `dataset_run_id` String,
        `dataset_item_id` String,
        `dataset_id` String,
        `trace_id` String,
        `observation_id` Nullable(String),
        -- error field
        `error` Nullable(String),
         -- timestamps
        `created_at` DateTime64(3) DEFAULT now(),
        `updated_at` DateTime64(3) DEFAULT now(),
        -- denormalized immutable dataset run fields
        `dataset_run_name` String,
        `dataset_run_description` Nullable(String),
        `dataset_run_metadata` Map(LowCardinality(String), String),
        `dataset_run_created_at` DateTime64(3),
        -- denormalized dataset item fields (mutable, but snapshots are relevant)
        `dataset_item_input` Nullable(String) CODEC(ZSTD(3)), -- json
        `dataset_item_expected_output` Nullable(String) CODEC(ZSTD(3)), -- json
        `dataset_item_metadata` Map(LowCardinality(String), String),
        -- clickhouse engine fields
        `event_ts` DateTime64(3),
        `is_deleted` UInt8,
        -- For dataset item lookups
        INDEX idx_dataset_item dataset_item_id TYPE bloom_filter(0.001) GRANULARITY 1,
    ) ENGINE = ReplicatedReplacingMergeTree(event_ts, is_deleted)
    ORDER BY (project_id, dataset_id, dataset_run_id, id); (details: code: 57, message: There was an error on [langfuse-clickhouse-shard0-2.langfuse-clickhouse-headless.monitoring.svc.cluster.local:9000]: Code: 57. DB::Exception: Table default.dataset_run_items_rmt already exists. (TABLE_ALREADY_EXISTS) (version 25.2.1.3085 (official build)))
    Applying clickhouse migrations failed. This is mostly caused by the database being unavailable.
    Exiting...
    ```
  - すべてのClickHouse Podで、`dataset_run_items_rmt`テーブルを削除  
    ```shell
    DROP TABLE dataset_run_items_rmt;
    ```
  - その後、langfuse-web podを再起動（削除）してもversionが1つ上がってdirtyエラーが出た。再度dirtyとなっている最新のversionだけを削除して、langfuse web podを削除（再起動）したら 'Running 1/1' になった！


#### 最終手段
- RDSのLangfuse用のDBを削除し、ClickHouseとZookeeper用のEFS（PV）もAccessPointを作り直して、すべてクリアした状態でデプロイしたら問題なく起動した。(既存のデータはすべて消えた)

---

# ■ ZookeeperでBindExceptionが出て、traceなどのデータが連携されない

### 事象
- ClickHouseのPodで以下のエラーが発生  
  ```shell
  2025.08.25 05:57:04.965588 [ 704 ] {} <Error> virtual bool DB::DDLWorker::initializeMainThread(): Code: 999. Coordination::Exception: All connection tries failed while connecting to ZooKeeper. nodes: 10.1.3.127:2181, 10.1.3.94:2181, 10.1.3.132:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.127:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.94:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.132:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.127:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.94:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.132:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.127:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.94:2181
  Code: 33. DB::Exception: Cannot read all data. Bytes read: 0. Bytes expected: 4.: while receiving handshake from ZooKeeper. (CANNOT_READ_ALL_DATA) (version 25.2.1.3085 (official build)), 10.1.3.132:2181

  . (KEEPER_EXCEPTION), Stack trace (when copying this message, always include the lines below):

  1. DB::Exception::Exception(DB::Exception::MessageMasked&&, int, bool) @ 0x000000000eec3c3b
  2. DB::Exception::Exception(PreformattedMessage&&, int) @ 0x0000000009f03fec
  3. DB::Exception::Exception<String&>(int, FormatStringHelperImpl<std::type_identity<String&>::type>, String&) @ 0x0000000009f19eeb
  4. Coordination::Exception::Exception<String&>(Coordination::Error, FormatStringHelperImpl<std::type_identity<String&>::type>, String&) @ 0x0000000013987baa
  5. Coordination::ZooKeeper::connect(std::vector<zkutil::ShuffleHost, std::allocator<zkutil::ShuffleHost>> const&, Poco::Timespan) @ 0x0000000014eb32a1
  6. Coordination::ZooKeeper::ZooKeeper(std::vector<zkutil::ShuffleHost, std::allocator<zkutil::ShuffleHost>> const&, zkutil::ZooKeeperArgs const&, std::shared_ptr<DB::ZooKeeperLog>) @ 0x0000000014eabebe
  7. zkutil::ZooKeeper::init(zkutil::ZooKeeperArgs, std::unique_ptr<Coordination::IKeeper, std::default_delete<Coordination::IKeeper>>) @ 0x0000000014e4e684
  8. zkutil::ZooKeeper::ZooKeeper(Poco::Util::AbstractConfiguration const&, String const&, std::shared_ptr<DB::ZooKeeperLog>) @ 0x0000000014e525ba
  9. zkutil::ZooKeeper::create(Poco::Util::AbstractConfiguration const&, String const&, std::shared_ptr<DB::ZooKeeperLog>) @ 0x0000000014e5f3ee
  10. DB::Context::getZooKeeper() const @ 0x0000000012a5e0f9
  11. DB::DDLWorker::getAndSetZooKeeper() @ 0x0000000012b085fc
  12. DB::DDLWorker::initializeMainThread() @ 0x0000000012b200a2
  13. DB::DDLWorker::runMainThread() @ 0x0000000012b05451
  14. void std::__function::__policy_invoker<void ()>::__call_impl[abi:ne190107]<std::__function::__default_alloc_func<ThreadFromGlobalPoolImpl<true, true>::ThreadFromGlobalPoolImpl<void (DB::DDLWorker::*)(), DB::DDLWorker*>(void (DB::DDLWorker::*&&)(), DB::DDLWorker*&&)::'lambda'(), void ()>>(std::__function::__policy_storage const*) @ 0x0000000012b24a03
  15. ThreadPoolImpl<std::thread>::ThreadFromThreadPool::worker() @ 0x000000000efa11ef
  16. void* std::__thread_proxy[abi:ne190107]<std::tuple<std::unique_ptr<std::__thread_struct, std::default_delete<std::__thread_struct>>, void (ThreadPoolImpl<std::thread>::ThreadFromThreadPool::*)(), ThreadPoolImpl<std::thread>::ThreadFromThreadPool*>>(void*) @ 0x000000000efa84da
  17. ? @ 0x00007fc40f9141c4
  18. ? @ 0x00007fc40f993ac0
   (version 25.2.1.3085 (official build))
  ```
- ZookeeperのPodで以下のエラーが発生  
  ```shell
  2025-08-25 06:06:42,187 [myid:] - ERROR [ListenerHandler-langfuse-zookeeper-1.langfuse-zookeeper-headless.monitoring.svc.cluster.local/10.1.3.127:3888:o.a.z.s.q.QuorumCnxManager$Listener$ListenerHandler@1099] - Exception while listening to address langfuse-zookeeper-1.langfuse-zookeeper-headless.monitoring.svc.cluster.local/10.1.3.127:3888
  java.net.BindException: Cannot assign requested address (Bind failed)
          at java.base/java.net.PlainSocketImpl.socketBind(Native Method)
          at java.base/java.net.AbstractPlainSocketImpl.bind(Unknown Source)
          at java.base/java.net.ServerSocket.bind(Unknown Source)
          at java.base/java.net.ServerSocket.bind(Unknown Source)
          at org.apache.zookeeper.server.quorum.QuorumCnxManager$Listener$ListenerHandler.createNewServerSocket(QuorumCnxManager.java:1141)
          at org.apache.zookeeper.server.quorum.QuorumCnxManager$Listener$ListenerHandler.acceptConnections(QuorumCnxManager.java:1070)
          at org.apache.zookeeper.server.quorum.QuorumCnxManager$Listener$ListenerHandler.run(QuorumCnxManager.java:1039)
          at java.base/java.util.concurrent.Executors$RunnableAdapter.call(Unknown Source)
          at java.base/java.util.concurrent.FutureTask.run(Unknown Source)
          at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(Unknown Source)
          at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(Unknown Source)
          at java.base/java.lang.Thread.run(Unknown Source)
  ```

### 原因
- Zookeeperのクラスター構成で、ZooKeeperクラスター内の各ノード（サーバー）を一意に識別する数値識別子があり、`data/myid`（bitnamiのHelmチャートを使っている場合は`/bitnami/zookeeper/data/myid`）にそのIDが保存されている。
- それが正しくは以下のようにマッピングされているはずだが、マッピングがずれてしまっている（例えばzookeeper-0のmyidが2になっている）場合に上記のようなエラーが発生する。  
  ```
  zookeeper-0 → 1
  zookeeper-1 → 2
  zookeeper-2 → 3
  ```
- Zookeeperの設定ファイルは`zoo.cfg`で、BitnamiのHelmチャートを使っている場合は`/opt/bitnami/zookeeper/conf/zoo.cfg`にある

### 対処方法
- ZookeeperのPodに入り、`data/myid`（bitnamiのHelmチャートを使っている場合は`/bitnami/zookeeper/data/myid`）を確認し、正しいIDが設定されていることを確認する。正しくは以下になっているはず。  
  ```
  zookeeper-0 → 1
  zookeeper-1 → 2
  zookeeper-2 → 3
  ```
- もし間違っていたら、`echo "正しい数字" > /bitnami/zookeeper/data/myid`で`myid`の値を修正してPodを再起動する。

---

# ■ ClickHouse podが再起動を繰り返す
## 事象
- ClickHouseのPodで以下のログが出て、再起動を繰り返す  
  ```shell
  clickhouse 02:44:45.18 INFO  ==>
  clickhouse 02:44:45.18 INFO  ==> Welcome to the Bitnami clickhouse container
  clickhouse 02:44:45.18 INFO  ==> Subscribe to project updates by watching https://github.com/bitnami/containers
  clickhouse 02:44:45.18 INFO  ==> Did you know there are enterprise versions of the Bitnami catalog? For enhanced secure software supply chain features, unlimited pulls from Docker, LTS support, or application customization, see Bitnami Premium or Tanzu Application Catalog. See https://www.arrow.com/globalecs/na/vendors/bitnami/ for more information.
  clickhouse 02:44:45.19 INFO  ==>
  clickhouse 02:44:45.19 INFO  ==> ** Starting ClickHouse setup **
  clickhouse 02:44:45.22 INFO  ==> Copying mounted configuration from /bitnami/clickhouse/etc
  clickhouse 02:44:45.24 INFO  ==> ** ClickHouse setup finished! **
  clickhouse 02:44:45.26 INFO  ==> ** Starting ClickHouse **
  Processing configuration file '/opt/bitnami/clickhouse/etc/config.xml'.
  Merging configuration file '/opt/bitnami/clickhouse/etc/conf.d/00_default_overrides.xml'.
  2025.09.03 02:44:45.346052 [ 1 ] {} <Information> Application: Will watch for the process with pid 42
  2025.09.03 02:44:45.346129 [ 1 ] {} <Warning> Application: Logging to console but received signal to close log file (ignoring).
  2025.09.03 02:44:45.346075 [ 42 ] {} <Information> Application: Forked a child process to watch
  2025.09.03 02:44:45.346437 [ 42 ] {} <Information> SentryWriter: Sending crash reports is disabled
  2025.09.03 02:44:45.409767 [ 42 ] {} <Information> Application: Starting ClickHouse 25.2.1.3085 (revision: 54495, git hash: 09989205d6fd661fb2683cbb1390fe4fcedaa548, build id: 431ADE7E477C326322B417A43CF8C4FFE14395D1), PID 42
  2025.09.03 02:44:45.409947 [ 42 ] {} <Information> Application: starting up
  2025.09.03 02:44:45.409968 [ 42 ] {} <Information> Application: OS name: Linux, version: 6.12.40, architecture: aarch64
  2025.09.03 02:44:45.410173 [ 42 ] {} <Information> Jemalloc: Value for background_thread set to true (from false)
  2025.09.03 02:44:45.416092 [ 42 ] {} <Information> Application: Available RAM: 12.00 GiB; logical cores: 4; used cores: 4.
  2025.09.03 02:44:45.416889 [ 42 ] {} <Information> CgroupsReader: Will create cgroup reader from '/sys/fs/cgroup/' (cgroups version: v2)
  2025.09.03 02:44:45.419012 [ 42 ] {} <Information> StatusFile: Status file /bitnami/clickhouse/data/status already exists - unclean restart. Contents:
  PID: 42
  Started at: 2025-09-03 02:43:40
  Revision: 54495
  2025.09.03 02:44:45.441423 [ 42 ] {} <Warning> Application: Integrity check of the executable skipped because the reference checksum could not be read.
  2025.09.03 02:44:45.441461 [ 42 ] {} <Information> Application: It looks like the process has no CAP_IPC_LOCK capability, binary mlock will be disabled. It could happen due to incorrect ClickHouse package installation. You could resolve the problem manually with 'sudo setcap cap_ipc_lock=+ep /opt/bitnami/clickhouse/bin/clickhouse'. Note that it will not work on 'nosuid' mounted filesystems.
  2025.09.03 02:44:45.441478 [ 42 ] {} <Information> MemoryWorker: Starting background memory thread with period of 50ms, using Cgroups as source
  2025.09.03 02:44:45.441591 [ 42 ] {} <Information> BackgroundSchedulePool/BgSchPool: Create BackgroundSchedulePool with 512 threads
  2025.09.03 02:44:45.476843 [ 42 ] {} <Information> Application: Lowered uncompressed cache size to 6.00 GiB because the system has limited RAM
  2025.09.03 02:44:45.492296 [ 52 ] {} <Information> MemoryTracker: Correcting the value of global memory tracker from 11.12 MiB to 474.14 MiB
  2025.09.03 02:44:45.536091 [ 42 ] {} <Information> Application: Changed setting 'max_server_memory_usage' to 10.80 GiB (12.00 GiB available memory * 0.90 max_server_memory_usage_to_ram_ratio)
  2025.09.03 02:44:45.536122 [ 42 ] {} <Information> Application: Setting merges_mutations_memory_usage_soft_limit was set to 6.00 GiB (12.00 GiB available * 0.50 merges_mutations_memory_usage_to_ram_ratio)
  2025.09.03 02:44:45.536130 [ 42 ] {} <Information> Application: Merges and mutations memory limit is set to 6.00 GiB
  2025.09.03 02:44:45.537492 [ 42 ] {} <Information> Application: Setting max_remote_read_network_bandwidth_for_server was set to 0
  2025.09.03 02:44:45.537515 [ 42 ] {} <Information> Application: Setting max_remote_write_network_bandwidth_for_server was set to 0
  2025.09.03 02:44:45.537524 [ 42 ] {} <Information> Application: ConcurrencyControl limit is set to 8
  2025.09.03 02:44:45.537537 [ 42 ] {} <Information> BackgroundSchedulePool/BgBufSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:45.538519 [ 42 ] {} <Information> BackgroundSchedulePool/BgMBSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:45.539377 [ 42 ] {} <Information> BackgroundSchedulePool/BgDistSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:45.541089 [ 42 ] {} <Warning> Application: Listen [::]:9009 failed: Poco::Exception. Code: 1000, e.code() = 98, Net Exception: Address already in use: [::]:9009 (version 25.2.1.3085 (official build)). If it is an IPv6 or IPv4 address and your host has disabled IPv6 or IPv4, then consider to specify not disabled IPv4 or IPv6 address to listen in <listen_host> element of configuration file. Example for disabled IPv6: <listen_host>0.0.0.0</listen_host> . Example for disabled IPv4: <listen_host>::</listen_host>
  2025.09.03 02:44:45.541244 [ 42 ] {} <Information> Application: Listening for replica communication (interserver): http://0.0.0.0:9009
  2025.09.03 02:44:45.580163 [ 42 ] {} <Information> CgroupsMemoryUsageObserver: Started cgroup current memory usage observer thread
  2025.09.03 02:44:45.580366 [ 618 ] {} <Information> CgroupsMemoryUsageObserver: Memory amount initially available to the process is 12.00 GiB
  2025.09.03 02:44:45.581185 [ 42 ] {} <Information> Context: Initialized background executor for merges and mutations with num_threads=16, num_tasks=32, scheduling_policy=round_robin
  2025.09.03 02:44:45.581615 [ 42 ] {} <Information> Context: Initialized background executor for move operations with num_threads=8, num_tasks=8
  2025.09.03 02:44:45.582747 [ 42 ] {} <Information> Context: Initialized background executor for fetches with num_threads=16, num_tasks=16
  2025.09.03 02:44:45.583172 [ 42 ] {} <Information> Context: Initialized background executor for common operations (e.g. clearing old parts) with num_threads=8, num_tasks=8
  2025.09.03 02:44:45.585133 [ 42 ] {} <Warning> Context: Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `echo 1 > /proc/sys/kernel/task_delayacct` or by using sysctl.
  2025.09.03 02:44:45.585974 [ 42 ] {} <Information> DNSCacheUpdater: Update period 15 seconds
  2025.09.03 02:44:45.586015 [ 42 ] {} <Information> Application: Loading metadata from /bitnami/clickhouse/data/
  2025.09.03 02:44:45.586062 [ 42 ] {} <Information> Context: Database disk name: default
  2025.09.03 02:44:45.586089 [ 42 ] {} <Information> Context: Database disk name: default, path: /bitnami/clickhouse/data/
  2025.09.03 02:44:45.643604 [ 42 ] {} <Information> DatabaseAtomic (system): Metadata processed, database system has 13 tables and 0 dictionaries in total.
  2025.09.03 02:44:45.643637 [ 42 ] {} <Information> TablesLoader: Parsed metadata of 13 tables in 1 databases in 0.050269464 sec
  clickhouse 02:44:55.09 INFO  ==>
  clickhouse 02:44:55.09 INFO  ==> Welcome to the Bitnami clickhouse container
  clickhouse 02:44:55.09 INFO  ==> Subscribe to project updates by watching https://github.com/bitnami/containers
  clickhouse 02:44:55.09 INFO  ==> Did you know there are enterprise versions of the Bitnami catalog? For enhanced secure software supply chain features, unlimited pulls from Docker, LTS support, or application customization, see Bitnami Premium or Tanzu Application Catalog. See https://www.arrow.com/globalecs/na/vendors/bitnami/ for more information.
  clickhouse 02:44:55.09 INFO  ==>
  clickhouse 02:44:55.10 INFO  ==> ** Starting ClickHouse setup **
  clickhouse 02:44:55.12 INFO  ==> Copying mounted configuration from /bitnami/clickhouse/etc
  clickhouse 02:44:55.14 INFO  ==> ** ClickHouse setup finished! **
  clickhouse 02:44:55.16 INFO  ==> ** Starting ClickHouse **
  Processing configuration file '/opt/bitnami/clickhouse/etc/config.xml'.
  Merging configuration file '/opt/bitnami/clickhouse/etc/conf.d/00_default_overrides.xml'.
  2025.09.03 02:44:56.394134 [ 1 ] {} <Information> Application: Will watch for the process with pid 42
  2025.09.03 02:44:56.394148 [ 42 ] {} <Information> Application: Forked a child process to watch
  2025.09.03 02:44:56.394218 [ 1 ] {} <Warning> Application: Logging to console but received signal to close log file (ignoring).
  2025.09.03 02:44:56.394496 [ 42 ] {} <Information> SentryWriter: Sending crash reports is disabled
  2025.09.03 02:44:56.454535 [ 42 ] {} <Information> Application: Starting ClickHouse 25.2.1.3085 (revision: 54495, git hash: 09989205d6fd661fb2683cbb1390fe4fcedaa548, build id: 1A3F7CA3FA57AA5910DEB2DA4440126BF05F69CD), PID 42
  2025.09.03 02:44:56.454680 [ 42 ] {} <Information> Application: starting up
  2025.09.03 02:44:56.454704 [ 42 ] {} <Information> Application: OS name: Linux, version: 6.12.40, architecture: x86_64
  2025.09.03 02:44:56.454929 [ 42 ] {} <Information> Jemalloc: Value for background_thread set to true (from false)
  2025.09.03 02:44:56.460035 [ 42 ] {} <Information> Application: Available RAM: 12.00 GiB; logical cores: 8; used cores: 6.
  2025.09.03 02:44:56.460068 [ 42 ] {} <Information> Application: Available CPU instruction sets: SSE, SSE2, SSE3, SSSE3, SSE41, SSE42, F16C, POPCNT, BMI1, BMI2, PCLMUL, AES, AVX, FMA, AVX2, SHA, ADX, RDRAND, RDSEED, RDTSCP, CLFLUSHOPT, CLWB, XSAVE, OSXSAVE
  2025.09.03 02:44:56.461116 [ 42 ] {} <Information> CgroupsReader: Will create cgroup reader from '/sys/fs/cgroup/' (cgroups version: v2)
  2025.09.03 02:44:56.463031 [ 42 ] {} <Information> StatusFile: Status file /bitnami/clickhouse/data/status already exists - unclean restart. Contents:
  PID: 42
  Started at: 2025-09-03 02:43:49
  Revision: 54495
  2025.09.03 02:44:56.667342 [ 42 ] {} <Information> Application: Integrity check of the executable successfully passed (checksum: 22606E44C93DF5D27026808E5D6560BA)
  2025.09.03 02:44:56.667399 [ 42 ] {} <Information> Application: It looks like the process has no CAP_IPC_LOCK capability, binary mlock will be disabled. It could happen due to incorrect ClickHouse package installation. You could resolve the problem manually with 'sudo setcap cap_ipc_lock=+ep /opt/bitnami/clickhouse/bin/clickhouse'. Note that it will not work on 'nosuid' mounted filesystems.
  2025.09.03 02:44:56.667419 [ 42 ] {} <Information> MemoryWorker: Starting background memory thread with period of 50ms, using Cgroups as source
  2025.09.03 02:44:56.667545 [ 42 ] {} <Information> BackgroundSchedulePool/BgSchPool: Create BackgroundSchedulePool with 512 threads
  2025.09.03 02:44:56.706629 [ 42 ] {} <Information> Application: Lowered uncompressed cache size to 6.00 GiB because the system has limited RAM
  2025.09.03 02:44:56.721185 [ 52 ] {} <Information> MemoryTracker: Correcting the value of global memory tracker from 11.09 MiB to 118.66 MiB
  2025.09.03 02:44:57.358151 [ 42 ] {} <Information> Application: Changed setting 'max_server_memory_usage' to 10.80 GiB (12.00 GiB available memory * 0.90 max_server_memory_usage_to_ram_ratio)
  2025.09.03 02:44:57.358189 [ 42 ] {} <Information> Application: Setting merges_mutations_memory_usage_soft_limit was set to 6.00 GiB (12.00 GiB available * 0.50 merges_mutations_memory_usage_to_ram_ratio)
  2025.09.03 02:44:57.358195 [ 42 ] {} <Information> Application: Merges and mutations memory limit is set to 6.00 GiB
  2025.09.03 02:44:57.359736 [ 42 ] {} <Information> Application: Setting max_remote_read_network_bandwidth_for_server was set to 0
  2025.09.03 02:44:57.359762 [ 42 ] {} <Information> Application: Setting max_remote_write_network_bandwidth_for_server was set to 0
  2025.09.03 02:44:57.359768 [ 42 ] {} <Information> Application: ConcurrencyControl limit is set to 12
  2025.09.03 02:44:57.359778 [ 42 ] {} <Information> BackgroundSchedulePool/BgBufSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:57.361194 [ 42 ] {} <Information> BackgroundSchedulePool/BgMBSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:57.362393 [ 42 ] {} <Information> BackgroundSchedulePool/BgDistSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:57.364167 [ 42 ] {} <Warning> Application: Listen [::]:9009 failed: Poco::Exception. Code: 1000, e.code() = 98, Net Exception: Address already in use: [::]:9009 (version 25.2.1.3085 (official build)). If it is an IPv6 or IPv4 address and your host has disabled IPv6 or IPv4, then consider to specify not disabled IPv4 or IPv6 address to listen in <listen_host> element of configuration file. Example for disabled IPv6: <listen_host>0.0.0.0</listen_host> . Example for disabled IPv4: <listen_host>::</listen_host>
  2025.09.03 02:44:57.364249 [ 42 ] {} <Information> Application: Listening for replica communication (interserver): http://0.0.0.0:9009
  2025.09.03 02:44:57.405137 [ 42 ] {} <Information> CgroupsMemoryUsageObserver: Started cgroup current memory usage observer thread
  2025.09.03 02:44:57.405392 [ 618 ] {} <Information> CgroupsMemoryUsageObserver: Memory amount initially available to the process is 12.00 GiB
  2025.09.03 02:44:57.406801 [ 42 ] {} <Information> Context: Initialized background executor for merges and mutations with num_threads=16, num_tasks=32, scheduling_policy=round_robin
  2025.09.03 02:44:57.407535 [ 42 ] {} <Information> Context: Initialized background executor for move operations with num_threads=8, num_tasks=8
  2025.09.03 02:44:57.408790 [ 42 ] {} <Information> Context: Initialized background executor for fetches with num_threads=16, num_tasks=16
  2025.09.03 02:44:57.409413 [ 42 ] {} <Information> Context: Initialized background executor for common operations (e.g. clearing old parts) with num_threads=8, num_tasks=8
  2025.09.03 02:44:57.411775 [ 42 ] {} <Warning> Context: Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `echo 1 > /proc/sys/kernel/task_delayacct` or by using sysctl.
  2025.09.03 02:44:57.412426 [ 42 ] {} <Information> DNSCacheUpdater: Update period 15 seconds
  2025.09.03 02:44:57.412467 [ 42 ] {} <Information> Application: Loading metadata from /bitnami/clickhouse/data/
  2025.09.03 02:44:57.412532 [ 42 ] {} <Information> Context: Database disk name: default
  2025.09.03 02:44:57.412543 [ 42 ] {} <Information> Context: Database disk name: default, path: /bitnami/clickhouse/data/
  2025.09.03 02:44:57.448629 [ 42 ] {} <Information> DatabaseAtomic (system): Metadata processed, database system has 13 tables and 0 dictionaries in total.
  2025.09.03 02:44:57.448664 [ 42 ] {} <Information> TablesLoader: Parsed metadata of 13 tables in 1 databases in 0.030098104 sec
  clickhouse 02:44:57.59 INFO  ==>
  clickhouse 02:44:57.60 INFO  ==> Welcome to the Bitnami clickhouse container
  clickhouse 02:44:57.60 INFO  ==> Subscribe to project updates by watching https://github.com/bitnami/containers
  clickhouse 02:44:57.60 INFO  ==> Did you know there are enterprise versions of the Bitnami catalog? For enhanced secure software supply chain features, unlimited pulls from Docker, LTS support, or application customization, see Bitnami Premium or Tanzu Application Catalog. See https://www.arrow.com/globalecs/na/vendors/bitnami/ for more information.
  clickhouse 02:44:57.60 INFO  ==>
  clickhouse 02:44:57.61 INFO  ==> ** Starting ClickHouse setup **
  clickhouse 02:44:57.64 INFO  ==> Copying mounted configuration from /bitnami/clickhouse/etc
  clickhouse 02:44:57.66 INFO  ==> ** ClickHouse setup finished! **
  clickhouse 02:44:57.68 INFO  ==> ** Starting ClickHouse **
  Processing configuration file '/opt/bitnami/clickhouse/etc/config.xml'.
  Merging configuration file '/opt/bitnami/clickhouse/etc/conf.d/00_default_overrides.xml'.
  2025.09.03 02:44:57.775905 [ 1 ] {} <Information> Application: Will watch for the process with pid 42
  2025.09.03 02:44:57.775982 [ 1 ] {} <Warning> Application: Logging to console but received signal to close log file (ignoring).
  2025.09.03 02:44:57.775935 [ 42 ] {} <Information> Application: Forked a child process to watch
  2025.09.03 02:44:57.776326 [ 42 ] {} <Information> SentryWriter: Sending crash reports is disabled
  2025.09.03 02:44:57.839549 [ 42 ] {} <Information> Application: Starting ClickHouse 25.2.1.3085 (revision: 54495, git hash: 09989205d6fd661fb2683cbb1390fe4fcedaa548, build id: 431ADE7E477C326322B417A43CF8C4FFE14395D1), PID 42
  2025.09.03 02:44:57.839729 [ 42 ] {} <Information> Application: starting up
  2025.09.03 02:44:57.839800 [ 42 ] {} <Information> Application: OS name: Linux, version: 6.12.40, architecture: aarch64
  2025.09.03 02:44:57.840012 [ 42 ] {} <Information> Jemalloc: Value for background_thread set to true (from false)
  2025.09.03 02:44:57.846135 [ 42 ] {} <Information> Application: Available RAM: 12.00 GiB; logical cores: 4; used cores: 4.
  2025.09.03 02:44:57.847027 [ 42 ] {} <Information> CgroupsReader: Will create cgroup reader from '/sys/fs/cgroup/' (cgroups version: v2)
  2025.09.03 02:44:57.848858 [ 42 ] {} <Information> StatusFile: Status file /bitnami/clickhouse/data/status already exists - unclean restart. Contents:
  PID: 42
  Started at: 2025-09-03 02:43:46
  Revision: 54495
  2025.09.03 02:44:57.869381 [ 42 ] {} <Warning> Application: Integrity check of the executable skipped because the reference checksum could not be read.
  2025.09.03 02:44:57.869413 [ 42 ] {} <Information> Application: It looks like the process has no CAP_IPC_LOCK capability, binary mlock will be disabled. It could happen due to incorrect ClickHouse package installation. You could resolve the problem manually with 'sudo setcap cap_ipc_lock=+ep /opt/bitnami/clickhouse/bin/clickhouse'. Note that it will not work on 'nosuid' mounted filesystems.
  2025.09.03 02:44:57.869428 [ 42 ] {} <Information> MemoryWorker: Starting background memory thread with period of 50ms, using Cgroups as source
  2025.09.03 02:44:57.869546 [ 42 ] {} <Information> BackgroundSchedulePool/BgSchPool: Create BackgroundSchedulePool with 512 threads
  2025.09.03 02:44:57.903776 [ 42 ] {} <Information> Application: Lowered uncompressed cache size to 6.00 GiB because the system has limited RAM
  2025.09.03 02:44:57.920151 [ 52 ] {} <Information> MemoryTracker: Correcting the value of global memory tracker from 11.13 MiB to 474.15 MiB
  2025.09.03 02:44:57.960137 [ 42 ] {} <Information> Application: Changed setting 'max_server_memory_usage' to 10.80 GiB (12.00 GiB available memory * 0.90 max_server_memory_usage_to_ram_ratio)
  2025.09.03 02:44:57.960175 [ 42 ] {} <Information> Application: Setting merges_mutations_memory_usage_soft_limit was set to 6.00 GiB (12.00 GiB available * 0.50 merges_mutations_memory_usage_to_ram_ratio)
  2025.09.03 02:44:57.960183 [ 42 ] {} <Information> Application: Merges and mutations memory limit is set to 6.00 GiB
  2025.09.03 02:44:57.961950 [ 42 ] {} <Information> Application: Setting max_remote_read_network_bandwidth_for_server was set to 0
  2025.09.03 02:44:57.961985 [ 42 ] {} <Information> Application: Setting max_remote_write_network_bandwidth_for_server was set to 0
  2025.09.03 02:44:57.961994 [ 42 ] {} <Information> Application: ConcurrencyControl limit is set to 8
  2025.09.03 02:44:57.962009 [ 42 ] {} <Information> BackgroundSchedulePool/BgBufSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:57.964505 [ 42 ] {} <Information> BackgroundSchedulePool/BgMBSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:57.966216 [ 42 ] {} <Information> BackgroundSchedulePool/BgDistSchPool: Create BackgroundSchedulePool with 16 threads
  2025.09.03 02:44:57.968467 [ 42 ] {} <Warning> Application: Listen [::]:9009 failed: Poco::Exception. Code: 1000, e.code() = 98, Net Exception: Address already in use: [::]:9009 (version 25.2.1.3085 (official build)). If it is an IPv6 or IPv4 address and your host has disabled IPv6 or IPv4, then consider to specify not disabled IPv4 or IPv6 address to listen in <listen_host> element of configuration file. Example for disabled IPv6: <listen_host>0.0.0.0</listen_host> . Example for disabled IPv4: <listen_host>::</listen_host>
  2025.09.03 02:44:57.968559 [ 42 ] {} <Information> Application: Listening for replica communication (interserver): http://0.0.0.0:9009
  2025.09.03 02:44:58.001844 [ 42 ] {} <Information> CgroupsMemoryUsageObserver: Started cgroup current memory usage observer thread
  2025.09.03 02:44:58.002408 [ 618 ] {} <Information> CgroupsMemoryUsageObserver: Memory amount initially available to the process is 12.00 GiB
  2025.09.03 02:44:58.003388 [ 42 ] {} <Information> Context: Initialized background executor for merges and mutations with num_threads=16, num_tasks=32, scheduling_policy=round_robin
  2025.09.03 02:44:58.005495 [ 42 ] {} <Information> Context: Initialized background executor for move operations with num_threads=8, num_tasks=8
  2025.09.03 02:44:58.007310 [ 42 ] {} <Information> Context: Initialized background executor for fetches with num_threads=16, num_tasks=16
  2025.09.03 02:44:58.008015 [ 42 ] {} <Information> Context: Initialized background executor for common operations (e.g. clearing old parts) with num_threads=8, num_tasks=8
  2025.09.03 02:44:58.011304 [ 42 ] {} <Warning> Context: Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `echo 1 > /proc/sys/kernel/task_delayacct` or by using sysctl.
  2025.09.03 02:44:58.012173 [ 42 ] {} <Information> DNSCacheUpdater: Update period 15 seconds
  2025.09.03 02:44:58.012207 [ 42 ] {} <Information> Application: Loading metadata from /bitnami/clickhouse/data/
  2025.09.03 02:44:58.012254 [ 42 ] {} <Information> Context: Database disk name: default
  2025.09.03 02:44:58.012287 [ 42 ] {} <Information> Context: Database disk name: default, path: /bitnami/clickhouse/data/
  2025.09.03 02:44:58.060087 [ 42 ] {} <Information> DatabaseAtomic (system): Metadata processed, database system has 13 tables and 0 dictionaries in total.
  2025.09.03 02:44:58.060117 [ 42 ] {} <Information> TablesLoader: Parsed metadata of 13 tables in 1 databases in 0.042572037 sec
  2025.09.03 02:44:59.840075 [ 686 ] {} <Information> AsyncLoader: Processed: 7.7%
  2025.09.03 02:45:08.579722 [ 683 ] {} <Information> AsyncLoader: Processed: 30.8%
  2025.09.03 02:45:11.936703 [ 683 ] {} <Information> AsyncLoader: Processed: 7.7%
  2025.09.03 02:45:12.023781 [ 688 ] {} <Information> AsyncLoader: Processed: 7.7%
  2025.09.03 02:45:17.398607 [ 688 ] {} <Information> AsyncLoader: Processed: 15.4%
  2025.09.03 02:45:19.670715 [ 683 ] {} <Information> AsyncLoader: Processed: 30.8%
  2025.09.03 02:45:22.681162 [ 685 ] {} <Information> AsyncLoader: Processed: 23.1%
  2025.09.03 02:45:24.351651 [ 1 ] {} <Warning> Application: Logging to console but received signal to close log file (ignoring).
  2025.09.03 02:45:24.351759 [ 43 ] {} <Information> Application: Received termination signal (Terminated)
  2025.09.03 02:45:29.284103 [ 685 ] {} <Information> AsyncLoader: Processed: 61.5%
  2025.09.03 02:45:29.952348 [ 683 ] {} <Information> AsyncLoader: Processed: 38.5%
  2025.09.03 02:45:34.204869 [ 1 ] {} <Warning> Application: Logging to console but received signal to close log file (ignoring).
  2025.09.03 02:45:34.204965 [ 43 ] {} <Information> Application: Received termination signal (Terminated)
  2025.09.03 02:45:34.430865 [ 1 ] {} <Warning> Application: Logging to console but received signal to close log file (ignoring).
  2025.09.03 02:45:34.430944 [ 43 ] {} <Information> Application: Received termination signal (Terminated)
  2025.09.03 02:45:40.712308 [ 685 ] {} <Information> AsyncLoader: Processed: 46.2%
  2025.09.03 02:45:43.982822 [ 688 ] {} <Information> AsyncLoader: Processed: 38.5%
  2025.09.03 02:45:46.322454 [ 683 ] {} <Information> AsyncLoader: Processed: 53.8%
  2025.09.03 02:45:49.405731 [ 685 ] {} <Information> AsyncLoader: Processed: 46.2%
  2025.09.03 02:45:51.649598 [ 684 ] {} <Information> AsyncLoader: Processed: 61.5%
  2025.09.03 02:45:59.537166 [ 685 ] {} <Information> AsyncLoader: Processed: 69.2%
  2025.09.03 02:46:00.408139 [ 686 ] {} <Information> AsyncLoader: Processed: 53.8%
  ```

## 原因
- ClickHouse起動時、初期化処理が完了する前にProbeの時間切れで再起動を繰り返す

## 対処方法
- `startupProbe`を追加し、30mの猶予を与えた  
  ```yaml
  clickhouse:
    deploy: true
    startupProbe:
      enabled: true
      failureThreshold: 180
  ```