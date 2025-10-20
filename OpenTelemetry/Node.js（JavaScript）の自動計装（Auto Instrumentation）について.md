- 参考URL
  - https://opentelemetry.io/ja/docs/zero-code/js/
  - https://github.com/open-telemetry/opentelemetry-js-contrib

## 手順
- 以下を実行し、必要なパッケージをインストールする  
  ```bash
  npm install --save @opentelemetry/api
  npm install --save @opentelemetry/auto-instrumentations-node
  npm install --save @opentelemetry/sdk-node
  npm install --save @opentelemetry/exporter-trace-otlp-grpc
  npm install --save @opentelemetry/exporter-metrics-otlp-grpc
  npm install --save @opentelemetry/exporter-logs-otlp-grpc
  npm install --save @opentelemetry/instrumentation-runtime-node
  npm install --save winston ## loggingライブラリとしてwinstonを使う場合
  ```
- 以下の通り、環境変数を設定して、`--require @opentelemetry/auto-instrumentations-node/register`をつけて実行する  
  ```shell
  OTEL_TRACES_EXPORTER=otlp OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=your-endpoint \
  node --require @opentelemetry/auto-instrumentations-node/register app.js
  ```

## メトリクス(metrics)、ログ(logs)
- 一部のinstrumentationはメトリクスの収集にも対応しているけど、大部分はトレースのみ対応しているっぽい
- **ログに関しては、サポートしているloggingライブラリは以下の3つ（これら以外のライブラリや`console.log`ではログにtrace idやspan idは自動注入されない）**
  - **winston**
    - https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-winston
  - **pino**
    - https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-pino
  - **bunyan**
    - https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-bunyan

> [!IMPORTANT]  
> `console.log`などの標準出力へのログ出力は自動計装では対応していないので、以下のようにloggingライブラリのloggerを使ってログ出力を行う必要がある（`logger.info`、`logger.warn`、`logger.error`）  
> ```javascript
> const winston = require('winston');
> // Winston ロガーの設定
> const logger = winston.createLogger({
>   level: 'info',
>   format: winston.format.combine(
>     winston.format.timestamp(),
>     winston.format.json()
>   ),
>   transports: [
>     new winston.transports.Console()
>   ]
> });
>
> app.get('/', (req, res) => {
>   logger.info('Node.js service root endpoint called');
>
>   res.json({ service: 'nodejs-express', status: 'running' });
> });
> ```

> [!CAUTION]  
> **デフォルトで用意されている `auto-instrumentations-node/register`を使う場合、トレースしか連携されない。**  
> ※メトリクスとログは連携されない。  
> メトリクスとログを連携するためには、デフォルトの`auto-instrumentations-node/register`（`packages/auto-instrumentations-node/src/register.ts`）の代わりに、自分でカスタムのregisterスクリプトを作成して、`metricReader`と`logRecordProcessor`を初期化する必要がある。  
> - `instrumentation.js`（ファイル名は任意）  
>   ```javascript
>   const { NodeSDK, metrics, logs } = require('@opentelemetry/sdk-node');
>   const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-grpc');
>   const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-grpc');
>   const { OTLPLogExporter } = require('@opentelemetry/exporter-logs-otlp-grpc');
>   const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
>   const { RuntimeNodeInstrumentation } = require('@opentelemetry/instrumentation-runtime-node');
>
>   const sdk = new NodeSDK({
>     traceExporter: new OTLPTraceExporter(),
>     metricReader: new metrics.PeriodicExportingMetricReader({
>       exporter: new OTLPMetricExporter(),
>     }),
>     logRecordProcessor: new logs.SimpleLogRecordProcessor(
>       new OTLPLogExporter()
>     ),
>     instrumentations: [
>       getNodeAutoInstrumentations(),  // 自動計装を有効化
>       new RuntimeNodeInstrumentation(),
>     ],
>   });
>
>   sdk.start();
>
>   process.on('SIGTERM', async () => {
>     await sdk.shutdown();
>     process.exit(0);
>   });
>   ```
> - そして、以下のように実行する  
>   ```shell
>   OTEL_TRACES_EXPORTER=otlp OTEL_METRICS_EXPORTER=otlp OTEL_LOGS_EXPORTER=otlp \
>   OTEL_EXPORTER_OTLP_ENDPOINT=your-endpoint OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=your-endpoint \
>   node --require ./instrumentation.js app.js

> [!NOTE]  
> 2025/10現在、exemplarsには対応していない

### 自動計装がサポートされるライブラリ
- 以下から確認可能
  - https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/auto-instrumentations-node#supported-instrumentations