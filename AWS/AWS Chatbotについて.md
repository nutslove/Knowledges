## AWS ChatBot（Slack）とBedrock Agentの設定
- https://qiita.com/moritalous/items/b63d976c2c40af1c39e5
- SlackにてAWS ChatBotからBedrock Agentを使って直接Bedrockを呼び出すことができる
- 会話履歴も踏まえて回答してくれる
- Knowledge Baseとの連動もできる

## ChatBotからその他AWSサービスの実行
- AWS CLIやLambda関数などを実行することもできる
  - https://docs.aws.amazon.com/ja_jp/chatbot/latest/adminguide/intro-to-the-aws-cli-in-slack.html
  - https://docs.aws.amazon.com/ja_jp/chatbot/latest/adminguide/common-use-cases.html
  - https://docs.aws.amazon.com/ja_jp/chatbot/latest/adminguide/chatbot-run-lambda-function-remotely-tutorial.html
- **`@aws help services`** で使えるサービス一覧を確認できる
- あるサービスの使い方を知りたい場合は`@aws <service名> --help`で確認できる
- `@aws switch-role`でスイッチロールもできる？
