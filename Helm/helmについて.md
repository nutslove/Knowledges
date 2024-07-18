- `helm create <Helmチャート名>`で`<Helmチャート名>`ディレクトリと、その中に`charts`、`templates`ディレクトリと`Chart.yaml`、`values.yaml`などが作成される
  - nginxのHelmチャートが作成される

## Helmリポジトリを追加
- `helm repo add <Repo名> <RepoのURL>`

## Chartのテンプレート
- `helm pull <Chart名> --untar `でインストールせず、templatesやvalues.yml等をダウンロードすることができる
  - e.g. `helm pull grafana/loki-distributed --version 0.69.9 --untar`
  - **ディレクトリごとにダウンロードする**
- `helm template <リリース名> <Chart名> > chart.yml`で値の入った完成されたYamlファイルをダウンロードできる

## `helm install` or `helm upgrade`時に特定のChartバージョンを使いたい場合は`--version`フラグで指定する
- e.g. `helm upgrade multi-tenant grafana/loki-distributed --version 0.69.9 --namespace=monitoring --values=loki-2.7.yml`

## テンプレート内でのif文
- 

## Helm各種コマンド
- repoリスト表示
  - `helm repo list`
- ChartのバージョンとChartバージョンに紐づくソフトフェアのバージョンの一覧を表示
  - `helm search repo <Chart名> --versions`
    - e.g. `helm search repo grafana/loki-distributed --versions`
- Chartバージョンアップ
  - `helm repo update`

## 自作のHelmチャート/リポジトリ作成方法
- `charts`ディレクトリがあるところで`helm package charts/<チャート名>`コマンドでHelmチャートの`*.tgz`にアーカイブする（e.g. `helm package charts/opensearch`）
  - `charts/<チャート名>/Chart.yaml`ファイルにある`version`、`appVersion`などをもとに作成される

- `*.tgz`があるディレクトリで`helm repo index . --url <HelmリポジトリURL>`で`index.yaml`を生成
  - `*.tgz`をもとに作成される

- **デフォルトでは`helm repo index`コマンドは既存の`index.yaml`を上書きしてしまうため、既存の内容を残して新しい内容を追記するためには既存の`index.yaml`を退避しといて`--merge`フラグでそれを指定する必要がある。**
  - **`helm repo index . --url <HelmリポジトリURL> --merge <退避しといた既存のindex.yaml>`**

- チャートを更新し、`*.tgz`と`index.yaml`を更新した場合は、`helm chart update`で反映が必要

- Nginx/ApacheなどのWebサーバやS3/SwiftなどのObject Storageに`*.tgz`と`index.yaml`をアップロードして、`index.yaml`内の`urls`の値をそのエンドポイントに修正
