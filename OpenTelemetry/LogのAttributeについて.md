## Attributes
- Attributesは、ログイベントに関する追加的なコンテキスト情報を提供するものであり、Key-Valueのペアで構成される。

## Attributeの種類
- Resource AttributesとLogRecord Attributesの2種類がある。

### Resource Attributes
- ログイベントが生成されたエンティティに関する情報を提供する。
- 例: `service.name`, `service.version`, `host.name`, `cloud.provider`

### LogRecord Attributes 
- 個々のログレコードで異なるKey-Valueペアを持つことができる。
- 例: `severity`, `timestamp`, `trace_id`, `span_id`