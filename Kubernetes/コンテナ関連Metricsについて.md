#### 前提知識
- コンテナ関連メトリクスはcAdvisorから取得できる
  - cAdvisorはkubeletに内蔵されている

### Memory
- 主要メトリクスとして`container_memory_rss`と`container_memory_working_set_bytes`がある

- `kubectl top pods`は`container_memory_working_set_bytes`の値