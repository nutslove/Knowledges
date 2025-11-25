- 参考URL
  - https://docs.langchain.com/oss/python/deepagents/overview

# 概要
- DeepAgentsもSubAgentsもLangchainの`create_agent`で作られていて、どちらもReact Agentである
　- https://github.com/langchain-ai/deepagents/blob/master/libs/deepagents/deepagents/graph.py#L149  
    ```python
    from langchain.agents import create_agent

    ・・・中略・・・

    def create_deep_agent(
        model: str | BaseChatModel | None = None,
        tools: Sequence[BaseTool | Callable | dict[str, Any]] | None = None,
        *,
        system_prompt: str | None = None,
        middleware: Sequence[AgentMiddleware] = (),
        subagents: list[SubAgent | CompiledSubAgent] | None = None,
        response_format: ResponseFormat | None = None,
        context_schema: type[Any] | None = None,
        checkpointer: Checkpointer | None = None,
        store: BaseStore | None = None,
        backend: BackendProtocol | BackendFactory | None = None,
        interrupt_on: dict[str, bool | InterruptOnConfig] | None = None,
        debug: bool = False,
        name: str | None = None,
        cache: BaseCache | None = None,
    ) -> CompiledStateGraph:

        ・・・中略・・・

        return create_agent(
            model,
            system_prompt=system_prompt + "\n\n" + BASE_AGENT_PROMPT if system_prompt else BASE_AGENT_PROMPT,
            tools=tools,
            middleware=deepagent_middleware,
            response_format=response_format,
            context_schema=context_schema,
            checkpointer=checkpointer,
            store=store,
            debug=debug,
            name=name,
            cache=cache,
        ).with_config({"recursion_limit": 1000})
    ```
  - https://github.com/langchain-ai/deepagents/blob/master/libs/deepagents/deepagents/middleware/subagents.py#L244