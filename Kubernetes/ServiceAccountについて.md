- 参考URL
  - https://kubernetes.io/ko/docs/reference/access-authn-authz/service-accounts-admin/

#### 基本的な知識
- ServiceAccountはプラグラム(Podで実行されるプロセス)がkube-apiserverへ認証するためのもの
- ServiceAccountはNamespacedリソース
- Podに`serviceAccountName`による明示的なServiceAccountの指定がなければ、Namespace内の`default` ServiceAccountを使用する
- 

#### v1.24以前
- v1.23まではServiceAccountを作成すると自動的にTokenが作成された
  - このTokenがPodがkube-apiserverへの認証の際に使われる
  ![ServiceAccount_Token](https://github.com/nutslove/Knowledges/blob/main/Kubernetes/image/ServiceAccout_Token.jpg)
- このTokenはTypeが`kubernetes.io/service-account-token`のSecretが作成され、
- 順番的に

#### v1.24以降