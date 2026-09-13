https://swagger.io/specification/

## 1. OpenAPIとは

**OpenAPI Specification（OAS）** は、REST APIの仕様（エンドポイント、パラメータ、リクエスト/レスポンスの形、認証方式など）を **プログラムで読み書きできる形式（YAML/JSON）** で記述するための標準規格。

- 旧称「Swagger」。2015年にSmartBearがSwagger仕様をLinux Foundation配下の **OpenAPI Initiative** に寄贈し、以降「OpenAPI Specification」と呼ばれる（Swagger UI/Editor/Codegenなどのツール名にはSwaggerの名が残っている）
- 目的は「**人間にもマシンにも読める、単一のAPI仕様書**」を作ること
- この仕様書（＝ドキュメント1本）があれば、そこから以下が自動生成できる
  - APIドキュメント（Swagger UI, ReDoc）
  - サーバースタブ・クライアントSDK（openapi-generator等）
  - モックサーバー
  - テストコード
  - AI Agentが呼び出すツール定義（MCP, Function Callingなど）

### バージョンの変遷

| バージョン | 特徴 |
|---|---|
| Swagger 2.0 | 旧名称。`swagger: "2.0"` から始まる。今も現役プロジェクトが多い |
| OpenAPI 3.0 | `components` の導入、`requestBody` の分離など大幅刷新 |
| OpenAPI 3.1 | JSON Schema 2020-12と完全互換に。`webhooks`対応。現時点の最新系統 |

FastAPIなど最近のフレームワークは3.0系（一部3.1系）を自動生成する。

---

## 2. ドキュメントの基本構造

OpenAPIドキュメントはYAMLかJSONで書く（内容は同一、表記が違うだけ）。トップレベルの主要フィールドは以下:

```yaml
openapi: 3.0.3
info:
  title: Sample API
  version: 1.0.0
  description: サンプルAPIの説明
servers:
  - url: https://api.example.com/v1
paths:
  /users:
    get:
      ...
components:
  schemas: {}
  securitySchemes: {}
tags:
  - name: users
security:
  - bearerAuth: []
```

| フィールド | 役割 |
|---|---|
| `openapi` | OpenAPI Specification自体のバージョン（例: `3.0.3`）。`info.version`（APIのバージョン）とは別物なので注意 |
| `info` | API名・バージョン・説明などのメタ情報 |
| `servers` | ベースURL（複数指定可。本番/検証などの切り替え） |
| `paths` | **エンドポイント定義の本体**（パス→HTTPメソッド→詳細） |
| `components` | 再利用可能な部品置き場（スキーマ、パラメータ、認証方式など） |
| `tags` | エンドポイントのグルーピング（Swagger UI上の見出し） |
| `security` | API全体に適用するデフォルトの認証要件 |

---

## 3. `paths`: エンドポイントの定義

`paths` はURLパスごとに、HTTPメソッド（`get`/`post`/`put`/`delete`等）ごとの「Operation Object」を持つ。

```yaml
paths:
  /users/{user_id}:
    get:
      summary: ユーザー取得
      operationId: getUser
      tags: [users]
      parameters:
        - name: user_id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: 成功
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "404":
          description: ユーザーが存在しない
    put:
      summary: ユーザー更新
      operationId: updateUser
      tags: [users]
      parameters:
        - name: user_id
          in: path
          required: true
          schema:
            type: integer
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/UserUpdate"
      responses:
        "200":
          description: 更新成功
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "404":
          description: ユーザーが存在しない
```

`put`（更新）の例のように、`parameters`（対象を特定するpathパラメータ）と `requestBody`（更新したいデータ）は**同じOperationに同時に書ける**。両者は役割が違うため排他ではなく、「対象の特定はparameters、送信データはrequestBody」という組み合わせは実務でごく普通に使われる。

| キー | 意味 |
|---|---|
| `summary` / `description` | ドキュメント上の説明文 |
| `operationId` | このOperationの一意なID。コード生成時の関数名になる |
| `tags` | Swagger UIでの章立て |
| `parameters` | パス/クエリ/ヘッダー/Cookieパラメータ（4章） |
| `requestBody` | リクエストボディの形（5章） |
| `responses` | ステータスコードごとのレスポンス形（6章） |
| `security` | このOperation固有の認証要件（未指定ならトップレベルの`security`を継承） |

---

## 4. パラメータ（`parameters`）

