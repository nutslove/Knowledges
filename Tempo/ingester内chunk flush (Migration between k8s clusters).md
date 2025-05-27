- **https://grafana.com/docs/tempo/latest/api_docs/**

## `/flush`エンドポイント
- Lokiの場合、Ingesterの`/flush`エンドポイントで腹持ちしているchunkをバックエンドにflushしてくれるけど、Tempoの場合、バックエンドではなくWALにだけ移動してくれるっぽい  
  > Triggers a flush of all in-memory traces to the WAL. Useful at the time of rollout restarts and unexpected crashes.

## `/shutdown`エンドポイント
- WALではなく、バックエンドにflushするためには`/shutdown`エンドポイントを利用する必要がある  
  > Flushes all in-memory traces and the WAL to the long term backend. Gracefully exits from the ring. Shuts down the ingester service.
- 以下は実際の例  
  ```shell
  curl -X POST http://<Tempo IngesterのIP>:3100/shutdown
  # "shutdown job acknowledged"が返ってくる
  ```
  - Tempo Ingester Podのログ  
    ```shell
    level=info ts=2025-05-27T11:36:21.552354049Z caller=flush.go:94 msg="shutdown handler called"
    level=info ts=2025-05-27T11:36:21.552432383Z caller=lifecycler.go:594 msg="lifecycler loop() exited gracefully" ring=ingester
    level=info ts=2025-05-27T11:36:21.552457515Z caller=lifecycler.go:977 msg="changing instance state from" old_state=ACTIVE new_state=LEAVING ring=ingester
    level=info ts=2025-05-27T11:36:21.552544506Z caller=lifecycler.go:1056 msg="transfers are disabled"
    level=info ts=2025-05-27T11:36:21.552653979Z caller=lifecycler.go:1073 msg="lifecycler entering final sleep before shutdown" final_sleep=0s
    level=info ts=2025-05-27T11:36:21.552886799Z caller=lifecycler.go:646 msg="instance removed from the KV store" ring=ingester
    level=info ts=2025-05-27T11:36:21.553049614Z caller=module_service.go:120 msg="module stopped" module=ingester
    level=info ts=2025-05-27T11:36:21.553071636Z caller=flush.go:108 msg="shutdown handler complete"
    ```
  - Ingester Podは以下のように`0/1`の状態になる（Pod自体を削除してPodを再度起動させると、しばらくしたら1/1になる）  
    ```shell
    multi-tenant-tempo-ingester-0   0/1   Running
    ```
  - `0/1`状態になっているIngester Podを終了させると以下のログが出る  
    ```shell
    level=info ts=2025-05-27T11:39:32.00306131Z caller=app.go:236 msg="=== received SIGINT/SIGTERM ===\n*** exiting"
    level=info ts=2025-05-27T11:39:32.004924013Z caller=module_service.go:120 msg="module stopped" module=usage-report
    level=info ts=2025-05-27T11:39:32.004979561Z caller=memberlist_client.go:742 msg="leaving memberlist cluster"
    level=info ts=2025-05-27T11:39:32.00507209Z caller=module_service.go:120 msg="module stopped" module=overrides
    level=info ts=2025-05-27T11:39:32.005152361Z caller=module_service.go:120 msg="module stopped" module=store
    level=info ts=2025-05-27T11:39:32.006135491Z caller=module_service.go:120 msg="module stopped" module=cache-provider
    level=info ts=2025-05-27T11:39:33.800567547Z caller=module_service.go:120 msg="module stopped" module=memberlist-kv
    level=info ts=2025-05-27T11:39:33.800725528Z caller=server_service.go:164 msg="server stopped"
    level=info ts=2025-05-27T11:39:33.800786714Z caller=module_service.go:120 msg="module stopped" module=server
    level=info ts=2025-05-27T11:39:33.80142792Z caller=module_service.go:120 msg="module stopped" module=internal-server
    ``` 