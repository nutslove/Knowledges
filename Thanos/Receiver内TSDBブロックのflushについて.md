- Receiverに、内部で持っているTSDBをバックエンドのObject Storeにflushするためのエンドポイントはないっぽい（2025/05時点）  
  - ただ、自動でPod終了時にflushされるようになっているっぽい
![](./image/receiver_flush_1jpg.jpg)