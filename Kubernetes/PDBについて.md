# PDB（Pod Disruption Budget）とは
- Voluntary Disruptions（意図的・自発的な中断）からアプリケーションの可用性を保護するKubernetesのリソース
- Involuntary Disruption（ノード障害、OOMKillなど）には効果なし
- Voluntary Disruption の具体例
  - `kubectl drain` でのNodeからのPod退避
  - Nodeスケールダウン時のPod退避
  - Node間のPodの移動
- 設定方法
  - `minAvailable`: 常に稼働していなければならない最小Pod数（or割合）
  - `maxUnavailable`: 同時に停止してよい最大Pod数（or割合）
  - `minAvailable`と`maxUnavailable`は**同時に指定不可**（どちらか一方のみ）
- **Eviction APIを経由する操作にのみ機能する**（`kubectl delete pod`の直接実行には効かない）
- **namespace scoped**なので、namespace単位で管理が必要

- 参考URL
  - https://engineering.mercari.com/blog/entry/20231204-k8s-understanding-pdb/
  - https://kubernetes.io/ja/docs/tasks/run-application/configure-pdb/

## パーセント指定の注意点
- `minAvailable`/`maxUnavailable` はパーセント指定も可能で、**両方とも切り上げ（ceiling）**
  - 例：Pod数が7で `minAvailable: 50%` → `7 × 0.5 = 3.5` → 切り上げで **4**
  - 例：Pod数が3で `maxUnavailable: 50%` → `3 × 0.5 = 1.5` → 切り上げで **2**
  - `maxUnavailable`の切り上げは「停止を許容するPod数が増える」方向なので、意図より保護が緩くなる可能性があるため注意

## 設定値の注意点
- `minAvailable: 100%` や `maxUnavailable: 0` を設定すると `kubectl drain` が**完全にブロック**される
  - 特にReplicaが1のDeploymentに `maxUnavailable: 0` を設定すると、drainが永遠に終わらなくなる
- StatefulSetに適用する際は、PDBのselectorと`matchLabels`が正しく対応しているか確認が必要

## unhealthyPodEvictionPolicy（Kubernetes 1.31でstable・デフォルト有効）
- デフォルト（`IfHealthyBudget`）：UnhealthyなPodは、**currentHealthy（Ready状態のPod数）がdesiredHealthy以上を満たしている場合にのみ**退避可能。Budgetが満たされていなければ退避不可
- `AlwaysAllow`：UnhealthyなPodはBudgetの状態にかかわらず常に退避可能（CrashLoopBackOffなど異常Podによるdrainブロックを防ぎたい場合に有用）
- 障害時にdrainが詰まるケースへの対処として有用