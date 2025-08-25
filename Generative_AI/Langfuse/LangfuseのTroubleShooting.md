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

### 原因
- https://github.com/langfuse/langfuse/issues/6679

### 対処方法
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