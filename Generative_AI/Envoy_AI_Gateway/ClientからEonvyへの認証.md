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
- JWT Authenticationでは、クライアントが事前に Entra ID から取得した
  アクセストークンを `Authorization: Bearer <token>` で Gateway に送り、
  Gateway 側で JWT を検証する。トークン取得のフロー（クライアントクレデンシャル / 認可コード）は
  クライアント側の選択であり、Gateway側のJWT検証ロジックは共通。

### EntraIDのアプリケーションについて
- EntraIDで「アプリケーション」というと、`Microsoft Entra ID` → `App registrations` で登録するもの。
- 本まとめでは、用途別に2種類のEntraIDアプリケーションを区別する：
  - **API側のEntraIDアプリケーション**（Envoy AI GatewayがAPIを公開する側）  
    例: `envoy-ai-gateway` という名前で登録するEntraIDアプリケーション
  - **クライアント側のEntraIDアプリケーション**（Envoy AI Gatewayを呼び出す側）  
    例: バッチサービス用、Webアプリ用などのEntraIDアプリケーション

> [!NOTE]
> EntraIDの認可モデル（App roles、Delegated permissions、Application permissions、Expose an API のスコープ等）の
> 詳細については別ドキュメント [EntraID認可について.md] を参照。
> 本まとめでは Envoy AI Gateway の文脈に絞って説明する。

### EntraIDアプリケーションの構成パターン

API側とクライアント側のEntraIDアプリケーションは、**1つにまとめても、別々に作っても動く**。  
用途・規模に応じて選ぶ。

#### パターン1: 単一のEntraIDアプリケーションで兼任

検証目的やシンプルな構成向け。1つのEntraIDアプリケーションが「APIを公開する」「APIを呼ぶ」の両役割を担う。

##### 最小構成（JWT検証のみで運用）
```
EntraIDアプリケーション (envoy-ai-gateway): 1つ
  ├─ Expose an API → Application ID URI: api://kinto-technologies.com/envoy-ai-gateway
  └─ Certificates & secrets → Client secret
```

これだけでクライアントクレデンシャルフロー（および認可コードフロー）は動く。  
`scope=api://.../.default` でトークン取得 → Gateway側のJWT検証は `iss` + `aud` + 署名チェックだけなのでパスする。  
ただしトークンに `roles` クレームは入らないため、ロールベース認可は不可。

##### ロールベース認可付き構成（クライアントクレデンシャルフロー用）
```
EntraIDアプリケーション (envoy-ai-gateway): 1つ
  ├─ Expose an API → Application ID URI: api://kinto-technologies.com/envoy-ai-gateway
  ├─ App roles → App role: Gateway.Access (Allowed member types: Applications)
  ├─ Certificates & secrets → Client secret
  └─ API permissions → Application permissions: 自分自身の Gateway.Access
      └─ Grant admin consent
```

App roles を定義し、Application permissions として自分自身に付与することで、
取得したトークンに `roles: ["Gateway.Access"]` が含まれるようになる。  
将来 `SecurityPolicy.authorization` でロールベース認可を入れたいならこの構成。

##### 認可コードフロー併用構成
```
EntraIDアプリケーション (envoy-ai-gateway): 1つ
  ├─ Expose an API → Application ID URI: api://kinto-technologies.com/envoy-ai-gateway
  ├─ Expose an API → Scope: Gateway.Access (認可コードフロー用)
  ├─ Authentication → Web platform + Redirect URI
  ├─ Certificates & secrets → Client secret
  └─ API permissions → Delegated permissions: 自分自身の Gateway.Access
      └─ Grant admin consent
```

> [!NOTE]
> パターン1で **同じアプリにDelegated permissions と Application permissions の両方を入れる**のは、
> 後述の Azure ポータルUI制約により設定が困難。両方欲しい場合は、
> パターン2（用途別にクライアント側のEntraIDアプリケーションを分ける）への切替を推奨。

**動作:**
- 「アプリが自分自身のAPIを呼ぶ」形になる
- トークン取得すると `aud`（=自分のApplication ID URI）と `appid`（=自分のCLIENT_ID）が両方とも同じアプリを指す
- Envoy Gateway 側のJWT検証は普通にパスする

**向いているケース:**
- PoC、検証、個人プロジェクト
- 呼び出し元クライアントが1種類しかない
- 「とりあえず動かしたい」

**向いていないケース:**
- 複数のクライアント（バッチA、バッチB、CIなど）が呼ぶ → 全部同じCLIENT_IDを使うので識別困難
- 認可コードフローも併用したい → 同一アプリでDelegated/Application両permissionの設定問題に当たる
- セキュリティ境界を明確にしたい本番運用

#### パターン2: API側 + クライアント側の2アプリケーション構成（本番推奨）

本番運用・スケールする構成向け。役割を明確に分離する。

```
API側のEntraIDアプリケーション (envoy-ai-gateway):
  ├─ Expose an API → Application ID URI: api://kinto-technologies.com/envoy-ai-gateway
  ├─ App roles → App role: Gateway.Access (Applications)  ← クライアントクレデンシャル用 (任意、roles認可入れる場合)
  └─ Expose an API → Scope: Gateway.Access                ← 認可コード用 (認可コード使う場合)

クライアント側のEntraIDアプリケーションA (バッチ用):
  ├─ Certificates & secrets → Client secret
  └─ API permissions → Application permissions: API側の Gateway.Access (任意)
      └─ Grant admin consent

クライアント側のEntraIDアプリケーションB (Webアプリ用):
  ├─ Authentication → Web platform + Redirect URI
  ├─ Certificates & secrets → Client secret
  └─ API permissions → Delegated permissions: API側の Gateway.Access
      └─ Grant admin consent
```

**動作:**
- API側とクライアント側で `appid` が異なるため、「どのクライアントが呼んだか」が監査ログで明確に判別できる
- クライアントごとにClient secretが独立 → 漏洩時の影響範囲を限定可能
- 認可コードフローとクライアントクレデンシャルフローを別々のクライアントアプリで運用可能

**向いているケース:**
- 本番運用
- 複数のクライアントサービスが呼ぶ
- 認可コードフロー + クライアントクレデンシャルフローを併用する
- セキュリティ境界・監査要件がある

#### 比較表

| 項目 | パターン1 (単一) | パターン2 (別々) |
|---|---|---|
| 設定の手数 | ✅ 少ない（1個分） | ❌ 多い（2個分以上） |
| Client secret 管理 | ✅ 1つだけ | ❌ クライアント数だけ管理 |
| 概念的な分かりやすさ | ⚠️ 「APIと呼び出し元が同じ」は直感的でない | ✅ 役割が明確 |
| 監査ログでの識別 | ❌ `appid` で「どのクライアントか」を区別不可 | ✅ `appid` で識別可能 |
| 複数クライアント対応 | ❌ 全クライアントが同じCLIENT_IDを共有 | ✅ クライアントごとに分離 |
| Secret漏洩時の影響範囲 | ❌ API側の設定にも影響 | ✅ 該当クライアントのみ無効化で済む |
| 認可コードフローとの共存 | ❌ Delegated/Application両permissionの設定問題が再発しやすい | ✅ クライアントアプリを分けて共存可能 |

### Callback URL（Redirect URI）はどちらのEntraIDアプリケーションに設定するか

認可コードフローを使う場合、**Redirect URI は「クライアント側のEntraIDアプリケーション」に設定する**。  
API側（Envoy AI Gateway用）には設定不要。

#### 理由：Redirect URI の役割を踏まえると自明

