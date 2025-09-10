- `InjectedState`を使えばtoolsの中でgraph stateにアクセスできる
- https://langchain-ai.github.io/langgraph/how-tos/tool-calling/#short-term-memory
```python
from typing import Annotated
from langchain_core.tools import tool
from langgraph.prebuilt import InjectedState

class CustomState(AgentState):
    user_id: str

@tool
def get_user_info(
    state: Annotated[CustomState, InjectedState]
) -> str:
    """Look up user info."""
    user_id = state["user_id"]
    return "User is John Smith" if user_id == "user_123" else "Unknown user"
```

## stateの中の一部の項目だけToolに渡す方法
- `InjectedState`を渡すとStateの中のすべての項目が連携される。しかし、Toolの中で必要なデータはStateの中の一部の項目(データ)だけの場合が多く、State全体を連携すると無駄なデータを連携することになる。
- **以下のように特定の項目だけを連携することができる**（`foo: Annotated[str, InjectedState("foo")]`の部分）  
  ```python
  from typing import List
  from typing_extensions import Annotated, TypedDict
  from langchain_core.messages import BaseMessage, AIMessage
  from langchain_core.tools import tool
  from langgraph.prebuilt import InjectedState, ToolNode

  class AgentState(TypedDict):
      messages: List[BaseMessage]
      foo: str

  @tool
  def state_tool(x: int, state: Annotated[dict, InjectedState]) -> str:
      '''Do something with state.'''
      if len(state["messages"]) > 2:
          return state["foo"] + str(x)
      else:
          return "not enough messages"

  @tool
  def foo_tool(x: int, foo: Annotated[str, InjectedState("foo")]) -> str:
      '''Do something else with state.'''
      return foo + str(x + 1)

  node = ToolNode([state_tool, foo_tool])

  tool_call1 = {"name": "state_tool", "args": {"x": 1}, "id": "1", "type": "tool_call"}
  tool_call2 = {"name": "foo_tool", "args": {"x": 1}, "id": "2", "type": "tool_call"}
  state = {
      "messages": [AIMessage("", tool_calls=[tool_call1, tool_call2])],
      "foo": "bar",
  }
  node.invoke(state)
  ```