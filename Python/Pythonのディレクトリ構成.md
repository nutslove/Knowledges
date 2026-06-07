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

### なぜ `src/` の中にさらに同名のディレクトリを作るのか

`my_project/src/my_package/` のように外側と内側で同名になるのは、**プロジェクト名とパッケージ名がたまたま同じ**なだけで、両者の役割は別物。

```
my_app/                  ← プロジェクト（リポジトリ）全体の入れ物
├── pyproject.toml       #   設定・テスト・docsなどを含む
├── README.md
├── tests/
└── src/                 ← ソースコードを隔離する境界
    └── my_app/          ← 実際にimport / インストールされるパッケージ本体
        ├── __init__.py
        └── main.py
```

| ディレクトリ | 役割 |
|---|---|
| `my_app/`（外側） | プロジェクト／リポジトリ全体（`pyproject.toml`・`tests/`・`docs/`などを含む入れ物） |
| `src/` | ソースコードを隔離する境界 |
| `src/my_app/`（内側） | `import my_app` で読み込まれる実際のパッケージ本体（`__init__.py`を持つ） |

`import my_app` の対象になるのは内側の `src/my_app/` のみ。外側は単なる入れ物。

#### 内側に同名パッケージを置く理由

1. **import可能な「パッケージ」の単位だから**
   `__init__.py` を持つ内側のディレクトリがパッケージの実体。`import my_app` で読み込まれるのはこの内側であり、外側は単なる入れ物にすぎない。

2. **`src/` を挟むことで「インストールせずに誤ってローカルのソースをimportする事故」を防げる（src layout最大の利点）**
   `src/`を挟まないフラットレイアウト（`my_app/my_app/`）だと、プロジェクトルートで作業しているときカレントディレクトリの `my_app/` がそのままimportできてしまう。すると「パッケージング設定が壊れていてインストールでは動かないのに、開発環境ではたまたま動いていた」という不整合に気づけない。`src/`を挟むとルートから直接importできなくなり、**必ず `uv sync` / `pip install -e .`（editable install）してからimportする**形になるため、テストやCIが「実際にユーザーがインストールした状態」と同じ条件で走る。

3. **テストが確実にインストール済みパッケージを対象にする**
   `tests/`から `import my_app` したとき、src layoutならカレントディレクトリのソースではなくインストール済みパッケージを参照する。「テストは通るのに配布したら壊れている」を防げる。

4. **ルートが散らからない**
   ルート直下にパッケージコードを置くと、`tests/`・`docs/`・設定ファイルとコードが混在する。`src/`に閉じ込めることでコードと周辺ファイルが明確に分離される。

> シンプルなスクリプトや小さなツールであれば `src/` を挟まないフラットレイアウトでもよい。**ライブラリとして配布する・きちんとテストしたい場合**に src layout が推奨される（PyPA公式も推奨）。

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

## 各サブディレクトリの役割と実装例

LangGraphベースのRCA Agentを題材に、各レイヤーのコード例を示す。

### 全体の依存関係

```
api  →  services  →  agents  →  infrastructure
              ↓           ↓
           domain     schemas
              ↑
          config（全レイヤーから参照可）
```

矢印は「依存する方向」。`services`は`infrastructure`を呼ぶが、逆はNG。`domain`は誰にも依存しない純粋なモデル層。

### `config.py` — 設定管理

環境変数や設定値を一元管理する。`pydantic-settings`を使うのが定番。

```python
# src/my_app/config.py
from functools import lru_cache
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # AWS
    aws_region: str = "ap-northeast-1"
    bedrock_model_id: str = "anthropic.claude-3-5-sonnet-20241022-v2:0"

    # Langfuse
    langfuse_public_key: str
    langfuse_secret_key: str
    langfuse_host: str = "https://cloud.langfuse.com"

    # Prometheus
    prometheus_url: str = "http://prometheus:9090"

    # App
    log_level: str = "INFO"
    max_iterations: int = Field(default=10, ge=1, le=50)


@lru_cache
def get_settings() -> Settings:
    return Settings()
```

