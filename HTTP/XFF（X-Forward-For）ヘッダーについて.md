## XFF（X-Forward-For）ヘッダーとは
- https://developer.mozilla.org/ja/docs/Web/HTTP/Headers/X-Forwarded-For
- HTTP プロキシーサーバーを通過してウェブサーバーへ接続したクライアントの、送信元 IP アドレスを特定するために事実上の標準となっているヘッダー
  - XFFがない場合、クライアントからのリクエストがプロキシーやロードバランサーを経由する時、送信元IPがプロキシーやロードバランサーのIPになってて、クライアントのIPアドレスを知ることができないため。
- XFF（X-Forward-For）ヘッダーはプロキシーやロードバランサーなどで付与される。

####  XFF（X-Forward-For）の構文
```
X-Forwarded-For: <client IP>, <proxy1 IP>, <proxy2 IP>
```