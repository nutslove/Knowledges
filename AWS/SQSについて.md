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
