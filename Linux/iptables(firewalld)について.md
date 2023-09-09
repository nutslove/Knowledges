- firewalldは裏でiptablesを使っている
  - なのでfirewalldを止めるとfirewalldによって設定されたすべてのiptablesルールもなくなる

## iptables
#### NAT
- iptablesはPacket filteringだけではなく、NATも行ってくれる。  
  DockerやKubernetes(default)もiptablesを使ってNATを行っている
- iptablesで設定できるNATの種類  
  | `target` (種類) | 説明 | 設定可能な`table` | 確認コマンド例 | 確認コマンド出力例 |
  | --- | --- | --- | --- | --- |
  | SNAT | SourceIPを変換 | POSTROUTING | MASQUERADE | |
  | DNAT | DestinationIPを変換 |PREROUTING | `iptables -t nat -L DOCKER -v -n` | Chain DOCKER (2 references)<br> pkts bytes target     prot opt in     out     source               destination<br>    0     0 RETURN     all  --  docker0 *       0.0.0.0/0            0.0.0.0/0<br> 149K 7574K DNAT       tcp  --  !docker0 *       0.0.0.0/0            0.0.0.0/0            tcp dpt:80 to:172.17.0.3:8000<br>  142  7392 DNAT       tcp  --  !docker0 *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8080 to:172.17.0.4:80 |
  | MASQUERADE | SourceIPを**動的**に変換 |  | |
  - `iptables`コマンドオプション
    - `-v`：詳細情報を出力
    - `-n`：ipアドレスやport番号を数字(numeric)で出力
- 参考URL
  - https://qiita.com/Shakapon/items/d29b0af036bf6796feb2
#### Chain
- 