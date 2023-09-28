## VPC Endpoint ServicesとNLBを使って(クロスアカウントの)異なるVPC上のリソースにアクセスする方法
- VPC PeeringなどでVPC間がNW的につながってない状態でアクセスできる
  - VPCのIPレンジが重複したりしてVPC Peering/Transit Gatewayによる繋ぎができない場合有用
  - 他にも **_VPC Lattice_** を使ってVPCのIPレンジ重複問題を回避できる

### ■ クロスアカウントでの設定方法
- 構成のイメージ ( 出典：https://www.yamamanx.com/aws-privatelink-nlb/ )  
  ![](image/Cross_Account_PrivateLink.jpg)
#### 手順
##### DBアカウント側
1. **Endpoint services**からendpoint serviceを作成
   ![](image/endpoint_service_1.jpg)
2. Load balancer typeはNetworkを選択し、Nameを記入し、繋げたいNLBを選択してCreateを押下。  
   Availableになったことを確認し、APPアカウント側で必要なので****を押さえておく。  
   **Allow principals**タブの**Allow principals**を押下。
   ![](image/endpoint_service_3.jpg)
- 参考URL
  - https://www.yamamanx.com/aws-privatelink-nlb/
  - https://docs.aws.amazon.com/ja_jp/vpc/latest/privatelink/privatelink-share-your-services.html
  - https://dev.classmethod.jp/articles/cross-account-rds-access-vial-privatelink-nlb/
