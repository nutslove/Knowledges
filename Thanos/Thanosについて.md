## アーキテクチャ（Sidecar方式）
![](./image/Thanos_sidecar.jpg)

## Sidecar
- PrometheusコンテナのSidecarコンテナとして起動され、定期的に (defaultでは2時間ごと) Prometheusのメトリクスをオブジェクトストレージに送信する
- Store APIも実装されており、オブジェクトストレージに送信されてない直近のメトリクスについてPrometheusにクエリーを投げる

## Store (Store Gateway)
- https://thanos.io/tip/components/store.md/

## Querier (Query)
- https://thanos.io/tip/components/query.md/

## Compactor
- https://thanos.io/tip/components/compact.md/
- Down Sampling、Retention、Compactionを担当するコンポーネント
- *Compaction*
  - responsible for **compacting multiple blocks into one to reduce the number of blocks and compact index indices.**
  - https://thanos.io/tip/components/compact.md/#compaction

## Store API
- https://thanos.io/tip/thanos/integrations.md/#storeapi
- https://thanos.io/tip/thanos/quick-tutorial.md/#store-api