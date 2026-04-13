# 非機能要件
- セキュリティ: ユーザーデータの保護、アクセス制御、認証と認可の実装
- パフォーマンス: **応答時間（TTFT, 総応答時間）** の最適化、スケーラビリティ
- 可観測性: トレース取得、プロンプト・LLM応答/Toolレスポンスロギング、コスト追跡
- 安全性: Guardrails (PII、プロンプトインジェクション、禁止トピック)

# 成功基準
- 何をもって成功とするかの基準を定義する。リリースしていいかの判断基準にもなる。
- 3つの次元を組み合わせて設計する: **品質** / **レイテンシ** / **コスト**
- 基準は2種類に分ける:
  - Gating criteria: リリースの絶対条件
  - Target metrics: 継続的に追う目標指標
- 参考URL
  - https://cloud.google.com/blog/topics/developers-practitioners/a-methodical-approach-to-agent-evaluation?hl=en
  - https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents

# 評価戦略
- 評価主導開発（Evaluation-Driven Development）を採用し、先に評価基準を決めて評価パイプラインを構築する。
- 評価手法を組み合わせる: LLM-as-Judge / 人手評価 / オンラインA/B
  - LLM-as-Judgeは人手評価との整合性を定期的に検証する
- Agent特有の評価軸も含める:
  - 最終出力の正しさ (Task completion)
  - 実行過程の品質 (Trajectory / Tool use efficiency)