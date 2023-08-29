## MicroServices Mode
- githubリポジトリ
  - https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed
- Volumesは`/var/loki`にマウントされるので、各設定上のdirectoryは`/var/loki`配下(e. g. `/var/loki/index`, `/var/loki/cache`)に設定すること
- Gateway、Ingester以外は個別(e.g. IndexGateway)に変数化されてなくて`livenessProbe`の設定ができない  
  `loki.livenessProbe`のようにLoki全体に設定する必要がある