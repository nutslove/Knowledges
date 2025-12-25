## `--cascade`オプションについて
- 参考URL
  - https://kubernetes.io/ja/docs/concepts/architecture/garbage-collection/
  - https://v1-33.docs.kubernetes.io/ja/docs/tasks/administer-cluster/use-cascading-deletion/
  - https://kubernetes.io/docs/tasks/run-application/delete-stateful-set/

- `kubectl delete`コマンドでリソースを削除する際に、依存する子リソース（例：`Deployment`の`ReplicaSet`や`Pod`）をどのように扱うかを制御するオプション
- `--cascade`オプションには以下の3つのモードがある
  - `--cascade=background`（デフォルト）: 親リソースを即座に削除し、子リソースはバックグラウンドでGarbage Collectorによって削除される
  - `--cascade=foreground`: 子リソースが完全に削除されるまで待ってから、親リソースを削除
  - `--cascade=orphan`: 親リソースを削除すると、依存する子リソースは親から切り離されて孤立状態になるが、削除はされない  
    > **When deleting a StatefulSet through `kubectl`, the StatefulSet scales down to 0. All Pods that are part of this workload are also deleted. If you want to delete only the StatefulSet and not the Pods, use `--cascade=orphan`.**

### どういう時に使うか
- 例えば、`StatefulSet`の`volumeClaimTemplates.spec.resources.requests.storage`の値を変更したい場合、そのまま変更すると、以下のようなエラーが出る。  
  ```shell
  one or more objects failed to apply, reason: error when patching "/dev/shm/2670892989": StatefulSet.apps "thanos-ingesting-receiver" is invalid: spec: Forbidden: updates to statefulset spec for fields other than 'replicas', 'ordinals', 'template', 'updateStrategy', 'revisionHistoryLimit', 'persistentVolumeClaimRetentionPolicy' and 'minReadySeconds' are forbidden
  ```  
  そういう場合に、`--cascade=orphan`オプションを使って`StatefulSet`を削除し、`PVC`と`Pod`を残したままにしておき、その後に再度`StatefulSet`を作成することで、`PVC`の容量を拡張できる。

> [!NOTE]  
> 上記の例は`StorageClass`で`allowVolumeExpansion: true`に設定されていて、StatefulSetの再作成の前に`kubectl patch`コマンドでPVCの容量を拡張する必要がある。