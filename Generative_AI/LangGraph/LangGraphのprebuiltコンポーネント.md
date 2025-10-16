# Prebuiltコンポーネント
- https://langchain-ai.github.io/langgraph/agents/overview/
- React AgentやSupervisor Multi-Agentなど、スクラッチから作成せずに、LangGraphのprebuiltコンポーネントを使って簡単に作成できる

## LangGraphでSupervisorAgent作成方法（`create_supervisor`） 
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

---

## `create_react_agent`関数からReact Agentを作成する方法
- **https://langchain-ai.github.io/langgraph/agents/agents/**

### 最後のoutputを構造化された形式で受け取る方法
- `create_react_agent`関数の`response_format`パラメータに`BaseModel`を継承したPydanticモデルを指定する  
  ```python
  from pydantic import BaseModel
  from langgraph.prebuilt import create_react_agent

  class WeatherResponse(BaseModel):
      conditions: str

  agent = create_react_agent(
      model="anthropic:claude-3-7-sonnet-latest",
      tools=[get_weather],
      response_format=WeatherResponse  
  )

  response = agent.invoke(
      {"messages": [{"role": "user", "content": "what is the weather in sf"}]}
  )

  response["structured_response"]
  ```

### 全体的なコード例
- `BaseModel`で複数のモデルを定義し、それらを組み合わせて複雑な構造のOutputを強制させることも可能
- `NewType`を使って、複数のモデルで同じ型を使うようにすることも可能

~~~python
from pydantic import Field, BaseModel
from langgraph.prebuilt import create_react_agent
from typing import Literal, Annotated, NewType, Optional
import subprocess
from langchain_aws import ChatBedrock
from langfuse.callback import CallbackHandler

NodeId = NewType('NodeId', int)

