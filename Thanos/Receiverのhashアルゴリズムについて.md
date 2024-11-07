- hashringはどのデータ(メトリクス)をどのReceiverに転送するかを決める仕組み
- hashringで使えるアルゴリズムには`Hashmod`と`Ketama`がある
- 歴史的な理由で`Hashmod`がデフォルトのアルゴリズムになっているが、現在は`Ketama`が推奨されている。`Ketama`は一貫性のあるhashアルゴリズムでデータの再分散を最小限に抑える。
  - https://thanos.io/tip/components/receive.md/#hashmod-discouraged  
    > its usage for new Receive installations is discouraged since adding new Receiver nodes leads to series churn and memory usage spikes.
- `Hashmod`はデータのハッシュ値をノード数で割ったあまりを使用してデータの割り振り先(ノード)を決める。そのため、ノード(Receiver)数に変動があった場合、すべてのデータの連携(割り振り)先が変わり、データ(series)変動とメモリのスパイクが発生する。
  - メモリのスパイク発生理由は、新しいseriesのデータをメモリにロードする必要があるため
- **Receiverの増減で既存のReceiver上のデータの移動は発生しない。あくまでReceiverが追加/削除された後に連携されるメトリクスの割り振り先に影響があるという話。**
- **ThanosではメトリクスのLabel Setでhash値を求める。**

## 流れ
- 以下のようなReceiverの`--receive.hashrings-file`で指定する`hashring.json`ファイルがあるとする  
  ```json
  [
      {
          "hashring": "soft-tenants",
          "endpoints": [
              "thanos-ingesting-receiver-0.thanos-ingesting-receiver.metrics.svc.cluster.local:10901",
              "thanos-ingesting-receiver-1.thanos-ingesting-receiver.metrics.svc.cluster.local:10901",
              "thanos-ingesting-receiver-2.thanos-ingesting-receiver.metrics.svc.cluster.local:10901"
          ]
      }
  ]    
  ```
- `endpoints`内の各（Ingesting）Receiver（上記の例だと`"thanos-ingesting-receiver-0.thanos-ingesting-receiver.metrics.svc.cluster.local:10901"`～`"thanos-ingesting-receiver-2.thanos-ingesting-receiver.metrics.svc.cluster.local:10901"`）はそれぞれハッシュ化され、リング状に配置される。例えばreceiver-0のハッシュ値は5で、receiver-1のハッシュ値は10で、receiver-2のハッシュ値は15だとする。
- Receiverに連携されるメトリクスがラベルでハッシュ化される。例えばメトリクス1～3があるとして、メトリクス1のハッシュ値が4で、メトリクス2のハッシュ値が7で、メトリクス3のハッシュ値が13だとする
- 各メトリクスは自分のハッシュ値と近いReceiverに（Routing Receiverによって）フォーワーディングされる。上記の例だとメトリクス1はreceiver-0に、メトリクス2はreceiver-1に、メトリクス3はreceiver-2にフォーワーディングされる。