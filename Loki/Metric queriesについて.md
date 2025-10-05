# Metric Queries (`count_over_time`等) の処理フロー
- **Claude Codeによる調査結果なので、自分でも追ってみること！**

HTTP Request: `count_over_time({app="foo"}[5m])`
    ↓
1. パース (`pkg/logql/syntax/parser.go:71`)
   `ParseExpr()` → AST構築
    ↓  

  ```go
   RangeAggregationExpr {
       Operation: "count_over_time"
       Left: LogRangeExpr {
           Selector: {app="foo"}
           Interval: 5m
       }
   }
   ```

1. エンジン (`pkg/logql/engine.go`)
   `query.Exec()` → 実行コンテキスト設定
   `evalSample()` → StepEvaluator作成

2. Evaluator (`pkg/logql/evaluator.go:328`)
   `NewStepEvaluator()`
   → RangeAggregationExprの場合:

3. データ取得 (`pkg/querier/querier.go:219`)
   `SelectSamples()`
       ├→ Ingesterクエリ (最新データ)
       ├→ Storeクエリ (履歴データ)
       └→ イテレータをマージ

4. Range集約 (`pkg/logql/range_vector.go`)
   `newRangeAggEvaluator()`
   → RangeVectorEvaluator作成

   各タイムステップで:
       ├→ [step-interval, step]のサンプルをロード
       ├→ countOverTime()集約を適用 (line 367)
       └→ Sample(metric, timestamp, count)を返す

5. 結果組み立て (`pkg/logql/engine.go:546`)
   `JoinSampleVector()`
   → 全ステップのサンプルを収集
   → メトリックラベルでグループ化
   → `promql.Matrix` (ソート済み時系列)を返す

## サポートされるRange集約操作
- `pkg/logql/syntax/ast.go` (lines 1236-1250)で定義:
  - `count_over_time`, `rate`, `bytes_over_time`
  - `avg_over_time`, `sum_over_time`, `min_over_time`, `max_over_time`
  - `quantile_over_time`, `first_over_time`, `last_over_time`
  - その他