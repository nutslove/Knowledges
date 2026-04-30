## Envoy Admin API

### 概要

Envoy プロキシ自身が公開している管理用 HTTP API。Envoy がどんな設定で動いているか、何が起きているか、現在の状態をリアルタイムに観察・操作できる。

主な用途:
- 動作中の設定 (cluster / listener / route / endpoint) の確認
- 統計情報 (stats) の取得
- ログレベルの動的変更
- ヘルスチェックの強制 / Drain などの運用操作

### Envoy Gateway / Envoy AI Gateway での Admin API のポート

デフォルトで Envoy Pod 内の `localhost:19000` で起動している。**外部からはアクセスできない (それで安全)**。

確認:

```bash
ENVOY_POD=$(kubectl get pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -o jsonpath='{.items[0].metadata.name}')

# Pod内のリスナーポート確認
kubectl get pod -n envoy-gateway-system $ENVOY_POD -o yaml | grep -A 3 containerPort
```

### アクセス方法: kubectl port-forward

```bash
kubectl port-forward -n envoy-gateway-system $ENVOY_POD 19000:19000 &

# あとはローカルからアクセスし放題
curl -s http://localhost:19000/help
```

### 主要エンドポイント

| エンドポイント | 用途 |
|---------------|------|
| `/ready` | Envoy が起動完了したか (LIVE が返る) |
| `/help` | 使えるエンドポイント一覧 |
| `/stats` | 統計情報 (テキスト形式) |
| `/stats/prometheus` | Prometheus 形式の統計 |
| `/clusters` | 動作中の cluster (upstream) 一覧と各 endpoint 状態 |
| `/listeners` | listener (downstream) 一覧 |
| `/config_dump` | 動作中の Envoy 設定全体 (JSON) |
| `/server_info` | Envoy のバージョンや起動オプション |
| `/logging` | ログレベルの確認・変更 |
| `/runtime` | runtime 設定 |
| `/healthcheck/fail` | ヘルスチェックを意図的に失敗させる (drain用) |
| `/healthcheck/ok` | ヘルスチェックを正常に戻す |

### よく使うコマンド集

#### 1. cluster 一覧と health 状態確認

```bash
# 全 cluster の概要
curl -s http://localhost:19000/clusters | grep observability_name

# 特定 cluster の詳細
curl -s http://localhost:19000/clusters | grep -A 30 "rca-llm"

# 各 cluster の health flag
curl -s http://localhost:19000/clusters | grep "health_flags"
```

#### 2. stats (統計) 取得

stats は数千行ある。grep で絞り込むのが基本。

```bash
# 全 stats
curl -s http://localhost:19000/stats

# 特定 cluster の stats
curl -s http://localhost:19000/stats | grep "rule/7"

# 0 以外の値だけ (実際に発生したイベント)
curl -s http://localhost:19000/stats | grep "rule/7" | grep -v ": 0$"

# エラー系の stats
curl -s http://localhost:19000/stats | grep -E "(fail|error|reset|protocol)" | grep -v ": 0$"

# Prometheus 形式 (メトリクス収集に便利)
curl -s http://localhost:19000/stats/prometheus
```

#### 3. 動作中の Envoy 設定確認

```bash
# 全設定 (大量、要 jq 整形)
curl -s http://localhost:19000/config_dump | jq '.'

# cluster 設定だけ抜粋
curl -s http://localhost:19000/config_dump | \
  jq '.configs[] | select(."@type" | contains("Cluster"))'

# 特定 cluster の設定
curl -s http://localhost:19000/config_dump | \
  jq '.configs[] | select(."@type" | contains("Cluster")) | .dynamic_active_clusters[] | select(.cluster.name | contains("rule/7"))'
```

#### 4. ログレベルの動的変更

Envoy 再起動なしで一時的にログレベル変更可能。

```bash
# 現在のログレベル確認
curl -s http://localhost:19000/logging

# 全部 debug にする
curl -X POST "http://localhost:19000/logging?level=debug"

# 特定コンポーネントのみ debug (例: HTTP/2)
curl -X POST "http://localhost:19000/logging?http2=debug"

# info に戻す
curl -X POST "http://localhost:19000/logging?level=info"
```

