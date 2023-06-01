- ECRリポジトリ上のイメージをクロスリージョン・クロスアカウントのECRにレプリケーションすることができる
  - **レプリケーション先のECRの`Private registry`の`Permissions`で事前に以下の許可を設定しておく必要がある**
    ~~~json
    {
      "Sid": "<任意の名前>",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::<レプリケーション元のAWSアカウントID>:root"
      },
      "Action": [
        "ecr:CreateRepository",
        "ecr:BatchImportUpstreamImage",
        "ecr:ReplicateImage"
      ],
      "Resource": "arn:aws:ecr:ap-northeast-1:<レプリケーション先(この設定をしているアカウント)のAWSアカウントID>:repository/*"
    }
    ~~~
    - https://docs.aws.amazon.com/ja_jp/AmazonECR/latest/userguide/registry-permissions-examples.html
- **レプリケーションされるのはレプリケーションの設定後にリポジトリにプッシュされたイメージのみ**
  - https://docs.aws.amazon.com/ja_jp/AmazonECR/latest/userguide/replication.html#replication-considerations
- 参考URL
  - https://docs.aws.amazon.com/ja_jp/AmazonECR/latest/userguide/replication.html