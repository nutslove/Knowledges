- https://aigateway.envoyproxy.io/docs/capabilities/security/
- https://aigateway.envoyproxy.io/docs/capabilities/security/upstream-auth/
- https://gateway.envoyproxy.io/docs/tasks/security/

## OIDC Authentication
- https://gateway.envoyproxy.io/docs/tasks/security/oidc/
- OIDC設定には `SecurityPolicy`というCRを適用する必要がある

> [!CAUTION]
> EnvoyのOIDCは **認可コードフロー** のみサポートしている。  
> クライアントクレデンシャルフローはサポートされていない。

## JWT Authentication
- https://gateway.envoyproxy.io/docs/tasks/security/jwt-authentication/
- 以下はIdPとして Entra ID を使う場合の例
- **基本的な流れ**  
  ```
  クライアント（呼び出し元アプリ）が Entra ID からクライアントクレデンシャルフローでアクセストークンを取得  
    ↓
  Authorization: Bearer <token> で Gateway に送信
    ↓
  Gateway が JWT を検証
    ↓
  問題なければリクエストをAzureなどのプロバイダーに転送
  ```

- JWT Authenticationにも `SecurityPolicy` というCRを適用する必要がある  
  ```yaml
  apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: SecurityPolicy
  metadata:
    name: entra-jwt-auth
  spec:
    targetRefs:
      - group: gateway.networking.k8s.io
        kind: Gateway
        name: envoy-ai-gateway
    jwt:
      providers:
        - name: entra
          issuer: "https://sts.windows.net/<TENANT_ID>/"
          audiences:
            - "api://<APP_ID_URI>"   # 例: api://custom/EnvoyGateway.OIDC
          remoteJWKS:
            uri: "https://login.microsoftonline.com/<TENANT_ID>/discovery/v2.0/keys"
  ```
  - 各フィールドの意味
    - `issuer`: JWTの発行元（Entra ID のテナント発行者URL）。v2 エンドポイントを使う場合は `/v2.0` 付き。トークンの `iss` クレームと一致する必要がある
    - `audiences`: JWTの対象。Entra IDアプリの「App ID URI」を指定。トークンの `aud` クレームと一致する必要があるリスト。Entra で登録した API の Application ID URI（例: `api://...`）または Application (client) ID。
    - `remoteJWKS.uri`: Entra ID の公開鍵セット（JWKS）エンドポイント。Envoy Gateway がここから鍵を取得して署名検証する。

> [!NOTE]
> `issuer`はEntraIDのトークンバージョンがv1.0かv2.0かでURLが異なる点に注意。
> - v1.0: `https://sts.windows.net/<TENANT_ID>/`
> - v2.0: `https://login.microsoftonline.com/<TENANT_ID>/v2.0`
>
> トークンバージョンはEntraIDの該当Applicationの「Manage」→「Manifest」で、
> `api.requestedAccessTokenVersion`で制御する：
> - `null` または `1` → v1.0 トークンが発行される（デフォルト）
> - `2` → v2.0 トークンが発行される
>
> 注意: `requestedAccessTokenVersion` の値と、トークン取得エンドポイント
> （`/oauth2/token` vs `/oauth2/v2.0/token`）は独立している。
> `/oauth2/v2.0/token` を叩いても `requestedAccessTokenVersion: null` なら
> v1.0 トークンが返るため、issuer は `sts.windows.net` 形式になる。
> 確実なのは、実際に取得したトークンを jwt.io でデコードして
> `iss` クレームの実値を確認すること。

### JWT Authenticationを使って Envoy AI Gatewayにリクエストを送る例

#### Curl
- まずは、Entra ID からクライアントクレデンシャルフローでアクセストークンを取得  
  ```bash
  TOKEN=$(curl -s -X POST "https://login.microsoftonline.com/<TENANT_ID>/oauth2/v2.0/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "client_id=<CLIENT_ID>" \
    -d "client_secret=<CLIENT_SECRET>" \
    -d "scope=api://<APP_ID_URI>/.default" \
    -d "grant_type=client_credentials" \
    | jq -r '.access_token')
  ```

> [!NOTE]
> - **scope**は、EntraIDの「App registrations」で該当アプリを選択 → 
>   「Expose an API」タブの上部にある「Application ID URI」（`api://...`）の値を使う
> - クライアントクレデンシャルフローでは末尾に `/.default` を付ける
> - 例: Application ID URI が `api://kinto-technologies.com/envoy-ai-gateway` なら、
>   scope は `api://kinto-technologies.com/envoy-ai-gateway/.default`
> - `.default` は「アプリに事前許可されたすべての権限」を意味する固定値で、
>   **クライアントクレデンシャルフローでは `.default` 固定**
> - 個別スコープ（`api://.../read` など）は委任されたアクセス許可
>   （ユーザー認証用）であり、クライアントクレデンシャルフローでは使わない

> [!NOTE]
> Application ID URI の設定でテナントポリシー違反のエラーが出る場合
> （`Failed to add identifier URI...` のようなエラー）、以下のいずれかで対処:
> - 検証済みドメイン形式: `api://kinto-technologies.com/envoy-ai-gateway`（推奨）
> - GUID 含み: `api://<APP_GUID>/envoy-ai-gateway`
> - マニフェストで `requestedAccessTokenVersion: 2` に変更してから再試行