### stat の読み方の基本

#### upstream 系 (= サーバーへの接続)

| stat 名 | 意味 |
|---------|------|
| `upstream_cx_total` | upstream への接続数(累計) |
| `upstream_cx_active` | upstream への接続数(現在) |
| `upstream_cx_http1_total` | HTTP/1.1 接続数 |
| `upstream_cx_http2_total` | HTTP/2 接続数 |
| `upstream_cx_destroy_local` | ローカル(自分)から切断した数 |
| `upstream_cx_destroy_remote` | リモート(相手)から切断された数 |
| `upstream_cx_connect_fail` | 接続失敗数 |
| `upstream_cx_protocol_error` | プロトコルエラーで切断された数 |
| `upstream_rq_total` | リクエスト総数 |
| `upstream_rq_2xx` / `5xx` | レスポンスコード別カウント |
| `upstream_rq_tx_reset` | 送信側リセット |
| `upstream_rq_rx_reset` | 受信側リセット |

#### TLS 系

| stat 名 | 意味 |
|---------|------|
| `ssl.handshake` | TLS ハンドシェイク成功数 |
| `ssl.connection_error` | TLS 接続エラー |
| `ssl.fail_verify_*` | 証明書検証失敗 |
| `ssl.versions.TLSv1.2` | TLS 1.2 で接続した数 |
| `ssl.versions.TLSv1.3` | TLS 1.3 で接続した数 |

#### HTTP/2 系

| stat 名 | 意味 |
|---------|------|
| `http2.rx_reset` | HTTP/2 stream RST 受信数 |
| `http2.tx_reset` | HTTP/2 stream RST 送信数 |
| `http2.goaway_sent` | GOAWAY フレーム送信 |
| `http2.streams_active` | 現在アクティブな stream 数 |

### トラブルシュート時の効率的な見方

エラー絞り込みワンライナー:

```bash
# 0 以外の値がある stats をすべて表示
curl -s http://localhost:19000/stats | grep -v ": 0$"

# エラー / 失敗系のみ
curl -s http://localhost:19000/stats | grep -E "(fail|error|reset|reject)" | grep -v ": 0$"

# 特定 hostname に関するもの
curl -s http://localhost:19000/clusters | grep "rca-llm"
curl -s http://localhost:19000/stats | grep "20.191.161.227" | grep -v ": 0$"
```

### 注意点

- **Pod 単位**: `replicas` が複数の場合、stats は Pod ごと。すべての Pod で見る必要がある
- **再起動でリセット**: stats は Pod 再起動でゼロから始まる。Prometheus 等で外部に蓄積するのが推奨
- **Admin API は外部公開しない**: 操作系エンドポイント (`/healthcheck/fail` など) があるため、kubectl port-forward 経由でのみアクセスする

---

## EnvoyPatchPolicy

### 概要

**Envoy Gateway がコントロールプレーンとして生成する Envoy 設定 (xDS) を、JSON Patch / Strategic Merge Patch でカスタマイズできる仕組み。**

Gateway API や Envoy Gateway 公式の CRD でカバーされていない、低レベルの Envoy 機能を使いたい時の最終手段。

### いつ使うか

通常は不要。以下のような **Envoy Gateway の標準 CRD では実現できない機能** が必要な時に使う:

- 特定 cluster の HTTP version を強制 (今回のケース)
- カスタム retry ポリシー
- WASM フィルター挿入
- ヘッダー操作の高度な制御
- Envoy 独自の Bootstrap 設定変更

> [!WARNING]
> EnvoyPatchPolicy は Envoy Gateway がいつでも仕様変更する可能性のある低レベル機能。標準 CRD で代替手段がある場合はそちらを優先。

### 有効化

デフォルトでは無効。EnvoyGateway 設定で明示的に有効化する必要がある。

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
extensionApis:
  enableEnvoyPatchPolicy: true   # ← これが必要