Redirect URI は OAuth 2.0 の認可コードフローにおいて、**「認可エンドポイントが認可コードを返す宛先」** を指す。
認可フローを開始するのはクライアントアプリ自身であり、認可コードを受け取って
アクセストークンに交換するのもクライアントアプリ。したがって Redirect URI は
**認可フローを実行するクライアント側のEntraIDアプリケーション**に紐付ける必要がある。

一方、Envoy AI Gateway（API側）は OAuth 認可フローには一切関与せず、
クライアントから受け取った JWT を `SecurityPolicy.jwt` で検証するだけ。
したがって API側のEntraIDアプリケーションに Redirect URI を設定する必要はない（設定しても使われない）。

```
[ユーザーのブラウザ]
    │
    │ 1. クライアントアプリにアクセス
    ↓
[クライアントアプリ]
    │
    │ 2. EntraID認可エンドポイントへリダイレクト
    │    client_id=<クライアント側のEntraIDアプリケーションのCLIENT_ID>
    │    redirect_uri=http://localhost:8765/callback  ← クライアント側に登録したURI
    │    scope=api://<APP_ID_URI>/Gateway.Access
    ↓
[EntraID 認可エンドポイント]
    │
    │ 3. ユーザー認証 + 同意
    │ 4. redirect_uri に認可コードを付けてリダイレクト
    ↓
[クライアントアプリの /callback]
    │
    │ 5. 認可コードをアクセストークンに交換
    │ 6. Bearer トークンで Envoy AI Gateway を呼ぶ
    ↓
[Envoy AI Gateway]
    │ JWT検証のみ（Redirect URI は無関係）
    ↓
[LLMプロバイダー]
```

#### 設定項目の対応表

| 設定項目 | API側 (envoy-ai-gateway) | クライアント側 (Webアプリ等) |
|---|---|---|
| Application ID URI | ✅ 設定する（例: `api://kinto-technologies.com/envoy-ai-gateway`） | ❌ |
| Expose an API → Scope | ✅ 設定する（例: `Gateway.Access`） | ❌ |
| App roles | ✅ 設定する（クライアントクレデンシャルでroles使う場合） | ❌ |
| Authentication → Platform | ❌ | ✅ Web または Mobile and desktop |
| **Redirect URI** | ❌ | ✅ 設定する（例: `http://localhost:8765/callback`） |
| Client secret | ❌（通常不要） | ✅ 設定する（Web プラットフォームの場合必須） |
| API permissions | ❌ | ✅ API側のScope/App roleを参照 |

#### クライアントクレデンシャルフロー併用時の補足

パターン2（API側 + 用途別のクライアント側）構成で、認可コードフローと
クライアントクレデンシャルフローを併用する場合：

- **バッチ用クライアントアプリ（クライアントクレデンシャル）**: Redirect URI **不要**。
  トークンエンドポイントを直接叩くだけなのでリダイレクトが発生しない
- **Webアプリ用クライアントアプリ（認可コード）**: Redirect URI **必須**

つまり Redirect URI を設定するのは **「認可コードフローを使うクライアント側のEntraIDアプリケーションだけ」**。

#### パターン1（単一アプリ兼任）の場合

API側＝クライアント側＝同じアプリなので、**その単一アプリに Redirect URI を設定する**。
「API側に設定する」というよりは「クライアント役割を兼ねているから設定する」と捉えると整理しやすい。

#### 複数のRedirect URIを登録するケース

EntraIDは1つのアプリに複数のRedirect URIを登録できる（最大256個）。
ローカル開発・ステージング・本番で別URIになる場合、同じクライアント側アプリに
複数登録するのが一般的：

```
http://localhost:8765/callback                # ローカル開発
https://myapp-staging.example.com/callback    # ステージング
https://myapp.example.com/callback            # 本番
```

ただし、**環境ごとにクライアント側EntraIDアプリケーション自体を分ける**運用もある
（Client secret 漏洩時の影響範囲を環境単位で限定したい場合や、テナント分離要件がある場合）。

> [!NOTE]
> ローカル開発で `http://localhost` を使うのは EntraID 側で例外的に許可されているが、
> 本番では必ず `https://` を使うこと（EntraIDのポリシーで強制される）。
> ワイルドカードURI（`https://*.example.com/callback` のような形）は基本的に使えない。

### API側とクライアント側の「紐づけ」について

EntraID上には**明示的な「紐づけ」ボタンや操作はない**。  
2つのEntraIDアプリケーションを使う場合（パターン2）、紐づけは以下の3要素の組み合わせで自動的に成立する：

1. **クライアント側の API permissions 追加**  
   `API permissions` → `Add a permission` → `My APIs` → API側のアプリ選択 → 必要なロール/スコープを追加  
   → クライアント側アプリのマニフェストに「API側アプリのCLIENT_ID + ロール/スコープのID」が記録される
```json
   // クライアント側アプリのマニフェスト (抜粋)
   "requiredResourceAccess": [
     {
       "resourceAppId": "<API側アプリのCLIENT_ID>",  // ← ここでAPI側を参照
       "resourceAccess": [
         {
           "id": "<App roleまたはScopeのGUID>",
           "type": "Role"   // Application permission (クライアントクレデンシャル用)
                            // Delegated permission (認可コード用) の場合は "Scope"
         }
       ]
     }
   ]
```

   > [!NOTE]
   > マニフェストの `type` の値とAzureポータルの画面表示の対応：
   > - 画面表示`Application` ↔ マニフェスト `"Role"`
   > - 画面表示`Delegated` ↔ マニフェスト `"Scope"`

2. **Grant admin consent**  
   管理者が「クライアント側にAPI側のロール/スコープを与える」ことを明示承認  
   → テナントレベルの consent 記録が作成され、紐づけが**実効性を持つ**

3. **トークン取得時の scope 指定**  
   クライアント側のコードで `scope=api://<API側のApp ID URI>/.default` を指定  
   → EntraIDがこのURIから「どのAPI向けのトークンか」を判定してトークンを発行  
   → 発行されたトークンの `aud` がAPI側のApp ID URIになり、Gateway側で検証される

つまり「紐づけ」は単一の操作ではなく、上記の組み合わせとして分散して実現されている。  
パターン1（単一アプリ）の場合は、これらが全部同じアプリ内で完結するため、表面的には「紐づけ操作」が不要に見える（実質的には自分自身に対する設定）。

### 共通: SecurityPolicy 設定
- JWT Authenticationには `SecurityPolicy` というCRを適用する  
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
            - "api://<APP_ID_URI>"   # 例: api://kinto-technologies.com/envoy-ai-gateway
          remoteJWKS:
            uri: "https://login.microsoftonline.com/<TENANT_ID>/discovery/v2.0/keys"
```
  - 各フィールドの意味
    - `issuer`: JWTの発行元（Entra ID のテナント発行者URL）。トークンの `iss` クレームと一致する必要がある
    - `audiences`: JWTの対象。**API側のEntraIDアプリケーション**の「Application ID URI」を指定。トークンの `aud` クレームと一致する必要があるリスト
    - `remoteJWKS.uri`: Entra ID の公開鍵セット（JWKS）エンドポイント。Envoy Gateway がここから鍵を取得して署名検証する

> [!NOTE]
> パターン1（単一アプリ）の場合、`audiences` には**そのアプリ自身**のApplication ID URIを指定する。  
> パターン2（別々）の場合、`audiences` には**API側のアプリ**のApplication ID URIを指定する。  
> どちらにしても、`audiences` はトークンの `aud` クレームと一致する必要がある（jwt.ioで実値を確認推奨）。

> [!NOTE]
> `issuer`はEntraIDのトークンバージョンがv1.0かv2.0かでURLが異なる点に注意。
> - v1.0: `https://sts.windows.net/<TENANT_ID>/`
> - v2.0: `https://login.microsoftonline.com/<TENANT_ID>/v2.0`
>
> トークンバージョンは**API側のEntraIDアプリケーション**の「Manage」→「Manifest」で、
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

