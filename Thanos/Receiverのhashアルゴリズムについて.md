- hashringはどのデータ(メトリクス)をどのReceiverに転送するかを決める機構
- hashringで使えるアルゴリズムには`Hashmod`と`Ketama`がある
- 歴史的な理由で`Hashmod`がデフォルトのアルゴリズムになっているが、現在は`Ketama`が推奨されている。`Ketama`は一貫性のあるhashアルゴリズムでデータの再分散を最小限に抑える。
  - https://thanos.io/tip/components/receive.md/#hashmod-discouraged  
    > its usage for new Receive installations is discouraged since adding new Receiver nodes leads to series churn and memory usage spikes.
- `Hashmod`はデータのハッシュ値をノード数で割ったあまりを使用してデータの割り振り先(ノード)を決める。そのため、ノード(Receiver)数に変動があった場合、すべてのデータの連携(割り振り)先が変わり、データ(series)変動とメモリのスパイクが発生する。
  - メモリのスパイク発生理由は、新しいseriesのデータをメモリにロードする必要があるため
- **Receiverの増減で既存のReceiver上のデータの移動は発生しない。あくまでReceiverが追加/削除された後に連携されるメトリクスの割り振り先に影響があるという話。**