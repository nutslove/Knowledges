# Amazon Bedrock Guardrails について

## 1. 概要

- 生成AIアプリ向けの設定可能な安全対策。ユーザー入力とモデル出力の両方を検査、モデルが異なっても同じ基準で適用可能。モデルの reasoning content は対象外。
- 使い方は大きく2系統。

**アプリのコードから呼ぶ(SDK / API)**

| # | 方法 | 内容 | 向く場面 | 検出したとき |
|---|---|---|---|---|
| 1 | モデル呼び出しに同梱 | `Converse` / `InvokeModel` に guardrail ID とバージョンを渡すと、入力と出力を自動で検査 → [具体例(3.1)](#31-モデル呼び出しに同梱converse--invokemodel) | Bedrock のモデルを使うアプリ | guardrail に設定した強度・閾値・アクションに従って、自動でブロックまたはマスク。Converse は `stopReason: guardrail_intervened` になり、定型メッセージが返る |
| 2 | `ApplyGuardrail` を単体で呼ぶ | 検査したいテキストを渡して判定結果だけ受け取る。guardrail リソースを使う。モデルは呼ばない → [具体例(3.2)](#32-applyguardrail-api) | Bedrock 以外のモデル(OpenAI、自前モデルなど)、RAG の検索前など好きなタイミング | guardrail の設定に従った結果(`action: GUARDRAIL_INTERVENED`、`outputs` に定型メッセージまたはマスク済みテキスト)を返す。実際に止めるかはアプリが判断 |
| 3 | `InvokeGuardrailChecks` を単体で呼ぶ | guardrail を作らず、呼ぶたびに `checks`(どの検査をするかの指定。`contentFilter` / `promptAttack` / `sensitiveInformation` から選ぶ)をインラインで指定する。検出のみで、判定はアプリ側 → [具体例(3.4)](#34-invokeguardrailchecks-api2026-06-16-発表) | エージェントのループ中の任意の地点 | スコアだけ返す。基準(閾値)も動作もアプリが決める |

**コード変更なし(AWS 側の設定)**

| # | 方法 | 内容 | 向く場面 | 検出したとき |
|---|---|---|---|---|
| 4 | 強制適用 | アプリのコードに guardrail を書かなくても、アカウント/組織の全 Bedrock 呼び出しに適用 → [詳細(3.3)](#33-強制適用enforcement) | 全社で統制したい場合 | guardrail に設定した強度・閾値・アクションに従って、モデル呼び出し側で自動でブロックまたはマスク |

- LiteLLM は内部で 2 または 3 を呼ぶ → [5 章](#5-litellm-との組み合わせ)。AgentCore Gateway に Policy Engine を attach すると、Gateway 経由のリクエストに対して Policy が 3 を呼ぶ(Gateway を通らない呼び出しは対象外) → [4.1](#41-agentcore-policy-の-guardrail-条件公式機能)
- 3.3 には具体的なコード例はなく、説明のみ

**運用**
- `DRAFT` で試し、数値バージョンを作って本番で固定する

**料金**
- 有効にしたフィルタごとの従量課金。1 text unit は最大1,000文字。Standard と Classic は同じ料金

| フィルタ | 料金 |
|---|---|
| Content filters(テキスト) | $0.15 / 1,000 units |
| Content filters(画像) | $0.00075 / 枚 |
| Prompt Attack | 独立した行なし(Content filters の一部として扱われると読めるが、明記はない) |
| Denied topics | $0.15 / 1,000 units |
| Sensitive information(PII) | $0.10 / 1,000 units |
| Sensitive information(regex) | 無料 |
| Word filters | 無料 |
| Contextual grounding | $0.10 / 1,000 units |
| Automated Reasoning | $0.17 / 1,000 units(ポリシー単位) |

`InvokeGuardrailChecks` を使う場合の専用料金(1,000 units あたり):

| フィルタ(`checks` のキー名) | 料金 |
|---|---|
| Content filters(`contentFilter`) | $0.07 |
| Prompt Attack(`promptAttack`) | $0.08 |
| Sensitive information(PII)(`sensitiveInformation`) | $0.10 |

## 2. 防御できる種類

AWS ではこれらの個々の検査機能を **safeguard**(= フィルタ)と呼ぶ。

**`InvokeGuardrailChecks` の `checks` との対応**(3.4)。検出する対象は同じだが、`checks` は guardrail を作らず毎回指定し、スコアを返すだけ(ブロックやマスクはしない)。判定の実装まで同一かは未確認。

| フィルタ(この章) | `checks` のキー名 |
|---|---|
| Content filters | `contentFilter` |
| Prompt Attack | `promptAttack` |
| Sensitive information(PII) | `sensitiveInformation` |

`checks` で指定できるのはこの 3 つだけ。Denied topics、Word filters、Contextual grounding、Automated Reasoning は含まれない。
「safeguard tier」は Standard/Classic の区別(対象は Content filters、Prompt attacks、Denied topics のみ)、料金の「safeguard ごと」は有効にした検査機能の数だけ課金される、という意味。

| 種類 | 内容 |
|---|---|
| **Content filters** | Hate / Insults / Sexual / Violence / Misconduct の5カテゴリ。強度は NONE / LOW / MEDIUM / HIGH。画像対応。 |
| **Prompt Attack** | Jailbreak / Prompt Injection / **Prompt Leakage(Standard のみ)**。 |
| **Denied topics** | 拒否トピックを自然言語で定義(最大30個)。単語やエンティティ検出には不向き。 |
| **Word filters** | 完全一致の単語・フレーズ。競合名や profanity 用。 |
| **Sensitive information** | PII を ML で検出、**ブロックまたはマスク**(`{EMAIL}` のように置換)。独自 regex 対応。 |
| **Contextual grounding** | RAG の幻覚検出。**Grounding**(ソースに基づくか)と **Relevance**(質問に答えているか)のスコアを閾値判定。 |
| **Automated Reasoning** | 形式論理でポリシーとの整合性を数学的に検証。detect モードのみ、ブロックはしない。 |

### 2.1 各フィルタの言語対応と注意点

- **Tier の違い**(Content filters、Prompt attacks、Denied topics のみ):

| 項目 | Standard | Classic |
|---|---|---|
| 性能 | より堅牢 | 従来の性能 |
| Denied topics 定義長 | 最大 1,000 文字 | 最大 200 文字 |
| 言語対応 | 多言語対応 | en/fr/es のみ |
| Cross-Region inference | 使用 | 非対応 |
| Prompt leakage 検出 | 対応(Standard only) | 非対応 |
| コード対応 | コメント、変数名、関数名、文字列内の害悪コンテンツ | なし |

- **日本語対応**:
  - Content filters、Prompt Attack(Jailbreak/Injection/Leakage)、Denied topics: Standard で "Optimized and supported"
  - PII: 日本語 "Optimized and supported"(PII に tier の概念はなく、tier 非依存)
  - Word filter、Contextual grounding: **日本語未対応**(en/fr/es のみ)
  - AgentCore Policy 経由(ContentFilter / PromptAttack / SensitiveInformation)の言語対応は、公式に記載なし
- **Prompt Attack は入力タグ必須**(`InvokeModel` / `InvokeModelWithResponseStream`): ユーザー入力を `<amazon-bedrock-guardrails-guardContent_xyz>` 形式のタグで囲まないと、Prompt Attack は検査されない。他のフィルタはタグがなければ全文が検査される
  - Converse API ではタグではなく `guardContent` ブロックを使う(XML タグは Converse では使えない)
  - 開発者の system prompt とユーザー入力は見た目が似ていて区別できないため、タグでユーザー入力だけを指定する
  - 具体例は 3.1 の InvokeModel
- **tool use は全フィルタで評価対象外**(`Converse` / `InvokeModel` の `guardrailConfig` 経由の場合):
  - tool result(`toolResult`)、tool 定義(`toolSpec` の description / inputSchema)、モデルが生成した tool call 引数(`toolUse.input`)の 3 つとも、Prompt Attack・PII・Content filter・Denied topics・Word filter のいずれでも検査されない
  - エージェント用途での重要な穴。ツール入出力を検査したい場合は、`ApplyGuardrail` / `InvokeGuardrailChecks` / AgentCore Policy などで、検査したいテキストを明示的に渡す(これらの経路に「tool 除外」の記述はないが、渡したテキストが評価される旨の明文も未確認)
- **Automated Reasoning の制約**: 
  - 英語(US)のみ
  - ストリーミング非対応
  - 対応リージョン限定(US 3、EU 3、シドニー)
  - **プロンプトインジェクション対策にはならない**
  - ソース文書5MB、50,000文字以内
- **ストリーミング**: 既定 sync モードはチャンクをバッファ検査するため遅延あり。async モードは即時に返すが、検査完了までは不適切な内容が出る可能性。**async では PII マスク不可**。
- **Contextual grounding の上限**: ソース100,000文字、クエリ1,000文字、応答5,000文字。会話型 QA は非対応。リージョン別の公式一覧は見つからず、東京で使えるかは明文なし(Guardrails 自体は東京対応)。
- **ログの注意**: PII マスクはモデルへの入出力だけ。Model invocation logs の input には元の PII が残る、trace の `match` にも元の値が入る。
- **ドキュメント間の食い違い**: Denied topics の定義長は資料によって異なる。
  - Tier ページ: Standard 1,000文字 / Classic 200文字
  - Denied topics ページ、Bedrock API Reference(`GuardrailTopicConfig`): 200文字
  - CloudFormation(`TopicConfig`): 1,000文字
  - Standard で 201 文字以上が実際に通るかは未検証。実装時に確認推奨

## 3. 基本的な使い方

### 3.1 モデル呼び出しに同梱(Converse / InvokeModel)

#### Converse API

```python
import boto3

bedrock_runtime = boto3.client("bedrock-runtime", region_name="us-east-1")
model_id = "<モデルID>"
gid = "<guardrail ID または ARN>"

resp = bedrock_runtime.converse(
    modelId=model_id,
    messages=[{"role": "user", "content": [
        {"text": "背景情報(検査しない)"},
        {"guardContent": {"text": {"text": "ユーザーの質問"}}},
    ]}],
    guardrailConfig={
        "guardrailIdentifier": gid,
        "guardrailVersion": "1",
        "trace": "enabled",
    },
)
# resp["stopReason"] == "guardrail_intervened" なら介入あり
```

- `guardContent` が1つでもあると、**それ以外は検査されない**
- `guardContent` 不使用の場合、メッセージ全体検査、system prompt は検査されない

#### InvokeModel / InvokeModelWithResponseStream

ユーザー入力を入力タグで囲み、`amazon-bedrock-guardrailConfig` に `tagSuffix` を入れる。タグの仕様は公式(guardrails-tagging)で確認済み。

```python
# 引数名(guardrailIdentifier / guardrailVersion / trace)は boto3 1.43.114 の入力定義で確認済み。
# 実際に AWS を呼んだ検証はしていない。body の形式はモデルごとに異なるため省略している
import json
import secrets

import boto3

bedrock_runtime = boto3.client("bedrock-runtime", region_name="us-east-1")
model_id = "<モデルID>"
gid = "<guardrail ID または ARN>"

suffix = secrets.token_hex(8)   # 英数字 1〜20 文字。リクエストごとにランダムにする
tag = f"amazon-bedrock-guardrails-guardContent_{suffix}"

user_input = "ユーザーの質問"
prompt = f"""あなたは銀行アシスタントです。(system prompt。検査しない)
次の質問に答えてください:
<{tag}>{user_input}</{tag}>"""

resp = bedrock_runtime.invoke_model(
    modelId=model_id,
    guardrailIdentifier=gid,
    guardrailVersion="1",
    trace="ENABLED",
    body=json.dumps({
        # モデル固有の body。プロンプト本文に上の prompt を入れる
        "amazon-bedrock-guardrailConfig": {"tagSuffix": suffix},
    }),
)
```

- タグの接頭辞 `amazon-bedrock-guardrails-guardContent` は予約語。`_` の後ろが `tagSuffix` と一致する必要がある
- **`tagSuffix` はリクエストごとにランダムにする**(推奨)。固定だと、悪意のあるユーザーがタグを閉じて、その後ろに攻撃文を付けられる
- タグは複数回使えるが、**入れ子は不可**
- **Prompt Attack を使うなら、タグは必須**。タグがないと Prompt Attack だけ検査されない
- ストリーミング(`InvokeModelWithResponseStream`)でも同じタグを使う

### 3.2 ApplyGuardrail API

```python
import boto3

bedrock_runtime = boto3.client("bedrock-runtime", region_name="us-east-1")
gid = "<guardrail ID または ARN>"

r = bedrock_runtime.apply_guardrail(
    guardrailIdentifier=gid, guardrailVersion="1",
    source="INPUT",   # LLM応答検査なら "OUTPUT"
    content=[{"text": {"text": "..."}}],
)
# r["action"]: "GUARDRAIL_INTERVENED" | "NONE"
# マスク→outputs に加工後テキスト、ブロック→定型メッセージ1件、介入なし→空配列
```

- `outputScope=FULL` で検出されなかった項目も返す
- Grounding使用時は `qualifiers` に `grounding_source` / `query` / `guard_content` を指定

### 3.3 強制適用(Enforcement)

- **アカウント単位**: `PutEnforcedGuardrailConfiguration` で設定。リージョンごとに必要。
- **組織単位**: AWS Organizations の Bedrock policy(`BEDROCK_POLICY`)で guardrail ARN を指定。
  - 強制対象: Bedrock のモデル呼び出し(InvokeModel / Converse 系)
  - **複数の Bedrock ポリシーが重なった場合**: Organizations の宣言型ポリシー継承ルールで 1 つの effective policy に解決される（単純な和集合ではない）。`DescribeEffectivePolicy` で確認
  - **組織・アカウント・リクエスト の 3 レベルが同時にある場合**: すべて適用、最も厳しい設定が優先
  - **非対応ポリシー**: Automated Reasoning を含むと実行時エラー
  - モデルの `included_models` / `excluded_models` で適用範囲を制限可能
  - `selective_content_guarding`(system / messages)でタグ対象外のコンテンツをスキップ可能（既定: comprehensive）
- 組織レベル強制(クロスアカウント)は 2026-04-03 に GA。AWS 商用リージョンと GovCloud のうち Guardrails 対応リージョンが対象
- 強制中の guardrail は `DeleteGuardrail` で削除不可

### 3.4 InvokeGuardrailChecks API(2026-06-16 発表)

エージェントのループ中の任意の地点で、個別の safeguard を単発で呼ぶための API。

- **resourceless**: guardrail の作成・ID・バージョン不要。リクエストごとに `checks` をインライン指定
- **detect-only**: ブロック・マスク・書き換えはしない。スコアだけ返し、block / retry / log の判断はアプリ側
- 1 回の呼び出し内の checks は並列実行
- ApplyGuardrail との違い:

| | InvokeGuardrailChecks | ApplyGuardrail |
|---|---|---|
| 用途 | ワークフローの特定地点での個別チェック | アプリ全体への一律適用 |
| リソース | 不要(インライン) | guardrail の作成・バージョン管理が必要 |
| 検査内容の項目 | 呼ぶたびに `checks` で明示的に指定 | guardrail を作るときに設定項目を選ぶ。guardrail に設定した項目はすべて検査される(呼び出し側で絞れない) |
| 選べる範囲 | 3 種類(`contentFilter` / `promptAttack` / `sensitiveInformation`)から選ぶ | 2 章のフィルタを guardrail に設定できる(Denied topics、Word filters、Contextual grounding、Automated Reasoning も含む) |
| 判定 | スコアを返すだけ | 設定した閾値で自動 block / mask |
| 向く構成 | エージェント型 | 通常のリクエスト/レスポンス型 |

- 対応 checks:

| check | カテゴリ | スコア |
|---|---|---|
| `contentFilter` | HATE / VIOLENCE / SEXUAL / INSULTS / MISCONDUCT | severity |
| `promptAttack` | JAILBREAK / PROMPT_INJECTION / PROMPT_LEAKAGE | severity |
| `sensitiveInformation` | PII 31 種類(下記) | confidence |

`sensitiveInformation` の `type` に指定できる PII の全タイプ(31 種類)。boto3 1.43.114 の入力定義と、AgentCore Policy の公式ページ(policy-guardrails-in-policies)の一覧が、過不足なく一致することを確認済み。

| 分類 | タイプ |
|---|---|
| 個人情報(11) | `ADDRESS`, `AGE`, `NAME`, `EMAIL`, `PHONE`, `USERNAME`, `PASSWORD`, `PIN`, `DRIVER_ID`, `LICENSE_PLATE`, `VEHICLE_IDENTIFICATION_NUMBER` |
| 金融(7) | `CREDIT_DEBIT_CARD_NUMBER`, `CREDIT_DEBIT_CARD_CVV`, `CREDIT_DEBIT_CARD_EXPIRY`, `US_BANK_ACCOUNT_NUMBER`, `US_BANK_ROUTING_NUMBER`, `INTERNATIONAL_BANK_ACCOUNT_NUMBER`, `SWIFT_CODE` |
| 政府・医療(8) | `US_SOCIAL_SECURITY_NUMBER`, `US_PASSPORT_NUMBER`, `US_INDIVIDUAL_TAX_IDENTIFICATION_NUMBER`, `CA_SOCIAL_INSURANCE_NUMBER`, `CA_HEALTH_NUMBER`, `UK_NATIONAL_INSURANCE_NUMBER`, `UK_NATIONAL_HEALTH_SERVICE_NUMBER`, `UK_UNIQUE_TAXPAYER_REFERENCE_NUMBER` |
| 認証情報・ネットワーク(5) | `AWS_ACCESS_KEY`, `AWS_SECRET_KEY`, `IP_ADDRESS`, `MAC_ADDRESS`, `URL` |

- 各タイプの定義(形式、桁数など)は、公式の [Sensitive information filters](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-sensitive-filters.html) に載っている
- 同ページの PII 一覧と `checks` の 31 種類が完全に同じかは未確認(Policy の一覧とは一致)。ただし独自 regex は `checks` では使えない(`type` は上記の列挙値のみ)

- スコアはどちらも {0, 0.2, 0.4, 0.6, 0.8, 1.0} の離散値。severity は内容の強さ、confidence は PII が存在する確からしさ
- **Prompt Attack を `checks` の独立した項目として指定でき、`JAILBREAK` / `PROMPT_INJECTION` / `PROMPT_LEAKAGE` のカテゴリも選べる**(boto3 の定義で確認済み)
  - guardrail を使う方式(ApplyGuardrail など)では、Prompt Attack は Content filters の設定の 1 項目(`type: PROMPT_ATTACK`)で、設定できるのは強度とアクションのみ。カテゴリを個別に選ぶ項目は、読んだ範囲では見つからず
- PII 結果には `messageIndex`、`contentIndex`、`beginOffset`、`endOffset` が付くので、クライアント側でマスク処理できる
- IAM: `bedrock:InvokeGuardrailChecks`、Resource は `"*"`(ブログの説明)。`aws:RequestedRegion` 等の条件で絞れる
- 対応リージョン(発表ページ): us-east-1、us-east-2、us-west-2、London、Stockholm、Tokyo、Sydney
- 料金は 1 節の表(content $0.07 / prompt attack $0.08 / PII $0.10、1,000 units あたり)

**閾値と違反時の扱い**

- `InvokeGuardrailChecks` 自体には閾値の引数がない。API はスコアを返すだけで、ブロックもエラーもしない。違反かどうかの判定と、その後の動作(ブロック / 再試行 / ログ / 人に回す)はすべてアプリ側で実装する
- 閾値は呼び出し側が決める。ブログの例は、0.8 以上でブロック、0.4 以上で escalate、それ未満はログのみ。金融系アプリは 0.4 でブロック、創作系は 0.8 でブロック、のように用途で変える
- 検出がない(スコア 0 のみ)場合の `results` が空配列になるか、0 を返すかは未確認
- 閾値を API に渡す形式にしたい場合は、LiteLLM の checks モード(5.8)を使う。LiteLLM が閾値を持ち、超えたら HTTP 400 を返す

判定のコード例（ブログの「ブロック / escalate / ログ」の考え方を元にした例。実際に AWS を呼んでの検証はしていない）:

```python
BLOCK_AT = 0.8      # 以上ならブロック
ESCALATE_AT = 0.4   # 以上なら人に回す / 再確認

def judge(response):
    worst = 0.0
    for name in ("contentFilter", "promptAttack"):
        for f in response["results"].get(name, {}).get("results", []):
            worst = max(worst, f["severityScore"])
    pii = [
        p for p in response["results"].get("sensitiveInformation", {}).get("results", [])
        if p["confidenceScore"] >= ESCALATE_AT
    ]
    if worst >= BLOCK_AT:
        return "block"
    if worst >= ESCALATE_AT or pii:
        return "escalate"
    return "allow"

action = judge(r)   # r は下の invoke_guardrail_checks の戻り値
```

```python
# AWS ブログの記述に基づく例。引数名と構造は boto3 1.43.114 の入力定義で確認済み
# (古い boto3 には invoke_guardrail_checks がないので、boto3 を最新にすること)
# 実際に AWS を呼んだ検証はしていない
import boto3

# 対応リージョン: us-east-1、us-east-2、us-west-2、London、Stockholm、Tokyo、Sydney
bedrock_runtime = boto3.client("bedrock-runtime", region_name="us-east-1")

r = bedrock_runtime.invoke_guardrail_checks(
    messages=[{"role": "user", "content": [{"text": "..."}]}],
    checks={
        "contentFilter": {"categories": [{"category": "VIOLENCE"}]},
        "promptAttack": {"categories": [{"category": "PROMPT_INJECTION"}]},
        "sensitiveInformation": {"entities": [{"type": "EMAIL"}]},
    },
)
# r["results"]["promptAttack"]["results"] -> [{"category": ..., "severityScore": ...}]
```

## 4. AgentCore Gateway との組み合わせ

Gateway 自体に guardrail ID を指定する設定はない。公式の連携は AgentCore Policy の guardrail 条件(4.1)で、guardrail リソースは使わず `InvokeGuardrailChecks` をインラインで呼ぶ。Policy で使えない Denied topics・Word filter・PII マスクなどが必要なら、Interceptor Lambda から `ApplyGuardrail` を呼ぶ(4.2、自前実装)。

### 4.1 AgentCore Policy の guardrail 条件(公式機能)

Gateway に Policy Engine を attach、Cedar/Dogwood ポリシー内で guardrail スコアを条件に使用。ツール呼び出しの事前チェックに最適。

```
# MCP ツール呼び出し時の入力検査例
forbid (principal,
  action == AgentCore::Action::"<TargetName>___<ToolName>",
  resource == AgentCore::Gateway::"<gateway-arn>")
when guardrails {
  BedrockGuardrails::ContentFilter(["VIOLENCE"], [context.input.<ツールの引数名>])
    ["VIOLENCE"].confidenceScore.greaterThan(decimal("0.2"))
};
# data path は文字列に解決される必要あり。オブジェクトや配列を指す field path は作成時に拒否される
# 配列内のテキストは "$.messages[*].content" のような JSONPath で指定(作成時に検証されない)

# AgentCore Runtime（エージェント）への入力検査例
forbid (principal,
  action == AgentCore::Action::"MyTarget___POST:/invocations",
  resource == AgentCore::Gateway::"<gateway-arn>")
when guardrails {
  BedrockGuardrails::PromptAttack(["PROMPT_INJECTION"], [context.input.prompt])
    ["PROMPT_INJECTION"].confidenceScore.greaterThan(decimal("0.6"))
};
```

**対応範囲と制限**:
- 使える guardrail は `ContentFilter`、`PromptAttack`、`SensitiveInformation` の3種類のみ
  - Denied topics、Word filter、Contextual grounding、Automated Reasoning は不可
- 内部呼び出しは `bedrock:InvokeGuardrailChecks`（`InvokeGuardrailChecks` API）。Gateway 実行ロールに `bedrock:InvokeGuardrailChecks` 権限が必要、credential は FAS(forward access session)を使用
- `ContentFilter` と `PromptAttack` は confidence score を返す。値は {0, 0.2, …, 1.0} の離散値
  - `SensitiveInformation` はスコアを持たず、検出ごとの結果を `count()` / `categories()` などで判定する
  - 推奨初期閾値: ContentFilter=0.2、PromptAttack=0.4、SensitiveInformation=0.2
- guardrail リソースは不要と考えられる。明文はなく、構文に guardrail ID を渡す場所がないこと、呼んでいる API(`InvokeGuardrailChecks`)が resourceless であることからの推論
- 2026-06-17 の AWS 発表で「Policy が Bedrock Guardrails に対応」と告知。発表では「各ツール呼び出しの入力」と「許可された各アクションの出力」の両方を検査すると説明
- IAM の Resource 指定に食い違いあり: AgentCore ドキュメントの例は guardrail ARN を指定、Bedrock 側のドキュメントとブログは `InvokeGuardrailChecks` を resourceless として `Resource: "*"` と説明。`"*"` で付与するのが安全と思われるが、ARN 指定時の挙動は未検証
- 効果は3種類:
  - `forbid`: 閾値超過でリクエスト拒否
  - `permit`: 閾値以下で許可。Policy は deny-by-default なので `permit` で書くのが自然
  - `suppressOutput`: ツール実行後、応答だけ握りつぶし

**入出力の制約**:
- `permit` / `forbid` は `context.input.*` のみ、`suppressOutput` は `context.output.*` のみを読み込み可能
- 1回の guardrail 呼び出しで入出力を混ぜ不可

**対象ターゲット**:
- **MCP ツール呼び出し**: `action == AgentCore::Action::"<TargetName>___<ToolName>"` — TargetName は Gateway に登録したターゲット名、ToolName は MCP ツール名
- **AgentCore Runtime（エージェント）**: `action == AgentCore::Action::"<TargetName>___POST:/invocations"`
- **推論ターゲット**: `/v1/chat/completions`、`/v1/responses`、`/v1/messages`、`GET:/v1/models`
- **HTTP proxy**: `action == AgentCore::Action::"<TargetName>___<METHOD>:<uri>"`

**対応リージョン(guardrail 条件に限った話)**: us-east-1、us-east-2、us-west-2、Sydney、Tokyo、London、Stockholm（7 つ、AgentCore ドキュメントの記載）
- Policy 自体は、これより広い 22 リージョンで使える。guardrail 条件が使えるのが上記 7 つ
- 2026-01-15 に Oregon と Ohio へ拡大された経緯あり
- 2026-06-17 の発表ページには N. Virginia / London / Stockholm / Sydney / Tokyo の 5 つしか載っていない。ドキュメントと発表ページで数が異なる理由は不明(ドキュメントの方が新しいかは未確認)
- ソウル・シンガポール・フランクフルトなどは、ドキュメント上は未対応

**運用上の注意**:
- 最初は `LOG_ONLY` で閾値調整、その後本番適用に移行。モードは 2 階層ある
  - Policy Engine(Gateway 側の `policyEngineConfiguration.mode`): `LOG_ONLY` → `ENFORCE`
  - 個別 policy(`enforcementMode`): `LOG_ONLY` → `ACTIVE`(既定 `ACTIVE`)
  - engine が `LOG_ONLY` だと、個別 policy が `ACTIVE` でも強制されない
  - `bedrock-agentcore:UpdateGateway` 権限があれば engine を `LOG_ONLY` に変えられ、全 forbid が効かなくなる
- JSONPath データパスは作成時に検証されない。誤るとリクエスト全体が拒否(fail closed)またはそのチェックだけスキップされる
- guardrail は非決定的。閾値付近の入力は結果が揺れる
- `SensitiveInformation` は per-category スコア (カテゴリ別インデックス) を持たず、`count()`、`categories()`、`maxConfidenceScore()` などで扱う

### 4.2 Interceptor Lambda から `ApplyGuardrail` を呼ぶ(自前実装)

- Gateway にはリクエスト/レスポンス用インターセプター1つずつ設定可(Lambda のみ)
- リクエスト側の出力に `transformedGatewayResponse` を入れると、**ターゲット呼び出しなしで即座に返答可**(ブロック)
- この中で `ApplyGuardrail` を呼べば、**Policy では使えない Denied topics、Word filter、PII マスクも使用可**。ドキュメント内に公式サンプルなし、設計案。
- スコアで自前判定したいだけなら `InvokeGuardrailChecks`(3.4)も使える。guardrail リソース不要で、判定ロジックを Lambda に書ける

**注意点**:
- MCP ターゲットはボディが parsed JSON、HTTP/推論ターゲットは base64 文字列
- Lambda 同期呼び出しは6MBが上限。推論応答は大きくなるため `RESPONSE_BODY` を payload filter で除外推奨

### 4.3 推論ターゲットでの利用(4.1 の補足)

- `/inference/v1/...` で OpenAI/Anthropic 互換エンドポイント提供、`bedrock-mantle`・`openai`・`anthropic` のコネクタに model 名で振り分け
- Policy の guardrail 条件(4.1)は推論ターゲットの `/v1/chat/completions` 等にも使える
- それ以外の guardrail 適用方法(Gateway 自体の設定など)は読んだ範囲では確認できず

## 5. LiteLLM との組み合わせ

LiteLLM は内部で `ApplyGuardrail`(`guardrailIdentifier` 指定時)または `InvokeGuardrailChecks`(`checks` 指定時、5.8)を呼び出す。対象モデルは Bedrock に限定されない。

### 5.1 設定例

```yaml
guardrails:
  - guardrail_name: "bedrock-pre-guard"
    litellm_params:
      guardrail: bedrock
      mode: "during_call"          # pre_call / post_call / during_call
      guardrailIdentifier: ff6ujrregl1q
      guardrailVersion: "1"        # "DRAFT" も可
      aws_region_name: os.environ/AWS_REGION
      aws_role_name: os.environ/AWS_ROLE_ARN
      default_on: true             # 全リクエストに適用
```

### 5.2 mode オプション

- `pre_call`: LLM 呼び出し前に入力検査
- `post_call`: 呼び出し後に入出力検査
- `during_call`: LLM と並列で入力検査、完了まで応答保留（汎用ページでは「ブロックのみ」。LLM と並列なので、マスクが間に合う保証がない）
- `pre_mcp_call`、`during_mcp_call`: MCP ツール呼び出し用
- 上記 5 つが、LiteLLM のソース(bedrock_guardrails.py の `get_supported_event_hooks`)に書かれた Bedrock 対応 mode
- `logging_only`(スキャン結果をログに記録するだけでブロックしない)と `post_mcp_call` は LiteLLM 全体の mode としては存在するが、Bedrock の対応リストには含まれない。Bedrock で使えるかは未検証
- ソースは GitHub の main ブランチを確認したもので、手元で使う LiteLLM のバージョンと一致するとは限らない
- Enterprise では map ベース(`tags` と `default`)でリクエスト属性に応じて mode を切り替え可能

**複数の mode を指定する場合**

`mode` はリストで書ける(LiteLLM quick_start の原文に「A list is supported to run multiple modes, e.g. mode: [pre_call, post_call]」とあり、確認済み)。guardrail のブロックを mode ごとに分けて書く必要はなく、1 ブロックで足りる。

```yaml
guardrails:
  - guardrail_name: "bedrock-guard"
    litellm_params:
      guardrail: bedrock
      mode: ["pre_call", "post_call"]   # LLM に送る前と、回答を受け取った後の両方を検査
      guardrailIdentifier: ff6ujrregl1q
      guardrailVersion: "1"
      aws_region_name: os.environ/AWS_REGION
```

- 入力と出力の両方を検査したい場合の典型例。`pre_call` で入力を LLM に送る前に止め、`post_call` で回答を検査する
- `post_call` は入力も検査対象(Bedrock 個別ページの記載)。ただし LLM を呼んだ後の検査になるので、入力を LLM に送る前に止めたい場合は `pre_call` が必要
- 入力用と出力用で別の guardrail(閾値や有効なフィルタが違うもの)を使いたい場合に限り、ブロックを 2 つに分ける。その場合はリクエストの `guardrails` に両方の `guardrail_name` を並べる
- 未確認: `["pre_call", "post_call"]` にしたとき、入力が 2 回検査されて料金が二重にかかるかは、読んだ範囲では不明。Bedrock 個別ページの設定例で、リスト指定を実際に使った例は見つかっていない

### 5.3 リクエスト単位の指定

```json
{"guardrails": ["bedrock-pre-guard"]}
```

- 違反時は HTTP 400 と `bedrock_guardrail_response` が返る
- `disable_exception_on_block: true` で HTTP 200 で返す(OpenWebUI 向け)

### 5.4 マスク機能

- Bedrock 個別ページの記載: マスクには `mode: pre_call` で `mask_request_content: true` / `mask_response_content: true` を設定
- LiteLLM の汎用ページ(custom_guardrail)の mode 表では、`post_call` もマスク可、`during_call` はブロックのみ
- 迷ったら `pre_call` を使う(Bedrock 個別ページで明記されているのはこれだけ)
- ドキュメント(PDF、ファイル)は書き換え不可。マスク対象になるとリクエストごとブロック
- **マスクは `guardrailIdentifier` 方式でのみ有効**。`checks` モード(5.8)では効果がなく、警告ログが出るだけ（`InvokeGuardrailChecks` が detect-only で、書き換えた内容を返さないため。ソース bedrock_guardrails.py のコンストラクタで確認）。`checks` で PII を見つけたときの動作は、ブロックかログのみの二択

設定例（LiteLLM 公式ページ「PII Masking with Bedrock Guardrails」の例）:

```yaml
guardrails:
  - guardrail_name: "bedrock-pre-guard"
    litellm_params:
      guardrail: bedrock
      mode: "pre_call"  # Important: must use pre_call mode for masking
      guardrailIdentifier: wf0hkdb5x07f
      guardrailVersion: "DRAFT"
      aws_region_name: os.environ/AWS_REGION
      aws_role_name: os.environ/AWS_ROLE_ARN
      mask_request_content: true    # Enable masking in user requests
      mask_response_content: true   # Enable masking in model responses
```

- Bedrock の guardrail が介入したとき、LiteLLM は guardrail が返したマスク済みの出力を読み取り、それをモデルに送る
- マスクの例: 「My email is john.doe@example.com and my phone number is 555-123-4567」を送ると、モデルには「My email is [EMAIL] and my phone number is [PHONE_NUMBER]」と送られる
- この例の `guardrailIdentifier` は guardrail 側で PII の ANONYMIZE を設定している前提（マスクの判断基準は、guardrail に設定した PII のタイプとアクション）
- `mask_response_content: true` にすると、ストリーミングはチャンクを逐次流さず、バッファしてから返す（ソースの `_streams_incrementally` で確認）

**Bedrock　Guardrails側のアクション設定と LiteLLM の挙動**

Bedrock Guardrails側のアクションは、コンソールの表示名と API の値が違う（コンソールは実機で確認済み）。

| 動作 | コンソールの表示 | Bedrock Guardrailsの設定値(API) | 判定結果(ApplyGuardrail のレスポンス) |
|---|---|---|---|
| ブロック | Block | `BLOCK` | `BLOCKED` |
| マスク | Mask | `ANONYMIZE` | `ANONYMIZED` |
| 検出のみ | Detect (no action) | `NONE` | (介入なし) |

- LiteLLM が HTTP 400 を返すのは、Bedrock 側のアクションが Block のとき（判定結果では `BLOCKED`）だけ。マスクや検出のみのときは 400 にならない（ソースの `_should_raise_guardrail_blocked_exception` で確認）
- **マスクにするには、Bedrock 側で PII のアクションを Mask にする**。Block のままだと、LiteLLM で `mask_request_content` / `mask_response_content` を `true` にしてもブロックされる
- PII は入力と出力でアクションを別々に設定できる（コンソールの Input action / Output action）。マスクしたい側を Mask にする
- Prompt Attack は Block と Detect (no action) の 2 択で、Mask はない
- ソースには、guardrail がマスク済みの出力を返したときはフラグの値に関係なくそれを適用する、というコメントがある。ただしこれが全経路（`pre_call`、`post_call`、ストリーミング）で成り立つかは未確認。公式の手順どおり、フラグも `true` にしておく

### 5.5 ストリーミング

- `post_call` で検査。既定 `streaming_buffer_until_moderated: true` はチャンクをバッファし、モデレーション通過後にまとめて返す
- `streaming_buffer_until_moderated: false` で即時返送（検査完了まで不適切な内容が含まれる可能性）
- `streaming_end_of_stream_only: true` で全チャンク受信後に1回検査（監査モード）
- `streaming_sampling_rate`(既定5): 累積テキスト Nth チャンクごとに検査（`streaming_buffer_until_moderated: false` 時）
- バッファ無し + `streaming_end_of_stream_only: false` だと、ブロック判定後も既に配信済みのチャンクは戻らない

### 5.6 画像・添付ファイル

- 検査可能: base64 の PNG/JPEG(4MB、8000px まで)、テキストソース文書
- 検査不可: PDF、音声、動画、URL 画像など → HTTP 400
- `skip_unscannable_attachments: true` で素通し可能

### 5.7 Contextual grounding

- `contextual_grounding_from_messages: true` で system prompt をソース、最新 user メッセージをクエリとして送信。クエリ長制限(約1,000文字)に注意

### 5.8 `checks` モード(guardrail リソース不要)

- `guardrailIdentifier` の代わりに `checks` を指定、`InvokeGuardrailChecks` API で contentFilter・promptAttack・sensitiveInformation を直接使用可
- **`guardrailIdentifier` と `checks` は併用できない**(LiteLLM 公式: "the two can't be combined")。どちらか一方だけを書く
- 設定例(LiteLLM 公式ページの例):

```yaml
guardrails:
  - guardrail_name: "bedrock-checks"
    litellm_params:
      guardrail: bedrock
      mode: "pre_call"
      aws_region_name: os.environ/AWS_REGION
      checks:                          # guardrailIdentifier / guardrailVersion は書かない
        contentFilter:
          categories:
            - category: VIOLENCE
        promptAttack:
          categories:
            - category: JAILBREAK
        sensitiveInformation:
          entities:
            - type: EMAIL
      content_filter_threshold: 0.5
      prompt_attack_threshold: 0.5
      pii_confidence_threshold: 0.5
```

- `checks` は使いたい項目だけ書く。最低 1 つ必須。`promptAttack: {}` のように中身を空にすると、AWS の既定の設定でその検査が有効になる
- 閾値キー: `content_filter_threshold`、`prompt_attack_threshold`、`pii_confidence_threshold`
  - **3 つとも指定は任意**。省略すると 0.5 になる（LiteLLM 公式の記述と、ソースの既定値 `= 0.5` の両方で確認済み）
  - 上の例のように 0.5 を明示するのは、値を変えやすくするため。省略しても同じ動作
  - `checks` に書いていない検査の閾値キーは使われない(ソースで確認済み。`checks` にない検査は判定をスキップする)
- スコアが閾値以上だと HTTP 400 でブロック
- **`null` にするのは、閾値のキー(`litellm_params` 直下の `content_filter_threshold` など)の値**。その検査だけ、スコアを記録してブロックしない。他の検査は通常どおりブロックする

```yaml
      checks:
        contentFilter:
          categories:
            - category: VIOLENCE
        promptAttack:
          categories:
            - category: JAILBREAK
      content_filter_threshold: null   # contentFilter はスコアを記録するだけ(ブロックしない)
      prompt_attack_threshold: 0.5     # promptAttack は 0.5 以上でブロック
```

- ソース(bedrock_guardrails.py)では、閾値が `None` の検査は判定をスキップする実装になっている
- PII の結果が途中で切れた(truncated)場合はブロックする(fail closed)
- このモードでは、リクエスト側の guardrail パラメータは無視される(呼び出し側が検査を弱められない)。入力はすべて user の内容として検査される
- テキストしか読めないため、画像はすべて検査不可の添付として扱われ HTTP 400 になる(`skip_unscannable_attachments: true` で回避可)
- ブロック時のレスポンスには `bedrock_guardrail_checks`(`check`、`category`、`severityScore`)が入る
- 料金(1,000 units あたり): contentFilter $0.07、promptAttack $0.08、PII $0.10
  - 通常の guardrail より安いのは content filter だけ($0.15 → $0.07)。PII は同額
  - prompt attack は通常の料金表に独立した行がない。content filter + prompt attack を併用すると $0.15 で、通常の content filter 単価と同じ
  - ApplyGuardrail 専用の料金表はなく、通常表が適用されると読める(明文なし)
- Prompt Attack は `checks` の独立した項目(`promptAttack`)で、カテゴリも選べる。通常の guardrail では Content filters の 1 項目
- `guardrailIdentifier` と併用不可
- `bedrock:InvokeGuardrailChecks` 権限が必要

### 5.9 キー・チーム・モデル単位の制御

- キー単位・モデル単位の指定は Enterprise 機能

### 5.10 ログ

- spend log の `guardrail_information` に結果が残る

## 6. 使い分けの目安

| やりたいこと | 向く方法 |
|---|---|
| Bedrock の呼び出しを全社で強制的にガード | Organizations の強制適用 |
| エージェントのツール呼び出しの入出力を統制 | Gateway Policy の guardrail 条件(4.1) |
| Gateway で PII マスク・トピック制限もしたい | Interceptor Lambda + ApplyGuardrail(4.2) |
| 複数プロバイダモデルを1か所でガード | LiteLLM の Bedrock guardrail(5) |
| RAG の幻覚抑制 | Contextual grounding(日本語未対応) |
| ルールの厳密な遵守検証 | Automated Reasoning(英語のみ) |

## 7. 2026年の主な更新

| 日付 | 内容 |
|---|---|
| 2026-01-15 | AgentCore Policy の Guardrails 連携が US West (Oregon) / US East (Ohio) に拡大 |
| 2026-04-03 | クロスアカウント強制適用(組織・アカウント単位)が GA |
| 2026-06-16 | `InvokeGuardrailChecks` API 登場(resourceless / detect-only) |
| 2026-06-17 | AgentCore Policy が Bedrock Guardrails に対応と発表(ツール入力と出力を検査) |
| 2026-06-23 | Automated Reasoning に 2 つのポリシー改善ワークフロー追加(テストを通すための反復改善、曖昧さを減らす変数定義の自動改善) |

- 発表ページが「GA」と明記していないものがある(`InvokeGuardrailChecks`、Policy 連携)。本番採用前に最新ドキュメントで状態を確認する
- 2026-06 以降の更新は未確認

## 8. 参考資料

- [What's New: InvokeGuardrailChecks](https://aws.amazon.com/about-aws/whats-new/2026/06/amazon-bedrock-guardrails-api-ai/)
- [Blog: InvokeGuardrailChecks API](https://aws.amazon.com/blogs/machine-learning/safeguard-your-agentic-ai-applications-with-the-amazon-bedrock-guardrails-invokeguardrailchecks-api/)
- [What's New: クロスアカウント強制適用 GA](https://aws.amazon.com/about-aws/whats-new/2026/04/bedrock-guardrails-cross-account-safeguards/)
- [What's New: AgentCore Policy と Guardrails](https://aws.amazon.com/about-aws/whats-new/2026/06/amazon-bedrock-agentcore-policy-guardrails-generally-available/)
- [What's New: Automated Reasoning の改善ワークフロー](https://aws.amazon.com/about-aws/whats-new/2026/06/amazon-bedrock-guardrails/)
- [Amazon Bedrock Guardrails](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails.html)
- [Safeguard tiers](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-tiers.html)
- [ApplyGuardrail API](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-use-independent-api.html)
- [Guardrails enforcements](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-enforcements.html)
- [Supported languages](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-supported-languages.html)
- [Automated Reasoning checks](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-automated-reasoning-checks.html)
- [Bedrock pricing](https://aws.amazon.com/bedrock/pricing/)
- [AgentCore Policy: Guardrails in policies](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/policy-guardrails-in-policies.html)
- [Gateway interceptors](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-interceptors-types.html)
- [Gateway inference targets](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-target-inference-connector.html)
- [LiteLLM Bedrock Guardrails](https://docs.litellm.ai/docs/proxy/guardrails/bedrock)
- [LiteLLM Guardrails quick start](https://docs.litellm.ai/docs/proxy/guardrails/quick_start)