> [!NOTE]
> Application ID URI の設定でテナントポリシー違反のエラーが出る場合
> （`Failed to add identifier URI...` のようなエラー）、以下のいずれかで対処:
> - 検証済みドメイン形式: `api://kinto-technologies.com/envoy-ai-gateway`（推奨）
> - GUID 含み: `api://<APP_GUID>/envoy-ai-gateway`
> - マニフェストで `requestedAccessTokenVersion: 2` に変更してから再試行

---

## クライアントクレデンシャルフロー版

### ユースケース
- M2M（マシン間）通信
- バッチジョブ、サービス間通信
- ユーザー識別が不要な自動化処理
- CI/CDパイプライン

### 特徴
- ブラウザ不要
- `client_secret` で認証
- トークンの `sub` / `oid` は **呼び出し元のEntraIDアプリケーション（クライアント側）** のObject ID（ユーザー情報なし）
- トークンの `appid` は **呼び出し元のEntraIDアプリケーション（クライアント側）** のApplication (client) ID
- トークンに `roles` クレーム（アプリロール）を含めることが可能（App role + Application permission の設定が必要）
- `idtyp` クレームが `"app"` になる（これでクライアントクレデンシャル由来か判別可能）

### EntraID側の設定

クライアントクレデンシャルフローを動かすだけなら **App roles / API permissions の設定は不要**。
ただし将来 `SecurityPolicy.authorization` でロールベース認可を入れたい場合は、追加で App role + Application permission の設定が必要。

#### 最小構成（JWT検証のみ）

JWT検証（`iss` + `aud` + 署名）だけで認可制御は行わない、という前提なら以下の最小設定で動く。

##### API側のEntraIDアプリケーション
1. **Expose an API**: Application ID URI を設定（例: `api://kinto-technologies.com/envoy-ai-gateway`）

##### クライアント側のEntraIDアプリケーション
1. **Certificates & secrets**: Client secret を作成 → 値を控える

> [!NOTE]
> 最小構成では API permissions / App roles は不要。  
> `scope=api://.../.default` でトークン取得すると、トークンの `aud` がAPI側のApp ID URIに設定され、
> Gateway側のJWT検証はパスする。
> ただしトークンに `roles` クレームは入らないため、`SecurityPolicy.authorization` で
> ロールベース認可を書きたい場合は次の「ロールベース認可付き構成」を採用する。

> [!NOTE]
> パターン1（単一アプリ）の場合：API側＝クライアント側＝同じアプリで上記設定を行う。  
> パターン2（別々）の場合：上記の通りAPI側とクライアント側それぞれを設定する。

#### ロールベース認可付き構成

トークンに `roles` クレームを含めて、Gateway側で認可制御したい場合。

##### API側のEntraIDアプリケーション
1. **Expose an API**: Application ID URI を設定
2. **App roles**: アプリロールを定義
   - **Display name**: 例 `Gateway Access`
   - **Allowed member types**: **Applications** を選択（Users/GroupsやBothではない）
   - **Value**: 例 `Gateway.Access`（トークンの `roles` クレームに入る値）
   - **Description**: 説明
   - **Do you want to enable this app role?**: チェック

##### クライアント側のEntraIDアプリケーション
1. **Certificates & secrets**: Client secret を作成 → 値を控える
2. **API permissions** → Add a permission → **My APIs** → API側のEntraIDアプリケーションを選択
   - **Application permissions** タブを選択（**Delegated permissions ではない**）
   - 上で定義したアプリロール（例 `Gateway.Access`）にチェック
3. **Grant admin consent** を実行

> [!NOTE]
> パターン1（単一アプリ）の場合：  
> 同じアプリ内で、`Expose an API` で Application ID URI を設定 → `App roles` を定義 → 
> 同じアプリ内の `API permissions` で `My APIs` から自分自身を選択 → 自分のApp roleを選択 → Grant admin consent。  
> 「自分自身のロールを自分のアプリに付与する」という見た目になる。

> [!CAUTION]
> 「Add a permission」→「My APIs」で **Application permissions タブがグレーアウト**している場合は、
> API側のApp roles で **Allowed member types: Applications**（または Both）のロールが定義されていない可能性が高い。  
> API側で App role を定義してから、ブラウザをハードリロード（Ctrl+Shift+R）して再度試す。

### 基本的な流れ
```
クライアント側のEntraIDアプリケーションが Entra ID からクライアントクレデンシャルフローでアクセストークンを取得  
  ↓
Authorization: Bearer <token> で Gateway に送信
  ↓
Gateway が JWT を検証
  ↓
問題なければリクエストをAzureなどのプロバイダーに転送
```

### Curl
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
> - **client_id / client_secret**: **クライアント側のEntraIDアプリケーション**の値を使用  
>   （パターン1の単一アプリ構成では、API側＝クライアント側＝同じアプリのCLIENT_ID/secret）
> - **scope**: **API側のEntraIDアプリケーション**の「Expose an API」タブで設定した
>   Application ID URI（`api://...`）の末尾に `/.default` を付ける
> - 例: API側のEntraIDアプリケーションの Application ID URI が
>   `api://kinto-technologies.com/envoy-ai-gateway` なら、
>   scope は `api://kinto-technologies.com/envoy-ai-gateway/.default`
> - `.default` は「アプリに事前許可されたすべての権限」を意味する固定値で、
>   **クライアントクレデンシャルフローでは `.default` 固定**
> - 個別スコープ（`api://.../read` など）は委任されたアクセス許可
>   （ユーザー認証用）であり、クライアントクレデンシャルフローでは使わない

> [!NOTE]
> <CLIENT_SECRET>は、Azureの**クライアント側のEntraIDアプリケーション**を選択 → 
> 「Certificates & secrets」タブ → 「Client secrets」の「Value」を確認

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

### LangChain
```python
import os
import time
import httpx
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser

# === Entra ID トークン取得 ===
# AZURE_TENANT_ID: EntraIDテナントID
# AZURE_CLIENT_ID: クライアント側のEntraIDアプリケーションのApplication (client) ID
#                  (パターン1単一アプリ構成の場合は、その単一アプリのCLIENT_ID)
# AZURE_CLIENT_SECRET: クライアント側のEntraIDアプリケーションのClient secret
TENANT_ID = os.environ["AZURE_TENANT_ID"]
CLIENT_ID = os.environ["AZURE_CLIENT_ID"]
CLIENT_SECRET = os.environ["AZURE_CLIENT_SECRET"]
# SCOPE: API側のEntraIDアプリケーションのApplication ID URI + /.default
#        (パターン1単一アプリ構成の場合は、その単一アプリのApplication ID URI)
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

### Strands Agents
```python
import os
import time
import asyncio
import httpx
from strands import Agent
from strands.models.openai import OpenAIModel

# === Entra ID トークン取得 ===
# AZURE_TENANT_ID: EntraIDテナントID
# AZURE_CLIENT_ID: クライアント側のEntraIDアプリケーションのApplication (client) ID
#                  (パターン1単一アプリ構成の場合は、その単一アプリのCLIENT_ID)
# AZURE_CLIENT_SECRET: クライアント側のEntraIDアプリケーションのClient secret
TENANT_ID = os.environ["AZURE_TENANT_ID"]
CLIENT_ID = os.environ["AZURE_CLIENT_ID"]
CLIENT_SECRET = os.environ["AZURE_CLIENT_SECRET"]
# SCOPE: API側のEntraIDアプリケーションのApplication ID URI + /.default
#        (パターン1単一アプリ構成の場合は、その単一アプリのApplication ID URI)
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

