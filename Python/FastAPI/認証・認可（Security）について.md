# FastAPIの認証・認可（Security）について

[APIRouterについて](APIRouterについて.md)の`Depends`の説明で `verify_token` / `verify_api_key` / `get_current_user` といった認証系の依存関数が繰り返し登場したが、ここではFastAPIが用意している **`fastapi.security`** モジュール（OAuth2・APIキー・Basic認証などのスキーム）を使った実装方法をまとめる。

> 出典: [FastAPI公式 - Security](https://fastapi.tiangolo.com/tutorial/security/)

---

## 1. なぜ `fastapi.security` を使うのか

認証は「トークンを受け取ってチェックする」だけなら普通の `Depends` 関数でも書けるが、`fastapi.security` のクラスを使うと以下が自動で付いてくる。

- **OpenAPI（Swagger UI）に認証方式が反映される**（`/docs`右上に鍵アイコン・「Authorize」ボタンが出る）
- ヘッダー/クエリ/Cookieからトークンを取り出す定型処理を自前で書かなくてよい
- トークンが無い場合の `401`/`403` を自動で返してくれる

```
[クライアント] → Authorizationヘッダー等にトークンを付与
      ↓
[fastapi.security の Scheme] → トークンを取り出す（Depends経由）
      ↓
[自前の検証ロジック] → トークンをデコード/検証し、ユーザー情報を返す
      ↓
[エンドポイント] → 検証済みユーザー情報を受け取る
```

---

## 2. APIキー認証（`APIKeyHeader` / `APIKeyQuery` / `APIKeyCookie`）

最もシンプルな方式。固定のキーをヘッダー・クエリ・Cookieのいずれかで受け取る。

```python
from fastapi import Depends, FastAPI, HTTPException, Security
from fastapi.security import APIKeyHeader

app = FastAPI()

api_key_header = APIKeyHeader(name="X-API-Key")

VALID_KEYS = {"secret-key-123"}


def verify_api_key(api_key: str = Security(api_key_header)) -> str:
    if api_key not in VALID_KEYS:
        raise HTTPException(status_code=401, detail="Invalid API Key")
    return api_key


@app.get("/protected", dependencies=[Depends(verify_api_key)])
async def protected_route():
    return {"message": "OK"}
```

- `Security(...)` は `Depends(...)` の薄いラッパーで、OAuth2の **スコープ**（後述）を渡せる点だけが違う。スコープ不要なら `Depends()` と同じ感覚で使える
- `APIKeyHeader(name="X-API-Key")` は「`X-API-Key` ヘッダーからキーを取り出す」ことだけを担当し、キーが正しいかどうかの判定は自分で書く

---

## 3. OAuth2 パスワードフロー + JWT（実務で最も一般的な構成）

「ユーザー名/パスワードでトークンを発行し、以後そのトークンで認証する」流れ。`OAuth2PasswordBearer` は**トークンを`Authorization: Bearer <token>`ヘッダーから取り出すだけ**のクラスで、トークンの発行・検証ロジックは自分で実装する。

### 3-1. 全体像

```
1. POST /token  (username, password)
       ↓ 検証OKなら
2. JWTアクセストークンを発行して返す
       ↓ クライアントは以後このトークンを保持
3. GET /users/me  (Authorization: Bearer <token>)
       ↓
4. OAuth2PasswordBearerがトークンを取り出す
       ↓
5. JWTをデコードしてユーザーを特定 → get_current_user
```

### 3-2. トークン発行エンドポイント

```python
from datetime import datetime, timedelta, timezone
from typing import Annotated

import jwt  # PyJWT (pip install pyjwt)
from fastapi import Depends, FastAPI, HTTPException, status
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from pwdlib import PasswordHash  # pip install "pwdlib[argon2]"
from pydantic import BaseModel

SECRET_KEY = "change-me"  # 実運用では環境変数/Secrets Managerから読む
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

app = FastAPI()
password_hash = PasswordHash.recommended()   # 現時点の推奨アルゴリズム（Argon2）を使う

# tokenUrl: Swagger UIの「Authorize」ボタンがトークンを取得しにいくエンドポイント
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

FAKE_USERS_DB = {
    "alice": {"username": "alice", "hashed_password": password_hash.hash("secret")},
}


class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"


def create_access_token(data: dict) -> str:
    expire = datetime.now(timezone.utc) + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    return jwt.encode({**data, "exp": expire}, SECRET_KEY, algorithm=ALGORITHM)


@app.post("/token", response_model=Token)
async def login(form_data: Annotated[OAuth2PasswordRequestForm, Depends()]):
    user = FAKE_USERS_DB.get(form_data.username)
    if not user or not password_hash.verify(form_data.password, user["hashed_password"]):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    token = create_access_token({"sub": user["username"]})
    return Token(access_token=token)
```

- `OAuth2PasswordRequestForm` は `application/x-www-form-urlencoded` の `username` / `password` フィールドを受け取る（**JSONではない**。OAuth2仕様のフォーム形式）
- パスワードは平文で保存・比較しない。ハッシュ化して保存し、検証時は `password_hash.verify()` を使う

> [!WARNING]
> #### パスワードハッシュライブラリは `passlib` ではなく `pwdlib` を使う（2025年時点）
>
> 以前のFastAPI公式チュートリアルは `passlib`（`CryptContext(schemes=["bcrypt"])`）を使っていたが、**`passlib`は2020年以降メンテナンスが止まっており**、`bcrypt>=4.1`と組み合わせると`AttributeError: module 'bcrypt' has no attribute '__about__'`が発生する既知の非互換問題がある（動作自体は継続するが将来的なリスクが高い）。
>
> FastAPI公式チュートリアルも現在は**`pwdlib`（Argon2）**へ移行済み。新規実装では`passlib`を避け、`pwdlib`を使うこと。既存コードで`passlib`を使っている場合は、`bcrypt`のバージョンを`<4.1`に固定するか、`pwdlib`への移行を検討する。

### 3-3. トークンを検証してユーザーを取得する依存関数

```python
def get_current_user(token: Annotated[str, Depends(oauth2_scheme)]) -> dict:
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username = payload.get("sub")
        if username is None:
            raise credentials_exception
    except jwt.PyJWTError:
        raise credentials_exception

    user = FAKE_USERS_DB.get(username)
    if user is None:
        raise credentials_exception
    return user


@app.get("/users/me")
async def read_users_me(current_user: Annotated[dict, Depends(get_current_user)]):
    return {"username": current_user["username"]}
```

- `Depends(oauth2_scheme)` が `Authorization: Bearer <token>` からトークン文字列を取り出す。ヘッダーが無ければ自動で `401` を返す
- `get_current_user` → `oauth2_scheme` というサブ依存の連鎖は、[APIRouterについて](APIRouterについて.md)の「サブ依存」と同じ構造

> [!NOTE]
> `Annotated` で別名を定義しておくと、複数エンドポイントで使い回しやすい（[APIRouterについて](APIRouterについて.md)の「`Annotated`を使う現代的な書き方」と同じ考え方）。
> ```python
> CurrentUser = Annotated[dict, Depends(get_current_user)]
>
> @app.get("/users/me")
> async def read_users_me(current_user: CurrentUser):
>     return current_user
> ```

---

## 4. スコープ（権限の範囲）を使う場合

`OAuth2PasswordBearer` の代わりに `scopes` を宣言できる形式もある。「このエンドポイントは `items:read` スコープが必要」のように**権限を細分化**したいときに使う。単純なロールチェックだけなら、依存関数内で `if user.role != "admin": raise HTTPException(403)` のように書く方が簡単なことも多い。

```python
from fastapi.security import OAuth2PasswordBearer, SecurityScopes

oauth2_scheme = OAuth2PasswordBearer(
    tokenUrl="token",
    scopes={"items:read": "アイテムの閲覧", "items:write": "アイテムの編集"},
)


def get_current_user(
    security_scopes: SecurityScopes,
    token: Annotated[str, Depends(oauth2_scheme)],
):
    # security_scopes.scopes にエンドポイント側で要求されたスコープ一覧が入る
    ...


@app.get("/items/")
async def read_items(
    user: Annotated[dict, Security(get_current_user, scopes=["items:read"])],
):
    ...
```

---

## 5. Basic認証（`HTTPBasic`）

簡易な用途（社内ツール、管理画面など）向け。ブラウザ標準のBasic認証ダイアログが出る。

```python
import secrets
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBasic, HTTPBasicCredentials

security = HTTPBasic()


def verify_credentials(credentials: Annotated[HTTPBasicCredentials, Depends(security)]):
    correct = secrets.compare_digest(credentials.password, "secret")  # タイミング攻撃対策
    if not (credentials.username == "admin" and correct):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect credentials",
            headers={"WWW-Authenticate": "Basic"},
        )
    return credentials.username
```

> [!WARNING]
> 文字列比較に `==` を使うと、比較にかかる時間の差から情報が漏れる**タイミング攻撃**の余地がある。パスワード等の比較には `secrets.compare_digest()` を使うこと。

---

## 6. どの方式を選ぶか

| 方式 | クラス | 向いている用途 |
|---|---|---|
| APIキー | `APIKeyHeader` / `APIKeyQuery` / `APIKeyCookie` | サーバー間連携、シンプルな外部公開API |
| OAuth2 + JWT | `OAuth2PasswordBearer` | ユーザーログインを伴うWebアプリ・SPA |
| Basic認証 | `HTTPBasic` | 社内ツール・管理画面など簡易な用途 |
| 独自ヘッダー + 自前検証 | 普通の `Depends` 関数 | 社内トークン、Keystoneトークン等、既存の認証基盤に合わせる場合 |

> [!NOTE]
> [APIRouterについて](APIRouterについて.md)の例に出てきた `verify_internal_token` / `verify_api_key` のような「ヘッダーを直接読んで検証するだけの`Depends`関数」も**それ自体で正しく機能する**。`fastapi.security`のクラスを使うのは主に「OpenAPI上に認証UIを出したい」「OAuth2/JWTのような標準フローに乗りたい」場合であり、必須ではない。

---

## 7. `dependencies=[...]` に置くか、引数で受け取るか

これは[APIRouterについて](APIRouterについて.md)の「適用レベル」の話がそのまま当てはまる。

| やりたいこと | 書き方 |
|---|---|
| 認証は必須だが、ユーザー情報はエンドポイント内で使わない | `dependencies=[Depends(verify_api_key)]` |
| 認証済みユーザーの情報（`user_id`など）をロジックで使いたい | `user: CurrentUser`（引数として受け取る） |
| ルーター配下すべてに同じ認証をかけたい | `APIRouter(dependencies=[Depends(...)])` |

---

## ポイントまとめ

- `fastapi.security` のクラス（`APIKeyHeader`, `OAuth2PasswordBearer`, `HTTPBasic` 等）は「トークンをどこから取り出すか」を定型化し、**OpenAPIに認証方式を反映**してくれるもの。トークンの正当性チェック自体は自分で実装する
- OAuth2 + JWTが最も一般的な構成: `/token` でパスワード検証してJWT発行 → 以後 `Authorization: Bearer <token>` で認証
- パスワードは`pwdlib`（Argon2）等でハッシュ化して保存・検証する（`passlib`は非推奨・メンテナンス停止済み）。平文比較・単純な`==`比較は避ける（タイミング攻撃対策には`secrets.compare_digest`）
- 認証結果を使わない（ガードだけ）なら `dependencies=[...]`、使うなら引数で受け取る（[APIRouterについて](APIRouterについて.md)の使い分けと同じ）
- シンプルな社内トークン検証なら、`fastapi.security`を使わず普通の`Depends`関数でも十分
