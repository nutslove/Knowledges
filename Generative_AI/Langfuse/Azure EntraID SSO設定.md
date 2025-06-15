# 手順
- 参考URL
  - https://langfuse.com/self-hosting/authentication-and-sso#azure-adentra-id
  - https://github.com/orgs/langfuse/discussions/4591
  - https://github.com/orgs/langfuse/discussions/4764

### 1. Enterpise Applicationsで新規アプリケーション作成/設定
- Entra IDで、「Manage」タグで「Enterpise applications」を開く
- 画面左上の「＋New Application」をクリック
- 「Create your own application」を選択し、アプリケーション名を入力して「Create」をクリック
  - 「What are you looking to do with your application?」は「Integrate any other application you don't find in the gallery（Non-gallery）」を選択
- 作成後のApplicationに入って「Manage」タグの「Users and groups」を開く
  - 画面左上の「+Add user/group」をクリックし、紐づけたいGroupやUserを選択しAssignをクリック
- 「Manage」タグの「Owners」タブに入って画面左上の「＋Add」をクリックし、自分を追加
- 「Manage」タグで「Single sign-on」を選択し、「SAML」をクリック
- １の「Basic SAML Configuration」の「Edit」をクリックし、以下を設定後「Save」を押下
  - Identifier (Entity ID): 任意の値（e.g. "plat-sandbox-langfuse"）
  - Reply URL (Assertion Consumer Service URL): `https://<Langfuseにアクセスするために設定したドメイン名>/api/auth/callback/azure-ad`

### 2. App registrations設定
- Entra IDで、「Manage」タグで「App registrations」を開く
- 作成したApplicationの画面に入る
- 「Overview」タブの「Redirect URIs」をクリックし、「Redirect URIs」に１．で設定した「Reply URL」の値が設定されていることを確認
- 「Certificates & secrets」タブに入り、「Client secrets」タブで「＋New client secret」を押下
  - 「Description」には任意の値を入力し、「Expires」はTokenの有効期限を選択し、「Add」をクリック
    - **このとき、生成されるValue列のシークレットの値をコピーしておくこと！（後からは確認できない）**（後でHelmの`values.yaml`に設定する必要がある）
- 「Overview」タブで「Application (client) ID」と「Directory (tenant) ID」をコピーしておく（後でHelmの`values.yaml`に設定する必要がある）

### 3. Helmの`values.yaml`に以下を追加する
- `langfuse.additionalEnv`の下に以下の環境変数を追加
  - `AUTH_AZURE_AD_CLIENT_ID`
    - 手順2.で確認した「Application (client) ID」
  - `AUTH_AZURE_AD_TENANT_ID`
    - 手順2.で確認した「Directory (tenant) ID」
  - `AUTH_AZURE_AD_CLIENT_SECRET`
    - 手順2.で確認した「Client secret」の値
- `langfuse.nextauth.url`と`NEXTAUTH_URL`環境変数両方設定しないといけない（設定する値は「*https://<Langfuseにアクセスするために設定したドメイン名>*」で一緒）
  - `langfuse.nextauth.url`を設定しないとログイン時「*http://localhost:3000/api/auth/callback/azure-ad* 」 にリダイレクトされてしまう
- `values.yaml`の例（External Secret OperatorでSecretを設定した場合の設定例）  
  ```yaml
  langfuse: 
    encryptionKey: # `openssl rand -hex 32`で生成
      secretKeyRef:
        name: langfuse-auth
        key: encryptionKey
    salt: # `openssl rand -base64 32`で生成
      secretKeyRef:
        name: langfuse-auth
        key: salt
    nextauth: # `openssl rand -base64 32`で生成
      secret:
        secretKeyRef:
          name: langfuse-auth
          key: nextauth-secret
      url: https://<Langfuseにアクセスするために設定したドメイン名>
    serviceAccount:
      create: true
      name: langfuse-serviceaccount # Pod IdentityでこのServiceAccountに対してS3の権限を与えること
    additionalEnv:
      - name: AUTH_AZURE_AD_CLIENT_ID
        valueFrom:
          secretKeyRef:
            name: langfuse-auth
            key: azure-client-id
      - name: AUTH_AZURE_AD_CLIENT_SECRET
        valueFrom:
          secretKeyRef:
            name: langfuse-auth
            key: azure-client-secret
      - name: AUTH_AZURE_AD_TENANT_ID
        valueFrom:
          secretKeyRef:
            name: langfuse-auth
            key: azure-tenant-id
      - name: NEXTAUTH_URL
        value: https://<Langfuseにアクセスするために設定したドメイン名>

  postgresql:
    deploy: false
    host: <RDSのエンドポイント>
    auth:
      database: postgres_langfuse # default database name for langfuse（事前にRDSに入って作成しておく必要がある）
      username: postgres
      existingSecret: langfuse-rds-auth
      secretKeys:
        userPasswordKey: password

  clickhouse:
    deploy: true
    auth:
      existingSecret: langfuse-auth
      existingSecretKey: clickhouse-password

  redis:
    deploy: true
    auth:
      existingSecret: langfuse-auth
      existingSecretPasswordKey: redis-password

  s3:
    deploy: false
    bucket: <Langfuseのデータを保存するS3バケット名>
  ```