async def main():
    async for event in agent.stream_async(user_input):
        if "data" in event:
            print(event["data"], end="", flush=True)
    print()

asyncio.run(main())
```

---

## 認可コードフロー版

### ユースケース
- ユーザー認証が必要なWebアプリ
- 社内向けAIツール（誰がリクエストしたか追跡したい）
- Slack/Teams Bot 経由でのAI利用（ユーザーIDの伝搬）
- ユーザーごとの利用ログ・コスト配賦

### 特徴
- ブラウザでのユーザーログインが必要
- トークンの `sub` / `oid` は**ログインしたユーザー**のObject ID
- トークンの `appid` は **呼び出し元のEntraIDアプリケーション（クライアント側）** のApplication (client) ID
- トークンに `name`, `preferred_username`（メアド）などユーザー情報を含む
- トークンに `scp` クレーム（委任されたスコープ）を含む
- Gateway 側の `SecurityPolicy.jwt` 設定は**クライアントクレデンシャル版と全く同じでOK**

### EntraID側の設定

#### API側のEntraIDアプリケーション
1. **Expose an API**: Application ID URI を設定（クライアントクレデンシャル版と同じでOK）
2. **Add a scope**: 個別スコープを定義（**認可コードフローでは必須**）
   - **Scope name**: 例 `Gateway.Access` または `user_impersonation`
   - **Who can consent?**: `Admins and users`
   - **Admin consent display name**: 例 `Access Envoy AI Gateway`
   - **Admin consent description**: 適切な説明文
   - **State**: `Enabled`
   - フルスコープ例: `api://kinto-technologies.com/envoy-ai-gateway/Gateway.Access`

#### クライアント側のEntraIDアプリケーション
1. **Authentication** → **Add a platform**:
   - **Web**（Confidential Client、`client_secret`必須）→ サーバーサイドWebアプリ向け
   - **Mobile and desktop applications**（Public Client、PKCE）→ CLI、デスクトップ向け
   - リダイレクトURIを登録: `http://localhost:8765/callback`（ローカルテスト用）
2. **Allow public client flows**（Mobile and desktop の場合のみ）: **Yes** に設定
3. **Certificates & secrets**: Client secret を作成（Web プラットフォームの場合）
4. **API permissions** → Add a permission → **My APIs** → API側のEntraIDアプリケーションを選択
5. **Delegated permissions** タブを選択（**Application permissions ではない**）
6. 上で定義したスコープ（例 `Gateway.Access`）にチェック
7. **Grant admin consent** を実行

> [!CAUTION]
> 「Add a permission」→「My APIs」でAPI側のEntraIDアプリケーションが**表示されない**場合は、
> **API側のEntraIDアプリケーションの「Expose an API」でスコープを1つ以上定義**する必要がある。
> Application ID URI だけでは「My APIs」に表示されない。

> [!CAUTION]
> Add permission の画面で **Delegated permissions タブがグレーアウト**している場合は、
> API側で「Expose an API」のScopeが定義されていない。  
> API側で `Add a scope` を実行してから、ブラウザをハードリロード（Ctrl+Shift+R）して再度試す。

> [!CAUTION]
> プラットフォーム種別（Web vs Mobile and desktop）を間違えるとエラーになる：
> - **「Web」プラットフォーム**: `client_secret` 必須 → MSALで `ConfidentialClientApplication` を使う
> - **「Mobile and desktop applications」**: PKCE使用、`client_secret` 不要 → MSALで `PublicClientApplication` を使う
>
> プラットフォーム種別を間違えると以下のエラーが出る:
> - Web プラットフォームに登録 + `PublicClientApplication` で実行 →
>   `AADSTS7000218: The request body must contain the following parameter: 'client_assertion' or 'client_secret'`

### 基本的な流れ
```
1. クライアント側のEntraIDアプリケーションがブラウザ経由でEntra IDの認可エンドポイントにリダイレクト
   ↓
2. ユーザーがEntra IDでログイン＆同意
   ↓
3. Entra IDが認可コード（code）を redirect_uri にリダイレクトで返す
   ↓
4. クライアント側のEntraIDアプリケーションが認可コード + (client_secret or PKCE verifier) でトークンエンドポイントを叩く
   ↓
5. アクセストークン + リフレッシュトークン取得
   ↓
6. Authorization: Bearer <token> で Gateway に送信
   ↓
7. Gateway が JWT を検証
   ↓
8. 問題なければリクエストをAzureなどのプロバイダーに転送
```

### 必要なライブラリ
```bash
uv pip install msal httpx
```

### Curl（参考：手動でフローを実行する例）

> [!NOTE]
> 認可コードフローはブラウザを介する必要があるため、curl だけで完結させるのは難しい。
> 通常はライブラリ（msal等）を使う。以下は学習用に手動でフローを追う場合の例。

#### Step 1: ブラウザで認可URLを開いて認可コード取得
以下のURLをブラウザで開いてログイン:
```
https://login.microsoftonline.com/<TENANT_ID>/oauth2/v2.0/authorize?
  client_id=<クライアント側のEntraIDアプリケーションのCLIENT_ID>&
  response_type=code&
  redirect_uri=http://localhost:8765/callback&
  scope=api://<APP_ID_URI>/Gateway.Access offline_access openid profile&
  state=random_string
```

リダイレクト先のURLから `code=xxx` の部分を取り出す。

#### Step 2: 認可コードをアクセストークンに交換
```bash
TOKEN=$(curl -s -X POST "https://login.microsoftonline.com/<TENANT_ID>/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=<クライアント側のEntraIDアプリケーションのCLIENT_ID>" \
  -d "client_secret=<クライアント側のEntraIDアプリケーションのCLIENT_SECRET>" \
  -d "code=<取得した認可コード>" \
  -d "redirect_uri=http://localhost:8765/callback" \
  -d "grant_type=authorization_code" \
  | jq -r '.access_token')
```

#### Step 3: Gateway にリクエスト送信
```bash
curl -v -H "Authorization: Bearer $TOKEN" \
  -H "x-ai-eg-model: gpt-5-4" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5-4",
    "messages": [{"role": "user", "content": "こんにちは"}]
  }' \
  http://localhost:8080/v1/chat/completions
```

### 踏み台サーバ経由で実行する場合（SSM Session Manager）

ブラウザを介する必要があるため、踏み台で実行する場合はポート転送が必要。

#### Mac側で SSM ポートフォワード起動

```bash
# ターミナル1: コールバック用 (8765)
aws ssm start-session \
  --target <INSTANCE_ID> \
  --document-name AWS-StartPortForwardingSession \
  --parameters '{"portNumber":["8765"],"localPortNumber":["8765"]}' \
  --region ap-northeast-1

# ターミナル2: AI Gateway用 (8080)
aws ssm start-session \
  --target <INSTANCE_ID> \
  --document-name AWS-StartPortForwardingSession \
  --parameters '{"portNumber":["8080"],"localPortNumber":["8080"]}' \
  --region ap-northeast-1
```

#### 踏み台側で AI Gateway への port-forward を起動
```bash
# ターミナル3: 踏み台にログイン
aws ssm start-session --target <INSTANCE_ID> --region ap-northeast-1

# 踏み台で
ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system \
  --selector=gateway.envoyproxy.io/owning-gateway-namespace=envoy-gateway-system,gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n envoy-gateway-system svc/$ENVOY_SERVICE 8080:80 &
```

