# LiteLLM Auto Router

LiteLLMには「Auto Router」という機能があり、ゲートウェイがリクエストを分類してモデルを選ぶ仕組み。クライアント側は1つのモデル名(例: `smart-router`)だけを使えばよく、リクエスト内容に応じて裏側で適切なモデルに振り分けられる。

> [!WARNING]
> 2026-09時点ではまだベータ版。
> 設定キーやデフォルト値は今後のリリースで変わる可能性がある、とドキュメントに明記されている。v1.94.x系で提供開始、最速の開発版リリースは2026-07-14。

## 仕組み

`auto_router/complexity_router` というモデルタイプを使い、SIMPLE/MEDIUM/COMPLEX/REASONINGの4段階の複雑度ティアごとに、任意のプロバイダの任意のモデル(単一モデル、ランダムプール、Thompsonサンプリング(`adaptive: true`)されたプール)を割り当てられる。

分類(`classifier_type`)は以下の4種類 + 決定的な短絡ルールから選べる:

- **`heuristic`(既定)**: 外部LLM/APIには一切問い合わせず、LiteLLM内部(プロキシプロセス自身)のロジックでtokenCount・codePresence・reasoningMarkers・technicalTerms・simpleIndicators・multiStepPatterns・questionComplexityの7次元をスコアリングして判定する。この分類処理自体の所要時間がサブミリ秒(1ミリ秒未満)であり、追加のAPI呼び出しコスト・レイテンシは発生しない。
  - 7次元は加重平均され、0〜1のスコアになる。このスコアを`tier_boundaries`という3つの閾値(既定0.15, 0.35, 0.60)で4段階に区切る: SIMPLE(〜0.15未満)/ MEDIUM(0.15〜0.35)/ COMPLEX(0.35〜0.60)/ REASONING(0.60以上)。
  - **COMPLEX**(スコア0.35〜0.60): 技術用語・コード・長文などでスコアはそれなりに高いが、明示的な推論要求はまだ少ない状態(例: 「分散システムのアーキテクチャを説明して」)。
  - **REASONING**(スコア0.60以上): 上記より高いスコアの場合に加え、reasoningMarkers(「step by step」「think through」「analyze」等)が2つ以上出てくると、スコアが0.60未満でも例外的に強制昇格する(例: 「このバグの原因をstep by stepで分析して」)。つまりCOMPLEXは"内容が難しい"、REASONINGは"難易度が高い、または多段階の思考プロセスを明示的に要求している"という違い。
- **`llm`**: 小型/高速モデルを使い構造化出力で分類。エージェント系トラフィックでの精度が上がるが、1リクエストあたり僅かな追加コストがかかる。`classifier_llm_config`(model, timeout_ms, system_prompt)で設定。
- **`jev`**: TypeSafeの「System One Choice」評価を使う分類器。外部エンドポイント`/v1/systemone`にリクエストを送るため、プロキシサーバー側に`TYPESAFE_API_KEY`が必要。タイムアウト、サーキットブレーカー、コンテキストウィンドウ/バジェット設定などが可能。Test Routing/Test Connection時にプロバイダ課金が発生し得る。
- **`custom`**: ユーザー定義の非同期プラグイン(`classifier_plugin`)を使う。設定ファイルのみで指定可、API/UIからは不可。

これとは別に、`keyword_tier_rules`という決定的な短絡ルールがあり、どのclassifierよりも先に評価される(`semantic_keyword_matching`でパラフレーズ一致もオプション対応)。複数ルールがマッチした場合は最も高いティアにエスカレーションされる。マッチ判定は直近のユーザーターンのテキストのみを見る(ツール出力や過去ターンは対象外)。

```yaml
keyword_tier_rules:
  - keywords: ["hi", "hello", "thanks"]
    tier: SIMPLE
  - keywords: ["kubernetes", "k8s", "istio"]
    tier: REASONING
```

## config.yamlでの設定例(OSSプロキシでそのまま動く)

