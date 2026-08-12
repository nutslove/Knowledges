# FastAPIの設定管理（pydantic-settings）について

環境変数や`.env`ファイルから設定値（DB接続文字列、APIキー、デバッグフラグなど）を読み込むには、**`pydantic-settings`** の `BaseSettings` を使うのが定番。[APIRouterについて](APIRouterについて.md)の推奨ディレクトリ構成に出てくる `core/config.py` がまさにこの用途。

> 出典: [FastAPI公式 - Settings and Environment Variables](https://fastapi.tiangolo.com/advanced/settings/)

---

## 1. なぜ必要か

設定値をコードにベタ書きすると、環境（ローカル/ステージング/本番）ごとに差し替えるたびにコード変更が必要になる。環境変数から読み込む形にしておけば、**コードは変えずデプロイ環境の設定だけ変える**ことができる。`pydantic-settings`を使うメリットは、これを**型付き・バリデーション付き**で行える点。

```bash
pip install pydantic-settings
```

> [!NOTE]
> Pydantic v1では`BaseSettings`は`pydantic`本体に含まれていたが、**Pydantic v2からは`pydantic-settings`という別パッケージに分離**された。`from pydantic import BaseSettings`は動かなくなるので注意（`from pydantic_settings import BaseSettings`が正しい）。

---

## 2. 基本形

```python
# core/config.py
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_name: str = "My API"
    debug: bool = False
    database_url: str                  # デフォルト値なし → 必須（環境変数に無いと起動時エラー）
    secret_key: str
    allowed_hosts: list[str] = ["localhost"]

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


settings = Settings()
```

- フィールド名（`database_url`）は**大文字小文字を区別せず**環境変数名（`DATABASE_URL`）にマッチする（デフォルトの挙動）
- デフォルト値の無いフィールドは**必須**。環境変数にも`.env`にも無ければ、アプリ起動時に`ValidationError`で落ちる（→設定漏れを早期に検知できる）
- `list[str]`のような複合型も、環境変数側は`ALLOWED_HOSTS='["a.com", "b.com"]'`のようにJSON文字列で渡せば自動でパースされる

```bash
# .env
DATABASE_URL=postgresql://user:pass@localhost/db
SECRET_KEY=super-secret
DEBUG=true
```

---

## 3. `Depends`経由で使う（キャッシュ付き）

`settings`をモジュールグローバルとして直接importして使うこともできるが、公式は**依存性注入 + `lru_cache`**を推奨している。テスト時に差し替えやすくなるため（→[テスト（TestClient）について](テスト（TestClient）について.md)の`dependency_overrides`）。

```python
# core/config.py
from functools import lru_cache
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "My API"
    database_url: str


@lru_cache
def get_settings() -> Settings:
    return Settings()   # 一度だけインスタンス化し、以後はキャッシュを返す
```

```python
# main.py
from typing import Annotated
from fastapi import Depends, FastAPI
from core.config import Settings, get_settings

app = FastAPI()

SettingsDep = Annotated[Settings, Depends(get_settings)]


@app.get("/info")
async def info(settings: SettingsDep):
    return {"app_name": settings.app_name}
```

- `@lru_cache`により、`Settings()`（環境変数読み込み・バリデーション）は**リクエストごとではなく初回の1回だけ**実行される
- テストでは`app.dependency_overrides[get_settings] = lambda: Settings(database_url="sqlite:///test.db")`のように差し替えられる

---

## 4. `lifespan`との役割分担

[lifespanについて](lifespanについて.md)の起動処理は「重いリソースの初期化」、`Settings`は「設定値そのものの読み込み」という役割分担になる。多くの場合、`lifespan`内で`get_settings()`を呼んで、その値を使ってDB接続プール等を作る。

```python
from contextlib import asynccontextmanager
from core.config import get_settings


@asynccontextmanager
async def lifespan(app: FastAPI):
    settings = get_settings()
    app.state.db_pool = create_pool(settings.database_url)   # 設定値を使ってリソースを準備
    yield
    await app.state.db_pool.close()


app = FastAPI(lifespan=lifespan)
```

---

## 5. 環境ごとに`.env`を切り替える

```python
import os
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_name: str = "My API"
    database_url: str

    model_config = SettingsConfigDict(
        env_file=f".env.{os.getenv('ENVIRONMENT', 'development')}",
        env_file_encoding="utf-8",
    )
```

```
.env.development
.env.staging
.env.production
```

- `ENVIRONMENT=production`のようにOS環境変数で切り替え先を指定し、対応する`.env.*`ファイルを読み込ませる
- **本番では`.env`ファイル自体を使わず、K8sのSecret/ConfigMapやAWS Secrets Managerから環境変数として直接注入する**構成も多い。`pydantic-settings`は「環境変数があればそれを優先し、無ければ`.env`を見る」という優先順位なので、両方の運用に対応できる

### 値の優先順位（デフォルト）

1. 実際に設定されている**OS環境変数**
2. `.env`ファイルの値
3. `Settings`クラスに書いたデフォルト値

---

## 6. ネストした設定（グループ化）

```python
from pydantic import BaseModel
from pydantic_settings import BaseSettings


class DatabaseSettings(BaseModel):
    host: str = "localhost"
    port: int = 5432
    name: str = "app"


class Settings(BaseSettings):
    app_name: str = "My API"
    database: DatabaseSettings = DatabaseSettings()

    model_config = {"env_nested_delimiter": "__"}   # DATABASE__HOST=... のように渡す
```

```bash
DATABASE__HOST=db.internal
DATABASE__PORT=5432
```

---

## 7. Secretな値の扱い

パスワードやAPIキーなど、ログに誤って出力したくない値には`pydantic.SecretStr`を使う。

```python
from pydantic import SecretStr

class Settings(BaseSettings):
    secret_key: SecretStr

# 使うときは .get_secret_value() で取り出す
settings.secret_key.get_secret_value()

# print(settings) や repr() では "**********" と表示され、値が漏れない
```

---

## ポイントまとめ

- 環境変数/`.env`からの設定読み込みは`pydantic-settings`の`BaseSettings`が定番（Pydantic v2では別パッケージ、インポート元に注意）
- デフォルト値の無いフィールドは必須になり、**設定漏れを起動時に検知**できるのが利点
- `@lru_cache`を付けた`get_settings()`を`Depends`経由で使うと、リクエストごとの再読込を避けつつテスト時の差し替え（[テスト（TestClient）について](テスト（TestClient）について.md)）もしやすい
- `.env`は開発用途、本番は環境変数を直接注入（K8s Secret等）という組み合わせがよくある。優先順位は「OS環境変数 > `.env` > デフォルト値」
- パスワード等は`SecretStr`で扱い、ログ出力時の値漏洩を防ぐ
- 起動時に読み込んだ設定値は、[lifespanについて](lifespanについて.md)のリソース初期化にそのまま渡すのが典型的な流れ
