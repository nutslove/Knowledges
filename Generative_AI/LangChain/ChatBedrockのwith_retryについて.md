## `with_retry`について
- LangChainの`ChatBedrock` Classに`with_retry`メソッドがあり、Retry設定を追加できる
  - https://reference.langchain.com/python/integrations/langchain_aws/

## 使い方
- Runnable Interfaceを持つクラスであれば`with_retry`メソッドが使える
- 例えば、`ChatBedrock`や`ChatVertexAI`、`ChatOpenAI`などで使える

```python
from langchain_google_vertexai import ChatVertexAI
from langchain_aws import ChatBedrock

# ChatVertexAI の例
llm_vertex = ChatVertexAI(model="gemini-2.5-flash")
llm_vertex_with_retry = llm_vertex.with_retry(
    stop_after_attempt=5,  # リトライ回数
    wait_exponential_jitter=True,  # 指数バックオフ + ジッター
    exponential_jitter_params={
        "initial": 1,      # 初期待機時間（秒）
        "max": 60,         # 最大待機時間（秒）
        "exp_base": 2,     # 指数の底
        "jitter": 1,       # ジッターの最大値
    },
    retry_if_exception_type=(Exception,),  # リトライする例外タイプ
)

# ChatBedrock の例
llm_bedrock = ChatBedrock(model_id="anthropic.claude-3-sonnet-20240229-v1:0")
llm_bedrock_with_retry = llm_bedrock.with_retry(
    stop_after_attempt=3,
    exponential_jitter_params={
        "initial": 2,
        "max": 30,
    },
)

# 使用例
response = llm_bedrock_with_retry.invoke("Hello")
```

> [!CAUTION]  
> - `with_retry`は`RunnableRetry`オブジェクトを返すため、`with_retry`後にChatModel固有のメソッドである`bind_tools`や`with_structured_output`は使えない。
> - なので組み合わせて使う場合は、先に`bind_tools`や`with_structured_output`を呼び出し、その後に`with_retry`を呼び出す必要がある。  
> ```python
> from botocore.exceptions import ClientError
> from google.api_core.exceptions import (
>    GoogleAPIError,       # 基底クラス（すべてキャッチ）
>    ResourceExhausted,    # 429 - スロットリング/クォータ超過
>    ServiceUnavailable,   # 503 - サービス一時停止
>    DeadlineExceeded,     # 504 - タイムアウト
>    InternalServerError,  # 500 - 内部エラー
> )
>
> def runnable_with_retry(runnable):
>    return runnable.with_retry(
>        stop_after_attempt=5,
>        retry_if_exception_type=(
>            ClientError,          # AWS Bedrock（AWSのClientErrorはGCPのように細かく分類されていないため、すべてClientErrorでキャッチする）
>            ResourceExhausted,    # GCP スロットリング
>            ServiceUnavailable,   # GCP 503
>            DeadlineExceeded,     # GCP タイムアウト
>            InternalServerError,  # GCP 500
>        ),
>        wait_exponential_jitter=True,
>        exponential_jitter_params={
>            "initial": 4,
>            "max": 30,
>            "exp_base": 2, # 指数関数の基数
>            "jitter": 5 # 0〜5秒のランダムな遅延が追加される
>        }
>    )
>
> if use_vertex_ai:
>    llm = ChatVertexAI(
>      model=vertex_model,
>      include_thoughts=True, # 思考プロセスを有効化（https://ai.google.dev/gemini-api/docs/thinking?hl=ja、https://reference.langchain.com/python/integrations/langchain_google_vertexai/#langchain_google_vertexai.ChatVertexAI）
>      temperature=0,
>      max_tokens=15000,
>      # max_retries=5, # LangChainのwith_retryを使うのでここでは指定しない
>      stop=None,
>      project="test-prj",
>                        location="global",
>                        credentials=gcp_credentials,
>                    )
>                elif use_bedrock:
>                    llm = ChatBedrock(
>                        model_id=bedrock_model_id,
>                        region_name=bedrock_llm_region,
>                        provider="anthropic",
>                        model_kwargs={"temperature": 0.1},
>                        max_tokens=15000, # limit: 64,000
>                    )
>
>                llm_with_structured_output = runnable_with_retry(llm.with_structured_output(STRUCTURED_OUTPUT_SCHEMA))
>                llm_with_tools = runnable_with_retry(llm_with_structured_output.bind_tools(tools))