- Grafana ad codeできるツールとしてTerraform,Ansible,Crossplaneなど複数ある
- ツールによってはAlertingなどサポートしてないコンポーネントもある
- 参考URL
  - https://grafana.com/blog/2022/12/06/a-complete-guide-to-managing-grafana-as-code-tools-tips-and-tricks/
  - https://grafana.com/blog/2020/02/26/how-to-configure-grafana-as-code/
## Grafana alerts as code
- **Terraformなどコードで作成したアラートは手動では修正できない**
  - https://grafana.com/docs/grafana/latest/alerting/set-up/provision-alerting-resources/terraform-provisioning/
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