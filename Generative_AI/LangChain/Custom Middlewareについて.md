## 概要

現状のLangChain/LangGraphには**時間ベースの制御用の組み込みミドルウェアは存在しない**が、`ModelCallLimitMiddleware`
のパターンを参考に、カスタムミドルウェアを作成できる。

## 実装方法

### 基本的な TimeoutMiddleware

```python
import time
from typing import Annotated, Any, NotRequired
from langchain_core.messages import AIMessage
from langgraph.channels.untracked_value import UntrackedValue
from langchain.agents.middleware import AgentMiddleware, AgentState, hook_config


class TimeoutState(AgentState[Any]):
    """State schema for TimeoutMiddleware."""
    run_start_time: NotRequired[Annotated[float, UntrackedValue]]


class TimeoutMiddleware(AgentMiddleware[TimeoutState, Any]):
    """時間ベースでエージェントを制御するミドルウェア."""

    state_schema = TimeoutState

    def __init__(
        self,
        *,
        timeout_seconds: float,
        graceful_exit_message: str | None = None,
    ) -> None:
        super().__init__()
        self.timeout_seconds = timeout_seconds
        self.graceful_exit_message = graceful_exit_message or (
            f"タイムアウト（{timeout_seconds}秒）に達しました。"
            "これまでの情報を元に回答します。"
        )

    def before_agent(self, state: TimeoutState, runtime) -> dict[str, Any] | None:
        """エージェント開始時に開始時間を記録."""
        return {"run_start_time": time.time()}

    async def abefore_agent(self, state: TimeoutState, runtime) -> dict[str, Any] | None:
        return self.before_agent(state, runtime)

    @hook_config(can_jump_to=["end"])
    def before_model(self, state: TimeoutState, runtime) -> dict[str, Any] | None:
        """モデル呼び出し前に経過時間をチェック."""
        start_time = state.get("run_start_time")
        if start_time is None:
            return None

        elapsed = time.time() - start_time
        if elapsed >= self.timeout_seconds:
            return {
                "jump_to": "end",
                "messages": [AIMessage(content=self.graceful_exit_message)],
            }
        return None

    @hook_config(can_jump_to=["end"])
    async def abefore_model(self, state: TimeoutState, runtime) -> dict[str, Any] | None:
        return self.before_model(state, runtime)
```

### 使用方法

```python
from langchain.agents import create_agent

agent = create_agent(
    model="anthropic:claude-sonnet-4-20250514",
    tools=[...],
    middleware=[
        TimeoutMiddleware(
            timeout_seconds=60.0,
            graceful_exit_message="時間制限に達しました。これまでの調査結果を元に回答します。",
        ),
    ],
)
```

## より高度な実装

### 残り時間に応じた動的プロンプト

```python
from langchain.agents.middleware import dynamic_prompt, ModelRequest

@dynamic_prompt
def time_aware_prompt(request: ModelRequest) -> str:
    start_time = request.state.get("run_start_time")
    timeout = 60.0

    if start_time is None:
        return "あなたは役立つアシスタントです。"

    elapsed = time.time() - start_time
    remaining = timeout - elapsed

    if remaining < 10:
        return (
            "あなたは役立つアシスタントです。\n\n"
            "【重要】時間制限が迫っています。"
            "これまでの調査結果を元に、すぐに最終回答を出してください。"
            "追加のツール呼び出しは行わないでください。"
        )
    elif remaining < 30:
        return (
            "あなたは役立つアシスタントです。\n\n"
            "【注意】残り時間が少なくなっています。"
            "必要最小限のツール呼び出しのみ行ってください。"
        )

    return "あなたは役立つアシスタントです。"
```

### モデルに「まとめ」を指示するミドルウェア

`jump_to: "end"` + `AIMessage` だと即座に終了して固定メッセージを返すだけ。モデルに「今までの結果でまとめて」と指示したい場合は `HumanMessage`
を注入する方が良い：

```python
from langchain_core.messages import HumanMessage

class GracefulTimeoutMiddleware(AgentMiddleware[TimeoutState, Any]):
    """タイムアウト前にモデルに「まとめ」を指示するミドルウェア."""

    state_schema = TimeoutState

    def __init__(
        self,
        *,
        timeout_seconds: float,
        warning_threshold: float = 10.0,
    ) -> None:
        super().__init__()
        self.timeout_seconds = timeout_seconds
        self.warning_threshold = warning_threshold
        self._warned = False

    def before_agent(self, state: TimeoutState, runtime) -> dict[str, Any] | None:
        self._warned = False
        return {"run_start_time": time.time()}

    def before_model(self, state: TimeoutState, runtime) -> dict[str, Any] | None:
        start_time = state.get("run_start_time")
        if start_time is None:
            return None

        elapsed = time.time() - start_time
        remaining = self.timeout_seconds - elapsed

        if remaining <= self.warning_threshold and not self._warned:
            self._warned = True
            return {
                "messages": [
                    HumanMessage(
                        content=(
                            "【システム通知】時間制限に達しました。"
                            "これ以上のツール呼び出しは行わず、"
                            "これまでの調査結果を元に最終回答を出してください。"
                        )
                    )
                ]
            }
        return None
```

## アプローチの比較

| アプローチ | 動作 | 適用場面 |
|-----------|------|----------|
| `jump_to: "end"` + `AIMessage` | 即座に終了、固定メッセージを返す | 強制終了したい場合 |
| `HumanMessage` を注入 | モデルが指示を読んで回答を生成 | 今までの結果を活かした回答が欲しい場合 |
| `dynamic_prompt` で残り時間を伝える | システムプロンプトで行動を誘導 | より自然な誘導をしたい場合 |

## 推奨：ハイブリッドアプローチ

残り時間に余裕がある段階で `HumanMessage` や `dynamic_prompt` で「まとめて」と促し、本当にタイムアウトしたら `jump_to: "end"` で強制終了する。

## 注意点

| 項目 | 説明 |
|------|------|
| チェックタイミング | `before_model` でチェックするため、ツール実行中はタイムアウトしない |
| ツール実行時間 | 長時間かかるツール（API呼び出しなど）の実行中はチェックされない |
| 正確性 | 厳密なリアルタイム制御ではない。次のモデル呼び出し時にチェックされる |

## 厳密な制御が必要な場合

`asyncio.timeout()` や `asyncio.wait_for()` でエージェント全体をラップする：

```python
import asyncio

async def run_with_timeout(agent, input_data, timeout_seconds=60):
    try:
        return await asyncio.wait_for(
            agent.ainvoke(input_data),
            timeout=timeout_seconds
        )
    except asyncio.TimeoutError:
        return {"messages": [AIMessage(content="タイムアウトしました")]}
```
