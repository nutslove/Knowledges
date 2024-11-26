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

## ControllerのReconcileループのロジックを実装
- **`internal/controller/<--kindで指定した名前>_controller.go`ファイルにControllerのReconcileループのロジックを実装する**

## `make manifests`コマンドでCRDのマニフェストファイルを作成
- `config/crd/bases/`ディレクトリが作成され、その中にサンプルのマニフェストファイルが作成される
- `config/rbac/role.yaml`ファイルが作成される
- **`make manifests`コマンドは`api/<--versionで指定したバージョン>/<--kindで指定した名前>_types.go`の`Spec`と`Status`の構造体(`struct`)の部分を見てマニフェストファイルを生成する**

