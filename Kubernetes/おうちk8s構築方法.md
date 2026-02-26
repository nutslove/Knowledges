# kubeadmを使ってインストール
- https://kubernetes.io/ja/docs/setup/production-environment/tools/kubeadm/install-kubeadm/

## containerdインストール
- https://github.com/containerd/containerd/blob/main/docs/getting-started.md

## CNIインストール
### Calico
- Quickstart on Kubernetesではなく、以下URLの「Manifest」の「Install Calico with Kubernetes API datastore, 50 nodes or less」の部分の手順でインストールすること
- https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises#install-calico

#### Calico IPIPモード
- 現在使っているのはCalicoのIPIPモードで、各ノード上の `tunl0` デバイス（IPIPトンネルインターフェース）がIPパケットを別のIPヘッダーで包む（IP-in-IP encapsulation）。これもLinuxカーネルのIPIPモジュールが処理する（VXLANとは異なる）。
  - VXLANはL2オーバーレイで、L2フレームをUDPパケットでカプセル化する。一方、IPIPはL3オーバーレイで、IPパケットを別のIPヘッダーで包む。
- 各ノード上で動作する `calico-node` 内の **BIRD BGPデーモン**が、「このPod CIDRへのパケットはこのノードに送れ」というルート情報を他のノードと交換する。これにより各ノードのルーティングテーブルが構築される。
- **IPIPトンネルの役割**： BGPで学習したルートのNext Hopに到達する際に、元のIPパケットを外側のIPヘッダーで包んでノード間を転送する。これにより、異なるノード上のPod同士が通信できるようになる。
- **流れ**
  1. BIRD（BGP） がノード間でPodのルート情報を交換
  2. 各ノードのルーティングテーブルに 10.244.1.0/24 via tunl0 のようなエントリが作られる
  3. パケット送信時に tunl0 デバイス（カーネルのIPIPモジュール） がカプセル化を実行
- BIRD（BGP）によって学習されたルート情報はノード内で `ip route` コマンドで確認できる
  - 例  
    ```shell
    root@workernode01:~# ip route
    default via 192.168.0.1 dev enp1s0 proto static 
    blackhole 172.16.52.128/26 proto bird 
    172.16.52.132 dev calib472a44443d scope link # 同じノード上のPodのIPアドレス（veth pair のホスト側端点に直接ルーティングされる） 
    172.16.52.134 dev calia863a3ed0c4 scope link # 同じノード上のPodのIPアドレス（veth pair のホスト側端点に直接ルーティングされる） 
    172.16.52.135 dev cali6973aefed8d scope link # 同じノード上のPodのIPアドレス（veth pair のホスト側端点に直接ルーティングされる） 
    172.16.103.64/26 via 192.168.0.146 dev tunl0 proto bird onlink # 別ノード上のPodのIPアドレス（IPIPトンネル経由でルーティングされる。172.16.103.64/26が別ノードのPod CIDRで、192.168.0.146がそのノードの物理NICのIPアドレス）
    172.16.246.192/26 via 192.168.0.241 dev tunl0 proto bird onlink # 別ノード上のPodのIPアドレス（IPIPトンネル経由でルーティングされる。172.16.246.192/26が別ノードのPod CIDRで、192.168.0.241がそのノードの物理NICのIPアドレス）
    192.168.0.0/24 dev enp1s0 proto kernel scope link src 192.168.0.176 # 「192.168.0.0/24 宛の通信は、NIC enp1s0 から直接送信し、送信元IPとして 192.168.0.176 を使う」という意味
    ```
    - ポイント：
      - `proto bird` — BIRDデーモン（BGP）によって追加されたルートであることを示す
      - `dev tunl0` — IPIPトンネルデバイス経由で送信される（カプセル化される）ことを意味する
      - `via 192.168.0.xx` — 宛先ノードの実IPアドレス（外側IPヘッダーのdstになる）
      - `blackhole` — 自ノードに割り当てられたPod CIDR全体に対するブラックホールルート。個々のPodが起動すると `/32` のより具体的なルート（例：`172.16.52.132 dev caliXXXX`）が追加され、ロンゲストマッチによりそちらが優先される。結果として、どのPodにも該当しないアドレス宛のパケットだけがこのblackholeルートで破棄される
      - `onlink` — Next HopがL2的に直接到達可能でなくても強制的にそのデバイスから送出する指示

