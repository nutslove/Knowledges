## Module構成での`output`
- １回`modules`ディレクトリ配下のtfで`output`ブロックで定義した後に、`module`の方で再度`output`ブロックを定義し、そこで`module.<module名>.<output名>`で指定する
### 例
- `modules`ディレクトリ配下のtfファイル  
  ```tf
  resource "aws_iam_policy" "test_policy" {
    name        = "test-policy"
    description = "A test policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "ec2:Describe*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }

  output "test_policy_id" {
    value = aws_iam_policy.test_policy.id
  }
  ```
- `module`側  
  ```
  module "iam" {
    source = "../../../modules/iam"
  }

  output "test_policy_id" {
    value = module.iam.test_policy_id
  }
  ```