## namespace
- Linuxで使われるProcessやNetworkスタックを隔離するための技術
- namespaceにはいくつか種類がある

### Network namespace
- Linuxのネットワークスタックを隔離するための機能
  - これにより、1つの物理的または仮想的なマシン上で複数の独立したネットワークインターフェイス、IPアドレス、ルーティングテーブル、ファイアウォールのルール、その他のネットワーク関連の状態を持つことができる。各network namespaceは他から独立しており、それぞれが独自のネットワーク環境を持つことができる。
- Network namespaceの確認
  - `ip netns list`
- Network namespaceの作成
  - `ip netns add <namespace名>`
- 特定のNetwork namespace内でのコマンド実行
  - `ip netns exec <namespace名> <コマンド>`
  - 例
    - `ip netns exec <namespace名> arp -a`
    - `ip netns exec <namespace名> ip route`

### PID namespace
- Process(群)を隔離し、同じnamespaceに存在するProcess間だけお互いが見えて(疎通できて)、  
  他のnamespaceにあるProcessは見えない(干渉できない)ようにする

- `/proc`ディレクトリ


## cgroup




## overlayfs