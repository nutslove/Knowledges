- **Gatewayタイプ**はAWSサービスにアクセスするためのInternet Gatewayのようなもので、**Interfaceタイプ**とは異なりVPC内IPを消費しない。

## Interfaceタイプ VPC Endpoint（PrivateLink）
- ENI（Elastic Network Interface）として実現されている
- VPC外のサービス(e.g. IAM)が VPC で直接ホストされているかのように、VPC内のリソースがVPC内プライベート IP アドレスを使用して、VPC外のサービスに接続することを許可
  - VPC内のIPアドレスが１つ消費される(払い出される)
#### ENI
- VPC内の仮想ネットワークインターフェースであり、プライベートIPアドレス、セキュリティグループ、MACアドレスなどを持つ。
- ENIはNIC（Network Interface Card）の仮想的な表現。物理的なNICがサーバーに接続されているのと同様に、ENIは仮想マシンやコンテナに接続される。
- InterfaceタイプのVPC Endpointを作成すると、指定したサブネット内に新しいENIが作成される。このENIを通じて、VPC内のリソースはAWSのサービスに直接アクセスできるようになる。これにより、Internet Gateway、NAT Gateway、VPNコネクション、または AWS Direct Connect 接続を必要とせずに、VPC内のリソースからAWSサービスにプライベートにアクセスできる。

## VPC Endpoint Policy
- VPC Endpointごとにポリシーを適用できて、該当Endpointでアクセスできるリソースを制限できたりする
  - GatewayタイプとInterfaceタイプ両方に存在する
- あるS3バケットへのアクセスは許可して、特定のS3バケットへのアクセスは拒否するポリシー例（S3用VPC Endpointにアタッチ）
  ~~~json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Effect": "Allow",
              "Principal": "*",
              "Action": "s3:List*",
              "Resource": "arn:aws:s3:::*"
          },
          {
              "Effect": "Allow",
              "Principal": "*",
              "Action": "s3:*",
              "Resource": [
                  "arn:aws:s3:::<許可するS3バケット名>",
                  "arn:aws:s3:::<許可するS3バケット名>/*"
              ]
          },
          {
              "Effect": "Deny",
              "Principal": "*",
              "Action": "s3:*",
              "Resource": [
                  "arn:aws:s3:::<拒否するS3バケット名>",
                  "arn:aws:s3:::<拒否するS3バケット名>/*"
              ]
          }
      ]
  }
  ~~~