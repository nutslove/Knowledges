# Agent Evaluation
- 参考URL
  - https://www.langchain.com/conceptual-guides/agent-observability-powers-agent-evaluation
  - https://langfuse.com/guides/cookbook/example_pydantic_ai_mcp_agent_evaluation
  - https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents
  - https://aws.amazon.com/jp/blogs/machine-learning/evaluating-ai-agents-real-world-lessons-from-building-agentic-systems-at-amazon/

---

# Agentの評価観点（戦略）
主に **「Final Response（最終出力）」**、**「Single step（各ステップ）」**、**「Trajectory（軌跡）」** の3つの観点があるとされている。

- https://langfuse.com/guides/cookbook/example_pydantic_ai_mcp_agent_evaluation  
  > 1）**Final Response Evaluation (Black-Box)**:
  This method evaluates only the user’s input and the agent’s final answer, ignoring the internal steps entirely. It’s the simplest to set up and works with any agent framework, but it cannot tell you why a failure occurred.
  >
  > 2）**Trajectory Evaluation (Glass-Box)**:
  This method checks whether the agent took the “correct path.” It compares the agent’s actual sequence of tool calls against the expected sequence from a benchmark dataset. When the final answer is wrong, trajectory evaluation pinpoints exactly where in the reasoning process the failure occurred.
  >
  > 3）**Single Step Evaluation (White-Box)**:
  This is the most granular evaluation strategy, acting like a unit test for agent reasoning. Instead of running the whole agent, it tests each decision-making step in isolation to see if it produces the expected next action. This is especially useful for validating that search queries, API parameters, or tool selections are correct.
- https://docs.langchain.com/langsmith/evaluate-complex-agent  
  > - **Final response**: Evaluate the agent’s final response.
  > - **Trajectory**: Evaluate whether the agent took the expected path (e.g., of tool calls) to arrive at the final answer.
  > - **Single step**: Evaluate any agent step in isolation (e.g., whether it selects the appropriate first tool for a given step).

---

# Agentの評価基準
- 参考URL
  - https://docs.cloud.google.com/vertex-ai/generative-ai/docs/models/evaluation-agents
  - https://aws.amazon.com/jp/blogs/machine-learning/evaluating-ai-agents-real-world-lessons-from-building-agentic-systems-at-amazon/

> - **Final response quality**:
>   - **Correctness**: The factual accuracy and correctness of an AI assistant’s response to a given task.
>   - **Faithfulness**: Whether an AI assistant’s response remains consistent with the conversation history.
>   - **Helpfulness**: How effectively an AI assistant’s response helps users appropriately address query and progress toward their goals.
>   - **Response relevance**: How well an AI assistant’s response addresses the specific question or request.
>   - **Conciseness**: How efficiently an AI assistant communicates information, for instance, whether the response is appropriately brief without missing key information.
> - **Task completion**: 
>   - **Goal success**: Did the AI assistant successfully complete all user goals within a conversation session.
>   - **Goal accuracy**: Compares the output to the ground truth.
> - **Tool use**:
>   - **Tool selection accuracy**: Did the AI assistant choose the appropriate tool for a given situation.
>   - **Tool parameter accuracy**: Did the AI assistant correctly use contextual information when making tool calls.
>   - **Tool call error rate**: The frequency of failures when an AI assistant makes tool calls.
>   - **Multi-turn function calling accuracy**: Are multiple tools being called and how often the tools are called in the correct sequence.
> - **Memory**:
>   - **Context retrieval**: Assesses the accuracy of findings and surfaces the most relevant contexts for a given query from memory, prioritizing relevant information based on similarity or ranking, and balancing precision and recall.
> - **Multi-turn**: 
>   - **Topic adherence classification**: If a multi-turn conversation includes multiple topics, assesses whether the conversation stays on predefined domains and topics during the interaction.
>   - **Topic adherence refusal**: Determines if the AI agent refuse to answer questions about a topic.
> - **Reasoning**:
>   - **Grounding accuracy**: Does the model understand the task, appropriately select tools, and is the CoT aligned with the provided context and data returned by external tools.
>   - **Faithfulness score**: Measures logical consistency across the reasoning process.
>   - **Context score**: Is each step taken by the agent contextually grounded.
> - **Responsibility and safety**:
>   - **Hallucination**: Do the outputs align with established knowledge, verifiable data, logical inference, or include any elements that are implausible, misleading, or entirely fictional.
>   - **Toxicity**: Do the outputs contain language, suggestions, or attitudes that are harmful, offensive, disrespectful, or promote negativity. This include content that might be aggressive, demeaning, bigoted, or excessively critical without constructive purpose.
>   - **Harmfulness**: Is there potentially harmful content in an AI assistant’s response, including insults, hate speech, violence, inappropriate sexual content, and stereotyping.

