## 概要

ArgoCD はデフォルトで `kubectl apply`（ClientSide Apply）を使って Git のマニフェストをクラスタに適用する。ClientSide Apply ではローカル（ArgoCD 側）で 3-way merge を計算し、そのパッチを API Server に送る方式。

Server-Side Apply (SSA) を有効にすると、ArgoCD は内部的に `kubectl apply --server-side --force-conflicts` 相当のコマンドを発行する。差分計算とマージが **API Server 側で実行** される。

- Kubernetes の SSA は v1.22 で GA（managed fields のフルサポートは v1.18 から beta）
- ArgoCD は SSA 時、field manager として `argocd-controller` で登録される
- `--force-conflicts` が常に付くため、他の manager が保有していた field も強制的に奪う

## ClientSide Apply との違い（なぜ SSA を使うのか）

### 1. `last-applied-configuration` annotation の 262144 bytes 制限の回避
ClientSide Apply は前回の状態を `kubectl.kubernetes.io/last-applied-configuration` annotation に保存する。大きな CRD（Prometheus Operator, Istio, Cilium など）ではこれが 262144 bytes を超えて、以下のエラーが発生：

```
metadata.annotations: Too long: must have at most 262144 bytes
```

SSA はこの annotation を使わず、API Server 側の `managedFields` で状態管理するので制限を回避できる。

> [!NOTE] 
>
> ArgoCD 本体のインストール・アップグレードでも、一部 CRD が 262144 bytes を超えるため、公式ドキュメント上 `kubectl apply --server-side --force-conflicts` で install.yaml / ha/install.yaml を適用することが推奨されている。

### 2. Field Ownership（複数コントローラとの共存）
SSA は field 単位でオーナーを追跡する（`managedFields`）。ArgoCD が `spec` を、Operator が `status` や一部の `spec` を管理、といった共同管理が可能になる。ClientSide Apply だと last writer wins で上書き合戦になる。

### 3. Admission Webhook を Diff 段階で効かせられる（Server-Side Diff と組み合わせ）
ValidatingWebhook / MutatingWebhook はサーバサイドでのみ実行される。SSA（特に Server-Side Diff と組み合わせ）を使うと、Sync の前段階で webhook の検証結果を得られる。

### 4. 部分 YAML でのパッチ適用
リソース全体を書かずに、変えたいフィールドだけを書いた YAML で patch できる。例：Deployment の replicas だけを ArgoCD で管理する。ただし schema 違反になるので `Validate=false` を併用する必要がある（後述）。

## 有効化方法

### Application レベル

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  syncPolicy:
    syncOptions:
      - ServerSideApply=true
```

### リソースレベル（annotation）

```yaml
metadata:
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=true
```

Application レベルで有効化していない場合に、特定リソース（巨大な CRD など）だけ SSA にしたいときに使う。

### リソース単位での無効化

Application レベルで `ServerSideApply=true` になっているが、特定リソースだけ ClientSide に戻したい場合：

```yaml
metadata:
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=false
```

> 過去に Issue [#20306](https://github.com/argoproj/argo-cd/issues/20306)（2024/10）で一部バージョンでは動作しない不具合が報告されていた。現行の公式ドキュメントには正式機能として記載されているが、使用時は自環境での挙動確認を推奨。

### 部分マニフェスト適用の注意点

replicas だけを書いた Deployment のような部分マニフェストは、スキーマ検証に通らない。そのため `Validate=false` の併用が必要：

```yaml
spec:
  syncPolicy:
    syncOptions:
      - ServerSideApply=true
      - Validate=false
```

この場合 ArgoCD が発行するコマンドは `kubectl apply --server-side --force-conflicts --validate=false` となる。

## Replace との関係（重要）

**`Replace=true` は `ServerSideApply=true` より優先される**（公式ドキュメント明記）。両方指定すると `Replace=true` が効いて、`kubectl delete/create` による破壊的な再作成が行われる。両者は同時指定しない。

## Client-Side Apply Migration

既存で ClientSide Apply 管理だったリソースを SSA に切り替える際、`managedFields` のオーナーシップを ArgoCD (`argocd-controller`) に移行する機能。**デフォルトで有効**。

動作：
1. 指定の field manager（指定されない場合はデフォルトの `kubectl-client-side-apply`）が `managedFields` に存在するかチェック
2. その manager の field ownership を ArgoCD の SSA manager (`argocd-controller`) に付け替え
3. 古い manager エントリを除去
4. SSA を実行

### 無効化

```yaml
spec:
  syncPolicy:
    syncOptions:
      - ClientSideApplyMigration=false
```

### カスタム field manager の指定

別の operator が管理していたフィールドのオーナーシップを ArgoCD に移したいとき：

```yaml
metadata:
  annotations:
    argocd.argoproj.io/client-side-apply-migration-manager: 'my-custom-manager'
