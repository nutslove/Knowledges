- https://langfuse.com/docs/evaluation/evaluation-methods/llm-as-a-judge

## LanfuseでのLLM as a Judge

### TraceとObservationの関係
- **Trace** = Agentの1回の実行全体（最上位の単位）
- **Observation** = Trace内の個別のステップ（LLM呼び出し、Tool呼び出し、Retrieval等）
- 1つのTraceの中に複数のObservationが含まれる親子関係
```
  Trace（Agent実行全体）
  ├── Observation: LLM呼び出し（推論・計画）
  ├── Observation: Tool呼び出し（検索API）
  ├── Observation: Retrieval（RAGのドキュメント取得）
  └── Observation: LLM呼び出し（最終回答生成）
```

### Evaluatorのタイプ
- Observations、Traces、Experimentsの3種類のタイプがある
  > LLM-as-a-Judge evaluators can run on three types of data: **Observations** (individual operations), **Traces** (complete workflows), or **Experiments** (controlled test datasets).
- TracesはLegacy（非推奨となる予定）で、Observationsが推奨されている

### Observations （個別のオペレーション）
- Trace内の特定のObservation（ステップ）に対して評価を行う
- **Liveデータのみ対応**：Observationがingest（取り込み）されたタイミングで自動的に評価が実行される仕組み。既存のトレースに対して後から遡って評価を実行すること（バックフィル）は現時点（2026/03）で非対応
- **フィルタで評価対象を絞り込む**：observation type、trace name、tags、userId、metadata等の条件を組み合わせて、どのObservationを評価するかを制御する。例えば「最終回答のLLM呼び出しだけ」や「特定タグのトレース内のgeneration stepだけ」といった指定が可能
- **Sampling設定**：全件ではなくX%だけ評価する設定ができ、コスト管理に有効

> [!CAUTION]  
> Samplingの設定は、Trace単位ではなく、Observation単位で行われる。  
> つまり、Sampling率が50%に設定されている場合、2つのTrace（２回の独立したAgentの実行）があったとして、片方のTraceのすべてのObservationが評価され、もう片方のTraceのすべてのObservationが評価されない、ということはなく、両方のTrace内のObservationが（Samplingに設定した確率に基づいて）ランダムに選ばれて評価されることになる。

### Experiments （オフライン）
- データセットに対してモデルやプロンプトのバリエーションを比較評価


### SDKからのスコア取得 
Langfuseに保存されたスコア（LLM-as-a-Judge、API経由、手動アノテーション等）はPython SDKで取得可能。
評価結果をThanosやLokiにメトリクス・ログとして転送する用途などに使える。

#### スコア取得方法
##### 方法1: score_v_2 APIでフィルタして取得
```python
from langfuse import Langfuse

langfuse = Langfuse()

# trace_idとスコア名でフィルタ
response = langfuse.api.score_v_2.get(trace_id="<trace_id>", name="Helpfulness")
for score in response.data:
    print(score.value)      # スコア値（例: 0.95）
    print(score.comment)    # 評価理由
    print(score.source)     # ScoreSource.EVAL / ScoreSource.API / ScoreSource.ANNOTATION
    print(score.data_type)  # NUMERIC / CATEGORICAL / BOOLEAN
    print(score.timestamp)
```

##### 方法2: トレースごと取得（全スコア埋め込み）
```python
trace = langfuse.api.trace.get(trace_id="<trace_id>")
for s in trace.scores:
    print(s.name, s.value, s.comment)
```

##### フィルタパラメータ（`api.score_v_2.get()`）

| パラメータ | 説明 |
|------------|------|
| `trace_id` | 特定トレースのスコアのみ取得 |
| `name` | スコア名でフィルタ（例: `"Helpfulness"`） |
| `source` | スコアソース（`EVAL`: Managed Evaluator, `API`: SDK経由, `ANNOTATION`: UI手動） |
| `data_type` | `NUMERIC` / `CATEGORICAL` / `BOOLEAN` |
| `from_timestamp` / `to_timestamp` | 期間指定 |
| `trace_tags` | トレースのタグでフィルタ |
| `page` / `limit` | ページネーション |