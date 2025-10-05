# Lokiのデータ構造
- https://grafana.com/docs/loki/latest/get-started/architecture
- Grafana Loki has two main file types: **index** and **chunks**.
  - The **index** is a table of contents of where to find logs for a specific set of labels.
  - The **chunk** is a container for log entries for a specific set of labels.

## Index Format
- 2025/10現在、実質 **`TSDB`** がLokiのIndex Formatとして使われている
  - TSDB: Time Series Database (or short TSDB) is an [index format](https://github.com/prometheus/prometheus/blob/main/tsdb/docs/format/index.md) originally developed by the maintainers of Prometheus for time series (metric) data.
- ラベルからChunkへのmapping table

### TSDB Index Format
- https://github.com/prometheus/prometheus/blob/main/tsdb/docs/format/index.md

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
|                      ※以下の       |
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
- MemChunk (`pkg/chunkenc/memchunk.go`):
  - blocks: 完了した圧縮ブロック
  - head: 現在追加中のメモリ内ブロック
  - encoding: 圧縮アルゴリズム

### 圧縮
- サポートされる圧縮コーデック(`pkg/compression/codec.go`):
  - LZ4-4M (デフォルト・推奨)
  - GZIP, Snappy, Flate, Zstd

## Block Format
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

