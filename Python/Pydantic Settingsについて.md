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