```yaml
model_list:
  - model_name: claude-haiku-4-5
    litellm_params:
      model: anthropic/claude-haiku-4-5
      api_key: os.environ/ANTHROPIC_API_KEY
  - model_name: claude-sonnet-5
    litellm_params:
      model: anthropic/claude-sonnet-5
      api_key: os.environ/ANTHROPIC_API_KEY
  - model_name: claude-opus-5
    litellm_params:
      model: anthropic/claude-opus-5
      api_key: os.environ/ANTHROPIC_API_KEY

  - model_name: smart-router
    litellm_params:
      model: auto_router/complexity_router
      complexity_router_config:
        tiers:
          SIMPLE: claude-haiku-4-5
          MEDIUM: claude-sonnet-5
          COMPLEX: claude-opus-5
          REASONING: claude-opus-5
        classifier_type: heuristic
      complexity_router_default_model: claude-sonnet-5
```

- `model_name`(ティア側の値)は`model_list`内の既存のdeployment名を参照する。
- `complexity_router_default_model`は、分類に失敗した場合や(`classifier_fallback: default_model`設定時などに)使われるフォールバックモデル。
- `classifier_type`を省略するとheuristic(既定)が使われる。

## 詳細設定オプション

- **`tier_labels`**: SIMPLE/MEDIUM/COMPLEX/REASONINGという既定のティア名を表示用にリネームできる(省略時は既定名のまま)。
  ```yaml
  tier_labels:
    SIMPLE: Cheap
  ```
- **`session_affinity`**(既定`false`): セッションの最初のターンで分類したモデルを、以降の全ターンに固定する。`session_id`メタデータをキーに、再分類をスキップする。プロバイダ側のプロンプトキャッシュを維持したり、Anthropicの`thinking`ブロックのようなモデル間リプレイエラーを防ぐのに有効。トレードオフ: 後続ターンが単純な内容でも、最初のターンのティアのコストをセッション全体で引きずる。
  ```yaml
  session_affinity: true
  session_affinity_ttl_seconds: 3600
  ```
- **`classifier_fallback`**: `llm`/`jev`/`custom`分類器が失敗(タイムアウト・空応答・スキーマ不一致・未知のティアなど)した際の挙動。既定は`heuristic`(組み込みスコアラーにフォールバック)、代替として`default_model`(`complexity_router_default_model`へ直行)も選べる。「複雑度以外の基準で判定する分類器」には`default_model`が向いている、とされている。
- **分類器へ渡す会話コンテキストの制御**(`llm`/`jev`分類器向け):
  - `classifier_context_window_size`(既定`3`): 分類器に渡す過去の会話ターン数。`0`で無効化。
  - `classifier_context_budget_chars`(既定`8000`): 過去ターン分のテキストに割く文字数上限(新しいターン優先)。現在の質問やシステムプロンプトは含まない。`classifier_context_per_turn_chars`でターンごとの上限も指定可能。
  - `classifier_context_include_assistant_turns`(既定`false`): trueにするとアシスタントの返答もコンテキストに含める(例: 「これは複雑な内容ですが進めますか?」のようにアシスタント側が難易度を示すケース向け)。既存運用でtrueに変えるとティア判定・コストが変動する点、分類器プロバイダへ送るデータが増える点に注意。

## コスト削減の記録方法

各リクエストのコスト削減額は以下の式で算出・記録される:

```
savings = 基準モデルのコスト - 実際に使ったモデルのコスト - 分類器自体のコスト
```

基準モデルは「設定済みティアの中で最もコストが高いティアの最高額モデル」。この値はUIの**Cost Optimization**ページ(「Auto-router savings」)、`LiteLLM_DailyUserSpend`テーブルの`autorouter_savings_spend`列、`GET /user/daily/activity` APIの`total_autorouter_savings_spend`などで確認できる。リクエスト単位の詳細(`routed_model`, `cause`, `tier`, `classifier_cost`, `autorouter_savings`)は`LiteLLM_SpendLogs`のmetadataに記録される。