#### スクリプト実行 → 表示されたURLをMacのブラウザで開いてログイン

これで:
- Mac の `localhost:8765` → SSM経由 → 踏み台の `localhost:8765`（コールバック受信）
- Mac の `localhost:8080` → SSM経由 → 踏み台の `localhost:8080`（AI Gateway接続）

### LangChain
```python
import os
import time
import threading
import http.server
import socketserver
import urllib.parse
import msal
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser


class TokenProvider:
    """メモリ内でトークンキャッシュを管理（プロセス再起動で消える）"""
    
    def __init__(self):
        # AZURE_TENANT_ID: EntraIDテナントID
        # AZURE_CLIENT_ID: クライアント側のEntraIDアプリケーションのApplication (client) ID
        # AZURE_CLIENT_SECRET: クライアント側のEntraIDアプリケーションのClient secret
        self.tenant_id = os.environ["AZURE_TENANT_ID"]
        self.client_id = os.environ["AZURE_CLIENT_ID"]
        self.client_secret = os.environ["AZURE_CLIENT_SECRET"]
        # scopes: API側のEntraIDアプリケーションのApplication ID URI + /<個別スコープ名>
        self.scopes = ["api://<APP_ID_URI>/Gateway.Access"]
        self.redirect_uri = "http://localhost:8765/callback"
        
        # token_cache を指定しない → デフォルトでメモリ内キャッシュ
        self.app = msal.ConfidentialClientApplication(
            client_id=self.client_id,
            client_credential=self.client_secret,
            authority=f"https://login.microsoftonline.com/{self.tenant_id}",
        )
        self._initial_login_done = False
    
    def get_token(self) -> str:
        # 既にログイン済みならサイレント取得を試みる（リフレッシュトークン使用）
        if self._initial_login_done:
            accounts = self.app.get_accounts()
            if accounts:
                result = self.app.acquire_token_silent(self.scopes, account=accounts[0])
                if result and "access_token" in result:
                    return result["access_token"]
        
        return self._login_via_browser()
    
    def _login_via_browser(self) -> str:
        auth_response = {}
        
        class CallbackHandler(http.server.BaseHTTPRequestHandler):
            def do_GET(self):
                query = urllib.parse.urlparse(self.path).query
                params = urllib.parse.parse_qs(query)
                for key, value in params.items():
                    auth_response[key] = value[0]
                
                if "code" in auth_response:
                    self.send_response(200)
                    self.send_header("Content-Type", "text/html; charset=utf-8")
                    self.end_headers()
                    self.wfile.write("認証成功！このタブを閉じてください。".encode("utf-8"))
                else:
                    self.send_response(400)
                    self.send_header("Content-Type", "text/html; charset=utf-8")
                    self.end_headers()
                    error_msg = f"認証失敗: {auth_response.get('error_description', 'unknown error')}"
                    self.wfile.write(error_msg.encode("utf-8"))
            
            def log_message(self, format, *args):
                pass
        
        server = socketserver.TCPServer(("localhost", 8765), CallbackHandler)
        server_thread = threading.Thread(target=server.serve_forever, daemon=True)
        server_thread.start()
        
        flow = self.app.initiate_auth_code_flow(
            scopes=self.scopes,
            redirect_uri=self.redirect_uri,
        )
        
        print("=" * 80)
        print("以下のURLをMacのブラウザで開いてください:")
        print()
        print(flow["auth_uri"])
        print()
        print("=" * 80)
        print("\nブラウザでログインを待機中...")
        
        while "code" not in auth_response and "error" not in auth_response:
            time.sleep(0.1)
        server.shutdown()
        
        if "error" in auth_response:
            raise RuntimeError(f"認証エラー: {auth_response.get('error_description')}")
        
        result = self.app.acquire_token_by_auth_code_flow(
            auth_code_flow=flow,
            auth_response=auth_response,
        )
        
        if "access_token" not in result:
            raise RuntimeError(f"トークン取得エラー: {result.get('error_description')}")
        
        self._initial_login_done = True
        print("✓ ログイン成功(以後はメモリキャッシュから取得)")
        return result["access_token"]


# === LangChain (インタラクティブループ) ===
token_provider = TokenProvider()
initial_token = token_provider.get_token()
print(f"アクセストークン取得成功（最初の50文字）: {initial_token[:50]}...\n")

prompt = ChatPromptTemplate.from_messages([
    ("system", "あなたは親切なアシスタントです。日本語で回答してください。"),
    ("user", "{question}"),
])

while True:
    user_input = input("\n質問を入力してください (quit で終了): ")
    if user_input.strip().lower() in ("quit", "exit", ""):
        print("終了します")
        break
    
    # 都度トークン取得(期限切れなら自動リフレッシュ)
    token = token_provider.get_token()
    
    llm = ChatOpenAI(
        #model="claude-haiku-4-5",
        model="gemini-3.1-flash-lite-preview",
        #model="gpt-5-4",
        base_url="http://localhost:8080/v1",
        api_key="dummy",
        default_headers={"Authorization": f"Bearer {token}"},
        streaming=True,
        timeout=120.0,
    )
    
    chain = prompt | llm | StrOutputParser()
    
    for chunk in chain.stream({"question": user_input}):
        print(chunk, end="", flush=True)
    print()
```