パラメータは **どこに乗るか（`in`）** で4種類に分かれる。

| `in` の値 | 由来 | 例 |
|---|---|---|
| `path` | URLパスの一部（`{user_id}`） | `/users/{user_id}` |
| `query` | クエリ文字列 | `?skip=0&limit=10` |
| `header` | HTTPヘッダー | `X-Request-Id: abc` |
| `cookie` | Cookie | `session=xyz` |

```yaml
parameters:
  - name: limit
    in: query
    required: false
    schema:
      type: integer
      default: 10
      minimum: 1
      maximum: 100
    description: 取得件数の上限
```

- `path` パラメータは常に `required: true`
- `schema` に型・デフォルト値・バリデーション制約（`minimum`, `pattern`, `enum`等）を書ける

---

## 5. リクエストボディ（`requestBody`）

OpenAPI 3.0で `parameters` から分離された。Content-Typeごとにスキーマを持てる（JSON以外に`multipart/form-data`等も表現可能）。

```yaml
requestBody:
  required: true
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/UserCreate"
```

---

## 6. レスポンス（`responses`）

ステータスコードをキーにしたマップ。`default` でその他すべてのコードを表現できる。

```yaml
responses:
  "200":
    description: 成功
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/User"
  "422":
    description: バリデーションエラー
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/HTTPValidationError"
  default:
    description: 予期しないエラー
```

> [!NOTE]
> **`default` は特定のステータスコードを表すキーではない**
>
> `default` は、他に明示的に定義されていない**すべてのステータスコードに対するフォールバック**を意味する。実際のレスポンスコードが何であれ、そのコード専用の定義（`"200"`, `"404"` など）が無ければ `default` が適用される。主に「個別に書ききれない/書く必要のないエラー全般（500系や想定外のエラーなど）」をまとめて表現する用途で使う。

---

## 7. `components`: 再利用可能な部品

`paths` の各所から `$ref` で参照される定義をまとめて置く場所。重複記述を避け、保守性を上げるための仕組み。

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
        email:
          type: string
          format: email
      required: [id, name]
    UserUpdate:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
          format: email
  parameters:
    LimitParam:
      name: limit
      in: query
      schema:
        type: integer
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
    apiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/oauth/authorize
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: 読み取り権限
            write: 書き込み権限
```

| サブフィールド | 内容 |
|---|---|
| `schemas` | データモデル（JSON Schemaベース） |
| `parameters` | 使い回すパラメータ定義 |
| `requestBodies` | 使い回すリクエストボディ定義 |
| `responses` | 使い回すレスポンス定義 |
| `securitySchemes` | 認証方式の定義（後述） |

### `$ref` による参照

```yaml
schema:
  $ref: "#/components/schemas/User"
```

`#/components/schemas/User` は「このドキュメント内の `components.schemas.User` を参照する」という意味（JSON Pointer構文）。別ファイルを参照する分割構成も可能（`./schemas/user.yaml#/User`など）。

---

## 8. `schema`: データ構造の表現（JSON Schemaベース）

OpenAPIのスキーマは **JSON Schema** をベースに、OpenAPI独自の拡張を加えたもの（3.1では素のJSON Schemaと完全互換）。

```yaml
User:
  type: object
  properties:
    id:
      type: integer
    role:
      type: string
      enum: [admin, member, guest]
    tags:
      type: array
      items:
        type: string
    address:
      $ref: "#/components/schemas/Address"
  required: [id, role]
```

| 型 | 例 |
|---|---|
| `string` | `format: date-time`, `format: email`, `pattern` |
| `integer` / `number` | `minimum`, `maximum` |
| `boolean` | - |
| `array` | `items` で要素の型を指定 |
| `object` | `properties` + `required` |

継承・合成には `allOf`（全条件を満たす）、`oneOf`（いずれか1つ）、`anyOf`（1つ以上）が使える。

```yaml
Pet:
  oneOf:
    - $ref: "#/components/schemas/Dog"
    - $ref: "#/components/schemas/Cat"
```

---

## 9. 認証・認可の表現（`securitySchemes` + `security`）

`components.securitySchemes` で「どんな認証方式が存在するか」を定義し、`security` で「どのOperationにどれを要求するか」を指定する。

```yaml
security:
  - bearerAuth: []      # トップレベル: デフォルトで全エンドポイントに要求

paths:
  /public/health:
    get:
      security: []      # このエンドポイントだけ認証不要にする
  /admin/users:
    get:
      security:
        - oauth2: [read, write]   # OAuth2はscopeを配列で指定
```

