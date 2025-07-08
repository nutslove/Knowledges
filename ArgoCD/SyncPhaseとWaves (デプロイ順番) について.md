- https://argo-cd.readthedocs.io/en/stable/user-guide/sync-waves/

> [!IMPORTANT]  
> Sync PhaseとSync Waveは、1つ(単一)の`Application`内でのリソースの同期を制御するための仕組みであり、複数の`Application`間の同期を制御するものではない。一つのApplicationの中で、そのApplicationが管理するKubernetesリソースの同期順序を制御するためのもの。

# Sync Phase
- `PreSync` → `Sync` → `PostSync` の3段階に分かれて実行される
  - `PreSync`: 主にセットアップ処理（データベースマイグレーション、設定の準備など）
  - `Sync`: 通常のアプリケーションリソースのデプロイ
  - `PostSync`: 後処理（テスト、通知、クリーンアップなど）
- `PreSync`の`Application`が全て同期される（Healthyになる）と、`Sync`フェーズが開始される
- `Sync`の`Application`が全て同期される（Healthyになる）と、`PostSync`フェーズが開始される
 
# Sync Waves
- `Sync Waves`は、`Application`が管理するKubernetesリソースの同期を順番に行うための仕組み
- `Application`そのものではなく、実際にデプロイされる`Job`などのリソースに対して、`sync-wave`というannotationを付与することで、同期の順番を制御できる
- 数値が小さい順に実行される（例: "0" → "1" → "2"）
- 例  
  ```yaml
  # Wave 1: Secret作成（PreSync）
  apiVersion: v1
  kind: Secret
  metadata:
    name: db-secret
    annotations:
      argocd.argoproj.io/hook: PreSync
      argocd.argoproj.io/sync-wave: "1"

  ---
  # Wave 2: DB初期化Job（PreSync）
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: db-init
    annotations:
      argocd.argoproj.io/hook: PreSync
      argocd.argoproj.io/sync-wave: "2"

  ---
  # Wave 0: アプリケーション本体（Sync - デフォルト）
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: app-deployment
  # sync-waveの指定なし = wave 0

  ---
  # Wave 1: ヘルスチェック（PostSync）
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: health-check
    annotations:
      argocd.argoproj.io/hook: PostSync
      argocd.argoproj.io/sync-wave: "1"
  ```