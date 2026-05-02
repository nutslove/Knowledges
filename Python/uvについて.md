## 1. uvとは

- **Astral** (Ruff の開発元) が開発する Rust 製の Python パッケージマネージャー
- `pip` / `pip-tools` / `pipx` / `poetry` / `pyenv` / `virtualenv` / `twine` を **1つに統合**
- pip比 **10〜100倍** 高速 (warm cache 時はほぼ瞬時)
- Python 不要で curl 1発でインストール可能 (単一バイナリ)
- macOS / Linux / Windows 対応

### なぜ速いのか
- Rust の並列処理 (parallel metadata fetch & wheel download)
- グローバルキャッシュ + ハードリンクでdeduplication(重複排除)
- 効率的な依存解決アルゴリズム

---

## 2. インストール

### macOS / Linux (推奨: 公式インストーラ)
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### Homebrew
```bash
brew install uv
```

### pip / pipx 経由
```bash
pip install uv
# または
pipx install uv
```

### アップデート / アンインストール
```bash
uv self update         # 公式インストーラで入れた場合
uv self uninstall
```

---

## 3. Python バージョン管理

uvがCPython本体を管理してくれる (pyenv不要)。

```bash
# 利用可能なPythonバージョンを確認
uv python list

# 複数バージョンを一括インストール
uv python install 3.12 3.13 3.14

# プロジェクトで使うバージョンを固定 (.python-version を作成)
uv python pin 3.13

# 特定バージョンでvenvを作成
uv venv --python 3.12
```

メモ: `.python-version` ファイルでパッチバージョンまで指定すると、自動アップグレードが無効化される (uv 0.11.8〜)。

---

## 4. プロジェクト管理 (推奨ワークフロー)

### プロジェクト作成
```bash
uv init my-project
cd my-project
```

生成されるファイル:
```
my-project/
├── .git
├── .gitignore
├── .python-version    # Pythonバージョン固定
├── README.md
├── main.py
└── pyproject.toml     # プロジェクトメタデータ&依存
```

最初に `uv run` / `uv sync` / `uv lock` を実行すると `.venv/` と `uv.lock` が生成される。

### 依存追加・削除
```bash
uv add requests                       # 通常の依存
uv add 'httpx>=0.27,<1.0'             # バージョン制約付き
uv add ruff --dev                     # devグループ (--group dev と同等)
uv add pytest --group test            # 任意のグループ
uv add 'fastapi[standard]'            # extras付き
uv add git+https://github.com/psf/requests   # gitから
uv add -r requirements.txt            # requirements.txtから移行

uv remove requests
uv remove pytest --group test
```

### 環境同期 (lock + install)
```bash
uv sync                               # lock + 環境を最新化 (基本これでOK)
uv sync --frozen                      # ロックファイルを変更せずインストール
uv sync --no-dev                      # devグループを除外
uv sync --extra build                 # optional-dependenciesを含める
```

### lockファイル操作
```bash
uv lock                               # lockファイルを生成・更新
uv lock --check                       # pyproject.tomlとの整合性チェック (CI向け)
uv lock --upgrade                     # 全パッケージをアップグレード
uv lock --upgrade-package fastapi     # 特定パッケージのみアップグレード
```

### コマンド実行
```bash
uv run python main.py                 # 自動でsyncしてから実行
uv run pytest
uv run --no-project ruff check        # プロジェクトを無視して実行
uv run --with rich python -c "..."    # 一時的に依存を追加して実行
```

> `uv run` は実行前に毎回 `pyproject.toml` ↔ `uv.lock` ↔ 環境 の整合性を自動チェックする。venvのactivate不要。

---

## 5. pip互換インターフェース

既存のpipワークフローをそのまま高速化したい場合:

```bash
# venv作成
uv venv                               # .venv/ を作成
uv venv --python 3.13

# install/uninstall
uv pip install requests
uv pip install -r requirements.txt
uv pip install -e .                   # editable install
uv pip uninstall requests

# requirements.txt生成 (pip-compile相当)
uv pip compile requirements.in -o requirements.txt
uv pip compile pyproject.toml --universal -o requirements.txt   # クロスプラットフォーム

# 環境同期 (pip-sync相当)
uv pip sync requirements.txt          # ロックと完全一致させる (余分なのは削除)

# 確認系
uv pip list
uv pip tree
uv pip tree --outdated                # 古いパッケージを確認
uv pip show requests
uv pip freeze
```

---

## 6. ツール管理 (pipx代替)

CLIツールを隔離環境にインストール。

```bash
# 一時実行 (uvxは uv tool run のエイリアス)
uvx ruff check .
uvx --from 'huggingface_hub[cli]' huggingface-cli

# 永続インストール
uv tool install ruff
uv tool install black
uv tool list
uv tool upgrade ruff
uv tool uninstall ruff
```

---

## 7. スクリプト管理 (PEP 723対応)

単一ファイルスクリプトに依存を埋め込んで実行できる。

```bash
# 依存をインライン宣言
uv add --script example.py requests rich

# 実行 (隔離venvで実行される)
uv run example.py
```

スクリプト先頭に以下のようなメタデータが追加される:
```python
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "requests",
#     "rich",
# ]
# ///
import requests
...
```

---

## 8. 依存グループ (dev / production / test 等)