- `security: []`（空配列）は「認証不要」を意味する
- 配列内に複数のスキーム名を並べると **AND**（両方満たす必要あり）
- `security` のリスト自体を複数要素にすると **OR**（いずれか1つ満たせばよい）

FastAPIでの認証実装の具体例（OAuth2/JWT、APIキー等）は→[Python/FastAPI/認証・認可（Security）について.md](../Python/FastAPI/認証・認可（Security）について.md)を参照。

---

## 10. ツールエコシステム

| ツール | 用途 |
|---|---|
| **Swagger UI** | OpenAPIドキュメントから対話的なAPIドキュメント画面を生成（`/docs`） |
| **ReDoc** | Swagger UIとは別デザインのドキュメント生成ツール（`/redoc`） |
| **Swagger Editor** | ブラウザ上でOpenAPIドキュメントを書きながらプレビュー |
| **openapi-generator** / **swagger-codegen** | OpenAPIドキュメントからサーバースタブ・クライアントSDKを多言語で自動生成 |
| **Prism** | OpenAPIドキュメントからモックサーバーを立てる |
| **Spectral** | OpenAPIドキュメントのLint（命名規則・必須フィールドチェック等） |

### コードファースト vs スキーマファースト

| アプローチ | 説明 |
|---|---|
| **コードファースト** | コード（型注釈やデコレータ）を書くと、そこからOpenAPIドキュメントを自動生成する。FastAPIはこちら |
| **スキーマファースト（Design First）** | 先にOpenAPI YAML/JSONを手書き・設計し、そこからサーバースタブやクライアントを生成する。API設計をチーム間で先に合意したい場合に向く |

FastAPIは型注釈（Pydanticモデル、`Depends`等）から自動的にOpenAPIドキュメントを組み立てる代表例（→[Python/FastAPI/request, responseのスキーマについて.md](../Python/FastAPI/request,%20responseのスキーマについて.md)、[Python/FastAPI/Path・Query・Bodyパラメータの検証について.md](../Python/FastAPI/Path・Query・Bodyパラメータの検証について.md)）。

---

## 11. AI Agent文脈でのOpenAPIの使われ方

近年は「既存のREST APIをAI Agentのツールとして使わせる」ための橋渡しとしてOpenAPIが使われる場面が増えている。

- **Function Calling / Tool定義の自動生成**: OpenAPIドキュメントの `paths` + `schemas` から、LLMに渡すtool schema（名前・パラメータ・型）を機械的に生成できる
- **MCP（Model Context Protocol）**: 既存のOpenAPI仕様を持つAPIをMCPサーバー化するツールが多数存在する（OpenAPIドキュメントを読み込み、各Operationを1つのMCP toolとして公開）。詳細は→[Generative_AI/MCP・Agentの認証認可.md](../Generative_AI/MCP・Agentの認証認可.md)
- **A2A（Agent-to-Agent）**: Agent自身の能力を外部に公開する`AgentCard`はOpenAPIそのものではないが、「機械可読なAPI/能力記述」という設計思想はOpenAPIと共通する。詳細は→[Generative_AI/A2A/A2Aについて.md](../Generative_AI/A2A/A2Aについて.md)

`operationId` や `description` を丁寧に書いておくと、そのままLLMへのtool説明文として転用できるため、AI Agent連携を見据える場合はOpenAPIドキュメントの説明文の質がそのままツール呼び出し精度に影響する。

---

## 12. まとめ

- OpenAPIは **REST APIの仕様を機械可読な形（YAML/JSON）で記述する標準規格**（旧Swagger）
- 本体は `paths`（エンドポイント）+ `components`（再利用部品）の2本柱
- `parameters`（path/query/header/cookie）と `requestBody` でリクエストの形、`responses` でレスポンスの形を表現
- `schema` はJSON Schemaベースで、型・制約・`$ref`参照・`oneOf`/`allOf`による合成ができる
- `securitySchemes` + `security` で認証方式とその適用範囲を宣言的に表現できる
- 1つのドキュメントからドキュメントサイト・SDK・モック・テスト・AI Agentのtool定義まで自動生成できるのが最大の価値
- FastAPIのような「コードファースト」フレームワークでは、型注釈から自動生成される（手書き不要）
