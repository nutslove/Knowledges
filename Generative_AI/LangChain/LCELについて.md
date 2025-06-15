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
- 該当コードの部分  
  - https://github.com/langchain-ai/langchain/blob/master/libs/core/langchain_core/runnables/base.py#L564  
  ```python
  class Runnable(Generic[Input, Output], ABC):
      """A unit of work that can be invoked, batched, streamed, transformed and composed.

      Key Methods
      ===========

      - **invoke/ainvoke**: Transforms a single input into an output.
      - **batch/abatch**: Efficiently transforms multiple inputs into outputs.
      - **stream/astream**: Streams output from a single input as it's produced.
      - **astream_log**: Streams output and selected intermediate results from an input.

      Built-in optimizations:

      - **Batch**: By default, batch runs invoke() in parallel using a thread pool executor.
        Override to optimize batching.

      - **Async**: Methods with "a" suffix are asynchronous. By default, they execute
        the sync counterpart using asyncio's thread pool.
        Override for native async.

      All methods accept an optional config argument, which can be used to configure
      execution, add tags and metadata for tracing and debugging etc.

      Runnables expose schematic information about their input, output and config via
      the input_schema property, the output_schema property and config_schema method.

      LCEL and Composition
      ====================

      The LangChain Expression Language (LCEL) is a declarative way to compose Runnables
      into chains. Any chain constructed this way will automatically have sync, async,
      batch, and streaming support.

      The main composition primitives are RunnableSequence and RunnableParallel.

      **RunnableSequence** invokes a series of runnables sequentially, with
      one Runnable's output serving as the next's input. Construct using
      the `|` operator or by passing a list of runnables to RunnableSequence.

      **RunnableParallel** invokes runnables concurrently, providing the same input
      to each. Construct it using a dict literal within a sequence or by passing a
      dict to RunnableParallel.


      For example,

      .. code-block:: python

          from langchain_core.runnables import RunnableLambda

          # A RunnableSequence constructed using the `|` operator
          sequence = RunnableLambda(lambda x: x + 1) | RunnableLambda(lambda x: x * 2)
          sequence.invoke(1) # 4
          sequence.batch([1, 2, 3]) # [4, 6, 8]


          # A sequence that contains a RunnableParallel constructed using a dict literal
          sequence = RunnableLambda(lambda x: x + 1) | {
              'mul_2': RunnableLambda(lambda x: x * 2),
              'mul_5': RunnableLambda(lambda x: x * 5)
          }
          sequence.invoke(1) # {'mul_2': 4, 'mul_5': 10}

      Standard Methods
      ================

      All Runnables expose additional methods that can be used to modify their behavior
      (e.g., add a retry policy, add lifecycle listeners, make them configurable, etc.).

      These methods will work on any Runnable, including Runnable chains constructed
      by composing other Runnables. See the individual methods for details.

      For example,

      .. code-block:: python

          from langchain_core.runnables import RunnableLambda

          import random

          def add_one(x: int) -> int:
              return x + 1


          def buggy_double(y: int) -> int:
              \"\"\"Buggy code that will fail 70% of the time\"\"\"
              if random.random() > 0.3:
                  print('This code failed, and will probably be retried!')  # noqa: T201
                  raise ValueError('Triggered buggy code')
              return y * 2

          sequence = (
              RunnableLambda(add_one) |
              RunnableLambda(buggy_double).with_retry( # Retry on failure
                  stop_after_attempt=10,
                  wait_exponential_jitter=False
              )
          )

          print(sequence.input_schema.model_json_schema()) # Show inferred input schema
          print(sequence.output_schema.model_json_schema()) # Show inferred output schema
          print(sequence.invoke(2)) # invoke the sequence (note the retry above!!)

      Debugging and tracing
      =====================

      As the chains get longer, it can be useful to be able to see intermediate results
      to debug and trace the chain.

      You can set the global debug flag to True to enable debug output for all chains:

          .. code-block:: python

              from langchain_core.globals import set_debug
              set_debug(True)

      Alternatively, you can pass existing or custom callbacks to any given chain:

          .. code-block:: python

              from langchain_core.tracers import ConsoleCallbackHandler

              chain.invoke(
                  ...,
                  config={'callbacks': [ConsoleCallbackHandler()]}
              )

      For a UI (and much more) checkout LangSmith: https://docs.smith.langchain.com/
      """

  ## ・・中略・・

    def __or__(
        self,
        other: Union[
            Runnable[Any, Other],
            Callable[[Iterator[Any]], Iterator[Other]],
            Callable[[AsyncIterator[Any]], AsyncIterator[Other]],
            Callable[[Any], Other],
            Mapping[str, Union[Runnable[Any, Other], Callable[[Any], Other], Any]],
        ],
    ) -> RunnableSerializable[Input, Other]:
        """Compose this Runnable with another object to create a RunnableSequence."""
        return RunnableSequence(self, coerce_to_runnable(other))

    def __ror__(
        self,
        other: Union[
            Runnable[Other, Any],
            Callable[[Iterator[Other]], Iterator[Any]],
            Callable[[AsyncIterator[Other]], AsyncIterator[Any]],
            Callable[[Other], Any],
            Mapping[str, Union[Runnable[Other, Any], Callable[[Other], Any], Any]],
        ],
    ) -> RunnableSerializable[Other, Output]:
        """Compose this Runnable with another object to create a RunnableSequence."""
        return RunnableSequence(coerce_to_runnable(other), self)
  ```

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
