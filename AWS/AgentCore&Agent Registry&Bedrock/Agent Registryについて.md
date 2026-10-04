# Agent Registryとは
- **AWS Agent Registry**は、組織内の**MCPサーバ、Agent、Skills、Customリソース**を一元的にカタログ化・審査(承認)・検索するための**フルマネージドなディスカバリサービス**
  - 「チームごとにMCPサーバやAgentを作ってしまい、既にあるものが見つけられず重複開発になる」問題を解決するためのもの
- 公式ドキュメント: https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry.html
- 主な特徴
  - **承認ワークフロー**: 管理者/キュレーターに承認されたレコードだけが検索対象になる(自動承認も設定可)
  - **ハイブリッド検索**: セマンティック検索(ベクトル)とキーワード検索を並列実行して結果をマージ
  - **カタログブラウジング**: 承認済みレコードをページネーション/フィルタ付きで一覧、IDで一括取得
  - **MCPネイティブ**: Registry自体がリモートMCPエンドポイントを持つので、MCPクライアント(Claude Code、Kiroなど)から直接検索できる
  - **認可方式**: インバウンドは **IAM(SigV4)** か **JWT(社内IdPのトークン)** を選択(対象はsearch/list/batch-get/MCPエンドポイントといったdata planeのみ。Registryやレコードの作成などcontrol planeは**常にIAM**)
  - **外部ソースとの同期**: MCPサーバ/A2A AgentはURLからメタデータを自動取得(Synchronize)できる

