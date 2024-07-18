- `helm create <Helmチャート名>`で`<Helmチャート名>`ディレクトリと、その中に`charts`、`templates`ディレクトリと`Chart.yaml`、`values.yaml`などが作成される
  - nginxのHelmチャートが作成される

## Helmリポジトリを追加
- `helm repo add <Repo名> <RepoのURL>`

## Chartのテンプレート
- `helm pull <Chart名> --untar `でインストールせず、templatesやvalues.yml等をダウンロードすることができる
  - e.g. `helm pull grafana/loki-distributed --version 0.69.9 --untar`
  - **ディレクトリごとにダウンロードする**
- `helm template <リリース名> <Chart名> > chart.yml`で値の入った完成されたYamlファイルをダウンロードできる

### テンプレート内のvalueの代入

### テンプレート内でのif文

### テンプレート内でのrange文
- 反復処理(for文)
- 例  
  ```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: example-configmap
  data:
  {{ range $key, $value := .Values.configs }}
    {{ $key }}: {{ $value }}
  {{ end }}
  ```

### テンプレート内の`include`と`template`について
- 両方とも通常`_helpers.tpl`ファイルから定義されたテンプレートを参照するけど、`include`は`{{ include "myapp.container" . | nindent 4 }}`のように他の関数と組み合わせて使うことができる

### `{{ }}`内の`-`について
- `{{-`のように`-`が付く場合、テンプレートのレンダリング時に余計なスペースや改行を除去される
- **`-`がない場合**
  - テンプレート  
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-configmap
    data:
    {{ if .Values.enabled }}
    key1: value1
    {{ else }}
    key2: value2
    {{ end }}
    ```
  - レンダリング後のテンプレート（`Values.enabled`が`true`の場合）  
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-configmap
    data:

    key1: value1
    ```
- **`-`がある場合**
  - テンプレート  
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-configmap
    data:
    {{- if .Values.enabled }}
    key1: value1
    {{- else }}
    key2: value2
    {{- end }}
    ```
  - レンダリング後のテンプレート（`Values.enabled`が`true`の場合）  
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-configmap
    data:
    key1: value1
    ```

### `indent <数字>`、`nindent <数字>`について
1. `n`: 新しい行（newline）を追加する
2. `indent`: インデントを追加する
3. 数字`: 数字の分、スペースでインデントを行う

- 例えば、`nindent 4` は次の動作を行う：
  1. まず、新しい行を追加する
  2. その後、その新しい行から始まるすべての行に4つのスペースのインデントを追加する

- 例  
  ```yaml
  spec:
    {{ include "myapp.container" . | nindent 4 }}
  ```

  上記の例の場合、`myapp.container` テンプレートの内容が新しい行から始まり、その内容全体が4スペースでインデントされる。

  結果は以下のようになる

  ```yaml
  spec:
      container:
        name: myapp
        image: myapp:1.0.0
        ports:
          - containerPort: 8080
  ```

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

## 自作のHelmチャート/リポジトリ作成方法
- `charts`ディレクトリがあるところで`helm package charts/<チャート名>`コマンドでHelmチャートの`*.tgz`にアーカイブする（e.g. `helm package charts/opensearch`）
  - `charts/<チャート名>/Chart.yaml`ファイルにある`version`、`appVersion`などをもとに作成される

- `*.tgz`があるディレクトリで`helm repo index . --url <HelmリポジトリURL>`で`index.yaml`を生成
  - `*.tgz`をもとに作成される

- **デフォルトでは`helm repo index`コマンドは既存の`index.yaml`を上書きしてしまうため、既存の内容を残して新しい内容を追記するためには既存の`index.yaml`を退避しといて`--merge`フラグでそれを指定する必要がある。**
  - **`helm repo index . --url <HelmリポジトリURL> --merge <退避しといた既存のindex.yaml>`**

- チャートを更新し、`*.tgz`と`index.yaml`を更新した場合は、`helm chart update`で反映が必要

- Nginx/ApacheなどのWebサーバやS3/SwiftなどのObject Storageに`*.tgz`と`index.yaml`をアップロードして、`index.yaml`内の`urls`の値をそのエンドポイントに修正