> [!NOTE]
> Prometheus callbackを有効にしていても、Auto Router関連(savings/tier/classifier_cost等)はメトリクスとして開示されない(2026-09時点、ソース`litellm/integrations/prometheus.py`で`autorouter`/`complexity_router`/`savings`関連の実装なしを確認済み)。savings確認は上記のUI/API/DBのみ。

## 事前検証・テスト機能

- **Test Routing**(`POST /auto_router/test_routing`): サンプルプロンプトを分類器に通し、実際にどのモデルが選ばれるかを確認できる安全なドライラン(補完リクエスト自体は発行されない)。ただしJEV分類器やセマンティックキーワードマッチ使用時は、分類のための呼び出し自体に課金が発生し得る。
- **Test Connection**: 各ティアのモデルグループに対して疎通確認(最小リクエスト)を行い、認証情報が有効かをチェックする。JEV利用時は分類プローブも別途実行され、補完モデルへの疎通が成功していても分類器側がフォールバックを使っていればエラーとして報告される。
- **`POST /auto_router/validate_complexity_router_config`**: `/model/new`で本番反映する前に、設定内容をAPI経由でバリデーションできる(CI/CD向け)。
- **`lite autoroute`(CLIツール)**: 本番プロキシ設定を変更せずに、ローカルの使い捨てプロキシ経由でリクエストを実際のプロキシに転送し、ルーティングをその場で試せる。

## ストリーミング対応

ストリーミングリクエストにも対応。`return_raw_model_name: true`を設定すると、実際に選ばれたモデル名が非ストリーミング応答だけでなく、全ストリーミングチャンクにも含まれる。

## エラー・障害時の挙動

- `llm`/`jev`/`custom`分類器の失敗時は`classifier_fallback`の設定に従う(既定はheuristicへのフォールバック)。
- JEV分類器にはサーキットブレーカーがあり、タイムアウトを検知すると一定のクールダウン期間はフォールバックを使い、その後回復を確認しにいく。
- カスタムプラグイン分類器が判定を拒否(`None`を返す)またはエラーになった場合は「可用性の問題」として扱われ、ルーティング精度が落ちるだけでリクエスト自体は失敗しない。
- 設定ミス(例: `classify`メソッドの欠落や非async実装)は起動時に検知される。

## Admin UI経由のセットアップ(config.yaml以外の方法)

管理UIの**Models + Endpoints → Auto Router**タブからもGUIで設定可能:

1. ルーターに名前を付け、「Configure automatically」(既存デプロイからティアを自動生成)またはテンプレート(1M Context / Anthropic Family / OpenAI Family / Gemini Family / Lite)を選択。
2. 生成されたティア構成を確認し、Test Routingで動作確認。
3. 保存。
4. 「Detailed Configuration」から、キーワードルール・LLM分類器とプロンプト・エスカレーションキーワード・adaptiveプール、(JEV利用時は)分類方式・JEVモデル・タイムアウト・サーキットブレーカー・コンテキストウィンドウサイズ/文字数バジェットなどの詳細設定が可能。

## 補足

- カスタム分類ルール(`instructions`の置き換えや`tier_definitions`によるカスタムティア定義)はEnterprise向けの既存カスタム分類器機能を使うもの。ただし**組み込みのJEV分類器自体はEnterprise不要**で、組み込みLLM分類器と同じライセンスポリシーで利用できる(誤解しやすい点: JEV = Enterprise限定ではない)。
- コスト削減効果として、ドキュメントには実運用ケーススタディ(27.2万リクエスト、450+ユーザーで4ヶ月間に51.1%削減・$12,249節約)や、品質/コストベンチマーク(フロンティアモデル比87.3%の品質で74.5%安価)、プロンプトキャッシュとの比較(キャッシュのみより37〜69%安価)などが紹介されている。

## Claude Codeでの利用時のハマりポイント

config.yamlで`auto-router`というdeploymentを定義しておけば、Claude Codeの`/model`ピッカーにも選択肢として出てきて選べるようになるはず、と当初想定していたが、実際にはそう単純にはいかなかった。以下はその過程で分かったこと。

