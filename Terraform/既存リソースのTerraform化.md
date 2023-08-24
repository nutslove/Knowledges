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

#### Terraform v1.5からはTFファイルの中で`import`blockを使ってimportできるようになった
- 例
  ~~~t
  import {
    to = aws_ec2_transit_gateway.example --> <リソースカテゴリ名>.<リソース名>
    id = "tgw-12345678" --> リソース識別子
  }
  ~~~
- moduleの中のリソースをimportする例
  - moduleをincludeする側
    ~~~t
    import {
      to = module.tgw.aws_ec2_transit_gateway.my_transit_gateway --> module.<module名>.<リソースカテゴリ名>.<リソース名>
      id = "tgw-12345678" --> リソース識別子
    }

    module "tgw" {
      source             = "../../modules/AWS/tgw"
      my_tgw_description = "for VPC Peering, Site to Site VPN, Direct Connect"
      my_tgw_tag_name    = "My_TGW"
    }
    ~~~
  - module側
    ~~~t
    resource "aws_ec2_transit_gateway" "my_transit_gateway" {
      description = var.my_tgw_description
      tag = {
        Name      = var.my_tgw_tag_name
      }
    }

    variable "my_tgw_tag_name" {}
    variable "my_tgw_description" {}
    ~~~