##　`chat_stream`メソッド

- 参考 URL
  - **https://docs.slack.dev/changelog/2025/10/7/chat-streaming/**
- `slack_sdk`のバージョン 3.37.0 に AI(LLM) response 向けに追加された機能
- `chat_stream`メソッドは`ChatStream`オブジェクトを返す
  - https://github.com/slackapi/python-slack-sdk/blob/main/slack_sdk/web/client.py

### 基本的な使い方

- **https://docs.slack.dev/changelog/2025/10/7/chat-streaming/**
- `append`メソッドでメッセージをストリーミング出力し、最後に`stop`メソッドでメッセージ送信を完了する

```python
streamer = client.chat_stream(
    channel=channel_id,
    recipient_team_id=team_id,
    recipient_user_id=user_id,
    thread_ts=thread_ts,
)

for event in returned_message:
    streamer.append(markdown_text=f"{chunk-received-from-llm}")

streamer.stop(blocks=feedback_block)
```

### `ChatStream`オブジェクト

- https://github.com/slackapi/python-slack-sdk/blob/main/slack_sdk/web/chat_stream.py
- `chat_stream`メソッドで作成された`ChatStream`オブジェクトの`append`メソッドと`markdown_text`パラメータでストリーミング出力できる
  - `append`メソッドは非同期メソッド(async method)なので、`await`で呼び出す必要がある
- `append`メソッドの`markdown_text`パラメータに指定する文字列の上限は 12,000 文字。
  - `append`なので全体のメッセージの長さ（上限）ではなく、例えば Token 単位でストリーミング出力する場合は１つの Token 分の長さの上限が 12,000 文字になるため、現実的には問題ないと思われる。
  - https://docs.slack.dev/tools/python-slack-sdk/reference/web/chat_stream.html
  - 全体の text の上限は明示的に書いてないが、通常の chat.postMessage の text の上限と同じく 40,000 文字と思われる。
