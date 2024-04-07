- OpenStackはCloudを管理するための**CMS（Cloud Management System）**
- OpenStackはVM(仮想化)の実現のためにHypervisorとして **KVM（Kernel-based Virtual Machine）** を使っている
  - A hypervisor is a software that you can use to run multiple virtual machines on a single physical machine. Every virtual machine has its own operating system and applications. The hypervisor allocates the underlying physical computing resources such as CPU and memory to individual virtual machines as required.
- **KVM（Kernel-based Virtual Machine）**
  - Linuxカーネルに統合されている仮想化技術。この技術により、Linux上で複数の仮想マシンを実行することが可能になる。各仮想マシンには、プライベートな仮想化されたハードウェア（仮想CPU、仮想メモリ、ディスク、ネットワークインターフェースなど）が提供され、実際の物理ハードウェア上で独立したオペレーティングシステムを実行できる。
- **OpenStackはKVMをベースとしているため、OpenStackがインストールされるサーバ（コントロールノード）だけではなく、実際にVMが実行されるサーバ（コンピュートノード）もLinux OSが必要である**

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
