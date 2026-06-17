## 1. uvとは

- **Astral** (Ruff の開発元) が開発する Rust 製の Python パッケージ＆プロジェクトマネージャー
- `pip` / `pip-tools` / `pipx` / `poetry` / `pyenv` / `virtualenv` / `twine` を **1つに統合**
- pip比 **10〜100倍** 高速 (warm cache 時はほぼ瞬時)
- Python 不要で curl 1発でインストール可能 (単一バイナリ)
- macOS / Linux / Windows 対応
- ライセンス: MIT OR Apache-2.0

### なぜ速いのか
- Rust の並列処理 (parallel metadata fetch & wheel download)
- グローバルキャッシュ + ハードリンクでdeduplication（重複排除）
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

> [!NOTE]
> `.python-version` ファイルでパッチバージョンまで指定すると、自動アップグレードが無効化される (uv 0.11.8〜)。

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
├── .git/              # gitリポジトリ (親がgitリポジトリでない場合のみ作成)
├── .gitignore         # Python向けの定番エントリが自動投入される
├── .python-version    # Pythonバージョン固定
├── README.md
├── main.py
└── pyproject.toml     # プロジェクトメタデータ&依存
```

最初に `uv run` / `uv sync` / `uv lock` を実行すると `.venv/` と `uv.lock` が生成される。

> [!NOTE]
> #### `uv init` のVCS関連オプション
> ※ VCS = Version Control System (バージョン管理システム)。git / Mercurial (hg) / Subversion (svn) などの総称。今はほぼgitと同義。
>
> - デフォルトで `git init` 相当が走り `.git/` と `.gitignore` が作られる
> - 既に親ディレクトリがgitリポジトリ配下の場合は新規作成しない (既存リポジトリを尊重)
> - `--vcs none` でgit初期化を無効化
> - `--vcs git` で明示的に有効化 (既存リポジトリ内でも強制したい場合に使う)
> - `--bare` で `.git/`、`.gitignore`、`.python-version`、`README.md`、`main.py` も含めて全部スキップし `pyproject.toml` のみ生成

### 依存追加・削除

`uv add` / `uv remove` を実行すると、**3つが連動して更新される**:

1. **`pyproject.toml`** ← `[project].dependencies` (またはグループ指定時は `[dependency-groups]`) が自動更新される
2. **`uv.lock`** ← 解決済みバージョンと依存ツリーを記録
3. **`.venv/`** ← 実際にインストール (またはアンインストール)

> [!NOTE]
> `uv pip install` は **環境にインストールするだけで `pyproject.toml` は更新しない**。プロジェクトでは基本 `uv add` を使う。

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

例: `uv add fastapi` の前後の `pyproject.toml`:
```toml
# 実行前
[project]
dependencies = []

# 実行後
[project]
dependencies = [
    "fastapi>=0.115.0",
]
```

### 環境同期 (lock + install)

`uv sync` は **`pyproject.toml` と `uv.lock` の内容に従って `.venv/` を完全に同期させる** コマンド。具体的には以下の3ステップを連続で実行する:

1. **lockファイルの整合性チェック** — `pyproject.toml` と `uv.lock` がズレていないか確認 (ズレていれば `uv.lock` を更新)
2. **不足しているパッケージをインストール** — `.venv/` に無いものを追加
3. **余分なパッケージを削除** — `uv.lock` に無いのに `.venv/` にあるものを削除

つまり `.venv/` の状態を `uv.lock` に **完全に一致させる** (= synchronize する) のが `uv sync`。

#### いつ使うか
- `git pull` した後 (他の人が依存を追加・削除している可能性あり)
- `pyproject.toml` を手で編集した後
- プロジェクトを clone した直後

> [!NOTE]
> `uv add` も `uv run` も内部で `uv sync` 相当の処理を呼んでいるので、明示的に `uv sync` を打つ機会はそんなに多くない。

| コマンド | sync | lockに反映 | venvに反映 |
|---|---|---|---|
| `uv add pkg` | ✅ | 追加 | 追加 |
| `uv remove pkg` | ✅ | 削除 | 削除 |
| `uv run cmd` | ✅ | (変更なし) | 必要なら更新 |
| `uv sync` | ✅ | (変更なし) | 完全同期 |

```bash
uv sync                               # lock + 環境を最新化 (基本これでOK)
uv sync --frozen                      # ロックファイルを変更せずインストール (CI/Docker向け)
uv sync --no-dev                      # devグループを除外 (本番Docker向け / 8章参照)
uv sync --extra build                 # optional-dependenciesを含める
```

### lockファイル (`uv.lock`) とは

**プロジェクトの依存関係を、解決済みの正確なバージョンで記録したファイル。再現性のために存在する。**

#### なぜ必要か (pyproject.tomlだけでは足りない理由)
`pyproject.toml` に書くのは「ゆるい制約」:
```toml
dependencies = [
    "fastapi>=0.115",   # 0.115以上ならOK = 0.115.0でも0.118.2でも入る
    "requests",         # バージョン無指定 = 最新が入る
]
```
これだけだと、インストールするタイミングで違うバージョンが入ってしまい、「自分のマシンでは動くのに本番では動かない」現象が起きる。

#### `uv.lock` が解決すること
**実際に解決された全パッケージ (transitive dependencies含む) の正確なバージョン** を記録:
```toml
# uv.lock (抜粋・簡略化)
[[package]]
name = "fastapi"
version = "0.115.0"           # ← 正確なバージョン
source = { registry = "https://pypi.org/simple" }
dependencies = [
    { name = "pydantic" },
    { name = "starlette" },
]

