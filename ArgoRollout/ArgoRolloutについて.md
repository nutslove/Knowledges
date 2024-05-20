- https://argo-rollouts.readthedocs.io/en/stable/

## Argo Rolloutsインストール
- **Openshiftには`Red Hat OpenShift GitOps`にArgo Rolloutsも含まれている**
  - https://docs.openshift.com/gitops/1.12/argo_rollouts/using-argo-rollouts-for-progressive-deployment-delivery.html
- https://argoproj.github.io/argo-rollouts/installation/
- デフォルトでは`argo-rollouts`namespaceにインストールされる
```shell
kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
```
- kubectl pluginもあり、必須ではないけどargo rolloutsをCLIで管理できるのでインストールしておく  
  ```shell
  curl -LO https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64
  chmod +x ./kubectl-argo-rollouts-linux-amd64
  sudo mv ./kubectl-argo-rollouts-linux-amd64 /usr/local/bin/kubectl-argo-rollouts
  ```
  - `kubectl argo rollouts <サブコマンド>`の形式で使える
    - 例： `kubectl argo rollouts version`
  - `oc`コマンドでも使える

## Argo Rolloutsアーキテクチャ
- https://argoproj.github.io/argo-rollouts/architecture/
![](./image/architecture.jpg)

### Argo Rollouts controller
- `Rollout`リソースを監視し、`Rollout`を定義した状態に収束させるコアコンポーネント

### `Rollout`リソース
- `Deployment`リソースを拡張したもので、`Deployment`リソースと同様に配下に`ReplicaSet`リソースを管理する。  
  `spec.strategy.canary`など`Rollout`リソース独自のフィールド(機能)を持っている
- `Rollout`リソースで利用できるすべてのフィールド
  - https://argoproj.github.io/argo-rollouts/features/specification/

## Argo Rollouts CLIチートシート
- rollouts一覧確認  
  ```shell
  kubectl/oc argo rollouts list rollouts
  ```
- 特定のrolloutの状態確認  
  ```shell
  kubectl/oc argo rollouts get rollouts <rollouts名>
  ```

## Argo Rollouts Dashboard
- 以下のCLIでport forwardingが必要  
  ```shell
  kubectl argo rollouts dashboard
  ```
  **※defaultではlocalhostからしかアクセスできなくて、外部からアクセスするためには追加の設定が必要**