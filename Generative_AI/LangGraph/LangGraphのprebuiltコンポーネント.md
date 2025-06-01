# Prebuiltコンポーネント
- https://langchain-ai.github.io/langgraph/agents/overview/
- React AgentやSupervisor Multi-Agentなど、スクラッチから作成せずに、LangGraphのprebuiltコンポーネントを使って簡単に作成できる

## LangGraphでSupervisorAgent作成方法 
- 参考URL
  - https://blog.kinto-technologies.com/posts/2025-02-28-building-Multi-Agent-system-by-using-langgraph-supervisor

### 設定手順
- `from langgraph_supervisor import create_supervisor`の`create_supervisor`関数から作成する方法と、スクラッチから作成する方法がある
- https://langchain-ai.github.io/langgraph/tutorials/multi_agent/agent_supervisor/

### `create_supervisor`関数から作成する方法
- `langgraph-supervisor`をインストール  
  ```shell
  pip install langgraph-supervisor
  ```
- `from langgraph.prebuilt import create_react_agent`の`create_react_agent`関数からWorker React Agentを作成
- `create_supervisor`関数からSupervisor Agentを作成  
  ```python
  from langgraph_supervisor import create_supervisor
  from langchain.chat_models import init_chat_model

  supervisor = create_supervisor(
      model=init_chat_model("openai:gpt-4.1"),
      agents=[research_agent, math_agent],
      prompt=(
          "You are a supervisor managing two agents:\n"
          "- a research agent. Assign research-related tasks to this agent\n"
          "- a math agent. Assign math-related tasks to this agent\n"
          "Assign work to one agent at a time, do not call agents in parallel.\n"
          "Do not do any work yourself."
      ),
      add_handoff_back_messages=True,
      output_mode="full_history",
  ).compile()

  for chunk in supervisor.stream(
      {
          "messages": [
              {
                  "role": "user",
                  "content": "find US and New York state GDP in 2024. what % of US GDP was New York state?",
              }
          ]
      },
  ):
      pretty_print_messages(chunk, last_message=True)

  final_message_history = chunk["supervisor"]["messages"]
  ```