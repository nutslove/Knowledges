## Deployment 対 StatefulSet
- ___Deployment___

- ___StatefulSet___
  - StatefulSetはReplicaSetを作成せず、`<statefulset name>-<ordinal index>`のuniqueな名前のPodを作成する
  - StatefulSetは2つの観点で固定化することができる
    > StatefulSets provides to each pod in it two stable unique identities.
    > 1. the Network Identity enables us to assign the same DNS name to the pod regardless of the number of restarts.
    > 2. the Storage Identity remains the same.
  - StatefulSetはscaling(up/down)時に以下が保証される
    > - For a StatefulSet with N replicas, when Pods are being deployed, they are created sequentially, in order from {0..N-1}.
    > - When Pods are being deleted, they are terminated in reverse order, from {N-1..0}.
    > - Before a scaling operation is applied to a Pod, all of its predecessors must be Running and Ready.
    > - Before a Pod is terminated, all of its successors must be completely shutdown.
    - https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#deployment-and-scaling-guarantees
- 参考URL
  - https://cloud.netapp.com/blog/cvo-blg-kubernetes-deployment-vs-statefulset-which-is-right-for-you
  - https://www.baeldung.com/ops/kubernetes-deployment-vs-statefulsets
  - https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/


podはデフォルト(ユーザ指定なし)ではrootで実行される
→pod内のrootはLinux上のrootとは違う
　→「securityContext.capabilities.add[drop]」に["NET_ADMIN", "SYS_TIME"]のように配列形式で追加/削除


kubectl describe po <POD名>で表示される「Last State」で前回の状態とその下の「Reason」その理由が分かる
