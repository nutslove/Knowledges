# 自作パッケージの作成と管理
- Pythonでは、自作のパッケージを作成して管理することができる。パッケージは、関連するモジュールをまとめたディレクトリ構造であり、コードの再利用性と整理に役立つ。
- `setup.py`を使う方法と`pyproject.toml`を使う方法がある。

## `pyproject.toml`を使う方法（推奨）
- https://packaging.python.org/ja/latest/guides/writing-pyproject-toml

- `pyproject.toml`は、Pythonパッケージ管理の標準的な設定ファイルで、ビルド、依存関係、メタデータなどを一元管理
  - `setup.py`や`setup.cfg`、`requirements.txt`を1つのファイルに統合できる

- このファイルの中には３個の TOML テーブルを置くことが可能
  1. **`[build-system]`**: どのビルドバックエンドを使うのか、また、そのプロジェクトをビルドするためには他のどんな依存関係が必要なのかを宣言する
     > **The `[build-system]` table should always be present, regardless of which build backend you use (`[build-system]` defines the build tool you use).**）
     - `requires`: ビルドに必要なパッケージのリスト
     - `build-backend`: 使用するビルドバックエンド（以下が指定できる）
         - `setuptools.build_meta`: setuptoolsを使用する場合
         - `flit_core.buildapi`: Flitを使用する場合
         - `poetry.core.masonry.api`: Poetryを使用する場合
  2. **`[project]`**: パッケージ(プロジェクト)のメタデータ（名前、バージョン、依存関係など）を定義する
     - `name`: パッケージ名（必須 / PyPIに登録される(pip installで使う)名前）
     - `version`: バージョン番号（必須）
     - `description`: パッケージの説明
     - `requires-python`: 対応するPythonの最小バージョン
     - `authors`: 作成者情報（名前とメールアドレス）
     - `dependencies`: パッケージが依存する他のパッケージのリスト
  3. **`[tool.<tool-name>]`**: 各ツール固有の設定

- 例: `pyproject.toml`
  ```toml
  [build-system]
  requires = ["setuptools>=42", "wheel"]
  build-backend = "setuptools.build_meta"

  [project]
  name = "my_package" # PyPIに登録される名前 / pip installで使う名前
  version = "0.1.0"
  description = "A sample Python package"
  authors = [{name = "Your Name", email = "your.email@example.com"}]
  license = {text = "MIT"}
  dependencies = ["requests>=2.20.0"]
  ```

- パッケージのディレクトリ構成例:
  ```bash
  my_package/ # importで使う名前（`import my_package`）
      __init__.py
      module1.py
      module2.py
  ```

> [!CAUTION]  
> 通常は、importで使う名前はディレクトリ名になるが、もしディレクトリ名とimportで使う名前を別にしたい場合は、`[tool.setuptools]`の`packages`で指定する必要がある  
> ```toml
> [tool.setuptools]
> packages = ["my_import_name"] # importで使う名前
> package-dir = {"my_import_name" = "."} # そのパッケージの実体の場所（.=カレントディレクトリ）
> ```
> 
> 例えば、以下のようにディレクトリ名は大文字の`Common`で、importで使う名前を小文字の`common`にしたい場合：
> ```
> Common/            ← ディレクトリ名（大文字）
> ├── __init__.py
> ├── module1.py
> ├── module2.py
> └── pyproject.toml
> ```
> ```toml
> [tool.setuptools]
> packages = ["common"]            # Pythonパッケージ名（小文字）
> package-dir = {"common" = "."}   # commonの実体は"."（=Common/ディレクトリ）
> ```
> ```python
> # 使用例
> import common  # ← 小文字でimport
> from common import module1
> ```

## `__init__.py`ファイル
- Python 3.3以降では、`__init__.py`ファイルがなくてもパッケージとして認識されるが、以下の理由から含めることが推奨される
  1. 互換性: 古いPythonバージョンとの互換性を保つため
  2. 明示的なパッケージ定義: このディレクトリがパッケージであることを明示的に示す
  3. 初期化処理: パッケージ読み込み時の初期化コードを書ける
  4. 公開APIの制御: `__all__` で「外からimport可能なもの」を制御する
- `__init__.py`の例:  
  ```python
  """
  共通モジュール
  """

  __version__ = "0.1.0" # バージョン情報

  # 各モジュールをimportして利用可能にする
  from . import auth
  from . import config
  from . import customlogger

  __all__ = [
      "auth",
      "config",
      "customlogger",
  ]
  ```
- `__init__.py`に`__all__`を定義することで、`from package import *`でimportされるモジュールを制御できる
  - 例えば、**上記例の場合、`from my_package import *`とすると、`auth`, `config`, `customlogger`が自動的にimportされる**
- `__init__.py`に初期化コードを書くことで、パッケージがimportされたときに特定の処理を実行できる
- `from . import <module_name>`のように相対importを使うことで、パッケージがimportされたときに、<module_name>に指定したモジュールを自動的にimportされて利用可能になる。
  - 例えば、`from . import auth`とすると、`my_package.auth.login()`のように`auth`モジュールを利用できる。
  - これがないと、`import my_package`しただけでは`my_package.auth`は利用できず、`from my_package import auth`のように明示的にimportする必要がある。