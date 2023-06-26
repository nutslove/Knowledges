## DjangoでAuto Instrumentationは`opentelemetry-instrument`コマンドを使って実行する
- 書式
  - `opentelemetry-instrument [オプション] python manage.py runserver 0.0.0.0:80`
- 参考URL
  - https://opentelemetry-python.readthedocs.io/en/latest/examples/django/README.html
  - https://opentelemetry.io/docs/instrumentation/python/automatic/
- **`opentelemetry-instrument`付きで実行すると`settings.py`に`ALLOWED_HOSTS`を設定しているにも関わらず、以下のエラーが出る**
  - `CommandError: You must set settings.ALLOWED_HOSTS if DEBUG is False`
  - **以下の環境変数の設定が必要**
    - `export DJANGO_SETTINGS_MODULE=<アプリ名>.settings`
    - https://opentelemetry-python.readthedocs.io/en/latest/examples/django/README.html


## ConfigはCLIに直接渡す方法と環境変数に設定する2通りの方法がある
- https://opentelemetry.io/docs/instrumentation/python/automatic/agent-config/
### CLIに直接渡す方法
- https://opentelemetry.io/docs/instrumentation/python/automatic/agent-config/
### 環境変数
- https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/