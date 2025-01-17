# Tempo
## Traces → Logs
- https://grafana.com/docs/grafana/next/datasources/tempo/configure-tempo-data-source/#trace-to-logs
- 設定例  
  ![](./image/trace_to_log_1.jpg)  
- TraceのService Name（`otel.service.name`等で設定する名前）がLogQLのラベルとして設定される  
  ![](./image/trace_to_log_2.jpg)
- `Filter by trace ID`や`Filter by span ID`にチェックを入れると対象のTraceID、SpanIDに絞るようにLogQLが生成される  
  ![](./image/trace_to_log_3.jpg)

## Traces → Metrics
- https://grafana.com/docs/grafana/next/datasources/tempo/configure-tempo-data-source/#trace-to-metrics
- `Link Label`に紐づけるメトリクスの概要が分かる名前を入れて、`Query`にPromQLを記述
- 設定例  
  ![](./image/trace_to_metric_1.jpg) 

## Traces → Profiles
- https://grafana.com/docs/grafana/next/datasources/tempo/configure-tempo-data-source/#trace-to-profiles

# Loki
## Logs → Traces
- https://grafana.com/docs/grafana/next/datasources/loki/configure-loki-data-source/#derived-fields
- Typeが`Label`と`Regex in log line`の２つある
### `Label` Type
- ログのラベルにtrace id用のラベルがある場合
  - otlpでログを送信された場合など
- 設定例  
  ![](./image/log_to_trace_1.jpg)
### `Regex in log line` Type
- ログの中身にtrace id項目が含まれている場合
- 設定例  
  ![](./image/log_to_trace_2.jpg)