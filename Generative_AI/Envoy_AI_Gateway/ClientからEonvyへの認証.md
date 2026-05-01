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