## install
- https://zenn.dev/zenogawa/articles/try_k8s_vault
- https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-raft-deployment-guide

- standaloneモードとHAモードがあり、standaloneモードでは単一の `vault-0` podで稼働され、データはPVに保存、永続化される

1. helmでインストール
    ```shell
    kubectl create namespace vault
    helm repo add hashicorp https://helm.releases.hashicorp.com
    helm search repo hashicorp/vault --versions
    helm install vault hashicorp/vault --namespace vault [--version <CHART VERSION>]
    ```
2. initializeとunseal  
    - helm installだけでは`vault-0` podは0/1状態となっていて、以下の手順を実施する必要がある
    - **`vault operator init`時に表示されるRoot Tokenは最初の初期化時しか分からないので、押さえておくこと！（`vault login`時に必要）**
    ```shell
    kubectl exec -it vault-0 -n vault -- vault operator init
    Unseal Key 1: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    Unseal Key 2: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    Unseal Key 3: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    Unseal Key 4: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    Unseal Key 5: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    
    kubectl exec -ti vault-0 -n vault -- vault operator unseal
    Unseal Key (will be hidden): # Unseal key 1を入力
    
    ## vaultからの出力

    kubectl exec -ti vault-0 -n vault -- vault operator unseal
    Unseal Key (will be hidden): # Unseal key 2を入力

    ## vaultからの出力

    kubectl exec -ti vault-0 -n vault -- vault operator unseal
    Unseal Key (will be hidden): # Unseal key 3を入力
    ```

## secretの追加
- `vault-0` Podにアクセス
    ```shell
    kubectl exec -it vault-0 -n vault -- sh
    ```
- ログイン後`vault` CLIでSecrets Engineを有効にする
    - `vault secrets enable -path=<有効にしたいpath> kv`
    - 以下の例だと`secret`パスが有効になり、`secret/myapp/config`などのパスにsecretを追加できる
        ```shell
        vault secrets enable -version=2 -path=secret kv
        ## Success! Enabled the kv secrets engine at: secret/
        ```
    - 無効化は`vault secrets disable <path>`
- secretを追加
    - `vault kv put <任意のsecretのパス(e.g. secret/minio/config)> <key=value [key=value]>`    
    ```shell
    vault kv put secret/myapp/config username='myuser' password='mypassword'
    ```

## その他`vault`コマンド
- Secret Engines(Vault Path)一覧確認
    ```shell
    vault secrets list
    ```
- 指定したVault Pathのsecretの値を確認
    ```shell
    vault kv get <Vault Path>
    ```
    ```shell
    / $ vault kv get secret/minio/config/
    ====== Data ======
    Key          Value
    ---          -----
    AccessKey    xxxxxxxxxxxxxxxxxxxxx
    SecretKey    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    ```


## Web UI
- `kubectl port-forward vault-0 -n vault 8200:8200`でPort Forwardingしてから、ブラウザから`localhost:8200`でVaultのGUIにアクセスできる