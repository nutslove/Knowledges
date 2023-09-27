### Private HostZoneについて
- https://docs.aws.amazon.com/ja_jp/Route53/latest/DeveloperGuide/hosted-zone-private-considerations.html
- https://docs.aws.amazon.com/ja_jp/vpc/latest/userguide/vpc-dns.html

### Route53 Outbound Resolverについて
- VPCに関連付けて、特定のドメインのみ別のNameServerに転送するように設定できる

### Route53 Inbound Resolverについて
- VPC内のDNSサーバと連携されるInbound Resolverを作成して、オンプレ上のサーバの`/etc/resolv.conf`に、もしくはオンプレ上のDNSサーバに特定のドメインはInbound Resolverで作成されたNameServer(2つ)を指定することで、オンプレからVPC内のAWSサービスのDNS名(e.g. ALBのDNS名)を名前解決するができる
  - もちろんオンプレとAWS間はDirect ConnectやSite-to-Site VPNでつながっている必要がある