```

確認:

```bash
kubectl get configmap -n envoy-gateway-system envoy-gateway-config -o yaml | grep -i patch
# enableEnvoyPatchPolicy: true
```

### 基本構造

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: <ポリシー名>
  namespace: envoy-gateway-system
spec:
  type: JSONPatch                 # ← パッチ方式 (JSONPatch / StrategicMerge)
  targetRef:                       # ← 対象 Gateway
    group: gateway.networking.k8s.io
    kind: Gateway
    name: <Gateway名>
  jsonPatches:                     # ← パッチ内容のリスト
    - type: <Envoyリソース型>
      name: <リソース名>
      operation:
        op: add | replace | remove
        path: <JSONパス>
        value: <値>
```

### 主な要素

#### `spec.type`

| 値 | 説明 |
|----|------|
| `JSONPatch` | RFC 6902 JSON Patch 形式。op (add/replace/remove) で操作 |
| `StrategicMerge` | Kubernetes と同じ Strategic Merge Patch |

通常は `JSONPatch` を使う。

#### `spec.targetRef`

どの Gateway の設定をパッチするか。Gateway 単位でしか指定できない (個別の HTTPRoute や Backend は指定不可)。

#### `spec.jsonPatches[].type`

どの Envoy リソースをパッチするか。Envoy のプロトコルバッファ型を指定する。

| type 値 | 対象リソース |
|---------|-------------|
| `type.googleapis.com/envoy.config.cluster.v3.Cluster` | Cluster (upstream) |
| `type.googleapis.com/envoy.config.listener.v3.Listener` | Listener (downstream) |
| `type.googleapis.com/envoy.config.route.v3.RouteConfiguration` | Route 設定 |
| `type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment` | Endpoint |

#### `spec.jsonPatches[].name`

パッチ対象の Envoy リソース名。Envoy Gateway が動的に生成するため、命名規則を理解しておく必要がある。

##### Cluster 名の規則

```
httproute/<HTTPRouteのnamespace>/<HTTPRouteのname>/rule/<ルール番号>
```

例:
- `httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/7` → `envoy-ai-gateway` HTTPRoute の 8 番目のルール (0始まり)

`AIGatewayRoute` は内部で `HTTPRoute` に変換されるため、AIGatewayRoute のルール順番がそのまま `rule/N` になる。

確認方法:

```bash
# Pod内のEnvoy Admin APIで現在のcluster名一覧
curl -s http://localhost:19000/clusters | grep observability_name | sort -u
```

#### `spec.jsonPatches[].operation`

JSON Patch 標準の操作。

| op | 説明 |
|----|------|
| `add` | 新しいフィールドを追加。既存の場合は置き換え |
| `replace` | 既存フィールドを置き換え (フィールドが存在しないと失敗) |
| `remove` | フィールドを削除 |
| `copy` | フィールドをコピー |
| `move` | フィールドを移動 |
| `test` | フィールドの値を検証 |

`path` は JSON Pointer 形式。例: `/typed_extension_protocol_options/some_key`

### 今回適用したパッチの解説

```yaml
- type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
  name: httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/7
  operation:
    op: add
    path: "/typed_extension_protocol_options"
    value:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
        explicit_http_config:
          http_protocol_options: {}
```

意味を分解すると:

1. `type: "...Cluster"` → Cluster リソースをパッチ対象に
2. `name: httproute/.../rule/7` → 8 番目のルールに対応する cluster
3. `op: add` + `path: /typed_extension_protocol_options` → Cluster の `typed_extension_protocol_options` フィールドを追加
4. `value` → そのフィールドに以下を設定:
   - `HttpProtocolOptions` 型を使う
   - `explicit_http_config` で**プロトコルを明示指定**する
   - `http_protocol_options: {}` (= HTTP/1.1) を選ぶ

これにより、Envoy はこの cluster に対して **TLS ALPN で h2 が提案されても無視して HTTP/1.1 で接続する** ようになる。

### `typed_extension_protocol_options` の選択肢

`explicit_http_config` で指定できるプロトコル:

| 設定 | プロトコル |
|------|-----------|
| `http_protocol_options: {}` | HTTP/1.1 |
| `http2_protocol_options: {}` | HTTP/2 |
| `http3_protocol_options: {}` | HTTP/3 |