[[package]]
name = "pydantic"             # ← 間接依存も記録
version = "2.9.2"
...
```
これがあれば、誰がどこで `uv sync` しても **完全に同じバージョン** がインストールされる。

#### `pyproject.toml` と `uv.lock` の役割の違い

| ファイル | 内容 | 役割 |
|---|---|---|
| `pyproject.toml` | ゆるい制約 (`>=`, `~=`) | 「何を」使いたいか宣言 |
| `uv.lock` | 厳密なバージョン (全依存ツリー) | 「正確に何を」インストールするかの記録 |

#### 運用ルール
1. **`uv.lock` は必ずgit commitする** ← 再現性のため
2. **手動で編集しない** ← uvが管理する
3. **CIでは `uv sync --frozen`** ← lockを変更させずに厳密同期
4. **依存を更新したいときだけ `uv lock --upgrade`** ← 通常はlockを尊重

#### 他ツールとの対比
- pip → lockなし (or `pip-tools` で `requirements.txt` を生成)
- go → `go.sum`
- Terraform → `.terraform.lock.hcl`
- **uv → `uv.lock`** ← 同じ思想

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

> [!NOTE]
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

PEP 735 に準拠した `[dependency-groups]` をサポート。アプリ本体の依存と、開発時だけ必要な依存を分けて管理できる。

```toml
[project]
dependencies = [
    "fastapi>=0.115",        # 本番でも必要
    "sqlalchemy>=2.0",       # 本番でも必要
]

[dependency-groups]
dev = [
    "pytest>=8",             # テスト用 (本番には不要)
    "ruff",                  # linter (本番には不要)
    "mypy",                  # type check (本番には不要)
]
production = [
    "gunicorn>=23",
]
```

- `[project].dependencies` → アプリ本体が動くのに必要なもの
- `[dependency-groups].dev` → 開発時だけ必要なもの

### `--no-dev` とは

**devグループの依存をインストールしない** オプション。**本番Docker や CI の本番ビルド** で使う。

> [!NOTE]
> ここでの **「本体依存」= `[project].dependencies` に書かれた依存**（アプリが動くのに必須なもの）のこと。uvでは慣用的に「main」「mainグループ」とも呼ばれる。`main.py` やgitの `main` ブランチとは無関係。

| | `uv sync` | `uv sync --no-dev` |
|---|---|---|
| インストールされるもの | 本体依存 + dev | 本体依存のみ |
| イメージサイズ | 大きい (pytest, ruff等込み) | 小さい |
| 攻撃対象面 | 広い (devツールが攻撃経路に) | 狭い |
| 用途 | 開発者のローカル | 本番Docker / CIの本番ビルド |

> [!IMPORTANT]
> Dockerでは **`uv sync --frozen --no-dev`** の組み合わせがお作法 (lockを変更せず + devを除外)。

### 関連オプション一覧

```bash
uv sync                       # 本体依存 + dev (デフォルト)
uv sync --no-dev              # 本体依存のみ
uv sync --only-dev            # dev のみ
uv sync --group production    # 本体依存 + dev + production
uv sync --only-group test     # test グループだけ
uv sync --no-default-groups   # グループを一切入れない
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

