# Install
## Flux CLI install
- https://fluxcd.io/flux/installation/#install-the-flux-cli

```shell
curl -s https://fluxcd.io/install.sh | sudo bash
```

## Flux install
- https://fluxcd.io/flux/get-started/

```shell
export GITHUB_TOKEN=<your-token>
export GITHUB_USER=<your-username> # https://github.com/<ここに入るもの>

flux bootstrap github \
  --owner=$GITHUB_USER \
  --repository=<fluxマニフェストファイルを置くGitリポジトリ> \
  --branch=main \
  --path=./<fluxマニフェストファイルを置くディレクトリ> \
  --personal --token-auth --verbose
```
- **`--repository`で指定したGitリポジトリの`--path`で指定したディレクトリに`flux-system`というディレクトリが作成され、その配下に以下のファイルが生成される**
  - `gotk-components.yaml`
  - `gotk-sync.yaml`
  - `kustomization.yaml`

- 例  
  ```shell
  flux bootstrap github --owner=$GITHUB_USER \
  --repository=IaC --branch=main \
  --path=./AWS/k8s/flux --personal --token-auth --verbose
  ```

- 以下のDeployment（Pod）が`flux-system` namespace上に作成される
  - `helm-controller`
  - `kustomize-controller`
  - `notification-controller`
  - `source-controller`

> [!NOTE]
> Github token情報は`flux-system` namespace上に`flux-system`という名前で作成される

### Flux uninstall
```
flux uninstall
```

## `GitRepository`リソース
- https://fluxcd.io/flux/components/source/gitrepositories/
- Gitリポジトリの接続に使うTokenは`spec.secretRef.name`で明示的にSecretリソースを指定できる
  ```yaml
  apiVersion: source.toolkit.fluxcd.io/v1
  kind: GitRepository
  metadata:
    name: pack-local-test
    namespace: flux-system
  spec:
    interval: 30s
    url: https://github.com/xxx/xxx.git
    secretRef:
      name: flux-system # ここ
    ref:
      branch: lee-flux-test
  ```
