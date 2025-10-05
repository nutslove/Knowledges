# Lokiのデータ構造
- https://grafana.com/docs/loki/latest/get-started/architecture
- Grafana Loki has two main file types: **index** and **chunks**.
  - The **index** is a table of contents of where to find logs for a specific set of labels.
  - The **chunk** is a container for log entries for a specific set of labels.

## Index Format
- 2025/10現在、実質 **`TSDB`** がLokiのIndex Formatとして使われている
  - TSDB: Time Series Database (or short TSDB) is an [index format](https://github.com/prometheus/prometheus/blob/main/tsdb/docs/format/index.md) originally developed by the maintainers of Prometheus for time series (metric) data.

### TSDB Index Format
- https://github.com/prometheus/prometheus/blob/main/tsdb/docs/format/index.md

## Chunk Format
- A chunk is a container for log lines of a stream (unique set of labels) of a specific time range.  
```
----------------------------------------------------------------------------
|                        |                       |                         |
|     MagicNumber(4b)    |     version(1b)       |      encoding (1b)      |
|                        |                       |                         |
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