> - **最終回答の品質（Final response quality）**:
>   - **正確性（Correctness）**: 与えられたタスクに対するAIアシスタントの回答の事実的な正確さと正しさ。
>   - **忠実性（Faithfulness）**: AIアシスタントの回答が会話履歴と一貫しているかどうか。
>   - **有用性（Helpfulness）**: AIアシスタントの回答が、ユーザーの質問に適切に対処し、目標達成に向けて効果的に支援できているか。
>   - **回答の関連性（Response relevance）**: AIアシスタントの回答が、特定の質問やリクエストにどれだけ的確に対応しているか。
>   - **簡潔性（Conciseness）**: AIアシスタントが情報を効率的に伝達しているか。例えば、重要な情報を漏らさずに適切に簡潔な回答になっているか。
> - **タスク完了（Task completion）**:
>   - **目標達成（Goal success）**: 会話セッション内で、AIアシスタントがユーザーの全ての目標を正常に完了したか。
>   - **目標精度（Goal accuracy）**: 出力をグラウンドトゥルース（正解データ）と比較した結果。
> - **ツール使用（Tool use）**:
>   - **ツール選択精度（Tool selection accuracy）**: 与えられた状況に対して、AIアシスタントが適切なツールを選択したか。
>   - **ツールパラメータ精度（Tool parameter accuracy）**: ツール呼び出し時に、AIアシスタントがコンテキスト情報を正しく使用したか。
>   - **ツール呼び出しエラー率（Tool call error rate）**: AIアシスタントがツール呼び出しを行った際の失敗頻度。
>   - **マルチターン関数呼び出し精度（Multi-turn function calling accuracy）**: 複数のツールが呼び出されているか、またツールが正しい順序で呼び出されている頻度。
> - **メモリ（Memory）**:
>   - **コンテキスト取得（Context retrieval）**: メモリから与えられたクエリに対して最も関連性の高いコンテキストを正確に検出・提示できているかを評価する。類似度やランキングに基づいて関連情報を優先し、適合率（Precision）と再現率（Recall）のバランスを取る。
> - **マルチターン（Multi-turn）**:
>   - **トピック遵守分類（Topic adherence classification）**: マルチターン会話に複数のトピックが含まれる場合、対話中に事前定義されたドメインやトピックに沿った会話が維持されているかを評価する。
>   - **トピック遵守拒否（Topic adherence refusal）**: AIエージェントがあるトピックに関する質問への回答を適切に拒否するかどうかを判定する。
> - **推論（Reasoning）**:
>   - **根拠精度（Grounding accuracy）**: モデルがタスクを理解し、適切にツールを選択しているか、またCoT（Chain of Thought）が提供されたコンテキストや外部ツールから返されたデータと整合しているか。
>   - **忠実性スコア（Faithfulness score）**: 推論プロセス全体にわたる論理的一貫性を測定する。
>   - **コンテキストスコア（Context score）**: エージェントが行った各ステップがコンテキストに基づいているか。
> - **責任と安全性（Responsibility and safety）**:
>   - **ハルシネーション（Hallucination）**: 出力が確立された知識、検証可能なデータ、論理的推論と一致しているか、または不自然、誤解を招く、もしくは完全に架空の要素が含まれていないか。
>   - **毒性（Toxicity）**: 出力に有害、攻撃的、無礼、またはネガティブさを助長する言語、提案、態度が含まれていないか。攻撃的、侮辱的、偏見に満ちた、または建設的な目的なく過度に批判的なコンテンツを含む。
>   - **有害性（Harmfulness）**: AIアシスタントの回答に、侮辱、ヘイトスピーチ、暴力、不適切な性的コンテンツ、ステレオタイプ化など、潜在的に有害なコンテンツが含まれていないか。

