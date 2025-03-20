- Chainに独自の関数を挟んだり、複数のChainを並列につないで実行することもできる

## Runnable / RunnableSequence
- 参考URL
  - https://python.langchain.com/docs/concepts/runnables/
  - https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.base.Runnable.html
- Prompt Template、Chat Model、Output Parser、Retrieverなどはすべて「**Runnable**」という抽象基底クラスを継承していて、**RunnableクラスはRunnableの実行方法として`invoke`、`stream`、`batch`メソッドを持つ**
- **Runnableを`|`でつなぐと**「**RunnableSequence**」となる
  - RunnableSequenceもRunnableの一種
- **RunnableSequenceをinvokeすると連結したRunnableが順にinvokeされる**

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

### `RunnablePassthrough`について