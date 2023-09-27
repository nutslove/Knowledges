- **Gatewayタイプ**はAWSサービスにアクセスするためのInternet Gatewayのようなもので、**Interfaceタイプ**とは異なりVPC内IPを消費しない。

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