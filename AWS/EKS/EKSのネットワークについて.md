- 一般的なKubernetesのネットワークについては[KubernetesフォルダのKubernetesのNetworkについて.md](../../Kubernetes/KubernetesのNetworkについて.md)にまとめている

# IP Mode と　Instance Modeについて
- `type: LoadBalancer`の`Service`リソースと、`Ingress`リソースで、NLB \ALBからPodへトラフィックを転送する際のモードとして、IP ModeとInstance Modeがある
- EKS Auto ModeではdefaultでIP Modeが使用される
  - https://docs.aws.amazon.com/eks/latest/userguide/auto-networking.html  
    > ## Load balancing
    > You configure AWS Elastic Load Balancers provisioned by EKS Auto Mode using annotations on Service and Ingress resources.
    > For more information, see [Create an IngressClass to configure an Application Load Balancer](https://docs.aws.amazon.com/eks/latest/userguide/auto-configure-alb.html) or [Use Service Annotations to configure Network Load Balancers](https://docs.aws.amazon.com/eks/latest/userguide/auto-configure-nlb.html).
    > ### Considerations for load balancing with EKS Auto Mode
    > - The default targeting mode is IP Mode, not Instance Mode.
    > - EKS Auto Mode only supports Security Group Mode for Network Load Balancers.
    > - AWS does not support migrating load balancers from the self managed AWS load balancer controller to management by EKS Auto Mode.
    > - The `networking.ingress.ipBlock` field in `TargetGroupBinding` spec is not supported.
    > - If your worker nodes use custom security groups (not `eks-cluster-sg-` naming pattern), your cluster role needs additional IAM permissions. The default EKS-managed policy only allows EKS to modify security groups named `eks-cluster-sg-`. Without permission to modify your custom security groups, EKS cannot add the required ingress rules that allow ALB/NLB traffic to reach your pods.

## IP ModeとInstance Modeの違い
### IP Mode
- NodePortとiptablesルールによるNATなどを経由せず、NLB/ALBからPodのIPアドレスに直接トラフィックを転送するモード
- Podに直接トラフィックが届くため、レイテンシーが低くなる

### Instance Mode
- NLB/ALBからワーカーノードのNodePortにトラフィックを転送し、
  kube-proxyが作成したiptablesルールによってPodにルーティングされるモード