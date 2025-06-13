## Node名とStateのKey名は同じものを使えない
- `add_node`で追加するNode名と、Stateの中のKey名は重複が許されないっぽい。
- 例  
  ```python
  from langgraph.graph import MessagesState, StateGraph

  class State(MessagesState):
    force_respond: bool = False

  workflow = StateGraph(State)
  workflow.add_node("force_respond", force_respond)
  ```

  ```shell
  ValueError: 'force_respond' is already being used as a state key
  Traceback
  ```

## React Agentで、llmがtool callと判断したのに無理やり回答させようとするとエラーとなる
- `bind_tools`でToolを紐づけたReact Agentで、LLMが次のアクションとして`tool_calls`を選んだのに、強制的にToolではなく、回答をさせようとすると以下のようなエラーが出る  
  ```shell
  ValidationException('An error occurred (ValidationException) when calling the InvokeModel operation: messages.36: Did not find 1 `tool_result` block(s) at the beginning of this message. Messages following `tool_use` blocks must begin with a matching number of `tool_result` blocks.')Traceback (most recent call last):
  ```
  - その時は以下のようにMessageステートから最後のメッセージを除外すればうまくいったりする  
    ```python
    history_messages = state["messages"][:-1]
    ```

## StateのMessageにToolによるメッセージ（e.g. `tool_use`または`tool_result`）が含まれているのに、`bind_tools`してないLLMで`invoke`すると以下のようなエラーが出る
```
Error raised by bedrock service: An error occurred (ValidationException) when calling the InvokeModel operation: Requests which include `tool_use` or `tool_result` blocks must define tools.
```

## Node並列実行時、各NodeでState内の同じ項目を更新すると以下のエラーになる
- 複数のNodeでState内の`system_name`項目を更新しようとしたときのエラー  
  ```shell
  At key 'system_name': Can receive only one value per step. Use an Annotated key to handle multiple values.
  ```
- 各Nodeで同じ項目を更新しないように修正するか、add Reducerを使って上書きではなく、追記する形にする
  - https://langchain-ai.github.io/langgraph/troubleshooting/errors/INVALID_CONCURRENT_GRAPH_UPDATE/  
  - 例  
    ```python
    import operator
    from typing import Annotated

    # add Reducerでappendすることによって回避
    class State(TypedDict):
        # The operator.add reducer fn makes this append-only
        some_key: Annotated[list, operator.add]

    # エラーになる
    class State(TypedDict):
        some_key: str

    def node(state: State):
        return {"some_key": "some_string_value"}

    def other_node(state: State):
        return {"some_key": "some_string_value"}

    builder = StateGraph(State)
    builder.add_node(node)
    builder.add_node(other_node)
    builder.add_edge(START, "node")
    builder.add_edge(START, "other_node")
    graph = builder.compile()
    ```

## Nodeの中でstateを使わない場合でもNode関数の引数にstateは必要
- `get_aws_health_dashboard_info`Node関数に引数を定義してない場合、以下のようなエラーが出る  
  ```shell
  get_aws_health_dashboard_info() takes 0 positional arguments but 1 was given
  ```  
  - Node関数内で使わないとしても以下のようにstateを設定してあげる  
    ```python
    def get_aws_health_dashboard_info(state: State) -> dict:
    ```