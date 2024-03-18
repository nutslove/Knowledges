- firewalldは裏でiptablesを使っている
  - なのでfirewalldを止めるとfirewalldによって設定されたすべてのiptablesルールもなくなる

## iptables
- **https://christina04.hatenablog.com/entry/iptables-outline?utm_source=pocket_saves**
#### NAT
- iptablesはPacket filteringだけではなく、NATも行ってくれる。  
  DockerやKubernetes(default)もiptablesを使ってNATを行っている
- iptablesで設定できるNATの種類  
  | `target` (種類) | 説明 | Chain | 確認コマンド例 | 確認コマンド出力例 |
  | --- | --- | --- | --- | --- |
  | SNAT | SourceIPを変換 | POSTROUTING |  | |
  | DNAT | DestinationIPを変換 | PREROUTING | `iptables -t nat -L DOCKER -vn` | Chain DOCKER (2 references)<br> pkts bytes target     prot opt in     out     source               destination<br>    0     0 RETURN     all  --  docker0 *       0.0.0.0/0            0.0.0.0/0<br> 149K 7574K DNAT       tcp  --  !docker0 *       0.0.0.0/0            0.0.0.0/0            tcp dpt:80 to:172.17.0.3:8000<br>  142  7392 DNAT       tcp  --  !docker0 *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8080 to:172.17.0.4:80 |
  | MASQUERADE | SourceIPを**動的**に変換 | POSTROUTING | `iptables -t nat -L POSTROUTING -nv` | Chain POSTROUTING (policy ACCEPT 720K packets, 54M bytes)<br> pkts bytes target     prot opt in     out     source               destination<br>  22M 1537M MASQUERADE  all  --  *      !docker0  172.17.0.0/16        0.0.0.0/0<br><br>**→outが`!docker`となってて、つまりbridge以外の外部NW向けの通信をSNATするということ** |
  - `iptables`コマンドオプション
    - `-L [<chain名>]`：chain名を指定した場合はそのchainの設定されたルール(ルールセット)が表示される。chain名を指定してない場合はすべてのchainのルールセットが表示される
    - `-t <table名>`：指定したtableの情報のみ表示する
    - `-v`：詳細情報を出力
    - `-n`：ipアドレスやport番号を数字(numeric)で出力

#### Table
- 関連するルールのセットをグループ化するための構造。  
  iptablesにはいくつかの異なるtableがあり、それぞれが異なる種類のtraffic処理に特化している
- defaultで利用可能なtable  
  | 種類 | 説明 | `-t`オプション |
  | --- | --- | --- |
  | Filter Table (default) | 基本的なpacket filtering（許可/拒否）を行うためのtable | `filter` |
  | NAT Table | NAT(SNAT,DNAT,MASQUERADE)のルールを管理するtable | `nat` |
  | Mangle Table | パケットのヘッダー情報（TTL, QoSビットなど）を変更するルールを管理するためのtable | `mangle` |
  | Raw Table | Conntrack（接続追跡）をバイパスするルールを管理するためのtable | `raw` |

#### Chain
- table内のルールの順序付けられたリスト。  
  iptablesにはいくつかのbuilt-in(デフォルト)chainがあり、ユーザーがcustom chainを作成することもできる。  
  Chainは、特定のtableに関連付けられ、そのtableが扱うtrafficの種類に関連するルールを含む。
- 追加時は`-A`オプションでChainを指定
- defaultで利用可能なChain  
  | 種類 | 説明 |
  | --- | --- |
  | `INPUT` | ローカルシステムへの**入力**trafficを処理するルールが含まれる。 |
  | `OUTPUT` | ローカルシステムからの**出力**trafficを処理するルールが含まれる。 |
  | `FORWARD` | ローカルシステムを**通過する**trafficを処理するルールが含まれる（ルータとして動作する場合など）。 |
  | `PREROUTING` | パケットがローカルシステムに到達する前に適用されるルールを持つchain。 このchainは主に、宛先アドレス変換（DNAT: Destination NAT）を実行する。 |
  | `POSTROUTING` | パケットがローカルシステムを通過または出発する直前に適用されるルールを持つchain。このchainは主に、送信元アドレス変換（SNAT: Source NAT または MASQUERADE）を実行するために使用。 |