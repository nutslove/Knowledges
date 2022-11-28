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

## Probeについて
- k8sには3種類のprobeが存在する
  - https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
1. ___livenessProbe___
   > Indicates whether the container is running. If the liveness probe fails, the kubelet kills the container, and the container is subjected to its restart policy. If a container does not provide a liveness probe, the default state is Success.
2. ___readinessProbe___
   > Indicates whether the container is ready to respond to requests. If the readiness probe fails, the endpoints controller removes the Pod's IP address from the endpoints of all Services that match the Pod. The default state of readiness before the initial delay is Failure. If a container does not provide a readiness probe, the default state is Success.
3. ___startupProbe___
   > Indicates whether the application within the container is started. All other probes are disabled if a startup probe is provided, until it succeeds. If the startup probe fails, the kubelet kills the container, and the container is subjected to its restart policy. If a container does not provide a startup probe, the default state is Success.