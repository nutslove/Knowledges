# App of Apps Pattern
- 親となるApplicationが複数の子Applicationを管理
- 1つの親Applicationをデプロイすることで、複数の子Applicationも同時にデプロイされる

# ApplicationSet
- https://argo-cd.readthedocs.io/en/latest/user-guide/application-set/
- Generator（Git、Cluster、Listなど）を使い、パラメータに応じて`Application`リソースを動的に自動生成・管理
- 例  
  ```yaml
  apiVersion: argoproj.io/v1alpha1
  kind: ApplicationSet
  metadata:
    name: guestbook
  spec:
    goTemplate: true
    goTemplateOptions: ["missingkey=error"]
    generators:
    - list:
        elements:
        - cluster: engineering-dev
          url: https://1.2.3.4
        - cluster: engineering-prod
          url: https://2.4.6.8
        - cluster: finance-preprod
          url: https://9.8.7.6
    template:
      metadata:
        name: '{{.cluster}}-guestbook'
      spec:
        project: my-project
        source:
          repoURL: https://github.com/infra-team/cluster-deployments.git
          targetRevision: HEAD
          path: guestbook/{{.cluster}}
        destination:
          server: '{{.url}}'
          namespace: guestbook
  ```
