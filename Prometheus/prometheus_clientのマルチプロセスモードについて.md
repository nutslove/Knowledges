#### 前提知識
- **prometheus_client**
  - Python 用の公式 Prometheus クライアントライブラリ。
  - `Counter`, `Gauge`, `Histogram`, `Summary` などのメトリクスを定義し、HTTP エンドポイント（通常 `/metrics`）で公開する。
  - https://github.com/prometheus/client_python
- **Gunicorn / Uvicorn のワーカーモデル**
  - Gunicorn (preforkモデル) は master プロセスが複数の worker プロセスを fork し、各 worker が独立した Python プロセスとしてリクエストを捌く。
  - 各 worker は **完全に別プロセス** であり、メモリは共有されない。
- **prometheus-fastapi-instrumentator**
  - FastAPI 用に `/metrics` エンドポイントの公開を簡単にしてくれるラッパー。
  - 内部で `prometheus_client` を使う。

---

### 現象: Counter なのに Grafana で値が前後する

`Counter` は単調増加するはずなのに、Grafana のグラフを見ると以下のような挙動になる:

- `0 → 1 → 2 → 1 → 0 → 2 → 1 → ...` のように **値が増減** する
- ECS タスクや Pod は 1 つしか動いていないのに不自然な値になる

---

### 原因: マルチワーカー × プロセスローカルなメトリクス

`prometheus_client` の `Counter` / `Gauge` は **プロセス内メモリ** に値を保持する。Gunicorn を `workers=N` で起動すると、N 個の worker プロセスがそれぞれ独立した Counter インスタンスを持つ。

例: Gunicorn 6 ワーカーで動いている場合、

```
Master process
├─ Worker #1  → Counter("foo_total") = 3
├─ Worker #2  → Counter("foo_total") = 1
├─ Worker #3  → Counter("foo_total") = 0
├─ Worker #4  → Counter("foo_total") = 2
├─ Worker #5  → Counter("foo_total") = 1
└─ Worker #6  → Counter("foo_total") = 0
```

Prometheus が `/metrics` をスクレイプすると、Gunicorn master がそのリクエストを **どれか1つの worker にラウンドロビンで振り分ける**。応答するのは振り分け先の worker だけなので、スクレイプごとに違う値が返る。

```
スクレイプ① → Worker #1 → 3
スクレイプ② → Worker #2 → 1   ← 「3 から 1 に減った」
スクレイプ③ → Worker #4 → 2
スクレイプ④ → Worker #6 → 0   ← 「2 から 0 に減った」
```

Prometheus 側は同じ `instance` ラベルとして扱うため、これを **Counter のリセット** と誤検知してしまう。Grafana のグラフが「Counter なのに値が前後する」状態になる。

> 同じことは ALB / Ingress 経由でスクレイプした場合にも発生する（タスク/Pod 単位での分散）。今回はその worker 版。

---

### 解決策: prometheus_client のマルチプロセスモード

`prometheus_client` には公式のマルチプロセス対応機能があり、各 worker のメトリクスをファイルベースで集約できる。

公式ドキュメント: https://prometheus.github.io/client_python/multiprocess/

#### 仕組み

1. 各 worker プロセスは、メトリクス値を **mmap ファイル** に書き込む（プロセス内メモリではなく）。
2. `/metrics` のスクレイプ時、`MultiProcessCollector` が **共有ディレクトリ内の全 mmap ファイルを読み込んで集約** する。
3. ファイル名に worker の pid が含まれており、worker 終了時に「dead」マークを付けることで、新しい pid と分離して管理できる。

#### 設定手順

##### 1. 環境変数 `PROMETHEUS_MULTIPROC_DIR` を設定

mmap ファイルを格納するディレクトリ。コンテナ環境なら `/tmp/prometheus_multiproc` などの ephemeral storage 配下が無難。

```dockerfile
ENV PROMETHEUS_MULTIPROC_DIR=/tmp/prometheus_multiproc
```

##### 2. Gunicorn のフックを設定

`gunicorn.conf.py` に以下を追加。

```python
import os
import shutil
from typing import Any

from prometheus_client import multiprocess


def on_starting(server: Any) -> None:
    # 前回起動時の mmap ファイルが残っていると死んだ pid の値が混ざるため、毎回クリアする
    multiproc_dir = os.environ.get("PROMETHEUS_MULTIPROC_DIR")
    if multiproc_dir:
        shutil.rmtree(multiproc_dir, ignore_errors=True)
        os.makedirs(multiproc_dir, exist_ok=True)


def child_exit(server: Any, worker: Any) -> None:
    # worker 終了時にそのプロセスのメトリクスファイルを「dead」マークしてマージ対象から外す
    multiprocess.mark_process_dead(worker.pid)
```

