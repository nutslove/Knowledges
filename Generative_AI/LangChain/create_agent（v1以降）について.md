## `create_agent`について

- Langchain v1 から`create_react_agent`が`create_agent`に変更された。

## `astream_events`メソッドについて

- `astream_events`メソッドは非同期ストリーミングイベントを処理するためのメソッド
- **該当コード**
  - https://github.com/langchain-ai/langchain/blob/master/libs/core/langchain_core/runnables/base.py#L1273
- `astream_events`メソッドの戻り値`StreamEvent`は辞書型で、以下のスキーマを持つ（コードの Docstring より抜粋）

  > `StreamEvent` is a dictionary with the following schema:
  >
  >      - `event`: Event names are of the format:
  >          `on_[runnable_type]_(start|stream|end)`.
  >      - `name`: The name of the `Runnable` that generated the event.
  >      - `run_id`: Randomly generated ID associated with the given execution of the
  >          `Runnable` that emitted the event. A child `Runnable` that gets invoked as
  >          part of the execution of a parent `Runnable` is assigned its own unique ID.
  >      - `parent_ids`: The IDs of the parent runnables that generated the event. The
  >          root `Runnable` will have an empty list. The order of the parent IDs is from
  >          the root to the immediate parent. Only available for v2 version of the API.
  >          The v1 version of the API will return an empty list.
  >      - `tags`: The tags of the `Runnable` that generated the event.
  >      - `metadata`: The metadata of the `Runnable` that generated the event.
  >      - `data`: The data associated with the event. The contents of this field
  >          depend on the type of event. See the table below for more details.

- `event`フィールドの値は以下のいずれかになる（以下コードの Docstring より抜粋）

  > **Note:** This reference table is for the v2 version of the schema.

| event                  | name                 | chunk                             | input                                           | output                                            |
| ---------------------- | -------------------- | --------------------------------- | ----------------------------------------------- | ------------------------------------------------- |
| `on_chat_model_start`  | `'[model name]'`     |                                   | `{"messages": [[SystemMessage, HumanMessage]]}` |                                                   |
| `on_chat_model_stream` | `'[model name]'`     | `AIMessageChunk(content="hello")` |                                                 |                                                   |
| `on_chat_model_end`    | `'[model name]'`     |                                   | `{"messages": [[SystemMessage, HumanMessage]]}` | `AIMessageChunk(content="hello world")`           |
| `on_llm_start`         | `'[model name]'`     |                                   | `{'input': 'hello'}`                            |                                                   |
| `on_llm_stream`        | `'[model name]'`     | `'Hello'`                         |                                                 |                                                   |
| `on_llm_end`           | `'[model name]'`     |                                   | `'Hello human!'`                                |                                                   |
| `on_chain_start`       | `'format_docs'`      |                                   |                                                 |                                                   |
| `on_chain_stream`      | `'format_docs'`      | `'hello world!, goodbye world!'`  |                                                 |                                                   |
| `on_chain_end`         | `'format_docs'`      |                                   | `[Document(...)]`                               | `'hello world!, goodbye world!'`                  |
| `on_tool_start`        | `'some_tool'`        |                                   | `{"x": 1, "y": "2"}`                            |                                                   |
| `on_tool_end`          | `'some_tool'`        |                                   |                                                 | `{"x": 1, "y": "2"}`                              |
| `on_retriever_start`   | `'[retriever name]'` |                                   | `{"query": "hello"}`                            |                                                   |
| `on_retriever_end`     | `'[retriever name]'` |                                   | `{"query": "hello"}`                            | `[Document(...), ..]`                             |
| `on_prompt_start`      | `'[template_name]'`  |                                   | `{"question": "hello"}`                         |                                                   |
| `on_prompt_end`        | `'[template_name]'`  |                                   | `{"question": "hello"}`                         | `ChatPromptValue(messages: [SystemMessage, ...])` |
