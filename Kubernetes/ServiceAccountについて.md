- 参考URL
  - https://kubernetes.io/ko/docs/reference/access-authn-authz/service-accounts-admin/

#### 基本的な知識
- ServiceAccountはプラグラム(Podで実行されるプロセス)がkube-apiserverへ認証するためのもの
- ServiceAccountはNamespacedリソース
- Podに`serviceAccountName`による明示的なServiceAccountの指定がなければ、Namespace内の`default` ServiceAccountを使用する
- 

#### v1.24以前
- v1.23まではServiceAccountを作成すると自動的にToken(Secret)が作成された  
  ![ServiceAccount_Token](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/ServiceAccout_Token.jpg)
  - このTokenがPodがkube-apiserverへの認証の際に使われる
  - このTokenは**無期限**だった
- このTokenはTypeが`kubernetes.io/service-account-token`の`Secret`として作成される  
  ![Secret](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret.jpg)  
  ![Secret2](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret2.jpg)
- Podが作成される時にこのTokenがPod内の`/var/run/secrets/kubernetes.io/serviceaccount`にvolumeとして自動的にMountされる  
  ![Secret_Mount](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Secret_Mount.jpg)  
  - Podの中で`/var/run/secrets/kubernetes.io/serviceaccount`ディレクトリを見ると`token`がファイルとして存在していることを確認できる  
      ![Token_insidepod](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/Token_InsidePod.jpg)  
  - `/var/run/secrets/kubernetes.io/serviceaccount`内の`ca.crt`はkube-apiserverが提供する証明書の検証に使われる
    - https://kubernetes.io/ko/docs/tasks/run-application/access-api-from-pod/

#### v1.24以降
- v1.24からはServiceAccountを作成しても自動的にToken(Secret)が作成されなくなった
- 参考URL
  - https://zaki-hmkc.hatenablog.com/entry/2022/07/27/002213