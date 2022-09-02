## Disk関連
- 参考URL  
  - https://brian-candler.medium.com/interpreting-prometheus-metrics-for-linux-disk-i-o-utilization-4db53dfedcfc
  - https://christina04.hatenablog.com/entry/prometheus-node-monitoring
  - https://devconnected.com/monitoring-disk-i-o-on-linux-with-the-node-exporter/
  - https://qiita.com/Esfahan/items/01833c1592910fb11858
- node_exporterはdisk関連メトリクスを`/proc/diskstats`から取得する
- I/Oスループット  
  - Read  
    `node_disk_read_bytes_total`
  - Write  
    `node_disk_written_bytes_total`