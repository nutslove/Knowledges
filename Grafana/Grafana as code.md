- Grafana ad codeできるツールとしてTerraform,Ansible,Crossplaneなど複数ある
- ツールによってはAlertingなどサポートしてないコンポーネントもある
- 参考URL
  - https://grafana.com/blog/2022/12/06/a-complete-guide-to-managing-grafana-as-code-tools-tips-and-tricks/
  - https://grafana.com/blog/2020/02/26/how-to-configure-grafana-as-code/
## Terraform
- **Terraformなどコードで作成したアラートはデフォルトでは手動では修正できない**
  - https://grafana.com/docs/grafana/latest/alerting/set-up/provision-alerting-resources/terraform-provisioning/
  - **ただ、`disable_provenance`項目を`true`にするとTerraformから作成したアラートも手動で修正できるようになる**
    - https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/rule_group
  - コードで作成したアラートの設定の中身は見れないので、中身を見たい場合は、  
    コードで作成されたアラートをCopyしてCopyしたやつから中身を確認できる
- **Admin権限のAPI Keyが必要**
- 手動で作ったアラートをexportしたyamlを参照してコードを作成
- 参考URL
  - https://grafana.com/docs/grafana/latest/alerting/set-up/provision-alerting-resources/
  - https://grafana.com/blog/2022/09/20/grafana-alerts-as-code-get-started-with-terraform-and-grafana-alerting/
- `reduce`や`math`など、TimeRangeがないものについても`relative_time_range`を定義する必要がある

### module化している場合の`Provider`,`module`の書き方
- importされる`module/`配下の`.tf`側とproviderやimportする側両方に`terraform.required_providers`の設定が必要！
- Providerやmoduleをimportする側の設定例
  ~~~terraform
  terraform {
    required_providers {
        grafana = {
            source = "grafana/grafana"
            version = ">= 1.37.1"
        }
    }
  }
 
  provider "grafana" {
    url  = "http://grafana:3000"
    auth = "**************************" ## Admin API Key
  }
 
  module "grafana_alert" {
    source                             = "../../../modules/grafana_alert"
 
    folder_uid                         = "zyOWMLa4k"
    org_id                             = "1"
    victoriametrics_datasource_uid     = "EDzSloZVz"
  }
  ~~~
- importされる`module/`配下の`.tf`側の設定例
  ~~~terraform
  terraform {
    required_providers {
      grafana = {
        source = "grafana/grafana"
        version = ">= 1.37.1"
      }
    }
  }
 
  resource "grafana_rule_group" "deployment_pod_count_alert" {
  　　　　　　　　　　　　　　・
  　　　　　　　　　　　　　　・
  　　　　　　　　　　　　　　・
  }
  ~~~

> [!NOTE]  
> ## tfstateファイルをS3に管理し、Grafana TokenをSecrets Managerで管理する場合の例
> ```hcl
> terraform {
>   required_providers {
>     grafana = {
>       source  = "grafana/grafana"
>       version = "~> 4.0"
>     }
>    aws = {
>      source  = "hashicorp/aws"
>      version = "~> 5.0"
>    }
>  }
> }
>
> terraform  {
>  backend "s3" {
>    bucket         = "lee-test-grafana-terraform-bucket"
>    key            = "dev/grafana/terraform.tfstate"
>    region         = "ap-northeast-1"
>    encrypt        = true
>  }
> }
>
> provider "aws" {
>  region = "us-west-2" # Secrets Managerが存在するリージョンを指定
> }
>
> data "aws_secretsmanager_secret" "grafana_token" {
>  name = "lee-grafana-test" # Secrets ManagerのSecret nameを指定
> }
>
> data "aws_secretsmanager_secret_version" "grafana_token" {
>  secret_id = data.aws_secretsmanager_secret.grafana_token.id
> }
>
> provider "grafana" {
>  url = "https://dev-grafana.com/"
>  auth = jsondecode(data.aws_secretsmanager_secret_version.grafana_token.secret_string)["dev_grafana_token"]
> }
> ```
#### Data Source
- https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source

#### Dashboard
- https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/dashboard
- jsonファイルにダッシュボードの中身を定義しておいて、`config_json`に`file(<jsonファイル>)`でファイルを指定することもできる
  - https://www.devopstricks.in/generate-grafana-dashboard-using-terraform/
    ~~~terraform
    resource "grafana_dashboard" "general" {
       config_json = file("tf_dashboard.json") 
    }
    ~~~
  - またはJSONの中身をそのまま記載することもできる
    ~~~terraform
    resource "grafana_dashboard" "general" {
       config_json = <<EOF
    <JSONの中身をそのままコピペ>
    EOF
    }
    ~~~
- **`config_json`には、既存のGrafanaダッシュボードの「Dashboard settings」(歯車マーク) →「JSON Model」の値をそのままコピーしてファイル(拡張子は`.json`)として保存して、それをそのまま`file`で指定すればできる**
  - 「JSON Model」のデータの末尾の`title`と`uid`は既存のダッシュボードと被らないように修正が必要

#### Contact Point
- https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/contact_point

#### Notification Policy
- https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/notification_policy