### Strands Agents
```python
import os
import time
import asyncio
import threading
import http.server
import socketserver
import urllib.parse
import msal
from strands import Agent
from strands.models.openai import OpenAIModel


class TokenProvider:
    """メモリ内でトークンキャッシュを管理（プロセス再起動で消える）"""
    
    def __init__(self):
        # AZURE_TENANT_ID: EntraIDテナントID
        # AZURE_CLIENT_ID: クライアント側のEntraIDアプリケーションのApplication (client) ID
        # AZURE_CLIENT_SECRET: クライアント側のEntraIDアプリケーションのClient secret
        self.tenant_id = os.environ["AZURE_TENANT_ID"]
        self.client_id = os.environ["AZURE_CLIENT_ID"]
        self.client_secret = os.environ["AZURE_CLIENT_SECRET"]
        # scopes: API側のEntraIDアプリケーションのApplication ID URI + /<個別スコープ名>
        self.scopes = ["api://<APP_ID_URI>/Gateway.Access"]
        self.redirect_uri = "http://localhost:8765/callback"
        
        self.app = msal.ConfidentialClientApplication(
            client_id=self.client_id,
            client_credential=self.client_secret,
            authority=f"https://login.microsoftonline.com/{self.tenant_id}",
        )
        self._initial_login_done = False
    
    def get_token(self) -> str:
        if self._initial_login_done:
            accounts = self.app.get_accounts()
            if accounts:
                result = self.app.acquire_token_silent(self.scopes, account=accounts[0])
                if result and "access_token" in result:
                    return result["access_token"]
        
        return self._login_via_browser()
    
    def _login_via_browser(self) -> str:
        auth_response = {}
        
        class CallbackHandler(http.server.BaseHTTPRequestHandler):
            def do_GET(self):
                query = urllib.parse.urlparse(self.path).query
                params = urllib.parse.parse_qs(query)
                for key, value in params.items():
                    auth_response[key] = value[0]
                
                if "code" in auth_response:
                    self.send_response(200)
                    self.send_header("Content-Type", "text/html; charset=utf-8")
                    self.end_headers()
                    self.wfile.write("認証成功！このタブを閉じてください。".encode("utf-8"))
                else:
                    self.send_response(400)
                    self.send_header("Content-Type", "text/html; charset=utf-8")
                    self.end_headers()
                    error_msg = f"認証失敗: {auth_response.get('error_description', 'unknown error')}"
                    self.wfile.write(error_msg.encode("utf-8"))
            
            def log_message(self, format, *args):
                pass
        
        server = socketserver.TCPServer(("localhost", 8765), CallbackHandler)
        server_thread = threading.Thread(target=server.serve_forever, daemon=True)
        server_thread.start()
        
        flow = self.app.initiate_auth_code_flow(
            scopes=self.scopes,
            redirect_uri=self.redirect_uri,
        )
        
        print("=" * 80)
        print("以下のURLをMacのブラウザで開いてください:")
        print()
        print(flow["auth_uri"])
        print()
        print("=" * 80)
        print("\nブラウザでログインを待機中...")
        
        while "code" not in auth_response and "error" not in auth_response:
            time.sleep(0.1)
        server.shutdown()
        
        if "error" in auth_response:
            raise RuntimeError(f"認証エラー: {auth_response.get('error_description')}")
        
        result = self.app.acquire_token_by_auth_code_flow(
            auth_code_flow=flow,
            auth_response=auth_response,
        )
        
        if "access_token" not in result:
            raise RuntimeError(f"トークン取得エラー: {result.get('error_description')}")
        
        self._initial_login_done = True
        print("✓ ログイン成功(以後はメモリキャッシュから取得)")
        return result["access_token"]


def silent_callback(**kwargs):
    """何もしない（デフォルトの stdout 出力を抑制）"""
    pass


# === Strands Agent (インタラクティブループ) ===
async def main():
    token_provider = TokenProvider()
    initial_token = token_provider.get_token()
    print(f"アクセストークン取得成功（最初の50文字）: {initial_token[:50]}...\n")
    
    while True:
        user_input = input("\n質問を入力してください (quit で終了): ")
        if user_input.strip().lower() in ("quit", "exit", ""):
            print("終了します")
            break
        
        token = token_provider.get_token()
        
        model = OpenAIModel(
            client_args={
                "base_url": "http://localhost:8080/v1",
                "api_key": "dummy",
                "default_headers": {"Authorization": f"Bearer {token}"},
                "timeout": 120.0,
            },
            model_id="claude-haiku-4-5",
            #model_id="gemini-3.1-pro-preview",
            #model_id="gpt-5-4",
        )
        
        agent = Agent(
            model=model,
            system_prompt="あなたは親切なアシスタントです。日本語で回答してください。",
            callback_handler=silent_callback,
        )
        
        async for event in agent.stream_async(user_input):
            if "data" in event:
                print(event["data"], end="", flush=True)
        print()


asyncio.run(main())
```

### トークンキャッシュ：メモリ vs ファイル

上記のコードはメモリキャッシュを使っているため、**プロセス再起動のたびに再ログインが必要**。
スクリプトを毎回起動するCLI用途なら、ファイルキャッシュにすると最大90日間ブラウザログイン不要になる。

#### ファイルキャッシュ版（永続化したい場合）
```python
from pathlib import Path
import os
import stat

class TokenProvider:
    def __init__(self):
        # ... 既存の設定 ...
        
        cache_file = Path.home() / ".envoy_ai_gateway_token_cache.bin"
        cache = msal.SerializableTokenCache()
        if cache_file.exists():
            cache.deserialize(cache_file.read_text())
        
        self.cache = cache
        self.cache_file = cache_file
        
        self.app = msal.ConfidentialClientApplication(
            client_id=self.client_id,
            client_credential=self.client_secret,
            authority=f"https://login.microsoftonline.com/{self.tenant_id}",
            token_cache=cache,  # ← ファイルキャッシュを渡す
        )
    
    def _save_cache(self):
        if self.cache.has_state_changed:
            self.cache_file.write_text(self.cache.serialize())
            os.chmod(self.cache_file, stat.S_IRUSR | stat.S_IWUSR)  # 600
    
    # _login_via_browser の最後で self._save_cache() を呼ぶ
```

> [!CAUTION]
> ファイルキャッシュにはアクセストークンとリフレッシュトークンが含まれる。
> - パーミッションを600にする
> - Git管理下に置かない（`.gitignore`に追加）
> - 共有ディレクトリに置かない

---

## JWTクレームの由来と認可制御での活用

Envoy AI Gateway の `SecurityPolicy.authorization` でロールベース認可を実装する際、
JWTのどのクレームが「何の設定」「誰の属性」から来るのかを理解しておくと、
認可ルールの設計や設定ミスのデバッグがスムーズになる。

### クレームの由来は4種類に分類できる

JWTのクレームは厳密に分けると以下のいずれかから来る:

| 由来 | 意味 |
|---|---|
| **API側のEntraIDアプリケーション** | API側のApp registrationの設定(App rolesのValue、Application ID URIなど)から決まる |
| **クライアント側のEntraIDアプリケーション** | クライアント側のApp registrationから決まる(主に「誰が呼んでいるか」の識別情報) |
| **テナント** | Entra IDテナント自体の属性。App registrationとは無関係に、テナントが決まれば自動的に決まる |
| **ユーザー** | ログインしたユーザーアカウント自体の属性。App registrationとは無関係に、ユーザーアカウントが決まれば自動的に決まる(認可コードフローのみ) |

#### 「テナント」由来とは

Entra IDは組織ごとに「テナント」という単位で分かれており、テナントには一意のGUID(Tenant ID)が振られている。
App registrationを作る前から、**テナントそのものの属性**として決まっているものが「テナント由来」のクレーム。

| クレーム | 値の例 | なぜテナント由来か |
|---|---|---|
| `iss` | `https://sts.windows.net/<TENANT_ID>/` または `https://login.microsoftonline.com/<TENANT_ID>/v2.0` | トークンの発行者URL。テナントが決まれば自動的に決まる |
| `tid` | `<TENANT_ID>` (GUID) | トークンを発行したテナントのID |

これらは、同じテナント内ならAPI側アプリを変えても、クライアント側アプリを変えても、ユーザーが違っても**同じ値**になる。

> [!NOTE]
> 厳密には `iss` の v1/v2 形式は **API側アプリ**の `requestedAccessTokenVersion` 設定の影響も受ける。
> ホスト部分(`sts.windows.net` vs `login.microsoftonline.com`)はトークンバージョンで決まり、
> テナントID部分はテナントで決まる、という分担。

#### 「ユーザー」由来とは

**認可コードフロー(ユーザーがブラウザでログインするフロー)でのみ**入るクレーム。
Entra IDのディレクトリに登録されているユーザーアカウント自体の属性から来る。

| クレーム | 値の例 | なぜユーザー由来か |
|---|---|---|
| `oid` (認可コードフロー時) | `<USER_OBJECT_ID>` (GUID) | ログインしたユーザーアカウントの一意なObject ID |
| `sub` (認可コードフロー時) | (アプリごとにハッシュされた値、またはoidと同じ) | ログインユーザーのsubject識別子 |
| `name` | "Lee Joon-ki" | Entra IDにそのユーザーアカウントが登録されているときの表示名 |
| `preferred_username` | "leejoonki@kinto-technologies.com" | ユーザーアカウントのメアド / UPN |
| `email` | "leejoonki@kinto-technologies.com" | ユーザーアカウントのメール属性(Optional Claim) |
| `upn` | UPN形式 | ユーザーのプリンシパル名 |
| `roles` (認可コードフロー時) | `["Gateway.User"]` | API側で**定義**されたApp roleのうち、**そのユーザーに割り当てられている**もの |

