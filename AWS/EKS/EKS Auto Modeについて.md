- https://speakerdeck.com/kashinoki38/eks-auto-mode

# Node管理
- defaultのNodeClassとNodePoolを使う場合は追加の設定なしでノードが利用可能で、Podをデプロイすると自動でEC2インスタンスが起動される
- `NodeClass`、`NodePool`、`NodeClaim`はKarpenterの概念

## NodeClass
- k8sの`NodeClass`リソースで管理
  - `kubectl get nodeclass`でリソースを確認

## NodePool
- マネージドNodePoolとカスタムNodePoolがある
- EKS Auto Mode作成時、"*general-purpose*"と"*system*"うち、どれかは有効にする必要がある（カスタムNodePoolだけはできない）
- k8sの`NodePool`リソースで管理
  - `kubectl get nodepool`でリソースを確認
  - https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/create-node-pool.html

### マネージドNodePool
- "*general-purpose*"と"*system*"の２つが用意されている
- "*general-purpose*"は一般的なさまざまなワークフロー向け、"*system*"はCoreDNSなどクラスター運用にクリティカルなPodを配置するためのNode

### カスタムNodePool
- サンプル  
  ```yaml
  apiVersion: karpenter.sh/v1
  kind: NodePool
  metadata:
    name: default
  spec:
    template:
      metadata:
        labels:
          billing-team: my-team
      spec:
        nodeClassRef:
          group: eks.amazonaws.com
          kind: NodeClass
          name: default

        requirements:
          - key: "karpenter.sh/capacity-type"
            operator: In
            values: ["on-demand", "spot"]
          - key: "eks.amazonaws.com/instance-category"
            operator: In
            values: ["c", "m", "r"]
          - key: "eks.amazonaws.com/instance-cpu"
            operator: In
            values: ["4", "8", "16", "32"]
          - key: "topology.kubernetes.io/zone"
            operator: In
            values: ["us-west-2a", "us-west-2b"]
          - key: "kubernetes.io/arch"
            operator: In
            values: ["arm64", "amd64"]

    limits:
      cpu: "1000"
      memory: 1000Gi
  ```
