- Security ListというのとNetwork Security Groupという２つがある


## Network Security Group
- AWSでいうとSecurity Groupのようなものでインスタンス(VNIC)などのリソースにアタッチする
- `Destination Port`が受け付けるポートで、許可するポートを指定する
  - `Source Port`は基本Allにする

## Security List（SL）
- AWSでいうとNACLのようなもので**サブネット**にアタッチする
> [!NOTE]  
> Security Listで許可しているものは、該当Security Listが存在するサブネット上にあるリソース(e.g. インスタンス)で許可していない通信もできてしまうので基本Security Listは使わずにNSGを使うこと
