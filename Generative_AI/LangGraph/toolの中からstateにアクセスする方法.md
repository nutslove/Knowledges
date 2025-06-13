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