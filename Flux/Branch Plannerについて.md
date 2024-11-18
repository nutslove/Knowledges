## 概要
- **PR作成時にTerraformのPlanが実行され、Plan結果をGitリポジトリのPRにコメントとして通知してくれる機能**
- 参照URL
  - https://qiita.com/da-ishi10/items/51dafbd52195d90560cb
  - https://flux-iac.github.io/tofu-controller/branch-planner/
  - https://flux-iac.github.io/tofu-controller/getting_started/
- Branch PlannerはTF-Controllerの１コンポーネント
- TF-Controllerはv0.16.0-rc.2以降のバージョンを使う必要がある
- Branch Plannerを有効にしてインストールする必要がある  
  ```shell
  kubectl apply -f https://raw.githubusercontent.com/flux-iac/tofu-controller/main/docs/branch-planner/release.yaml
  ```
  - `branchPlanner.enabled`を`true`にする必要がある
    ```yaml
    branchPlanner:
      enabled: true
    ```