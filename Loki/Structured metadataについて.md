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


