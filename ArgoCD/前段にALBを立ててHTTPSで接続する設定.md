## 前提
- TLSの終端はALB側でして、ArgoCD側は80のHTTPで受信する

## ArgoCD側設定
- https://techstep.hatenablog.com/entry/2020/11/15/121503
- ArgoCD serverを`--insecure`で起動する必要がある  
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    namespace: argocd
    labels:
      app.kubernetes.io/component: server
      app.kubernetes.io/name: argocd-server
      app.kubernetes.io/part-of: argocd
    name: argocd-server
  spec:
    replicas: 2
    selector:
      matchLabels:
        app.kubernetes.io/name: argocd-server
    template:
      metadata:
        labels:
          app.kubernetes.io/name: argocd-server
      spec:
        nodeSelector:
          karpenter.sh/nodepool: arm64-nodepool
          karpenter.sh/capacity-type: on-demand
          kubernetes.io/os: linux
        affinity:
          podAntiAffinity:
            preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/name: argocd-server
                topologyKey: topology.kubernetes.io/zone
              weight: 100
            requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchLabels:
                  app.kubernetes.io/name: argocd-server
              topologyKey: kubernetes.io/hostname
        containers:
        - args:
          - /usr/local/bin/argocd-server
          - --insecure ## ここ！
                
                ・・・中略・・・
  ```