ポイント:

- `@lru_cache`でシングルトン化
- 利用側は `from my_app.config import get_settings` で取得

### `domain/` — ドメインモデル

ビジネスの本質を表現する層。**外部ライブラリやフレームワークに依存しない** 純粋なPythonコード。

```python
# src/my_app/domain/models.py
from dataclasses import dataclass
from datetime import datetime
from enum import Enum


class IncidentSeverity(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


@dataclass(frozen=True)
class Incident:
    """インシデントを表すドメインモデル"""
    id: str
    title: str
    severity: IncidentSeverity
    started_at: datetime
    affected_services: list[str]

    def is_critical(self) -> bool:
        return self.severity == IncidentSeverity.CRITICAL


@dataclass(frozen=True)
class RootCause:
    """RCA分析の結果"""
    incident_id: str
    summary: str
    contributing_factors: list[str]
    confidence_score: float  # 0.0 - 1.0

    def is_high_confidence(self) -> bool:
        return self.confidence_score >= 0.8
```

ポイント:

- `@dataclass(frozen=True)`で不変オブジェクトに
- ビジネスルール（`is_critical`、`is_high_confidence`）はモデル自身に持たせる
- ここでPydanticを使ってもよいが、I/O境界（API/DB）とドメインを混ぜたくない場合は分ける

### `schemas/` — Pydanticスキーマ（I/O境界）

API入出力やLLM出力のバリデーション用。**domainとは別物** として扱う。

```python
# src/my_app/schemas/api_schemas.py
from datetime import datetime
from pydantic import BaseModel, Field


class RCARequest(BaseModel):
    """RCA分析リクエスト"""
    incident_id: str = Field(..., description="インシデントID")
    time_range_minutes: int = Field(default=60, ge=5, le=1440)
    services: list[str] = Field(default_factory=list)


class RCAResponse(BaseModel):
    """RCA分析レスポンス"""
    incident_id: str
    summary: str
    contributing_factors: list[str]
    confidence_score: float
    analyzed_at: datetime
```

```python
# src/my_app/schemas/agent_schemas.py
from pydantic import BaseModel, Field


class RCASummary(BaseModel):
    """LLMに構造化出力させるためのスキーマ"""
    root_cause: str = Field(..., description="根本原因の説明")
    evidence: list[str] = Field(..., description="根拠となる観測事実")
    recommended_actions: list[str] = Field(..., description="推奨される対応策")
    confidence: float = Field(..., ge=0.0, le=1.0)
```

`domain`と`schemas`を分ける理由は、API仕様の変更がドメインに波及しないようにするため。両者の変換は`services`層で行う。

### `infrastructure/` — 外部連携

外部API、DB、メッセージキュー等とのやり取りを担当。**実装の詳細を隠蔽** することが目的。

```python
# src/my_app/infrastructure/bedrock_client.py
from langchain_aws import ChatBedrock
from my_app.config import get_settings


def create_bedrock_chat() -> ChatBedrock:
    """LangChain用のBedrockクライアントを生成"""
    settings = get_settings()
    return ChatBedrock(
        model_id=settings.bedrock_model_id,
        region_name=settings.aws_region,
        model_kwargs={"temperature": 0.0, "max_tokens": 4096},
    )
```

```python
# src/my_app/infrastructure/prometheus_client.py
import httpx
from datetime import datetime
from my_app.config import get_settings


class PrometheusClient:
    def __init__(self, base_url: str | None = None) -> None:
        self.base_url = base_url or get_settings().prometheus_url
        self.client = httpx.AsyncClient(base_url=self.base_url, timeout=30.0)

    async def query_range(
        self,
        query: str,
        start: datetime,
        end: datetime,
        step: str = "30s",
    ) -> dict:
        """PromQLクエリをレンジで実行"""
        response = await self.client.get(
            "/api/v1/query_range",
            params={
                "query": query,
                "start": start.timestamp(),
                "end": end.timestamp(),
                "step": step,
            },
        )
        response.raise_for_status()
        return response.json()

    async def close(self) -> None:
        await self.client.aclose()
```

