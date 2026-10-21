# Kustomizeとは
- Kubernetesマニフェストを**テンプレートを使わずに**環境ごとにカスタマイズするためのツール
  - Helmのように`{{ }}`で変数埋め込みをするのではなく、**素のYAML（base）に対して差分（patch）を重ねる**アプローチ
- `kubectl`に組み込まれている（`kubectl apply -k`、`kubectl kustomize`）。スタンドアロンの`kustomize`バイナリもある
  - **kubectl内蔵版はスタンドアロン版よりバージョンが古いことが多い**ので、新しい機能（`replacements`や`helmCharts`の一部オプションなど）を使う場合は`kustomize version`と`kubectl version --client`で確認する
- 設定は`kustomization.yaml`（`kustomization.yml`、`Kustomization`でも可）に書く
- 各リソースのYAMLは**そのままkubectl applyできる有効なYAMLのまま**にしておける（これがHelmとの大きな違い）

## Helmとの違い
| | Kustomize | Helm |
|---|---|---|
| 方式 | 素のYAML + overlay/patch | Goテンプレート + values |
| 変数 | 基本なし（`replacements`で代用） | `{{ .Values.xxx }}` |
| 配布 | Gitのディレクトリ/URL | Chartリポジトリ(OCI含む) |
| リリース管理 | なし（apply/pruneのみ） | あり（`helm list`、rollback） |
| 向いている用途 | 自社マニフェストの環境差分管理 | サードパーティ製品の配布・インストール |

- 両者は排他ではなく、`helmCharts`フィールドでKustomizeからHelm Chartを展開したり、`helm template`の出力をbaseにしてKustomizeでpatchする運用もよくある

## インストール
```bash
# Mac
brew install kustomize

# インストールスクリプト（カレントディレクトリにバイナリが置かれる）
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash

# Go
go install sigs.k8s.io/kustomize/kustomize/v5@latest
```

# 基本的なコマンド
```bash
# kustomization.yamlがあるディレクトリを指定して、最終的なマニフェストを標準出力に出す（applyはしない）
kustomize build <ディレクトリ>
kubectl kustomize <ディレクトリ>     # 同じ（kubectl内蔵版）

# 適用
kubectl apply -k <ディレクトリ>
kustomize build <ディレクトリ> | kubectl apply -f -

# 差分確認（適用前に必ずやる）
kubectl diff -k <ディレクトリ>

# 削除
kubectl delete -k <ディレクトリ>

# 確認・デバッグ用に出力をファイルに保存
kustomize build overlays/dev -o /tmp/dev.yaml     # ファイル指定
kustomize build overlays/dev -o /tmp/out          # 既存ディレクトリを指定すると1リソース1ファイルで出力
```

- **`kustomize build`は何も適用しない**。まずbuild結果を目視で確認する癖をつけるのが大事
- `kubectl apply -k`と`kubectl apply -f`は同時に指定できない
- `kubectl get -k <ディレクトリ>`のように`-k`は`get`/`delete`/`diff`などでも使える

## `kustomize edit`
- CLIから`kustomization.yaml`を書き換えるコマンド。CI/CDでイメージタグを更新するときによく使う
```bash
cd overlays/dev
kustomize edit set image myapp=123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/myapp:v1.2.3
kustomize edit set namespace dev
kustomize edit set nameprefix dev-
kustomize edit add resource ../../base
kustomize edit add patch --path patch.yaml --kind Deployment --name myapp
kustomize edit fix      # 非推奨フィールドを新しい書き方に変換（後述）
```

## `kustomize build`の主なオプション
| オプション | 説明 |
|---|---|
| `--enable-helm` | `helmCharts`フィールドを有効化（`helm`コマンドが必要） |
| `--load-restrictor LoadRestrictionsNone` | kustomization.yamlのあるディレクトリ配下以外のファイルを参照できるようにする |
| `--enable-alpha-plugins` | exec/containerized function等のプラグインを有効化 |
| `-o <path>` | 出力先 |

# ディレクトリ構成（base / overlays）
- 一般的なのは`base`（共通の定義）と`overlays`（環境ごとの差分）に分ける構成
```
my-app/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── configmap.yaml
└── overlays/
    ├── dev/
    │   ├── kustomization.yaml
    │   └── replica-patch.yaml
    ├── stg/
    │   └── kustomization.yaml
    └── prod/
        ├── kustomization.yaml
        ├── replica-patch.yaml
        └── resources-patch.yaml
```

