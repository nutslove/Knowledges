- https://langfuse.com/docs/evaluation/evaluation-methods/llm-as-a-judge

## LanfuseでのLLM as a Judge
- Observations、Traces、Experimentsの3種類のタイプがある
  > LLM-as-a-Judge evaluators can run on three types of data: **Observations** (individual operations), **Traces** (complete workflows), or **Experiments** (controlled test datasets).
- TracesはLegacy（非推奨となっている）で、Observationsが推奨されている

### Observations　（個別のオペレーション）
- LLM呼び出し、Retrieval、ツールコールなど個別のステップに対して評価を行う

### Experiments （オフライン）
- データセットに対してモデルやプロンプトのバリエーションを比較評価