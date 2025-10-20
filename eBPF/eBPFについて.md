## eBPFの概要
- eBPF (Extended Berkeley Packet Filter) プログラムは、何かしらのイベントにアタッチされる必要がある
  - カーネルにロードされ、特定のイベントにアタッチし、そのイベントが発生したときに実行される
- kprobe（kernel probe）は、Linuxカーネル内のほとんどの命令に対してフックを入れて、追加の命令を実行できる
- eBPFプログラムもkprobeにアタッチできる（2015年に導入）
- eBPFは非常に強力なため、使う上で特別なLinux Capabilitiesが必要
  - 例えば、`CAP_BPF`と`CAP_PERFMON`が、トレース関連のeBPFプログラムをロードして実行するために必要で、`CAP_NET_ADMIN`と`CAP_BPF`は、ネットワーク関連のeBPFプログラムをロードして実行するために必要
  - なので、rootユーザーで実行するか、適切なCapabilitiesを持つ必要がある
- eBPFプログラムは、マシンやプロセスの再起動は不要で、動的にロードおよびアンロードできる

## eBPF MAP
- eBPF MAPは、eBPFプログラムと user space アプリケーション間や、eBPFプログラム同士でデータを共有するためのデータ構造