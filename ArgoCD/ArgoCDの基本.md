### Install
- https://github.com/argoproj/argo-cd/blob/master/manifests/ha/install.yaml はnamespaceの指定(argocd)がなかったり、GUIのための`argocd-server`の`Service`がClusterIPになっていたりして、修正が必要
- HAの構成で必要なcontainer imageは以下４つ
  - `quay.io/argoproj/argocd:<version>`
  - `ghcr.io/dexidp/dex:<version>`
  - `haproxy:<version>`
  - `redis:<version>`
- 参考URL
  - https://argo-cd.readthedocs.io/en/stable/getting_started/
  - https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/

### 最初に払い出されるadminのPW確認方法
- `kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d`
- ログイン後GUI上でパスワード変更ができる

## 構成
- https://hiroki-hasegawa.hatenablog.jp/entry/2023/05/02/145115?utm_source=pocket_reader

## HelmチャートのApplications登録方法
- `index.yaml`と`*.tgz`があるHelmリポジトリを使う方法と、`charts`ディレクトリや`Chart.yaml`があるGitリポジトリを使う方法がある

### `repoURL`に`index.yaml`と`*.tgz`があるWebサーバ/Object Storageのエンドポイントを指定する場合
- https://argo-cd.readthedocs.io/en/stable/user-guide/helm/
- argocdバージョン2.6前までは、values fileは必ずHelm chartと同じgit repositoryに存在している必要があったが、2.6からは別のrepository上のvalues fileも扱えるようになった。  
  > Before v2.6 of Argo CD, Values files must be in the same git repository as the Helm chart. The files can be in a different location in which case it can be accessed using a relative path relative to the root directory of the Helm chart. As of v2.6, values files can be sourced from a separate repository than the Helm chart by taking advantage of multiple sources for Applications.
- HelmチャートのApplication設定例
  ~~~yaml
  apiVersion: argoproj.io/v1alpha1
  kind: Application
  metadata:
    name: sealed-secrets
    namespace: argocd
  spec:
    project: default
    source:
      chart: sealed-secrets ---> Helm chart名
      repoURL: https://bitnami-labs.github.io/sealed-secrets
      targetRevision: 1.16.1 --> 使用するHelm Chartバージョン(「helm search repo <Helm chart名> --versions」で確認可能)
      helm:
        releaseName: sealed-secrets ----> Helmリリース名
        valueFiles: 
        - values-production.yaml ★-----> spec.source.helm.valueFiles配下でリスト形式でvalues fileを指定
    destination:
      server: "https://kubernetes.default.svc"
      namespace: kubeseal
  ~~~

> [!NOTE]  
> - `repoURL`に指定するのは`charts`ディレクトリや`Chart.yaml`等があるgitリポジトリではなく、**`index.yaml`と`*.tgz`ファイルがあるWebサーバやObject Storageのエンドポイント**
> - `chart`には`index.yaml`の`entries`の下の階層のチャート名を指定
> - `targetRevision`には`index.yaml`の`version`の部分のチャートVersionを指定

#### **Helm ChartとValues fileが別々のGit Repositoryにある場合の設定方法**
- https://argo-cd.readthedocs.io/en/stable/user-guide/multiple_sources/#helm-value-files-from-external-git-repository
- これで3rd partyのHelm Chart(e.g. Loki Helm chart)と自組織Git Repository上のvalues fileを組み合わせることができる
- 設定例
  ~~~yaml
  apiVersion: argoproj.io/v1alpha1
  kind: Application
  spec:
    sources:
    - repoURL: 'https://prometheus-community.github.io/helm-charts'
      chart: prometheus
      targetRevision: 15.7.1 ★--->helmの場合はChartバージョン
      helm:
        valueFiles:
        - $values/charts/prometheus/values.yaml ★---> 「$values」には下のGit Repositoryのrootが入る
    - repoURL: 'https://git.example.com/org/value-files.git'
      targetRevision: dev ★---> gitの場合はgit repositoryのブランチ
      ref: values
  ~~~  
  > In the above example, the prometheus chart will use the value file from git.example.gom/org/value-files.git. \$values resolves to the root of the value-files repository. The $values variable may only be specified at the beginning of the value file path.

### `repoURL`にGitリポジトリを指定する場合
- **`targetRevision`にはGitリポジトリの(1)Commit ID、(2)ブランチ名、(3)tag名のいずれを指定する**
- **Gitリポジトリの場合は、`chart`の代わりに`path`パラメータを使用。`path`にはGitリポジトリのトップディレクトリから`charts`ディレクトリへの相対パスを指定する**
- 設定例  
  ```yaml
  ---
  apiVersion: argoproj.io/v1alpha1
  kind: Application
  metadata:
    name: 'opensearch'
    namespace: openshift-gitops
    finalizers:
      - resources-finalizer.argocd.argoproj.io
  spec:
    destination:
      namespace: opensearch
      server: https://kubernetes.default.svc
    ignoreDifferences:
    - jsonPointers:
      - /spec/clusterIP
      kind: Service
    project: default
    source:
      helm:
        values: |
          clusterName: 'mship3'
          nodeGroup: client
          roles:
            - ingest
            　　・
            　　・
            　　・
        releaseName: 'opensearch-client'
      path: charts/opensearch
      repoURL: https://somegitrepository/opensearch-helm-charts.git
      targetRevision: 'opensearch-2.14.1'
    syncPolicy:
      automated:
        prune: true
  ```

### values fileを設けず、直接`Application`の中でvalueを定義することもできる
- https://argo-cd.readthedocs.io/en/stable/user-guide/helm/#values
- 設定例(1)
  ~~~yaml
  source:
    helm:
      valuesObject:
        ingress:
          enabled: true
          path: /
          hosts:
            - mydomain.example.com
          annotations:
            kubernetes.io/ingress.class: nginx
            kubernetes.io/tls-acme: "true"
          labels: {}
          tls:
            - secretName: mydomain-tls
              hosts:
                - mydomain.example.com
  ~~~
- 設定例(2)
  ~~~yaml
  source:
    helm:
      values: |
        ingress:
          enabled: true
          path: /
          hosts:
            - mydomain.example.com
          annotations:
            kubernetes.io/ingress.class: nginx
            kubernetes.io/tls-acme: "true"
          labels: {}
          tls:
            - secretName: mydomain-tls
              hosts:
                - mydomain.example.com
  ~~~

## その他
- `targetRevision: HEAD`はデフォルトのブランチ(masterまたはmain)を意味する