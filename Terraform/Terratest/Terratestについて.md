# Terratestとは
- goの`testing`パッケージを使って、Terraform PlanやApplyを実行することができる
- goのTestと同様に`_test.go`ファイル内に`"github.com/gruntwork-io/terratest/modules/terraform"`パッケージを使って`terraform init/plan/apply/destroy`処理を実装し、`go test -v [_test.go]`で実行できる

## 例
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

## `terraform.Options`で指定できるもの
- https://github.com/gruntwork-io/terratest/blob/main/modules/terraform/options.go  
  ```go
  // Options for running Terraform commands
  type Options struct {
  	TerraformBinary string // Name of the binary that will be used
  	TerraformDir    string // The path to the folder where the Terraform code is defined.

  	// The vars to pass to Terraform commands using the -var option. Note that terraform does not support passing `null`
  	// as a variable value through the command line. That is, if you use `map[string]interface{}{"foo": nil}` as `Vars`,
  	// this will translate to the string literal `"null"` being assigned to the variable `foo`. However, nulls in
  	// lists and maps/objects are supported. E.g., the following var will be set as expected (`{ bar = null }`:
  	// map[string]interface{}{
  	//     "foo": map[string]interface{}{"bar": nil},
  	// }
  	Vars map[string]interface{}

  	VarFiles                 []string               // The var file paths to pass to Terraform commands using -var-file option.
  	Targets                  []string               // The target resources to pass to the terraform command with -target
  	Lock                     bool                   // The lock option to pass to the terraform command with -lock
  	LockTimeout              string                 // The lock timeout option to pass to the terraform command with -lock-timeout
  	EnvVars                  map[string]string      // Environment variables to set when running Terraform
  	BackendConfig            map[string]interface{} // The vars to pass to the terraform init command for extra configuration for the backend
  	RetryableTerraformErrors map[string]string      // If Terraform apply fails with one of these (transient) errors, retry. The keys are a regexp to match against the error and the message is what to display to a user if that error is matched.
  	MaxRetries               int                    // Maximum number of times to retry errors matching RetryableTerraformErrors
  	TimeBetweenRetries       time.Duration          // The amount of time to wait between retries
  	Upgrade                  bool                   // Whether the -upgrade flag of the terraform init command should be set to true or not
  	Reconfigure              bool                   // Set the -reconfigure flag to the terraform init command
  	MigrateState             bool                   // Set the -migrate-state and -force-copy (suppress 'yes' answer prompt) flag to the terraform init command
  	NoColor                  bool                   // Whether the -no-color flag will be set for any Terraform command or not
  	SshAgent                 *ssh.SshAgent          // Overrides local SSH agent with the given in-process agent
  	NoStderr                 bool                   // Disable stderr redirection
  	OutputMaxLineSize        int                    // The max size of one line in stdout and stderr (in bytes)
  	Logger                   *logger.Logger         // Set a non-default logger that should be used. See the logger package for more info.
  	Parallelism              int                    // Set the parallelism setting for Terraform
  	PlanFilePath             string                 // The path to output a plan file to (for the plan command) or read one from (for the apply command)
  	PluginDir                string                 // The path of downloaded plugins to pass to the terraform init command (-plugin-dir)
  	SetVarsAfterVarFiles     bool                   // Pass -var options after -var-file options to Terraform commands
  	WarningsAsErrors         map[string]string      // Terraform warning messages that should be treated as errors. The keys are a regexp to match against the warning and the value is what to display to a user if that warning is matched.
  }
	```
### `Targets`オプション
- 以下のように`Targets`オプションで作成/削除するリソースを特定することもできる  
  ```go
  terraformOptions := &terraform.Options{
      TerraformDir: "../AWS/dev/tokyo/",
      Targets: []string{"aws_instance.example"},  // 特定のリソースを指定
  }
	```

### `Vars`オプション
- 以下のように`Vars`オプションでTerraformに変数を渡すこともできる
  - HCL  
    ```
    // main.tf
    variable "create_resource" {
      type    = bool
      default = true
    }

    resource "aws_instance" "example" {
      count = var.create_resource ? 1 : 0
      // ... 他の設定
    }
		```
  - terratest  
		```go
    terraformOptions := &terraform.Options{
        TerraformDir: "../AWS/dev/tokyo/",
        Vars: map[string]interface{}{
            "create_resource": false,
        },
    }
		```

## `terraform`メソッド種類
### `terraform.InitE()`

### `terraform.PlanE()`

### `terraform.ApplyE()`