ALPN による自動選択をやめて固定したい場合に使う。

### 適用後の確認

```bash
# パッチが正しく適用されたか
kubectl describe envoypatchpolicy -n envoy-gateway-system <名前>

# Conditionsを見る
# Programmed: True が出ていれば Envoy に反映されている
# False の場合、Message に失敗理由 (例: "cluster not found")
```

実際の Envoy 設定で反映されているかは Admin API で確認:

```bash
curl -s http://localhost:19000/config_dump | \
  jq '.configs[] | select(."@type" | contains("Cluster")) | .dynamic_active_clusters[] | select(.cluster.name | contains("rule/7")) | .cluster.typed_extension_protocol_options'
```

### よくある落とし穴

#### 1. `name` のスペルミス

`rule/N` の番号が間違っているとパッチは適用されないが、エラーも出にくい。Conditions で必ず確認する。

#### 2. パッチ対象の動的性

Envoy Gateway が生成する cluster 名は、HTTPRoute の構成に応じて動的に変わる。`rule/N` の番号は不安定。

#### 3. 適用順序

EnvoyPatchPolicy は **Envoy Gateway が標準で生成した設定の後**に適用される。生成されないリソースは patch できない。

#### 4. バージョン互換性

Envoy Gateway / Envoy のバージョンが上がると、パッチ対象のフィールド構造が変わる可能性がある。Envoy Gateway をアップグレードする際はパッチの互換性確認が必要。

### デバッグ方法

#### 1. パッチ適用ステータス確認

```bash
kubectl get envoypatchpolicy -A
kubectl describe envoypatchpolicy -n envoy-gateway-system <名前>
```

`Status.Conditions`:
- `Accepted: True` → CRD validation 通過
- `Programmed: True` → Envoy に実際に反映された

`Programmed: False` の場合、`Message` に失敗理由:
- `cluster not found` → `name` の指定ミス
- `invalid path` → `path` の JSON Pointer 構文エラー
- `validation error` → `value` の中身が Envoy 設定として不正

#### 2. 実 Envoy 設定の確認

```bash
# 該当 cluster の現在の設定
curl -s http://localhost:19000/config_dump | \
  jq '.configs[].dynamic_active_clusters[]?.cluster | select(.name | contains("rule/7"))'
```

#### 3. パッチを試行錯誤する時

`kubectl apply -f` で何度も試すより、`kubectl edit envoypatchpolicy <名前> -n envoy-gateway-system` で直接編集して、即座にステータスを確認する方が早い。

### 公式ドキュメント

- Envoy Gateway EnvoyPatchPolicy: https://gateway.envoyproxy.io/docs/tasks/extensibility/envoy-patch-policy/
- Envoy HttpProtocolOptions: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/upstreams/http/v3/http_protocol_options.proto

---

## まとめ: 今回の障害対応で学んだ実践フロー

### 段階的な切り分け手法

1. **アプリケーションログ (ExtProc) で問題箇所の特定**
   - `kubectl logs` で正常動作ログとエラー位置の確認

2. **依存先サービスを直接叩いて切り分け**
   - curl で Azure OpenAI に直接接続 → サービス側の問題か、Envoy 経由の問題かを切り分け

3. **Envoy Admin API で内部状態を観察**
   - `/clusters` と `/stats` でリクエストがどこまで到達しているかを確認

4. **複数 Pod の場合は両方確認**
   - replicas が 2 以上なら、必ず両方の Pod の stats を見る

5. **HTTP プロトコルバージョン別の挙動確認**
   - `curl --http1.1` と `curl --http2` で動作差を確認

### Admin API + EnvoyPatchPolicy のセット運用

| 役割 | ツール |
|------|--------|
| 観察 (read) | Envoy Admin API (`/stats`, `/clusters`, `/config_dump`) |
| 介入 (write) | EnvoyPatchPolicy |

問題が起きたら:
- Admin API で何が起きているかを正確に把握 →
- 標準 CRD で対処できないか検討 →
- どうしてもダメなら EnvoyPatchPolicy で介入 →
- Admin API で介入が効いているか確認

この往復が Envoy Gateway 系のトラブルシュートの基本パターン。