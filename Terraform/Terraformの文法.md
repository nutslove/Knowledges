### `count` vs `for_each` vs `for`
- Terraformでloop処理のために用意されているものは`count`,`for_each`,`for`がある
- loop処理だけではなく、**環境ごとに差分を吸収するためにも使う**
- `count`と`for_each`は *Meta-Arguments* であり、`for`は *Expression* である
- 参考URL
  - https://zenn.dev/kyo2bay/articles/a6f98473141f36
  - https://tellme.tokyo/post/2022/06/12/terraform-count-for-each/
  - https://zenn.dev/wim/articles/terraform_loop
  - https://developer.hashicorp.com/terraform/language/meta-arguments/for_each

## `count`
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

## `for_each`
- `for_each`はリソースをmapとして作成する
- 参考URL
  - https://developer.hashicorp.com/terraform/language/meta-arguments/for_each

## `for`
- 参考URL
  - https://developer.hashicorp.com/terraform/language/expressions/for

### `count`が0か1以外の場合は`for_each`を使うべき？
- `count`は上述の通りリソースを配列として作成するため、途中で配列の中のリソースが削除された場合、indexが変わる。  
  indexが変わると
- 詳細については以下URLを参照
  - https://developer.hashicorp.com/terraform/language/meta-arguments/count#when-to-use-for_each-instead-of-count
  - https://zenn.dev/kyo2bay/articles/a6f98473141f36

## その他の文法
#### `length`
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

#### 条件分岐
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