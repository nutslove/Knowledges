## NLB
- Target Groupの方に`クライアントIPアドレスの保持`という項目があって、これをオンにするかオフにするかによって、ターゲットに連携される送信元IPアドレスが変わる
  - **オンの時はクライアントのIPアドレスがターゲットにそのまま連携されて、オフにするとNLBのIPがターゲットに連携される**
- **ターゲットの種類によって挙動が異なるらしい**
  - https://dev.classmethod.jp/articles/pondering-source-ip-address-of-network-load-balancer/