```python
# src/my_app/infrastructure/langfuse_client.py
from langfuse import Langfuse
from my_app.config import get_settings


def create_langfuse() -> Langfuse:
    settings = get_settings()
    return Langfuse(
        public_key=settings.langfuse_public_key,
        secret_key=settings.langfuse_secret_key,
        host=settings.langfuse_host,
    )
```

ポイント:

- 外部ライブラリ（boto3、httpx、langfuse）への依存はここに閉じ込める
- `services`や`agents`からは抽象化されたインターフェースだけ見せる

### `agents/` — LangGraphエージェント定義

エージェントのグラフ構造、ノード、ツールを置く。

#### `tools.py` — Toolの定義

```python
# src/my_app/agents/tools.py
from datetime import datetime, timedelta
from langchain_core.tools import tool
from my_app.infrastructure.prometheus_client import PrometheusClient


@tool
async def query_metrics(promql: str, minutes_ago: int = 60) -> str:
    """Prometheusにメトリクスクエリを実行する。

    Args:
        promql: PromQLクエリ式
        minutes_ago: 何分前から取得するか（デフォルト60分）
    """
    client = PrometheusClient()
    try:
        end = datetime.now()
        start = end - timedelta(minutes=minutes_ago)
        result = await client.query_range(promql, start, end)
        return _format_metrics_result(result)
    finally:
        await client.close()


def _format_metrics_result(result: dict) -> str:
    # LLMに渡しやすい形式に整形
    ...
```

#### `nodes.py` — グラフのノード

```python
# src/my_app/agents/nodes.py
from langchain_core.messages import SystemMessage
from langgraph.graph import MessagesState
from my_app.infrastructure.bedrock_client import create_bedrock_chat
from my_app.agents.tools import query_metrics, search_logs


def planner_node(state: MessagesState) -> dict:
    """次に取るべきアクションを計画する"""
    llm = create_bedrock_chat().bind_tools([query_metrics, search_logs])
    system = SystemMessage(content=(
        "あなたはSREエンジニアです。インシデントの根本原因を特定するため、"
        "必要なメトリクスやログを順次調査してください。"
    ))
    response = llm.invoke([system, *state["messages"]])
    return {"messages": [response]}


def summarizer_node(state: MessagesState) -> dict:
    """調査結果をまとめる"""
    from my_app.schemas.agent_schemas import RCASummary

    llm = create_bedrock_chat().with_structured_output(RCASummary)
    summary = llm.invoke(state["messages"])
    return {"summary": summary}
```

#### `graph.py` — グラフの組み立て

```python
# src/my_app/agents/graph.py
from langgraph.graph import StateGraph, START, END
from langgraph.prebuilt import ToolNode
from my_app.agents.nodes import planner_node, summarizer_node
from my_app.agents.tools import query_metrics, search_logs
from my_app.agents.state import RCAState


def should_continue(state: RCAState) -> str:
    """ループ継続判定"""
    last_message = state["messages"][-1]
    if last_message.tool_calls:
        return "tools"
    if state.get("iteration", 0) >= 10:
        return "summarize"
    return "summarize"


def build_rca_graph():
    workflow = StateGraph(RCAState)

    workflow.add_node("planner", planner_node)
    workflow.add_node("tools", ToolNode([query_metrics, search_logs]))
    workflow.add_node("summarize", summarizer_node)

    workflow.add_edge(START, "planner")
    workflow.add_conditional_edges("planner", should_continue)
    workflow.add_edge("tools", "planner")
    workflow.add_edge("summarize", END)

    return workflow.compile()
```

### `services/` — ビジネスロジック

ユースケースを実装する層。**APIとエージェントの橋渡し** をする。

