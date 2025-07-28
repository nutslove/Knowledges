## 制約
- https://aws.amazon.com/sqs/faqs/
- キューに保存できるメッセージ数は無限。ただ、consumerによって受信されたけどまだ削除はされてないメッセージ(inflight messages)はlimitがある
  > A single Amazon SQS message queue can contain an unlimited number of messages. **However, there is a quota of 120,000 for the number of inflight messages for a standard queue and 20,000 for a FIFO queue.** Messages are inflight after they have been received from the queue by a consuming component, but have not yet been deleted from the queue.

## 各種メトリクス
- `ApproximateNumberOfMessagesVisible`
  - 現在処理(削除)されずにキューに残っているメッセージ数
  - **キュー内のメッセージ数をもとにAutoScalingする場合はこのメトリクスを使用**
- `ApproximateAgeOfOldestMessage`
  - キュー内に削除されずに残っているメッセージの中で一番古いメッセージの存続期間(秒)
- `NumberOfEmptyReceives`
  - メッセージを受信する側で`ReceiveMessage` APIを呼び出したけどキュー内にメッセージが存在しなかった回数

## Dead Letter Queue（デッドレターキュー）
- 処理に失敗したメッセージを保存するためのキュー
- 一定回数の再試行後も処理に失敗したメッセージをこのキューに送信することで、問題のあるメッセージを隔離し、後で分析や再処理を行うことができる

### 処理の流れと`maxReceiveCount`について
1. メッセージがSQSキューに送信される
2. コンシューマーがメッセージを受信して処理を試みる
3. 処理が成功すればメッセージは削除される
4. 処理が失敗すると、メッセージは再びキューに戻る（visibility timeoutの期間後）

- 上記の受信 → 失敗のサイクルが`maxReceiveCount`で指定した回数に達すると、そのメッセージは自動的にDLQに移動される
例：`maxReceiveCount: 3`の場合、3回処理に失敗したメッセージがDLQに送られる

## 使用上の注意
### 処理後メッセージ削除
- consumer側でメッセージを受信し、正常に処理した後はキューからメッセージを削除する必要がある。(受信だけではメッセージはキューから削除されない)

### `WaitTimeSeconds`について
- `WaitTimeSeconds`は１回のポーリングのtimeout秒数。
- Long Polling
  - `WaitTimeSeconds`を最大20秒まで設定できる
  - SQSはリクエスト数の従量課金なので、Long Pollingでリクエスト数を減らした方が良い
  - Long Pollingを使っても**キュー内に利用可能なメッセージがある場合は待たずにすぐに受信する**  
  - https://docs.aws.amazon.com/ja_jp/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html#SQS-ReceiveMessage-request-WaitTimeSeconds
    > The duration (in seconds) for which the call waits for a message to arrive in the queue before returning. If a message is available, the call returns sooner than WaitTimeSeconds. If no messages are available and the wait time expires, the call returns successfully with an empty list of messages.
  - **キュー内にメッセージがある場合でも一定間隔でポーリングしたい場合は、アプリのロジックで`ReceiveMessage`を呼び出す間隔を調整する必要がある**

### `MessageGroupId`について
- FIFOキューの場合、`MessageGroupId`単位でIn/Outされる。例えば同じ`MessageGroupId`のキューが複数ある場合、前のキューが何らかの理由(そのキューを扱って処理するLambdaが処理に失敗しているとか)で詰まっている場合、その後のキューも取得できなくなる
