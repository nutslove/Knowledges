## ASGI とは

**ASGI (Asynchronous Server Gateway Interface)** は、Pythonの**ASGIサーバーとASGIアプリケーション間のインターフェース仕様**。
WSGIの非同期版にあたる。

- ソフトウェアの種類ではなく「規約（specification）」
- 「ASGIサーバーはASGIアプリにこういう形でデータを渡す」「ASGIアプリはこういう形で返す」というプロトコルを定義
- 非同期処理 (`async/await`)、WebSocket、HTTP/2 などに対応するために設計された

> [!NOTE]
> #### 補足: ここで言う「サーバー」とは？**
>
> ASGI仕様の文脈での「サーバー」は **ASGIサーバー（Uvicorn, Hypercorn, Daphne）** を指す。
> NginxやALB、EC2のような物理/仮想マシンは含まれない。
>
> | 文脈での「サーバー」 | 具体例 |
> |------------------|-------|
> | 物理/仮想マシン | EC2インスタンス、オンプレのマシン |
> | HTTPサーバー（Webサーバー） | Nginx, Apache, ALB |
> | **ASGI/WSGIサーバー** | **Uvicorn, Gunicorn, uWSGI** |
>
> ASGI仕様が定めているのは、**ASGIサーバーとASGIアプリの間でやりとりするデータの形**:
> - HTTPリクエストをどんなPythonの辞書（`scope`）としてアプリに渡すか
> - アプリがレスポンスをどんな形（`send`関数の呼び出し）でサーバーに返すか
> - WebSocketやライフスパン（起動・終了）イベントをどう扱うか

## なぜ必要か

FastAPIなどのASGIアプリは、リクエストを処理するロジックは持つが、
**TCPソケットをlistenしてHTTPプロトコルを話す機能は持たない**。

ALBやNginxは普通のHTTPを喋るので、その間で「HTTP ⇔ ASGI」の変換をするASGIサーバーが必要。

```
[HTTPクライアント / ALB / Nginx]
        ↓ HTTP
[ASGIサーバー: Uvicorn等]   ← HTTPをパースしてASGI形式に変換
        ↓ ASGI仕様（Pythonの関数呼び出し）
[ASGIアプリ: FastAPI等]     ← ASGI仕様に従って書かれている
```

## 登場人物の分類

### ASGIサーバー（HTTPを受けてASGI仕様でアプリを呼ぶ）

| サーバー | 特徴 |
|---------|------|
| **Uvicorn** | 最も一般的。軽量・高速。本番でも使える |
| **Hypercorn** | HTTP/2、HTTP/3対応 |
| **Daphne** | Django Channelsの公式サーバー |

### ASGIアプリケーション / フレームワーク（ASGI仕様に従って呼ばれる側）

- **FastAPI**
- **Starlette**（FastAPIのベース）
- **Django** (3.0以降)
- **Quart**（Flaskの非同期版）

## Gunicorn の位置づけ

**Gunicorn自体はASGIサーバーではない**（元々はWSGIサーバー兼プロセスマネージャ）。

ただし、worker classを差し替える仕組みがあり、`uvicorn.workers.UvicornWorker` を指定すると
各workerプロセスがUvicornとして動く。

```
Gunicorn（プロセスマネージャ）
  ├─ Worker 1: Uvicorn（ASGIサーバー） → FastAPI
  ├─ Worker 2: Uvicorn（ASGIサーバー） → FastAPI
  └─ Worker 3: Uvicorn（ASGIサーバー） → FastAPI
```

- **Gunicorn の役割**: workerの起動・監視・再起動・graceful shutdown
- **Uvicorn の役割**: 実際にHTTPを話してASGIでアプリを呼ぶ

起動コマンド例:
```bash
gunicorn main:app -k uvicorn.workers.UvicornWorker -w 4 --bind 0.0.0.0:8000
```

## WSGI との対比

同じ構造がWSGIの世界にもある。

| 区分 | ASGI（非同期） | WSGI（同期） |
|------|--------------|-------------|
| 仕様 | ASGI | WSGI |
| サーバー | Uvicorn, Hypercorn, Daphne | Gunicorn, uWSGI |
| フレームワーク | FastAPI, Starlette, Django(3.0+) | Flask, Django(従来) |

## 本番構成の例

### シンプル構成（EKS/ECS等）
```
ALB → Service → Pod (Uvicornコンテナ + FastAPIアプリ)
```

起動例:
```bash
uvicorn main:app --host 0.0.0.0 --port 8000
```

### Gunicorn + Uvicorn workers構成（プロセス管理を強化）
```
ALB → Service → Pod (Gunicorn → Uvicorn workers × N + FastAPIアプリ)
```

### Nginxを挟むケース
- 静的ファイル配信
- 追加のレート制限
- TLS終端の追加レイヤー
- ALB → Uvicornで十分なケースも多く、Nginxは**必須ではない**

## ポイント整理

- **ASGI**は仕様であって、ソフトウェアではない
- **Uvicorn / Hypercorn / Daphne** がASGIサーバー（HTTPを話す側）
- **FastAPI / Starlette / Django** がASGIアプリ（呼ばれる側）
- **Gunicorn** はASGIサーバーではないが、Uvicorn workerと組み合わせて本番で使える
- FastAPI単体ではHTTPを話せないので、ASGIサーバーは**必須**
- Nginxやリバースプロキシは**必須ではない**（要件次第）