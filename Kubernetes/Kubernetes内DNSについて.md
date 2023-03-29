- 参考URL
  - https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
  - https://www.nslookup.io/learning/the-life-of-a-dns-query-in-kubernetes/
  - https://zenn.dev/toversus/articles/d9faba80f68ea2

### DNS Server in k8s
- Add-onとしてDNSサーバ(Pod)が作成される
- v1.11以降から*CoreDNS*が推奨されていて、EKSもCoreDNSがAdd-onとしてデフォルトで払い出される
  > Kubernetes offers a DNS cluster addon, which most of the supported environments enable by default. In Kubernetes version 1.11 and later, CoreDNS is recommended and is installed by default with kubeadm.
  - https://kubernetes.io/docs/tasks/access-application-cluster/configure-dns-cluster/

### 概要
![k8s_dns](image/k8s_dns.jpg)
![k8s_dns2](image/k8s_dns2.jpg)
https://speakerdeck.com/hhiroshell/kubernetes-network-fundamentals-69d5c596-4b7d-43c0-aac8-8b0e5a633fc2?slide=42
https://speakerdeck.com/hhiroshell/kubernetes-network-fundamentals-69d5c596-4b7d-43c0-aac8-8b0e5a633fc2?slide=44