```

この manager の操作種別が `Update`（ClientSide Apply の印）のものを、ArgoCD の SSA manager に付け替える。

### ArgoCD v3.3 での注意（自身を Application で管理している場合）

- v3.3 で自身を Application で管理している場合、**`ServerSideApply=true` が必須**
- **v3.3.0 / v3.3.1** には ClientSideApplyMigration に起因する不具合があり、一時回避策として `ClientSideApplyMigration=false` の設定が推奨されていた（[Issue #26279](https://github.com/argoproj/argo-cd/issues/26279) 参照）
- **v3.3.2 で修正済み**。回避策として `ClientSideApplyMigration=false` を設定していた場合は **削除する必要がある**（そのままにしておくと将来 field manager の衝突を引き起こすリスク）

## 関連：Server-Side Diff (SSD)

SSA と **別機能** だが関連が深い。v2.10 で導入。Diff 計算のときに SSA の dry-run を API Server で実行し、その結果をライブ状態と比較する方式。

※「`ServerSideApply=true` で SSD も自動有効化される」と説明している二次情報もあるが、公式ドキュメントでは SSA と SSD は別オプションとして記載されているので、SSD は明示的に有効化する必要がある。

### 有効化

Application レベル（annotation）：
```yaml
metadata:
  annotations:
    argocd.argoproj.io/compare-options: ServerSideDiff=true
```

全体（`argocd-cmd-params-cm` ConfigMap）：
```yaml
data:
  controller.diff.server.side: "true"
```

### Mutating Webhook を Diff に含める

デフォルトでは mutating webhook の変更は diff に入らない。含めるには：

```yaml
metadata:
  annotations:
    argocd.argoproj.io/compare-options: ServerSideDiff=true,IncludeMutationWebhook=true
```

### 全体有効時にリソース単位で無効化

```yaml
metadata:
  annotations:
    argocd.argoproj.io/compare-options: ServerSideDiff=false
```

### 注意点
- **新規リソース作成時には SSD は実行されない**（ライブ状態がないため）。そのタイミングでは webhook 検証が diff 段階では走らない
- CRD がデフォルト値を定義している場合の false OutOfSync 問題は SSD で解消されるケースが多い
- Webhook が dry-run に対応していないと SSD が失敗する（コントローラログに `dry-run` 関連のエラーが出る）

## 既知の問題・トラブル事例

- **SSA 有効化後の diff 計算エラー**: [Issue #17358](https://github.com/argoproj/argo-cd/issues/17358) — 一部の CRD で `error calculating structured merge diff: ... field not declared in schema` が発生するケースがある
- **SSA + SSD + Kyverno 等の webhook 併用で FATA エラー**: [Issue #22562](https://github.com/argoproj/argo-cd/issues/22562) — webhook がリクエストを拒否すると diff 計算が落ちる
- **Rollback 時に syncOptions が消える**: [Issue #20183](https://github.com/argoproj/argo-cd/issues/20183) — automated sync 有効時のロールバックで `ServerSideApply=true` が失われることがある

## 使いどころの指針（個人的整理）

| ケース | 推奨 |
|---|---|
| 巨大 CRD（Prometheus Operator, Istio, Cilium, kube-prometheus-stack など） | SSA 必須 |
| Operator と共存するリソース（CR の status を operator が書く） | SSA |
| 部分管理したいリソース（replicas だけ管理など） | SSA + `Validate=false` |
| Webhook が多くて diff 精度を上げたい | SSA + SSD |
| シンプルなマニフェスト、単独管理 | ClientSide でも問題なし |

## 実運用の注意

- Application 全体で SSA を有効化すると、想定外のリソースにも影響する。まず annotation で特定 CRD だけ有効化 → 問題ないことを確認 → Application レベルに広げる、という段階的導入が無難
- `managedFields` がリソースに蓄積される。`kubectl get <resource> -o yaml` で見るとかなりの量になるので、デバッグ時は `--show-managed-fields=false` を使う
- 既存環境から切り替える際は `ClientSideApplyMigration` がデフォルト有効なので基本そのままで良いが、別 operator が握っていた field を引き取りたいなら `client-side-apply-migration-manager` annotation を活用
- Conflict 発生時：ArgoCD は `--force-conflicts` を常用するので、他の manager が持っていた field も強制的に奪う。意図しない結果になる可能性があるため、事前に field ownership を整理する
- `Replace=true` は SSA より優先されるため、混在させない

## 参考リンク

- [Sync Options - ArgoCD 公式ドキュメント](https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/)
- [Diff Strategies - ArgoCD 公式ドキュメント](https://argo-cd.readthedocs.io/en/stable/user-guide/diff-strategies/)
- [Server-Side Apply Proposal](https://argo-cd.readthedocs.io/en/stable/proposals/server-side-apply/)
- [v3.2 to v3.3 Upgrade Notes](https://argo-cd.readthedocs.io/en/stable/operator-manual/upgrading/3.2-3.3/)
- [Kubernetes Server-Side Apply 公式](https://kubernetes.io/docs/reference/using-api/server-side-apply/)
- [KEP-555 Server-Side Apply](https://github.com/kubernetes/enhancements/tree/master/keps/sig-api-machinery/555-server-side-apply)