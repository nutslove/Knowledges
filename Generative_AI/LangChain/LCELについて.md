- Chainに独自の関数を挟んだり、複数のChainを並列につないで実行することもできる

## Runnable / RunnableSequence
- 参考URL
  - https://python.langchain.com/docs/concepts/runnables/
  - https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.base.Runnable.html
- Prompt Template、Chat Model、Output Parser、Retrieverなどのモジュールは「**Runnable**」という抽象基底クラスを継承していて、**RunnableクラスはRunnableの実行方法として`invoke`、`stream`、`batch`メソッドを持つ**
- **Runnableを`|`でつなぐと**「**RunnableSequence**」となる
  - RunnableSequenceもRunnableの一種
- **RunnableSequenceをinvokeすると連結したRunnableが順にinvokeされる**
- このようにRunnableを`|`でつないで新たなRunnable（RunnableSequence）を作り、それをinvokeしたときに、内部のRunnableが順に実行(invoke)される仕組みを **LCEL（LangChain Expression Language）** と呼ぶ
- Runnableを`|`で連結する時は、出力の型と入力の型の整合性に注意する必要がある
  - 例えば、`RunnableSequence`の最初のRunnableの出力が`str`型で、次のRunnableの入力が`int`型の場合はエラーになる

### LCELを使ったRAGのChainの実装例
- `{"context": retriever, "question": RunnablePassthrough()}`は入力が`retriever`に渡されつつ、`prompt`にも渡されるイメージ  
  ```python
  from langchain_core.prompts import ChatPromptTemplate
  from langchain_openai import ChatOpenAI
  from langchain_core.output_parsers import StrOutputParser
  from langchain_core.runnables import RunnablePassthrough
  from langchain_community.retrievers import TavilySearchAPIRetriever

  prompt = ChatPromptTemplate.from_template('''\
  以下の文脈だけを踏まえて質問に回答してください。

  文脈: """
  {context}
  """

  質問: {question}
  ''')

  model = ChatOpenAI(model_name="gpt-4o-mini", temperature=0)
  retriever = TavilySearchAPIRetriever(k=3) # kは検索する件数

  chain = (
      {"context": retriever, "question": RunnablePassthrough()}
      | prompt
      | model
      | StrOutputParser()
  )

  output = chain.invoke("東京の今日の天気は？")
  print(output)
  ```

### `RunnablePassthrough`について
- `RunnablePassthrough`は、入力をそのまま出力として返すRunnable
- 上記の例で、`{"context": retriever, "question": RunnablePassthrough()}`の部分は、`retriever`からの出力（検索結果）を`context`に渡しつつ、入力された質問をそのまま`question`に渡す役割を果たす