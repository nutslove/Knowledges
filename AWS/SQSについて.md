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

## 使用上の注意
- consumer側でメッセージを受信し、正常に処理した後はキューからメッセージを削除する必要がある。(受信だけではメッセージはキューから削除されない)
- `WaitTimeSeconds`は１回のポーリングのtimeout秒数。**キュー内に利用可能なメッセージがある場合は待たずに受信する**  
  - https://docs.aws.amazon.com/ja_jp/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html#SQS-ReceiveMessage-request-WaitTimeSeconds
    > The duration (in seconds) for which the call waits for a message to arrive in the queue before returning. If a message is available, the call returns sooner than WaitTimeSeconds. If no messages are available and the wait time expires, the call returns successfully with an empty list of messages.
  - **キュー内にメッセージがある場合でも一定間隔でポーリングしたい場合は`WaitTimeSeconds`を0にしてLong Pollingを無効にして、アプリのロジックで`ReceiveMessage`を呼び出す間隔を調整する必要がある**