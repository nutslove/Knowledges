## 基本的な知識
- ServiceAccountはプラグラム(Podで実行されるプロセス)がkube-apiserverへ認証するためのもの
- ServiceAccountはNamespacedリソース
- Podに`serviceAccountName`による明示的なServiceAccountの指定がなければ、Namespace内の`default` ServiceAccountを使用する
- 参考URL
  - https://kubernetes.io/ko/docs/reference/access-authn-authz/service-accounts-admin/

## v1.24以前
- v1.23まではServiceAccountを作成すると自動的にToken(Secret)が作成された  
  ![ServiceAccount_Token](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/ServiceAccout_Token.jpg)
  - このTokenがPodがkube-apiserverへの認証の際に使われる
  - このTokenは**無期限**だった
- このTokenはTypeが`kubernetes.io/service-account-token`の`Secret`として保存される  
  ![Secret](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret.jpg)  
  <img src="https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret2.jpg" width="1000" height="300">
  <!-- ![Secret2](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret2.jpg =250x250) -->
- Podが作成される時にこのTokenがPod内の`/var/run/secrets/kubernetes.io/serviceaccount`にvolumeとして自動的にMountされる  
  ![Secret_Mount](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret_Mount.jpg)  
  - Podの中で`/var/run/secrets/kubernetes.io/serviceaccount`ディレクトリを見ると`token`がファイルとして存在していることを確認できる  
      ![Token_insidepod](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Token_InsidePod.jpg)  
  - `/var/run/secrets/kubernetes.io/serviceaccount`内の`ca.crt`はkube-apiserverが提供する証明書の検証に使われる
    - https://kubernetes.io/ko/docs/tasks/run-application/access-api-from-pod/

## v1.24以降
- v1.24からはServiceAccountを作成しても自動的にToken(Secret)が作成されなくなった  
  ![v1.24_ServiceAccount](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/v1.24_ServiceAccount.jpg) 
- ServiceAccountとは別で`kubectl create token`コマンドでTokenを作成する必要がある
  > **Note**  
  > ただ、`kubectl create token`コマンドで生成されたTokenには有効期限がある（defaultは1時間）  
  > https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#manually-create-an-api-token-for-a-serviceaccount
- **v1.23までのように無期限のTokenを作成したい場合はSecretを作成する必要がある**
  - Secretを作成するとTokenが自動的に作成されてSecretに保存される
    - https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#manually-create-a-long-lived-api-token-for-a-serviceaccount
    - https://kubernetes.io/docs/concepts/configuration/secret/
  - Secretを作成した後のServiceAccountにTokenが付いていることが分かる  
    ![v1.24_ServiceAccount_After_Secret](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/v1.24_ServiceAccount_After_Secret.jpg) 
  - Secretマニフェストファイルの例
    ~~~yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: sa-token
      namespace: monitoring
      annotations:
        kubernetes.io/service-account.name: default ---> ServiceAccount名に合せる
    type: kubernetes.io/service-account-token
    ~~~
    > **Note**  
    > ServiceAccountを先に作成してからSecretを作成すること

    > **Note**  
    > `metadata.annotations.kubernetes.io/service-account.name`と  
    > `type: kubernetes.io/service-account-token`を忘れないこと！

- 参考URL
  - https://zaki-hmkc.hatenablog.com/entry/2022/07/27/002213