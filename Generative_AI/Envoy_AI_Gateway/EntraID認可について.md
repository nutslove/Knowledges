# EntraID の認可（Authorization）モデル解説

EntraID（Azure AD）でアプリケーション間の認可を設定するときに混乱しがちな
「App roles」「Delegated permissions」「Application permissions」「Expose an API のスコープ」などの
概念を整理する。

## 目次
- [前提：認証 vs 認可](#前提認証-vs-認可)
- [前提：APIを呼ぶ2つのパターン](#前提apiを呼ぶ2つのパターン)
- [認可モデルの全体像](#認可モデルの全体像)
- [4つの設定項目の対応関係](#4つの設定項目の対応関係)
- [各設定項目の詳細](#各設定項目の詳細)
  - [1. 「Expose an API」のスコープ（Scope）](#1-expose-an-api-のスコープscope)
  - [2. 「App roles」のアプリロール](#2-app-roles-のアプリロール)
  - [3. 「Delegated permissions」](#3-delegated-permissions-クライアント側)
  - [4. 「Application permissions」](#4-application-permissions-クライアント側)
- [Grant admin consent とは](#grant-admin-consent-とは)
- [設定画面でグレーアウトする理由](#設定画面でグレーアウトする理由)
- [実際の設定の流れ](#実際の設定の流れ)
- [用途別ベストプラクティス](#用途別ベストプラクティス)
- [トークンによる動作確認](#トークンによる動作確認)
- [まとめ](#まとめ)
- [トラブルシュート早見表](#トラブルシュート早見表)

---

## 前提：認証 vs 認可

EntraIDで扱うのはこの2つ。混同しがちなので最初に整理しておく。

| 用語 | 意味 | 例 |
|---|---|---|
| **認証（Authentication）** | 「あなたは誰？」を確認する | パスワード、Client secret、PKCE |
| **認可（Authorization）** | 「あなたは何ができる？」を決める | スコープ、ロール、権限 |

このドキュメントで扱う「App roles」「Delegated permissions」「Application permissions」などは
**すべて認可の話**。

---

## 前提：APIを呼ぶ2つのパターン

EntraIDでは、APIを呼び出すパターンが大きく2種類ある。  
これが、後述するDelegated と Application の使い分けに直結する。

| パターン | 主体（=誰として動くか） | 代表的なフロー | 用途例 |
|---|---|---|---|
| **A. ユーザー認証パターン** | エンドユーザー（人間） | 認可コードフロー（OAuth 2.0 Authorization Code） | Webアプリ、CLI、モバイルアプリ |
| **B. M2M（マシン認証）パターン** | アプリ自身（人間は介在しない） | クライアントクレデンシャルフロー | バッチ、サービス間通信、CIパイプライン |

### パターンA: ユーザー認証
ブラウザでEntraIDのログイン画面が出て、ユーザーがログイン → アプリは「ログインしたユーザーの代わりに」APIを呼ぶ。  
トークンには「誰がログインしたか」（ユーザーのID、メアドなど）が含まれる。

### パターンB: M2M
ユーザーはいない。アプリが自分のClient ID + Client secretで直接認証 → アプリ自身がAPIを呼ぶ。  
トークンには「どのアプリが呼んでいるか」が含まれるが、ユーザー情報はない。

---

## 認可モデルの全体像

EntraIDの認可モデルは「**API側でメニューを定義して、クライアント側でメニューから注文する**」という構造。

```
┌─────────────────────────────────────────────────┐
│ 【API側のEntraIDアプリケーション】                │
│ どんな権限があるかを定義する（メニュー作成）       │
│                                                  │
│  「Expose an API」 → Scope を定義               │
│  「App roles」    → App role を定義             │
└─────────────────────────────────────────────────┘
                    ↓ 「この権限を使いたい」と要求
┌─────────────────────────────────────────────────┐
│ 【クライアント側のEntraIDアプリケーション】       │
│ どの権限を使うか宣言する（メニューから注文）      │
│                                                  │
│  「API permissions」                             │
│   ├─ Delegated permissions（ScopeをもとにAPI側のメニューから選択）│
│   └─ Application permissions（App roleをもとにAPI側のメニューから選択）│
└─────────────────────────────────────────────────┘
```

> [!NOTE]
> 本ドキュメントでは、用途別に2種類のEntraIDアプリケーションを区別する：
> - **API側のEntraIDアプリケーション**: APIを提供する側
> - **クライアント側のEntraIDアプリケーション**: APIを呼び出す側
>
> どちらもAzureポータルの「App registrations」に登録するEntraIDアプリケーションで、
> 違いは役割だけ。同じテナント内に複数登録して使い分けるのが一般的。

---

## 4つの設定項目の対応関係

ここがこのドキュメントで一番伝えたいポイント。  
**同じ「権限」を表すのに、用途別に2系統の仕組みが用意されている**。

| 用途 | API側で定義する場所 | クライアント側で要求する場所 | トークンに入るクレーム | 使うフロー |
|---|---|---|---|---|
| **A. ユーザー認証用** | **Expose an API** → Scope | **Delegated permissions** | `scp` | 認可コードフロー |
| **B. M2M用** | **App roles**（Applicationsタイプ） | **Application permissions** | `roles` | クライアントクレデンシャルフロー |

たとえば「データ読み取り権限」を作るなら、

- ユーザー認証用に欲しければ → API側で「Scope」を定義する
- M2M用に欲しければ → API側で「App roles」（Applications）を定義する

両方欲しければ、両方定義することになる（同じ名前で並行定義できる）。

---

## 各設定項目の詳細

### 1. 「Expose an API」のスコープ（Scope）

#### 何のためにあるか
**ユーザー認証用**の権限定義。  
「ユーザーの代わりにこのAPIにアクセスする権利」を表す。

#### 定義場所
API側のEntraIDアプリケーション → **Expose an API** タブ → **Add a scope**

#### 設定項目の意味

| 項目 | 意味 |
|---|---|
| **Scope name** | スコープ名。トークンの `scp` クレームに入る値（例: `Data.Read`、`User.Read.All`） |
| **Who can consent?** | ユーザーが自分で同意できるか、管理者承認が必要か |
| **Admin consent display name/description** | 管理者向けの同意画面に表示される説明 |
| **User consent display name/description** | ユーザー向けの同意画面に表示される説明 |
| **State** | `Enabled` / `Disabled` |

#### ユーザー同意画面のイメージ
ユーザーが認可コードフローでログインしたときに表示される画面：

```
"〇〇アプリ" は以下の権限を要求しています:
- ✓ あなたの代わりにデータを読み取る (Data.Read)
- ✓ あなたのプロフィール情報にアクセスする (User.Read)

[承諾] [キャンセル]
```

ユーザーがここで「承諾」すると、アプリは**ユーザーから委任を受けた**形でAPIを呼べる。  
英語で **Delegated**（委任された）と呼ばれるのはこのため。

#### 例
- Microsoft Graph の `User.Read` → 「ユーザーのプロフィールを読み取る」
- 自作API の `Data.Read` → 「ユーザーの代わりにデータを読み取る」
- 自作API の `Order.Write` → 「ユーザーの代わりに注文を作成する」

---

### 2. 「App roles」のアプリロール

#### 何のためにあるか
主に**M2M認証用**の権限定義。  
「アプリ自身がこのAPIにアクセスする権利」を表す。

> [!NOTE]
> App rolesには「ユーザー/グループへのロール割り当て」用途もあるが、
> 本ドキュメントではM2M用途に絞って解説する。

#### 定義場所
API側のEntraIDアプリケーション → **App roles** タブ → **Create app role**

#### 設定項目の意味

| 項目 | 意味 |
|---|---|
| **Display name** | 表示名（例: `Data Reader`） |
| **Allowed member types** | このロールを誰に割り当てられるか:<br>- **Applications**: アプリ自身に割り当て（M2M用）<br>- **Users/Groups**: ユーザーやグループに割り当て<br>- **Both**: 両方 |
| **Value** | トークンの `roles` クレームに入る値（例: `Data.Read`） |
| **Description** | 説明 |
| **Do you want to enable this app role?** | 有効化するか |

#### Allowed member types の使い分け

- **Applications** を選ぶ場合: クライアントクレデンシャルフロー（M2M）でアプリにロールを割り当てる
- **Users/Groups** を選ぶ場合: 認可コードフローで特定のユーザー/グループにのみロールを割り当てる
- **Both**: 両方の用途で使える

#### 例
- 自作API の `Data.Read` (Applications) → 「アプリがデータを読む権限」
- 自作API の `Data.Admin` (Users/Groups) → 「データ管理者ロール（特定のユーザーにのみ付与）」

---

### 3. 「Delegated permissions」 (クライアント側)

#### 何のためにあるか
クライアント側のEntraIDアプリケーションで、「**ユーザーの代わりに**このスコープでAPIを呼びます」と宣言する。

#### どこから来るか
API側で定義された「Expose an API」の Scope から選ぶ。  
**API側でScopeが定義されていないと、ここには何も表示されない**（タブが選択不可になる）。

#### 設定場所
クライアント側のEntraIDアプリケーション →  
**API permissions** → **Add a permission** → **My APIs** → API側を選択 → **Delegated permissions**

#### いつ使うか
- 認可コードフロー
- ユーザーが介在するアプリ（Webアプリ、SPA、CLI、デスクトップアプリ）

#### トークンへの反映
トークンの `scp` クレームに入る：
```json
"scp": "Data.Read User.Read"
```
（複数のスコープがスペース区切りで入る）

---

### 4. 「Application permissions」 (クライアント側)

#### 何のためにあるか
クライアント側のEntraIDアプリケーションで、「**アプリ自身が**このアプリロールでAPIを呼びます」と宣言する。  
ユーザーの代わりではなく、アプリ自身の権限として。

#### どこから来るか
API側で定義された「App roles」のうち、**Allowed member types が Applications**（または Both）のロールから選ぶ。  
**API側でApplicationsタイプのApp roleが定義されていないと、ここには何も表示されない**（タブが選択不可になる）。

#### 設定場所
クライアント側のEntraIDアプリケーション →  
**API permissions** → **Add a permission** → **My APIs** → API側を選択 → **Application permissions**

#### いつ使うか
- クライアントクレデンシャルフロー
- バッチ、サービス間通信、CIパイプラインなど（ユーザーが介在しない）

#### トークンへの反映
トークンの `roles` クレームに入る：
```json
"roles": ["Data.Read"]
```
（複数のロールが配列で入る）

---

## Grant admin consent とは

API permissions に追加しただけでは、まだ権限は有効化されていない。  
**管理者が「このアプリにこの権限を与えることを承認します」と明示的に同意する**必要がある。

### なぜ必要か
EntraIDのセキュリティモデルでは、デフォルトで「権限の要求」と「権限の付与」が別々のアクション：

1. アプリ開発者が `API permissions` に追加 = 「この権限を要求します」と宣言
2. テナント管理者が `Grant admin consent` をクリック = 「テナントとして許可します」

これを分けることで、開発者が勝手に強い権限を付与できないようにしている。  
たとえば「全ユーザーのメールを読む権限」みたいな強い権限は、必ず管理者の承認が必要になる。

### Grant admin consent が必要なケース

- **必ず必要**: Application permissions（アプリ自身の権限はユーザーには同意できない）
- **必要なことが多い**: Delegated permissions で「Admin consent required」が `Yes` のもの
- **不要**: Delegated permissions で「Admin consent required」が `No` のもの（ユーザーが個別に同意できる）

### 実行方法
クライアント側のEntraIDアプリケーション → **API permissions** → 画面上部の  
「**Grant admin consent for <テナント名>**」 ボタンをクリック  
（管理者権限が必要）

---

## 設定画面でグレーアウトする理由

クライアント側のEntraIDアプリケーションで権限を追加する画面：

「**Add a permission** → **My APIs** → APIを選択 → 次の画面でDelegated/Application を選ぶ」

この画面で **Delegated permissions / Application permissions タブが選択不可（グレーアウト）** になっていることがある。

### 原因と対処

| グレーアウトしているタブ | 原因 | 対処 |
|---|---|---|
| **Delegated permissions** | API側の「Expose an API」でScopeが1つも定義されていない | API側で `Add a scope` を実行 |
| **Application permissions** | API側の「App roles」で **Allowed member types: Applications**（またはBoth）のロールが1つも定義されていない | API側で `Create app role`（Applications指定）を実行 |
| **両方グレーアウト** | API側でScopeもApp rolesも未定義 | 両方定義する |
| **設定したのにグレーアウト** | ブラウザのキャッシュ問題 | Ctrl+Shift+R でハードリロードして再度開く |

つまり「クライアント側で何かを要求したい」なら、まずAPI側でその「何か」を定義する必要がある、ということ。  
このドキュメント冒頭の「メニューにないものは注文できない」が再度ここで効いてくる。

> [!NOTE]
> 「My APIs」タブに対象のAPIが表示されない場合は、別の問題：
> API側のEntraIDアプリケーションで「Expose an API」の **Application ID URI が未設定** か、
> Scopeが1つも定義されていない可能性が高い。

---

## 実際の設定の流れ

### パターンA: ユーザー認証用（認可コードフロー向け）

```
[API側のEntraIDアプリケーション]
  Expose an API タブ
    ├─ Application ID URI 設定 (例: api://my-api)
    └─ Add a scope
        ├─ Scope name: "Data.Read"
        └─ Who can consent: Admins and users
              ↓
         (これでメニューが用意された)

[クライアント側のEntraIDアプリケーション]
  API permissions タブ
    └─ Add a permission
        └─ My APIs → API側のEntraIDアプリケーションを選択
            └─ Delegated permissions タブ ← ここが選択可能になる
                └─ ☑ Data.Read を選択
                  ↓
              Add permissions
                  ↓
              Grant admin consent (or ユーザー同意)

[実際の認可フロー]
  ブラウザでEntraIDログイン → 同意画面 → 承諾
        ↓
  トークン取得、scp クレームに "Data.Read" が入る
```

### パターンB: M2M用（クライアントクレデンシャルフロー向け）

```
[API側のEntraIDアプリケーション]
  Expose an API タブ
    └─ Application ID URI 設定 (例: api://my-api)
        ※ Scopeは不要（.defaultで動く）
  
  App roles タブ
    └─ Create app role
        ├─ Display name: "Data Reader"
        ├─ Allowed member types: Applications  ← 重要
        └─ Value: "Data.Read"
              ↓
         (これでメニューが用意された)

[クライアント側のEntraIDアプリケーション]
  Certificates & secrets タブ
    └─ New client secret  ← クラクレフロー用に必須
  
  API permissions タブ
    └─ Add a permission
        └─ My APIs → API側のEntraIDアプリケーションを選択
            └─ Application permissions タブ ← ここが選択可能になる
                └─ ☑ Data.Read を選択
                  ↓
              Add permissions
                  ↓
              Grant admin consent ← 必須

[実際のトークン取得]
  scope=api://my-api/.default + grant_type=client_credentials
        ↓
  トークン取得成功、roles クレームに ["Data.Read"] が入る
```

---

## 用途別ベストプラクティス

### A. M2M（クラクレフロー）専用

API側で必要なもの：
```
Expose an API:
  └─ Application ID URI のみ設定
     （Scopeは不要、.defaultで動くため）

App roles:
  └─ App role: "Data.Read" (Allowed member types: Applications)
```

クライアント側で必要なもの：
```
Certificates & secrets:
  └─ Client secret

API permissions:
  └─ Application permissions: Data.Read
  └─ Grant admin consent (必須)
```

### B. ユーザー認証（認可コードフロー）専用

API側で必要なもの：
```
Expose an API:
  ├─ Application ID URI
  └─ Scope: "Data.Read"
```

クライアント側で必要なもの：
```
Authentication:
  └─ Web platform + Redirect URI (Confidential Client)
     または Mobile and desktop applications + Redirect URI (Public Client)

Certificates & secrets:
  └─ Client secret (Web platformの場合のみ必須)

API permissions:
  └─ Delegated permissions: Data.Read
  └─ Grant admin consent (or ユーザー同意)
```

### C. 両方サポート

API側で必要なもの：
```
Expose an API:
  ├─ Application ID URI
  └─ Scope: "Data.Read" (認可コード用)

App roles:
  └─ App role: "Data.Read" (Applications) (クラクレ用)
```

> [!NOTE]
> ScopeとApp roleは同じ名前 (`Data.Read`) で並行定義できる。
> トークンに入るクレームが違う（`scp` vs `roles`）ので、別物として扱われる。

クライアント側は**用途別にEntraIDアプリケーションを分ける**のがおすすめ：

```
クライアントアプリA (バッチ用):
  └─ Application permissions: Data.Read

クライアントアプリB (Webアプリ用):
  └─ Delegated permissions: Data.Read
```

> [!NOTE]
> 同じクライアント側のEntraIDアプリケーションで両方を持たせることも技術的には可能。
> ただし運用上は分けた方が安全：
> - 権限境界が明確になる
> - 片方だけ無効化が可能
> - 監査ログでフロー識別が容易（`appid` で判別可能）

---

## トークンによる動作確認

トークン取得後、[jwt.io](https://jwt.io) でデコードして以下を確認すると、設定が正しいか分かる。

### M2M（クラクレフロー）のトークン例

正しく Application permissions が効いている場合：
```json
{
  "aud": "api://my-api",
  "iss": "https://sts.windows.net/<tenant-id>/",
  "appid": "<クライアント側のEntraIDアプリケーションのclient_id>",
  "idtyp": "app",
  "roles": ["Data.Read"]   ← ★これが入っていればOK
}
```

### ユーザー認証（認可コードフロー）のトークン例

正しく Delegated permissions が効いている場合：
```json
{
  "aud": "api://my-api",
  "iss": "https://sts.windows.net/<tenant-id>/",
  "sub": "<ユーザーのObject ID>",
  "name": "山田太郎",
  "preferred_username": "yamada@example.com",
  "scp": "Data.Read"   ← ★これが入っていればOK
}
```

`roles` や `scp` が入っていない場合、API permissions 設定が正しく反映されていないか、
Grant admin consent が未実行の可能性が高い。

---

## まとめ

EntraIDの認可は「**API側でメニューを作って、クライアント側で注文する**」というモデル。  
「メニュー」が2種類（Scope と App roles）あって、用途別（ユーザー認証 vs M2M）で使い分ける。

| キーワード | 何の話か | 対応するフロー | トークンクレーム |
|---|---|---|---|
| **Expose an API → Scope** | API側で「ユーザー認証用」の権限を定義 | 認可コード | `scp` |
| **App roles (Applications)** | API側で「M2M用」の権限を定義 | クラクレ | `roles` |
| **Delegated permissions** | クライアント側で「ユーザーの代わりに」と要求 | 認可コード | `scp` |
| **Application permissions** | クライアント側で「アプリとして」と要求 | クラクレ | `roles` |
| **Grant admin consent** | 管理者が要求を承認する | 両方 | - |

このマッピングを覚えておけば、「Application permissions タブが選択不可 → API側にApp roles (Applications) が定義されてないんだな」と即座に判断できるようになる。

---

## トラブルシュート早見表

### 設定画面の問題

| 症状 | 原因 | 対処 |
|---|---|---|
| 「My APIs」にAPIが表示されない | API側で「Expose an API」のApplication ID URIが未設定、またはScope/App roleが未定義 | API側でApplication ID URIを設定し、必要に応じてScope/App roleを定義 |
| Delegated permissions タブが選択不可 | API側で「Expose an API」のScopeが未定義 | API側で `Add a scope` を実行 |
| Application permissions タブが選択不可 | API側で「App roles」のApplicationsタイプのロールが未定義 | API側で `Create app role`（Applications指定）を実行 |
| 両方のタブが選択不可 | API側でScopeもApp roleも未定義 | 必要な方を定義（または両方） |
| 設定したのにタブが選択不可 | ブラウザキャッシュ問題 | Ctrl+Shift+R でハードリロード |

### トークン関連の問題

| 症状 | 原因 | 対処 |
|---|---|---|
| トークン取得は成功するが `roles` クレームが入らない | Grant admin consent が未実行、またはApplication permissionsの追加忘れ | クライアント側のAPI permissions画面を確認、`Grant admin consent` を実行 |
| トークン取得は成功するが `scp` クレームが入らない | Delegated permissionsの追加忘れ、または同意画面でユーザーが拒否 | クライアント側のAPI permissions画面を確認、再度ログインして同意 |
| `AADSTS65001: admin consent required` | Grant admin consent が未実行 | クライアント側のAPI permissions画面で `Grant admin consent` をクリック |
| `AADSTS50105: user is not assigned to a role` | App role を Users/Groups タイプで定義していて、ユーザーへのロール割り当てが未実行 | EntraIDの「Enterprise applications」→ 該当アプリ → 「Users and groups」でユーザーにロールを割り当てる |