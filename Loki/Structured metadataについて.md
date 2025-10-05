# Structured metadataとは
- https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata/
- Labelにするにはカーディナリティが高すぎるが、ログに含めておきたい情報をStructured metadataとして扱うことができる
  - 例: ユーザーID、セッションID、リクエストIDなど
- **Structured metadataは、indexingされないため、高いカーディナリティでも検索パフォーマンスに影響を与えない**  
  > Selecting proper, low cardinality labels is critical to operating and querying Loki effectively. Some metadata, especially infrastructure related metadata, can be difficult to embed in log lines, and is too high cardinality to effectively store as indexed labels (and therefore reducing performance of the index).
  >
  > Structured metadata is a way to attach metadata to logs without indexing them or including them in the log line content itself. Examples of useful metadata are kubernetes pod names, process ID’s, or any other label that is often used in queries but has high cardinality and is expensive to extract at query time.

> [!WARNING]  
> Structured metadata was added to chunk format V4 which is used if the schema version is greater or equal to `13`. See [Schema Config](https://grafana.com/docs/loki/latest/configure/storage/#schema-config) for more details about schema versions.

---

# Structured metadataの有効化
- `limits_config`ブロックで`allow_structured_metadata: true`でStructured metadataを有効にする必要がある

---
# LabelsとStructured metadataの比較
## 1. Label (Stream Labels) - インデックス化される
- **TSDB Indexに保存される**
- 該当コード: `pkg/push/types.go`  
  ```go
  type Stream struct {
      Labels  string   // ← これがStream Label（インデックス化される）
      Entries []Entry
      Hash    uint64
  }
  ```

- データ例:  
  ```json
  {
    "streams": [
      {
        "labels": "{app=\"nginx\", env=\"prod\"}",  ← これがLabel
        "entries": [...]
      }
    ]
  }
  ```

- 特徴:
  - ✅ インデックス化される
  - ✅ Streamの識別に使用される
  - ✅ クエリのフィルタリングに使用: `{app="nginx"}`
  - ✅ 同じlabelセット = 同じstream
  - ✅ Cardinality制限がある（ユニークなlabelの組み合わせ数）

## 2. Structured Metadata - インデックス化されない
- **Chunk内に保存される**
- 該当コード: `pkg/push/types.go`  
  ```go
  type Entry struct {
      Timestamp          time.Time
      Line               string
      StructuredMetadata LabelsAdapter  // ← これがStructured Metadata（インデックス化されない）
      Parsed             LabelsAdapter
  }
  ```

- データ例:  
  ```json
  {
    "streams": [
      {
        "labels": "{app=\"nginx\"}",
        "entries": [
          {
            "ts": "2024-01-15T12:00:00Z",
            "line": "GET /api/users 200",
            "structuredMetadata": {       ← これがStructured Metadata
              "trace_id": "abc123",
              "user_id": "456",
              "response_time_ms": "150"
            }
          }
        ]
      }
    ]
  }
  ```

- 特徴:
  - ❌ インデックス化されない
  - ✅ エントリごとに異なる値を持てる（高cardinality OK）
  - ✅ ログエントリと一緒にチャンクに保存される
  - ✅ クエリ時にフィルタリング可能（ただし全チャンクをスキャン）
  - ✅ 抽出・集計が可能

### 具体的な例

#### ケース1: Labelのみでフィルタ（高速）

```logql
{app="nginx", env="prod"}
```

**1. TSDB Index検索（高速）**
```
Query: app="nginx" AND env="prod"
↓
Index: Labels → Fingerprint → ChunkRefs
{app="nginx", env="prod"} → fp:12345 → [chunk-1, chunk-2, chunk-3]
↓
結果: 3個のChunkRef
```

**2. Chunk取得（必要な分だけ）**
```
Object Storage から chunk-1, chunk-2, chunk-3 を取得
```

**3. Chunk展開してログ返却**

**効率**: ✅ Indexで絞り込み済み → 必要最小限のChunkのみ取得

---

#### ケース2: Label + Structured Metadataでフィルタ（低速）

```logql
{app="nginx"} | trace_id="abc123"
```

**1. TSDB Index検索（高速）**
```
Query: app="nginx"
↓
Index: Labels → Fingerprint → ChunkRefs
{app="nginx", env="prod"} → fp:12345 → [chunk-1, chunk-2, chunk-3]
{app="nginx", env="dev"}  → fp:67890 → [chunk-4, chunk-5]
↓
結果: 5個のChunkRef（全環境のnginxログ）
```

**2. 全Chunkを取得・展開（遅い）**
```
Object Storage から chunk-1, chunk-2, chunk-3, chunk-4, chunk-5 を取得
↓
各ChunkのStructured Metadata Sectionを展開
```

**3. 全エントリをスキャン（遅い）**
```
Chunk-1:
├─ Entry 1: trace_id="abc123" ✅ マッチ
├─ Entry 2: trace_id="def456" ❌
└─ Entry 3: trace_id="ghi789" ❌

Chunk-2:
├─ Entry 4: trace_id="abc123" ✅ マッチ
└─ Entry 5: trace_id="jkl012" ❌

... (全Chunkで同様にスキャン)
```

**4. マッチしたエントリのみ返却**

**効率**: ❌ Indexで絞り込めない → 全Chunk取得 → 全エントリスキャン

---

#### パフォーマンス比較

**データ構成**
```
TSDB Index:
{app="nginx", env="prod"} → 1000個のChunk
{app="nginx", env="dev"}  → 500個のChunk
{app="api", env="prod"}   → 800個のChunk
合計: 2300個のChunk
```

**クエリA: Labelのみ**
```logql
{app="nginx", env="prod"}
```

処理:
- Index検索: 1回（数ms）
- Chunk取得: 1000個
- 総データ量: 1000個分

**クエリB: Label + Structured Metadata**
```logql
{app="nginx"} | trace_id="abc123"
```

処理:
- Index検索: 1回（数ms）
- Chunk取得: 1500個（prod 1000 + dev 500）
- 全1500個のChunkを展開してスキャン
- 仮に `trace_id="abc123"` のログが100エントリしかなくても、1500個全部を調べる必要がある

**比較表**

| 項目         | Label のみ              | Label + Structured Metadata |
|--------------|-------------------------|------------------------------|
| Index検索    | ✅ 必要なChunkのみ特定  | ✅ Labelで粗い絞り込み       |
| Chunk取得数  | 1000個                  | 1500個                       |
| スキャン     | 不要                    | 全エントリスキャン           |
| 速度         | 高速                    | 低速                         |

---

#### なぜこの設計なのか？

**Structured Metadataの目的**

高Cardinality（多様性）のメタデータを扱うため

例: `trace_id`
- ユニークな値: 数百万〜数十億
- 全てをLabelにすると: 数百万のStream生成 → Indexが爆発 💥

**Labelにした場合の問題**
```
{app="nginx", trace_id="abc123"} → Stream 1
{app="nginx", trace_id="def456"} → Stream 2
{app="nginx", trace_id="ghi789"} → Stream 3
... (数百万Stream)

TSDB Index: 💥 巨大化・遅延・コスト増
```

**Structured Metadataにした場合**
```
{app="nginx"} → Stream 1 (1つだけ！)
  ├─ Entry 1: trace_id="abc123"
  ├─ Entry 2: trace_id="def456"
  ├─ Entry 3: trace_id="ghi789"
  └─ ... (数百万エントリ)

TSDB Index: ✅ 小さい・高速
Chunk内検索: 遅いが、Indexは保護される
```

---

#### ベストプラクティス

**Labelに適したもの（低Cardinality）**
- `app`, `env`, `cluster`, `namespace`
- `job`, `instance`, `region`
- `severity`, `log_level`

特徴: 値のバリエーションが少ない（10〜1000程度）

**Structured Metadataに適したもの（高Cardinality）**
- `trace_id`, `span_id`, `request_id`
- `user_id`, `session_id`
- `ip_address`, `user_agent`
- `http_status`, `response_time_ms`

特徴: 値のバリエーションが多い（数万〜数百万）

---

#### まとめ

✅ LabelはChunkに含まれない → TSDB Indexに保存
✅ Structured MetadataはChunk内 → Symbolizerとして保存
✅ Label検索は高速 → Indexで必要なChunkのみ特定
✅ Structured Metadata検索は低速 → 該当Labelの全Chunk取得・全エントリスキャン

これがLokiの設計思想の核心部分！

---

# Structured metadataのクエリー（Querying structured metadata）
> **Structured metadata is extracted automatically for each returned log line and added to the labels returned for the query. You can use labels of structured metadata to filter log line using a [label filter expression](https://grafana.com/docs/loki/latest/query/log_queries/#label-filter-expression).**
>
> For example, if you have a label `pod` attached to some of your log lines as structured metadata, you can filter log lines using:
> ```logql
> {job="example"} | pod="myservice-abc1234-56789"
> ```
> Of course, you can filter by multiple labels of structured metadata at the same time:
> ```logql
> {job="example"} | pod="myservice-abc1234-56789" | trace_id="0242ac120002"
> ```
> **Note that since structured metadata is extracted automatically to the results labels, some metric queries might return an error like `maximum of series (500) reached for a single query`.** **You can use the [Keep](https://grafana.com/docs/loki/latest/query/log_queries/#keep-labels-expression) and [Drop](https://grafana.com/docs/loki/latest/query/log_queries/#drop-labels-expression) stages to filter out labels that you don’t need.** For example:
> ```logql
> count_over_time({job="example"} | trace_id="0242ac120002" | keep job  [5m])
> ```


