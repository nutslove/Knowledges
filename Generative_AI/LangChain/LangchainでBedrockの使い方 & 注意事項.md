# Cross-region inference使用方法について
## Cross region inference（クロスリージョン推論）
- https://docs.aws.amazon.com/ja_jp/bedrock/latest/userguide/cross-region-inference.html
  > When running model inference in on-demand mode, your requests might be restricted by service quotas or during peak usage times. Cross-region inference enables you to seamlessly manage unplanned traffic bursts by utilizing compute across different AWS Regions. With cross-region inference, you can distribute traffic across multiple AWS Regions, enabling higher throughput.

## LangChainでCross region inferenceを使う方法
1. まずCross region inferenceのProfile ARNを確認する
   - 「Inference and Assessment」の「Cross-region inference」から確認可能  
    ![](./image/bedrock_cross_region_inference_1.png)
2. `ChatBedrock`などでインスタンスを初期化するときに`model_id`に1.で確認したProfile ARNを指定し、`provider`に`"anthropic"`を、`region_name`に対象対象リージョンを指定する  
   ```python
   llm = ChatBedrock(
       model_id="arn:aws:bedrock:us-west-2:1234567890:inference-profile/us.anthropic.claude-3-7-sonnet-20250219-v1:0",
       region_name="us-west-2",
       provider="anthropic",
       model_kwargs={"temperature": 0.1}
   )
   ```

# BedrockのConfig設定方法
- `ChatBedrock`初期化時に`config`に`Config(<Key>=<Value>)`形式で指定する  
  ```python
  from langchain_aws import ChatBedrock
  from botocore.config import Config

  llm_with_tools = ChatBedrock(
      model_id=model_id,
      region_name=llm_region,
      provider="anthropic",
      model_kwargs={"temperature": 0.1},
      max_tokens=4096, # default: 1024
      config=Config(read_timeout=60)
  ).bind_tools(tools)
  ```
- `config`に指定するのは`botocore.config.Config`のインスタンス
  - https://botocore.amazonaws.com/v1/documentation/api/latest/reference/config.html

# `Input is too long for requested model`エラーについて
- エラーメッセージ  
  ```shell
  Error processing messages: An error occurred (ValidationException) when calling the InvokeModel operation: Input is too long for requested model.
  ```
- TokenがLLM ModelのLimitを超えたため出るエラー
- 参考URL
  - https://repost.aws/questions/QUshd0uzCZRAy1TbudkUKhww/claude-on-bedrock-giving-input-is-too-long-for-requested-model-for-10k-token-inputs-edit-broken-in-eu-central-1-working-in-other-regions

# `input length and `max_tokens` exceed context limit` エラーについて
- エラーメッセージ  
  ```shell
  ValidationException('An error occurred (ValidationException) when calling the InvokeModel operation: input length and `max_tokens` exceed context limit: 201850 + 4096 > 204698, decrease input length or `max_tokens` and try again')Traceback (most recent call last):
  ```
- TokenがLLM ModelのLimitを超えたため出るエラー