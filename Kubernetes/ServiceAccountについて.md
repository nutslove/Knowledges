## 基本的な知識
- ServiceAccountはプラグラム(Podで実行されるプロセス)がkube-apiserverへ認証するためのもの
- ServiceAccountはNamespacedリソース
- 各Namespaceには`default` ServiceAccountがある（自動で作成される）
- Podに`serviceAccountName`による明示的なServiceAccountの指定がなければ、Namespace内の`default` ServiceAccountを使用する
- 参考URL
  - https://kubernetes.io/ko/docs/reference/access-authn-authz/service-accounts-admin/

## v1.22以前
- v1.21まではServiceAccount作成時に一緒に作成されるSecret Object内の**無期限**のTokenが、Pod作成時に自動でPod内の`/var/run/secrets/kubernetes.io/serviceaccount`にMountされて、kube-apiserverへの認証に使われていた
- Podの中で`/var/run/secrets/kubernetes.io/serviceaccount`ディレクトリを見ると`token`がファイルとして存在していることを確認できる  
  ![Token_insidepod](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Token_InsidePod.jpg)  
- `/var/run/secrets/kubernetes.io/serviceaccount`内の`ca.crt`はkube-apiserverが提供する証明書の検証に使われる
  > Pods can use these certificates to make sure that they are connecting to your cluster's kube-apiserver
  - https://kubernetes.io/ko/docs/tasks/run-application/access-api-from-pod/
- **上記(既存)方式はセキュリティ観点で(無期限である等)課題があった**

## v1.22以降
- `Bound Service Account Token`というのがデフォルトで有効になり、SecretのTokenの代わりに`TokenRequest API`によって取得された短命(Defaultで1時間)のTokenがPodに
  > By default, the Kubernetes control plane (specifically, the ServiceAccount admission controller) adds a projected volume to Pods, and this volume includes a token for Kubernetes API access.
  >
  > A serviceAccountToken source, that contains a token that the kubelet acquires from kube-apiserver. The kubelet fetches time-bound tokens using the TokenRequest API. A token served for a TokenRequest expires either when the pod is deleted or after a defined lifespan (by default, that is 1 hour). The token is bound to the specific Pod and has the kube-apiserver as its audience. 
- 参考URL
  - **https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#bound-service-account-token-volume**
  - https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/kubernetes-versions.html → *Kubernetes 1.22*の部分を参照

## v1.24以前
- v1.23まではServiceAccountを作成すると自動的にToken(Secret)が作成された  
  ![ServiceAccount_Token](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/ServiceAccout_Token.jpg)
  - **このTokenはPodがkube-apiserverへの認証の際に直接使われるわけではない**
    - TokenがなくてもPodの中からkubectlが打てた  
      → https://amateur-engineer-blog.com/kubernetes-service-account/
  - このTokenは**無期限**だった
- このTokenはTypeが`kubernetes.io/service-account-token`の`Secret`として保存される  
  ![Secret](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret.jpg)  
  <img src="https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret2.jpg" width="1800" height="300">
  <!-- ![Secret2](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret2.jpg =250x250) -->

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