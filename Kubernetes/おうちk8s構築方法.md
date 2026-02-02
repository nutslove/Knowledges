# kubeadmを使ってインストール
- https://kubernetes.io/ja/docs/setup/production-environment/tools/kubeadm/install-kubeadm/

## containerdインストール
- https://github.com/containerd/containerd/blob/main/docs/getting-started.md

## CNIインストール
### Calico
- Quickstart on Kubernetesではなく、以下URLの「Manifest」の「Install Calico with Kubernetes API datastore, 50 nodes or less」の部分の手順でインストールすること
- https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises#install-calico

## 注意事項
1. kubeletが正常に動作するためにはSwapを無効にしないといけない
   - Ubuntuの場合、`swapoff -a`コマンドでSwapを無効にできる
     - `swapon --show`でon/offの確認ができる
2. CalicoなどCNIを入れるまではCoreDNSのPodはRunningにならない  
   > Pod同士が通信できるようにするには、Container Network Interface(CNI)をベースとするPodネットワークアドオンをデプロイしなければなりません。ネットワークアドオンをインストールする前には、Cluster DNS(CoreDNS)は起動しません。

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