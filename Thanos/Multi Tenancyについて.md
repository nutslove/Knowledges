- `Receiver`を使う方式と`Sidecar`を使う方式がある
  - 一般的にはMulti Tenancyのために`Receiver`を使うところが多い

## Receiverを使う方式
- アーキテクチャ（https://thanos.io/v0.8/proposals/201812_thanos-remote-receive/）
![](./image/multi-tenancy-receiver.jpg)