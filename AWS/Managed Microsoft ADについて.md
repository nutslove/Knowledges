- __ハイブリッド環境向けに複数リージョンを跨ったAWS Managed Microsoft ADの設計__
  - [AWSブログ](https://aws.amazon.com/jp/blogs/news/multi-region-aws-managed-microsoft-ad-for-hybrid-environments-jp/)

## Best practices for AWS Managed Microsoft AD
- 使用の前に必ず目を通しておくこと！
- https://docs.aws.amazon.com/directoryservice/latest/admin-guide/ms_ad_best_practices.html

## AWS Managed Microsoft ADのクォーター
- https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/ms_ad_limits.html
- https://docs.aws.amazon.com/directoryservice/latest/admin-guide/ms_ad_limits.html

## AWS Managed Microsoft ADの制約
- 参考URL
  - https://dev.classmethod.jp/articles/re-introduction-2022-directory-service/
- __オンプレミスADなど既存のADとの接続についての制限事項__
  - オンプレミスADなど既存のドメインに対して、AWS Managed Microsoft ADを「追加のドメインコントローラー」として追加することはできない
  - AWS Managed Microsoft ADで構築したドメインに対して、オンプレミスAD等のドメインコントローラーを追加することはできない
- __ADの管理についての制限事項__
  - ドメイン管理者ユーザー「Administrator」がAWSの管理下となっており、AWS利用者はAdministratorを使うことができない  
   → 代わりに`Admin`というユーザーが用意されており、これを管理者ユーザーとして使用
  - ADディレクトリのデフォルトOU/コンテナ (ex.`Computers`, `Domain Controllers`, `Users`など) がAWSの管理下となっており、AWS利用者はこれらのOU/コンテナ内のオブジェクトを移動したりオブジェクトを新規作成することができない

## AWS Managed Microsoft ADで作成される対象
- 参考URL
  - https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/ms_ad_getting_started_what_gets_created.html

## ドメイン参加専用ユーザについて
- 参考URL
  - https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_join_privileges.html
- 利用者システム側にドメイン参加時に使うユーザも用意する必要がある
  - `Admins`と`AWS Delegated Server Administrators`グループのメンバーがその権限をもっているが、2つのグループは権限が強すぎるのでこれを利用者側に使わせるのは望ましくない
  - [AWSドキュメント](https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_join_privileges.html)の手順通りにドメイン参加用のユーザを作成する

## RODC(Read Only Domain Controllers)
- AWS Managed Microsoft AD自体はRODCをサポートしない
  - AWS Managed Microsoft ADからデプロイされるDCはwritable domain controllersになる
- ただ、自前でRODCを作成することはできて、特に制約事項もない
- https://docs.aws.amazon.com/whitepapers/latest/active-directory-domain-services/other-considerations.html


## AWS Managed Microsoft ADの設定
##### Password Policy
- ADのPassword Policyを管理する方法は以下2つ
  1. Group Policy Object(GPO)
     - Default Domain Policyのこと
     - 
  2. Password Setting Object(PSO)
- 5つの用意されているPassword Policyのいずれも適用しなかった場合、デフォルトのPassword Policy`Group Policy Object(GPO)`が適用される
  - [参考URL(AWS)](https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/assignpasswordpolicies.html)
  - [AWS Managed MSADのdefault Password Policy](https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/ms_ad_password_policies.html)
- PSOは`User`と __Group type__ が __Security__ 、__Group scoup__ が __Global__ の`Group`にだけ適用できる  
  > **Warning**  
  > Group scoupが`Domain local`のGroupには適用できない
- `Admin@<ドメイン>`ユーザにもデフォルトのPassword Policyが適用されるため、  
  Adminユーザを使って自動化をする場合などは無期限のPassword Policyを作成してAdminユーザを適用する
  - ドメイン参加用ユーザにも適用！

## その他
- AWS Managed ADを作成すると自動でSecurity Groupが作成され、ADのENI(Elastic Network Interfaces)にアタッチされる
  - https://docs.aws.amazon.com/directoryservice/latest/admin-guide/ms_ad_best_practices.html