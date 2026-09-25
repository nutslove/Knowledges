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

## 該当ドキュメントURL

- 概要: https://docs.litellm.ai/docs/auto_router/
- 管理者向けセットアップ(config.yaml例など): https://docs.litellm.ai/docs/auto_router/setup
- 設定リファレンス(全パラメータ): https://docs.litellm.ai/docs/proxy/auto_routing
