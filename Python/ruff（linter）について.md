# Ruff（Python Linter / Formatter）について

## 1. 概要

**Ruff（ラフ）** は、Rust製の**超高速なPython向けLinter兼Formatter**。

- **開発元**: [Astral](https://astral.sh/) 社（Pythonパッケージマネージャー `uv` と同じ会社）
- **言語**: Rust製（だから非常に高速）
- **速度**: 従来ツール（Flake8 / Black など）の **10〜100倍高速**
- **ルール数**: 900以上の組み込みルール。人気のFlake8プラグイン（flake8-bugbear 等）をネイティブ再実装
- **採用実績**: FastAPI, Pydantic, Airflow, Pandas, SciPy, Hugging Face Transformers など主要OSSで採用

### これ1つで複数ツールを代替できる

| 従来のツール | 役割 | Ruffが代替 |
|------------|------|:---------:|
| Flake8（+ 多数のプラグイン） | Lint（静的解析） | ✅ |
| Black | コードフォーマット | ✅ |
| isort | importの並び替え | ✅ |
| pyupgrade | 構文の近代化 | ✅ |
| pydocstyle | docstringチェック | ✅ |
| autoflake | 不要importの削除 | ✅ |
| bandit | セキュリティチェック（一部） | ✅ |

複数ツールの設定・依存をRuff 1つに集約できるのが大きなメリット。

---

## 2. 何ができるか

Ruffの機能は大きく2つ。

### ① Linter（`ruff check`）

コードの問題を静的に検出する。

- 未使用のimport・変数の検出
- importの並び順チェック（isort相当）
- 命名規則チェック（pep8-naming）
- バグになりやすいコードパターンの検出（flake8-bugbear）
- 古い構文の近代化提案（pyupgrade）
- 複雑なコードの簡略化提案（flake8-simplify）
- **自動修正（auto-fix）**: 検出した問題の多くを `--fix` で自動修正

### ② Formatter（`ruff format`）

Black互換のコード整形。

- インデント・改行・クォートなどを統一
- Blackとほぼ互換の整形結果（移行が容易）
- 2026年スタイルガイドに準拠

---

## 3. 導入方法（uv前提）

### プロジェクトに開発依存として追加

```bash
uv add --dev ruff
```

`pyproject.toml` の `[dependency-groups]`（または `dev` グループ）に追加され、`uv.lock` で固定される。

### 単発で使いたいだけなら（インストール不要で実行）

```bash
uvx ruff check        # uvx 経由で都度実行
```

### バージョン確認

```bash
uv run ruff version
```

---

## 4. 使用方法（uv run 経由）

uvプロジェクトでは `uv run` を付けて実行する（仮想環境のRuffが使われる）。

### Lint（チェック）

```bash
uv run ruff check                 # カレント配下をチェック
uv run ruff check path/to/code/   # パス指定
uv run ruff check --fix           # 検出した問題を自動修正（安全な修正のみ）
uv run ruff check --fix --unsafe-fixes   # 挙動が変わる可能性のある修正も適用
uv run ruff check --watch         # ファイル変更を監視して再チェック
uv run ruff check --statistics    # ルールごとの違反件数を集計表示
uv run ruff check --add-noqa      # 既存の違反箇所に noqa コメントを自動付与
```

### Format（整形）

```bash
uv run ruff format                # 整形を実行（ファイルを書き換え）
uv run ruff format --check        # 整形が必要かだけ確認（CI向け・書き換えなし）
uv run ruff format --diff         # 変更される差分を表示
```

### 典型的な開発フロー

```bash
uv run ruff check --fix    # 1. Lint + 自動修正
uv run ruff format         # 2. フォーマット
```

> ⚠️ 安全な修正（safe fix）はデフォルトで適用される。挙動を変えうる修正（unsafe fix）は `--unsafe-fixes` を付けたときだけ適用される。

---

## 5. 設定方法（pyproject.toml）

Ruffは `pyproject.toml` / `ruff.toml` / `.ruff.toml` を読む。
**uv前提なら `pyproject.toml` に集約するのがおすすめ**（uvと設定ファイルを共有できる）。

各セクションは `[tool.ruff]` プレフィックスを付ける（`ruff.toml` 単体なら省略可）。

### 設定の始め方（自動生成コマンドは無い）

`uv init` のような **config自動生成コマンドは Ruff には存在しない**（要望のIssue [#12111](https://github.com/astral-sh/ruff/issues/12111) はあるが未実装）。そのため設定は**手動で書く**のが基本。ただし以下のおかげで負担は小さい。

```bash
# ① そもそも設定ファイルが無くてもデフォルト値で動く（line-length=88, py310想定 等）
uv run ruff check

# ② 各オプションの意味・デフォルト値を調べる（手書きの参照に便利）
uv run ruff config                # 全オプション一覧
uv run ruff config line-length    # 特定オプションの説明

# ③ 現在適用中の全設定（デフォルト含む）をTOML出力 → コピペの土台にできる
uv run ruff check --show-settings
```

**運用方針**: 最初は設定を書かずデフォルトで使い、必要になった分だけ `[tool.ruff]` に追記していくのが現実的。エディタ拡張のスキーマ補完も効く。

### 基本設定

```toml
[tool.ruff]
line-length = 88          # 1行の最大文字数（デフォルト88、Black互換）
indent-width = 4          # インデント幅
target-version = "py313"  # 対象とするPythonの最低バージョン
exclude = [               # チェック/整形の対象外
    ".git",
    ".venv",
    "build",
    "dist",
]
```

### Lint設定（ルールの選択）

```toml
[tool.ruff.lint]
# 有効化するルール（プレフィックス or 個別コードで指定）
select = ["E", "F", "I", "UP", "B", "SIM", "N"]
# select を上書きせず追加したい場合は extend-select
extend-select = ["B"]
# 無効化するルール
ignore = ["E501"]         # 例: 行長チェックを無効化
# 自動修正を許可するルール
fixable = ["ALL"]
unfixable = ["B"]         # B系は自動修正させない
```

### ファイル別のルール除外

```toml
[tool.ruff.lint.per-file-ignores]
"__init__.py" = ["F401"]              # 再エクスポート用の未使用importを許可
"**/{tests,docs}/*" = ["E402"]
"*.ipynb" = ["T20"]                   # Notebookのprintを許可
```

### Formatter設定

```toml
[tool.ruff.format]
quote-style = "double"          # クォート: "double" / "single"
indent-style = "space"          # "space" / "tab"
skip-magic-trailing-comma = false
line-ending = "auto"
docstring-code-format = false   # docstring内のコードも整形するか
```

### プラグイン個別設定の例

```toml
[tool.ruff.lint.flake8-quotes]
docstring-quotes = "double"

[tool.ruff.lint.isort]
known-first-party = ["myapp"]
```

---

## 6. 主なルールカテゴリ（プレフィックス）

ルールコードは「1〜3文字のプレフィックス + 3桁の数字」（例: `F401`）。プレフィックスでまとめて選択できる。

| プレフィックス | 由来 | 内容 |
|:---:|------|------|
| `E` / `W` | pycodestyle | コーディングスタイル（エラー/警告） |
| `F` | Pyflakes | 未使用import・未定義名など論理的な問題 |
| `I` | isort | importの並び替え |
| `N` | pep8-naming | 命名規則（PEP8） |
| `UP` | pyupgrade | 古い構文の近代化 |
| `B` | flake8-bugbear | バグになりやすいパターン |
| `SIM` | flake8-simplify | コードの簡略化提案 |
| `C4` | flake8-comprehensions | 内包表記の改善 |
| `ANN` | flake8-annotations | 型アノテーション |
| `RUF` | Ruff独自 | Ruff固有のルール |
| `ALL` | （特殊） | すべてのルールを有効化 |

> **デフォルト挙動**: 何も指定しなければ `F`（Pyflakes）と `E` の一部のみ有効。Flake8と違い、`W`（警告）や複雑度（`C901`）はデフォルト無効。

---

## 7. 違反の抑制（suppression）

特定のルールを意図的に無視したい場合。

### 行単位

```python
import sys  # noqa: F401      # F401だけ抑制
result = x   # noqa            # その行の全ルールを抑制
```

### ブロック単位

```python
# ruff: disable[N803]
def foo(legacyArg1, legacyArg2):
    ...
# ruff: enable[N803]
```

### ファイル単位（ファイル先頭に記述）

```python
# ruff: noqa            # ファイル全体で全ルール抑制
# ruff: noqa: F841      # ファイル全体でF841のみ抑制
```

### 不要になった noqa を検出

```toml
[tool.ruff.lint]
extend-select = ["RUF100"]   # 使われていない noqa を警告
```

---

## 8. CI / pre-commit 連携

### CIでのチェック（書き換えずに失敗させる）

```bash
uv run ruff check                # Lint違反があれば非ゼロ終了
uv run ruff format --check       # 未整形があれば非ゼロ終了
```

### pre-commit

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.x.x          # 最新タグに合わせる
    hooks:
      - id: ruff         # Lint（--fix を付けたい場合は args: [--fix]）
        args: [--fix]
      - id: ruff-format  # Format
```

### エディタ連携

- **VS Code**: 公式拡張 [Ruff](https://marketplace.visualstudio.com/items?itemName=charliermarsh.ruff)（保存時の自動Lint/Format/import整列）
- 多くのエディタはLSP（`ruff server`）経由で連携可能

---

## 9. まとめ

- **Ruff = 速い・オールインワン**。Flake8 + Black + isort + pyupgrade などを1つで置き換えられる、今どきの定番ツール。
- uvプロジェクトなら **`uv add --dev ruff`** で導入し、**`uv run ruff check --fix` → `uv run ruff format`** が基本フロー。
- 設定は **`pyproject.toml` の `[tool.ruff]`** に集約。`select` でルールを選び、`per-file-ignores` で例外を管理する。

---

## 参考リンク

- [Ruff 公式ドキュメント](https://docs.astral.sh/ruff/)
- [設定リファレンス](https://docs.astral.sh/ruff/configuration/)
- [Linter（ルール一覧）](https://docs.astral.sh/ruff/linter/)
- [Formatter](https://docs.astral.sh/ruff/formatter/)
- [GitHub: astral-sh/ruff](https://github.com/astral-sh/ruff)
- [Astral（開発元）](https://astral.sh/ruff)
