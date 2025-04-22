# QoS Class
- 参考URL
  - https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/
  - https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/
- Podに設定されている `Requests`と`Limits`の値に応じて自動的に設定されるもので、3つのClassがあり、各Classは優先度があって、**Worker Nodeのメモリが足りなくなった（Node Pressure）時、Eviction ManagerがどのPodをEvictionするか決める際に使われる**
  |QoS Class|条件|優先度|
  |---|---|---|
  |`Guaranteed`|CPUとMemoryのRequestとLimitが設定されていて、CPU Request = CPU Limit、Memory Request = Memory Limitである|1|
  |`Burstable`|At least one Container in the Pod has a memory or CPU request or limit|2|
  |`BestEffort`|CPUとMemory両方で、RequestとLimitどっちも設定されてない|3|
- `BestEffort` → `Burstable` → `Guaranteed` の順に終了される
- 以下のコマンドで確認できる  
  ```shell
  kubectl get po -o custom-columns="NAME:{.metadata.name},QoS Class:{.status.qosClass}"
  ```
- QoS Classが適用されるのはWorker Nodeのメモリが足りなくなった場合であって、Pod単体（Pod個別）でPodのMemory使用量がPod Memory Limitに達した場合は即座にそのPodはOOM Killedされる

## CPUとMemoryで、`requests`==`limits`に設定すべきか
- CPUについては、一般的に`requests != limits`の設定が推奨されることが多い
  - CPUはコンプレッシブル（圧縮可能）なリソースで、使用量が制限を超えた場合はスロットリングされるだけで、Podが強制終了されることはないため
  - `requests`と`limits`に差を設けることで、CPU使用量がスパイクする場合でも柔軟に対応でき、リソース効率が向上するため
- Memoryについては、`requests == limits`の設定が推奨されることが多い
  - メモリはインコンプレッシブル（非圧縮）なリソースで、制限を超えるとOOMKiller（Out of Memory Killer）によってPodが強制終了されるため
  - Podのスケジューリングは`requests`をもとに行われるため、`requests != limits`の場合、実際のPodのMemory使用量は`limits`に近く、ノードのMemory使用率は限界なのに`requests`の値だけでPodがスケジューリングされて、結果Node Pressureが発生し、Podがevictionされる可能性があるため

# PriorityClass
- Schedulerがスケジューリングのために既存のPodを削除（Preemption）する際に使われるもの