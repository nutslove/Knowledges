## Chartのテンプレート
- `helm pull <Chart名> --untar `でインストールせず、templatesやvalues.yml等をダウンロードすることができる
  - e.g. `helm pull grafana/loki-distributed --version 0.69.9 --untar`
  - **ディレクトリごとにダウンロードする**
- `helm template <リリース名> <Chart名> > chart.yml`で値の入った完成されたYamlファイルをダウンロードできる

## `helm install` or `helm upgrade`時に特定のChartバージョンを使いたい場合は`--version`フラグで指定する
- e.g. `helm upgrade multi-tenant grafana/loki-distributed --version 0.69.9 --namespace=monitoring --values=loki-2.7.yml`

## Helm各種コマンド
- repoリスト表示
  - `helm repo list`
- ChartのバージョンとChartバージョンに紐づくソフトフェアのバージョンの一覧を表示
  - `helm search repo <Chart名> --versions`
    - e.g. `helm search repo grafana/loki-distributed --versions`
- Chartバージョンアップ
  - `helm repo update`