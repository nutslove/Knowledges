## Bedrockでは最後にAIMessageが入っているとエラーになる
- エラー内容  
  ```shell
  botocore.errorfactory.ValidationException: An error occurred (ValidationException) when calling the InvokeModel operation: Your API request included an `assistant` message in the final position, which would pre-fill the `assistant` response. When using tools, pre-filling the `assistant` response is not supported.
  ```
- https://github.com/langchain-ai/langchain-aws/issues/124  
  ![](../../AWS/image/Bedrock_AIMessage_in_the_last_error.jpg)
  - 上記には以下のように`beta_use_converse_api=False`にすればいいみたいなことが書いてあるけど、ダメだった。  
    ```python
    llm = ChatBedrock(
        model_id="anthropic.claude-3-5-sonnet-20240620-v1:0",
        model_kwargs={"temperature": 0.1},
        beta_use_converse_api=False,
    )
    ```
- **なのでMulti AgentsやToolなどで、最後にAIMessageが入っている場合は、そのままLLMに投げるのではなくて、System PromptとHumanMessageだけ抜き取ってLLMに投げるようにする必要がある**

## AIが生成する回答（Output）のmax tokens
- デフォルトでは*1024*になっているように見える
- max output tokenを超えるとAIが回答の生成をやめる。
- 以下のようなメッセージが出る  
  ```json
  {
  	usage: {
  		prompt_tokens: 45595
  		completion_tokens: 1024
  		total_tokens: 46619
  	}
  	stop_reason: "max_tokens"
  	model_id: "arn:aws:bedrock:us-west-2:123456789:inference-profile/us.anthropic.claude-3-7-sonnet-20250219-v1:0"
  }
  ```
- LangChainの`ChatBedrock`で、`max_tokens`で変更できる
  - https://python.langchain.com/api_reference/aws/chat_models/langchain_aws.chat_models.bedrock.ChatBedrock.html#langchain_aws.chat_models.bedrock.ChatBedrock.max_tokens  
  ```python
  llm = ChatBedrock(
      model_id=model_id,
      region_name=llm_region,
      provider="anthropic",
      model_kwargs={"temperature": 0.1},
      beta_use_converse_api=False,
      max_tokens=4096 # default: 1024
  )
  ```

## Tool use
- OpenAIのFunction Callingに該当するもの
- https://docs.aws.amazon.com/bedrock/latest/userguide/tool-use.html

## Bedrock Agents
- https://speakerdeck.com/minorun365/sutatoatupuderen-gazu-rinai-sorenaraaiezientowogu-ou
- Lambda関数の実行もできる
- Knowledge Baseとの連動もできる
- 会話履歴も踏まえて回答してくれる
- SlackにてAWS ChatBotからBedrock Agentを使って直接Bedrockを呼び出すことができる
  - https://qiita.com/moritalous/items/b63d976c2c40af1c39e5