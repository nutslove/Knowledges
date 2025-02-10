# 参考URL
- https://dev.classmethod.jp/articles/how-to-use-vscode-remote-ssh-with-aws-systems-manager/
- https://repost.aws/ja/articles/AR8Gk1UngsTpmpu7azMUiNvw/how-to-connect-to-a-private-ec2-instance-from-a-local-visual-studio-code-ide-with-session-manager-and-aws-sso-cli

# 前提
- EC2はPrivate Subnet上に存在
  - EC2にPublic IPをアタッチする必要はない
- EC2のセキュリティグループで自VPCに対してHTTPSの許可が必要？（定かではない）
- EC2が存在するPrivate Subnetのルートは`0.0.0.0/0`をNAT Gatewayに向けて、IGWがあるPublic Subnetのルートは`0.0.0.0/0`をIGWに向ける

# 手順
- Private Subnet上にEC2を作成する
- EC2に`AmazonSSMManagedInstanceCore`PolicyがアタッチされているIAM Roleをアタッチする
- 以下のVPC Endpointを作成し、Private Subnet(s)に関連付ける
  - `com.amazonaws.<region (e.g. ap-northeast-1)>.ec2messages`
  - `com.amazonaws.<region (e.g. ap-northeast-1)>.ssm`
  - `com.amazonaws.<region (e.g. ap-northeast-1)>.ssmmessages`
  - `com.amazonaws.ap-northeast-1.s3` (GatewayタイプとInterfaceタイプ)
    - S3はSSMと直接関係はないが、`dnf`でパッケージインストールする場合などに必要なのでアタッチしておく
- VPC Endpointのセキュリティグループにて該当VPCからの443ポートを許可する
- AWS CLI、SSM Session Manager Plugin、VSCODE(`code --version`コマンドで確認)をインストール
- profileがある`config`ファイルを用意する
- `~/.ssh/config`に以下のように追記する  
  ```
  host <任意のホスト名>
    HostName "インスタンスID"
    Port 22
    User ec2-user
    IdentityFile "EC2キーペア秘密鍵のフルパス" 
    ProxyCommand C:\Program Files\Amazon\AWSCLIV2\aws.exe ssm start-session --target %h --document-name AWS-StartSSHSession --parameters "portNumber=%p" --profile your_profile
  ```