PEP 735 に準拠した `[dependency-groups]` をサポート。

```toml
[project]
name = "my-project"
version = "0.1.0"
requires-python = ">=3.12"
dependencies = [
    "fastapi>=0.115",
]

[dependency-groups]
dev = [
    "pytest>=8",
    "ruff",
    "mypy",
]
production = [
    "gunicorn>=23",
]
```

```bash
uv sync                       # main + dev (デフォルト)
uv sync --no-dev              # main のみ
uv sync --group production    # production を含める
uv sync --only-group dev      # devだけ
```

---

## 9. ビルド & 公開

```bash
uv build                              # sdist + wheel をビルド
uv build --sdist
uv build --wheel

uv publish                            # PyPIへ公開
uv publish --token $PYPI_TOKEN
uv publish --index testpypi           # TestPyPIへ
```

---

## 10. ワークスペース (モノレポ)

Cargo風のワークスペースをサポート。複数プロジェクトで依存を共有。

```toml
# ルートの pyproject.toml
[tool.uv.workspace]
members = ["packages/*"]

[tool.uv.sources]
my-lib = { workspace = true }
```

---

## 11. requirements.txt 互換

```bash
# pyproject.tomlからrequirements.txtを生成
uv export --no-hashes --format requirements-txt > requirements.txt
uv export --no-dev > requirements.txt
```

---

## 12. CI/CD・Docker での使い方

### Docker (推奨パターン)
```dockerfile
FROM python:3.13-slim

# uvバイナリをコピー
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

WORKDIR /app
COPY pyproject.toml uv.lock ./

# bytecodeコンパイル有効化で起動高速化
ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy

RUN uv sync --frozen --no-dev

COPY . .
CMD ["uv", "run", "python", "-m", "myapp"]
```

### GitHub Actions
```yaml
- uses: astral-sh/setup-uv@v3
  with:
    enable-cache: true
- run: uv sync --frozen
- run: uv run pytest
```

---

## 13. よく使う環境変数

| 変数 | 用途 |
|------|------|
| `UV_COMPILE_BYTECODE=1` | インストール時に `.pyc` を生成 (Dockerで推奨) |
| `UV_LINK_MODE=copy` | hardlinkではなくcopyを使う (Dockerで推奨) |
| `UV_NO_CACHE=1` | キャッシュを使わない |
| `UV_INDEX_URL` | PyPIインデックスURL |
| `UV_EXTRA_INDEX_URL` | 追加のインデックスURL |
| `UV_PYTHON` | デフォルトPythonバージョン |
| `UV_CACHE_DIR` | キャッシュディレクトリ変更 |
| `UV_PROJECT_ENVIRONMENT` | プロジェクトvenvのパス変更 |
| `UV_KEYRING_PROVIDER=native` | OS keyringから認証情報を取得 |

---

## 14. pip / Poetry / Rye との比較

| 機能 | uv | Poetry | pip + venv | Rye |
|------|----|----|----|-----|
| 言語 | Rust | Python | Python | Rust |
| 速度 | ★★★★★ | ★★ | ★ | ★★★★ (内部でuv使用) |
| Lockfile | ✅ uv.lock | ✅ poetry.lock | ⚠️ pip-tools | ✅ |
| Pythonバージョン管理 | ✅ | ❌ | ❌ | ✅ |
| Tools実行 (pipx代替) | ✅ uvx | ❌ | ❌ | ❌ |
| Workspace | ✅ | ❌ | ❌ | ⚠️ |
| pip互換CLI | ✅ | ❌ | ✅ | ❌ |

**現状の結論**: 新規プロジェクトはほぼ uv で良い。Rye は内部で uv を使う形になっており、Poetry からの移行も `uvx pdm import` 等で容易。

---

## 15. 移行ガイド

### requirements.txt から
```bash
uv init
uv add -r requirements.txt
```

### Poetry から
```bash
# pdm-backend経由で変換するツールがある
# または手動で pyproject.toml の [project] セクションを書き換え
uvx migrate-to-uv
```

---

## 16. ハマりポイント / Tips

- **`.venv` のactivate不要**: `uv run` を癖にする
- **`uv.lock` はgitにcommitする**: 再現性のため
- **`uv pip install` と `uv add` の違い**: 前者は環境のみ更新、後者は `pyproject.toml` も更新。プロジェクトでは基本 `uv add` を使う
- **`uv run pkg` は `pyproject.toml` の制約に従ってバージョンを自動更新する**: 直接 `pkg` を呼ぶと旧バージョンのまま。意図せぬアップデートを避けたい場合は `--no-sync` か `--frozen`
- **`build isolation` を切りたい場合** (flash-attn, deepspeed 等):
  ```toml
  [tool.uv]
  no-build-isolation-package = ["flash-attn"]
  ```
- **bytecode compile**: 開発では不要、本番Dockerでは `UV_COMPILE_BYTECODE=1` 推奨
- **キャッシュ場所**: macOS は `~/.cache/uv`, Linuxは XDG準拠

---

## 17. 参考リンク

- 公式ドキュメント: https://docs.astral.sh/uv/
- GitHub: https://github.com/astral-sh/uv
- リリースノート: https://github.com/astral-sh/uv/releases
- PyPI: https://pypi.org/project/uv/
- Astral公式ブログ: https://astral.sh/blog