- Security ListというのとNetwork Security Groupという２つがある


## Network Security Group（NSG）
- AWSでいうとSecurity Groupのようなものでインスタンス(VNIC)などのリソースにアタッチする
- `Destination Port`が受け付けるポートで、許可するポートを指定する
  - `Source Port`は基本Allにする
#### Stateless/Stateful
- NSGにはStatelessとStatefulの2種類がある
- Statefulは許可したPortにアクセスされた通信のEgressの通信は無条件で許可する
- Statelessは許可したPortにアクセスされた通信のEgressの通信であってもEgressルールで明示的に許可している必要がある
- **なので基本Statefulを使うこと**
- 参考URL
  - https://docs.oracle.com/ja-jp/iaas/Content/Network/Concepts/securityrules.htm#stateful


## Security List（SL）
- AWSでいうとNACLのようなもので**サブネット**にアタッチする
> [!NOTE]  
> **Security Listで許可しているものは、該当Security Listが存在するサブネット上にあるリソース(e.g. インスタンス)で許可していない通信もできてしまうので基本Security Listは使わずにNSGを使うこと**
