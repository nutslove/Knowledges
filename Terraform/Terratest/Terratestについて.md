# Terratestとは
- goの`testing`パッケージを使って、Terraform PlanやApplyを実行することができる
- goのTestと同様に`_test.go`ファイル内に`"github.com/gruntwork-io/terratest/modules/terraform"`パッケージを使って`terraform init/plan/apply/destroy`処理を実装し、`go test -v [_test.go]`で実行できる

### 例
```go
package test

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"testing"
)

func TestTerratest(t *testing.T) {
	t.Parallel()
	awsRegion := "ap-northeast-1"
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../AWS/dev/tokyo/",
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	})
	if _, err := terraform.InitE(t, terraformOptions); err != nil {
		t.Errorf("Terraform Init Error: %v", err)
		return
	}
	if _, err := terraform.PlanE(t, terraformOptions); err != nil {
		t.Errorf("Terraform Plan Error: %v", err)
		return
	}
	if _, err := terraform.ApplyE(t, terraformOptions); err != nil {
		t.Errorf("Terraform Apply Error: %v", err)
		terraform.Destroy(t, terraformOptions)
	}
}
```