### Docker: シングルステージ (簡易・開発向け)

シンプルさ優先。`uv` バイナリも最終イメージに残るが、書きやすい。

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

### Docker: マルチステージ (本番推奨)

ビルド用ステージで `.venv/` を作り、ランタイムステージには **venvとアプリのソースだけ** をコピーする。`uv` バイナリは最終イメージに含めない。

```dockerfile
# ===== Builder stage =====
FROM python:3.13-slim AS builder

# uvバイナリをコピー
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# uvの環境変数 (Docker推奨設定)
ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PYTHON_DOWNLOADS=0

WORKDIR /app

# 1. 依存だけ先にインストール (キャッシュ効率化)
#    --no-install-project でプロジェクト本体は入れない
COPY pyproject.toml uv.lock ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-install-project --no-dev

# 2. ソースをコピーしてプロジェクト本体をインストール
COPY . .
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev

# ===== Runtime stage =====
FROM python:3.13-slim

WORKDIR /app

# builderから .venv/ とソースだけコピー (uvバイナリは含めない)
COPY --from=builder /app /app

# venvのbinをPATHに追加 → python が venv の python を指すようになる
ENV PATH="/app/.venv/bin:$PATH"

# uv run ではなく python を直接呼ぶ (uvが無いので)
CMD ["python", "-m", "myapp"]
```

#### 各設定の意味

- **`UV_COMPILE_BYTECODE=1`** — `.pyc` を事前生成して起動を高速化
- **`UV_LINK_MODE=copy`** — uvがhardlinkを使おうとしてDocker層間で警告が出るのを防ぐ
- **`UV_PYTHON_DOWNLOADS=0`** — uvがPythonを勝手にダウンロードしないようにする (ベースイメージのPythonを使う)
- **`--mount=type=cache`** — BuildKitのキャッシュマウント。`uv` のグローバルキャッシュをビルド間で再利用
- **`--locked`** — `uv.lock` と `pyproject.toml` の整合性を厳密に検証 (CI/本番ビルド向け、`--frozen` より安全)
- **`--no-install-project`** — プロジェクト本体は入れず、依存だけインストール (Docker層キャッシュを効かせるため)

#### `CMD` で `uv run` を使うか、`python` を直接呼ぶか

| | `CMD ["uv", "run", "python", "-m", "myapp"]` | `CMD ["python", "-m", "myapp"]` |
|---|---|---|
| 起動速度 | 整合性チェックが入る分わずかに遅い | 最速 (venvのpythonを直接実行) |
| イメージサイズ | `uv` バイナリ (~35MB) を含む | `uv` 不要 |
| 起動時の挙動 | `pyproject.toml` ↔ `lock` ↔ 環境 を毎回チェック | 何もチェックしない (immutable) |
| ネットワーク | lockがズレていると外部参照する可能性 | 一切しない |
| 推奨用途 | シングルステージ / 開発・CI | マルチステージ / 本番 (K8s等) |

**K8sのようなimmutableなコンテナ運用では、起動時に外部状態を参照しないほうが望ましい**ので、マルチステージ + `python` 直接呼びが推奨される。

#### マルチステージのメリット
- 最終イメージに `uv` バイナリ、ビルドキャッシュ、ビルド用ツールが含まれない → サイズが小さい
- 攻撃対象面が小さい (本番に不要なものが入っていない)
- `pyproject.toml` / `uv.lock` も最終イメージから除外できる (上記例ではコピーしているが、必要なら省ける)

#### 非rootユーザーで動かす (本番では推奨)
```dockerfile
# Runtime stageに追加
RUN useradd -m -u 1000 app && chown -R app:app /app
USER app
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
| `UV_PYTHON_DOWNLOADS=0` | uvがPython本体をダウンロードしない (Dockerで推奨) |
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