> [!NOTE]  
> `172.16.246.192/26 via 192.168.0.241 dev tunl0 proto bird onlink`は、**172.16.246.192/26**宛のパケットを`tunl0`でIPIPカプセル化し、外側IPヘッダーの宛先を**192.168.0.241**に設定するという意味。カプセル化後、外側IPヘッダのIP**192.168.0.241**で`192.168.0.0/24 dev enp1s0 proto kernel scope link src 192.168.0.176`のルートに従って、物理NIC `enp1s0` から送信される。
> #### 処理の流れ
> ① Pod → 172.16.103.65 宛のパケット送出  
> ② ルーティング検索 → 172.16.103.64/26 via 192.168.0.146 dev tunl0 にマッチ  
> ③ tunl0 がIPIPカプセル化  
> [Outer IP: 192.168.0.176 → 192.168.0.146] [Inner IP: PodA → 172.16.103.65] [Payload]  
> ④ カプセル化後のパケット（宛先: 192.168.0.146）が再度ルーティング検索される  
> ⑤ 192.168.0.0/24 dev enp1s0 にマッチ → enp1s0 から物理NWへ送出

> [!NOTE]  
> - Calico IPIPモードはBGPによるRoutingも使うけど、Routingモードではなく、Overlayモードに分類される
> - Routingモードの特徴として、カプセル化なしで、ルーティングのみで、Pod間通信を実現するのがあって、Calico BGP（Directモード）がそれに該当する  
> https://www.alibabacloud.com/blog/getting-started-with-kubernetes-%7C-kubernetes-cnis-and-cni-plug-ins_596330

## 注意事項
1. kubeletが正常に動作するためにはSwapを無効にしないといけない
   - Ubuntuの場合、`swapoff -a`コマンドでSwapを無効にできる
     - `swapon --show`でon/offの確認ができる
2. CalicoなどCNIを入れるまではCoreDNSのPodはRunningにならない  
   > Pod同士が通信できるようにするには、Container Network Interface(CNI)をベースとするPodネットワークアドオンをデプロイしなければなりません。ネットワークアドオンをインストールする前には、Cluster DNS(CoreDNS)は起動しません。

---

# IPIP / VXLAN メモ

## 1. デカプセル化の仕組み — ルートテーブルは使わない

カプセル化されたパケットを受信したとき、デカプセル化はルートテーブルではなく**カーネルのプロトコル処理レイヤー**で行われる。

### IPIP の場合
- 外側IPヘッダの**プロトコル番号が4**（IPIP）であることをカーネルが検出
- `ip_local_deliver_finish` でプロトコル番号を判定し、IPIPカーネルモジュール（tunl0）にパケットを直接渡す
- tunl0が外側ヘッダを剥がして（デカプセル化）、内側パケットがルートテーブルで再検索される

```
ip_rcv → ip_local_deliver → ip_local_deliver_finish
  → Protocol=4 を検出 → tunnel4_rcv → ip_tunnel_rcv（tunl0がデカプセル化）
  → 内側パケットをルートテーブルで検索 → caliXXX(veth) → Pod
```

参考: https://chenchun.github.io/network/2017/10/24/ipip

### VXLAN の場合
- 外側IPヘッダのプロトコル番号は**17（UDP）**
- UDPスタックに渡り、**ポート4789で待ち受けているVXLANモジュール（VTEP）** がパケットを受け取る
- VXLANヘッダのVNIを確認後、デカプセル化して内側パケットがルートテーブルで再検索される

```
ip_rcv → ip_local_deliver → UDPスタック（Protocol=17）
  → ポート4789 → VXLANモジュール（vxlan.calico）がデカプセル化
  → 内側パケットをルートテーブルで検索 → caliXXX(veth) → Pod
```

### 共通点
- どちらもデカプセル化にルートテーブルは関与しない
- ルートテーブルが使われるのは**デカプセル化後**の内側パケットに対して

---

## 2. カプセル化は送信元IPも書き換わる

カプセル化では宛先IPだけでなく、**送信元IPもノードのIPになる**。

```
Outer IP Header:  送信元 = 送信元ノードのIP    宛先 = 宛先ノードのIP
Inner IP Header:  送信元 = 送信元PodのIP       宛先 = 宛先PodのIP
```

