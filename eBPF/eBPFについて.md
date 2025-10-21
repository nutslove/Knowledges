## eBPFの概要
- eBPF (Extended Berkeley Packet Filter) プログラムは、何かしらのイベントにアタッチされる必要がある
  - カーネルにロードされ、特定のイベントにアタッチし、そのイベントが発生したときに実行される
- kprobe（kernel probe）は、Linuxカーネル内のほとんどの命令に対してフックを入れて、追加の命令を実行できる
- eBPFプログラムもkprobeにアタッチできる（2015年に導入）
- eBPFは非常に強力なため、使う上で特別なLinux Capabilitiesが必要
  - 例えば、`CAP_BPF`と`CAP_PERFMON`が、トレース関連のeBPFプログラムをロードして実行するために必要で、`CAP_NET_ADMIN`と`CAP_BPF`は、ネットワーク関連のeBPFプログラムをロードして実行するために必要
  - なので、rootユーザーで実行するか、適切なCapabilitiesを持つ必要がある
- eBPFプログラムは、マシンやプロセスの再起動は不要で、動的にロードおよびアンロードできる

> [!TIP]  
> ## kprobe（kernel probe）とは
> - Linuxカーネル内部の任意の関数に動的にフック（hook）を挿入して、その関数の実行をトレース（観測）するための仕組み
> - 「カーネルの動作を変更せずに、実行中のカーネルの中身を覗く（観測する）」ための機構
> - Linuxカーネルは、さまざまな関数で構成されている（e.g. `do_sys_open()`、`tcp_sendmsg()`など）
> - kprobeは、これら任意の関数の任意の命令位置にブレークポイントを設置し、設置したポイント（プローブ）に到達したとき、ユーザーが定義したハンドラ関数（例えば`pre_handler`や`post_handler`）が呼ばれる

## eBPF MAP
- eBPF MAPは、eBPFプログラムと user space アプリケーション間や、eBPFプログラム同士でデータを共有するためのデータ構造
- eBPF MAPには、ハッシュテーブル、配列、リングバッファ（perfリングバッファ、BPFリングバッファ）など、さまざまなタイプがある

## BCC (BPF Compiler Collection)
- eBPF (extended Berkeley Packet Filter) を使って、カーネルをトレース・モニタリング・デバッグするための強力なフレームワーク／ツール群