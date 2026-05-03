# Pythonプロジェクトのディレクトリ構成ベストプラクティス（uv版）

uvを使ったPythonプロジェクトの推奨構成と運用方法のまとめ。

## 基本構成（src layout）

PyPA（Python Packaging Authority）が推奨する **src layout** を採用する。

```
my_project/
├── src/
│   └── my_package/
│       ├── __init__.py
│       ├── main.py
│       ├── core/
│       │   ├── __init__.py
│       │   └── logic.py
│       └── utils/
│           ├── __init__.py
│           └── helpers.py
├── tests/
│   ├── __init__.py
│   ├── conftest.py
│   ├── unit/
│   │   └── test_logic.py
│   └── integration/
│       └── test_api.py
├── docs/
│   └── index.md
├── scripts/
│   └── deploy.sh
├── .gitignore
├── .python-version
├── .env.example
├── pyproject.toml
├── uv.lock
├── README.md
└── LICENSE
```

`src/`配下にパッケージを置くことで、インストールされた状態でテストが実行されることを保証できる。これによりパッケージング時の問題に早期に気づける。

## アプリケーション系（Web API、AI Agentなど）

ライブラリではなくアプリケーションの場合、ドメインごとにレイヤーを分ける。LangGraphベースのAIエージェントやFastAPIアプリでよく使う構成。

```
my_app/
├── src/
│   └── my_app/
│       ├── __init__.py
│       ├── main.py              # エントリポイント
│       ├── config.py            # 設定（pydantic-settings）
│       ├── api/                 # ルーター層
│       │   ├── __init__.py
│       │   ├── routes.py
│       │   └── deps.py
│       ├── domain/              # ドメインモデル
│       │   ├── __init__.py
│       │   └── models.py
│       ├── services/            # ビジネスロジック
│       │   ├── __init__.py
│       │   └── rca_service.py
│       ├── agents/              # LangGraphエージェント定義
│       │   ├── __init__.py
│       │   ├── graph.py
│       │   ├── nodes.py
│       │   └── tools.py
│       ├── infrastructure/      # 外部連携
│       │   ├── __init__.py
│       │   ├── bedrock_client.py
│       │   ├── prometheus_client.py
│       │   └── langfuse_client.py
│       └── schemas/             # Pydanticスキーマ
│           ├── __init__.py
│           └── api_schemas.py
├── tests/
│   ├── conftest.py
│   ├── unit/
│   └── integration/
├── docker/
│   └── Dockerfile
├── .env.example
├── pyproject.toml
├── uv.lock
└── README.md
```

ポイントは **api → services → infrastructure** の単方向依存にすること。services層が外部APIの詳細を知らないようにする。

## `__init__.py`について

`__init__.py`は、そのディレクトリを **Pythonパッケージとして扱うためのマーカーファイル**。空ファイルでも構わない。

### 厳密には必須ではない

Python 3.3以降、**Namespace Packages**（PEP 420）が導入され、`__init__.py`がなくてもパッケージとして機能するようになった。ただし実プロジェクトでは **明示的に置く（Regular Package にする）のがベストプラクティス**。

### 置くべき理由

1. **ツールの挙動が安定する**
   pytest、mypy、IDE（VSCode、PyCharm）は`__init__.py`があるディレクトリを明確にパッケージとして認識する。Namespace Packageだとツールによってはパッケージ検出に失敗したり、import解決がおかしくなることがある。特にpytestでは`__init__.py`の有無でテスト収集の挙動（rootdir判定、conftest.pyのスコープ）が変わる。

2. **意図せぬディレクトリがパッケージ化されるのを防ぐ**
   Namespace Packageだと`__init__.py`がないディレクトリも勝手にパッケージとして認識される。これがimportの曖昧さやバグの原因になることがある。

3. **パッケージレベルの初期化処理が書ける**
   パッケージ読み込み時に実行したい処理を記述できる。

   ```python
   # src/my_app/__init__.py
   from my_app.config import settings
   from my_app.logger import setup_logging

   __version__ = "0.1.0"

   setup_logging()
   ```

4. **パブリックAPIを定義できる**
   `__all__` や再エクスポートで、パッケージの外向けインターフェースを明示できる。

   ```python
   # src/my_app/services/__init__.py
   from my_app.services.rca_service import RCAService
   from my_app.services.summary_service import SummaryService

   __all__ = ["RCAService", "SummaryService"]
   ```

   これで利用側は `from my_app.services import RCAService` と簡潔に書ける。

### `tests/`配下の`__init__.py`は別

`tests/`ディレクトリには **`__init__.py`を置かないのが推奨**。pytestの挙動（rootdir方式）と相性が良く、テストをパッケージ化しないことで余計な依存関係が発生しない。

ただし、テストファイル間で同名のモジュールがある場合（例: `tests/unit/test_logic.py` と `tests/integration/test_logic.py`）は、衝突回避のために`__init__.py`を置く必要がある。