具体例:
```
[Outer: 192.168.0.176 → 192.168.0.146] [Inner: 172.16.52.132 → 172.16.103.65] [Payload]
         自ノード        相手ノード            Pod A              Pod B
```

### なぜ送信元もノードIPにするのか
- 物理ネットワーク上のルーターはPodのIP（172.16.x.x）への戻り経路を知らない
- 外側の送信元がPodのIPだと応答パケットが返ってこれない
- **内側のIPは一切書き換えていない**ので、デカプセル化すれば元のPod IPが取り出され、K8sの「NATなし通信」の要件を満たせる

---

## 3. IPIP と VXLAN の違い

| 項目 | IPIP | VXLAN |
|------|------|-------|
| カプセル化レイヤー | L3（IPヘッダのみ） | L2（Ethernetフレームごと） |
| オーバーヘッド | 20 bytes（小さい） | 50 bytes（大きい） |
| プロトコル | IP Protocol 4 | UDP ポート 4789 |
| MAC情報の保持 | なし | あり |
| マルチテナント（VNI） | 非対応 | 対応（最大16M） |
| ファイアウォール通過 | △ プロトコル4がブロックされることがある | ○ UDPなので通りやすい |
| パフォーマンス | ◎ ヘッダ小＆シンプル | ○ 良好 |
| CNIの例 | Calico（デフォルト） | Flannel, Calico, Cilium |
| トンネルデバイス名（Calico） | tunl0 | vxlan.calico |
| デカプセル化の判定 | IPヘッダのプロトコル番号 | UDPポート番号 |

### パケット構造の比較

**IPIP:**
```
[Outer IP: Node→Node] [Inner IP: Pod→Pod] [Payload]
```

**VXLAN:**
```
[Outer IP: Node→Node] [Outer UDP: :4789] [VXLAN Header(VNI)] [Inner Ethernet] [Inner IP: Pod→Pod] [Payload]
```

### 使い分けの目安
- **IPIP**: 同一L3ネットワーク内でシンプルかつ高速に通信したい場合
- **VXLAN**: クラウド環境やFW越え、L2情報の保持が必要な場合

---

# Trouble Shooting
## 1. WorkNodeがNotReadyの状態になる
### 事象
- Worknode01,02がNotReady状態になり、k8sクラスターが機能しなくなった

### やったこと
- `kubelet`からswap領域がonになっているとエラーが出て、`kubelet`が異常終了していた。(`swapon --show`で何かが表示されたらswap領域がonになっている状態)
  - **`swapoff -a`コマンドでSwap領域を無効にし、`systemctl restart kubelet`でkubeletは起動した。** しかし、それでもWorkerNodeは一瞬ReadyになるけどすぐNotReadyに戻ってしまった
- kubeletのログ（`journalctl -u kubelet`）から `kubelet[1817]: E0826 16:00:58.661382    1817 kubelet.go:2412] "Skipping pod synchronization" err="container runtime is down"` のエラーが出ていた。これは kubelet が container runtime（この場合 containerd）と通信できないと判断している状態だそう。
- `crictl info`で確認したら以下のようなエラーが出た  
  ```shell
  E0826 16:07:03.766916   10146 log.go:32] "Status from runtime service failed" err="rpc error: code = DeadlineExceeded desc = context deadline exceeded"
  FATA[0002] getting status of runtime: rpc error: code = DeadlineExceeded desc = context deadline exceeded
  ``` 
- **`systemctl restart containerd`でcontainerdを再起動**して、再度`crictl info`を実行したらjsonのconfigデータが返ってくるようになった。

> [!TIP] 
> ### CRI（Container Runtime Interface）
> - kubeletとコンテナランタイム（containerd、CRI-O、etc.）が通信するための「仕様・標準インターフェース」。このインターフェースにより、KubernetesはDockerだけでなく、CRI-Oやcontainerdなど様々なコンテナランタイムをプラグインのように利用できるようになった
> - kubeletはコンテナ実行基盤（containerd / CRI-O など）と直接ではなく CRI (gRPC API) を介してやり取りしている
> - `crictl`はその gRPC API を直接叩くコマンドラインツール

