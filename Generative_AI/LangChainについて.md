## *LLMs* vs *Chat models*
- https://python.langchain.com/docs/modules/model_io/#llms-vs-chat-models  
  > LLMs and chat models are subtly but importantly different. LLMs in LangChain refer to pure text completion models. The APIs they wrap take a string prompt as input and output a string completion. OpenAI's GPT-3 is implemented as an LLM. Chat models are often backed by LLMs but tuned specifically for having conversations. And, crucially, their provider APIs use a different interface than pure text completion models. Instead of a single string, they take a list of chat messages as input. Usually these messages are labeled with the speaker (usually one of "System", "AI", and "Human"). And they return an AI chat message as output. GPT-4 and Anthropic's Claude-2 are both implemented as chat models.
- **LLMs**
  - 1回の回答で終わるもの
  - inputとしてstringを受け付ける
- **Chat models**
  - 1回の回答で終わりではなく会話のためにtuningされたもの
  - inputとしてchat messageのリストを受け付ける

## Memory
- デフォルトではChat-GPTなどのLLMはチャットの履歴を保持しない
- LLMに会話履歴を踏まえて回答してほしい場合は、リクエストに過去のやり取り(会話履歴)を含める必要がある
- チャットの履歴を保持させるためのLangChainの機能がMemory
- **role**を持ち、ユーザからの入力なのかLLMからの回答なのかを記憶させる
  - OpenAIの場合
    - LLMへの指示内容のroleは **`system`**
    - ユーザからの入力のroleは **`user`**
    - LLMからの出力のroleは **`assistant`**
    - e.g.
      ~~~json
      {
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "Hi! My name is Lee."},
          {"role": "assistant", "content": "Hi Lee! How can I help you today?"},
          {"role": "user", "content": "Do you remember my name?"}
        ]
      }
      ~~~

### AWS DynamoDBをMemoryとして利用
- 参考URL
  - https://python.langchain.com/docs/integrations/memory/aws_dynamodb
  - https://tech.nri-net.com/entry/lambda_dynamodb_create_chatbot
  - https://qiita.com/HayaP/items/389ff57e8a204ef403c1
- `DynamoDBChatMessageHistory`クラスを使ってDynamoDBで会話履歴を管理する
  - `DynamoDBChatMessageHistory`クラスのパラメータ一覧  
    > - **table_name** : name of the DynamoDB table
    > - **session_id** : arbitrary key that is used to store the messages
            of a single chat session.
    > - **endpoint_url**: URL of the AWS endpoint to connect to. This argument
            is optional and useful for test purposes, like using Localstack.
            If you plan to use AWS cloud service, you normally don't have to
            worry about setting the endpoint_url.
    > - **primary_key_name**: name of the primary key of the DynamoDB table. This argument
            is optional, defaulting to "SessionId".
    > - **key**: an optional dictionary with a custom primary and secondary key.
            This argument is optional, but useful when using composite dynamodb keys, or
            isolating records based off of application details such as a user id.
            This may also contain global and local secondary index keys.
    > - **kms_key_id**: an optional AWS KMS Key ID, AWS KMS Key ARN, or AWS KMS Alias for client-side encryption
    - https://github.com/langchain-ai/langchain/blob/master/libs/community/langchain_community/chat_message_histories/dynamodb.py
- Memoryでチャット履歴を参照し使うためには`ConversationBufferMemory`クラスで`memory_key`と`chat_memory`パラメータを指定する必要がある  
  - `ChatPromptTemplate.from_messages`の`MessagesPlaceholder`の`variable_name`の値を`memory_key`の値と合わせる必要がある
  - `ConversationBufferMemory`クラスの`chat_memory`パラメータに会話履歴を渡す
    ~~~python
    memory = ConversationBufferMemory(memory_key = "chat_history", chat_memory = history, return_messages = True)
    prompt = ChatPromptTemplate.from_messages([
        MessagesPlaceholder(variable_name = "chat_history"), ## variable_nameはmemory_keyと同じにする必要がある
        HumanMessagePromptTemplate.from_template("""{input}""")
    ])
    ~~~
- DynamoDBへのユーザからの入力の追加は`add_user_message(user_input)`、AIからの回答の追加は`add_ai_message(ai_output)`で行う  
  ~~~python
  history = DynamoDBChatMessageHistory(table_name=tablename, session_id=sess_id)
  print(f'■■■■history.messages■■■■\n{history.messages}')
  llm_chain = ConversationChain(llm = llm, prompt = prompt, memory = memory)
  output_text = llm_chain.run(user_input)

  history.add_user_message(user_input)
  history.add_ai_message(output_text)
  ~~~
  - `DynamoDBChatMessageHistory`クラスのinstanceの`messages`には、ユーザInputは`HumanMessage(content='<入力内容>')`、AIからと回答は`AIMessage(content='<AIからの回答内容>'`という形で格納される