### ハマりポイント1: Claude Codeの`/model`はAI Gateway（LiteLLM）のモデル一覧をそのままでは表示しない

config.yaml側で`auto-router`のdeploymentが正しく定義され、`curl http://localhost:4000/v1/models`でも一覧に出てくる状態にもかかわらず、Claude Codeの`/model`には`auto-router`が出てこなかった。Claude Code CLIバイナリ(v2.1.274、`/opt/homebrew/Caskroom/claude-code/2.1.274/claude`)を直接読んで解析した結果、判明した仕組み:

- AI Gateway（LiteLLM。`ANTHROPIC_BASE_URL`で指定）からモデル一覧を取得する「gateway model discovery」という処理は、環境変数`CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY`を有効化しない限りそもそも動かない(既定OFF)。有効化する値は`"1"`(例: `~/.claude/settings.json`の`env`に`"CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY": "1"`)。
- この変数を立てても、discoveryの実際のfetchが走るには、別途以下のいずれかで解決できる認証情報が必要:
  - `ANTHROPIC_AUTH_TOKEN`
  - `apiKeyHelper`
  - **「承認済み」の`ANTHROPIC_API_KEY`**(`~/.claude.json`の`customApiKeyResponses.approved`/`rejected`に記録される。初回検出時に出る「Detected custom API key in environment」ダイアログでの回答が保存される)
- `ANTHROPIC_CUSTOM_HEADERS`(例: `x-litellm-api-key: ...`)経由でLiteLLM向けの認証ヘッダーを渡していても、discovery用の認証情報解決ロジックからは一切見られない(これは実際のchatリクエスト送信時にのみ使われる別経路)。認証ヘッダー名として認識されるのは`x-api-key`/`authorization`のみ。

つまりgateway model discoveryを機能させるには`ANTHROPIC_API_KEY`を承認済みの状態で設定する必要があるが、これがclaude.aiのOAuthログインと**根本的に衝突する**:

```
⚠ Both claude.ai and ANTHROPIC_API_KEY set · auth may not work as expected
  · To use claude.ai: Unset the ANTHROPIC_API_KEY environment variable, or run `claude /logout`
    and then say "No" to API key approval before logging in.
  · To use ANTHROPIC_API_KEY: run `claude /logout` to sign out of claude.ai.
```

これは誤検出ではなく実際に効く警告で、claude.aiのPro/MaxサブスクリプションによるOAuth認証と`ANTHROPIC_API_KEY`ベースの認証は同時に成立しない。つまり「claude.aiのOAuthログインを維持したまま`/model`にLiteLLMのauto-routerを表示させる」ことは、現行バージョン(2.1.274)のClaude Codeでは実現不可能という結論に至った。

#### 対処

> [!IMPORTANT]
> ##### 採用した方針
> `/model`にauto-routerを表示させること自体は諦め、代わりに`~/.claude/settings.json`の`env.ANTHROPIC_MODEL`に直接`"auto-router"`を指定することで、毎回のセッション起動時に自動的にauto-router経由でリクエストされるようにした(その場限りのコマンドライン指定ではなく、恒久的なデフォルトとして設定)。

### ハマりポイント2: `ANTHROPIC_MODEL`に未知のモデル名を指定すると警告が出る

`ANTHROPIC_MODEL=auto-router`に変更した直後、以下の警告が出るようになった:

```
"auto-router" isn't described by this version's model catalog; update Claude Code, or map it via
behavesAs on a modelPicker row (or modelOverrides, if this is a provider id for a model version it
knows). Until then, auto-compact keeps the session within 200k tokens (the context window it
assumes); if the model accepts more, append [1m] to the model name or set CLAUDE_CODE_MAX_CONTEXT_TOKENS
to the real window; CLAUDE_CODE_DISABLE_UNKNOWN_MODEL_WINDOW_ENFORCEMENT=1 restores the previous
wait-for-the-API behavior.
```