## base
`base/kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml
```

`base/deployment.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: myapp
          image: myapp:latest
          ports:
            - containerPort: 8080
```

## overlay
`overlays/prod/kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: prod
namePrefix: prod-

resources:
  - ../../base          # baseディレクトリを参照（ディレクトリ or YAMLファイル or URL）

images:
  - name: myapp
    newName: 123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/myapp
    newTag: v1.2.3

replicas:
  - name: myapp
    count: 5

patches:
  - path: resources-patch.yaml
```

```bash
kustomize build overlays/prod
kubectl apply -k overlays/prod
```

- `resources`には**ファイル、ディレクトリ（その中のkustomization.yamlが評価される）、リモートURL**を書ける
  - 古いドキュメントに出てくる`bases:`は非推奨。`resources:`に統一する
- overlayからさらに別のoverlayをresourcesで参照して多段にすることも可能だが、深くなると追いにくいので2階層程度に留めるのが無難

# kustomization.yamlの主要フィールド

## resources
```yaml
resources:
  - deployment.yaml                 # ローカルファイル
  - ../base                         # ディレクトリ
  - https://github.com/org/repo//path/to/dir?ref=v1.0.0   # リモート(gitのタグ/ブランチ/コミットを固定する)
  - https://raw.githubusercontent.com/xxx/yyy/v1.0.0/install.yaml   # YAML直接
```
- リモートを参照する場合は**`?ref=`で必ずバージョンを固定**する（固定しないとデフォルトブランチの最新になり再現性がなくなる）
- 同じリソース（apiVersion/kind/name/namespaceが同じ）が複数のresourcesに重複するとエラー（`may not add resource with an already registered id`）になる

## namespace
```yaml
namespace: prod
```
- 全ての**namespaceスコープのリソース**のnamespaceを上書き（設定）する
- **Namespaceリソース自体は作成されない**ので、必要なら`resources`に`namespace.yaml`を追加する
- ClusterRoleBindingの`subjects[].namespace`など、参照しているnamespaceも更新される

## namePrefix / nameSuffix
```yaml
namePrefix: dev-
nameSuffix: -v1
```
- 全リソースの`metadata.name`に接頭辞/接尾辞を付ける
- `Service`名や`ConfigMap`/`Secret`の参照（`envFrom`、`volumes`、`serviceAccountName`など）は**自動で追従して書き換えられる**（nameReference機能）
- CRDなどKustomizeが参照関係を知らないフィールドは追従しない場合がある（`configurations`で設定可能）

## labels / commonLabels
```yaml
# 推奨
labels:
  - pairs:
      app.kubernetes.io/part-of: myapp
      env: prod
    includeSelectors: false     # デフォルトfalse。trueだとselector/Service.spec.selectorにも付与
    includeTemplates: true      # Pod templateのlabelsにも付与

# 非推奨（deprecated）
commonLabels:
  env: prod
```
- **`commonLabels`は`selector.matchLabels`にも付与される**ため、既存のDeployment/StatefulSetに後から追加するとselectorが変わり、`spec.selector: field is immutable`エラーになる
  - 新しい`labels`で`includeSelectors: false`（デフォルト）にしておけば、selectorには触らずlabelsだけ付けられる
  - `includeSelectors: true`にするとcommonLabels同様にselectorも書き換わるので注意

## commonAnnotations / annotations
```yaml
commonAnnotations:
  owner: platform-team
```
- 全リソース（`includeTemplates`相当でPod templateにも）に付与される

## images
```yaml
images:
  - name: myapp                  # マニフェスト内のimage名（タグ部分は除く）にマッチ
    newName: ecr.example.com/myapp
    newTag: v1.2.3
  - name: nginx
    digest: sha256:xxxxxxxx...   # digestで固定（newTagとは併用不可）
```
- Deployment/StatefulSet/DaemonSet/Job/CronJobなどのPod specの`containers`/`initContainers`のimageを書き換える
- baseのDeploymentは`image: myapp`（or `myapp:latest`）のままにして、overlayでタグを差し替えるのが定番
- CIから`kustomize edit set image`でタグを更新 → Gitにcommit → ArgoCDが同期、という流れが多い

## replicas
```yaml
replicas:
  - name: myapp
    count: 5
```
- Deployment/StatefulSet/ReplicationController/ReplicaSetの`spec.replicas`を上書きする
  - HPAを使っている場合はreplicasを固定するとHPAと競合するので注意

