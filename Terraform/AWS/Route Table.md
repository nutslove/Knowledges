- `aws_route_table`リソースは中に定義されているルートのうち、1つでも変わるとすべてのルートを1回`-`で外して`+`でアタッチする動きになる
  - https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table.html