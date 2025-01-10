## `locals` vs `variable`
- 基本的に`locals`の使用が推奨されているっぽい。
- 参考URL
  - https://febc-yamamoto.hatenablog.jp/entry/2018/01/30/185416
### `locals`（Local Values）
- https://developer.hashicorp.com/terraform/language/values/locals
- module内でのみ使える変数
- 以下のように`locals`の中で関数が使える（`variable`はできない）  
  ~~~tf
  locals {
    load_balancer_count = var.use_load_balancer == "" ? 1 : 0
    switch_count        = local.load_balancer_count
  }
  ~~~
### `variable`（Input Variables）
- https://developer.hashicorp.com/terraform/language/values/variables
- 以下の方法で値の上書きができる
  - コマンドラインで`-var`オプションや`-ver-file`オプションで指定
  - `terraform.tfvars`ファイルで指定
  - 環境変数(`TF_VAR_xxx`など)で指定
  - `variable`の定義時にデフォルト値を明示

## `count` vs `for_each` vs `for`
- Terraformでloop処理のために用意されているものは`count`,`for_each`,`for`がある
- loop処理だけではなく、**環境ごとに差分を吸収するためにも使う**
- `count`と`for_each`は *Meta-Arguments* であり、`for`は *Expression* である
- 参考URL
  - https://zenn.dev/kyo2bay/articles/a6f98473141f36
  - https://tellme.tokyo/post/2022/06/12/terraform-count-for-each/
  - https://zenn.dev/wim/articles/terraform_loop
  - https://developer.hashicorp.com/terraform/language/meta-arguments/for_each
  - **https://zenn.dev/kasa/articles/8fe998e04cb916**

### `count`
- `count`はリソースを配列(リスト)として作成する
- 配列内のリソースは`[count.index]`で参照
- 例
  ~~~tf
  variable "subnet_ids" {
    type = list(string)
  }

  resource "aws_instance" "server" {
    # Create one instance for each subnet
    count = length(var.subnet_ids)

    ami           = "ami-a1b2c3d4"
    instance_type = "t2.micro"
    subnet_id     = var.subnet_ids[count.index]

    tags = {
      Name = "Server ${count.index}"
    }
  }
  ~~~
- 参考URL
  - https://developer.hashicorp.com/terraform/language/meta-arguments/count

### `for_each`
- `for_each`はリソースをmapとして作成する
- **リスト(List)** にしたい場合は`toset`を使う必要がある
  ~~~tf
  resource "aws_iam_user" "the-accounts" {
    for_each = toset( ["Todd", "James", "Alice", "Dottie"] )
    name     = each.key
  }
  ~~~
- **map**の場合
  ~~~tf
  resource "aws_route53_zone" "my_zone" {
    name = "example.com."
  }

  variable "dns_records" {
    description = "Map of DNS records to create"
    default = {
      "www" = {
        type    = "A"
        ttl     = 300
        records = ["123.123.123.123"] # 仮のIPアドレス
      },
      "api" = {
        type    = "CNAME"
        ttl     = 300
        records = ["api.example.com"]
      }
      # 必要に応じて他のレコードを追加
    }
  }

  resource "aws_route53_record" "my_records" {
    for_each = var.dns_records

    zone_id = aws_route53_zone.my_zone.zone_id
    name    = "${each.key}.example.com."
    type    = each.value.type
    ttl     = each.value.ttl
    records = each.value.records
  }

  output "dns_record_names" {
    description = "The names of the DNS records"
    value       = [for r in aws_route53_record.my_records : r.name]
  }
  ~~~
- 参考URL
  - https://developer.hashicorp.com/terraform/language/meta-arguments/for_each

### `for`
- 例(1)  
  ~~~tf
  variable "TEST_NLB_target_groups" = {
    [
      {
        name      = "Aurora",
        targets   = [
          {target = "100.100.100.1"},
          {target = "100.100.100.2"}
        ],
        port      = 5432
      },
      {
        name      = "OEM",
        targets   = [
          {target = "100.100.100.5"}
        ],
        port      = 7802
      }
    ]
  }

  locals {
    TEST_NLB_target_group_config = flatten([
      for i, group in var.TEST_NLB_target_groups : [
        for target in group.targets : {
          group_index = i
          target      = target
          port        = group.port
        }
      ]
    ])
  }

  resource "aws_lb_target_group_attachment" "TEST_NLB_TG" {
    count                      = length(local.TEST_NLB_target_group_config)
    target_group_arn           = aws_lb_target_group.TEST_NLB[local.TEST_NLB_target_group_config[count.index].group_index].arn
    port                       = local.TEST_NLB_target_group_config[count.index].port
    target_id                  = local.TEST_NLB_target_group_config[count.index].target.target
    availability_zone          = "all"
  }
  ~~~
- 例(2) `for`文であるmapから新たなmapを作成  
  **`name`がkey、`=>`右側がvalueになる**
  ~~~tf
  locals {
    # サンプルのマップ: 人の名前とその年齢
    people = {
      "john" = 28
      "jane" = 34
      "mike" = 25
    }

    # 既存のマップから新しいマップを生成し、何らかの変換を適用
    people_with_greetings = {
      for name, age in local.people:
      name => "Hello, my name is ${name} and I am ${age} years old."
    }
  }

  output "people_with_greetings" {
    value = local.people_with_greetings
  }
  ~~~
  - 上のoutputの出力は以下  
    ~~~tf
    people_with_greetings = {
      "jane" = "Hello, my name is jane and I am 34 years old."
      "john" = "Hello, my name is john and I am 28 years old."
      "mike" = "Hello, my name is mike and I am 25 years old."
    }
    ~~~
- 参考URL
  - https://developer.hashicorp.com/terraform/language/expressions/for

### `count`が0か1以外の場合は`for_each`を使うべき？
- `count`は上述の通りリソースを配列として作成するため、途中で配列の中のリソースが削除された場合、indexが変わる。  
  indexが変わると
- 詳細については以下URLを参照
  - https://developer.hashicorp.com/terraform/language/meta-arguments/count#when-to-use-for_each-instead-of-count
  - https://zenn.dev/kyo2bay/articles/a6f98473141f36

## その他の文法
### `length`
- indexの長さを取得
- 例
  ```tf
  resource "aws_lb_target_group_attachment" "nlb_test_target" {
    count = length(var.nlb_targets)

    target_group_arn = aws_lb_target_group.nlb_targets.arn
    port             = var.nlb_target_port
    target_id        = var.nlb_targets[count.index].target
    availability_zone = "all"
  }
  ```

### 三項演算子
- 例１）num変数の値が１の場合は`yes`が、1以外の場合は`no`が入る
  - `var.num == "1" ? "yes" : "no"`
- 例２）  
  ~~~tf
  variable "environment" {
    type    = string
    default = "prd"
  }

  resource "aws_instance" "example" {
    instance_type = var.environment == "prd" ? "t2.large" : "t2.micro"
    ami           = var.environment == "prd" ? "ami-0123456789" : "ami-9876543210"
    subnet_id     = var.environment == "prd" ? "subnet-12345678" : "subnet-87654321"
  }
  ~~~
- 以下のように条件文をネストすることもできる。  
  以下の場合、最初の条件`try(v.domain_name, "") != ""`がfalseの場合、次の条件文`v.is_s3_origin`が判定される  
  ```tf
  try(v.domain_name, "") != "" ? v.domain_name : v.is_s3_origin ? module.s3[v.origin_sub_sid].regional_domain_name : module.alb[v.origin_sub_sid].dns_name
  ```


### `flatten`


### `dynamic`
