## import
### module化しているリソースのimport
１．backend.tf、main.tf、provider.tfがあるディレクトリに移動する
２．以下フォーマットでimportする
　　terraform import module.<module名>.<リソースカテゴリ名>.<リソース名> <存在するリソース識別子>
　　EC2インスタンスの例）module.ec2.aws_instance.cicd i-01d391bd44a0b047e
　　★countでネストしているリソースは[]で要素数を指定
　　terraform import module.ec2.aws_instance.cicd[0] i-01d391bd44a0b047e
３．terraform state show module.ec2.aws_instance.cicd[0]

- 参考URL
  - https://developer.hashicorp.com/terraform/cli/commands/import
  - https://tech.layerx.co.jp/entry/improve-iac-development-with-terraform-import
  - https://qiita.com/masato930/items/f5707be8077dba995978
