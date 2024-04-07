## horizon
- OpenStackのダッシュボードのコンポーネント（AWSのマネコンのようなもの）

## keystone
- OpenStackの認証周りを担当するコンポーネント（AWSのIAMのようなもの）
- 以下のような機能を提供
  - ユーザー、グループ、ロールの管理
  - 認証（Authentication）
  - 認可（Authorization）
  - サービスカタログの管理
    - 利用可能なサービスの情報提供
  - トークンの発行と検証
- OpenStack環境内の各サービス（Nova、Cinder、Neutronなど）に対する認証と認可を一元管理し、ユーザーはSSOでOpenStackの各サービスを利用できるようになる

## nova
- OpenStackのコンピュートサービスを担当するコンポーネント（AWSのEC2のようなもの）

## glance
- OpenStackのマシンイメージサービスを担当するコンポーネント（AWSのAMIのようなもの）
