- defaultではLLMの1回の推論で複数のToolの呼び出しができる。
- LLMモデルによっては、1回の推論で複数のTool呼び出しができないように制御できるものもある。
  - OpenAIとAnthropicのモデルは`parallel_tool_calls`で制御可能
    - https://github.com/langchain-ai/langchain/blob/master/libs/partners/anthropic/langchain_anthropic/chat_models.py#L2051  
      ```python
      def bind_tools(
          self,
          tools: Sequence[dict[str, Any] | type | Callable | BaseTool],
          *,
          tool_choice: dict[str, str] | str | None = None,
          parallel_tool_calls: bool | None = None,
          strict: bool | None = None,
          **kwargs: Any,
      ) -> Runnable[LanguageModelInput, AIMessage]:
          r"""Bind tool-like objects to this chat model.

          Args:
              tools: A list of tool definitions to bind to this chat model.

                  Supports Anthropic format tool schemas and any tool definition handled
                  by `langchain_core.utils.function_calling.convert_to_openai_tool`.
              tool_choice: Which tool to require the model to call. Options are:

                  - Name of the tool as a string or as dict `{"type": "tool", "name": "<<tool_name>>"}`: calls corresponding tool
                  - `'auto'`, `{"type: "auto"}`, or `None`: automatically selects a tool (including no tool)
                  - `'any'` or `{"type: "any"}`: force at least one tool to be called
              parallel_tool_calls: Set to `False` to disable parallel tool use.

                  Defaults to `None` (no specification, which allows parallel tool use).

                  !!! version-added "Added in `langchain-anthropic` 0.3.2"
              strict: If `True`, Claude's schema adherence is applied to tool calls.

                  See the [Claude docs](https://platform.claude.com/docs/en/build-with-claude/structured-outputs#when-to-use-json-outputs-vs-strict-tool-use).
              kwargs: Any additional parameters are passed directly to `bind`.
      ```