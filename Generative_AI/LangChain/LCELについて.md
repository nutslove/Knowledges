- Chainに独自の関数を挟んだり、複数のChainを並列につないで実行することもできる

### LCELを使ったRAGのChainの実装例
- `{"context": retriever, "question": RunnablePassthrough()}`は入力が`retriever`に渡されつつ、`prompt`にも渡されるイメージ  
  ```python
  from langchain_core.prompts import ChatPromptTemplate
  from langchain_openai import ChatOpenAI
  from langchain_core.output_parsers import StrOutputParser
  from langchain_core.runnables import RunnablePassthrough

  prompt = ChatPromptTemplate.from_template('''\
  以下の文脈だけを踏まえて質問に回答してください。

  文脈: """
  {context}
  """

  質問: {question}
  ''')

  model = ChatOpenAI(model_name="gpt-4o-mini", temperature=0)

  chain = (
      {"context": retriever, "question": RunnablePassthrough()}
      | prompt
      | model
      | StrOutputParser()
  )

  output = chain.invoke(query)
  print(output)
  ```
