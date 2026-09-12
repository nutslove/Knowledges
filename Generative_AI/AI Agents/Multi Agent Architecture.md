# Multi Agent パターン・アーキテクチャ

- https://www.langchain.com/blog/choosing-the-right-multi-agent-architecture

## 参考記事・ブログ（2026年以降）

> 2026年に入り、Multi-Agentは「研究デモ」から「プロダクション基盤」へ移行。各社・各記事が収束しつつある代表的なオーケストレーション/協調パターンと、実運用で見えてきた知見をまとめる。

### ベンダー公式（一次情報）

- **[Multi-agent coordination patterns: Five approaches and when to use them](https://claude.com/blog/multi-agent-coordination-patterns)** — Anthropic (Claude) 公式ブログ、2026/4/10
  - 5つの協調パターンを整理：**Generator-Verifier**（品質基準が明確な場合。生成役と検証役を分離）／**Orchestrator-Subagent**（タスク分解が明確な場合。リードエージェントが委譲・集約）／**Agent Teams**（並列・独立・長時間実行のサブタスク向け。ワーカーが永続し文脈を蓄積）／**Message Bus**（イベント駆動。エージェント数が増えるほど柔軟性が必要な場合）／**Shared State**（エージェント間で発見事項を共有しながら協働する場合）
  - 推奨：「まず Orchestrator-Subagent から始める。最も広い問題に対応でき、協調オーバーヘッドも最小」→ 観測された限界に応じて他パターンへ進化させる
- **[Multi-agent patterns](https://learn.microsoft.com/en-us/agents/architecture/multi-agent-patterns)** — Microsoft Learn、2026/7/6更新
  - Microsoft Agent Framework（AutoGen + Semantic Kernel統合、2026/2 RC）を前提に、A2A（Agent2Agent）とMCPの使い分けを整理。A2Aは異プラットフォーム間のエージェント連携（Agent Card / Task / Artifactモデル）、MCPはツール・データアクセスの標準化に用途分担
  - 設計原則：並列性を設計に組み込む／エージェント間コンテキストは必要最小限に絞る／高インパクトな操作は人間承認を必須にする
- **[AI Agent Orchestration Patterns](https://learn.microsoft.com/en-us/azure/architecture/ai-ml/guide/ai-agent-design-patterns)** — Azure Architecture Center、2026/2/12公開（2026/5/12更新）
  - Sequential（線形パイプライン。決定的な順序。前段の失敗が伝播しやすく並列性なし）／Concurrent（同一入力を複数エージェントが並列処理。矛盾する結果の調整が必要でリソース消費大）／Group Chat（共有スレッド上でエージェントが発言、chat managerが発言順を制御。合意形成やmaker-checker検証向き。会話ループに注意）／Handoff（1エージェントずつアクティブになる動的委譲。適任者が処理中に判明するタスク向き。無限ハンドオフループに注意）／**Magentic**（後述）の5パターンを比較表付きで整理
  - **[Magentic orchestration](https://learn.microsoft.com/en-us/azure/architecture/ai-ml/guide/ai-agent-design-patterns#magentic-orchestration)** の詳細：別名 *dynamic orchestration* / *task-ledger-based orchestration* / *adaptive planning*。事前に解法が決まっていないオープンエンドな問題向けで、**manager agent**が専門エージェント群と対話しながら「タスク台帳（task ledger）」を動的に構築・更新し、ゴール／サブゴールを持つ実行計画を反復的に洗練（必要ならバックトラックも）。Group Chatの発展形だが、各エージェントがツールで外部システムに実際の変更を加える点が異なる
    - **向く場面**：解法パスが未確定で複雑／複数の専門エージェントの知見を統合する必要がある／実行前後に人間がレビューできる計画書が欲しい／外部システムに影響するツールを使うため事前に手順を可視化したい
    - **避けるべき場面**：決定的に解けるタスク／台帳が不要な単純作業／低複雑度で他のシンプルなパターンで十分／時間的制約が厳しい（速度より計画の妥当性を優先するため遅い）／曖昧なゴールで停滞・無限ループが懸念される
    - **実例**：SREチームのインシデント対応自動化 — manager agentが初期タスク台帳を作成し、diagnostics／infrastructure／communication／rollbackの各エージェントと連携しながら復旧計画を動的に組み替える
    - 比較表での位置づけ：「Coordination＝Plan-build-execute（manager agentが台帳を構築・調整）」「Routing＝manager agentが動的にタスクを割当・並べ替え」「Best for＝解法パス未確定のオープンエンドな問題」「Watch out for＝収束が遅い、曖昧なゴールで停滞」
- **[Guidance for Multi-Agent Orchestration on AWS](https://docs.aws.amazon.com/solutions/multi-agent-orchestration-on-aws/)** — AWS公式
  - カスタマーサポートを題材に4種の参照アーキテクチャを提示：①Bedrock AgentCore + Strands（Supervisor Agentが専門SubAgentへルーティング）②Bedrock Multi-Agent Collaboration（Bedrockネイティブ機能でSupervisorが委譲・集約）③Agent Squad（Bedrock Classifierによる意図分類→専門エージェント／人間へメッセージキュー経由でルーティング）④LangGraph（ECS上のSupervisor Agentが4つの専門SubAgentを調整）。いずれもSupervisor中心の設計で、認証・Observability・会話メモリまで含めたリファレンス実装

### 総合ガイド・タクソノミー

- **[Multi-Agent Architecture Guide](https://www.openlayer.com/blog/post/multi-agent-system-architecture-guide)** — Openlayer、2026/3/9
  - **Supervisor**（中央集権・単一コーディネーターがボトルネックになりうる）／**Hierarchical**（多層の監督構造。レイテンシ増加とのトレードオフ）／**Peer-to-Peer**（中央管理なし、合意形成。調整コストがエージェント数に対して二次関数的に増大：10体で45経路）を比較
- **[Agent Architecture Patterns: 2026 Taxonomy Guide](https://www.digitalapplied.com/blog/agent-architecture-patterns-taxonomy-2026)** — Digital Applied、2026/4/30
  - Single-Agent／協調型Multi-Agent／競争型Multi-Agent／オーケストレーション・トポロジーの4象限×8パターンに整理：①Single-Agent＝ReAct・Reflexion ②協調型＝Plan-and-execute・Supervisor-worker ③競争型＝Multi-agent debate・Verifier-critic ④トポロジー＝Graph orchestration・Swarm/blackboard
- **[Multi-Agent Orchestration: 5 Patterns That Work in 2026](https://www.digitalapplied.com/blog/multi-agent-orchestration-5-patterns-that-work)** — Digital Applied、2026/5/17
  - Fan-Out（並列scatter-gather）／Pipeline（逐次連鎖）／Debate（同一プロンプトに複数エージェントが独立回答→相互批評、コストは単一モデル比約2.5倍）／**Supervisor（「2026年のプロダクションデフォルト」**）／Swarm（動的ピア、共有メモリで協調。Kimi K2.6で300エージェントまでスケール事例）
- **[6 Multi-Agent Orchestration Patterns for Production](https://beam.ai/agentic-insights/multi-agent-orchestration-patterns-production)** — Beam.ai、2026/4/15（2026/9/10更新）
  - Orchestrator-Worker／Sequential Pipeline／Fan-out-Fan-in／Multi-Agent Debate／Dynamic Handoff／Adaptive Planningの6パターン
- **[Agentic Design Patterns: The 2026 Guide to Building Autonomous Systems](https://www.sitepoint.com/the-definitive-guide-to-agentic-design-patterns-in-2026/)** — SitePoint（本文は未取得、タイトルのみ確認）
- **[What Are Agentic Design Patterns? 2026 Pattern Catalog](https://www.augmentcode.com/guides/agentic-design-patterns)** — Augment Code、2026/5/18公開（2026/6/18更新）
  - Andrew Ngの4パターン（Reflection／Tool Use／Planning／Multi-Agent Collaboration）と、Anthropicの5つのワークフローパターン（Prompt Chaining／Routing／Parallelization／Orchestrator-Workers／Evaluator-Optimizer）を統合した12パターンのタクソノミーとして整理
- **[Agent Platforms Architecture — 2026 Patterns](https://vdf.ai/blog/enterprise-ai-agent-platform-architecture-patterns-2026/)** — vdf.ai、2026/6/5
  - エンタープライズ向けに7つの基盤パターンを提示：Orchestrator-Worker分解／Supervisor・Routerエージェント／RAGグラウンディング／Model Gateway（モデルルーティングとポリシー一元化）／Human-in-the-Loop承認ゲート／評価・フィードバックループ／Observability・監査プレーン。「個々に機能するのではなく1つの統制プレーンとして積み重なる」設計を強調
- **[Multi-Agent Orchestration Patterns Explained](https://levelop.dev/blog/multi-agent-orchestration-patterns-2026)** — Levelop、2026/8/21
  - Sequential／Parallel／Hierarchicalの3方式を基本に、オーケストレーターの4責務（ルーティング／コンテキスト管理／集約／終了判定）を整理。「オーケストレーションパターンは実装の詳細ではなくアーキテクチャ上の意思決定」と強調

### 実践知見・プロダクション事例

- **[How Anthropic Built Multi-Agent Deep Research](https://theaiengineer.substack.com/p/how-anthropic-built-multi-agent-deep)** — The AI Engineer (Paolo Perrone)、2026/5/23
  - Claude Research の Orchestrator-Worker構成を解説：LeadResearcher（Opus、拡張思考でプラン策定・外部メモリに書き出し）→ 並列Subagent（Sonnet 3〜5体、相互不可視で独立実行）→ 統合フェーズ → CitationAgentによる引用検証。Single-Agent比+90.2%改善だがトークン消費は約15倍。**幅優先で独立並列に分解できる高価値タスク**に向く。共有状態やリアルタイム協調が必要なタスクには不向き
- **[Multi-Agent in Production in 2026: What Actually Survived](https://medium.com/@Micheal-Lanham/multi-agent-in-production-in-2026-what-actually-survived-f86de8bb1cd1)** — Medium (Micheal Lanham)
  - 「生き残った2026年の協調システムは例外なくフェーズゲート・共有アーティファクト・最終スーパーバイザーのいずれかを持つ」という指摘
- **[Claude Code Multi-Agent Orchestration: 6 Patterns (2026)](https://thepromptshelf.dev/blog/claude-code-multi-agent-orchestration-patterns-2026/)** — The Prompt Shelf、2026/6/5
  - Claude Code上での実践6パターン：Orchestrator-Worker／Parallel Reviewers（同一成果物を複数視点で並列レビュー）／Pipeline Chain／Debate and Convergence／Fan-Out-Fan-In／Self-Healing Pipeline（バリデータ付きリトライ）。コスト試算やAGENTS.md設定例も掲載
- **[Multi-Agent AI Systems in Production: The Architecture Patterns That Actually Work at Scale](https://www.aimagicx.com/blog/multi-agent-ai-production-architecture-patterns-2026)** — AI Magicx、2026/4/12
  - 4つのコアパターン（Hierarchical＝Orchestrator-Worker／Peer-to-Peer／Pipeline／Event-driven）とMCP・A2A・ACPのプロトコル比較、無限委譲ループ・合意形成デッドロック・コンテキスト汚染・コスト爆発など本番特有の失敗モードと対策を整理。「シングルエージェントAIは2025年にピークを迎えた」と主張

### 留意点（複数記事で共通して指摘）

- **タスク形状で向き不向きが分かれる**：[Getting Up to Speed on Multi-Agent Systems, Part 7: Benchmarks and What They Miss](https://christophermeiklejohn.com/ai/agents/mas-series/2026/04/30/mas-series-07-benchmarks.html)（Christopher Meiklejohn、2026/4/30）によれば、Multi-Agentは幅優先探索（BrowseComp：+90%改善）、複数ステップの高難度推論（GAIA Level 3：2倍改善）、状態共有を伴う制約付きプランニング（TravelPlanner：3.3倍改善）、独立並列サブタスクで優位。一方、集中的なコーディングタスクや単純な関数生成、統合が必要な分散推論ではSingle-Agentが上回る。また「Single-Agent向けベンチマーク（HumanEval・SWE-bench・WebArena等）でMulti-Agentを測ると、測定対象自体が誤っている」と指摘し、標準化されたMulti-Agent評価基準の欠如を問題視
- **オーバーエンジニアリングへの警鐘**：「まずシンプルな構成から始め、実測で複雑化の根拠を得てから拡張する」という設計哲学はAnthropic/Microsoft/OpenAI/Google共通
- **コストとのトレードオフ**：リサーチ型オーケストレーションはチャット比で約15倍のトークンを消費するため、コスト構造が見合わない用途では不採用が妥当
- **信頼性の担保**：本番で生き残る構成は共通してフェーズゲート／共有アーティファクト／最終スーパーバイザーのいずれかを備える
- **プロトコル標準化**：Google A2A（Agent2Agent）、Anthropic MCP（Model Context Protocol）が相互運用の標準として定着しつつある（MCPは月間SDKダウンロード9,700万件規模との報告）
