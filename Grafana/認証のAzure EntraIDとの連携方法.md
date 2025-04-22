- 参考URL
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/azuread/
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/
  - https://grafana.com/blog/2024/07/31/an-overview-of-grafana-sso-benefits-recent-updates-and-best-practices-to-get-started/

# 設定内容
- 主に以下のページの通り設定すればOK
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/azuread/
- 認証方法として「__*Client secrets*__」と「__*Federated credentials*__」の2つがある
  - 本手順は「__*Client secrets*__」方式の設定手順
- `[auth.azuread]`フィールドの`allow_assign_grafana_admin`は`true`に設定する必要がある
- 以下の5つ以外は環境変数としては設定できず、`.ini`ファイルで設定しないといけないっぽい  
  - https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/azuread/#enable-azure-ad-oauth-in-grafana  
  ```
  GF_AUTH_AZUREAD_CLIENT_AUTHENTICATION
  GF_AUTH_AZUREAD_CLIENT_ID
  GF_AUTH_AZUREAD_CLIENT_SECRET
  GF_AUTH_AZUREAD_MANAGED_IDENTITY_CLIENT_ID
  GF_AUTH_AZUREAD_FEDERATED_CREDENTIAL_AUDIENCE
  ```
- `root_url`（環境変数だと`GF_SERVER_ROOT_URL`）にはEntraIDのAPP Registrationsの*Redirect URIs*で設定するURLを設定する（`/login/azuread`がない方のURL）  
  > Verify that the Grafana `root_url` is set in your Azure Application Redirect URLs.


> [!CAUTION]
> Negative potential consequences of an action.