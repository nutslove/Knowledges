## Task
- KubernetesのPodに該当するもの
- ECSの最小の実行単位
- Task定義に基づいて起動されるコンテナ群
- 同一Task内のコンテナは同一ホスト上で実行される

### Task Definition
- ECSタスクの設定を記述するテンプレート
- Kubernetesのマニフェストファイルのようなもの
- 以下のようなものを定義
  - コンテナイメージ
  - メモリとCPUのLimit（Task Size）
    - Task全体とコンテナ単位の設定ができる
    - KubernetesのようにRequestとLimitは分かれてない
  - タスクに割り当てるIAMロール
  - ネットワークモード
  - etc.

#### Task Size
- Task全体(Task内のすべてのコンテナが分け合ってして使う)リソースサイス
- Fargateの場合、指定できるCPUとメモリの値の組み合わせが決まっている
  - https://docs.aws.amazon.com/ja_jp/AmazonECS/latest/developerguide/task_definition_parameters.html#task_size
- `cpu`、`memory`２つの項目がある
- Fargateの場合、設定必須

#### `containerDefinitions`の中で定義できるコンテナ単位のリソース割り当て
- 以下３つの項目がある
  - `cpu`
    - CPUの場合は、Limitを超えてもコンテナが強制終了されず、CPUスロットリングがかかる
  - `memory`
    - コンテナに適用されるメモリの量（ハードリミット）
    - コンテナがここで指定したメモリを超えようとすると、強制終了される
  - `memoryReservation`
    - コンテナ用に予約するメモリのソフトリミット（タスクが保証する最低限のメモリ量）
- `memory`コンテナレベルと`memoryReservation`値の両方を指定する場合、`memory`値は`memoryReservation`値より大きくする必要がある

## Service
- KubernetesのDeploymentsに近い概念
- 実行する対象のECS Clusterや維持するTask(コンテナ)の数、実行するTask、ELBなどを指定
- ELBと連携できる
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

## DataPlaneとしてFargateを使う場合の注意事項
- Taskの定義にCPUとメモリのLimitの設定が必須（EC2の場合は省略可）
- 使用できるネットワークモードは`awsvpc`のみ