> [!CAUTION]
> - 2026/4に`bedrock-agentcore`ネームスペースでパブリックプレビューとして公開されたが、2026/8/6に新しい **`agent-registry`ネームスペース**(CLIは`aws agent-registry-control` / `aws agent-registry`)が公開された。旧ネームスペースは**2026/10/30で終了**し、読み書きできなくなる。
> - 旧ネームスペースの既存Registry/レコードは**自動では移行されない**。自分で移行ツール([agentcore-samples](https://github.com/awslabs/agentcore-samples/tree/main/01-features/07-centralize-and-govern-your-ai-infrastructure/03-registry/04-migrate-to-new-namespace))を実行する必要がある。API/IAM/EventBridgeなどの変更点は末尾の「旧ネームスペースからの移行メモ」を参照。
> - 2026/8/6時点で既存Registryがなかった新規顧客は、旧ネームスペースを使えない。
> - このノートは`agent-registry`ネームスペースを前提に書く。

## 主要な概念
- **Registry**: AWSアカウント内に作るカタログ。Registryごとに名前、説明、インバウンド認可(IAM or JWT)、承認設定(自動承認 or 手動)を持つ。組織全体で1つ、リソースタイプ別、環境別(prod/QA/dev)、チーム別など自由に分けられる
- **Registry Record**: Registryに登録される個々のリソースのメタデータ。`name` + `recordVersion`の組み合わせがRegistry内でユニークである必要がある(`name`は英数字始まりで`a-zA-Z0-9_-./`が使え、255文字まで。`description`は1〜4,096文字)
- **recordType**: `AGENT` / `MCP` / `SKILL` / `CUSTOM`(手動登録で指定できるのはこの4つ。`GATEWAY`は後述の自動検出でGatewayから作られたレコード専用で、他のタイプには変更できない)
- **descriptor**: レコードの中身。1レコードにつきプライマリdescriptorは1つだけ指定できる

| recordType | 指定できるdescriptor | 中身 |
| --- | --- | --- |
| `MCP` | `mcpServer`, `custom` | MCPの`server.json`(サーバ定義) + tools定義 |
| `AGENT` | `a2aAgentCard`, `mcpServer`, `custom` | A2AのAgent Card(v0.3) |
| `SKILL` | `agentSkillsDefinition`, `custom` | SKILL.md(任意) + スキル定義(任意) |
| `CUSTOM` | `custom` | 任意のJSON |

- 自動検出(後述の「発展トピック」)で作られたRuntimeのレコードでは、descriptorに`http` / `agui`も入る(上の表は手動登録で指定できるdescriptor)

- **Persona**
  - **Administrator**: Registryの作成、認可/承認設定、IAM権限管理
  - **Publisher**: レコードを作成してDraft→承認申請
  - **Curator/Approver**: 申請されたレコードをApprove/Reject/Deprecate
  - **Consumer**: 承認済みレコードの検索/ブラウズ(人間 or Agent)

## レコードのライフサイクル
```
Create → DRAFT → Submit → PENDING_APPROVAL → Approve → APPROVED
                               │                          │
                               │ Reject                   │ Edit (新しいDRAFTリビジョン。
                               ▼                          │  承認済みリビジョンは検索可能なまま)
                          REJECTED ── Approve(直接) ──────┘
                               │
                               └── Edit → DRAFT

         どのステータスからでも → DEPRECATED (終端。元に戻せない)
```
- **検索/一覧/MCPエンドポイントに出てくるのはAPPROVEDなレコードだけ**(Draft/Pending/Rejected/Deprecatedは出ない)
- 承認済みレコードを編集すると**新しいDRAFTリビジョン**ができ、再承認されるまでは**古い承認済みリビジョンが検索され続ける**(dual-revision)
- 承認済みレコードを一時的に隠したいときは、Deprecateではなく**Reject**する(Deprecateは元に戻せないため)
- 承認後、検索インデックスに反映されるまで**数秒〜数分**かかる(結果整合)。直後に出なくても`GetRegistryRecord`でステータスを確認する
- Registryの自動承認設定がONなら、Submitした時点で即APPROVEDになる。OFFなら`PENDING_APPROVAL`になりEventBridgeに通知が飛ぶ
- 同期(Synchronize)を伴う作成/更新では、処理中の一時的なステータスがある。作成中は`CREATING`、再同期中は`UPDATING`で、成功すると`DRAFT`になる。失敗すると`CREATE_FAILED` / `UPDATE_FAILED`になり、`statusReason`に原因が出る

## 承認・却下・非推奨(Curator)
- 申請中(`PENDING_APPROVAL`)のレコードを一覧する

```bash
aws agent-registry-control list-registry-records \
  --registry-id "<registryId>" \
  --filters '[{"name": "status", "values": ["PENDING_APPROVAL"]}]' \
  --region us-east-1
```

- 承認/却下/非推奨は同じAPI(`update-registry-record-status`)で`--status`を変えるだけ。コンソールでは**Update status**から行う。`--status-reason`に理由を残せる

```bash
# 承認
aws agent-registry-control update-registry-record-status \
  --registry-id "<registryId>" \
  --record-id "<recordId>" \
  --status APPROVED \
  --status-reason "Reviewed and approved" \
  --region us-east-1

# 却下 (Publisherは修正して再申請できる。Curatorが却下済みを直接承認することも可能)
aws agent-registry-control update-registry-record-status \
  --registry-id "<registryId>" \
  --record-id "<recordId>" \
  --status REJECTED \
  --status-reason "Missing tool input schemas" \
  --region us-east-1

# 非推奨 (どのステータスからでも可。終端で、元に戻せない)
aws agent-registry-control update-registry-record-status \
  --registry-id "<registryId>" \
  --record-id "<recordId>" \
  --status DEPRECATED \
  --status-reason "Replaced by v2" \
  --region us-east-1
```

- 非推奨にしたレコードも`ListRegistryRecords` / `GetRegistryRecord`では見える(監査用)が、編集も他のステータスへの遷移もできない
- レコードの削除は`delete-registry-record`(永久削除で、元に戻せない)

## Registryの作成
- レコードを登録する前に、まずRegistryを作る。以降の例は`$REGISTRY_ID`にRegistryのIDが入っている前提
- 名前は英数字始まりで最大64文字、ステータスは`Creating` → `Ready`
- **自動承認(Auto-approval)** のON/OFFと、**インバウンド認可**(IAM or JWT)を作成時に決める
  - インバウンド認可のタイプ(IAM/JWT)と、JWTのdiscovery URLは**作成後に変更できない**(JWTの`allowedClients`/`allowedAudience`等は更新可能)
  - 自動承認のON/OFFは後から変更できるが、**変更後に申請されたレコードにだけ**効く(既に`PENDING_APPROVAL`のものは対象外)
  - APIのデータモデルでは`approvalConfiguration.autoApprovalRules: ["APPROVE_ALL"]`で自動承認、空配列`[]`で手動承認(旧ネームスペースの`autoApproval: true/false`の置き換え)。`create-registry`のCLIフラグ名は公式ドキュメントに例がなく未確認なので、`aws agent-registry-control create-registry help`で確認すること
- KMSのカスタマーマネージドキーによる暗号化は作成時のみ指定可能(後から変更不可。デフォルトはAWS所有キー)
- 削除するには、先にRegistry内の**全レコードを削除**しておく必要がある

```bash
# IAM認可のRegistry
aws agent-registry-control create-registry \
  --name "MyRegistry" \
  --description "Production registry" \
  --region us-east-1
```

```bash
# JWT認可のRegistry (例: Cognito)
aws agent-registry-control create-registry \
  --name "MyOAuthRegistry" \
  --discovery-configuration '{"authorizerType": "CUSTOM_JWT", "authorizerConfiguration": {"customJWTAuthorizer": {"discoveryUrl": "https://cognito-idp.us-east-1.amazonaws.com/<poolId>/.well-known/openid-configuration", "allowedClients": ["<appClientId>"]}}}' \
  --region us-east-1
```

- JWTの場合は`allowedAudience` / `allowedClients` / `allowedScopes` / カスタムクレームの**いずれか1つ以上**が必須(複数指定したら全て検証される)
- コンソールでの検索/ブラウズ(Record directory)は**IAM認可のRegistryのみ**対応。JWT認可の場合はAPIを直接(curl等)叩くか、MCPエンドポイントを使う
- AWS CLI/SDKはSigV4署名を使うため、**JWT認可のRegistryにはCLI/SDKでdata plane(検索など)を呼べない**。Bearerトークン付きでHTTPを直接叩く(後述の「Skillsの取得」にcurl例がある)
- レコードに付ける独自の属性は、Registryに**カスタムメタデータスキーマ**を定義して扱う(次節)

## カスタムメタデータ
- Registryごとに、レコードに付けられる独自の属性(担当チーム、社内レビュー状況、コンプライアンスタグなど)のスキーマを定義できる。検索の`filters`で`customMetadata.{field}`として絞り込める
- スキーマは**JSON Schema (draft-07) のごく限られたサブセット**。フラットなobjectで、各フィールドは次の4形式のみ(ネスト、配列、`$ref`、`default`/`pattern`などは不可。`$schema`キーも不可)
  - Text: `{"type":"string"}`
  - Enum: `{"type":"string","enum":[...]}`
  - URL: `{"type":"string","format":"uri"}`
  - Boolean: `{"type":"boolean"}`
- 全レコードタイプ共通の`defaultSchema`と、レコードタイプ別の`recordTypeSchemaOverrides`(最大5つ)を持てる。overrideがあるタイプはそちらが優先される
- **追加のみ可能(additive-only)**。フィールドの追加、Enum値の追加、requiredの変更はできるが、保存済みフィールドの削除/型変更、Enum値の削除はできない。**一度設定したスキーマは消せない**ので、設計は慎重に行う
- スキーマを変えても既存レコードは変更されず、準拠しなくなったレコードは`customMetadataSchemaComplianceStatus: NON_COMPLIANT`と報告される(承認ステータスには影響しない)
- 値は文字列(128文字まで)かJSONのbooleanのみ(Booleanに`"true"`という文字列は不可)。1スキーマあたり最大15フィールド。スキーマにないキーは拒否される
- 値の指定はレコードの作成/更新時は任意だが、**`SubmitRegistryRecordForApproval`で必ずスキーマ検証される**。requiredフィールドがあるのに値がないと、承認申請で弾かれる

```bash
# スキーマ付きでRegistryを作成
aws agent-registry-control create-registry \
  --name "my-registry" \
  --custom-metadata-schema-configuration '{
    "defaultSchema": "{\"type\":\"object\",\"properties\":{\"owner\":{\"type\":\"string\"}},\"required\":[\"owner\"]}",
    "recordTypeSchemaOverrides": [
      {
        "recordType": "MCP",
        "schema": "{\"type\":\"object\",\"properties\":{\"tier\":{\"type\":\"string\",\"enum\":[\"internal\",\"partner\",\"public\"]}}}"
      }
    ]
  }' \
  --region us-east-1
```

```bash
# レコードの作成時に値を指定 (公式の例をスキーマに合わせて変更したもの)
aws agent-registry-control create-registry-record \
  --registry-id "<registryId>" \
  --name "my-mcp-server" \
  --record-type MCP \
  --descriptors '{"mcpServer": {"data": "{\"name\":\"my/mcp-server\",\"description\":\"My MCP server\",\"version\":\"1.0.0\"}", "dataSchemaVersion": "2025-12-11"}}' \
  --record-version "1.0" \
  --custom-metadata '{"owner": "search-team", "tier": "internal"}' \
  --region us-east-1

# 既存レコードの値を更新 (マップ全体の置き換え。optionalValueで包む)
aws agent-registry-control update-registry-record \
  --registry-id "<registryId>" \
  --record-id "<recordId>" \
  --custom-metadata '{"optionalValue": {"owner": "search-team", "tier": "partner"}}' \
  --region us-east-1
```

- スキーマを後から変える`update-registry`は、既存のフィールドとoverrideを**すべて含めた**完全な置き換えで、`optionalValue`で包む(破壊的な変更は拒否される)

# Skills
- `recordType: SKILL`、descriptorは`agentSkillsDefinition`
- 構成
  - **skillMd** (任意): `SKILL.md`の中身。[AgentSkills仕様](https://agentskills.io/home)に対して検証される。`descriptors.agentSkillsDefinition.additionalData.skillMd.data`に入れる
  - **スキル定義** (任意): `descriptors.agentSkillsDefinition.data`に入れる。`dataSchemaVersion`は`0.1.0`。全フィールド任意
    - `repository` (`url`, `source`は必須。例: `github`, `gitlab`, `codecommit`)
    - `websiteUrl`
    - `packages[]` (`registryType`, `identifier`は必須。`version`は具体的なバージョン)
- **重要な制約**: Registryに保存されるのはあくまで**ディスカバリ用のメタデータ**。`SKILL.md`以外のスキルのファイル(scripts等)は保存されない。実体はリポジトリやパッケージ側に置き、そこへのポインタ(`repository`/`packages`)を登録する
- **URLからの同期(Synchronize)は非対応**(同期はMCPとAgentのみ)。手動(Manual)で登録する

## Skillsの登録/更新
### 登録
- コンソール: Registry詳細 → **Create record** → **Manual** → Typeに**Skills** → Descriptorに**Agent skills definition**
- CLIの例(descriptorの構造は、[CreateRegistryRecordのAPIリファレンス](https://docs.aws.amazon.com/agent-registry-control/latest/APIReference/API_CreateRegistryRecord.html)のリクエスト構文と一致することを確認した。公式にSkill用のCLI例はなく、MCPの例と同じ形式で組み立てたもので、実際には実行していない)

```bash
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "my-skill" \
  --display-name "My Skill" \
  --description "Brief description of what this skill does." \
  --record-type SKILL \
  --record-version "1.0.0" \
  --descriptors '{
    "agentSkillsDefinition": {
      "data": "{\"websiteUrl\": \"https://example.com/my-skill\", \"repository\": {\"url\": \"https://github.com/example/my-skill\", \"source\": \"github\"}}",
      "dataSchemaVersion": "0.1.0",
      "additionalData": {
        "skillMd": {
          "data": "---\nname: my-skill\ndescription: Brief description of what this skill does.\n---\n\n# My Skill\n\nDescribe your skill purpose, usage, and capabilities here."
        }
      }
    }
  }' \
  --region us-east-1
```

- `SKILL.md`の最小例

```markdown
---
name: my-skill
description: Brief description of what this skill does.
---

# My Skill

Describe your skill's purpose, usage, and capabilities here.
```

### 承認申請
- 作成直後は`DRAFT`。公開(検索可能に)するには承認申請が必要
- コンソールなら**Create and submit for approval**で作成と申請を同時にできる

```bash
aws agent-registry-control submit-registry-record-for-approval \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --region us-east-1
```

- Curatorは`UpdateRegistryRecordStatus`(コンソールの**Update status**)でApprove/Reject/Deprecateする

### 更新
- `update-registry-record`で変更する

```bash
aws agent-registry-control update-registry-record \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --description '{"optionalValue": "Updated description"}' \
  --region us-east-1
```

- `UpdateRegistryRecord`は、`description` / `displayName` / `customMetadata` / `descriptors`を`optionalValue`で包む必要がある(PATCH形式。省略したフィールドは変更されない)。公式のAPIリファレンスで確認済み。詳細は下の「UpdateRegistryRecordの落とし穴」を参照
  - 開発者ガイドのCLI例には、包まない`--description "Updated description"`という書き方があるが、APIリファレンスとは食い違う。包む形式を使うこと
- 更新時のステータスの挙動は前述の「レコードのライフサイクル」のとおり
  - `DRAFT`: その場で更新
  - `APPROVED`: 新しいDRAFTリビジョンができ、再承認までは旧リビジョンが検索される
  - `DEPRECATED`: 編集不可
- Skillsは同期非対応なので、スキルの内容が変わるたびに**自分で`update-registry-record`して再度承認申請**する必要がある(CI/CDに組み込むのが現実的)

## Skillsの取得
- 承認済みのものだけが対象。`recordType`で`SKILL`に絞る
- **検索**(自然言語、ハイブリッド検索。応答の各レコードに**`descriptors`も含まれる**ので、`skillMd`などの本体までこの1回で取れる。APIリファレンスの応答構文で確認済み)

```bash
aws agent-registry search-discoverable-registry-records \
  --search-query "PDFからテキストを抽出する" \
  --registry-ids "<registryARN>" \
  --filters '{"recordType": {"$eq": "SKILL"}}' \
  --max-results 10 \
  --region us-east-1
```

- **一覧**(ブラウズ。サマリのみで、`descriptors`は含まれない。代わりに、使われているdescriptorの種別名が`descriptorTypes`という配列で付く)

```bash
aws agent-registry list-discoverable-registry-records \
  --registry-id "<registryARN>" \
  --filters '[{"name": "recordType", "values": ["SKILL"]}]' \
  --max-results 50 \
  --region us-east-1
```

- **詳細取得**(descriptor込みで最大100件まで一括。一覧で絞ったあとに、必要なレコードの本体をまとめて取るのに使う)

```bash
aws agent-registry batch-get-discoverable-registry-record \
  --entries '[{"registryId": "<registryARN>", "recordIds": ["<recordId1>", "<recordId2>"]}]' \
  --region us-east-1
```

- 検索のポイント
  - `searchQuery`は最大256文字(開発者ガイドは1〜256文字、APIリファレンスは0〜256文字と書いていて食い違う)、`maxResults`は1〜20(デフォルト10)
  - `registryIds`は**現時点では1件だけ指定できる**(配列の長さは固定で1。複数のRegistryを横断した検索はできない)
  - 一覧の`maxResults`は1〜100、`filters`は最大10件。上の例の`--max-results 50`は例の値で、上限ではない
  - 「SKILLだけ」のような属性での絞り込みはクエリ文に入れず、**`filters`で指定する**(クエリ文に混ぜるとセマンティック検索が意図とずれる)
  - フィルタで使える演算子は`$eq` / `$ne` / `$in`、論理演算は`$and` / `$or`。対象フィールドは`name` / `recordType` / `recordVersion` / `customMetadata.{field}`
  - 検索対象は`name`(キーワードで最も重み大)、`description`、descriptorの中身。**descriptionを「何をするか/どんな課題を解決するか」で書くと見つかりやすい**
- 取得できるのはあくまでメタデータ(SKILL.mdと`repository`/`packages`情報)。`SKILL.md`以外のスキルのファイルはRegistryに保存されないので、実際にスキルを使うには、そこにある`repository`/`packages`をたどって取得することになる
- 公式のCLI例は`--search-query`と`--registry-ids`だけ。上の検索例の`--filters`は、APIリファレンスの`filters`パラメータ(JSON値。`{"recordType": {"$eq": "SKILL"}}`のような形式)をCLIに渡す形で組み立てたもので、CLIの例は公式にない。実行前に`aws agent-registry search-discoverable-registry-records help`で確認すること(一覧の`--filters`は公式の例どおり)
- 上の検索/一覧/一括取得のCLI例は**IAM認可のRegistry**用。**JWT認可のRegistry**は、Bearerトークンを付けてHTTPで直接叩く(公式の例)

```bash
# 検索
curl -X POST "https://agent-registry.<region>.api.aws/discoverable-records-search" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <accessToken>" \
  -d '{"registryIds": ["<registryARN>"], "searchQuery": "weather", "maxResults": 10}'

# 一覧
curl -X POST "https://agent-registry.<region>.api.aws/registries/<registryARN>/discoverable-records-list" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <accessToken>" \
  -d '{"filters": [{"name": "recordType", "values": ["MCP"]}], "maxResults": 50}'

# 一括取得
curl -X POST "https://agent-registry.<region>.api.aws/discoverable-records-batch" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <accessToken>" \
  -d '{"entries": [{"registryId": "<registryARN>", "recordIds": ["<recordId1>", "<recordId2>"]}]}'
```

## Skillsの補足: APIの仕様と一式配布の扱い (2026-10-04時点)
### Registryに入るもの/入らないもの
- **`SKILL.md`以外のスキルのファイルは、Registryに保存されない**。`scripts/`、`references/`、`assets/`をRegistryに登録して、取得側で受け取る構成は、仕様上できない
  - 公式の記述(Supported record types): "The markdown is only used as metadata for discovery purpose. Registry does not support storing other agent skill files."
- Registryはスキルの**配布先ではなく、発見(discovery)用のカタログ**

### `agentSkillsDefinition`の構造 (APIリファレンス)
| フィールド | 内容 |
| --- | --- |
| `additionalData.skillMd.data` | `SKILL.md`のテキスト |
| `additionalData.skillMd.dataSchemaVersion` | `skillMd`のスキーマバージョン |
| `additionalData.skillMd.source.fromUrl` | `url`と`credentialProviderConfigurations`を持つ項目。ただし下の注意を参照 |
| `data` | skill definition(JSON文字列)。1〜102,400文字 |
| `dataSchemaVersion` | 1〜255文字。skill definitionの現在の値は`0.1.0` |

- `agentSkillsDefinition`の中のフィールドは**すべて任意**(APIリファレンスで`Required: No`)。ただしトップレベルの`descriptors`は必須
- `skillMd`(`AgentSkillsMdDescriptor`)の`data`も、1〜102,400文字。`dataSchemaVersion`は1〜255文字。`data` / `dataSchemaVersion` / `source`はすべて任意(APIリファレンスで確認済み)
- skill definitionのスキーマ(`repository` / `websiteUrl` / `packages` / `_meta`)は、前述の「Skills」の節のとおり

> [!NOTE]
> - `skillMd.source.fromUrl`があるので、URLから`SKILL.md`を取得できそうに見えるが、開発者ガイドは「`skillMd`の`source`は保存されるが、同期の実行には使われない。SKILLレコードは自動同期できない」と書いている。`SKILL.md`をURLから自動取得する用途には使えない

### CreateRegistryRecordの制約と応答
- 必須は`name` / `recordType` / `descriptors`の3つ。`recordVersion` / `description` / `displayName`はAPI上は任意
- `name`: `[a-zA-Z0-9][a-zA-Z0-9_\-\.\/]*`、最大255文字
- `recordVersion`: `[a-zA-Z0-9.-]+`(**アンダースコアとスラッシュは使えない**。`name`とは文字種が違う)、最大255文字
- `description`: 1〜4,096文字
- `clientToken`(冪等性。33〜256文字)と`tags`(最大50個)も指定できる
- 作成は**非同期**。応答はHTTP 202で、返るのは`recordArn`と`status`(`CREATING`)だけ。**`recordId`は返らない**ので、ARNの末尾(12文字)から取り出す

```bash
RECORD_ARN=$(aws agent-registry-control create-registry-record ... --query recordArn --output text)
RECORD_ID=${RECORD_ARN##*/}
```

- `status`の値: `DRAFT` / `PENDING_APPROVAL` / `APPROVED` / `REJECTED` / `DEPRECATED` / `CREATING` / `UPDATING` / `CREATE_FAILED` / `UPDATE_FAILED`
- 主なエラー: `ConflictException`(409)、`ServiceQuotaExceededException`(402)、`ValidationException`(400。`fieldList`と`reason`が付く)

### UpdateRegistryRecordの落とし穴 (`optionalValue`)
- `UpdateRegistryRecord`は`CreateRegistryRecord`と**同じ形では渡せない**。PATCH形式で、フィールドを`optionalValue`で包む。省略したフィールドは変更されない(APIリファレンスで確認済み)
  - 包む: `description` / `displayName` / `customMetadata` / `descriptors`と、その中の各キー(`agentSkillsDefinition`、`additionalData`、`skillMd`、`data`など)
  - 包まない: `name` / `recordType` / `recordVersion` / `triggerSynchronization`
  - 値の消し方: `description` / `displayName`は空のラッパーで解除、`customMetadata`は`null`で全消し
- 応答はHTTP 202で、`status`は`UPDATING`
- Createと同じ形で渡すと、クライアント側の検証(`ParamValidationError`)で弾かれる(メモの検証による。こちらでは再確認できていない)

```python
# NG: Createと同じ形 -> ParamValidationError
client.update_registry_record(registryId=..., recordId=..., description="d",
                              descriptors={"agentSkillsDefinition": {...}})

# OK: optionalValueで包む (SKILL.mdだけを差し替える例)
client.update_registry_record(
    registryId=..., recordId=...,
    description={"optionalValue": "d"},
    descriptors={"optionalValue": {"agentSkillsDefinition": {"optionalValue": {
        "additionalData": {"optionalValue": {"skillMd": {"optionalValue": {
            "data": {"optionalValue": "<SKILL.mdの中身>"}}}}}}}}},
)
```

- CLIで同じことをする場合の形(APIリファレンスの構文から組み立てたもので、実行していない)

```bash
aws agent-registry-control update-registry-record \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --descriptors '{"optionalValue":{"agentSkillsDefinition":{"optionalValue":{"additionalData":{"optionalValue":{"skillMd":{"optionalValue":{"data":{"optionalValue":"<SKILL.mdの中身>"}}}}}}}}}' \
  --region us-east-1
```

### 取得側
- data planeのAPIは`SearchDiscoverableRegistryRecords` / `ListDiscoverableRegistryRecords` / `GetDiscoverableRegistryRecord` / `BatchGetDiscoverableRegistryRecord`の4つ
- `descriptors`を返すのは、**`Search`と`BatchGet`**。構造は作成時と同じ(`additionalData.skillMd.data`、`data`、`dataSchemaVersion`)。`List`は`descriptors`を返さず、サマリと`descriptorTypes`(種別名の配列)だけ
  - `GetDiscoverableRegistryRecord`の応答の中身は、APIリファレンスのページを取得できず確認していない
- `source`は`url`だけで、`credentialProviderConfigurations`は返らない。**ファイル群を返すフィールドはない**
- 返るのは**承認済み(`APPROVED`)のレコードだけ**(開発者ガイドの記述)
- **公式ドキュメント内に食い違いがある**: ライフサイクルのページは「承認済みレコードを編集して新しいDRAFTができても、承認済みリビジョンは検索され続ける」と書く。一方、検索/ブラウズのページは「最新リビジョンが`Approved`のレコードだけ」「`BatchGet`の`RESOURCE_NOT_FOUND`は、最新リビジョンが`Approved`でない場合にも返る」と書く。編集中の承認済みSkillが取得できるか(旧リビジョンが見えるか、`RESOURCE_NOT_FOUND`になるか)は、**実機で確認していない**
- IAMのアクションとARNは、後述の「IAM権限」を参照(`BatchGet`専用のアクションはなく、`GetDiscoverableRegistryRecord`で認可される)
- JWT認可のRegistryには、AWS CLI/SDKからdata planeを呼べない(SigV4のため)。公式の「Get started」にも明記がある

### 一式配布の代替案 (設計案。AWSの公式手順ではない)
- **Gitリポジトリを実体にする**: `repository.url`をRegistryに登録する。取得側は`SKILL.md`を取得したあと、`repository.url`から`git clone`などでディレクトリごと取得する
- **S3などに置く**: アーカイブを置き、`websiteUrl`や`_meta`に場所を入れる
- **`SKILL.md`だけで運用する**: 参照資料を本文に取り込み、`scripts/`は使わない

### 未確認として残るもの
- skill definition(`data`)を省略して、`SKILL.md`だけで作成した場合の実際の挙動(API上はどちらも任意だが、実機で確認していない)
- `packages`(npm、pypi)の使われ方: スキーマは配布情報のメタデータとして定義されているだけで、Registryがこれを解決して取得する仕組みの記載は見つからなかった
- Registryレコードの`name`と、`SKILL.md`のFrontmatterの`name`の関係: 一致を求める記載は見つからなかった。文字種は違う(レコードの`name`は大文字、`/`、`.`が使え、仕様上の`SKILL.md`の`name`は小文字とハイフンのみ)
- 旧ネームスペースの形(`descriptorType`が`AGENT_SKILLS`、`descriptors.agentSkills`の`skillMd.inlineContent`と`skillDefinition.inlineContent` + `schemaVersion`)は、移行ガイドの記述で確認した。ただし、移行ガイドのJSON例のコメントは`AGENT_SKILL`、本文は`AGENT_SKILLS`と、表記が揺れている
- [Classmethodの記事(Strands Agentsへの動的ロード)](https://dev.classmethod.jp/en/articles/aws-agent-registry-dynamic-skills-strands-agents/)は、本文を読めていない。検索結果の要約では、`register_skill.py`(登録)、`registry_skill_loader.py`(検索してSkillに変換)、`agent.py`の3つのスクリプトで構成されている。スクリプトの挙動は未確認

# Agents
- `recordType: AGENT`、基本のdescriptorは`a2aAgentCard`
  - A2Aプロトコルの**Agent Card**(`dataSchemaVersion: 0.3`)に対して検証される(JSON SchemaのAgentCard定義)
  - A2Aに従わないAgentでも`AGENT`タイプで登録可能。MCPで話すAgentなら`mcpServer`、それ以外(HTTPエンドポイントのみ等)は`custom`descriptorを使う
- Agent Cardの最小例

```json
{
  "name": "My Agent",
  "description": "Brief description of what this agent does",
  "version": "1.0.0",
  "protocolVersion": "0.3.0",
  "url": "https://api.example.com/a2a",
  "capabilities": {},
  "defaultInputModes": ["text/plain"],
  "defaultOutputModes": ["text/plain"],
  "skills": [
    {
      "id": "default-skill",
      "name": "Default Skill",
      "description": "Description of what this skill does",
      "tags": ["general"]
    }
  ]
}
```

## Agentsの登録/更新
### 登録(方法は2つ)
#### 1. URLから同期(Synchronize)する
- Agent Cardの公開URL(`https://<agent>/.well-known/agent-card.json`)を指定すると、Registryがカードを取得してdescriptorを埋める
- 作成すると`CREATING` → 同期完了で`DRAFT`(失敗時は`CREATE_FAILED`でStatus Reasonにエラーが出る)
- エンドポイントはHTTPSのみ
- アウトバウンド認証(Registry → Agentへのアクセス)は3種類
  - **None**: 公開されているAgent Card
  - **IAM**: AgentCore Runtime/Gateway上のAgent。`roleArn`と`service`を指定してSigV4で署名される。公式の同期ページは`service`に`agent-registry`を使うよう書いているが、疑わしい点がある(後述の「同期の制約・注意点」を参照)
  - **OAuth**: AgentCore Identityのcredential providerのARNを指定

```bash
# 公開Agent Cardから同期
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "travel-agent" \
  --display-name "Travel Agent" \
  --record-type AGENT \
  --descriptors '{
    "a2aAgentCard": {
      "source": {
        "fromUrl": {
          "url": "https://agent.example.com/.well-known/agent-card.json"
        }
      }
    }
  }' \
  --region us-east-1
```

```bash
# AgentCore Runtime上のAgentからIAMで同期
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "a2a_agent_record" \
  --display-name "A2A Agent Record" \
  --record-type AGENT \
  --descriptors "{
    \"a2aAgentCard\": {
      \"source\": {
        \"fromUrl\": {
          \"url\": \"$A2A_URL\",
          \"credentialProviderConfigurations\": [{
            \"credentialProviderType\": \"IAM\",
            \"credentialProvider\": {
              \"iamCredentialProvider\": {
                \"roleArn\": \"$IAM_INVOKER_ROLE\",
                \"service\": \"agent-registry\"
              }
            }
          }]
        }
      }
    }
  }"
```

- 同期に必要な追加のIAM権限(OAuthの場合)
  - `bedrock-agentcore:GetWorkloadAccessToken`、`bedrock-agentcore:GetResourceOauth2Token`(対象のcredential provider ARNに絞ること)
  - credential providerは同一アカウントのものである必要がある
- 同期に必要な追加のIAM権限(IAMの場合): 同期用ロールへの`iam:PassRole`(`iam:PassedToService`は`bedrock-agentcore.amazonaws.com`)

#### 2. 手動でAgent Cardを渡す
- コンソール: **Create record** → **Manual** → Typeに**Agent** → Descriptorに**A2A Agent Card**
- CLI(MCPの例と同じ形式で`a2aAgentCard`に置き換えたもの。Agent用の手動登録CLI例は公式ドキュメントに載っていなかったので、構造から組み立てている)

```bash
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "my-agent" \
  --display-name "My Agent" \
  --record-type AGENT \
  --record-version "1.0.0" \
  --descriptors '{
    "a2aAgentCard": {
      "data": "{\"name\": \"My Agent\", \"description\": \"Brief description of what this agent does\", \"version\": \"1.0.0\", \"protocolVersion\": \"0.3.0\", \"url\": \"https://api.example.com/a2a\", \"capabilities\": {}, \"defaultInputModes\": [\"text/plain\"], \"defaultOutputModes\": [\"text/plain\"], \"skills\": [{\"id\": \"default-skill\", \"name\": \"Default Skill\", \"description\": \"Description of what this skill does\", \"tags\": [\"general\"]}]}",
      "dataSchemaVersion": "0.3"
    }
  }' \
  --region us-east-1
```

### 承認申請

```bash
aws agent-registry-control submit-registry-record-for-approval \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --region us-east-1
```

### 更新
- メタデータの更新は`update-registry-record`(Skillsと同じ)
- **同期で更新**する場合は`--trigger-synchronization`を付ける
  - ステータスは`UPDATING` → 成功すると新しい`DRAFT`リビジョン(失敗時は`UPDATE_FAILED`)
  - ソース側にある値で、name / description / version / server定義 / tools定義(公式のMCPに関する記述)が**手動編集より優先して上書き**される(ソースが持っていないフィールドは変更されない)。Agentの場合もAgent Cardの内容がソースの値で更新されると思われるが、公式に明記があるのはMCPの説明だけ
  - APPROVED済みレコードの場合、旧リビジョンは検索可能なまま、新DRAFTが再承認されるまで切り替わらない
  - 自動では再同期されない。**ソースが変わったら手動で同期をトリガーする**必要がある

```bash
aws agent-registry-control update-registry-record \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --trigger-synchronization \
  --region us-east-1
```

## Agentsの取得
- Skillsと同じAPIで、`recordType`を`AGENT`にする

```bash
# 検索
aws agent-registry search-discoverable-registry-records \
  --search-query "出張の手配をしてくれるエージェント" \
  --registry-ids "<registryARN>" \
  --filters '{"recordType": {"$eq": "AGENT"}}' \
  --region us-east-1

# 一覧
aws agent-registry list-discoverable-registry-records \
  --registry-id "<registryARN>" \
  --filters '[{"name": "recordType", "values": ["AGENT"]}]' \
  --region us-east-1
```

- Agent Card(`url`や`skills`)は`descriptors.a2aAgentCard.data`に入っている。検索の応答には`descriptors`が含まれるので、**検索だけで取得できる**。一覧の応答は`descriptors`を含まないサマリ(`descriptorTypes`付き)なので、一覧で絞った場合は`batch-get-discoverable-registry-record`に`recordIds`を渡して取得する
- 呼び出す側のAgentは、取得したAgent Cardの`url`に対してA2Aで通信する。RegistryのAPI/MCPエンドポイントにあるのはsearch / list / batch-getだけで、Agentを呼び出すAPIはない。このことから、Registryは**見つけるためのカタログ**であって、呼び出しを仲介(プロキシ)するものではないと考えられる(公式に明記があるわけではなく、この点は推測)

# MCP
- `recordType: MCP`、descriptorは`mcpServer`
- 構成
  - **server**: MCPの[server.json](https://registry.modelcontextprotocol.io/)形式。`dataSchemaVersion`は`2025-12-11`、`2025-10-17`、`2025-10-11`、`2025-09-29`、`2025-09-16`、`2025-07-09`から選ぶ(`server.json`が無ければ最新で作成することが推奨されている)
  - **tools**: `descriptors.mcpServer.additionalData.tools`に入れる。MCPプロトコル仕様のtools定義。`dataSchemaVersion`は`2025-11-25`、`2025-06-18`、`2025-03-26`、`2024-11-05`
- toolの名前、説明、入力パラメータの説明は検索のセマンティックマッチに使われるため、**ツール定義を省略せずきちんと書く**と見つかりやすい
- serverの最小例

```json
{
  "name": "my-org/weather-server",
  "description": "Weather data and forecasts via OpenWeatherMap API",
  "version": "1.0.0"
}
```

- toolsの最小例

```json
{
  "tools": [
    {
      "name": "get_weather",
      "description": "Get the current weather for a given location",
      "inputSchema": {
        "type": "object",
        "properties": {
          "location": {
            "type": "string",
            "description": "City name, postal code, or latitude,longitude"
          }
        },
        "required": ["location"]
      }
    }
  ]
}
```

## MCPの登録/更新
### 登録(方法は2つ)
#### 1. 手動で定義を渡す(公式ドキュメントの例)

```bash
aws agent-registry-control create-registry-record \
  --registry-id <registryId> \
  --name "my-mcp-server" \
  --display-name "MyMCPServer" \
  --record-type MCP \
  --descriptors '{"mcpServer": {"data": "{\"name\": \"my/mcp-server\", \"description\": \"My MCP server\", \"version\": \"1.0.0\"}", "dataSchemaVersion": "2025-12-11"}}' \
  --record-version "1.0" \
  --region us-east-1
```

```python
import boto3
import json

client = boto3.client('agent-registry-control')

server_content = json.dumps({
    "name": "my/mcp-server",
    "description": "My MCP server",
    "version": "1.0.0"
})

response = client.create_registry_record(
    registryId='<registryId>',
    name='my-mcp-server',
    displayName='MyMCPServer',
    recordType='MCP',
    descriptors={
        'mcpServer': {
            'data': server_content,
            'dataSchemaVersion': '2025-12-11'
        }
    },
    recordVersion='1.0'
)
print(f"Record ARN: {response['recordArn']}")
print(f"Status: {response['status']}")  # CREATING
```

#### 2. 稼働中のMCPサーバから同期(Synchronize)する
- RegistryがMCPサーバに接続して、**server定義とtools定義を自動で取得**してくれる。手でtools JSONを書かなくて済むので、基本はこちらがおすすめ
- アウトバウンド認証は3パターン

```bash
# 認証なしの公開MCPサーバ
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "aws-knowledge-server" \
  --display-name "AWS Knowledge Server" \
  --record-type MCP \
  --descriptors '{
    "mcpServer": {
      "source": {
        "fromUrl": {
          "url": "https://knowledge-mcp.global.api.aws"
        }
      }
    }
  }' \
  --region us-east-1
```

```bash
# OAuth保護のMCPサーバ (AgentCore Identityのcredential provider)
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "oauth-mcp-server" \
  --display-name "OAuth MCP Server" \
  --record-type MCP \
  --descriptors '{
    "mcpServer": {
      "source": {
        "fromUrl": {
          "url": "$MCP_OAUTH_URL",
          "credentialProviderConfigurations": [
            {
              "credentialProviderType": "OAUTH",
              "credentialProvider": {
                "oauthCredentialProvider": {
                  "providerArn": "$OAUTH_PROVIDER_ARN",
                  "grantType": "CLIENT_CREDENTIALS"
                }
              }
            }
          ]
        }
      }
    }
  }' \
  --region us-east-1
```

```bash
# IAM保護のMCPサーバ (AgentCore Runtime / Gateway など)
aws agent-registry-control create-registry-record \
  --registry-id $REGISTRY_ID \
  --name "gateway-mcp-server" \
  --display-name "Gateway MCP Server" \
  --record-type MCP \
  --descriptors '{
    "mcpServer": {
      "source": {
        "fromUrl": {
          "url": "$MCP_IAM_URL",
          "credentialProviderConfigurations": [
            {
              "credentialProviderType": "IAM",
              "credentialProvider": {
                "iamCredentialProvider": {
                  "roleArn": "$IAM_ROLE_ARN",
                  "service": "$SIGNING_SERVICE",
                  "region": "$SIGNING_REGION"
                }
              }
            }
          ]
        }
      }
    }
  }' \
  --region us-east-1
```

- IAM保護の場合の`service`(SigV4署名先)は、公式ドキュメントでは次のとおり
  - AgentCore Runtime / Gateway上: `agent-registry`(**要検証**。後述の「同期の制約・注意点」を参照)
  - Amazon API Gateway上: `execute-api`
  - AWS Lambda上: `lambda`
  - `region`は省略可(省略時はRegistryと同じリージョンで署名)
  - 同期用ロールには対象サービスの呼び出し権限が必要(例: Runtimeなら`bedrock-agentcore:InvokeAgentRuntime`、Gatewayなら`bedrock-agentcore:InvokeGateway`)
- Gatewayで作ったMCP(Lambda/APIをMCP化したもの)もこの方法でRegistryに登録できる

### 承認申請

```bash
aws agent-registry-control submit-registry-record-for-approval \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --region us-east-1
```

### 更新
- 手動更新: `update-registry-record`
- 再同期: `--trigger-synchronization`(Agentと同じ。`UPDATING` → 新DRAFTリビジョン)

```bash
aws agent-registry-control update-registry-record \
  --registry-id $REGISTRY_ID \
  --record-id $RECORD_ID \
  --trigger-synchronization \
  --region us-east-1
```

- 同期失敗時は`CREATE_FAILED` / `UPDATE_FAILED`になり、`statusReason`に原因が出る(コンソールのレコード詳細にも表示される)

## MCPの取得
- Skills/Agentsと同じく`recordType`を`MCP`にして検索/一覧/一括取得する。`Search`と`BatchGet`の応答には`descriptors`(server定義と`additionalData.tools`)が含まれ、`List`はサマリのみ

```bash
# 検索
aws agent-registry search-discoverable-registry-records \
  --search-query "weather forecast" \
  --registry-ids "<registryARN>" \
  --filters '{"recordType": {"$eq": "MCP"}}' \
  --region us-east-1

# 一覧
aws agent-registry list-discoverable-registry-records \
  --registry-id "<registryARN>" \
  --filters '[{"name": "recordType", "values": ["MCP"]}]' \
  --region us-east-1

# 一括取得 (server/toolsを含む完全なdescriptor)
aws agent-registry batch-get-discoverable-registry-record \
  --entries '[{"registryId": "<registryARN>", "recordIds": ["<recordId1>"]}]' \
  --region us-east-1
```

```python
import boto3

client = boto3.client('agent-registry')

response = client.search_discoverable_registry_records(
    registryIds=['<registryARN>'],
    searchQuery='weather',
    maxResults=10
)
for record in response['registryRecords']:
    print(f"{record['displayName']} ({record['name']}) - {record['recordType']} - {record['status']}")
```

### RegistryのMCPエンドポイントから取得する
- Registryは自分自身がMCPサーバとして公開されている。MCPクライアント(Claude Code、Kiroなど)やAgentから**ツールとして呼ぶ**ことで、Agentが自律的に必要なMCP/Agent/Skillを探せる
- エンドポイント(MCP仕様`2025-11-25`準拠)

```
https://agent-registry.<region>.api.aws/registry/<registryId>/mcp
```

- 公開されるMCPツール
  - `search_discoverable_registry_records`: 自然言語で検索(`searchQuery`必須、`maxResults`、`filter`)。HTTP APIの応答には`descriptors`が含まれるが、MCPツールの説明は「Returns metadata for matching records」としか書いておらず、MCP経由でも含まれるかは確認していない
  - `list_discoverable_registry_records`: 承認済みレコードのページネーション一覧
  - `batch_get_discoverable_registry_record`: `recordIds`(最大100件)で完全なdescriptorを取得
- **JWT認可のRegistry**の場合

```bash
curl -s -X POST "https://agent-registry.<region>.api.aws/registry/<registryId>/mcp" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search_discoverable_registry_records","arguments":{"searchQuery":"weather"}}}'
```

  - MCPクライアントの設定方法は3種類: ①Bearerトークンをヘッダに設定、②事前登録クライアント(`allowedClients`にclient idを許可)、③動的クライアント登録(DCR。`allowedAudience`を設定)
- **IAM認可のRegistry**の場合は、[mcp-proxy-for-aws](https://github.com/aws/mcp-proxy-for-aws)経由でSigV4署名してつなぐ(`.mcp.json`の例)

```json
{
  "mcpServers": {
    "iam-based-registry": {
      "type": "stdio",
      "command": "uvx",
      "args": [
        "mcp-proxy-for-aws@latest",
        "https://agent-registry.<region>.api.aws/registry/<registryId>/mcp",
        "--service",
        "agent-registry",
        "--region",
        "<region>",
        "--profile",
        "my-profile"
      ]
    }
  }
}
```

  - 必要なIAM権限は`agent-registry:InvokeRegistryMcp`(MCPの初期化とツール一覧の取得まで)。検索ツールを実行するには、加えて`agent-registry:SearchDiscoverableRegistryRecords`も必要(公式のIAMページでは、MCPエンドポイントの呼び出しにはこの両方が必要とされている)
  - list / batch-getのツールの実行に必要な追加アクションは、公式に明記がない(`ListDiscoverableRegistryRecords` / `GetDiscoverableRegistryRecord`が必要になると思われる)

# 参考
- [AWS Agent Registry (公式ドキュメント)](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry.html)
- [Supported record types and descriptors](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-supported-record-types.html)
- [Create and manage records](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-create-manage-records.html)
- [Synchronize records from external sources](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-sync-records.html)
- [Search for registry records](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-search-records.html)
- [Using the Registry MCP endpoint](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry-mcp-endpoint.html)
- [AWS Agent Registry is now available in Preview (What's New, 2026/04)](https://aws.amazon.com/about-aws/whats-new/2026/04/aws-agent-registry-in-agentcore-preview/)
- [The future of managing agents at scale: AWS Agent Registry now in preview (AWS Blog)](https://aws.amazon.com/blogs/machine-learning/the-future-of-managing-agents-at-scale-aws-agent-registry-now-in-preview/)

# 同期の制約・注意点
- 同期(Synchronize)の対象は**MCPサーバとA2A Agentのみ**。Skills / Customは非対応
- エンドポイントは**HTTPS**必須
- 接続先は**パブリックIPのサーバのみ**(「The provided URL resolves to a non-public IP address」エラーになる)。社内のプライベートなMCPサーバ/AgentはURLからは同期できず、手動登録になる
- MCPサーバのtoolsのページネーション取得は**最大30秒**でタイムアウトする(「MCP server tools/list pagination timed out」)
- 同期でも手動でも、スキーマ検証に通らないとエラーになる。例えばA2Aの`dataSchemaVersion`は`0.3`で、`0.3.0`は不可
- 同期の代表的なエラー(`CREATE_FAILED` / `UPDATE_FAILED`の`statusReason`)
  - 権限: 認証情報の期限切れ、`GetWorkloadAccessToken` / `GetResourceOauth2Token`のエラー、credential provider ARNの形式不正、IAMロールをAssumeできない(`iam:PassRole`不足など)
  - 接続: HTTP 401/403(credential providerや権限が不正)、URL不正、接続失敗
  - 検証: JSON不正、レスポンスサイズ超過、MCPの`result`がない
  - その他: 「Unknown error」はサーバ側の問題なので、時間を置いて再試行するかAWSサポートに問い合わせる
- **IAM同期の`service`について(要検証)**: 公式の同期ページは、AgentCore Runtime/Gateway上のMCP/Agentでは`service`に`agent-registry`を使うよう書いている。一方、同じ公式ドキュメントの自動検出ページが示すRuntimeのエンドポイントURLは`bedrock-agentcore.<region>.amazonaws.com`で、旧ネームスペース向けの例でもこの`service`は`bedrock-agentcore`になっている。署名先は実際のエンドポイントのサービス名と一致する必要があると考えられるので、`agent-registry`で署名エラーになった場合は`bedrock-agentcore`を試す。このノートでは実機で確認していない
- **IAM同期のロール**: 信頼ポリシーのPrincipalは`agent-registry.amazonaws.com`(新ネームスペース)。旧ネームスペースの`bedrock-agentcore.amazonaws.com`のままだと同期が失敗する(移行時の注意点)

# IAM権限
- 管理用のマネージドポリシーは`AgentRegistryFullAccess`(旧ネームスペースの`BedrockAgentCoreFullAccess`には`agent-registry:*`は追加されない)
- ペルソナごとに必要なアクション(`arn:aws:agent-registry:{region}:{account}:registry/{registryId}`、レコードは末尾に`/record/{recordId}`)

| ペルソナ | 主なアクション |
| --- | --- |
| Administrator | `Create/Get/Update/DeleteRegistry`、`ListRegistries`、`Create/Get/Update/DeleteRegistryRecord`、`ListRegistryRecords`、`SubmitRegistryRecordForApproval`、`UpdateRegistryRecordStatus` |
| Curator | `ListRegistries`、`GetRegistry`、`ListRegistryRecords`、`GetRegistryRecord`、`UpdateRegistryRecordStatus` |
| Publisher | Curatorの読み取り系 + `CreateRegistryRecord`、`Update/DeleteRegistryRecord`、`SubmitRegistryRecordForApproval`(同期を使うなら下記の追加権限) |
| Consumer | `SearchDiscoverableRegistryRecords`、`ListDiscoverableRegistryRecords`、`GetDiscoverableRegistryRecord`、`InvokeRegistryMcp` |

- `BatchGetDiscoverableRegistryRecord`には**専用のIAMアクションがない**。各レコードが`GetDiscoverableRegistryRecord`で認可される
- Registryの作成時は、サービスが`bedrock-agentcore:CreateWorkloadIdentity` / `GetWorkloadIdentity` / `DeleteWorkloadIdentity`と`iam:CreateServiceLinkedRole`(`AWSServiceRoleForAgentRegistry`)を使うため、作成者にこれらの権限が必要
- 同期を使うPublisherの追加権限: `bedrock-agentcore:GetWorkloadAccessToken`、OAuthなら`bedrock-agentcore:GetResourceOauth2Token`(credential providerのARNに絞る)、IAMなら同期ロールへの`iam:PassRole`
- WorkloadIdentityやOAuth credential providerはAgentCore Identityのリソースなので、`bedrock-agentcore`のまま変わらない(移行時に全部`agent-registry`へ置換してはいけない)

# EventBridge通知
- イベントはデフォルトのイベントバスに、リソースのアカウントで送られる。ソースは`aws.agent-registry`
- レコード: `Registry Record State changed to Draft` / `Pending Approval` / `Approved` / `Rejected` / `Deprecated`(`detail`に`registryRecordId`と`registryId`)
- Registry: `Registry Creating` / `Ready` / `Create Failed` / `Updating` / `Update Failed` / `Deleting` / `Delete Failed`
- 例: 承認申請のイベント

```json
{
  "version":"0",
  "detail-type":"Registry Record State changed to Pending Approval",
  "source":"aws.agent-registry",
  "account":"123456789012",
  "region":"us-west-2",
  "resources":["arn:aws:agent-registry:us-west-2:123456789012:registry/REG_ID/record/REC_ID"],
  "detail":{"registryRecordId":"REC_ID","registryId":"REG_ID"}
}
```

- 既存の審査パイプライン(セキュリティ/コンプライアンスチェック)との連携: 承認申請のイベントでパイプラインを起動し、完了したら`update-registry-record-status`で承認/却下する

# 制限・クォータ・リージョン
- **対応リージョン**(公式のリージョン表): N. Virginia(us-east-1)、Oregon(us-west-2)、Ireland(eu-west-1)、Sydney(ap-southeast-2)、Tokyo(ap-northeast-1)の5つ
  - **東京リージョンで使える**
  - 自動検出も、1リージョンにつき1つ有効なRegistryを作る形でリージョン単位
- **クォータ**(いずれもリージョン単位。Service Quotasで引き上げ可能)
  - Registry数: 1アカウント・1リージョンあたり**5個**
  - APIのレート: `CreateRegistry` / `GetRegistry` / `UpdateRegistry` / `DeleteRegistry` / `ListRegistries` / `CreateRegistryRecord` / `UpdateRegistryRecord`は5 TPS
  - `Get/Delete/ListRegistryRecords`、`SubmitRegistryRecordForApproval`、`UpdateRegistryRecordStatus`、`Search/List/BatchGetDiscoverableRegistryRecord`は10 TPS
  - 旧ネームスペースで引き上げ申請をしていた場合は、`agent-registry`のサービスコードで再申請が必要

# 発展トピック
## AWS Organizationsによる自動検出(Auto-detection)
- AWS Organizationsのメンバーアカウント内の**AgentCore RuntimeとGateway**を、自動でRegistryにレコードとして登録できる(メンバーアカウント側の設定は不要)
- 流れ: ①管理アカウントで信頼されたアクセスを有効化 + サービスリンクロール作成 → ②委任管理者を登録 → ③委任管理者で`--auto-detection-configuration '{"scope":"ORGANIZATION","enabled":true}'`を指定したRegistryを作成
- 組織内で、**1リージョンにつき有効な組織スコープのRegistryは1つ**だけ
- 検出されたレコードは、名前が`aws-autodetected-<accountId>-<region>-<resourceId>`の形式で、`provenance`(`DETECTED_FROM`)でソースのリソースと紐づく。初回の検出には最大20分かかる
- RuntimeはrecordTypeが`AGENT`(MCPに絞り込み可)、Gatewayは`GATEWAY`で固定
- 自動検出が有効な間は、検出されたレコードを削除できない(自動検出を`false`にしてから削除する)。ソースのリソースが削除されたり、アカウントが組織を抜けるとレコードも消える
- `name` / `description` / `recordVersion`は自分で編集でき、再検出で上書きされない。承認済みのレコードが更新された場合は、新しいDRAFTリビジョンになる
- 検出されたレコードを`DEPRECATED`にすると、そのレコードはソースの変更に追従しなくなる

## RAMによるクロスアカウント共有
- Registryを他のアカウントとAWS RAMで共有できる(Registryの所有者だけが共有を開始でき、`UpdateRegistry` / `DeleteRegistry`は委譲できない)
- マネージド権限は4種類: `AWSRAMDefaultPermissionAgentRegistryReadOnly`(既定)、`...ForConsumer`(検索 + MCP呼び出し。通常はこれを推奨)、`...ForPublisher`(+ レコードの作成/更新/削除/承認申請)、`...ForAdmin`(+ 承認/却下/非推奨)
- 条件キー`agent-registry:RecordCreatorAccount` / `RecordSourceAccount`で、レコード単位のアクセス制御ができる(CloudFormationでレコードを管理するなら、`ListTagsForResource`を含むPublisher以上の権限が必要)

## その他
- **PrivateLink**: インターフェースエンドポイントは`com.amazonaws.<region>.agent-registry-control`(control plane)と`com.amazonaws.<region>.agent-registry`(data plane / MCPエンドポイント)の2つ。JWT認可のRegistryには、エンドポイントポリシーで`Principal: "*"`が必要
- **KMS**: カスタマーマネージドキーによる暗号化は、Registryの作成時にのみ指定できる
- **CloudTrail**: control planeのAPI呼び出しは管理イベントとして記録される。イベントソースは`agent-registry.amazonaws.com`

# 旧ネームスペース(bedrock-agentcore)からの移行メモ
- 移行ツールで新しいRegistry/レコードを作り直す(同じアカウント・同じリージョン)。自動では移行されない
- 主な変更点

| 項目 | 旧 | 新 |
| --- | --- | --- |
| CLI | `aws bedrock-agentcore` / `bedrock-agentcore-control` | `aws agent-registry` / `agent-registry-control` |
| エンドポイント | `bedrock-agentcore(-control).{region}.amazonaws.com` | `agent-registry(-control).{region}.api.aws` |
| IAMアクション / ARN | `bedrock-agentcore:*` | `agent-registry:*` |
| 検索API | `SearchRegistryRecords` | `SearchDiscoverableRegistryRecords`(+ List/BatchGetが新設) |
| EventBridge | `aws.bedrock-agentcore` | `aws.agent-registry`(Registryの`Registry State transitions from Creating to Ready`は`Registry Ready`に) |
| CloudWatch | `AWS/BedrockAgentCore` | `AWS/AgentRegistry` |
| `name` | レコードの表示名 | `displayName`に名称変更。新しい`name`が重複排除用のキー |
| descriptor | `descriptorType`(`MCP` / `A2A` / `AGENT_SKILLS` / `CUSTOM`) + `inlineContent` / `schemaVersion` | `recordType` + フラットな`mcpServer` / `a2aAgentCard` / `agentSkillsDefinition` / `custom` + `data` / `dataSchemaVersion` |
| 同期の設定 | `synchronizationType` / `synchronizationConfiguration` | descriptor内の`source.fromUrl` |
| 認可の設定 | `authorizerType` / `authorizerConfiguration` | `discoveryConfiguration`の中に移動 |
| Listのフィルタ | `--status`などの個別パラメータ(GET) | `filters`パラメータ(POST) |

- 同期用のIAMロールの信頼ポリシーのPrincipalを`agent-registry.amazonaws.com`に更新しないと、移行後のレコードの同期が失敗する(失敗したレコードは`CREATE_FAILED`になり、削除して再ロードが必要)

# 補足
- **提供状況**: 2026年4月にプレビューとして公開された(What's New記事より)。その後、`agent-registry`ネームスペースへ移行する形で提供されている
- **サードパーティ連携**: OSSの[MCP Gateway Registry](https://github.com/agentic-community/mcp-gateway-registry/blob/main/docs/aws-agent-registry-federation.md)は、AgentCoreのRegistryとフェデレーションできる。定期的にレコードを同期し、AgentCoreのdescriptor(MCP / A2A / CUSTOM / AGENT_SKILLS)をGateway側のネイティブなアセットに変換する
