# Streamlitと組み合わせて使う例
### LangChainの`invoke`の代わりに`stream`メソッドを使用
- https://python.langchain.com/docs/concepts/streaming/
- https://python.langchain.com/docs/how_to/streaming/

### pythonの`yield`とStreamlitの`st.write_stream`を使う
- https://docs.streamlit.io/develop/api-reference/write-magic/st.write_stream

## `invoke`の例
```python
import streamlit as st
import json
import urllib
import os
import datetime
import time
import string
import random
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.runnables.history import RunnableWithMessageHistory
from langchain_community.chat_message_histories import StreamlitChatMessageHistory

def init_state(key: str, value: any = None) -> None:
    if key not in st.session_state:
        st.session_state[key] = value

def general_llm_call(user_input: str) -> str:
    prompt = ChatPromptTemplate.from_messages(
        [
            ("system", "You are an AI chatbot having a conversation with a human. Answer must be in japanese."),
            MessagesPlaceholder(variable_name="history"),
            ("human", "{question}"),
        ]
    )
    llm = ChatOpenAI(model="gpt-4o",temperature=0.1,max_tokens=None)
    msgs = StreamlitChatMessageHistory(key="chat_messages")
    chain = prompt | llm
    chain_with_history = RunnableWithMessageHistory(
        chain,
        get_session_history = lambda message_history: msgs,
        input_messages_key = "question",
        history_messages_key = "history",
    )
    config = {"configurable": {"session_id": "any"}}
    res = chain_with_history.invoke({"question": user_input}, config)
    return res.content

def reset_chat():
    del st.session_state["messages"]
    del st.session_state["chat_messages"]
    del st.session_state["message_id"]

def main():
    init_state("messages", [{"role": "assistant", "content": "初めまして、Mship3のChatBotです。Mship3についてお答えします。"}])
    init_state("message_id", str(time.time_ns())+''.join(random.choice(string.ascii_lowercase) for _ in range(5)))
    for message in st.session_state.messages:
        with st.chat_message(message["role"]):
            st.markdown(message["content"])

    if user_input := st.chat_input("質問を入力してください"):
        st.session_state.messages.append({"role": "user", "content": user_input})
        with st.chat_message("user"):
            st.markdown(user_input)
        with st.spinner("Thinking..."):
            res = general_llm_call(user_input)
        with st.chat_message("assistant"):
            st.markdown(res)
            st.session_state.messages.append({"role": "assistant", "content": res})

        if st.button('会話リセット', type="primary", icon=":material/refresh:", on_click=reset_chat):
            st.rerun()

if __name__ == '__main__':
    st.set_page_config(
        page_title="Mship3 ChatBot",
        page_icon=":books:"
    )
    st.title("Mship3 ChatBot :books:")
    main()
```

## `stream`の例
- content（回答内容）だけ受け取るために`StrOutputParser`を使用
- LLMを叩く関数を`st.write_stream`で囲む
- LangChainの`stream`メソッドはChunkを少しずつ返すので、`for`文で回しながら`yield`でChunkを処理する
- LLMを叩く関数が明示的に`return`しなくても`yield`で処理したChunkの固まりがLLMを叩く関数の戻り値として入る（下記の例の`res`変数）

```python
import streamlit as st
import json
import urllib
import os
import datetime
import time
import string
import random
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.runnables.history import RunnableWithMessageHistory
from langchain_core.output_parsers import StrOutputParser
from langchain_community.chat_message_histories import StreamlitChatMessageHistory

def init_state(key: str, value: any = None) -> None:
    if key not in st.session_state:
        st.session_state[key] = value

def general_llm_call(user_input: str) -> None:
    prompt = ChatPromptTemplate.from_messages(
        [
            ("system", "You are an AI chatbot having a conversation with a human. Answer must be in japanese."),
            MessagesPlaceholder(variable_name="history"),
            ("human", "{question}"),
        ]
    )
    llm = ChatOpenAI(model="gpt-4o",temperature=0.1,max_tokens=None)
    msgs = StreamlitChatMessageHistory(key="chat_messages")
    parser = StrOutputParser()
    chain = prompt | llm | parser
    chain_with_history = RunnableWithMessageHistory(
        chain,
        get_session_history = lambda message_history: msgs,
        input_messages_key = "question",
        history_messages_key = "history",
    )
    config = {"configurable": {"session_id": "any"}}
    for chunk in chain_with_history.stream({"question": user_input}, config):
        yield chunk

def reset_chat():
    del st.session_state["messages"]
    del st.session_state["chat_messages"]
    del st.session_state["message_id"]

def main():
    init_state("messages", [{"role": "assistant", "content": "初めまして、Mship3のChatBotです。Mship3についてお答えします。"}])
    init_state("message_id", str(time.time_ns())+''.join(random.choice(string.ascii_lowercase) for _ in range(5)))

    for message in st.session_state.messages:
        with st.chat_message(message["role"]):
            st.markdown(message["content"])

    if user_input := st.chat_input("質問を入力してください"):
        st.session_state.messages.append({"role": "user", "content": user_input})
        with st.chat_message("user"):
            st.markdown(user_input)
        with st.chat_message("assistant"):
            res = st.write_stream(general_llm_call(user_input))

        st.session_state.messages.append({"role": "assistant", "content": res})

        if st.button('会話リセット', type="primary", icon=":material/refresh:", on_click=reset_chat):
            st.rerun()

if __name__ == '__main__':
    st.set_page_config(
        page_title="Mship3 ChatBot",
        page_icon=":books:"
    )
    st.title("Mship3 ChatBot :books:")
    main()
```
