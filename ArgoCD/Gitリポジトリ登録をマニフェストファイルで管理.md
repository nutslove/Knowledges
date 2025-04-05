- Gitリポジトリの登録はGUIだけではなく、マニフェストファイル(`Secret`リソース)で管理することもできる
- `metadata.labels`フィールドに`argocd.argoproj.io/secret-type: repository`を指定
- 例
  - https://argo-cd.readthedocs.io/en/stable/operator-manual/argocd-repositories-yaml/  
    ```yaml
    # Git repositories configure Argo CD with (optional).
    # This list is updated when configuring/removing repos from the UI/CLI
    # Note: the last example in the list would use a repository credential template, configured under "argocd-repo-creds.yaml".
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-private-https-repo
      namespace: argocd
      labels:
        argocd.argoproj.io/secret-type: repository
    stringData:
      url: https://github.com/argoproj/argocd-example-apps
      password: my-password
      username: my-username
      insecure: "true" # Ignore validity of server's TLS certificate. Defaults to "false"
      forceHttpBasicAuth: "true" # Skip auth method negotiation and force usage of HTTP basic auth. Defaults to "false"
      enableLfs: "true" # Enable git-lfs for this repository. Defaults to "false"
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-private-ssh-repo
      namespace: argocd
      labels:
        argocd.argoproj.io/secret-type: repository
    stringData:
      url: ssh://git@github.com/argoproj/argocd-example-apps
      sshPrivateKey: |
        -----BEGIN OPENSSH PRIVATE KEY-----
        ...
        -----END OPENSSH PRIVATE KEY-----
      insecure: "true" # Do not perform a host key check for the server. Defaults to "false"
      enableLfs: "true" # Enable git-lfs for this repository. Defaults to "false"
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: istio-helm-repo
      namespace: argocd
      labels:
        argocd.argoproj.io/secret-type: repository
    stringData:
      url: https://storage.googleapis.com/istio-prerelease/daily-build/master-latest-daily/charts
      name: istio.io
      type: helm
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: private-helm-repo
      namespace: argocd
      labels:
        argocd.argoproj.io/secret-type: repository
    stringData:
      url: https://my-private-chart-repo.internal
      name: private-repo
      type: helm
      password: my-password
      username: my-username
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: private-repo
      namespace: argocd
      labels:
        argocd.argoproj.io/secret-type: repository
    stringData:
      url: https://github.com/argoproj/private-repo
    ```

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: github-repo
      namespace: argocd
      labels:
        argocd.argoproj.io/secret-type: repository
    stringData:
      password: my-password
      username: my-username
      proxy: http://10.10.10.10:3128
      type: git
      url: https://github.com/argocd-example-apps.git
    ```

## ESOの`ExternalSecrets`でも登録できる
- 例  
  ```yaml
  apiVersion: external-secrets.io/v1beta1
  kind: ExternalSecret
  metadata:
    name: lee-repo
    namespace: argocd
  spec:
    refreshInterval: 1h
    secretStoreRef:
      name: aws-secrets-manager
      kind: ClusterSecretStore
    target:
      name: lee-repo
      creationPolicy: Owner
      template:
        engineVersion: v2
        templateFrom:
        - target: Labels
          literal: "argocd.argoproj.io/secret-type: repository"
        data:
          url: https://github.com/nutslove/IaC.git
          insecure: "false"
          username: "{{ .username }}"
          password: "{{ .password }}"
    data:
    - secretKey: username
      remoteRef:
        key: argocd-repo-creds
        property: username
    - secretKey: password
      remoteRef:
        key: argocd-repo-creds
        property: password
  ```