Claude Codeは組み込みのモデルカタログに載っていないモデルID(`auto-router`)についてはコンテキストウィンドウサイズなどのメタデータが分からないため、安全側に倒して200kトークンとみなし、auto-compact(自動要約)をその基準で発動させる。

#### 対処

> [!IMPORTANT]
> `~/.claude/settings.json`に`modelPicker.options`エントリを追加し、`behavesAs`で「未知のモデルIDだが実質的には既知のモデル(ここでは`claude-sonnet-5`)と同じように扱ってよい」と明示することで解消:

```json
{
  "modelPicker": {
    "options": [
      {
        "model": "auto-router",
        "label": "Auto Router",
        "description": "LiteLLM complexity router: picks haiku/sonnet/opus automatically",
        "behavesAs": "claude-sonnet-5"
      }
    ]
  }
}
```

> [!NOTE]
> ##### `behavesAs`の対象に`claude-sonnet-5`を選んだ理由
> config.yamlの`complexity_router_default_model: claude-sonnet-5`、つまりauto-router自体が定義しているデフォルト(フォールバック)モデルと一致させたため。実際にどのモデルにルーティングされるかはリクエストごとに変わり、Claude Code側からは事前に知りようがないので、auto-router自身が「基準」としているモデルに合わせておくのが妥当という判断。

関連する他の設定:
- `modelOverrides`: 「Anthropicモデル名 → プロバイダ固有のモデルID(例: BedrockのARN)」をマップする別の設定キー。今回のように「エイリアス名の裏側で実際に使われるモデルがリクエストごとに変わる」ケースには向かず、`modelPicker`+`behavesAs`の方が適切。
- `[1m]`サフィックス: モデル名の末尾に付けると「1Mトークンのコンテキストウィンドウを持つモデル」という扱いになる命名規約。
- `CLAUDE_CODE_MAX_CONTEXT_TOKENS`: 実際のコンテキストウィンドウサイズを直接指定する環境変数。
- `CLAUDE_CODE_DISABLE_UNKNOWN_MODEL_WINDOW_ENFORCEMENT=1`: 未知モデルに対する200k前提のenforcementを無効化し、APIレスポンス任せの挙動に戻す。

### 結論

- LiteLLMのAuto Router自体はAI Gateway側で正しく機能している(`curl http://localhost:4000/v1/models`や実際のリクエストで確認可能)。
- **Claude Code側の`/model`ピッカーにAI Gateway経由のカスタムモデルを表示する機能は、claude.aiの`/login`によるOAuthログインと両立しない設計になっている(2.1.274時点)。**
- 恒久的にauto-routerをデフォルトで使いたいだけなら`/model`に出す必要はなく、`env.ANTHROPIC_MODEL`への直接指定 + `modelPicker.behavesAs`によるコンテキストウィンドウ情報の補完、の2点で運用できる。
- `/model`にauto-routerを表示させること自体を諦めない道もある: `claude /logout`で claude.ai の`/login`によるOAuthログイン(Pro/Max/Enterpriseいずれのプランでも同様)を完全に解除し、`ANTHROPIC_API_KEY`ベースの認証のみで運用すればgateway model discoveryが動き、`/model`に`auto-router`が選択肢として出てくる。ただしこの場合、Pro/Maxサブスクリプション経由ではなくAnthropic API(従量課金)経由の請求に切り替わる。
  - なお、Enterpriseプランでも`/login`によるOAuthログインを使う限り同じ制約を受ける。SSO経由のAPIキー発行やBedrock/Vertex経由などの別認証方式を使うケースは今回の調査・検証の対象外で、衝突するかどうかは未確認。

## 該当ドキュメントURL

- 概要: https://docs.litellm.ai/docs/auto_router/
- 管理者向けセットアップ(config.yaml例など): https://docs.litellm.ai/docs/auto_router/setup
- 設定リファレンス(全パラメータ): https://docs.litellm.ai/docs/proxy/auto_routing