クライアント側アプリの設定にも、API側アプリの設定にも書かれていない、**ユーザー自身のプロパティ**が入る。

> [!NOTE]
> クライアントクレデンシャルフローは M2M 認証で、ブラウザもユーザーもログイン操作も介在しないため、
> `name` / `preferred_username` / `email` / `upn` のような**ユーザー属性は入らない**。
> `oid` / `sub` は入るが、これは**クライアント側のサービスプリンシパルのObject ID**であって、ユーザー由来ではない。

### クレーム別: 由来とフロー別の値

両フローのクレームを「何由来か」で整理した完全版:

| クレーム | 由来 | クライアントクレデンシャル | 認可コード |
|---|---|---|---|
| `iss` | テナント(+ API側のトークンバージョン設定) | テナント発行者URL | テナント発行者URL(同じ) |
| `tid` | テナント | テナントID | テナントID |
| `aud` | API側 | API側のApplication ID URI | API側のApplication ID URI |
| `appid` / `azp` | クライアント側 | クライアント側CLIENT_ID | クライアント側CLIENT_ID |
| `oid` (CC時) | クライアント側のサービスプリンシパル | クライアント側SPのObject ID | — |
| `sub` (CC時) | クライアント側のサービスプリンシパル | クライアント側SPのObject ID | — |
| `oid` (AuthCode時) | ユーザー | — | ログインユーザーのObject ID |
| `sub` (AuthCode時) | ユーザー(アプリ固有にハッシュ) | — | ユーザーのsubject識別子 |
| `name` | ユーザー | — | ユーザーの表示名 |
| `preferred_username` | ユーザー | — | ユーザーのUPN/メール |
| `scp` | API側で定義、ユーザーへの委任 | — | 委任されたスコープ名 |
| `roles` | API側で定義、誰に割り当てたかで内容が決まる | クライアント側SPに割り当てられたロール | ログインユーザーに割り当てられたロール |
| `idtyp` | Entra IDが自動付与(オプショナルクレーム) | `"app"` | `"user"` または存在しない |

### App roleとScopeの「定義場所」と「割り当て先」を分けて考える

認可制御で `roles` / `scp` クレームを使うとき、最も混乱しやすいのが
**「ロール/スコープを定義する場所」と「誰に割り当てるか」が別**という点。

| 観点 | クライアントクレデンシャル | 認可コード |
|---|---|---|
| **App role / Scope の定義場所** | API側 | API側 |
| **割り当て先(誰に付与するか)** | クライアント側のサービスプリンシパル | ユーザー(または所属グループ) |
| **割り当て方** | クライアント側で「Application permissions」として追加 + Grant admin consent | ユーザーをAPI側のEnterprise ApplicationでApp roleに直接割り当て、もしくはDelegated permissionsで委任 |
| **トークンに入るクレーム** | `roles` | `roles`(ユーザーへのApp role割り当て時) / `scp`(Delegated permission時) |

つまり「**保護対象API(=Gateway)が『こういうロールがあります』と宣言し、それを誰に与えるかを後から決める**」モデル。
直感的にも、`Gateway.Access` というロール名を定義する権限を持つのは Gateway 自身(=API側)と考えると自然。

> [!CAUTION]
> クライアント側のEntraIDアプリケーションにも「App roles」タブはあるが、
> **そこにロールを定義しても、Gatewayへのアクセス制御には使われない**。
> クライアント側のApp rolesは「このクライアント自身が**保護対象になる場合**(=別のクライアントから呼ばれる側になる場合)」のためのもの。
> 今回Gatewayを呼ぶバッチアプリは「呼ぶ側」なので、自分のApp rolesは使われない。

> [!CAUTION]
> **グループ経由のApp role割り当ては、クライアントクレデンシャルフローでは機能しない。**  
> クライアント側のサービスプリンシパルをセキュリティグループに入れて、そのグループにApp roleを割り当てても、
> `roles` クレームには入らない。**サービスプリンシパルに直接割り当てる必要がある**
> (「Grant admin consent」がこれを実行している)。
> 公式ドキュメント: "Currently, if you add a service principal to a group, and then assign an app role to that group, Microsoft Entra ID doesn't add the roles claim to tokens it issues"

> [!TIP]
> API側のEnterprise Applicationで「**Assignment required: Yes**」にしておくと、
> 明示的にApp roleを割り当てられていないクライアント/ユーザーは
> そもそもトークン取得自体ができなくなる(`AADSTS501051`)。
> 本番運用ではYesにしておくのが推奨。

### Envoy AI Gateway での認可ルール設計の指針

`SecurityPolicy.authorization` でクレームを使って制御する場合、目的別に「どのクレームを見るか」が変わる。

| 目的 | 見るべきクレーム | 備考 |
|---|---|---|
| どのテナントから来たか | `iss` または `tid` | マルチテナント運用時、または検証目的 |
| どのAPI向けのトークンか | `aud` | `SecurityPolicy.jwt.audiences` での照合で済むが、追加で見ることも可能 |
| どのクライアントアプリから来たか | `appid` / `azp` | M2M用途で「バッチA だけ許可、バッチB は不可」のような制御 |
| ログインユーザーが誰か | `oid`(永続ID、推奨) / `preferred_username`(メアドベース、人間に読みやすい) | 認可コードフロー時のみ |
| M2Mロールベース認可 | `roles` | API側で定義 + クライアント側SPに割り当て |
| ユーザーロールベース認可 | `roles` または `scp` | `roles`: API側で定義 + ユーザーに割り当て / `scp`: API側で定義 + ユーザーに委任 |
| フロー識別(M2M vs ユーザー) | `idtyp`(オプショナルクレーム) | v2.0トークンの場合、Token Configurationで明示的にOptional claimとして追加が必要なケースあり |

#### クライアント識別 + ロールベース認可の併用例

「クライアントAは `Gateway.Access` ロールを持っているが、Gatewayを呼べるのは特定のクライアントだけ」のような細かい制御:

```yaml
spec:
  authorization:
    defaultAction: Deny
    rules:
      # 特定のクライアントアプリ かつ Gateway.Access ロール保持者のみ許可
      - name: "allow-specific-batch-app"
        action: Allow
        principal:
          jwt:
            provider: entra
            claims:
              - name: appid  # クライアント側EntraIDアプリケーションのCLIENT_ID
                valueType: "String"
                values:
                  - "<batch-app-client-id>"
              - name: roles
                valueType: "StringArray"
                values:
                  - "Gateway.Access"
```

#### ハマりやすいポイント: 同名のScopeとApp role

同じValue名(例 `Gateway.Access`)をScopeとApp roleの両方で定義すると、
フローによって入るクレームが違う(クライアントクレデンシャル → `roles`、認可コード → `scp`)ため、
SecurityPolicy.authorization で両方の場合を書き分ける必要がある。

運用上のおすすめは:
- **名前を分ける**(例: App role → `Gateway.Access.App`、Scope → `Gateway.Access.User`)
- もしくは**同名のままにする場合は、`scp` と `roles` の両方を見る認可ルールを書く**

---

## クライアントクレデンシャルフロー vs 認可コードフロー まとめ

