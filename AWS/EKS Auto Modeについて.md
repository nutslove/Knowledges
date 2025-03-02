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

# `Ingress`周りについて
- 参考URL
  - https://zenn.dev/hanabusashun/articles/43572ae6e15366
- EKS Auto Modeで`IngressClassParams`、`IngressClass`を作成して、`Ingress`リソースを作成したら、ALBとリスナー、ターゲットグループまで作成してくれて、`TargetGroupBinding`も作成されてターゲットグループでのターゲットの登録までやってくれる。
  - https://docs.aws.amazon.com/eks/latest/userguide/auto-configure-alb.html

## `TargetGroupBinding`
- すでに作成されているALB、ターゲットグループとk8sの`Service`を紐づけるリソース
- EKS Auto Modeでは、リスナーとターゲットグループに以下のタグが設定されている必要がある
  - **リスナー**
    - `eks:eks-cluster-name`
      - EKSクラスター名
    - `ingress.eks.amazonaws.com/resource`
      - ターゲットグループのポート？要確認
    - `ingress.eks.amazonaws.com/stack`
      - "<Ingressのnamespace>/<Ingress名>"？要確認
  - **ターゲットグループ**
    - `eks:eks-cluster-name`
      - EKSクラスター名
    - `ingress.eks.amazonaws.com/resource`
      - "<Ingressのnamespace>/<Ingress名>-<ターゲットService名>:<ターゲットServiceのport>"？要確認
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
        - port: 8080 # Target ServiceのtargetPort？
          protocol: TCP
    serviceRef:
      name: argocd-server # route traffic to the awesome-service
      port: 80 # Target Serviceのport
    targetGroupARN: arn:aws:elasticloadbalancing:ap-northeast-1:123456789:targetgroup/argocd-tg/abcdefghijklmnop
    targetType: ip
  ```