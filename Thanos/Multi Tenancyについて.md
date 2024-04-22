- `Receiver`を使う方式と`Sidecar`を使う方式がある
  - 一般的にはMulti Tenancyのために`Receiver`を使うところが多い

# Receiverを使う方式
### アーキテクチャ
![](./image/multi-tenancy-receiver.jpg)  
※https://thanos.io/v0.8/proposals/201812_thanos-remote-receive/
![](./image/multi_tenancy_1.jpg)
### ingestion
- **Receiverは自動的にingestionされるデータにあるHTTPヘッダーの`THANOS-TENANT`の値を、`tenant_id`というラベルに変換して保存する**
  - **https://www.youtube.com/watch?v=SAyPQ2d8v4Q**  
  ![](./image/multi_tenancy_2.jpg)
### query
- **Querier(Query)実行時、`--query.enforce-tenancy`フラグを付けて実行すると、HTTPヘッダーの`THANOS-TENANT`の値を、`tenant_id`というラベルの値に変換/挿入してクエリーを投げてくれる**  
  ![](./image/multi_tenancy_4.jpg)
- `--query.enforce-tenancy`フラグをつけないと、Grafanaなどで`THANOS-TENANT`ヘッダーを設定してもすべてのテナント(`tenant_id`ラベル)のメトリクスが参照できてしまう


## 後で詳しく確認！
![](./image/multi_tenancy_3.jpg)
![](./image/slack-1.png)