- サンプルコード  
  ~~~python
  from boto3.session import Session
  from botocore.exceptions import ClientError
  import boto3
  from langchain.chat_models import BedrockChat
  from langchain.callbacks.streaming_stdout import StreamingStdOutCallbackHandler
  from langchain.chains import ConversationChain
  from langchain.memory import ConversationBufferMemory
  from langchain.memory.chat_message_histories import DynamoDBChatMessageHistory
  from langchain.prompts.chat import (
      ChatPromptTemplate,
      HumanMessagePromptTemplate,
      MessagesPlaceholder,
  )
  import os
  #dynamodb = boto3.resource("dynamodb", region_name="ap-northeast-1")
  dynamodb = boto3.resource("dynamodb")
  tablename = os.getenv("DYNAMODB_TABLENAME") ## 環境変数からDynamoDBテーブル名を取得

  # テーブルが存在しなければテーブルを作成
  try:
      response = boto3.client("dynamodb").describe_table(TableName=tablename)
  except ClientError as e:
      if e.response['Error']['Code'] == 'ResourceNotFoundException':
          print(f"テーブル '{table_name}' は存在しないので作成します。")
          # Create the DynamoDB table.
          table = dynamodb.create_table(
              TableName=tablename,
              KeySchema = [{"AttributeName": "SessionId", "KeyType": "HASH"}],
              AttributeDefinitions = [{"AttributeName": "SessionId", "AttributeType": "S"}],
              BillingMode = "PAY_PER_REQUEST",
          )

          # Wait until the table exists.
          table.meta.client.get_waiter("table_exists").wait(TableName=tablename)

          # Print out some data about the table.
          print(table.item_count)
      else:
          print(f"エラーが発生しました: {e}")
  else:
      # テーブルオブジェクトの取得
      table = dynamodb.Table(tablename)

      # テーブルをスキャン
      response = table.scan()
      #print(f'■■■■response■■■■\n{response}')

      # スキャンした結果の表示
      #for item in response['Items']:
      #    print(item)

  # クエリ内容を設定
  user_input = input("質問：")
  sess_id = input("Session ID: ") ## session idで会話履歴を管理

  #history = DynamoDBChatMessageHistory(table_name=tablename, session_id=sess_id, endpoint_url="https://dynamodb.ap-northeast-1.amazonaws.com")
  history = DynamoDBChatMessageHistory(table_name=tablename, session_id=sess_id)

  #print(f'■■■■history.messages■■■■\n{history.messages}')
  #print(f'■■■■history■■■■\n{history}')

  llm = BedrockChat(
      model_id = "anthropic.claude-v2:1",
      model_kwargs = {"temperature":0.1, "top_k": 10, "max_tokens_to_sample": 500},
  )

  memory = ConversationBufferMemory(memory_key = "chat_history", chat_memory = history, return_messages = True)
  prompt = ChatPromptTemplate.from_messages([
      MessagesPlaceholder(variable_name = "chat_history"), ## variable_nameはmemory_keyと同じにする必要がある
      HumanMessagePromptTemplate.from_template("""{input}""")
  ])

  llm_chain = ConversationChain(llm = llm, prompt = prompt, memory = memory)
  output_text = llm_chain.run(user_input)

  history.add_user_message(user_input)
  history.add_ai_message(output_text)

  print(output_text)
  ~~~

#### ■ **AWS Credentials(Config)で指定しているAWSリージョンにDynamoDBテーブルが存在する必要がある**
- `DynamoDBChatMessageHistory`クラスの`endpoint_url`パラメータで一応リージョンのエンドポイントを指定でいるけど、AWS credentials(config)に指定されているリージョンと異なる場合、以下のエラーが出る
  - `An error occurred (InvalidSignatureException) when calling the GetItem operation: Credential should be scoped to a valid region.`
- LangChainの該当コード
  - https://github.com/langchain-ai/langchain/blob/master/libs/community/langchain_community/chat_message_histories/dynamodb.py
- DynamoDBエンドポイント
  - https://docs.aws.amazon.com/ja_jp/general/latest/gr/ddb.html
- 参考URL
  - https://qiita.com/tottu22/items/260f75de737c21664ec7

## その他
- レスポンスのパラメータで指定する`top_k`は言語モデルがテキスト(回答文)を生成する際に、各ステップで考慮するトークン(文字列)の候補数を指定する。具体的には、モデルが次のトークンを選ぶ際に、確率が最も高い上位 k 個のトークンの中から選択を行うようになる。なので、`temperature`を低く設定している場合は、`top_k`を大きく設定してもあまり意味がない