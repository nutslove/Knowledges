# PromQL 二項演算における `on()` 句の必要性

## 結論

**両側のラベルセットが完全に一致していれば `on()` は不要。一致しない場合は必要。**

---

## PromQL 二項演算のデフォルト動作

PromQLの二項演算（`/`, `*`, `+`, `-` など）は、デフォルトで**すべてのラベルが一致**する要素同士でマッチングする。

---

## `on()` が必要なケース

```promql
sum by (node) (kube_pod_container_resource_requests{resource="cpu"})
/ on (node)
kube_node_status_allocatable{resource="cpu"}
```

### 理由

左側と右側でラベルセットが異なる：

| 側 | ラベル |
|---|---|
| 左側（`sum by` で集約済み） | `{node="node-1"}` |
| 右側（生のメトリクス） | `{node="node-1", resource="cpu", unit="core", instance="...", job="..."}` |

`on (node)` を指定しないと、ラベルセット全体が一致しないためマッチングが失敗し、**結果が空**になる。

---

## `on()` が不要なケース

両側を同様に集約すれば `on()` は不要：

```promql
sum by (node) (kube_pod_container_resource_requests{resource="cpu"})
/
sum by (node) (kube_node_status_allocatable{resource="cpu"})
```

両側とも `{node="..."}` だけになるので、`on()` なしでも正しくマッチする。

---

## ベストプラクティス

- 両側を明示的に `sum by` で集約する書き方の方が意図が明確で可読性が高い
- `on()` を使う場合は、どのラベルでマッチングしているか明示的になるメリットがある