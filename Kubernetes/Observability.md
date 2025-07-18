## その他
- **`kube_pod_container_status_last_terminated_reason`**
  - PodがTerminatedされた時にCountされる。`reason`ラベルに`OOMKilled`など、Terminatedされた理由が入る。
- **`kube_pod_container_status_restarts_total`** (counter)
  - PodがRestartされた時にCountされる。

## CPU関連のメトリクス
#### **`container_cpu_cfs_throttled_seconds_total`** (counter)
- CPU Limitを設定している場合、PodのCPU使用率がLimitを超えてthrottleされた時間
- 参考URL
  - https://www.metricfire.com/blog/top-10-cadvisor-metrics-for-prometheus/#containerfsiotimesecondstotal
  - https://medium.com/orangesys/a-deep-dive-into-kubernetes-metrics-part-3-7333fae67403

#### **`container_cpu_cfs_periods_total`** (counter)
- LinuxカーネルのCFS (Completely Fair Scheduler) において、コンテナに割り当てられたCPUクォータが適用される期間の総数
- CPUクォータが設定されているコンテナがCPUリソースを使用するために与えられた「期間（period）」の合計回数を表す

#### **`container_cpu_cfs_throttled_periods_total`** (counter)
- CPUリミットに達してスロットリングが発生した期間の回数

### `container_cpu_cfs_periods_total`と`container_cpu_cfs_throttled_periods_total`の関係  
```
時間軸: [期間1] [期間2] [期間3] [期間4] [期間5]
結果:   正常    スロットル  正常    スロットル  正常
```
上記の場合、
`container_cpu_cfs_periods_total` = 5（全期間数）
`container_cpu_cfs_throttled_periods_total` = 2（スロットリング発生期間数）

- CPUスロットリング率  
  ```
  rate(container_cpu_cfs_throttled_seconds_total[5m]) / rate(container_cpu_cfs_periods_total[5m]) * 100
  ```