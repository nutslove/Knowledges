- **https://opentelemetry.io/ja/docs/zero-code/python/**
- https://github.com/open-telemetry/opentelemetry-python-contrib

## 手順
- 以下を実行し、必要なパッケージをインストールする  
  ```bash
  pip install opentelemetry-distro opentelemetry-exporter-otlp
  opentelemetry-bootstrap -a install
  ```
  - `opentelemetry-bootstrap -a install`は、自動的に現在の環境のライブラリを検出し、対応するOpenTelemetryのinstrumentationパッケージをインストールする

> [!NOTE]  
> `opentelemetry-exporter-otlp`をインストールすると、`opentelemetry-exporter-otlp-proto-grpc`と`opentelemetry-exporter-otlp-proto-http`の両方が自動でインストールされる

> [!NOTE]  
> ## `opentelemetry-bootstrap -a install`の実際の処理内容
> 1. [opentelemetry-instrumentation/pyproject.toml](https://github.com/open-telemetry/opentelemetry-python-contrib/blob/main/opentelemetry-instrumentation/pyproject.toml)の`opentelemetry-bootstrap = "opentelemetry.instrumentation.bootstrap:run"`に記載されている`run`関数が実行される
> 2. その実態は[opentelemetry-instrumentation/src/opentelemetry/instrumentation/bootstrap.py](https://github.com/open-telemetry/opentelemetry-python-contrib/blob/main/opentelemetry-instrumentation/src/opentelemetry/instrumentation/bootstrap.py)の`run`関数であり、`bootstrap_gen.py`に定義されている`libraries`と`default_instrumentations`のリストにあるパッケージを引数として`_run_install`関数を実行する
> 3. `_run_install`関数は、`_find_installed_libraries`関数を呼び出す
> 4. `_find_installed_libraries`関数は、`default_instrumentations`リスト内のすべてのパッケージと、`libraries`リスト内のパッケージのうちインストールされているものを`_is_installed`関数で検出し、インストールされているパッケージだけを返す
> 5. `_run_install`関数に戻って、`_sys_pip_install`関数で、`default_instrumentations`リスト内のすべてのパッケージと、`libraries`リスト内でインストールされているパッケージのみのinstrumentationパッケージを、`_sys_pip_install`関数でインストールする

- 以下の環境変数を設定
  - `OTEL_EXPORTER_OTLP_ENDPOINT`
    - OTLPエクスポーターのエンドポイントURLを指定する。デフォルトは`http://localhost:4318`（HTTPプロトコルの場合）（gRPCの場合は`http://localhost:4317`）
  - `OTEL_TRACES_EXPORTER`
    - トレースエクスポーターを指定。`otlp`を指定するとOTLPエクスポーターが使用される
  - `OTEL_METRICS_EXPORTER`
    - メトリクスエクスポーターを指定。`prometheus`を指定するとPrometheusエクスポーターが使用される
  - `OTEL_SERVICE_NAME`
    - サービス名を指定

- その後は、Python実行時に`opentelemetry-instrument`コマンドを使用してアプリケーションを起動するだけで、自動的に計装が行われる  
  ```bash
  OTEL_SERVICE_NAME=your-service-name \
  OTEL_TRACES_EXPORTER=console,otlp \
  OTEL_METRICS_EXPORTER=console \
  OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=0.0.0.0:4317

  opentelemetry-instrument python myapp.py
  ```

---

## 自動計装の仕組み
- OpenTelemetryのPython Auto Instrumentationは、Pythonの標準機能である`sitecustomize.py`を利用して、自動的にアプリケーションの起動時に、インストールされているライブラリ/フレームワークを検出し、
 計装コードを挿入する（正確にはモンキーパッチを適用する）仕組みを採用している。
 - すべてのコードをモンキーパッチで置換するわけではなく、サポートしているライブラリ/フレームワークのコードに対してのみモンキーパッチを適用する

---

## Python Auto Instrumentationが対応(サポート)しているライブラリ/フレームワーク
- https://opentelemetry.io/ecosystem/registry/?language=python&component=instrumentation
- 以下のPythonのOpenTelemetryリポジトリの「**instrumentation**」ディレクトリ配下から確認可能（そのディレクトリ配下にあるのがAuto Instrumentationが対応しているライブラリ/フレームワーク）
  - https://github.com/open-telemetry/opentelemetry-python-contrib  
  ![](images/python_auto_instrumentation_list.jpg)

---
## メトリクス(metrics)
- Python Auto Instrumentationで、一部のinstrumentationパッケージはメトリクスも出してくれる
- Python Auto Instrumentationのリポジトリの 「**instrumentation**」ディレクトリ配下のREADME.mdに、「Metrics support」列にメトリクスも出してパッケージが確認
  - https://github.com/open-telemetry/opentelemetry-python-contrib/tree/main/instrumentation
- 以下の環境変数を設定  
  - `OTEL_METRICS_EXPORTER=prometheus`
    - default: `otlp`
  - 以下は必要に応じて設定（defaultのままで良い）
    - `OTEL_EXPORTER_PROMETHEUS_PORT`
      - default: `8080`
    - `OTEL_EXPORTER_PROMETHEUS_HOST`
      - default: `0.0.0.0`
- prometheus（もしくはOtel-Collectorなど）で`<ip_address>:8080/metrics`エンドポイントからスクレイピング

> [!TIP]  
> Python Auto Instrumentationが生成してくれるメトリクスは、exemplarsは出してくれないっぽい

---

## ログ(log)との連携
- 参考URL
  - https://signoz.io/docs/userguide/python-logs-auto-instrumentation/
  - https://opentelemetry.io/ja/docs/zero-code/python/logs-example/
  - https://opentelemetry.io/ja/docs/zero-code/python/configuration/#logging
  - https://github.com/open-telemetry/opentelemetry-python/blob/main/docs/examples/logs/example.py

- **以下の環境変数を設定して、`logging`標準ライブラリでloggerを設定し、作成したloggerでログを出力すれば、自動でTraceIDとSpanIDがログに含まれるようになる**  
  1. `OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED=true`
     - OpenTelemetryがログを自動計装
  2. `OTEL_PYTHON_LOG_CORRELATION=true`
     - TraceIDとSpanIDが自動的に追加
  3. `OTEL_PYTHON_LOG_FORMAT`
     - ログフォーマット
     - 例： `%(asctime)s [%(levelname)s] trace_id=%(otelTraceID)s span_id=%(otelSpanID)s - %(message)s`
- 上記の環境変数を設定した上で、以下のようにloggingライブラリを使ってログを出力すれば、自動的にTraceIDとSpanIDがログに含まれるようになる
  ```python
  import logging

  logging.basicConfig(level=logging.INFO)
  logger = logging.getLogger(__name__)
  
  logger.info("Creating order for user")
  ```
  - 上記のように書くだけで、実際の出力は以下のようになる  
    ```shell
      2025-10-20 12:34:56 [INFO] trace_id=1234567890abcdef1234567890abcdef span_id=1234567890abcdef - Creating order for user
    ```

> [!CAUTION]  
> `logging`標準ライブラリでloggerを設定し、作成したloggerでログを出力する必要がある。`print()`関数などでの出力ではTraceID/SpanIDは含まれない
