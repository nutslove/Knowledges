https://fastapi.tiangolo.com/ko/reference/apirouter/?h=apirouter

## 1. 基本: 通常の `@app.get` 方式 vs `APIRouter` 方式

### 通常の方式（小規模向け）

```python
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Hello"}

@app.get("/users")
async def get_users():
    return [...]

@app.post("/users")
async def create_user():
    return {...}
```

- すべてのエンドポイントを `main.py` 1ファイルに書く
- エンドポイントが少ないうちは見通しがいい
- 増えてくると一気にカオスになる

### APIRouter 方式（中〜大規模向け）

```python
# routers/users.py
from fastapi import APIRouter

router = APIRouter(
    prefix="/users",
    tags=["users"],
)

@router.get("/")
async def get_users():
    return [...]

@router.get("/{user_id}")
async def get_user(user_id: int):
    return {"user_id": user_id}
```

```python
# main.py
from fastapi import FastAPI
from routers import users, items, auth

app = FastAPI()

app.include_router(users.router)
app.include_router(items.router)
app.include_router(auth.router)
```

- 機能ごとにファイルを分けられる
- `main.py` は登録だけに専念できる
- 共通設定（prefix, tags, 認証）をまとめて適用できる

---

## 2. `@app.xxx` と `@router.xxx` の違い

| 項目 | `@app.get` など | `@router.get` など |
|------|----------------|-------------------|
| インポート元 | `FastAPI` | `APIRouter` |
| 用途 | アプリ全体に直接登録 | ルーターにまとめて、後で `include_router` で登録 |
| 共通 prefix | 個別に書くしかない | `APIRouter(prefix=...)` で一括 |
| 共通 tags | 個別に書くしかない | `APIRouter(tags=[...])` で一括 |
| 共通の依存関係（認証など） | 個別に書く or `app` 全体 | `APIRouter(dependencies=[...])` で一括 |
| ファイル分割 | 困難 | 自然にできる |

機能としては同じことができますが、**スケーラビリティと整理のしやすさ** が大きく違います。

---

## 3. `APIRouter` のよく使うオプション

```python
from fastapi import APIRouter, Depends, HTTPException

router = APIRouter(
    prefix="/users",                       # 全エンドポイントの先頭に付くパス
    tags=["users"],                        # OpenAPI のグルーピング用
    dependencies=[Depends(verify_token)],  # ルーター内全エンドポイントに適用される依存
    responses={404: {"description": "Not found"}},  # 共通レスポンス定義
)
```

`include_router` 側でも追加で指定できる:

```python
app.include_router(
    users.router,
    prefix="/api/v1",                      # 元の prefix と結合される → /api/v1/users
    tags=["v1"],                           # 追加される
    dependencies=[Depends(another_check)], # 追加される
)
```

→ パスは `/api/v1/users/{user_id}` のように結合される。

---

## 4. 推奨ディレクトリ構成

```
project/
├── main.py
├── routers/
│   ├── __init__.py
│   ├── users.py
│   ├── items.py
│   └── auth.py
├── dependencies.py        # 共通の依存関係（認証など）
├── models/                # Pydantic モデル
│   └── user.py
├── services/              # ビジネスロジック
│   └── user_service.py
└── core/
    └── config.py          # 設定
```

ポイント:

- `routers/` 配下は **ルーティングと薄い入出力変換のみ**
- ビジネスロジックは `services/` に切り出す（agent ロジックもここ）
- `models/` は Pydantic スキーマ専用

---

## 5. ルーターのネスト（router の中に router を含める）

`APIRouter` は他の `APIRouter` を `include_router` できる。バージョニングや階層化に便利。

```python
# routers/v1/__init__.py
from fastapi import APIRouter
from . import users, items

api_v1 = APIRouter(prefix="/v1")
api_v1.include_router(users.router)
api_v1.include_router(items.router)
```

```python
# main.py
from fastapi import FastAPI
from routers.v1 import api_v1

app = FastAPI()
app.include_router(api_v1, prefix="/api")
# → /api/v1/users, /api/v1/items
```

---

## 6. いつ `APIRouter` を使うべきか

### 使うべきケース

- エンドポイントが **5〜10個以上** になりそうなとき
- 機能ドメインが明確に分かれているとき（users, items, auth, agents など）
- **API バージョニング** が必要なとき（`/api/v1`, `/api/v2`）
- 一部のエンドポイントだけに **共通の認証/依存** を適用したいとき
- チーム開発で **ファイルの責任範囲を分けたい** とき
- マイクロサービスや agent ごとに API を切り出したいとき

### 使わなくていいケース

- プロトタイプ・PoC で、エンドポイント 2〜3個で済むとき
- 1ファイルのスクリプト的な API
- ヘルスチェック専用の極小サービス

### 実務的な目安

> **迷ったら最初から `APIRouter` で書く**

理由: 後から `@app.xxx` → `@router.xxx` に書き換えるのは地味に面倒（インポート、デコレーター、prefix の調整全部発生する）。最初から `APIRouter` で書いていれば、規模が大きくなっても自然に拡張できる。

特に AI agent 系の API では、エージェントごとに `APIRouter` を切る構成が見通しよく、Langfuse の trace タグや認証も agent 単位で管理しやすい。

---

## 7. 認証を agent ごとに変える例

```python
# routers/rca_agent.py
from fastapi import APIRouter, Depends
from dependencies import verify_internal_token

router = APIRouter(
    prefix="/agents/rca",
    tags=["rca-agent"],
    dependencies=[Depends(verify_internal_token)],  # 社内トークン必須
)

@router.post("/invoke")
async def invoke_rca_agent(...):
    ...
```

```python
# routers/public_agent.py
from fastapi import APIRouter, Depends
from dependencies import verify_api_key

router = APIRouter(
    prefix="/agents/public",
    tags=["public-agent"],
    dependencies=[Depends(verify_api_key)],  # API キー認証
)
```

→ それぞれ別の認証方式を「ルーター単位」で適用できる。`@app.xxx` 方式だと各エンドポイントに `Depends` を書かなければならない。

---

## 8. まとめ

- `@app.xxx` は **小規模・単一ファイル向け**
- `APIRouter` は **構造化・分割・共通設定** のための仕組み
- 共通の `prefix`, `tags`, `dependencies`, `responses` を一括で適用できるのが最大の利点
- バージョニング、認証分離、ドメイン分割など、ほぼすべての中規模以上のプロジェクトで `APIRouter` 一択
- 迷ったら最初から `APIRouter` で書く方が後悔しない