def awscli_tool(
    command: Annotated[str, "The shell command or awscli command to execute."],
) -> str:
    """
    Use this to execute shell, awscli commands.
    
    Args:
        command: The shell command or awscli command to execute.
    Returns:
        The output of the command.
    """

    try:
        result = subprocess.run(
            command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
    except BaseException as e:
        return f"Failed to execute. Error: {repr(e)}"
    result_str = f"""Successfully executed:\n```shell\n{command}\n```\nStdout: {result.stdout}\nStderr: {result.stderr}"""
    print(result_str)
    return result_str

class Node(BaseModel):
    nodeId: NodeId = Field(..., description="The unique identifier for the node.")
    service: Literal["route53", "s3", "cloudfront", "wafv2", "targetgroup", "rds", "elasticloadbalancing", "ecs"] = Field(..., description="The name of the AWS service.(e.g., route53, s3, cloudfront, wafv2, targetgroup, etc.)")
    arn: str = Field(..., description="The ARN of the resource. If the resource is route53, this field should be the domain name.")

class Edge(BaseModel):
    from_: NodeId = Field(..., alias="from", description="The starting node number of the edge.")
    to: NodeId = Field(..., description="The ending node number of the edge.")

class AffectedResources(BaseModel):
    nodes: list[Node] = Field(..., description="List of resources")
    edges: list[Edge] = Field(..., description="List of edges representing relationships between resources")
    roots: list[str] = Field(..., description="List of root domain names that are entry points to the system")

    class Config:
        populate_by_name = True  # aliasと元のフィールド名の両方を受け付ける

def main():
    system_prompt = """
## Role
You are an expert AWS cloud infrastructure engineer. You have deep knowledge of AWS services, architecture, and best practices. You are skilled at analyzing complex AWS environments and identifying relationships between resources.

## Task
Given a list of AWS resources (ARNs) for a specific system, your task is to analyze these resources and determine how they are interconnected. You will identify relationships such as which resources are connected to others, dependencies, and any hierarchical structures.
You can use AWS CLI commands to gather additional information about the resources.

## List of AWS resources for the system (ARN):
arn:aws:rds:ap-northeast-1:123456789012:cluster:cluster-ezhmbu3o4pnx3fh4dr7qaaaaaa
arn:aws:rds:ap-northeast-1:123456789012:db:aurora3-cluster-instance-ap-northeast-1-0
arn:aws:rds:ap-northeast-1:123456789012:cluster:aurora3-cluster-ap-northeast-1
arn:aws:ec2:ap-northeast-1:123456789012:instance/i-0373780175e22aaaa
arn:aws:elasticloadbalancing:ap-northeast-1:123456789012:loadbalancer/app/app-alb/44ffe6aaa839aaaa
arn:aws:ecs:ap-northeast-1:123456789012:cluster/cluster
arn:aws:elasticloadbalancing:ap-northeast-1:123456789012:targetgroup/frontend-tg/e02c0e498f26aaaa
arn:aws:elasticloadbalancing:ap-northeast-1:123456789012:targetgroup/backend-tg/265c2d67c8920aaa
arn:aws:cloudfront::123456789012:distribution/E2E0LMG089PAAA

## Important Instructions
- You MUST provide ALL three fields in your response: "nodes", "edges", and "roots"
- All fields are required and must be included even if the lists are empty
- The "edges" field must contain relationships between the resources
- The "roots" field must contain the entry points of the system (domain names)

## Example Output
```json
{
    "nodes": [
        {
            "nodeId": 1,
            "service": "route53",
            "arn": "example.jp"
        },
        {
            "nodeId": 2,
            "service": "cloudfront",
            "arn": "arn:aws:cloudfront::123456789012:distribution/E1185M26TSAAAA"
        },
        {
            "nodeId": 3,
            "service": "wafv2",
            "arn": "arn:aws:wafv2:us-east-1:123456789012:global/webacl/example-frontend-cloudfront-waf-global/e035f4cb-30a9-4d56-bc32-684cd274c037"
        },
        {
            "nodeId": 4,
            "service": "s3",
            "arn": "arn:aws:s3:::example-frontend-ap-northeast-1"
        },
        {
            "nodeId": 5,
            "service": "elasticloadbalancing",
            "arn": "arn:aws:elasticloadbalancing:ap-northeast-1:123456789012:loadbalancer/app/example-webapp-alb/363c0528627a298c"
        },
        {
            "nodeId": 6,
            "service": "wafv2",
            "arn": "arn:aws:wafv2:ap-northeast-1:123456789012:regional/webacl/example-webapp-alb-waf-ap-northeast-1/88a01ddc-108e-46f7-9f72-6c439dd9d0cb"
        },
        {
            "nodeId": 7,
            "service": "targetgroup",
            "arn": "arn:aws:elasticloadbalancing:ap-northeast-1:123456789012:targetgroup/example-webapp-tg/46123039102d7439"
        },
        {
            "nodeId": 8,
            "service": "ecs",
            "arn": "arn:aws:ecs:ap-northeast-1:123456789012:task/example-cluster/759ed7979e9d4cd4a5d7401bf89a46f5"
        },
        {
            "nodeId": 9,
            "service": "rds",
            "arn": "arn:aws:rds:ap-northeast-1:123456789012:cluster:example-aurora3-cluster-ap-northeast-1"
        }
    ],
    "edges": [
        {
            "from": 1,
            "to": 2
        },
        {
            "from": 2,
            "to": 3
        },
        {
            "from": 2,
            "to": 4
        },
        {
            "from": 2,
            "to": 5
        },
        {
            "from": 5,
            "to": 6
        },
        {
            "from": 5,
            "to": 7
        },
        {
            "from": 7,
            "to": 8
        },
        {
            "from": 8,
            "to": 9
        }
    ],
    "roots": [
        "example.jp"
    ]
}
```"""

    llm = ChatBedrock(
        model_id="arn:aws:bedrock:us-west-2:123456789012:inference-profile/global.anthropic.claude-sonnet-4-5-20250929-v1:0",
        region_name="us-west-2",
        provider="anthropic",
        model_kwargs={"temperature": 0.0},
        max_tokens=4096,
    )

    agent = create_react_agent(
        model=llm,
        tools=[awscli_tool],
        prompt=system_prompt,
        response_format=AffectedResources,
    )

    user_message = """Analyze the provided list of AWS resources and identify how they are interconnected."""

    langfuse_handler = CallbackHandler()
    agent.invoke(
        {"messages": [{"role": "user", "content": user_message}]},
        config={"callbacks": [langfuse_handler]},
        stream_mode="values", # default
    )

if __name__ == "__main__":
    main()
~~~
