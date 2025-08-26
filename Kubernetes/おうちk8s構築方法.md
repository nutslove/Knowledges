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
- `kubelet`からswap領域がonになっているとエラーが出て、`kubelet`が異常終了していた。
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