---

# Agent評価のライブラリ
## DeepEval
- https://deepeval.com/guides/guides-ai-agent-evaluation
- オープンソースのLLM評価フレームワーク。pytestライクなインターフェースでLLMの出力をユニットテストできる
- 50以上のリサーチベースのメトリクスを提供（G-Eval、Hallucination検出、Tool Correctness、Answer Relevancy等）
- Agent評価には`@observe`デコレータによるトレースが必須（End-to-End / Component-Level両方とも）
- Agent専用メトリクス: PlanQualityMetric、PlanAdherenceMetric、ToolCorrectnessMetric、ArgumentCorrectnessMetric、TaskCompletionMetric、StepEfficiencyMetric等
- End-to-End eval: `evals_iterator(metrics=[...])`にメトリクスを渡してAgent全体のトレースを分析
- Component-Level eval: `@observe(metrics=[...])`で特定のコンポーネントに直接メトリクスを付与して個別に評価
- **開発時のAgent評価はローカルで完結**（`@observe` + `evals_iterator`でConfident AIなしで動作）。Confident AIが必要なのは本番での非同期評価（`metric_collection`経由）と結果の可視化ダッシュボード。Confident AIには無料プランもあり
- **Langfuseとの連携**: GEvalなどの汎用メトリクスは`LLMTestCase`にinput/outputを手動で渡すだけでトレース不要で使える。ただしAgent専用メトリクス（PlanQuality等）はDeepEvalの`@observe`トレースが必須のため、Langfuseのトレースとは別系統になる

## AgentEvals
- https://github.com/langchain-ai/agentevals
- LangChain製。Agentのtrajectory（中間ステップの軌跡）評価に特化したパッケージ
- 主な評価方法:
  - **Trajectory Match**: 期待されるtrajectoryとの比較（strict / unordered / subset / superset のマッチングモード）
  - **Trajectory LLM-as-Judge**: LLMに軌跡を判定させる（参照trajectoryなしでも可能）
  - **Graph Trajectory**: LangGraphのノード単位でのtrajectory評価。LangGraphのスレッドからtrajectoryを抽出するユーティリティも提供
- モデルは`"プロバイダー:モデル名"`のLangChain形式で指定（OpenAI以外も使用可能。Anthropic、Bedrock等）
- **LangSmith統合が前提の設計**だが、evaluator自体はただのPython関数なので、スコアをLangfuseに`create_score()`で書き戻すことは可能（ただしtrajectoryのフォーマット変換などグルーコードが必要）

## OpenEvals
- https://github.com/langchain-ai/openevals
- LangChain製。AgentEvalsの兄弟パッケージ
- **汎用LLM評価**のプリビルトプロンプト: Correctness、Conciseness、Hallucination、Toxicity、Answer Relevance、Plan Adherence
- **RAG評価**: Correctness、Helpfulness、Groundedness、Retrieval Relevance
- **コード評価**: Pyright/Mypy型チェック、サンドボックス実行
- **Agent trajectory評価**も含む（Trajectory Match、Trajectory LLM-as-Judge）。ただしREADMEではAgent特化のevalsについてはAgentEvalsを参照するよう案内されている
- AgentEvalsと同様にLangSmith統合が前提だが、evaluator自体はPython関数なのでLangfuseにスコアを書き戻し可能
- Langfuse Managed Evaluatorと機能が重複する部分が多い。Langfuseにないもの（Plan Adherence等）を補完する使い方が効率的。プロンプトはGitHubで公開されているので、Langfuse Custom Evaluatorへの移植も容易

## LangfuseのLLM as a Judge
- https://langfuse.com/docs/evaluation/evaluation-methods/llm-as-a-judge
- 詳細は「Langfuse」フォルダの「LLM as a Judgeについて.md」を参照