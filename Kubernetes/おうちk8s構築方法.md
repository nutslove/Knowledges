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