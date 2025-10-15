# 自作パッケージの作成と管理
- Pythonでは、自作のパッケージを作成して管理することができる。パッケージは、関連するモジュールをまとめたディレクトリ構造であり、コードの再利用性と整理に役立つ。
- `setup.py`を使う方法と`pyproject.toml`を使う方法がある。

## `pyproject.toml`を使う方法（推奨）
- https://packaging.python.org/ja/latest/guides/writing-pyproject-toml

- `pyproject.toml`は、Pythonパッケージ管理の標準的な設定ファイルで、ビルド、依存関係、メタデータなどを一元管理
  - `setup.py`や`setup.cfg`、`requirements.txt`を1つのファイルに統合できる

- このファイルの中には３個の TOML テーブルを置くことが可能
  1. `[build-system]`: どのビルドバックエンドを使うのか、また、そのプロジェクトをビルドするためには他のどんな依存関係が必要なのかを宣言する
     > **The `[build-system]` table should always be present, regardless of which build backend you use (`[build-system]` defines the build tool you use).**）
     - `requires`: ビルドに必要なパッケージのリスト
     - `build-backend`: 使用するビルドバックエンド（以下が指定できる）
         - `setuptools.build_meta`: setuptoolsを使用する場合
         - `flit_core.buildapi`: Flitを使用する場合
         - `poetry.core.masonry.api`: Poetryを使用する場合
  2. `[project]`: パッケージ(プロジェクト)のメタデータ（名前、バージョン、依存関係など）を定義する
     - `name`: パッケージ名（必須 / PyPIに登録される(pip installで使う)名前）
     - `version`: バージョン番号（必須）
     - `description`: パッケージの説明
     - `requires-python`: 対応するPythonの最小バージョン
     - `authors`: 作成者情報（名前とメールアドレス）
     - `dependencies`: パッケージが依存する他のパッケージのリスト
  3. `[tool.<tool-name>]`: 各ツール固有の設定

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
> package-dir = {"my_import_name" = "."} # importで使う名前のディレクトリの場所
> ```
> 例えば、以下のようにディレクトリ名は大文字の`Common`で、importで使う名前を小文字の`common`にしたい場合は以下のようにする
> ```markdown
> Common/
> ├── __init__.py
> ├── module1.py
> ├── module2.py
> └── pyproject.toml
> ```
> ```toml
> [tool.setuptools]
> packages = ["common"]
> package-dir = {"common" = "."} # common パッケージの実体はこのディレクトリ（＝カレント）
> ```

## `__init__.py`ファイル
- パッケージディレクトリには`__init__.py`ファイルを含める必要がある
- `__init__.py`は、パッケージの初期化コードを含むことができる
- 空の`__init__.py`ファイルでもパッケージとして認識される