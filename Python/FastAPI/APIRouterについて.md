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

## 4. `Depends`（依存性注入）とは

`Depends` は FastAPI の **依存性注入（Dependency Injection, DI）** の仕組み。「エンドポイントを実行する前に共通の処理（関数）を呼び出し、その戻り値を引数として受け取る」ことができる。認証・DBセッション・共通パラメータ・設定取得などを **再利用可能・テスト可能** な形で切り出すために使う。

### 基本

```python
from fastapi import Depends, FastAPI

app = FastAPI()


# 依存関数（ただの呼び出し可能オブジェクト）
def pagination(skip: int = 0, limit: int = 10) -> dict:
    return {"skip": skip, "limit": limit}


@app.get("/users")
async def get_users(page: dict = Depends(pagination)):
    # /users?skip=20&limit=5 → page = {"skip": 20, "limit": 5}
    return page
```

- `Depends(pagination)` を書くと、FastAPIがリクエストごとに `pagination` を呼び、戻り値を `page` に注入する
- `pagination` のクエリパラメータ（`skip`, `limit`）はそのままエンドポイントのクエリパラメータとしてOpenAPIにも反映される
- 同じ依存を複数のエンドポイントで使い回せる

> **補足: `Depends` は「引数」に注入するだけで、レスポンスには関与しない**
>
> `Depends(pagination)` がするのは、`pagination()` の戻り値を **引数 `page` に渡す** ことだけ。レスポンスに自動で追記したり置き換えたりはしない。上の例で `page` がそのままレスポンスになっているのは、エンドポイントが `return page` と **明示的に返している** からにすぎない。**何がレスポンスになるかは「`return` で何を返すか」だけで決まり、`Depends` とは無関係**。

### 実際の使い方（`page` は入力として使う）

現実には `page`（skip/limit）は「DBから何件目をどれだけ取るか」の**入力**として使い、レスポンスは自分で組み立てて返す。

```python
# ダミーデータ（本来はDBから取得）
FAKE_USERS = [{"id": i, "name": f"user{i}"} for i in range(100)]


def pagination(skip: int = 0, limit: int = 10) -> dict:
    return {"skip": skip, "limit": limit}


@app.get("/users")
async def get_users(page: dict = Depends(pagination)):
    # page は「取得範囲」を決めるための入力として使う
    start = page["skip"]
    end = page["skip"] + page["limit"]
    users = FAKE_USERS[start:end]

    # レスポンスは自分で組み立てたもの。page そのものではない
    return {
        "items": users,
        "skip": page["skip"],
        "limit": page["limit"],
        "total": len(FAKE_USERS),
    }
```

リクエスト `/users?skip=20&limit=5` のレスポンス例:

```json
{
  "items": [
    {"id": 20, "name": "user20"},
    {"id": 21, "name": "user21"},
    {"id": 22, "name": "user22"},
    {"id": 23, "name": "user23"},
    {"id": 24, "name": "user24"}
  ],
  "skip": 20,
  "limit": 5,
  "total": 100
}
```

- `page`（= `{"skip": 20, "limit": 5}`）は **取得範囲を決める入力** として使っているだけ
- レスポンスに `skip` / `limit` が含まれるのは、こちらが `return` の中に **明示的に入れた** から。`Depends` が勝手に追記したわけではない
- もし `return users` とだけ書けば、レスポンスは `items` の配列だけになる（`page` はどこにも現れない）

### なぜ使うのか

- **共通ロジックの再利用**: 認証チェックやページネーションを各所にコピペせず1箇所に
- **関心の分離**: エンドポイントは「やりたいこと」だけに集中し、前処理は依存に逃がす
- **テストしやすさ**: `app.dependency_overrides` で依存を差し替えれば、認証やDBをモックできる

### 適用レベル（どこに書くか）

`Depends` は4つのレベルで適用できる。**戻り値を使うかどうか**で書き方が変わる。

