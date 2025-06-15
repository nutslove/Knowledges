- Chainに独自の関数を挟んだり、複数のChainを並列につないで実行することもできる

## Runnable / RunnableSequence
- 参考URL
  - https://python.langchain.com/docs/concepts/runnables/
  - https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.base.Runnable.html
- Prompt Template、Chat Model、Output Parser、Retrieverなどのモジュールは「**Runnable**」という抽象基底クラスを継承していて、**RunnableクラスはRunnableの実行方法として`invoke`、`stream`、`batch`メソッドを持つ**
- **Runnableを`|`でつなぐと**「**RunnableSequence**」となる
  - RunnableSequenceもRunnableの一種
- **RunnableSequenceを`invoke`すると連結したRunnableが順に`invoke`される**
  - RunnableSequenceを`stream`で呼び出すと、連結したRunnableが順に`stream`で実行される
  - RunnableSequenceを`batch`で呼び出すと、連結したRunnableが順に`batch`で実行される
- このようにRunnableを`|`でつないで新たなRunnable（RunnableSequence）を作り、それをinvokeしたときに、内部のRunnableが順に実行(invoke)される仕組みを **LCEL（LangChain Expression Language）** と呼ぶ
- **Runnableを`|`で連結する時は、出力の型と入力の型の整合性に注意する必要がある**
  - 例えば、`RunnableSequence`の最初のRunnableの出力が`str`型で、次のRunnableの入力が`int`型の場合はエラーになる
- 簡単な例  
  ```python
  from langchain_core.output_parsers import StrOutputParser
  from langchain_core.prompts import ChatPromptTemplate
  from langchain_aws import ChatBedrock


  prompt = ChatPromptTemplate.from_messages(
    [
      ("system", "ユーザが入力した料理名を元に、料理のレシピを考えてください。"),
      ("user", "{dish}"),
    ]
  )

  model = ChatBedrock(
    model_id="anthropic.claude-3-5-sonnet-20240620-v1:0",
    region_name="ap-northeast-1",
  )

  output_parser = StrOutputParser()

  chain = prompt | model | output_parser
  result = chain.invoke({"dish": "ハンバーグ"})
  ## 上記は以下と同じ
  # prompt_value = prompt.invoke({"dish": "ハンバーグ"})
  # model_value = model.invoke(prompt_value)
  # result = output_parser.invoke(model_value)
  print(result)
  ```

### `|`の仕組み
- pythonでは、演算子の動作を自分で定義できる
- `Runnable`クラスは`__or__`メソッドと`__ror__`メソッドをオーバーライドしていて、`|`演算子が呼ばれたときに、2つのRunnableを連結した新しいRunnableSequenceを返すようになっている
  - `__or__`メソッドは左側のRunnableから右側のRunnableへと結合
  - `__ror__`メソッドは右側のRunnableから左側のRunnableへと結合
  - `A | B`は、まず`A.__or__(B)`を試し、`A.__or__(B)` が `NotImplemented` を返すか、`__or__` メソッドが存在しない場合、次に `B.__ror__(A)` を試す。それも PNotImplemented` を返すか存在しない場合、`TypeError` が発生
- `prompt | model`は`prompt.__or__(model)`と同じ意味
- `|`つまり、`__or__`メソッドは、戻り値として`RunnableSequence`を返す

---

## `RunnablePassthrough`について
- `RunnablePassthrough`は、入力をそのまま出力として返すRunnable
- 以下の例で、`{"context": retriever, "question": RunnablePassthrough()}`の部分は、`retriever`からの出力（検索結果）を`context`に渡しつつ、入力された質問をそのまま`question`に渡す役割を果たす
- RunnablePassthroughを使ったRAGのChainの実装例    
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
