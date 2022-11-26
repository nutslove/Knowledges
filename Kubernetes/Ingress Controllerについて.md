#### rewrite-target
- defaultでは`spec.rules.http.paths.path`のパスまで後ろのPodに連携される  
  しかし、アプリPodはそのパスを持っていない(rootパスになっている)場合404エラーになるので、そういう場合は`annotations.nginx.ingress.kubernetes.io/rewrite-target`で`spec.rules.http.paths.path`のパスをRewriteする
- 参考URL
  - https://kubernetes.io/docs/concepts/services-networking/ingress/
  - https://kubernetes.github.io/ingress-nginx/examples/rewrite/
- 例
    ~~~yaml
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: test-ingress
      namespace: critical-space
      annotations:
        nginx.ingress.kubernetes.io/rewrite-target: /
    spec:
      rules:
      - http:
          paths:
          - path: /pay
            backend:
              serviceName: pay-service
              servicePort: 8282
    ~~~