- カスタムNodePoolにPodを配置するためには`nodeSelector`で指定する必要がある  
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    namespace: default
    name: nginx
  spec:
    selector:
      matchLabels:
        app.kubernetes.io/name: nginx
    replicas: 2
    template:
      metadata:
        labels:
          app.kubernetes.io/name: nginx
      spec:
        nodeSelector:
          karpenter.sh/nodepool: default # NodePoolの`metadata.name`
          karpenter.sh/capacity-type: spot # NodePoolの`karpenter.sh/capacity-type`を指定
        containers:
          - image: nginx:1.14.2
            imagePullPolicy: Always
            name: nginx
            ports:
              - containerPort: 80
            resources:
              requests:
                cpu: "0.5"
  ```

## Node Disruption Budgets


## Pod Disruption Budgets

---

# Storage
## `StorageClass`
- EKS Auto Modeではユーザが使うための`StorageClass`はデフォルトで作成されない。EKS Auto Modeのストレージ機能を使用するには、`ebs.csi.eks.amazonaws.com` を参照する`StorageClass`を作成する必要がある。  
  ```yaml
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: auto-ebs-sc
    annotations:
      storageclass.kubernetes.io/is-default-class: "true"
  provisioner: ebs.csi.eks.amazonaws.com
  volumeBindingMode: WaitForFirstConsumer
  parameters:
    type: gp3
    encrypted: "true"
  ```
- https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/sample-storage-workload.html

## Auto Modeで静的EBS PVを使うときの注意点
- https://neusstudio.blog/eks-auto-mode-static-ebs-troubleshooting/

1. `driver`を`ebs.csi.aws.com`ではなく、 **`ebs.csi.eks.amazonaws.com`** にする必要がある
    - https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/ebs-csi.html#ebs-csi-considerations
2. EBS Volumeに Key：`eks:eks-cluster-name`、Value：`<対象EKSクラスター名>`の設定が必要
    - http://repost.aws/questions/QUj5BAYq-xQgKmH-dkR5ajLQ/eks-automode-static-pv-can-t-be-attached-volume-attachment-is-being-deleted

---

# Network

## Prefix Delegation
- EKS Auto ModeではdefaultでPrefix Delegationを使う設定になっている
  - https://docs.aws.amazon.com/en_es/eks/latest/userguide/auto-networking.html  
    > EKS Auto Mode defaults to using prefix delegation (/28 prefixes) for pod networking and maintains a predefined warm pool of IP resources that scales based on the number of scheduled pods. 

- Prefix Delegationの仕組み  
  ```
  ┌─────────────────────────────────────────────────────────────┐
  │  Worker Node (EKS Auto Mode)                                │
  │                                                             │
  │  ENI                                                        │
  │  ├─ Primary IP: 10.0.1.10 (ノード自身)                        │
  │  └─ Prefix: 10.0.1.64/28（16個のIP: 10.0.1.64〜10.0.1.79）    │
  │       │                                                     │
  │       ├─ Pod A: 10.0.1.64                                   │
  │       ├─ Pod B: 10.0.1.65                                   │
  │       └─ Pod C: 10.0.1.66                                   │
  └─────────────────────────────────────────────────────────────┘
  ```

- ENIには10個のスロット制限があり、従来のセカンダリ(Secondary)IP方式では、1つのENIに最大10個のPodしか配置できなかったが、Prefix Delegation方式では/28プレフィックスが割り当てられるため、1つのENIに最大16個?のPodを配置できるようになる
- ただ、Prefix Delegation方式では、最初から/28プレフィックスが確保されるため、Pod数が少ない場合は逆にIPアドレスの無駄遣いになる可能性がある

---

# `Ingress`周りについて
- 参考URL
  - https://zenn.dev/hanabusashun/articles/43572ae6e15366
- EKS Auto Modeで`IngressClassParams`、`IngressClass`を作成して、`Ingress`リソースを作成したら、ALBとリスナー、ターゲットグループまで作成してくれて、`TargetGroupBinding`も作成されてターゲットグループでのターゲットの登録までやってくれる。
  - https://docs.aws.amazon.com/eks/latest/userguide/auto-configure-alb.html

## `TargetGroupBinding`
- すでに作成されているALB、ターゲットグループとk8sの`Service`を紐づけるリソース
- EKS Auto Modeでは、リスナーとターゲットグループに以下のタグが設定されている必要がある（https://github.com/aws/containers-roadmap/issues/2508）
  - **ターゲットグループ**
    - `eks:eks-cluster-name`
      - EKSクラスター名
- ターゲットはPodのIPでマッピングされる
  - Podが再作成されたりしてPodのIPが変わると、自動的にターゲットグループのターゲットのマッピングも変更される
- 参考URL
  - https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/guide/targetgroupbinding/spec/
  - https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/guide/targetgroupbinding/targetgroupbinding/
- マニフェストファイル例  
  ```yaml
  apiVersion: eks.amazonaws.com/v1
  kind: TargetGroupBinding
  metadata:
    name: argocd-target-group-binding
    namespace: argocd
    labels:
      ingress.eks.amazonaws.com/stack-name: ingress-argocd
      ingress.eks.amazonaws.com/stack-namespace: argocd
  spec:
    networking:
      ingress:
      - from:
        - securityGroup:
            groupID: sg-0542a085asdasa # ターゲットグループへのアクセスを許可するセキュリティグループ（ロードバランサーに設定されているSGを指定）
        ports:
        - port: 8080 # Target ServiceのtargetPort
          protocol: TCP
    serviceRef:
      name: argocd-server # route traffic to the awesome-service
      port: 80 # Target Serviceのport
    targetGroupARN: arn:aws:elasticloadbalancing:ap-northeast-1:123456789:targetgroup/argocd-tg/abcdefghijklmnop
    targetType: ip
  ```

---

# Add-ons
- https://docs.aws.amazon.com/eks/latest/userguide/eks-add-ons.html#addon-consider-auto
- https://docs.aws.amazon.com/eks/latest/userguide/workloads-add-ons-available-eks.html
- EKS Auto Modeでは以下のAdd-onはインストール不要
  - VPC CNI
  - kube-proxy
  - CoreDNS
  - EBS CNI Dribe
  - EKS Pod Identity Agent

> [!IMPORTANT] 
> EKS Auto Modeでは上記Add-onは各ワーカーノード上にsystemdプロセスとして配置されるため、`kube-system` Namespace上にPodとしては存在しない 

---

# その他Auto Modeの注意点
## CoreDNS、VPC CNI、EBS CSI driverなどのAdd-onの配置場所
- 通常のKubernetesではCoreDNSやCSI Driverなどは`kube-system` NamespaceにPodとして配置されるが、**EKS Auto Modeでは各ワーカーノード上にsystemdプロセスとして配置される**
  - https://medium.com/@gajaoncloud/simplify-kubernetes-management-with-amazon-eks-auto-mode-d26650bfc239

> [!TIP]  
> CoreDNSのログを確認するためには以下のコマンドで対象ワーカーノード上にdebugコンテナを起動して、そこに入ってjournalctlコマンドで確認する必要がある
> ```bash
> kubectl debug node/< エラーが発生した Pod が起動するワーカーノード名 > -it --profile=sysadmin --image=public.ecr.aws/amazonlinux/amazonlinux:2023
> # CoreDNS のログの確認
> bash-5.2# chroot /host journalctl -u coredns
> ```
> Systemdなので、DaemonSetのAlloyの`loki.source.journal
`を使ってLokiにログを送ることもできる（はず）  
> https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.journal/

## Hop Limit
- Auto ModeではワーカーノードのHop Limitを変更することはできない
  - https://github.com/aws/containers-roadmap/issues/2498
- なのでPod Identityを使う必要がある（Pod Identityを使うとIMDSを使わないため）
  - https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/pod-id-association.html

## LokiのHelmチャートでのインストールについて
- LokiをDistributedモードでHelmチャートからデプロイする場合、gatewayというNginxのPodが立ち上がる。**その中で`kube-system` Namespace上の`kube-dns`のServiceを`kube-dns.kube-system.svc.cluster.local.`として指定している。**  
  しかし、**EKS Auto Modeでは`kube-system` Namespace上にCoreDNSは作成されないため、デプロイが失敗する。**  
  回避策として、**以下の`kube-dns` Serviceだけデプロイしておけば解消される（CoreDNSのPod自体は不要）**
- `kube-dns` Serviceのマニフェストファイル  
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: kube-dns
    namespace: kube-system
    labels:
      k8s-app: kube-dns
      kubernetes.io/cluster-service: "true"
      kubernetes.io/name: "CoreDNS"
      eks.amazonaws.com/component: kube-dns
  spec:
    selector:
      k8s-app: kube-dns
    clusterIP: 172.20.0.10
    ports:
    - name: dns
      port: 53
      protocol: UDP
    - name: dns-tcp
      port: 53
      protocol: TCP
    - name: metrics
      port: 9153
      protocol: TCP
  ```