| 項目 | クライアントクレデンシャル | 認可コード |
|---|---|---|
| **認証対象** | クライアント側のEntraIDアプリケーション（M2M） | エンドユーザー |
| **ブラウザ** | 不要 | 必要 |
| **ユースケース** | バッチ、サービス間通信 | Webアプリ、社内ツール |
| **API側のEntraIDアプリケーションのスコープ定義** | 不要（`.default`で動く） | **必要**（個別スコープ） |
| **API側のEntraIDアプリケーションのApp role定義** | 任意（`roles`認可入れる場合のみ必要） | 通常不要 |
| **クライアント側のEntraIDアプリケーションの権限種別** | Application permissions（任意） | **Delegated permissions（必須）** |
| **`client_secret`** | 必要 | Web プラットフォームなら必要、Mobile/Desktop（PKCE）なら不要 |
| **scope の指定** | `api://.../.default` | `api://.../<個別スコープ>` |
| **トークンの `aud`** | API側のEntraIDアプリケーション | API側のEntraIDアプリケーション（同じ） |
| **トークンの `iss`** | 同じ | 同じ |
| **トークンの `sub` / `oid`** | クライアント側のEntraIDアプリケーションのObject ID | **ユーザーのObject ID** |
| **トークンの `appid`** | クライアント側のEntraIDアプリケーションのApplication (client) ID | クライアント側のEntraIDアプリケーションのApplication (client) ID |
| **トークンの `scp`** | なし | あり |
| **トークンの `roles`** | 設定すれば入る（App role + Application permission の付与が必要） | なし（通常） |
| **トークンの `name` / `preferred_username`** | なし | あり |
| **トークンの `idtyp`** | `"app"` | `"user"`（または存在しない） |

### Gateway側の SecurityPolicy はどちらも同じ
両フローとも、Gateway 側の `SecurityPolicy.jwt` の設定は変更不要。  
Gateway は受け取った JWT を検証するだけで、どのフローで取得されたかは識別しない。

### 共存可能（用途別にクライアント側のEntraIDアプリケーションを分ける）

両フローは**1つのAPI側のEntraIDアプリケーション**+ 別々の**クライアント側のEntraIDアプリケーション**で共存可能。

```
API側のEntraIDアプリケーション (envoy-ai-gateway): 1つだけ用意
  ├─ Expose an API → Scope: Gateway.Access (認可コード用)
  └─ App roles → App role: Gateway.Access (Applications) (クライアントクレデンシャルでrolesクレーム使う場合)

クライアント側のEntraIDアプリケーションA (バッチ・サービス間通信用):
  └─ API permissions → Application permissions: Gateway.Access

クライアント側のEntraIDアプリケーションB (Webアプリ・ユーザー認証用):
  └─ API permissions → Delegated permissions: Gateway.Access
```

> [!CAUTION]
> **同じクライアント側のEntraIDアプリケーションに、Delegated permissionsとApplication permissionsを後から両方追加しようとしても、Azureポータル上では片方しか選択できない。**  
> 「My APIs」で同じAPIを選んだあとの画面で、すでに片方のpermissionが追加されていると、もう片方のタブがグレーアウトされたり、選択しても効かなかったりする挙動がある。
>
> **対処：**
> - 「Add a permission」ダイアログを一度閉じて、改めて開き直す
> - ブラウザをハードリロード（Ctrl+Shift+R）
> - それでもダメなら、**用途ごとにクライアント側のEntraIDアプリケーションを分ける**のが最も確実
>
> 単一のクライアント側EntraIDアプリケーションで両方を持たせるのは技術的には可能だが、以下の理由で**用途別に分ける運用を強く推奨**する：
> - **権限境界が明確**: バッチ用とWebアプリ用が混在しない
> - **片方だけ無効化が可能**: バッチを止めたいときに Webアプリに影響しない
> - **監査ログでフロー識別が容易**: `appid` クレームで「どのアプリ経由で来たか」が即座に分かる
> - **Client Secret のローテーションが独立**: 片方の漏洩がもう片方に波及しない
> - **Azureポータルの設定UIの制約に引っかからない**: 上記のような表示問題が発生しない

### 認可ルールでフロー別に制御可能
`SecurityPolicy.authorization` を使って、`scp`（認可コード）と `roles`（クライアントクレデンシャル）で
別々の認可ルールを書ける（クライアントクレデンシャル側はAPI側のApp role + クライアント側のApplication permission の設定が前提）:

```yaml
spec:
  jwt:
    providers:
      - name: entra
        # ... 既存設定
  authorization:
    defaultAction: Deny
    rules:
      # クライアントクレデンシャル経由（M2M）
      - name: "allow-batch-services"
        action: Allow
        principal:
          jwt:
            provider: entra
            claims:
              - name: roles
                valueType: "StringArray"
                values:
                  - "Gateway.Access"
      # 認可コード経由（ユーザー）
      - name: "allow-user-access"
        action: Allow
        principal:
          jwt:
            provider: entra
            claims:
              - name: scp
                valueType: "String"
                values:
                  - "Gateway.Access"
```

---

## トラブルシュート

### Envoyログでよくある response_code_details

| メッセージ | 原因 | 対処 |
|---|---|---|
| `Jwt_issuer_is_not_configured` | `issuer` がトークンの `iss` と不一致 | jwt.ioで `iss` を確認して一致させる |
| `Audiences_in_Jwt_are_not_allowed` | `audiences` がトークンの `aud` と不一致 | jwt.ioで `aud` を確認して `audiences` に追加 |
| `Jwt_verification_fails` | 署名検証失敗 | `remoteJWKS.uri` の到達性を確認 |
| `Jwks_remote_fetch_is_failed` | JWKS取得失敗 | NetworkPolicy/egress proxy 設定を確認 |
| `Jwt_is_expired` | トークン期限切れ | 再取得 |

### EntraID 設定でよくあるエラー

| エラー | 原因 | 対処 |
|---|---|---|
| 「My APIs」にAPI側のEntraIDアプリケーションが表示されない | API側で「Expose an API」のApplication ID URIが未設定、またはScope/App roleが未定義 | API側でApplication ID URIを設定し、必要に応じてScope/App roleを定義 |
| Delegated permissions タブがグレーアウト | API側で「Expose an API」のScopeが未定義 | API側で `Add a scope` を実行 |
| Application permissions タブがグレーアウト | API側で「App roles」のApplicationsタイプのロールが未定義 | API側で `Create app role`（Allowed member types: Applications）を実行 |
| 設定したのにタブがグレーアウト | ブラウザキャッシュ、または同じクライアント側EntraIDアプリケーションに既に片方のpermissionが追加されている | Ctrl+Shift+R でハードリロード、ダイアログを閉じて開き直す、または**用途別にクライアント側EntraIDアプリケーションを分ける** |
| `AADSTS65001: admin consent required` | Grant admin consent が未実行 | クライアント側のAPI permissions画面で `Grant admin consent` をクリック |
| クライアントクレデンシャルフローでJWT検証は通るが `roles` クレームが入らない | App role 未定義、Application permissions 未追加、Grant admin consent 未実行のいずれか | API側で App role 定義 → クライアント側で Application permissions に追加 → Grant admin consent |
| 単一アプリ構成でトークンの `aud` と `appid` が同じ | パターン1（単一アプリ兼任）の正常な動作 | 問題なし。Gateway側の `audiences` には自分自身のApp ID URIを設定すればOK |

### 認可コードフローでよくあるエラー

| エラー | 原因 | 対処 |
|---|---|---|
| `state mismatch` | コールバックパラメータ全体（code+state）を `acquire_token_by_auth_code_flow` に渡していない | `auth_response` 全体を渡す |
| `client_assertion or client_secret required` | プラットフォーム種別が「Web」になっているのに PublicClientApplication を使っている | `ConfidentialClientApplication` + `client_secret` を使う、または プラットフォームを「Mobile and desktop applications」に変更 |
| `Public Client should not possess credentials` | `PublicClientApplication` に `client_credential` を渡している | `ConfidentialClientApplication` を使うか、`client_credential` 引数を削除 |