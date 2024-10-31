- 参考URL
  - https://python.langchain.com/docs/integrations/memory/streamlit_chat_message_history/
  - https://github.com/langchain-ai/streamlit-agent/blob/main/streamlit_agent/basic_memory.py
  - https://qiita.com/FukuharaYohei/items/a11628c1ae081c5b01ac

- `StreamlitChatMessageHistory`を使って、streamlitの**session state**に会話履歴を保持し、利用する
  - 会話履歴を保持する *session state* のKey名はdefaultでは`langchain_messages`だけど、`StreamlitChatMessageHistory(key="chat_messages")`のように`StreamlitChatMessageHistory`クラス初期化時に変更することもできる
  - `StreamlitChatMessageHistory`のKey名で**session state**が保存される
  - `RunnableWithMessageHistory`クラスの`get_session_history`変数に会話履歴（`HumanMessage`、`AIMessage`）が含まれた状態で`invoke`メソッドで実行する

#### サンプルコード（一部抜粋）
- `chain_with_history.invoke`で指定する`config`（`session_id`）はStreamlitの *session state* を使う場合は不要だけど、メソッド的に必要なため、ダミーを入れている気がする。。
  - "any"じゃなくて何を入れてもOK
- `MessagesPlaceholder`の`variable_name`名と、`RunnableWithMessageHistory`の`history_messages_key`名を一致させる必要がある
- 以下3つを一致させる必要がある
  1. `ChatPromptTemplate`の`human`の部分の変数名
  2. `RunnableWithMessageHistory`の`input_messages_key`の値
  3. `invoke`メソッドのユーザ入力値を入れるKey名

```python
import streamlit as st
import os
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.runnables.history import RunnableWithMessageHistory
from langchain_community.chat_message_histories import StreamlitChatMessageHistory

os.environ["OPENAI_ENDPOINT"] = "https://api.openai.com/v1/chat/completions"
os.environ["OPENAI_API_KEY"] = "xxxxx"

def general_llm_call(user_input: str) -> str:
    prompt = ChatPromptTemplate.from_messages(
        [
            ("system", "You are an AI chatbot having a conversation with a human. Answer must be in japanese."),
            MessagesPlaceholder(variable_name="history"),
            ("human", "{question}"),
        ]
    )
    msgs = StreamlitChatMessageHistory(key="chat_messages")
    chain = prompt | ChatOpenAI(model="gpt-4o",temperature=0.1,max_tokens=None)
    chain_with_history = RunnableWithMessageHistory(
        chain,
        get_session_history = lambda message_history: msgs,
        input_messages_key = "question",
        history_messages_key = "history",
    )
    config = {"configurable": {"session_id": "any"}} ## ないとエラーになる

    # Note: new messages are saved to history automatically by Langchain during run
    res = chain_with_history.invoke({"question": user_input}, config)

    ## 会話履歴が出力される
    print(f"history: {st.session_state}['chat_messages']")

    return res.content
```