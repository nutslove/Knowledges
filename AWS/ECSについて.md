## Service
- KubernetesのDeploymentsに近い概念
- 実行する対象のECS Clusterや維持するTask(コンテナ)の数、実行するTask、ELBなどを指定
- Terraform Resource
  - https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_service

## IAM Role
- Task RoleとTask Execution Roleの２つがある

### Task Role
- Task自体がAWSリソースにアクセスするためのRole
- タスク内で実行されるコンテナが直接AWSサービス（例：DynamoDB、S3、SNSなど）にアクセスするためのもの
- アプリケーションが特定のAWSリソースを操作する必要がある場合に利用

### Task Execution Role
- ECSサービス自体が使用するRoleで、Taskの実行に必要な操作を行うために使用される
- 以下のような用途で使われる
  - ECRからコンテナイメージの取得
  - CloudWatch Logsへのログ書き込み
  - Secrets Managerからの機密情報の取得