# Pydantic Settingsとは
- アプリケーションの設定管理を型安全に行うためのPythonライブラリ
- 環境変数や設定ファイル（e.g. `.env`）から設定を読み込んで自動的に検証できる
- `BaseSettings`、`SettingsConfigDict`などのビルトインクラスを使って設定モデルを定義できる

## 使い方
1. Pydantic Settingsをインストールする  
   ```bash
   pip install pydantic-settings
   ```
2. 設定モデルを定義する  
    ```python
    from pydantic_settings import BaseSettings, SettingsConfigDict

    class Settings(BaseSettings):
        # 基本的な設定
        app_name: str = "MyApp"
        debug: bool = False
        port: int = 8000
        
        # リストやその他の複雑な型も扱える
        allowed_hosts: list[str] = []
        
        model_config = SettingsConfigDict(
            env_file=".env",  # .envファイルから読み込む
            env_file_encoding="utf-8",
            case_sensitive=False  # 環境変数名の大文字小文字を区別しない
        )

    # インスタンス化すると自動的に環境変数を読み込む
    settings = Settings()
    print(settings.database_url)
    ```

- モデルで定義している環境変数が設定されていない場合はエラーが出る.  
  例えば、以下のように`database_url`と`api_key`フィールドを持つ設定モデルを定義した場合、対応する環境変数が設定されていないとエラーになる.  
  ```python
  from pydantic_settings import BaseSettings
  class Settings(BaseSettings):
    database_url: str
    api_key: str

  settings = Settings()
  ```
  - エラーメッセージ例:  
    ```
    Traceback (most recent call last):
      File "<python-input-2>", line 1, in <module>
        settings = Settings()
      File "/Users/joonki.lee/.pyenv/versions/3.13.9/lib/python3.13/site-packages/pydantic_settings/main.py", line 194, in __init__
        super().__init__(
        ~~~~~~~~~~~~~~~~^
            **__pydantic_self__._settings_build_values(
            ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        ...<27 lines>...
            )
            ^
        )
        ^
      File "/Users/joonki.lee/.pyenv/versions/3.13.9/lib/python3.13/site-packages/pydantic/main.py", line 250, in __init__
        validated_self = self.__pydantic_validator__.validate_python(data, self_instance=self)
    pydantic_core._pydantic_core.ValidationError: 2 validation errors for Settings
    database_url
      Field required [type=missing, input_value={}, input_type=dict]
        For further information visit https://errors.pydantic.dev/2.12/v/missing
    api_key
      Field required [type=missing, input_value={}, input_type=dict]
        For further information visit https://errors.pydantic.dev/2.12/v/missing
    ```

> [!NOTE]  
> - 環境変数の名前は、フィールド名を**大文字**に変換し、**アンダースコア**で区切った形式になる
>   - 例えば、`database_url`フィールドは`DATABASE_URL`環境変数から読み込まれる
> - 命名規則をカスタマイズできる  
>   ```python
>   class DatabaseSettings(BaseSettings):
>     host: str
>     port: int = 5432
>    
>     model_config = SettingsConfigDict(
>       env_prefix="DB_"  # DB_HOST, DB_PORT として読み込む
>     )
>   class Settings(BaseSettings):
>     database: DatabaseSettings
>    
>     model_config = SettingsConfigDict(
>       env_nested_delimiter="__"  # DATABASE__HOST のような形式
>     )
>   ```