### まとめ

- `src/my_package/`配下のディレクトリには **必ず`__init__.py`を置く**（空でOK）
- `tests/`配下は **基本置かない**。名前衝突がある場合のみ置く
- 必須ではないが、ツール互換性と意図の明確化のために置くのが現代的なベストプラクティス

## pyproject.tomlの設定例

`setup.py`、`setup.cfg`、`requirements.txt`は使わず、すべて`pyproject.toml`に集約する。

```toml
[project]
name = "my_app"
version = "0.1.0"
description = "AI Agent for Root Cause Analysis"
requires-python = ">=3.12"
dependencies = [
    "fastapi>=0.115",
    "pydantic>=2.9",
    "pydantic-settings>=2.5",
    "langgraph>=0.2",
    "langchain-aws>=0.2",
]

[dependency-groups]
dev = [
    "pytest>=8.3",
    "pytest-asyncio>=0.24",
    "pytest-cov>=5.0",
    "ruff>=0.7",
    "mypy>=1.13",
    "pre-commit>=4.0",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP", "B", "SIM", "RUF"]

[tool.mypy]
strict = true
python_version = "3.12"

[tool.pytest.ini_options]
testpaths = ["tests"]
asyncio_mode = "auto"
addopts = "--cov=src --cov-report=term-missing"
```

## 必須の補助ファイル

### .gitignore

```gitignore
# Python
__pycache__/
*.py[cod]
*$py.class
*.egg-info/
dist/
build/

# venv
.venv/

# 環境変数
.env
.env.local

# IDE
.vscode/
.idea/

# テスト・カバレッジ
.pytest_cache/
.coverage
htmlcov/
.mypy_cache/
.ruff_cache/
```

`.python-version`と`uv.lock`は **コミット対象**。除外しないように注意。

### .python-version

Pythonバージョンを固定するファイル（uvが自動生成）。

```
3.12
```

### .env.example

実際の`.env`はコミットせず、テンプレートだけ共有する。

```bash
# AWS
AWS_REGION=ap-northeast-1
BEDROCK_MODEL_ID=anthropic.claude-3-5-sonnet-20241022-v2:0

# Langfuse
LANGFUSE_PUBLIC_KEY=
LANGFUSE_SECRET_KEY=
LANGFUSE_HOST=https://cloud.langfuse.com

# Application
LOG_LEVEL=INFO
```

### pre-commit設定（.pre-commit-config.yaml）

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.7.0
    hooks:
      - id: ruff
        args: [--fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.13.0
    hooks:
      - id: mypy
        additional_dependencies: [pydantic, types-requests]
```

セットアップ:

```bash
uv add --dev pre-commit
uv run pre-commit install
```

## テストの構成

```
tests/
├── conftest.py          # pytest共通fixture
├── unit/                # 単体テスト（外部依存なし）
│   ├── test_logic.py
│   └── services/
│       └── test_rca_service.py
├── integration/         # 統合テスト（DB、API呼び出しあり）
│   └── test_bedrock.py
└── e2e/                 # E2Eテスト
    └── test_full_flow.py
```

テストはプロダクションコードを **ミラーする構造** にすると見通しが良い。`src/my_app/services/foo.py` に対して `tests/unit/services/test_foo.py` のように対応させる。

## Dockerでの利用

uvの公式イメージを使うと、本番デプロイ用のイメージも軽量に作れる。

```dockerfile
FROM python:3.12-slim

# uvをインストール
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

WORKDIR /app

# 依存関係を先にコピー（キャッシュ効率化）
COPY pyproject.toml uv.lock ./

# 本番用依存のみインストール
RUN uv sync --frozen --no-dev

# アプリケーションコードをコピー
COPY src/ ./src/

# 仮想環境のPythonを使う
ENV PATH="/app/.venv/bin:$PATH"

CMD ["python", "-m", "my_app.main"]
```

`--frozen`は`uv.lock`を変更しないモード、`--no-dev`は開発依存をスキップする。

## ベストプラクティスのまとめ

新規プロジェクトを始めるときの推奨フロー:

1. `uv init --lib my_project` でsrc layoutで作成
2. `uv python pin 3.12` でPythonバージョン固定
3. `pyproject.toml`にruff/mypy/pytestの設定を集約
4. `uv add --dev pytest ruff mypy pre-commit` で開発ツール導入
5. `pre-commit install` でコミット時の自動チェック設定
6. `.env.example`、`.gitignore`、`README.md`を整備
7. `uv.lock`は **必ずコミット** する

これで再現性のある開発環境と高速なツールチェーンが整う。アプリケーション系（FastAPI、LangGraphなど）の場合は、上記の「アプリケーション系」のレイヤー構成を採用することで、ドメインロジックと外部依存を綺麗に分離できる。