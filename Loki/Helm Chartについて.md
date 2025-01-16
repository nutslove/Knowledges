- 手順
  - https://grafana.com/docs/loki/latest/setup/install/helm/

# *Update!* LokiのHelm Chart
- すべてのModeで以下の共通の１つのHelmチャートを使うようになった模様
  - https://grafana.github.io/helm-charts の`grafana/loki`
- 以下で確認できる  
  ```shell
  helm repo add grafana https://grafana.github.io/helm-charts
  helm repo update
  helm search repo grafana/loki --versions
  ```

## MicroServices Mode
> [!CAUTION]
> 以下のHelmチャートはアップデートされなくなった。  
> `grafana/loki`チャートを使うこと！
- githubリポジトリ
  - https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed
- Volumesは`/var/loki`にマウントされるので、各設定上のdirectoryは`/var/loki`配下(e. g. `/var/loki/index`, `/var/loki/cache`)に設定すること
- Gateway、Ingester以外は個別(e.g. IndexGateway)に変数化されてなくて`livenessProbe`の設定ができない  
  `loki.livenessProbe`のようにLoki全体に設定する必要がある