## patches
**差分を当てる最も重要な機能。** 2種類のpatch方式があり、どちらも`patches`フィールドで書く

### Strategic Merge Patch
- 変更したいリソースの**一部分だけを書いたYAML**を用意する。`apiVersion`/`kind`/`metadata.name`でどのリソースかが特定される
- listは`containers`のように**mergeKey（name）を持つものはnameでマージ**され、それ以外のlistは**丸ごと置き換え**になる

`overlays/prod/resources-patch.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp          # ← base側の名前（namePrefix適用前の名前）を書く
spec:
  template:
    spec:
      containers:
        - name: myapp  # ← nameでマージされる
          resources:
            requests:
              cpu: 500m
              memory: 512Mi
            limits:
              memory: 1Gi
```

```yaml
# kustomization.yaml
patches:
  - path: resources-patch.yaml
```

インラインでも書ける
```yaml
patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: myapp
      spec:
        replicas: 3
```

#### フィールド/リソースの削除
```yaml
# 特定のフィールドを削除: nullを指定
spec:
  template:
    spec:
      containers:
        - name: myapp
          resources: null

# list要素を削除
spec:
  template:
    spec:
      containers:
        - name: sidecar
          $patch: delete

# リソースごと削除
apiVersion: v1
kind: Service
metadata:
  name: myapp
$patch: delete
```

#### CRD（カスタムリソース）に対する注意
- Kustomizeは組み込みリソースのスキーマ（OpenAPI）を持っているため、listのマージキーを知っている。**CRDはスキーマを持たないため、listは常に丸ごと置き換えになる**
- CRDのlistを要素単位でマージしたい場合は、次のJSON6902を使うか、`openapi`フィールドでCRDのスキーマを読み込ませる

### JSON 6902 Patch
- RFC 6902形式で`op`（add/remove/replace/move/copy/test）と`path`を指定する。**対象リソースは`target`で明示する必要がある**
- listの特定要素を**indexで**指定できるのが強み（CRD、Strategic Mergeで表現しづらい箇所に有効）

```yaml
patches:
  - target:
      group: apps
      version: v1
      kind: Deployment
      name: myapp
    patch: |-
      - op: replace
        path: /spec/replicas
        value: 3
      - op: add
        path: /spec/template/spec/containers/0/env/-     # "-" はlistの末尾に追加
        value:
          name: LOG_LEVEL
          value: debug
      - op: remove
        path: /spec/template/spec/containers/0/args/1
```

ファイルを使う場合
```yaml
patches:
  - path: jsonpatch.yaml
    target:
      kind: Deployment
      name: myapp
```
- `patch`/`path`の中身が`- op:`で始まるlistならJSON6902、リソース形式のYAMLならStrategic Mergeとして**自動判別**される

### targetの指定方法
```yaml
patches:
  - path: patch.yaml
    target:
      group: apps                 # 省略可
      version: v1                 # 省略可
      kind: Deployment            # 正規表現可
      name: myapp-.*              # 正規表現可
      namespace: default
      labelSelector: "app=myapp,tier!=db"
      annotationSelector: "foo=bar"
```
- `target`を指定すれば、Strategic Mergeでも**複数のリソースにまとめて同じpatchを当てる**ことができる（その場合patch内のmetadata.nameは何でもよい／省略可）
- ※`patchesStrategicMerge`と`patchesJson6902`は非推奨。`patches`に統一する

### patchの注意点
- Strategic Mergeのpatchに書く`metadata.name`は、**namePrefix/nameSuffixが適用される前のbaseでの名前**
- patchで`name`を変更したい場合は`options: { allowNameChange: true }`が必要
- 適用順は`patches`に書いた順

## configMapGenerator / secretGenerator
- ファイル・リテラルからConfigMap/Secretを生成する
- **生成されたConfigMap/Secretの名前には内容のハッシュ値が付く**（例: `app-config-7h2k8d9fmc`）。それを参照しているDeploymentなどの参照も自動で書き換わるため、**ConfigMapの内容を変更すると名前が変わり、Podが自動でローリングアップデートされる**

```yaml
configMapGenerator:
  - name: app-config
    literals:
      - LOG_LEVEL=info
      - TZ=Asia/Tokyo
    files:
      - application.properties                # ファイル名がkeyになる
      - config.json=./conf/prod.json          # key=パス でkey名を指定
    envs:
      - app.env                               # KEY=VALUE形式のファイル（1行1エントリ）
    options:
      disableNameSuffixHash: false            # trueでハッシュ無し
      labels:
        app: myapp
      annotations:
        note: generated

secretGenerator:
  - name: app-secret
    type: Opaque
    literals:
      - DB_PASSWORD=changeme
    files:
      - tls.crt
      - tls.key
```

