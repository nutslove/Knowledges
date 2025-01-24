# 目次 <!-- omit in toc -->
<!-- TOC -->

- [作成方法](#作成方法)
  - [Socket Modeについて](#socket-modeについて)
- [slack_bolt](#slack_bolt)
  - [`process_before_response`について](#process_before_responseについて)
- [Event契機でApp実行](#event契機でapp実行)
  - [Event一覧](#event一覧)
  - [共通設定](#共通設定)
  - [ボットがメンションされたときに反応するApp](#ボットがメンションされたときに反応するapp)
  - [特定のメッセージに反応するApp](#特定のメッセージに反応するapp)
- [AWS Lambdaと連携](#aws-lambdaと連携)

<!-- /TOC -->

# 作成方法
- https://tools.slack.dev/bolt-python/ja-jp/getting-started/
- https://tools.slack.dev/bolt-python/api-docs/slack_bolt/kwargs_injection/args.html
- https://dev.classmethod.jp/articles/amazon-bedrock-slack-chat-bot-part2/

## Socket Modeについて
- https://dev.classmethod.jp/articles/amazon-bedrock-slack-chat-bot-part1/

# slack_bolt
- A Python framework to build Slack apps
- 参考URL
  - https://github.com/slackapi/bolt-python?tab=readme-ov-file
  - https://tools.slack.dev/bolt-python/ja-jp/getting-started/
  - https://tools.slack.dev/bolt-python/api-docs/slack_bolt/index.html

## `process_before_response`について
- Bolt for Python では、デフォルトではすべてのリクエストを処理した後にレスポンスを返す
- しかし、`process_before_response=True`に設定するとリクエスト処理中にレスポンスを先に返し、その後バックグラウンドで処理を続けることができる
- 例  
  ```python
  from slack_bolt import App

  app = App(process_before_response=True)

  @app.event("message")
  def handle_message(event, say):
      say("Processing your message...")
      # 長時間の処理をここで行う
      print(f"Message: {event['text']}")

  if __name__ == "__main__":
      app.start(3000)
  ```

# Event契機でApp実行
## Event一覧
- https://api.slack.com/events

## 共通設定
- 環境変数で`SLACK_BOT_TOKEN`(`xoxb-***`)と`SLACK_APP_TOKEN`(`xapp-***`)を設定

## ボットがメンションされたときに反応するApp
- 参考URL
  - https://api.slack.com/events/app_mention
- 「Features」-「Event Subscriptions」の「Subscribe to bot events」にて`app_mention`を追加する  
  ![](./image/event_subscriptions_app_mention.jpg)
- 「Features」-「OAuth & Permissions」の「Scopes」にて`app_mentions:read`を追加する  
  ![](./image/permission_for_app_mention.jpg)
- `@app.event("app_mention")`でイベントをcatchする  
  ```python
    import os
    from slack_bolt import App
    from slack_bolt.adapter.socket_mode import SocketModeHandler

    # ボットトークンとソケットモードハンドラーを使ってアプリを初期化
    app = App(token=os.environ.get("SLACK_BOT_TOKEN"))

    @app.event("app_mention")
    def message_hello(event, say):
        # イベントがトリガーされたチャンネルへ say() でメッセージを送信
        text = event["text"]
        print(text)
        # say(f"メンションを受け取りました: {text}")
        say(f"こんにちは、<@{event['user']}> さん！")

    if __name__ == "__main__":
        SocketModeHandler(app, os.environ["SLACK_APP_TOKEN"]).start() # アプリを起動
  ```
- `/invite @<ボット名>`でチャネルにボットを参加させる

## 特定のメッセージに反応するApp
- https://tools.slack.dev/bolt-python/ja-jp/getting-started/
- 例  
  ```python
  import os
  from slack_bolt import App
  from slack_bolt.adapter.socket_mode import SocketModeHandler

  # ボットトークンを渡してアプリを初期化します
  app = App(token=os.environ.get("SLACK_BOT_TOKEN"))

  # 'こんにちは' を含むメッセージをリッスンします
  # 指定可能なリスナーのメソッド引数の一覧は以下のモジュールドキュメントを参考にしてください：
  # https://tools.slack.dev/bolt-python/api-docs/slack_bolt/kwargs_injection/args.html
  @app.message("こんにちは")
  def message_hello(message, say):
      # イベントがトリガーされたチャンネルへ say() でメッセージを送信します
      say(f"こんにちは、<@{message['user']}> さん！")

  if __name__ == "__main__":
      # アプリを起動して、ソケットモードで Slack に接続します
      SocketModeHandler(app, os.environ["SLACK_APP_TOKEN"]).start()
  ```

# AWS Lambdaと連携
- https://www.beex-inc.com/blog/slackbot-aws-lambda-python
