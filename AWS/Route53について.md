### Private HostZoneについて
- https://docs.aws.amazon.com/ja_jp/Route53/latest/DeveloperGuide/hosted-zone-private-considerations.html
- https://docs.aws.amazon.com/ja_jp/vpc/latest/userguide/vpc-dns.html

### Route53 Outbound Resolverについて
- VPCに関連付けて、特定のドメインのみ別のNameServerに転送するように設定できる

### Route53 Inbound Resolverについて
- VPC内のDNSサーバと連携されるInbound Resolverを作成して、オンプレ上のサーバの`/etc/resolv.conf`に、もしくはオンプレ上のDNSサーバに特定のドメインはInbound Resolverで作成されたNameServer(2つ)を指定することで、オンプレからVPC内のAWSサービスのDNS名(e.g. ALBのDNS名)を名前解決するができる
  - もちろんオンプレとAWS間はDirect ConnectやSite-to-Site VPNでつながっている必要がある

### ACMで作成したTLS証明書をALBに関連付けした場合
- **前提** 
  - Route53で、該当ドメインを登録して、Publicホストゾーンを作成しておく必要がある
- ACMで証明書を作成し、ACM管理画面で「_Route53でレコードを作成_」を押下すると、CNAMEレコードが該当ドメイン(ホストゾーン)に登録される
  - 該当ドメインにCNAMEレコードが登録されたら、ACMの証明書が発行済みステータスとなる
- ALBにACMを関連付けた後、**ALBのDNS名を該当Route53ホストゾーンにCNAMEレコードとして追加する必要がある**
  - **例えば、Route53で`nutslove.net`ドメイン（ホストゾーン）を持っていて、ACMで`portal.nutslove.net`というドメインで証明書を作成した場合、Route53の該当ホストゾーンにレコード名＝`portal`、値＝`<ALBのDNS名>`のCNAMEレコードを追加する必要がある**
- 参考URL
  - [Route53でドメイン取得からACMで証明書を作成し、ALBに関連付ける](https://zenn.dev/moko_poi/articles/d6d459a2b3ae30)