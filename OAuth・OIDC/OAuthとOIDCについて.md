```mermaid
sequenceDiagram
    participant U as 👤 ユーザー<br/>(Resource Owner)
    participant C as 📱 クライアント<br/>(アプリケーション)
    participant AS as 🔐 認証サーバー<br/>(Authorization Server)
    participant RS as 💾 リソースサーバー<br/>(Resource Server)

    Note over U,RS: OpenID Connect + OAuth 2.0 Authorization Code Flow

    U->>C: 1. ログイン要求<br/>「OpenID Connectでログイン」
    
    Note over C: scope=openid profile email<br/>response_type=code
    C->>AS: 2. 認可リクエスト<br/>GET /oauth/authorize?<br/>response_type=code&<br/>scope=openid profile email&<br/>client_id=xxx&<br/>redirect_uri=xxx

    AS->>U: 3. 認証画面表示<br/>ログインフォーム
    
    U->>AS: 4. 認証情報送信<br/>ユーザー名・パスワード
    
    AS->>U: 5. 認可確認画面<br/>「プロフィール情報の<br/>アクセスを許可しますか？」
    
    U->>AS: 6. 認可同意<br/>「許可」ボタンクリック
    
    AS->>C: 7. 認可コード発行<br/>302 Redirect<br/>https://app.example.com/callback?<br/>code=ABC123...
    
    Note over C: client_secret使用で<br/>バックエンドで安全に実行
    C->>AS: 8. 発行された認可コードでトークン要求<br/>POST /oauth/token<br/>grant_type=authorization_code&<br/>code=ABC123&<br/>client_id=xxx&<br/>client_secret=xxx
    
    AS->>C: 9. IDトークン・アクセストークン発行<br/>・access_token <br/>・id_token (JWT)<br/>・refresh_token
    
    Note over C: IDトークン（JWT）を検証<br/>・署名検証<br/>・有効期限確認<br/>・audience確認
    
    Note over C: IDトークン（JWT）のペイロードから<br/>基本的なユーザー情報を取得<br/>{"sub":"12345","name":"田中太郎","email":"tanaka@example.com"}
    
    opt さらに詳細な情報が必要な場合
        C->>RS: 10. ユーザー情報取得<br/>GET /userinfo<br/>Authorization: Bearer ACCESS_TOKEN
        RS->>C: 11. ユーザー情報レスポンス<br/>追加のユーザー属性情報
    end
    
    C->>U: 12. ログイン完了<br/>ユーザー情報表示
    
    Note over U,RS: 以降のリソースアクセス
    
    U->>C: 13. 保護されたリソース要求<br/>（例：マイページ表示）
    
    C->>RS: 14. アクセストークン付き APIアクセス<br/>GET /api/protected<br/>Authorization: Bearer ACCESS_TOKEN
    
    RS->>C: 15. リソース応答<br/>JSON データ
    
    C->>U: 16. 画面表示<br/>取得したデータを表示
    
    Note over U,RS: トークンリフレッシュ（必要に応じて）
    
    alt アクセストークンの有効期限切れ時
        C->>AS: 17. トークンリフレッシュ<br/>POST /oauth/token<br/>grant_type=refresh_token&<br/>refresh_token=REFRESH_TOKEN
        AS->>C: 18. 新しいトークン発行<br/>・新しいaccess_token<br/>・新しいrefresh_token
    end
```

- **IDトークン**: OIDC固有で、ユーザーの認証情報（`sub`, `email`, `name`など）を含むJWT
- **アクセストークン**: リソースへのアクセス権限（何ができるか）を表すもので、APIアクセス用
- **スコープ**: openidは必須、profileやemailで追加情報を要求
  - 大文字・小文字は区別される
  - 複数のスコープを指定する場合はスペースで区切る
  - 複数のスコープを指定時、スコープの順番は意味を持たない

# OAuth 2.0
- **認可**のプロトコル
- Third-Party アプリケーションに対して、ユーザーの代わりにリソースにアクセスする権限を安全に委譲する

## OAuthの登場人物
#### **リソースオーナー（Resource Owner）**
- 認可を与えるユーザー
#### **クライアント（Client）**
- リソースオーナーの代理としてリソースサーバーにアクセスするアプリケーション
- Scope（範囲）を指定して、アクセス権限を制限することができる
#### **認可サーバー（Authorization Server）**
- リソースオーナーの認証・認可を受けて、クライアントにアクセストークンを発行するサーバー
#### **リソースサーバー（Resource Server）**
- APIを通じてクライアントに保護されたリソースを提供するサーバー
- アクセストークンを検証して、クライアントのアクセスを許可する

# OIDC（OpenID Connect）
- **認証**のプロトコル
- OAuth 2.0の上に構築された認証レイヤー（OAuth 2.0に「ユーザー認証（Who you are）」の要素を追加したもの）

# OIDC/OAuth vs SAML
| 項目 | SAML | OIDC/OAuth |
| --- | --- | --- |
| **データ形式** | XML | JSON |
| **主な用途** | エンタープライズSSO、属性交換 | Web/モバイル認証（OIDC）、API認可（OAuth） |
| **登場人物** | ・Identity Provider（IdP）<br>・Service Provider（SP）<br>・ユーザー | ・Authorization Server<br>・Client（Relying Party）<br>・Resource Server<br>・Resource Owner（ユーザー） |
| **トークン** | **SAML Assertion**<br>・XML形式の認証/認可情報<br>・デジタル署名付き<br>・ユーザー属性を含む | **Access Token**：リソースアクセス用<br>**ID Token（OIDC）**：ユーザー認証情報（JWT）<br>**Refresh Token**：トークン更新用 |
| **通信方式** | ・HTTP POST/Redirect<br>・SOAP over HTTP<br>・ブラウザリダイレクトベース | ・HTTPS REST API<br>・JSON over HTTP<br>・Authorization Headerでトークン送信 |
| **SSO実現方式** | ・SP-Initiated SSO<br>・IdP-Initiated SSO<br>・メタデータ交換による信頼関係構築 | ・Authorization Code Flow<br>・Discovery URLによる動的設定<br>・複数サービス間でのトークン共有 |
| **パフォーマンス** | ・メッセージサイズ：大（2-10KB）<br>・処理速度：遅い（XML解析）<br>・ネットワーク：複数リダイレクト | ・メッセージサイズ：小（0.5-2KB）<br>・処理速度：速い（JSON解析）<br>・ネットワーク：効率的なAPI通信 |