- https://langchain-ai.github.io/langgraph/concepts/agentic_concepts/
- https://www.anthropic.com/research/building-effective-agents


## Multi Agent Pattern
- https://langchain-ai.github.io/langgraph/concepts/multi_agent/#multi-agent-architectures
- https://blog.langchain.dev/langgraph-multi-agent-workflows/

### いつMulti Agentが必要か
- https://langchain-ai.github.io/langgraph/concepts/multi_agent/#multi-agent-systems  
  > An agent is a system that uses an LLM to decide the control flow of an application. As you develop these systems, they might grow more complex over time, making them harder to manage and scale. For example, you might run into the following problems:
  > 
  > - agent has too many tools at its disposal and makes poor decisions about which tool to call next
  > - context grows too complex for a single agent to keep track of
  > - there is a need for multiple specialization areas in the system (e.g. planner, researcher, math expert, etc.)
  >
  > To tackle these, you might consider breaking your application into multiple smaller, independent agents and composing them into a **multi-agent system**. These independent agents can be as simple as a prompt and an LLM call, or as complex as a ReAct agent (and more!).
  >
  > The primary benefits of using multi-agent systems are:
  >
  > - **Modularity**: Separate agents make it easier to develop, test, and maintain agentic systems.
  > - **Specialization**: You can create expert agents focused on specific domains, which helps with the overall system performance.
  > - **Control**: You can explicitly control how agents communicate (as opposed to relying on function calling).