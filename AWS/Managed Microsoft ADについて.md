- AWS Managed Microsoft ADのクォーター
  - https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/ms_ad_limits.html
- __ハイブリッド環境向けに複数リージョンを跨ったAWS Managed Microsoft ADの設計__
  - [AWSブログ](https://aws.amazon.com/jp/blogs/news/multi-region-aws-managed-microsoft-ad-for-hybrid-environments-jp/)

## AWS Managed Microsoft ADの設定
##### Password Policy
- ADのPassword Policyを管理する方法は以下
  1. Group Policy Object(GPO)
     - Default Domain Policyのこと
     - 
  2. Password Setting Object(PSO)
- 5つの用意されているPassword Policyのいずれも適用しなかった場合、デフォルトのPassword Policy`Group Policy Object(GPO)`が適用される
  - [参考URL(AWS)](https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/assignpasswordpolicies.html)
  - [AWS Managed MSADのdefault Password Policy](https://docs.aws.amazon.com/ja_jp/directoryservice/latest/admin-guide/ms_ad_password_policies.html)
- PSOはUserと __Group type__ が __Security__ 、__Group scoup__ が __Global__ のGroupにだけ適用できる  
  > **Warning**  
  > Group scoupが`Domain local`のGroupには適用できない
