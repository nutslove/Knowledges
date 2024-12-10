# 概要
- Operator Frameworkを構成する以下３つの中の１つ
  - Operator SDK
  - OperatorLifecycleManager（OLM）
  - OperatorHub
- Kubebuilderをベースに拡張

# Install
- https://sdk.operatorframework.io/docs/installation/

# Operator SDK設定
## `operator-sdk init`コマンドによる初期化（Project作成）
- 書式  
  ```
  operator-sdk init --domain <API groupドメイン> --repo <name to use for go module (e.g. github.com/user/repo), defaults to the go package of the current working directory> 
  ```
  - The `--repo=<path>` flag is required when creating a project outside of `$GOPATH/src`
  - 以下のようなディレクトリとファイルが作成される  
    ```
    .
    |-- cmd
    |   `-- main.go
    |-- config
    |   |-- default
    |   |   |-- kustomization.yaml
    |   |   |-- manager_metrics_patch.yaml
    |   |   `-- metrics_service.yaml
    |   |-- manager
    |   |   |-- kustomization.yaml
    |   |   `-- manager.yaml
    |   |-- manifests
    |   |   `-- kustomization.yaml
    |   |-- prometheus
    |   |   |-- kustomization.yaml
    |   |   `-- monitor.yaml
    |   |-- rbac
    |   |   |-- kustomization.yaml
    |   |   |-- leader_election_role_binding.yaml
    |   |   |-- leader_election_role.yaml
    |   |   |-- metrics_auth_role_binding.yaml
    |   |   |-- metrics_auth_role.yaml
    |   |   |-- metrics_reader_role.yaml
    |   |   |-- role_binding.yaml
    |   |   |-- role.yaml
    |   |   `-- service_account.yaml
    |   `-- scorecard
    |       |-- bases
    |       |   `-- config.yaml
    |       |-- kustomization.yaml
    |       `-- patches
    |           |-- basic.config.yaml
    |           `-- olm.config.yaml
    |-- Dockerfile
    |-- go.mod
    |-- go.sum
    |-- hack
    |   `-- boilerplate.go.txt
    |-- Makefile
    |-- PROJECT
    |-- README.md
    `-- test
        |-- e2e
        |   |-- e2e_suite_test.go
        |   `-- e2e_test.go
        `-- utils
            `-- utils.go
    ```

## `operator-sdk api`コマンドでAPIとControllerを作成
- 書式  
  ```
  operator-sdk api --group <API Group名(e.g. apps/v1のappsの部分)> --version <API Version> --kind <リソースのKind> --resource --controller
  ```  
  - `--resource`フラグはCRDのコードを生成、`--controller`フラグはコントローラーのコードを生成する
  - 追加で`api`、`bin`、`internal`ディレクトリが作成される  
    ```
    .
    |-- api
    |   `-- v1alpha1
    |       |-- groupversion_info.go
    |       |-- memcached_types.go
    |       `-- zz_generated.deepcopy.go
    |-- bin
    |   |-- controller-gen -> /root/k8s-operator/bin/controller-gen-v0.15.0
    |   `-- controller-gen-v0.15.0
    |-- cmd
    |   `-- main.go
    |-- config
    |   |-- crd
    |   |   |-- kustomization.yaml
    |   |   `-- kustomizeconfig.yaml
    |   |-- default
    |   |   |-- kustomization.yaml
    |   |   |-- manager_metrics_patch.yaml
    |   |   `-- metrics_service.yaml
    |   |-- manager
    |   |   |-- kustomization.yaml
    |   |   `-- manager.yaml
    |   |-- manifests
    |   |   `-- kustomization.yaml
    |   |-- prometheus
    |   |   |-- kustomization.yaml
    |   |   `-- monitor.yaml
    |   |-- rbac
    |   |   |-- kustomization.yaml
    |   |   |-- leader_election_role_binding.yaml
    |   |   |-- leader_election_role.yaml
    |   |   |-- memcached_editor_role.yaml
    |   |   |-- memcached_viewer_role.yaml
    |   |   |-- metrics_auth_role_binding.yaml
    |   |   |-- metrics_auth_role.yaml
    |   |   |-- metrics_reader_role.yaml
    |   |   |-- role_binding.yaml
    |   |   |-- role.yaml
    |   |   `-- service_account.yaml
    |   |-- samples
    |   |   |-- cache_v1alpha1_memcached.yaml
    |   |   `-- kustomization.yaml
    |   `-- scorecard
    |       |-- bases
    |       |   `-- config.yaml
    |       |-- kustomization.yaml
    |       `-- patches
    |           |-- basic.config.yaml
    |           `-- olm.config.yaml
    |-- Dockerfile
    |-- go.mod
    |-- go.sum
    |-- hack
    |   `-- boilerplate.go.txt
    |-- internal
    |   `-- controller
    |       |-- memcached_controller.go
    |       |-- memcached_controller_test.go
    |       `-- suite_test.go
    |-- Makefile
    |-- PROJECT
    |-- README.md
    `-- test
        |-- e2e
        |   |-- e2e_suite_test.go
        |   `-- e2e_test.go
        `-- utils
            `-- utils.go
    ```

## CRDの定義
- **`api/<--versionで指定したバージョン>/<--kindで指定した名前>_types.go`の`Spec`と`Status`の構造体(`struct`)の部分を修正してCRDのSpecやStatusを定義する**
  - **`Spec`や`Status`に変更を加えたら、必ず`make manifests`を実行してマニフェストを更新すること**
- **`*_types.go`を修正した後、`make generate`コマンドを実行して`api/<--versionで指定したバージョン>/zz_generated.deepcopy.go`をUpdateすること**

## ControllerのReconcileループのロジックを実装
- **`internal/controller/<--kindで指定した名前>_controller.go`ファイルの`Reconcile`メソッド内にControllerのReconcileループのロジックを実装する**
  - `Reconcile`メソッドは、コントローラランタイム（controller-runtime）のコントローラループから呼び出される
- **デフォルトでは、`internal/controller/*_controller.go`の`SetupWithManager`メソッド内の`For`で指定されたカスタムリソースを監視するため、`Reconcile`メソッドは対象のカスタムリソースの変更（作成、更新、削除）の時に呼び出される**
  - **`SetupWithManager`メソッド内で、`Owns`や`Watches`メソッドを使用して他のリソースを監視対象に追加することもできる**
- `Reconcile`メソッドが返す`error`の中身によって`Reconcile`メソッドが繰り返して呼び出されるか１回で終わるかが決まる。  
  返す`error`が`nil`の場合は１回で終わる。
  - `return ctrl.Result{}, err`とすることで、再度`Reconcile`が呼び出される
  - `ctrl.Result{}`を返すと再試行されない
    - `error`型のZero Valueが`nil`のため
  - `ctrl.Result{Requeue: true}`を返すと、すぐに再試行される
  - `ctrl.Result{RequeueAfter: time.Duration}`を返すと、指定した時間後に再試行される
- **つまり、`internal/controller/*_controller.go`の、１．`SetupWithManager`メソッド内の`For`や`Owns`や`Watches`メソッドで指定されたリソースが更新された時、２．`Reconcile`メソッド内で`return ctrl.Result{}, err`や`ctrl.Result{Requeue: true}`などで返された時に、`Reconcile`が実行される**

### `Reconcile`メソッドの基本的な構造例
```go
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. カスタムリソースのインスタンスを取得
    var myApp myappv1.MyApp
    if err := r.Get(ctx, req.NamespacedName, &myApp); err != nil {
        if errors.IsNotFound(err) {
            // リソースが削除された場合の処理
            return ctrl.Result{}, nil
        }
        // 取得時にエラーが発生した場合
        return ctrl.Result{}, err
    }

    // 2. 関連するリソースの存在を確認（例: Deployment）
    var deployment appsv1.Deployment
    err := r.Get(ctx, types.NamespacedName{Name: myApp.Name, Namespace: myApp.Namespace}, &deployment)
    if err != nil && errors.IsNotFound(err) {
        // Deploymentが存在しない場合、作成する
        deployment = *r.constructDeployment(&myApp)
        if err := r.Create(ctx, &deployment); err != nil {
            return ctrl.Result{}, err
        }
    } else if err != nil {
        // その他のエラー処理
        return ctrl.Result{}, err
    }

    // 3. リソースの更新が必要か確認し、更新する
    if !reflect.DeepEqual(deployment.Spec.Replicas, myApp.Spec.Replicas) {
        deployment.Spec.Replicas = &myApp.Spec.Replicas
        if err := r.Update(ctx, &deployment); err != nil {
            return ctrl.Result{}, err
        }
    }

    // 4. ステータスの更新 `r.Status().Update(ctx, &myApp)`
    myApp.Status.AvailableReplicas = deployment.Status.AvailableReplicas
    if err := r.Status().Update(ctx, &myApp); err != nil {
        return ctrl.Result{}, err
    }

    // 5. 正常終了
    return ctrl.Result{}, nil
}
```

### `SetupWithManager`メソッドの設定例
#### `For`、`Owns`、`Watches`メソッド
- `For`メソッド
  - コントローラーが主要なリソースとして扱うカスタムリソースを指定
- `Owns`メソッド
  - コントローラーが作成・管理するリソースを指定
- `Watches`メソッド
  - 任意のリソースを監視対象として追加
  - イベントハンドラーを使用して、どのようにReconcileをトリガーするかを制御
- 参照URL
  - https://yash-kukreja-98.medium.com/develop-on-kubernetes-series-demystifying-the-for-vs-owns-vs-watches-controller-builders-in-c11ab32a046e

#### `For`、`Owns`メソッドの例
```go
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&myappv1.MyApp{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Complete(r)
}
```
- `For(&myappv1.MyApp{})`
  - `MyApp`リソースの変更を監視
- `Owns(&appsv1.Deployment{})`
  - コントローラーが所有する`Deployment`リソースの変更を監視
- `Owns(&corev1.Service{})`
  - コントローラーが所有する`Service`リソースの変更を監視

#### `Watches`メソッドで他のリソースをカスタムウォッチする例
```go
import "sigs.k8s.io/controller-runtime/pkg/handler"

func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&myappv1.MyApp{}).
        Watches(
            &source.Kind{Type: &corev1.ConfigMap{}},
            &handler.EnqueueRequestForOwner{
                OwnerType:    &myappv1.MyApp{},
                IsController: true,
            },
        ).
        Complete(r)
}
```
- `Watches`メソッドを使用して、`ConfigMap`の変更を監視し、`MyApp`の`Reconcile`をトリガーする


## `make manifests`コマンドでCRDのマニフェストファイルを作成
- `config/crd/bases/`ディレクトリが作成され、その中にサンプルのマニフェストファイルが作成される
- `config/rbac/role.yaml`ファイルが作成される
- **`make manifests`コマンドは`api/<--versionで指定したバージョン>/<--kindで指定した名前>_types.go`の`Spec`と`Status`の構造体(`struct`)の部分を見てマニフェストファイルを生成する**

## Operatorの実行
- https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/#run-the-operator
- 開発(PoC)の目的でKubernetesクラスター外でOperatorを実行する時は`make run`（`make install run`）
- `Deployment`としてKubernetesクラスター内で実行する場合
  - 以下でOperatorのイメージをビルドして、コンテナレジストリ（Docker HubやECRなど）にプッシュ  
    ```shell
    make docker-build IMG=<your-registry>/<your-operator>:<tag>
    make docker-push IMG=<your-registry>/<your-operator>:<tag>
    ```
  - その後`make deploy`でOperatorをデプロイ  
    ```shell
    make deploy IMG=<your-registry>/<your-operator>:<tag>
    ```

## Operatorのテスト
### `suite_test.go`
- `internal/controller/suite_test.go`で`envtest`、`ginkgo`、`gomega`というテストフレームワーク/ツールでOperator(`Reconcile`)の挙動テストのための環境初期化を行う
  - `BeforeSuite`ブロックで環境初期化を、`AfterSuite`ブロックで後処理を記述
- `envtest`（`envtest.Environment`と`testEnv.Start()`）は軽量な(疑似的な)API Serverとetcdをインメモリで起動し、擬似的なKubernetes APIサーバー環境をテスト中に構築する
- `AddToScheme`関数で`Scheme`にCRDの型情報を追加することで、`k8sClient`がCRオブジェクトを適切にシリアライズ・デシリアライズできるようにする
- `TestControllers`のエントリーポイント関数でGinkgoの`RunSpecs(t, "Suite名")`を呼び出すことで、Ginkgoによるテストスイートが開始される
### `*_controller_test.go`
- **実際のテストケースは`internal/controller/*_controller_test.go`の`Describe`や`It`ブロックで記述**
- テスト実行時に`suite_test.go`でセットアップした環境下でこれらのテストが実行される
### テストの実行
- `make test`コマンドでテストを実行する


# 各ファイルについて
## `cmd/main.go`
- managerを初期化して実行する
- managerは