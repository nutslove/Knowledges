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

## Terraform / Open Tofu Controller for Flux install
- https://flux-iac.github.io/tofu-controller/getting_started/

```shell
kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/docs/release.yaml
```
- Branch Plannerを有効にしてインストールする場合  
  ```shell
  kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/docs/branch-planner/release.yaml
  ```

> [!CAUTION]
> 2025/03/18 k8s 1.31 ver
> 上記のマニフェストファイルは、リソースのAPI versionが古くて警告が出るし、デプロイしても`flux-system` namespaceにtofu-controller PODは起動しない。
> [Getting Started](https://flux-iac.github.io/tofu-controller/getting_started/)の「Manual installation」のHelmでのインストール手順と https://github.com/flux-iac/tofu-controller/tree/main/charts/tofu-controller の手順も、`helm repo add tofu-controller https://flux-iac.github.io/tofu-controller`は合ってるけど、Chart名が`tofu-controller/tofu-controller`ではなく、**`tofu-controller/tf-controller`が正しい**  
> - 以下でインストールした
> ```shell
> helm repo add tofu-controller https://flux-iac.github.io/tofu-controller
> helm repo update
> helm install tofu-controller tofu-controller/tf-controller --namespace flux-system
> ```

### `Terraform`リソースを強制削除する方法
- https://flux-iac.github.io/tofu-controller/use-tf-controller/resource-deletion/
- コマンド例  
  ```shell
  kubectl patch terraforms.infra.contrib.fluxcd.io \
  -n stk helloworld \
  -p '{"metadata":{"finalizers":null}}' \
  --type=merge
  ```

## `flux-system` namespace削除方法
- https://github.com/fluxcd/terraform-provider-flux/issues/67
### 手順
- Dump the descriptor as JSON to a file  
  ```shell
  kubectl get namespace flux-system -o json > flux-system.json
  ```

- Edit flux-system.json and remove kubernetes from the finalizers array 
  - 修正前
    ```
          ・
          ・

    "spec": {
            "finalizers": [
                "kubernetes"
            ]
        },
          ・
          ・

    ```
  - 修正後
    ```
          ・
          ・

    "spec": {
            "finalizers": []
        },
          ・
          ・

    ```
- Executing our cleanup command  
  ```shell
  kubectl replace --raw "/api/v1/namespaces/flux-system/finalize" -f ./flux-system.json
  ```