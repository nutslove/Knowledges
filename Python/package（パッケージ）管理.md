# 自作パッケージの作成と管理
- Pythonでは、自作のパッケージを作成して管理することができる。パッケージは、関連するモジュールをまとめたディレクトリ構造であり、コードの再利用性と整理に役立つ。
- `setup.py`を使う方法と`pyproject.toml`を使う方法がある。

## 1. `pyproject.toml`を使う方法（推奨）
- `pyproject.toml`は、Pythonのビルドシステムとパッケージングのための設定ファイル。
- `setuptools`や`poetry`などのビルドツールを指定できる。
- 例: `pyproject.toml`
  ```toml
  [build-system]
  requires = ["setuptools>=42", "wheel"]
  build-backend = "setuptools.build_meta"

  [project]
  name = "my_package"
  version = "0.1.0"
  description = "A sample Python package"
  authors = [{name = "Your Name", email = "your.email@example.com"}]
  license = {text = "MIT"}
  dependencies = ["requests>=2.20.0"]
  ```

- パッケージのディレクトリ構成例:
  ```bash
  my_package/
      __init__.py
      module1.py
      module2.py
  ```

### `pyproject.toml`内の各セクションとフィールドの説明:
- `[build-system]`: ビルドシステムの要件とバックエンドを指定。
  - `requires`: ビルドに必要なパッケージ。
  - `build-backend`: 使用するビルドバックエンド。
- `[project]`: パッケージのメタデータを定義。
  - `name`: パッケージ名。
  - `version`: バージョン番号。
  - `description`: パッケージの説明。
  - `authors`: 作者情報（名前とメールアドレス）。
  - `license`: ライセンス情報。
  - `dependencies`: パッケージが依存する他のパッケージ。