### behavior（baseのgeneratorをoverlayで上書きする）
```yaml
# overlays/prod/kustomization.yaml
configMapGenerator:
  - name: app-config        # baseに同名のgeneratorがある
    behavior: merge         # create(デフォルト) | replace | merge
    literals:
      - LOG_LEVEL=warn
```
- `create`: 新規作成（同名があるとエラー）／`replace`: 丸ごと置換／`merge`: keyごとにマージ（上書き）

### generatorOptions（全generator共通）
```yaml
generatorOptions:
  disableNameSuffixHash: true
  labels:
    type: generated
```

### 注意点
- **`secretGenerator`のliteralsに平文のパスワードをGit管理するのは避ける**
  - External Secrets Operator、SOPS（KSOPS）、Sealed Secrets等を使う
  - **SOPS/KSOPS を使う場合の例**:
    ```yaml
    # kustomization.yaml の generators フィールドで指定
    generators:
      - kustomizePlugin:
          apiVersion: viaduct.ai/v1
          kind: KSOPSGenerator
          name: app-secret
    ```
    KSOPSプラグインはkustomizeが`secrets.enc.yaml`などのSOPS暗号化ファイルをデプロイ時に復号化します。詳細は[KSOPS GitHub](https://github.com/viaduct-ai/kustomize-sops)参照
  - **External Secrets Operator (ESO) を使う場合の例**:
    ```yaml
    # kustomization.yaml
    resources:
      - secretstore.yaml              # SecretStore（credentials定義）
      - external-secret.yaml          # ExternalSecret（取得指定）
    ```
    ESOは外部Secret管理（HashiCorp Vault、AWS Secrets Manager、Azure Key Vault等）からSecretを定期的に同期します
- `disableNameSuffixHash: true`にすると、ConfigMap変更時にPodが自動再起動されなくなる（代わりに名前が固定される）
  - ArgoCDなどでConfigMapを他のリソース（CRD等）から**固定名で参照**する場合はこちらを使うことがある
- ハッシュ付きの名前を**Kustomizeが知らないフィールド**（CRDの任意フィールドなど）で参照している場合、参照が追従しないので`configurations`でnameReferenceを設定するか`disableNameSuffixHash`を使う

## replacements（変数置換）
- あるリソースのフィールドの値を、別のリソースのフィールドにコピーする。Kustomizeには`${VAR}`のような環境変数展開が**ない**ため、その代替
- 旧`vars`は非推奨で、`replacements`を使う

```yaml
# 例: ConfigMapのhost値を、DeploymentのENV値に埋め込む
replacements:
  - source:
      kind: ConfigMap
      name: app-config
      fieldPath: data.DB_HOST
    targets:
      - select:
          kind: Deployment
          name: myapp
        fieldPaths:
          - spec.template.spec.containers.[name=myapp].env.[name=DB_HOST].value
```

```yaml
# 例: 文字列の一部だけ置換（delimiter + index）
replacements:
  - source:
      kind: Service
      name: myapp
      fieldPath: metadata.name
    targets:
      - select:
          kind: Ingress
        fieldPaths:
          - spec.rules.0.host
        options:
          delimiter: '.'
          index: 0              # "xxx.example.com" の0番目("xxx")を置換
          create: true          # フィールドが無ければ作る
```
- `fieldPaths`のlist要素は`[name=myapp]`のように**キーで指定**できる（indexでも可）
- `reject:`で対象から除外するリソースを指定できる

## components
- **再利用可能な機能（オプション機能）の部品**を作り、必要なoverlayだけに足せる仕組み
  - 例: 「Prometheus用のannotation追加」「Istioサイドカー有効化」「ログ収集sidecar追加」など
- `kind: Component`として定義し、overlayの`components:`で読み込む

`components/monitoring/kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component

patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: myapp
      spec:
        template:
          metadata:
            annotations:
              prometheus.io/scrape: "true"
```

`overlays/prod/kustomization.yaml`
```yaml
resources:
  - ../../base
components:
  - ../../components/monitoring
```
- `resources`との違い: componentは**それを読み込む側のリソース群に対して**patchなどを適用できる（`resources`で読み込んだbaseは、baseの中身に変更を加えることはできない独立したもの扱い）

## helmCharts
- Kustomizeの中でHelm Chartを展開して、その出力に対してpatchを当てられる
- **`--enable-helm`が必須**で、`helm`コマンドがインストールされている必要がある
```yaml
helmCharts:
  - name: ingress-nginx
    repo: https://kubernetes.github.io/ingress-nginx
    version: 4.10.0
    releaseName: ingress-nginx
    namespace: ingress-nginx
    includeCRDs: true
    valuesInline:
      controller:
        replicaCount: 2
    # valuesFile: values.yaml
```
```bash
kustomize build --enable-helm overlays/prod
```
- `helm template`相当の出力になるため、Helmのリリース管理（`helm list`/rollback）は使えない
- ArgoCDで使う場合は`argocd-cm`の`kustomize.buildOptions: --enable-helm`が必要（後述）

## openapi / configurations / transformers / generators
| フィールド | 用途 |
|---|---|
| `openapi` | CRDのOpenAPIスキーマを読み込ませ、Strategic Mergeでlistのmergeキーを認識させる |
| `configurations` | nameReference/commonLabelsの対象フィールドなど、Kustomizeの変換ルールを拡張するYAMLファイルを指定 |
| `transformers` | 自作/組み込みのtransformerプラグインを指定 |
| `generators` | 自作/組み込みのgeneratorプラグインを指定 |
| `sortOptions` | build出力のリソース順（`legacy`/`fifo`）を指定 |

# 非推奨フィールドと移行
| 非推奨 | 移行先 |
|---|---|
| `bases` | `resources` |
| `patchesStrategicMerge` | `patches` |
| `patchesJson6902` | `patches`（`target`を指定） |
| `commonLabels` | `labels`（`includeSelectors`で挙動を選択） |
| `vars` | `replacements` |

- 古い書き方のkustomization.yamlは`kustomize edit fix`で自動変換できる

# ファイルの参照制限（セキュリティ）
- デフォルトでは、kustomization.yamlが参照できるファイルは**そのkustomization.yamlのあるディレクトリ配下のみ**
  - `../../base`のように**ディレクトリ（kustomization）として**親を参照するのはOK
  - `patches.path: ../common/patch.yaml`のように、**ディレクトリ外の個別ファイル**を参照するとエラー（`security; file ... is not in or below ...`）
- 回避方法: `--load-restrictor LoadRestrictionsNone`を付ける、または参照先のディレクトリにもkustomization.yamlを置いてresources/componentsで読み込む形にする（後者推奨）

# ArgoCDとの連携
- ArgoCDは**applicationのsource pathにkustomization.yamlがあれば自動でKustomizeとして認識**し、`kustomize build`した結果をデプロイする
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp-prod
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: main
    path: my-app/overlays/prod        # kustomization.yamlがあるディレクトリ
    kustomize:                         # 省略可。Application側からもKustomize設定を上書きできる
      namePrefix: ""
      images:
        - myapp=123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/myapp:v1.2.3
  destination:
    server: https://kubernetes.default.svc
    namespace: prod
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```
- `helmCharts`を使う場合は、`argocd-cm`に以下を設定する
```yaml
data:
  kustomize.buildOptions: --enable-helm
```
- ArgoCDに同梱されているkustomizeのバージョンは`argocd version`、またはrepo-serverのイメージで確認できる。**ローカルのバージョンと違うとbuild結果が変わることがある**ので注意

## イメージタグの更新方法
### 方法1: CI/CD で Gitにcommit（推奨 / GitOps的に正統）
```bash
# CI でのコマンド例（GitHub Actions等）
kustomize edit set image myapp=123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/myapp:${BUILD_TAG}
git add overlays/prod/kustomization.yaml
git commit -m "Update myapp image to ${BUILD_TAG}"
git push
```
- Gitが真実のソースになり、デプロイ履歴がcommit履歴で追跡可能
- ArgoCDがGit変更を検知してsyncするだけ（監視の仕組みが単純）

### 方法2: ArgoCD Image Updater（自動化）
- **ArgoCD Image Updaterとは**: ArgoCDのaddOnで、定期的にレジストリをスキャンして新しいイメージタグを発見し、kustomization.yamlの`images`を自動更新しGitにcommitする
- **Application側のアノテーション例**:
  ```yaml
  metadata:
    annotations:
      argocd-image-updater.argoproj.io/image-list: myapp          # 監視するイメージ
      argocd-image-updater.argoproj.io/myapp.update-strategy: semver  # セマンティックバージョニングで最新を選ぶ
      argocd-image-updater.argoproj.io/myapp.helm.image-name: image.name    # (Helm使用時)
      argocd-image-updater.argoproj.io/myapp.kustomize.image-name: myapp    # (Kustomize使用時)
  ```
- **argocd-image-updater ConfigMap の設定例**:
  ```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: argocd-image-updater-config
    namespace: argocd
  data:
    log.level: info
    registries.conf: |
      registries:
        - name: ecr
          api_url: https://123456789012.dkr.ecr.ap-northeast-1.amazonaws.com
          type: aws
          credentials: pullsecret:argocd/aws-ecr    # (AWS認証情報)
  ```
- メリット: タグ更新のためのCIパイプラインが不要、新しいイメージが自動デプロイ
- デメリット: Gitでのcommit履歴が自動化されるため、レビュー・承認フローが入りにくい（一般的に開発環境/staging向け）
- **使い分け**: 本番環境はGitによる手動commit → argocd-imageUpdaterは開発/staging環境向けがベストプラクティス

# 実践的なTips・ベストプラクティス
- **baseは「そのままapplyできる状態」にしておく**（最低限動くデフォルト値を持たせる）。環境差分だけをoverlayに書く
- overlayに書く差分は最小限にし、patchが肥大化したらcomponentsに切り出す
- **`kustomize build`の結果をレビューする**。PRのCIで`kustomize build`が通ること、スキーマ検証、`kubectl diff`で差分確認をする
  - **kubeconform での検証**（推奨。複数フォーマット対応）:
    ```bash
    kustomize build overlays/prod | kubeconform -kubernetes-version 1.29
    # KubernetesCRDカタログを使う場合
    kustomize build overlays/prod | kubeconform -schema-location default -schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceApiVersion}}.json'
    ```
  - **kubeval での検証**:
    ```bash
    kustomize build overlays/prod > /tmp/manifest.yaml
    kubeval /tmp/manifest.yaml --kubernetes-version 1.29 --strict
    ```
  - **kubectl apply --dry-run を使ったサーバー側検証**:
    ```bash
    kubectl apply -k overlays/prod --dry-run=server
    ```
- リモートbaseは`?ref=タグ`で固定する
- ラベルはselectorを壊さないよう`labels`（`includeSelectors: false`）を使う。既存リソースに`commonLabels`を後から足さない
- イメージは`images`で一元管理し、baseのimageはプレースホルダ名にする
- ConfigMapは`configMapGenerator`でハッシュ付きにすると、**設定変更→自動ローリング再起動**ができる
- Secretの平文をGitに置かない（SOPS/KSOPS、External Secrets Operator、Sealed Secretsなど）
- 変数展開が必要なとき、`envsubst`でkustomizeの出力を後処理する運用もあるが、まずは`replacements`で足りないか検討する

# よくあるエラーとトラブルシューティング
| エラー/症状 | 原因・対処 |
|---|---|
| `accumulating resources: ... no such file or directory` | resourcesのパス誤り。相対パスは**そのkustomization.yamlからの相対** |
| `may not add resource with an already registered id` | 同じリソースが複数のresources/overlayに重複している |
| `no matches for Id ...; failed to find unique target for patch` | patchのtarget（kind/name/namespace）がbaseにマッチしない。**namePrefix適用前の名前**で書いているか確認 |
| `field is immutable`（spec.selector） | `commonLabels`でselectorが変わった。`labels`に移行するか、リソースを再作成 |
| `security; file ... is not in or below ...` | ファイル参照制限。`--load-restrictor LoadRestrictionsNone`かディレクトリ構成の見直し |
| `must build with flag '--enable-helm'` | `helmCharts`を使っている。buildに`--enable-helm`を付ける（ArgoCDは`kustomize.buildOptions`） |
| patchが効かない（CRD） | CRDのlistはmergeキーが無く丸ごと置換になる、またはtargetが間違い。JSON6902か`openapi`を検討 |
| `Warning: 'bases' is deprecated` 等 | `kustomize edit fix`で新しい書き方に変換 |
| CIとローカルでbuild結果が違う | kustomizeのバージョン差。`kustomize version`を揃える |

# 参考
- 公式ドキュメント: https://kubectl.docs.kubernetes.io/references/kustomize/
- GitHub: https://github.com/kubernetes-sigs/kustomize
- Kustomization APIリファレンス: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/
- ArgoCD Kustomize: https://argo-cd.readthedocs.io/en/stable/user-guide/kustomize/

