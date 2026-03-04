# MCP Adapterについて
- https://github.com/langchain-ai/langchain-mcp-adapters

## LangChain MCP AdaptersのToolのエラーハンドリング
- LangChain MCP Adaptersライブラリでは、StructuredTool作成時にhandle_tool_errorパラメータを渡していない（2026-03-03時点）
  - https://github.com/langchain-ai/langchain-mcp-adapters/issues/263
  - https://github.com/langchain-ai/langchain-mcp-adapters/pull/264 （未マージ）
- そのため、MCP Toolでエラーが発生すると`ToolException`がraiseされ、Agentが異常終了する
  - `langchain_mcp_adapters/tools.py`の`_convert_call_tool_result()`で`raise ToolException(error_msg)`している
- LangGraph側でも`ToolException`を適切にキャッチできていない問題がある
  - https://github.com/langchain-ai/langgraph/issues/6449
  - https://github.com/langchain-ai/langgraph/pull/6508 （未マージ）
  - これがマージされれば、`ToolException`が自動的にエラーメッセージに変換される
- 現時点の対策として、MCPツールをラップしてエラーを文字列として返すようにする必要がある

```python
from langchain_core.tools import StructuredTool

def wrap_mcp_tools_with_error_handling(mcp_tools: list) -> list:
    """MCPツールをエラーハンドリングでラップする"""
    wrapped_tools = []
    for tool in mcp_tools:
        original_coroutine = tool.coroutine

        async def wrapped_coroutine(*args, _original=original_coroutine, _name=tool.name, **kwargs):
            try:
                return await _original(*args, **kwargs)
            except Exception as e:
                return f"Tool '{_name}' failed: {repr(e)}"

        wrapped_tool = StructuredTool(
            name=tool.name,
            description=tool.description,
            coroutine=wrapped_coroutine,
            args_schema=tool.args_schema,
        )
        wrapped_tools.append(wrapped_tool)
    return wrapped_tools

# 使用例
client = _get_mcp_client()
mcp_tools = await client.get_tools()
mcp_tools = wrap_mcp_tools_with_error_handling(mcp_tools)  # ラップを適用
native_tools = [save_user_memory, get_user_memory, delete_user_memory]
tools = mcp_tools + native_tools
```