```python
# src/my_app/services/rca_service.py
from datetime import datetime
from my_app.agents.graph import build_rca_graph
from my_app.domain.models import RootCause
from my_app.schemas.api_schemas import RCARequest, RCAResponse


class RCAService:
    def __init__(self) -> None:
        self.graph = build_rca_graph()

    async def analyze(self, request: RCARequest) -> RCAResponse:
        # 1. リクエストをエージェント入力に変換
        initial_message = self._build_initial_prompt(request)

        # 2. エージェント実行
        result = await self.graph.ainvoke({
            "messages": [initial_message],
            "incident_id": request.incident_id,
        })

        # 3. ドメインモデルへ変換
        summary = result["summary"]
        root_cause = RootCause(
            incident_id=request.incident_id,
            summary=summary.root_cause,
            contributing_factors=summary.evidence,
            confidence_score=summary.confidence,
        )

        # 4. APIレスポンス形式に変換
        return RCAResponse(
            incident_id=root_cause.incident_id,
            summary=root_cause.summary,
            contributing_factors=root_cause.contributing_factors,
            confidence_score=root_cause.confidence_score,
            analyzed_at=datetime.now(),
        )

    def _build_initial_prompt(self, request: RCARequest) -> str:
        services = ", ".join(request.services) or "全サービス"
        return (
            f"インシデント {request.incident_id} の根本原因を分析してください。"
            f"対象サービス: {services}、"
            f"調査範囲: 過去{request.time_range_minutes}分。"
        )
```

ポイント:

- `services`は **ユースケース単位** で分ける（`rca_service.py`、`alert_service.py`等）
- API層とエージェント層を疎結合にする変換役
- ドメインモデル（`RootCause`）を経由することで、エージェント出力の変更がAPIに直接波及しないようにする

### `api/` — ルーター層（FastAPI）

HTTPエンドポイントの定義。**ロジックは持たず、`services`を呼ぶだけ**。

#### `deps.py` — 依存性注入

```python
# src/my_app/api/deps.py
from functools import lru_cache
from my_app.services.rca_service import RCAService


@lru_cache
def get_rca_service() -> RCAService:
    return RCAService()
```

#### `routes.py` — エンドポイント

```python
# src/my_app/api/routes.py
from fastapi import APIRouter, Depends, HTTPException
from my_app.api.deps import get_rca_service
from my_app.schemas.api_schemas import RCARequest, RCAResponse
from my_app.services.rca_service import RCAService

router = APIRouter(prefix="/api/v1")


@router.post("/rca/analyze", response_model=RCAResponse)
async def analyze_incident(
    request: RCARequest,
    service: RCAService = Depends(get_rca_service),
) -> RCAResponse:
    try:
        return await service.analyze(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)) from e


@router.get("/health")
async def health() -> dict:
    return {"status": "ok"}
```

### `main.py` — エントリポイント

アプリケーションの起動。

```python
# src/my_app/main.py
from contextlib import asynccontextmanager
from fastapi import FastAPI
from my_app.api.routes import router
from my_app.config import get_settings


@asynccontextmanager
async def lifespan(app: FastAPI):
    # 起動時の処理
    settings = get_settings()
    print(f"Starting app in {settings.aws_region}")
    yield
    # 終了時の処理


app = FastAPI(
    title="RCA Agent API",
    version="0.1.0",
    lifespan=lifespan,
)
app.include_router(router)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("my_app.main:app", host="0.0.0.0", port=8000, reload=True)
```

### レイヤー設計の効果

この分け方により以下が実現できる。

1. **テストしやすさ** — `services`をテストする際、`infrastructure`をモックすればLLMやPrometheusに繋がずテスト可能
2. **変更の局所化** — Bedrockから別のLLMに切り替えても、影響は`infrastructure`内に閉じる
3. **再利用性** — 同じ`RCAService`をHTTP API経由でもCLI経由でもSlack Bot経由でも呼べる
4. **責務の明確化** — レビュー時に「このコードはここに置くべきか？」の判断がしやすい

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