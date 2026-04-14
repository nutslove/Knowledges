# Push Gatewayとは
- Push Gatewayは、Prometheusが直接スクレイピングできない短命のジョブやバッチジョブからメトリクスを収集するためのコンポーネント。短命のジョブからメトリクスをプッシュするためのエンドポイントを提供し、Prometheusは定期的にPush Gatewayからこれらのメトリクスをスクレイピングする。

# Push Gatewayの設定について
- https://github.com/prometheus/pushgateway
- defaultではPush Gatewayはメトリクスを永続化しないため、Push Gatewayを再起動するとすべてのメトリクスが失われる。永続化を有効にするには、`--persistence.file`フラグを使用して永続化ファイルのパスを指定する必要がある。  
  > By default, Pushgateway does not persist metrics. However, the --persistence.file flag allows you to specify a file in which the pushed metrics will be persisted (so that they survive restarts of the Pushgateway).

## Push Gatewayのマニフェストファイル例
```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pushgateway
  namespace: monitoring
  labels:
    app: pushgateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pushgateway
  template:
    metadata:
      labels:
        app: pushgateway
    spec:
      containers:
      - name: pushgateway
        image: prom/pushgateway:v1.11.2
        args:
          - --persistence.file=/data/pushgateway.data
          - --persistence.interval=5m
          - --web.enable-admin-api  # DELETE API など、Admin APIを使う場合に必要
        ports:
        - name: http
          containerPort: 9091
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /-/healthy
            port: http
        readinessProbe:
          httpGet:
            path: /-/ready
            port: http
      volumes:
      - name: data
        emptyDir: {}  # 永続化したい場合は PVC に変更
---
apiVersion: v1
kind: Service
metadata:
  name: pushgateway
  namespace: monitoring
  labels:
    app: pushgateway
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 9091
    targetPort: 9091
  selector:
    app: pushgateway
```

> [!NOTE]
> - `--persistence.interval` は、Pushgatewayがメモリ上のメトリクスをディスク（`--persistence.file` で指定したファイル）に書き出す間隔を指定するフラグ。デフォルトは5分（5m）。
> - Pushgatewayは受け取ったメトリクスを基本的にメモリ上に保持する。そのままだとPod再起動やクラッシュですべてのメトリクスが失われるため、定期的にディスクにスナップショットを書き出し、起動時にそのファイルから復元するようになっている。

## Push GatewayのメトリクスをPrometheusでスクレイピングする方法
- Prometheusのscrape config例  
  ```yaml
  scrape_configs:
    - job_name: 'pushgateway'
      honor_labels: true  # 必須：push側のjob/instanceを保持
      static_configs:
        - targets: ['pushgateway.monitoring.svc.cluster.local:9091']
  ```

> [!IMPORTANT]
> - `honor_labels: true` を設定しないと、Push Gateway側の `job`、`instance` ラベルが `exported_job`、`exported_instance` にリネームされてしまい、Push Gatewayにプッシュしたときの `job`、`instance` ラベルはPrometheus側のConfigで指定した `job`、`instance` ラベルに上書きされてしまうため、Push Gatewayにプッシュしたときの `job`、`instance` ラベルを保持したい場合は必ず `honor_labels: true` を設定する必要がある。
> - https://github.com/prometheus/pushgateway?tab=readme-ov-file#about-the-job-and-instance-labels  
  > ### About the job and instance labels
  > The Prometheus server will attach a `job` label and an `instance` label to each scraped metric. The value of the `job` label comes from the scrape configuration. When you configure the Pushgateway as a scrape target for your Prometheus server, you will probably pick a job name like `pushgateway`. The value of the `instance` label is automatically set to the host and port of the target scraped. Hence, all the metrics scraped from the Pushgateway will have the host and port of the Pushgateway as the `instance` label and a `job` label like `pushgateway`. The conflict with the `job` and `instance` labels you might have attached to the metrics pushed to the Pushgateway is solved by renaming those labels to `exported_job` and `exported_instance`.
  >
  > However, this behavior is usually undesired when scraping a Pushgateway. Generally, you would like to retain the `job` and `instance` labels of the metrics pushed to the Pushgateway. That's why you have to set `honor_labels: true` in the scrape config for the Pushgateway. It enables the desired behavior. See the [documentation](https://prometheus.io/docs/operating/configuration/#scrape_config) for details.
  >
  > This leaves us with the case where the metrics pushed to the Pushgateway do not feature an `instance` label. This case is quite common as the pushed metrics are often on a service level and therefore not related to a particular instance. Even with `honor_labels: true`, the Prometheus server will attach an `instance` label if no `instance` label has been set in the first place. Therefore, if a metric is pushed to the Pushgateway without an instance label (and without instance label in the grouping key, see below), the Pushgateway will export it with an empty instance label (`{instance=""}`), which is equivalent to having no `instance` label at all but prevents the server from attaching one.



# Push Gatewayへのメトリクスのプッシュ方法