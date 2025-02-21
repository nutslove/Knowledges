## Knowledge Base
- Vector StoreとしてAurora PostgreSQL Serverlessを使う場合、自動でauto pause設定が入っていて、デフォルトだと3時間アイドルタイム後、一時停止される。このアイドル時間は最大24時間まで設定できる
- pause状態のときにKnowledge Baseを呼び出すと以下のようなエラーが出る  
  ```
  Error processing messages: An error occurred (ValidationException) when calling the Retrieve operation: The vector database encountered an error while processing the request: The Aurora DB instance db-xxxxxxxxx is resuming after being auto-paused. Please wait a few seconds and try again. (Service: RdsData, Status Code: 400, Request ID: cfc79014-f252-4757-a0ad-068147601b60)
  ```