> [!NOTE]  
> - 上記とは別で、WorkerNodeの物理NICのIPv4のIPアドレスがなくなっていた（`ip a`コマンドで確認）
> - `reboot`したらIPアドレスが復活して、上記のsawp無効化とcontainerd再起動でWorkerNodeは正常に動作するようになった


## 2. Master Node, Worker Nodeすべて１回落ちたあとの復旧手順
### 事象
- Master Node, Worker Nodeすべて電源問題で落ちた
- その後、ノードを起動したが、Master Nodeを含むノードのIPアドレスが変わってしまい、k8sクラスターが機能しなくなった
- kubectlコマンドで以下のようなエラーが出る  
  ```shell
  Unable to connect to the server: dial tcp <IP>:6443: connect: connection refused

  Unable to connect to the server: tls: failed to verify certificate: x509: certificate is valid for <旧IP>, not <新IP>
  ```

> [!NOTE]  
> - Ubuntuの場合、DHCPでIPアドレスが変わらないようにするには、`/etc/netplan/nn-xx.yaml`を以下のように修正する  
>  ```yaml
> network:
>   ethernets:
>       enp1s0:
>           dhcp4: no # ここをtrueからnoに変更
>           addresses:
>             - 192.168.0.176/24 # 固定IPアドレスを設定
>           routes:
>             - to: default
>               via: 192.168.0.1 # デフォルトゲートウェイを設定(ip routeで確認)
>           nameservers:
>             addresses:
>               - 8.8.8.8
>   version: 2
> ``` 
> - 修正後、`sudo netplan apply`コマンドで反映する

### やったこと
- クライアント側のkubeconfigを修正  
  ```shell
  grep "server:" ~/.kube/config
  sed -i 's/Master Nodeの旧IP/Master Nodeの新IP/g' ~/.kube/config
  ```
#### Master Node側の修正
- Master Node側で、etcdとkube-apiserverの設定ファイルを修正  
  ```shell
  # etcdの設定ファイル修正
  sudo sed -i 's/Master Nodeの旧IP/Master Nodeの新IP/g' /etc/kubernetes/manifests/etcd.yaml

  # kube-apiserverの設定ファイル修正
  sudo sed -i 's/Master Nodeの旧IP/Master Nodeの新IP/g' /etc/kubernetes/manifests/kube-apiserver.yaml
  ```
- Master Nodeのkubeletを再起動  
  ```shell
  sudo systemctl restart kubelet
  ```
- kube-apiserverとetcdが起動したことを確認  
  ```shell
  sudo crictl ps | grep kube-apiserver
  sudo crictl ps | grep etcd
  ```
- `/etc/kubernetes/admin.conf`の`server:`のIPアドレスを新しいIPアドレスに修正  
  ```shell
  sudo sed -i 's/Master Nodeの旧IP/Master Nodeの新IP/g' /etc/kubernetes/admin.conf
  ```
- API Serverの証明書を再生成  
  ```shell
  # 古い証明書をバックアップ
  sudo mv /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/apiserver.crt.bak
  sudo mv /etc/kubernetes/pki/apiserver.key /etc/kubernetes/pki/apiserver.key.bak

  # 新IPで証明書を再生成
  sudo kubeadm init phase certs apiserver --apiserver-advertise-address=新IP

  # API Serverを再起動
  sudo systemctl restart kubelet
  ```
- しばらくしてからkubectlコマンドが通ることを確認

## 3. 突然kubectlコマンドで認証エラーとなる
### 事象
- ある日突然（それまでは正常に動作していた）、kubectlコマンドで以下のような認証エラーが発生するようになった  
  ```shell
  E0227 01:09:47.860588   14918 memcache.go:265] "Unhandled Error" err="couldn't get current server API group list: the server has asked for the client to provide credentials"
  error: You must be logged in to the server (the server has asked for the client to provide credentials)
  ```

### 原因
- クラスター内の証明書の有効期限が切れていた
- master node上で以下のコマンドで証明書の有効期限を確認できる  
  ```shell
  kubeadm certs check-expiration
  ```

### 対処
- master node上でkubeadmコマンドを使用して証明書を更新する  
  ```shell
  kubeadm certs renew all
  ```
- kubeletを再起動する  
  ```shell
  systemctl restart kubelet
  ```
- master node上の`/etc/kubernetes/admin.conf`の内容をクライアント側の`~/.kube/config`に反映する  