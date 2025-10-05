# Lokiのデータ構造
- https://grafana.com/docs/loki/latest/get-started/architecture
- Grafana Loki has two main file types: **index** and **chunks**.
  - The **index** is a table of contents of where to find logs for a specific set of labels.
  - The **chunk** is a container for log entries for a specific set of labels.

## 主要データ構造
- [`pkg/ingester/stream.go`](https://github.com/grafana/loki/blob/main/pkg/ingester/stream.go)
#### Stream
- ログストリームを管理する構造体
```go
// 一部抜粋
type stream struct {
    chunks      []chunkDesc       // 複数のchunkDescを保持（最新チャンクはchunks[n-1]）
    fp          model.Fingerprint // ストリームを一意に識別するフィンガープリント(ラベルセットのハッシュ)
    labels      labels.Labels     // ストリームラベル
    highestTs   time.Time        // 最高タイムスタンプ
}
```
- 各項目の説明
  - `chunks`: このストリームに属するチャンクのリスト（`chunkDesc`構造体のスライス）
  - `fp`: ストリームを一意に識別するフィンガープリント（ラベルセットのハッシュ）
    - 例: `{app="foo"}`というラベルセット → `12345678901234567890` (fingerprint)
  - `labels`: ストリームに関連付けられたラベルのセット
  - `highestTs`: このストリームで観測された最新のタイムスタンプ
    - Unordered writes有効時: 受け入れ可能なタイムスタンプの範囲を制限するために使用
    - `highestTs` - (`max_chunk_age` / 2) より古いタイムスタンプのログエントリは拒否される
#### ChunkDesc
- チャンクのメタデータと状態を管理する構造体
```go
type chunkDesc struct {
    chunk   *chunkenc.MemChunk  // 実際のチャンク本体
    closed  bool                // チャンクがクローズ済みか
    synced  bool                // WALに同期済みか
    flushed time.Time          // Object Storageにフラッシュされた時刻
    reason  string             // フラッシュ理由 ("idle", "full", "forced"等)
    lastUpdated time.Time      // 最終更新時刻
}
```  
- 各項目の説明
  - `chunk`: 実際のMemChunkインスタンス(ログデータを含む)
  - `closed`: `true`の場合、新しいエントリは追加されない
  - `synced`: WALへの書き込みが完了したかどうか
  - `flushed`: フラッシュ済みの場合は時刻、未フラッシュはtime.Zero
  - `reason`: フラッシュ理由 (例: "idle" - アイドル時間超過, "full" - サイズ上限到達)
  - `lastUpdated`: 最後にエントリが追加された時刻 (アイドルタイムアウト判定に使用)

## Index Format
- 2025/10現在、実質 **`TSDB`** がLokiのIndex Formatとして使われている
  - TSDB: Time Series Database (or short TSDB) is an [index format](https://github.com/prometheus/prometheus/blob/main/tsdb/docs/format/index.md) originally developed by the maintainers of Prometheus for time series (metric) data.
- ラベルからChunkへのmapping table

### TSDB Index Format
- https://github.com/prometheus/prometheus/blob/main/tsdb/docs/format/index.md

### TSDB Indexのマッピングについて
これは2段階のマッピング

1. Labels → Fingerprint
   - Labels: `{app="foo", env="prod"}` のようなラベルセット
   - Fingerprint: ラベルセットに基づいて生成される一意の識別子（ラベルセットをハッシュ化した64ビットのユニーク識別子 (`model.Fingerprint`)）
   - 目的: ラベルセットを効率的に比較・検索するため
2. Fingerprint → ChunkRefs
   - Fingerprint: 上記のストリーム識別子
   - ChunkRefs: そのストリームに属する全チャンクのリスト
   - 各ChunkRef: チャンクの場所(start時刻, end時刻, checksum)を含む

これにより、クエリ時に:
1. ラベルマッチャー `{app="foo"}` → 該当するFingerprintを検索
2. Fingerprint → 該当するChunkRefsのリストを取得
3. ChunkRefsからチャンクを読み込み

## Chunk Format
- A chunk is a container for log lines of a stream (unique set of labels) of a specific time range.
- 圧縮されたログEntryのコンテナ
- The following ASCII diagram describes the chunk format in detail.  
```
----------------------------------------------------------------------------
|                        |                       |                         |
|     MagicNumber(4b)    |     version(1b)       |      encoding (1b)      |
|                        |                       |      ※圧縮形式          |
----------------------------------------------------------------------------
|                      #structuredMetadata (uvarint)                       |
----------------------------------------------------------------------------
|      len(label-1) (uvarint)      |          label-1 (bytes)              |
----------------------------------------------------------------------------
|      len(label-2) (uvarint)      |          label-2 (bytes)              |
----------------------------------------------------------------------------
|      len(label-n) (uvarint)      |          label-n (bytes)              |
----------------------------------------------------------------------------
|                      checksum(from #structuredMetadata)                  |
----------------------------------------------------------------------------
|           block-1 bytes          |           checksum (4b)               |
----------------------------------------------------------------------------
|           block-2 bytes          |           checksum (4b)               |
----------------------------------------------------------------------------
|           block-n bytes          |           checksum (4b)               |
----------------------------------------------------------------------------
|                           #blocks (uvarint)                              |
----------------------------------------------------------------------------
| #entries(uvarint) | mint, maxt (varint)  | offset, len (uvarint)         |
----------------------------------------------------------------------------
| #entries(uvarint) | mint, maxt (varint)  | offset, len (uvarint)         |
----------------------------------------------------------------------------
| #entries(uvarint) | mint, maxt (varint)  | offset, len (uvarint)         |
----------------------------------------------------------------------------
| #entries(uvarint) | mint, maxt (varint)  | offset, len (uvarint)         |
----------------------------------------------------------------------------
|                          checksum(from #blocks)                          |
----------------------------------------------------------------------------
| #structuredMetadata len (uvarint) | #structuredMetadata offset (uvarint) |
----------------------------------------------------------------------------
|     #blocks len (uvarint)         |       #blocks offset (uvarint)       |
----------------------------------------------------------------------------
```
- `mint` and `maxt` describe the minimum and maximum Unix nanosecond timestamp, respectively.
- The `structuredMetadata` section stores non-repeated strings. It is used to store label names and label values from [structured metadata](https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata/). Note that the labels strings and lengths within the `structuredMetadata` section are stored compressed.

- https://github.com/grafana/loki/tree/main/pkg/chunkenc
  - Chunk v4 format  
    ```
    // Header
    +-----------------------------------+
    | Magic Number (uint32, 4 bytes)    |
    +-----------------------------------+
    | Version (1 byte)                  |
    +-----------------------------------+
    | Encoding (1 byte)                 |
    +-----------------------------------+

    // Blocks
    +--------------------+----------------------------+
    | block 1 (n bytes)  | checksum (uint32, 4 bytes) |
    +--------------------+----------------------------+
    | block 1 (n bytes)  | checksum (uint32, 4 bytes) |
    +--------------------+----------------------------+
    | ...                                             |
    +--------------------+----------------------------+
    | block N (n bytes)  | checksum (uint32, 4 bytes) |
    +--------------------+----------------------------+

    // Metas
    +------------------------------------------------------------------------------------------------------------------------+
    | #blocks (uvarint)                                                                                                      |
    +--------------------+-----------------+-----------------+------------------+---------------+----------------------------+
    | #entries (uvarint) | minTs (uvarint) | maxTs (uvarint) | offset (uvarint) | len (uvarint) | uncompressedSize (uvarint) |
    +--------------------+-----------------+-----------------+------------------+---------------+----------------------------+
    | #entries (uvarint) | minTs (uvarint) | maxTs (uvarint) | offset (uvarint) | len (uvarint) | uncompressedSize (uvarint) |
    +--------------------+-----------------+-----------------+------------------+---------------+----------------------------+
    | ...                                                                                                                    |
    +--------------------+-----------------+-----------------+------------------+---------------+----------------------------+
    | #entries (uvarint) | minTs (uvarint) | maxTs (uvarint) | offset (uvarint) | len (uvarint) | uncompressedSize (uvarint) |
    +--------------------+-----------------+-----------------+------------------+---------------+----------------------------+
    | checksum (uint32, 4 bytes)                                                                                             | 
    +------------------------------------------------------------------------------------------------------------------------+

    // Structured Metadata
    +---------------------------------+
    | #labels (uvarint)               |
    +---------------+-----------------+
    | len (uvarint) | value (n bytes) |
    +---------------+-----------------+
    | ...                             |
    +---------------+-----------------+
    | checksum (uint32, 4 bytes)      |
    +---------------------------------+

    // Footer
    +-----------------------+--------------------------+
    | len (uint64, 8 bytes) | offset (uint64, 8 bytes) |   // offset to Structured Metadata
    +-----------------------+--------------------------+
    | len (uint64, 8 bytes) | offset (uint64, 8 bytes) |   // offset to Metas
    +-----------------------+--------------------------+
    ```

### InMemory Chunk Format
- IngesterがObject Storageにflushする前に自身のメモリ内に保持するChunk
- MemChunk (`pkg/chunkenc/memchunk.go`):
  - `blocks []block`: 完了した圧縮ブロック（圧縮済みの完了したブロック）
  - `head HeadBlock`: 現在追加中のメモリ内ブロック（現在書き込み中の未圧縮ブロック）
  - `encoding compression.Codec`: 圧縮アルゴリズム

### 圧縮
- サポートされる圧縮コーデック(`pkg/compression/codec.go`):
  - LZ4-4M (デフォルト・推奨)
  - GZIP, Snappy, Flate, Zstd

## Block Format
- 以下はLokiのChunk内に含まれるBlockの圧縮される前のフォーマット
  - 以下のBlockが圧縮されて、Chunkの一部として（`block.b []byte`に）格納される
  - `pkg/chunkenc/memchunk.go`の一部  
    ```go
    type block struct {
        b          []byte  // 複数のログエントリを含む圧縮済みのバイト列
        numEntries int     // このブロック内のエントリ数
        mint, maxt int64   // 最小・最大タイムスタンプ
        offset     int     // チャンク内のこのブロックのオフセット
        uncompressedSize int // 圧縮前のバイト数
    }
    ```
- A block is comprised of a series of entries, each of which is an individual log line. Note that the bytes of a block are stored compressed. The following is their form when uncompressed:  
```
-----------------------------------------------------------------------------------------------------------------------------------------------
|  ts (varint)  |  len (uvarint)  |  log-1 bytes  |  len(from #symbols)  |  #symbols (uvarint)  |  symbol-1 (uvarint)  | symbol-n*2 (uvarint) |
-----------------------------------------------------------------------------------------------------------------------------------------------
|  ts (varint)  |  len (uvarint)  |  log-2 bytes  |  len(from #symbols)  |  #symbols (uvarint)  |  symbol-1 (uvarint)  | symbol-n*2 (uvarint) |
-----------------------------------------------------------------------------------------------------------------------------------------------
|  ts (varint)  |  len (uvarint)  |  log-3 bytes  |  len(from #symbols)  |  #symbols (uvarint)  |  symbol-1 (uvarint)  | symbol-n*2 (uvarint) |
-----------------------------------------------------------------------------------------------------------------------------------------------
|  ts (varint)  |  len (uvarint)  |  log-n bytes  |  len(from #symbols)  |  #symbols (uvarint)  |  symbol-1 (uvarint)  | symbol-n*2 (uvarint) |
-----------------------------------------------------------------------------------------------------------------------------------------------
```
- `ts` is the Unix nanosecond timestamp of the logs, while `len` is the length in bytes of the log entry.
- Symbols store references to the actual strings containing label names and values in the `structuredMetadata` section of the chunk.
  - ラベル名や値の文字列自体は、Chunkの`structuredMetadata`セクションにあり、シンボルはその文字列にたどり着くための「案内役（参照）」として機能しているということ

