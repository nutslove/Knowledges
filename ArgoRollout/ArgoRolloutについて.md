- https://argo-rollouts.readthedocs.io/en/stable/

## Argo Rolloutsインストール
- https://argoproj.github.io/argo-rollouts/installation/

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