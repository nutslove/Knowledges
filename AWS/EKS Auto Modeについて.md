## `Ingress`周りについて
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