- 発行したトークンを使って、Gateway のエンドポイントにリクエストを送る  
  ```bash
  curl -v -H "Authorization: Bearer $TOKEN" \
    -H "x-ai-eg-model: some-cool-self-hosted-model" \
    -H "Content-Type: application/json" \
    -d '{
      "model": "some-cool-self-hosted-model",
      "messages": [{"role": "user", "content": "hi"}]
    }' \
    http://localhost:8080/v1/chat/completions
  ```

#### LangChain
```python
import os
import time
import httpx
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser

# === Entra ID トークン取得 ===
TENANT_ID = os.environ["AZURE_TENANT_ID"]
CLIENT_ID = os.environ["AZURE_CLIENT_ID"]
CLIENT_SECRET = os.environ["AZURE_CLIENT_SECRET"]
SCOPE = "api://<APP_ID_URI>/.default"

_token_cache = {"access_token": None, "expires_at": 0}

def get_access_token() -> str:
    now = time.time()
    if _token_cache["access_token"] and _token_cache["expires_at"] > now + 60:
        return _token_cache["access_token"]

    response = httpx.post(
        f"https://login.microsoftonline.com/{TENANT_ID}/oauth2/v2.0/token",
        data={
            "client_id": CLIENT_ID,
            "client_secret": CLIENT_SECRET,
            "scope": SCOPE,
            "grant_type": "client_credentials",
        },
        timeout=10.0,
    )
    response.raise_for_status()
    token_data = response.json()

    _token_cache["access_token"] = token_data["access_token"]
    _token_cache["expires_at"] = now + token_data["expires_in"]
    return token_data["access_token"]

# === LangChain ===
token = get_access_token()

llm = ChatOpenAI(
    #model="claude-haiku-4-5",  # AIGatewayRouteで定義したmodel名
    model="gemini-3.1-flash-lite-preview",  # AIGatewayRouteで定義したmodel名
    #model="gpt-5-4",  # AIGatewayRouteで定義したmodel名
    base_url="http://localhost:8080/v1",  # AI Gatewayのエンドポイント
    api_key="dummy",  # Gateway側で認証するのでdummyでOK(パラメータ自体は必須項目なので必要)
    default_headers={"Authorization": f"Bearer {token}"},
    streaming=True,
)

prompt = ChatPromptTemplate.from_messages([
    ("system", "あなたは親切なアシスタントです。日本語で回答してください。"),
    ("user", "{question}"),
])

chain = prompt | llm | StrOutputParser()

user_input = input("質問を入力してください: ")
for chunk in chain.stream({"question": user_input}):
    print(chunk, end="", flush=True)
print()
```

#### Strands Agents
```python
import os
import time
import asyncio
import httpx
from strands import Agent
from strands.models.openai import OpenAIModel

# === Entra ID トークン取得 ===
TENANT_ID = os.environ["AZURE_TENANT_ID"]
CLIENT_ID = os.environ["AZURE_CLIENT_ID"]
CLIENT_SECRET = os.environ["AZURE_CLIENT_SECRET"]
SCOPE = "api://<APP_ID_URI>/.default"

_token_cache = {"access_token": None, "expires_at": 0}

def get_access_token() -> str:
    now = time.time()
    if _token_cache["access_token"] and _token_cache["expires_at"] > now + 60:
        return _token_cache["access_token"]

    response = httpx.post(
        f"https://login.microsoftonline.com/{TENANT_ID}/oauth2/v2.0/token",
        data={
            "client_id": CLIENT_ID,
            "client_secret": CLIENT_SECRET,
            "scope": SCOPE,
            "grant_type": "client_credentials",
        },
        timeout=10.0,
    )
    response.raise_for_status()
    token_data = response.json()

    _token_cache["access_token"] = token_data["access_token"]
    _token_cache["expires_at"] = now + token_data["expires_in"]
    return token_data["access_token"]

# === Strands Agent ===
token = get_access_token()

model = OpenAIModel(
    client_args={
        "base_url": "http://localhost:8080/v1",  # kubectl port-forward svc/envoy-... 8080:80
        "api_key": "dummy",  # OpenAI SDKの必須パラメータ。Gateway側で認証する場合は実トークンを指定
        "default_headers": {"Authorization": f"Bearer {token}"},
    },
    model_id="claude-haiku-4-5",  # AIGatewayRouteで定義したmodel名(modelNameOverrideのエイリアス)
    #model_id="gemini-3.1-pro-preview",  # AIGatewayRouteで定義したmodel名(modelNameOverrideのエイリアス)
    #model_id="gpt-5-4",  # AIGatewayRouteで定義したmodel名(modelNameOverrideのエイリアス)
)

def silent_callback(**kwargs):
    """何もしない（デフォルトの stdout 出力を抑制）"""
    pass

agent = Agent(
    model=model,
    system_prompt="あなたは親切なアシスタントです。日本語で回答してください。",
    callback_handler=silent_callback,
)

user_input = input("質問を入力してください: ")

# stream_async は非同期ジェネレータでイベントを流してくる
import asyncio

async def main():
    async for event in agent.stream_async(user_input):
        if "data" in event:
            print(event["data"], end="", flush=True)
    print()

asyncio.run(main())
```