- `on_starting`: master プロセスの初期化前に呼ばれる。前回起動時の mmap ファイルが残ったままだと、消えた pid の値が永続的に集計に混ざってしまうので必ずクリアする。
- `child_exit`: worker 終了直後に master 側で呼ばれる。`worker_exit` (worker 側) と紛らわしいが、prometheus_client 公式ドキュメントは `child_exit` を推奨。

##### 3. `/metrics` エンドポイントの実装

`prometheus_client` を直接使う場合は `MultiProcessCollector` を手動で登録する。

```python
from prometheus_client import CollectorRegistry, generate_latest, multiprocess

def metrics():
    registry = CollectorRegistry()
    multiprocess.MultiProcessCollector(registry)
    return generate_latest(registry)
```

`prometheus-fastapi-instrumentator` (v6 以降) を使う場合は、**`PROMETHEUS_MULTIPROC_DIR` が設定されていれば自動でマルチプロセス対応する** ので、アプリ側のコード変更は不要。

```python
# v6 以降のソースで PROMETHEUS_MULTIPROC_DIR を検出して MultiProcessCollector を使う
Instrumentator().instrument(app).expose(app)
```

---

### 注意点

#### Gauge は `multiprocess_mode` の指定が必要

`Counter` / `Histogram` / `Summary` はマルチプロセスモードで自動的に sum 集約されるが、`Gauge` は集約方法を **明示的に指定** する必要がある。指定しないと `'all'` 扱いになり、worker ごとに `pid` ラベル付きの別系列として公開され、Grafana 側でクエリが破綻する。

| モード | 意味 |
|---|---|
| `livesum` | **生きている** worker の合計（推奨）|
| `livemax` | 生きている worker の最大値 |
| `livemin` | 生きている worker の最小値 |
| `liveall` | 生きている worker ごとに別系列で公開 |
| `sum` / `max` / `min` / `all` | 死んだ worker の値も含める |
| `mostrecent` / `livemostrecent` | 最新タイムスタンプの値を採用 |

```python
from prometheus_client import Gauge

# 全 worker の合計を取りたい場合
db_connections_current = Gauge(
    "db_connections_current",
    "Current active connections",
    multiprocess_mode="livesum",
)
```

#### 比率（Ratio）系の Gauge は sum できない

「現在使用中 / 上限」のような比率を Gauge で公開しているケースで、各 worker の比率を `livesum` で足すと **意味のない数字** になる。

例: 3 worker × pool_size=10、各 worker で 5 接続使用中
- 全体の正しい利用率: 15 / 30 = **0.5**
- 各 worker の比率: 0.5, 0.5, 0.5 → `livesum` で合算すると **1.5**（誤り）

正しい運用:
- 比率 Gauge は **削除する**
- 分子（current）と分母（max）をそれぞれ `livesum` で公開
- Grafana 側で `sum(numerator) / sum(denominator)` として算出する

#### `_created` メトリクスが消える

通常モードでは Counter / Summary / Histogram に `<metric>_created` という補助メトリクス（作成タイムスタンプ）が自動で付くが、マルチプロセスモードでは仕様により公開されなくなる。実害はないが、ダッシュボードや Alert ルールで参照していないか念のため確認する。

#### ディレクトリのライフサイクル

- mmap ファイルはコンテナ再起動で消えても問題ない（元々プロセスメモリだったので、再起動でリセットされるのは仕様通り）。
- 永続ボリュームに配置する **必要はない**。むしろ ephemeral storage（`/tmp` など）が推奨。
- 1 worker あたり数十 KB 程度のサイズ。容量心配は不要。

#### ローカル開発・テスト環境

`PROMETHEUS_MULTIPROC_DIR` を設定しなければ通常モードで動作するので、ローカル / pytest 実行には影響しない。`multiprocess_mode` 引数も通常モードでは無視されるだけで、エラーにはならない。

---

### 確認方法

#### 設定が効いているかの確認

```bash
# /metrics をスクレイプして、同じ instance なのに値が前後しないことを確認
for i in $(seq 1 10); do
  curl -s http://localhost:8080/metrics | grep '^comment_operations_total'
done
```

設定前: 値がスクレイプごとに変わる。
設定後: 値が単調増加する（または変わらない）。

#### multiproc_dir 内の状態

```bash
ls -la /tmp/prometheus_multiproc/
# counter_<pid>.db, gauge_livesum_<pid>.db などのファイルが worker 数分存在する
```

---

### 参考URL

- prometheus_client 公式 multiprocess ドキュメント: https://prometheus.github.io/client_python/multiprocess/
- Gunicorn server hooks リファレンス: https://docs.gunicorn.org/en/stable/settings.html#server-hooks
- prometheus-fastapi-instrumentator (multiprocess 対応箇所): https://github.com/trallnag/prometheus-fastapi-instrumentator