| レベル | 書き方 | 戻り値 | 用途 |
|---|---|---|---|
| エンドポイント引数 | `def ep(x = Depends(dep))` | **使う** | DBセッション、現在のユーザー、共通パラメータ |
| デコレータ | `@router.get(..., dependencies=[Depends(dep)])` | 使わない | 副作用だけ（認証・レート制限・監査ログ） |
| ルーター全体 | `APIRouter(dependencies=[Depends(dep)])` | 使わない | ルーター内の全エンドポイントに認証等を一括適用 |
| アプリ / include_router | `include_router(..., dependencies=[...])` / `FastAPI(dependencies=[...])` | 使わない | アプリ全体・特定ルーター群への一括適用 |

戻り値が欲しいとき（現在のユーザーなど）は**引数**で受け取り、戻り値が不要な単なるガード（認証で弾くだけ）は **`dependencies=[...]`** に置く、と覚えるとよい。セクション3・7で出てきた `dependencies=[Depends(verify_token)]` は後者（認証ガード）。

### `yield` による後処理（セットアップ / クリーンアップ）

`return` の代わりに `yield` を使うと、エンドポイント実行後に後処理を走らせられる。DBセッションのクローズなどに必須。

```python
def get_db():
    db = SessionLocal()
    try:
        yield db          # ← ここが注入される
    finally:
        db.close()        # ← レスポンス返却後に必ず実行される


@app.get("/items")
async def list_items(db = Depends(get_db)):
    return db.query(...).all()
```

### サブ依存（依存の中で依存を使う）

依存関数自身がさらに `Depends` を持てる。認証の積み上げなどに便利。

```python
from fastapi import Header, HTTPException


def get_token(authorization: str = Header()) -> str:
    return authorization.removeprefix("Bearer ")


def get_current_user(token: str = Depends(get_token)) -> User:
    user = decode_user(token)
    if user is None:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user


@app.get("/me")
async def me(user: User = Depends(get_current_user)):
    return user
```

`me` → `get_current_user` → `get_token` と連鎖して解決される。

### クラスベースの依存

`__init__` を持つクラスもそのまま依存にできる（呼び出し可能なため）。パラメータを保持したい依存に向く。

```python
class RateLimiter:
    def __init__(self, max_calls: int) -> None:
        self.max_calls = max_calls

    def __call__(self, user: User = Depends(get_current_user)) -> None:
        if over_limit(user, self.max_calls):
            raise HTTPException(status_code=429, detail="Too Many Requests")


@app.get("/heavy", dependencies=[Depends(RateLimiter(max_calls=10))])
async def heavy():
    ...
```

### 依存のキャッシュ

**同一リクエスト内**で同じ依存が複数回要求されても、デフォルトでは **1回しか実行されず結果がキャッシュ** される（サブ依存で共有される）。毎回実行したい場合は `Depends(dep, use_cache=False)` を指定する。

### `Annotated` を使う現代的な書き方（推奨）

FastAPI公式は `Annotated` での記法を推奨している。依存を「型」として定義でき、複数エンドポイントで使い回しやすい。

```python
from typing import Annotated
from fastapi import Depends

CurrentUser = Annotated[User, Depends(get_current_user)]
DbSession = Annotated[Session, Depends(get_db)]


@app.get("/me")
async def me(user: CurrentUser):
    return user


@app.get("/items")
async def list_items(user: CurrentUser, db: DbSession):
    ...
```

- デフォルト引数（`= Depends(...)`）方式より引数の見通しがよく、再利用しやすい
- `CurrentUser` のような別名を定義しておけば、各エンドポイントは型注釈を書くだけで依存が効く

---

## 5. 推奨ディレクトリ構成

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

## 6. ルーターのネスト（router の中に router を含める）

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

## 7. いつ `APIRouter` を使うべきか

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

## 8. 認証を agent ごとに変える例

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

## 9. まとめ

- `@app.xxx` は **小規模・単一ファイル向け**
- `APIRouter` は **構造化・分割・共通設定** のための仕組み
- 共通の `prefix`, `tags`, `dependencies`, `responses` を一括で適用できるのが最大の利点
- `dependencies` に渡す `Depends`（依存性注入）は、認証・DBセッション・共通パラメータを再利用可能な形で切り出す仕組み（セクション4）。戻り値が要るなら引数で、ガードだけなら `dependencies=[...]` で適用する
- バージョニング、認証分離、ドメイン分割など、ほぼすべての中規模以上のプロジェクトで `APIRouter` 一択
- 迷ったら最初から `APIRouter` で書く方が後悔しない