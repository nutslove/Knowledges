# Metric Queries (`count_over_time`等) の処理フロー

> **⚠️ 注意**: Claude Codeによる調査結果なので、自分でも追ってみること！

## 処理フロー
- LogQLクエリー: `count_over_time({app="foo"}[5m])`

### 1. パース
**場所**: `pkg/logql/syntax/parser.go:71`

- `ParseExpr()` でAST構築

```go
RangeAggregationExpr {
    Operation: "count_over_time"
    Left: LogRangeExpr {
        Selector: {app="foo"}
        Interval: 5m
    }
}
```

### 2. エンジン
**場所**: `pkg/logql/engine.go`

- `query.Exec()` → 実行コンテキスト設定
- `evalSample()` → StepEvaluator作成

### 3. Evaluator
**場所**: `pkg/logql/evaluator.go:328`

- `NewStepEvaluator()` を呼び出し
- RangeAggregationExprの場合、次のステップへ

### 4. データ取得
**場所**: `pkg/querier/querier.go:219`

`SelectSamples()` による並列データ取得:
- **Ingesterクエリ**: 最新データを取得
- **Storeクエリ**: 履歴データを取得
- **マージ**: イテレータを統合

### 5. Range集約
**場所**: `pkg/logql/range_vector.go`

- `newRangeAggEvaluator()` でRangeVectorEvaluatorを作成

各タイムステップで以下を実行:
1. `[step-interval, step]` の範囲のサンプルをロード
2. `countOverTime()` 集約を適用 (line 367)
3. `Sample(metric, timestamp, count)` を返す

### 6. 結果組み立て
**場所**: `pkg/logql/engine.go:546`

`JoinSampleVector()` による結果の構築:
1. 全ステップのサンプルを収集
2. メトリックラベルでグループ化
3. `promql.Matrix` (ソート済み時系列)を返す

---

# サポートされるRange集約操作
- `pkg/logql/syntax/ast.go` (lines 1236-1250)で定義:
  - `count_over_time`, `rate`, `bytes_over_time`
  - `avg_over_time`, `sum_over_time`, `min_over_time`, `max_over_time`
  - `quantile_over_time`, `first_over_time`, `last_over_time`
  - その他