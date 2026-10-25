# Cloud Mapとは
- サービスディスカバリーのためのサービス
- KubernetesのCoreDNSのようなもの
- **アプリケーションのリソース（EC2、コンテナ、Lambda、DynamoDBテーブル、SQSキューなど）に任意の名前を付けて登録し、その名前でリソースの場所（IP/ポート/ARN/任意の属性）を動的に検索できるようにするフルマネージドサービス**
  - 接続先のIPアドレスをハードコードしなくて済む
  - スケールイン/アウト、タスクの入れ替えで変わるIPを自動で追従できる
  - ヘルスチェックを設定すると、正常なリソースだけを返せる
- 参考URL
  - https://docs.aws.amazon.com/ja_jp/cloud-map/latest/dg/what-is-cloud-map.html
  - https://aws.amazon.com/cloud-map/pricing/
  - https://docs.aws.amazon.com/general/latest/gr/cloud_map.html

---

## 基本概念（Namespace / Service / Instance）
```
Namespace  (e.g. prod.local)              ← サービスの論理グループ。DNS名のドメイン部分になる
 └─ Service  (e.g. backend)               ← 同種のリソースのテンプレート（DNSレコード種別・TTL・ヘルスチェック等を定義）
     ├─ Instance (id: task-a, 10.0.1.10:8080)   ← 実際に登録されるリソース
     └─ Instance (id: task-b, 10.0.2.20:8080)
```
- DNSで解決する場合のFQDNは **`<Service名>.<Namespace名>`** (e.g. `backend.prod.local`)
- **Namespaceはリージョン単位**。複数リージョンで使いたい場合は各リージョンでNamespaceを作る

### Namespace
- サービスの論理的なグループ。作成時に「どう発見させるか」を指定する
- 3種類ある

| 種類 | Console上の表記 | 解決方法 | Route 53 Hosted Zone | 用途 |
| --- | --- | --- | --- | --- |
| **HTTP** | API calls | `DiscoverInstances` APIのみ（DNSでは引けない） | 作成されない | IP以外のリソース（SQS、DynamoDB等）や、属性で絞り込みたい場合。DNS不要な場合 |
| **Private DNS** | API calls and DNS queries in VPCs | DNSクエリ + API | **Private Hosted Zone**が自動作成される | VPC内のサービス間通信（最も一般的） |
| **Public DNS** | API calls and public DNS queries | DNSクエリ + API | **Public Hosted Zone**が自動作成される | インターネットから名前解決させたい場合 |

- DNS namespaceを作ると、**Cloud Mapが裏でRoute 53のHosted Zoneを作成・管理する**
  - Namespace名 = Hosted Zone名。レコード名は`<service-name>.<namespace-name>`
  - Namespaceを削除するとHosted Zoneも削除される
  - **このHosted Zoneは、アカウントのRoute 53 Hosted Zoneクォータにカウントされる**
  - **Cloud Mapが作ったRoute 53のレコードやHosted Zoneを直接編集/削除しないこと**（Cloud Map側と不整合になる）
- **Private DNS namespace**
  - 作成時にVPCを1つ指定する（`--vpc`）
  - VPC側で **`enableDnsHostnames` と `enableDnsSupport` が有効**である必要がある（ECSドキュメントに記載。VPC DNS属性が正しくないと名前解決できない）
  - VPC内のクエリはRoute 53 Resolverが解決する。一致するレコードがなければ`NXDOMAIN`
  - Hosted Zoneは複数VPCに関連付けられる（`AssociateVPCWithHostedZone`）
- **Public DNS namespace**
  - **Namespace名は自分が登録済みのドメイン名**で、TLDで終わっている必要がある（e.g. `.com`）
  - 属性やInstanceIdに**機密情報を入れない**（SRVの場合InstanceIdもDNSで見える）
- DNS namespaceの作成時に**SOAレコードのTTL**を指定できる（これはHosted ZoneのSOAのキャッシュ時間で、各Serviceのレコードの TTLとは別）
- Namespace作成は非同期で、`OperationId`が返る（`aws servicediscovery get-operation`で確認）
- 国際化ドメイン名(IDN)を使う場合は事前にPunycode変換が必要

### Service
- 同じ種類のリソースをまとめるテンプレート。以下を定義する
  - **DNSレコードの種類**（DNS namespaceの場合）
  - **ルーティングポリシー**（DNS namespaceの場合）
  - **TTL**
  - **ヘルスチェック設定**
- 種類（API上の`Type`）: `HTTP`（`DiscoverInstances`のみ）/ `DNS_HTTP`（DNSとAPI両方）
- **レコードの種類は後から変更できない**（Serviceを削除して作り直す必要がある）
- **`HealthCheckCustomConfig`は既存のServiceに追加/更新/削除できない**
- DNSでアクセスされるServiceは、大文字小文字だけが異なる名前のServiceを作れない（同じDNS名になるため）。API-onlyのNamespaceならOK
- SRVレコードをHAProxy等の特定形式で使いたい場合は、Service名を`_service._tcp`のような形式にする
- **Serviceレベルの属性**（`UpdateServiceAttributes`、最大30個）も持てる
  - インスタンス間で重複する属性をまとめたり、プログレッシブデプロイ時のエンドポイントの重み、APIタイムアウト/リトライ方針のような情報を持たせる用途

#### DNSレコードの種類
| 種類 | 返す値 | 備考 |
| --- | --- | --- |
| `A` | IPv4アドレス | |
| `AAAA` | IPv6アドレス | |
| `CNAME` | ドメイン名 | **`WEIGHTED`ルーティング必須**。**Route 53ヘルスチェックは設定不可** |
| `SRV` | `priority weight port service-hostname` | priority/weightは**1固定で変更不可**。portは`AWS_INSTANCE_PORT`。service-hostnameは`<InstanceId>.<Service名>.<Namespace名>` |

- 指定できる組み合わせは `A` / `AAAA` / `A`+`AAAA` / `CNAME` / `SRV` のいずれか
- SRVを指定した場合、IPv4/IPv6を登録すると**同名のA/AAAAレコードも自動作成**される

#### ルーティングポリシー
| ポリシー | 動作 |
| --- | --- |
| `MULTIVALUE` | 正常なインスタンスの値を**最大8つ**返す。ヘルスチェック未設定なら全て正常とみなして最大8つ返す |
| `WEIGHTED` | **ランダムに選ばれた1つ**を返す。全レコードの重みは同じで、重みの調整はできない |

- ヘルスチェックの結果、**全インスタンスが異常の場合はRoute 53が全て正常であるかのように応答**する（`WEIGHTED`なら1つ、`MULTIVALUE`なら最大8つ）

### Instance
- Serviceに登録される個々のリソース。`InstanceId`（Service内で一意、最大64文字、`[0-9a-zA-Z_/:.@-]`）と**属性（Attributes）**のkey-valueで構成される
- 既存の`InstanceId`と`ServiceId`で`RegisterInstance`すると**更新**になる（DNSレコード更新。ヘルスチェックがある場合は古いものを削除して作り直す）
- 同じService ID + Instance IDに対する`RegisterInstance`は、前のリクエストが完了してから次を送る必要がある
- 予約済みの主な属性

| 属性 | 内容 |
| --- | --- |
| `AWS_INSTANCE_IPV4` | IPv4アドレス（Aレコード用。Aレコードの場合は必須） |
| `AWS_INSTANCE_IPV6` | IPv6アドレス（AAAAレコード用。AAAAレコードの場合は必須） |
| `AWS_INSTANCE_PORT` | ポート番号。SRVレコード、またはRoute 53ヘルスチェックを使う場合は必須 |
| `AWS_INSTANCE_CNAME` | CNAMEレコードの値。CNAMEレコードの場合は必須 |
| `AWS_ALIAS_DNS_NAME` | **ELB（ロードバランサー）のDNS名**を指定してRoute 53のAliasレコードを作る。制約あり（後述） |
| `AWS_EC2_INSTANCE_ID` | EC2インスタンスID（**HTTP namespaceのみ**）。指定すると併用できるのは`AWS_INIT_HEALTH_STATUS`のみで、プライマリプライベートIPv4が`AWS_INSTANCE_IPV4`に自動設定される |
| `AWS_INIT_HEALTH_STATUS` | カスタムヘルスチェックの初期ステータス（`HEALTHY` / `UNHEALTHY`）。未指定は`HEALTHY` |

- `AWS_ALIAS_DNS_NAME`の制約
  - ELB以外のAWSリソースへのAliasは非対応
  - ServiceにA/AAAAレコードの設定が必要
  - **ルーティングポリシーは`WEIGHTED`必須**
  - `AWS_INSTANCE_*`属性とは併用不可
  - Serviceにヘルスチェックがあっても、Aliasレコードにはヘルスチェックが関連付けされない
  - **GovCloudでは非対応**
- **カスタム属性は最大30個**
  - キー最大255文字、値最大1,024文字、**全属性の合計が5,000文字以内**
  - `DiscoverInstances`でこの属性を使って絞り込める（DNSでは絞り込み不可）
- 登録できるリソースの例：EC2、DynamoDBテーブル、S3バケット、SQSキュー、API Gatewayで公開したAPI など

---

## リソースの探索方法（Discovery）
### 1. DNSクエリ
- DNS namespaceの場合に使える。アプリ側は普通のDNS名前解決をするだけでOK（SDK不要）
```bash
# VPC内から
dig backend.prod.local
```
- **TTLの間はクライアント/リゾルバにキャッシュされる**ため、インスタンスが消えても古いIPが返り続けることがある
  - → TTLを短くする、アプリ側でリトライ/不良バックエンドの除外をする、等の考慮が必要

### 2. `DiscoverInstances` API
- HTTP namespace / DNS namespaceどちらでも使える（**HTTP namespaceでは唯一の方法**）
- **結果はランダムな順序で返る**（トラフィックを均等に分散するため）
```bash
aws servicediscovery discover-instances \
  --namespace-name prod-http \
  --service-name backend \
  --query-parameters stage=prod \
  --health-status HEALTHY
```
- 主なパラメータ

| パラメータ | 内容 |
| --- | --- |
| `NamespaceName` / `ServiceName` | 必須。Namespaceは`HttpName`で指定する（通常`Name`と同じ。同名のNamespaceを作り直した場合はハッシュが付く） |
| `QueryParameters` | カスタム属性による絞り込み。**指定した全てのkey-valueに一致**したものだけ返る |
| `OptionalParameters` | 「できれば絞り込みたい」フィルタ。QueryParametersと合わせて一致するものがあればそれを返し、なければこのフィルタは無視される |
| `HealthStatus` | `HEALTHY`(デフォルト) / `UNHEALTHY` / `ALL` / `HEALTHY_OR_ELSE_ALL`。**ヘルスチェック未設定のServiceでは無視され全て返る** |
| `MaxResults` | 1〜1000。未指定は最大100件 |
| `OwnerAccount` | **共有Namespaceを使う場合は必須**（Namespace所有者のアカウントID） |

  - `HEALTHY_OR_ELSE_ALL`: 正常なものがあればそれだけ、全部異常なら全部返す（フェイルオープン）
- レスポンスの`InstancesRevision`は、**Instanceの登録/解除で更新される。ヘルスステータスの変化では更新されない**
- **データプレーンのエンドポイントは `data-servicediscovery.<region>.amazonaws.com`**（管理系APIの`servicediscovery.<region>.amazonaws.com`とは別）
- `DiscoverInstancesRevision` は、Namespace/Serviceのリビジョンだけを軽量に取得するAPI。変更があった時だけ`DiscoverInstances`を呼ぶ、というキャッシュ運用に使える
- IPv6クライアント用のデュアルスタックエンドポイントは `servicediscovery.<region>.api.aws`

### 3. インスタンスの一覧取得
- `ListInstances` / `GetInstance`（管理系API）。ディスカバリー用途ではなく管理・確認用

---

## ヘルスチェック
- 設定しない場合は、**ヘルス状態に関係なく全インスタンスが返される**
- 設定すると、デフォルトでは正常なものだけが返る
- `GetInstancesHealthStatus` APIで各インスタンスのヘルス状態を取得できる
- **Route 53ヘルスチェックとカスタムヘルスチェックのどちらか一方のみ指定可能**（両方は不可）

### 1. Route 53 ヘルスチェック（`HealthCheckConfig`）
- **Public DNS namespaceとHTTP namespaceのみ。Private DNS namespaceでは設定できない**
- Instanceを登録するたびにCloud MapがRoute 53のヘルスチェックを自動作成し、解除すると削除する。プロトコルは `HTTP` / `HTTPS` / `TCP`
- Route 53のヘルスチェッカーが各ヘルスチェックリージョンから30秒間隔で確認する
- Public DNS namespaceの場合：登録時に作られるRoute 53レコードにヘルスチェックを関連付ける
  - AとAAAA両方指定した場合、IPv4アドレスでチェックされる。IPv4が異常ならAもAAAAも異常扱い
  - **CNAMEレコードには設定不可**
- **HTTP namespaceの場合**：ヘルスチェックは作られるが、**関連付けるDNSレコードがない**
  - 状態の確認はRoute 53コンソールやCloudWatchアラーム（`PutMetricAlarm`）で行う
- 別途Route 53のヘルスチェック料金がかかる

### 2. カスタムヘルスチェック（`HealthCheckCustomConfig`）
- Cloud Map自身は**ヘルスチェックを行わない**。自分（もしくはサードパーティのヘルスチェッカー）でチェックし、結果を`UpdateInstanceCustomHealthStatus`でCloud Mapに通知する
```bash
aws servicediscovery update-instance-custom-health-status \
  --service-id srv-xxxxxxxx --instance-id task-a --status UNHEALTHY
```
- Cloud Mapは**最後に通知されたステータスを記録するだけ**
- 動作：`UNHEALTHY`を通知してから**30秒以内に`HEALTHY`の通知が来なければ**、そのリソースへのルーティングを止める
  - `FailureThreshold`は**非推奨で、常に1固定**（指定しても意味がない）
- VPC内のリソースの場合、ヘルスチェッカーも同じVPC内にある必要がある
- **Private DNS namespaceで使える唯一のヘルスチェック方式**
- **ECSのService Discoveryはこの方式**（ECSがコンテナのヘルスチェックとタスクの状態をCloud Mapに反映する）

### 動作まとめ
- DNSクエリ：異常と判定されたインスタンスは返されない
  - ただし**全インスタンスが異常の場合**は、Route 53が全て正常であるかのように応答する（`WEIGHTED`なら1つ、`MULTIVALUE`なら最大8つ）
- `DiscoverInstances`：`HealthStatus`で制御（デフォルトは`HEALTHY`のみなので、全部異常なら空リスト）

---

## インスタンスの登録方法
### 手動/スクリプト（CLI、SDK）
```bash
# Namespace作成（Private DNS）
aws servicediscovery create-private-dns-namespace \
  --name prod.local --vpc vpc-xxxxxxxx

# 作成完了の確認
aws servicediscovery get-operation --operation-id <OperationId>

# Service作成
aws servicediscovery create-service \
  --name backend \
  --namespace-id ns-xxxxxxxx \
  --dns-config 'RoutingPolicy=MULTIVALUE,DnsRecords=[{Type=A,TTL=10}]' \
  --health-check-custom-config FailureThreshold=1

# Instance登録
aws servicediscovery register-instance \
  --service-id srv-xxxxxxxx \
  --instance-id i-0123456789 \
  --attributes AWS_INSTANCE_IPV4=10.0.1.10,AWS_INSTANCE_PORT=8080

# 解除（DNSレコードとヘルスチェックも削除される）
aws servicediscovery deregister-instance --service-id srv-xxxxxxxx --instance-id i-0123456789
```
- EC2の場合は、起動時のユーザーデータ/起動スクリプトで`register-instance`、停止時に`deregister-instance`を実行する運用になる（自動登録はされない）
- 失敗したリクエストを安全にリトライするために`CreatorRequestId`を指定できる

### ECS Service Discovery（自動登録）
- https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-discovery.html
- ECS Serviceに **Service registry** を関連付けると、**Taskの起動/停止に合わせてCloud Mapに自動で登録/解除**される
- **ECSのService Discovery（`serviceRegistries`）は、Service作成後も`UpdateService`で追加/更新/削除できる**
  - 変更すると新しいデプロイが走り、新しいTaskに切り替わる
  - 削除する場合は空リストを渡す

#### ネットワークモードと使えるレコード
| ネットワークモード | 使えるレコード |
| --- | --- |
| `awsvpc` (Fargateは必ずこれ) | `A` / `SRV`（の組み合わせ）。デュアルスタックのサブネットなら`AAAA`も可。IPv6のみのサブネットでは`A`は不可 |
| `bridge` / `host` | `SRV`のみ。Task定義のコンテナ名とコンテナポートの組み合わせを指定する |
| `none` | 非対応 |

  - SRVを使う場合はポートが必須
  - `awsvpc`は**Taskごとに固有のIP**を持ち、**サービス単位でSecurity Groupを分けられる**のでService Discoveryと相性が良い

#### ECSがCloud Mapに登録するカスタム属性
`AWS_INSTANCE_IPV4` / `AWS_INSTANCE_IPV6` / `AWS_INSTANCE_PORT` に加えて、以下も付く

| 属性 | 内容 |
| --- | --- |
| `AVAILABILITY_ZONE` | Taskが起動したAZ |
| `REGION` | リージョン |
| `ECS_SERVICE_NAME` | ECS Service名 |
| `ECS_CLUSTER_NAME` | ECS Cluster名 |
| `EC2_INSTANCE_ID` | Taskが配置されたコンテナインスタンスのID（Fargateでは付かない） |
| `ECS_TASK_DEFINITION_FAMILY` | Task定義のファミリー |
| `ECS_TASK_SET_EXTERNAL_ID` | 外部デプロイのTask Setを使う場合のExternal ID |

#### 制約/注意点
- **Fargateはプラットフォームバージョン1.1.0以降**
- **Service Discoveryを設定したServiceは、1 Serviceあたり最大1,000 Task**（Route 53のクォータが理由）
- ECSコンソールでServiceを作成する場合、**Private DNS namespaceのみ**選択できる
- **ECSは共有（RAM）されたCloud Map Namespaceへの登録は非対応**（Service Connectは対応）
- **DNSレコードには、Public namespaceを使っていても、Taskの「プライベートIP」が登録される**
- ELBの背後にあるServiceにも設定できるが、**Service Discoveryのトラフィックは、ELBではなく常にTaskに直接ルーティング**される
- **Classic Load Balancerとは併用不可**
- DNSクエリは`<service-discovery-service-name>.<service-discovery-namespace>`の形式
- **ECSが作ったCloud Mapのリソースは、手動で掃除する必要がある**
- ヘルスチェックはECS管理のカスタムヘルスチェック（`HealthCheckCustomConfig`）を推奨
  - Taskは**コンテナのヘルスチェックが結果を返すまで`UNHEALTHY`として登録**され、パスすると`HEALTHY`になる。**失敗したらInstanceが解除（deregister）される**
  - 異常なエンドポイントはDNSから外れる
  - 8件以下しか正常なレコードがなければ、Route 53はその全てを返す。**全て異常なら最大8件の異常なレコードを返す**
- 料金：Route 53（Hosted Zone、クエリ）とCloud MapのDiscovery APIの料金がかかる。ECSによるコンテナレベルのヘルスチェック連携自体は追加料金なし。公開Taskに別途ネットワークヘルスチェックを設定した場合はその料金がかかる
- 他のサービスからは `http://backend.prod.local:8080` のようにアクセスできる

#### AWSServiceRoleForECS（サービスリンクロール）
- https://docs.aws.amazon.com/AmazonECS/latest/developerguide/using-service-linked-roles-for-clusters.html
- ECSのService Discoveryで、Cloud MapやRoute 53のリソースを操作しているのは**ユーザーのIAMロールではなく、このサービスリンクロール**
  - **Task Execution Role / Task Roleとは別物**。Task定義に`servicediscovery:*`の権限を付ける必要はない
- ロール名: **`AWSServiceRoleForECS`**
  - 信頼ポリシー: `ecs.amazonaws.com`のみがAssumeできる
  - 権限ポリシー: AWS管理の **`AmazonECSServiceRolePolicy`**（サービスリンクロール専用で、他のIAMユーザー/ロールにはアタッチ不可）
- **自動で作成される**
  - クラスターを作成したとき、またはServiceを作成/更新したときにECSが作る（コンソール、CLI、APIいずれも）
  - 2017/1/1以前からECSを使っていたアカウントには、すでに存在している
  - IAMコンソールの「AWSServiceRoleForECS」ユースケース、もしくはCLIで`ecs.amazonaws.com`を指定して手動作成もできる
- ECSのService Discovery/Service Connectに関係する権限（`AmazonECSServiceRolePolicy`より）

| 用途 | 権限 | 条件 |
| --- | --- | --- |
| Instanceの登録/解除、ヘルス更新 | `servicediscovery:RegisterInstance` / `DeregisterInstance` / `UpdateInstanceCustomHealthStatus` | なし（`Resource: *`） |
| 参照 | `servicediscovery:Get*` / `List*`、`route53:Get*` / `List*` | なし |
| Route 53レコード/ヘルスチェックの操作 | `route53:ChangeResourceRecordSets` / `CreateHealthCheck` / `DeleteHealthCheck` / `UpdateHealthCheck` | なし |
| Service Connect用のCloud Mapリソース作成 | `servicediscovery:CreateHttpNamespace` / `CreateService` | リクエストのタグキーが`AmazonECSManaged`のみ |
| タグ付け | `servicediscovery:TagResource` | リクエストタグに`AmazonECSManaged`がある場合のみ |
| Serviceの削除 | `servicediscovery:DeleteService` | **`AmazonECSManaged`タグが付いたリソースのみ** |
| ディスカバリー | `servicediscovery:DiscoverInstances` / `DiscoverInstancesRevision` | なし |
| Serviceの属性更新 | `servicediscovery:UpdateServiceAttributes` | **`AmazonECSManaged`タグが付いたリソースのみ** |

  - ポリシーに**`CreatePrivateDnsNamespace`や`DeleteNamespace`は含まれていない**（Service Connectが自動作成するのはHTTP namespace）。Private DNS namespaceは、ユーザー自身の権限で作成する必要がある、と読み取れる（これは公式の明記ではなく、ポリシー内容からの推測）
  - **`AmazonECSManaged`タグ**が付いたCloud Mapリソース = ECSが作った/管理しているもの。ECSが自分で作ったもの以外は、削除や属性更新ができないようにスコープされている
  - ポリシーにはほかに、ENI管理、ELBのターゲット登録/解除、Auto Scaling、CloudWatchアラーム/ログ、ECS Execなどの権限も含まれる
- **編集はできない**。名前も変更不可（説明文のみIAMで編集可能）
- 削除したい場合は、先に全リージョンでServiceを（desired count 0にして）削除、コンテナインスタンスの強制登録解除、クラスターの削除が必要。全リージョンのクラスターを削除すると、ECSが自動でこのロールも削除する
  - 削除後でも、クラスター/Serviceを作成/更新すると再び自動作成される
- ユーザー側で必要な権限：サービスリンクロールの作成/編集/削除を許可するIAM権限（`iam:CreateServiceLinkedRole`など）。初回に作成するユーザーに付与されている必要がある

### その他の連携
- **ECS Service Connect**：Cloud MapのNamespaceを利用する（後述の比較参照）
- **EKS**：[AWS Cloud Map MCS Controller for K8s](https://github.com/aws/aws-cloud-map-mcs-controller-for-k8s)を使うと、Multi-Cluster Services API(KEP-1645)の実装として、Cloud Mapを使ってクラスタ間でServiceを共有できる
  - `ServiceExport` / `ServiceImport` を使う
  - 2つ以上のEKSクラスタが必要で、**同一AWSアカウント・同一リージョン**の想定。クラスタ間はネットワーク的に到達できる必要がある（VPC Peering、共有VPC等）
  - リポジトリの更新は活発ではなく、v0.1.0が最新（要確認）
- **AWS App Mesh**：Cloud Mapを仮想ノードのService Discoveryとして利用していたが、**App Meshは2026/9/30にサポート終了済み**
  - ECSの場合の移行先はService Connectが推奨されている
  - App MeshとService Connectの両方に属することはできないため、ECS Serviceの作り直し（blue/green推奨）が必要

---

## 他の仕組みとの比較
### ECS Service Discovery vs ECS Service Connect
- https://docs.aws.amazon.com/AmazonECS/latest/developerguide/interconnecting-services.html
- https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-connect.html
- どちらも内部ではCloud Mapを使うが、**仕組みが全く違う**

| 項目 | Service Discovery | Service Connect |
| --- | --- | --- |
| 公式の位置づけ | 直接通信（DNS） | **AWS推奨**。Service Discovery、接続性、トラフィックの監視をECSの設定として提供 |
| Cloud Mapの使い方 | **Private DNS（またはPublic DNS）namespace** + Route 53 Hosted Zoneを使う | **HTTP namespaceを自動作成して使う**（Namespaceの種類は問わない）。**Route 53のHosted Zoneは作成も使用もしない** |
| 名前解決 | VPCのDNS（Route 53 Resolver）でDNSクエリ。**VPCのDNS設定に依存** | 各Task内の**Service Connectプロキシ**が解決する。**VPCのDNS設定に依存しない**。短い名前（`mysql`など）と標準ポートで接続可能 |
| トラフィックの経路 | Task → Task（直接）。間に余計なコンポーネントなし | Task内のプロキシ → 相手Task内のプロキシ → アプリ（**両方のTaskにプロキシが入る**） |
| レイテンシ | 低い（直接通信） | プロキシを2回経由する分のオーバーヘッドがある |
| ロードバランシング | DNSの応答依存（ラウンドロビン的、クライアントのキャッシュ依存） | プロキシが**ラウンドロビン**で分散 |
| 障害時の挙動 | **リトライや接続失敗の処理はアプリ側で実装**する必要がある。TTLの間は古いIPが返り続ける | プロキシが**外れ値検出(outlier detection)とリトライ**を行う |
| ヘルスチェック | ECS（コンテナ）のヘルスチェック → Cloud Mapカスタムヘルスチェック → DNSから除外 | 起動時のヘルスチェック（essential containerのヘルスチェック定義が必要。`UNKNOWN`の場合は最大8分で判断）+ 以降は**パッシブな外れ値検出** |
| 設定変更の反映 | DNSのTTL経過後（クライアントのキャッシュ次第） | クライアントServiceの**デプロイで設定が更新される**（Task入れ替え）。デプロイサーキットブレーカー等のデプロイ設定で制御できる |
| メトリクス/ログ | なし | プロキシがCloudWatchにメトリクスを出す。アプリケーションプロトコル(HTTP等)を指定するとプロトコル別のメトリクス(5xx等)も取れる。アクセスログ、ゾーンを考慮したルーティング(zone-aware routing)、TLS暗号化(AWS Private CA)にも対応 |
| 追加コンポーネント | 不要 | Taskごとにプロキシコンテナ（CPU/メモリを共有して消費） |
| ネットワークモード | awsvpc ○ / bridge ○(SRV) / host ○(SRV) | awsvpc ○ / bridge ○ / **host ×** |
| 使えるレコード/ポート | `A`は`awsvpc`のみ。bridge/hostではSRV必須（アプリがSRV対応している必要あり） | Task定義の**ポートに名前**を付ける（`portMappings`の`name`）。DNSレコードの種類は意識不要 |
| 共有Namespace(RAM) | **非対応** | 対応（アカウントをまたぐServiceも可） |
| 接続できる範囲 | VPC内でDNSが引ける範囲 | **同一Namespace内**のService同士（同一リージョン。クラスタ/VPCをまたげる）。**Namespace外、Service Connect未設定のTask、ECS外のアプリ**はエンドポイント名を解決できない |
| セキュリティグループ | Service単位で細かく分けやすい（awsvpc） | 同様にawsvpcで制御 |
| 料金 | Route 53（Hosted Zone/クエリ）+ Cloud MapのAPI料金 | **追加料金なし**（Cloud Map利用料も無料）。プロキシ分のvCPU/メモリは自分のTask側で消費。TLS暗号化を使う場合のみPrivate CA料金 |
| 制約 | 1 Serviceあたり最大1,000 Task | Task定義の`links`は非対応。ServiceはNamespaceを1つだけ持つ |
| 注意 | ECSが作ったCloud Mapリソースは手動で掃除 | **Service Connectが作ったCloud Mapリソースを手動で編集（登録/解除、属性変更、Service削除）すると、トラフィックやデプロイに予期しない影響が出る** |

- **使い分けの目安**
  - ECS外のクライアント（EC2、Lambda、オンプレなど）からも名前で引きたい、DNSだけで済ませたい、プロキシによるオーバーヘッドを避けたい → **Service Discovery**
  - ECS Service同士の通信で、リトライ/外れ値検出/メトリクス/切り替えの確実性が欲しい → **Service Connect**
- ECS外から名前解決したい場合は、**Service Connectのエンドポイント名は使えない**ので別の手段が必要
- そのほか、**VPC Lattice**（ECS Taskをターゲットとして登録できる、VPC/アカウントをまたぐアプリケーションネットワーキングサービス）という選択肢もある。VPC LatticeはServiceに対して`host`モードも含め対応している
- ECSでインターネットからの外部接続を受けるなら、ELBを使う

### Cloud Map vs Route 53 Private Hosted Zone（手動管理）
- Route 53 Private Hosted Zoneに自分でレコードを登録/更新する場合は、IP変更のたびに自前で更新処理が必要
- Cloud Mapは**登録/解除のAPIと、ヘルスチェック連動、属性による検索**が使えるのが利点
- DNS namespaceの実体はRoute 53 (Hosted Zone) なので、**VPC内の名前解決は結局Route 53 Resolver（VPC DNS）が行う**

### Cloud Map vs ELB
- ELBはL4/L7のロードバランシング + ヘルスチェックを提供する（トラフィックがELBを経由する）
- Cloud Mapは「宛先を教える」だけでトラフィックは経由しない。クライアントが直接宛先に接続する
- ELBをCloud Mapに登録する（`AWS_ALIAS_DNS_NAME`）こともできる

### Cloud Map vs Kubernetes (CoreDNS)
- Kubernetesでは `Service` を作ると自動で `<svc>.<ns>.svc.cluster.local` が引けるが、Cloud Mapは**明示的にNamespace/Service/Instanceを作って登録する**必要がある（ECSは自動登録）
- Kubernetesのクラスタ外（EC2、Lambda、別クラスタ）も同じ仕組みで扱える

---

## 共有Namespace（AWS RAM）
- https://docs.aws.amazon.com/cloud-map/latest/dg/sharing-namespaces.html
- **AWS RAM（Resource Access Manager）でNamespaceを他のアカウント/OU/Organization全体に共有**できる
  - 共有できるのは**自分が所有するNamespaceのみ**（共有されたNamespaceを再共有はできない）
  - Organizationで共有する場合は、RAMでOrganizationsとの共有を有効化しておく
  - GovCloudでも2025年10月からクロスアカウント共有に対応
- 共有Namespaceを使う側は、Service/NamespaceをID（またはARN）で指定する。共有Namespaceのものは**ARNで指定**する
  - `DiscoverInstances`では`OwnerAccount`が必須
- **DNSで共有Private DNS namespaceを使う場合の追加手順**
  - Namespace所有者：`aws route53 create-vpc-association-authorization --hosted-zone-id <ID> --vpc VPCRegion=...,VPCId=...`
  - 利用者：`aws route53 associate-vpc-with-hosted-zone --hosted-zone-id <ID> --vpc VPCRegion=...,VPCId=...`
  - 異なるVPCにある場合、別途VPC Peering等でネットワーク接続も必要
- 権限
  - 所有者：Namespace配下の全Service/Instanceにアクセス、共有の取り消し、他アカウントへの登録権限設定、Namespaceの更新/削除
  - 利用者：Serviceの作成/削除、Instanceの登録/解除、Instanceの発見。**Namespaceの更新/削除は不可**。共有が取り消されると、利用者が作ったServiceにもアクセスできなくなる
  - 共有には`servicediscovery:PutResourcePolicy` / `GetResourcePolicy` / `DeleteResourcePolicy`が必要
- 制約：共有Namespaceに対するタグ付けは不可。他アカウントが実行した操作は`ListOperations`で見られない
- 課金：Instance登録分と、ヘルスチェック、`DiscoverInstances`は各アカウントが自分の分を課金される。DNS Namespaceの場合、Serviceに対して作られるRoute 53のDNSレコード分は所有者の課金
- クォータ：共有Namespaceは所有者の「Namespaces per Region」にカウント。利用者が登録したInstanceは所有者の「Instances per namespace」にカウントされるが、「Instances per service」はServiceを作った側のカウント

---

## 料金
- https://aws.amazon.com/cloud-map/pricing/
- **登録リソース数** と **ディスカバリーAPIコール数** に対する従量課金（初期費用なし）

| 項目 | 料金 (USD) |
| --- | --- |
| 登録リソース（Instance） | 1リソースあたり **$0.10/月** |
| Discovery APIコール（`DiscoverInstances`、HTTP namespace） | **$1.00/100万コール** |

- DNSクエリで解決する場合、代わりに **Route 53の料金**がかかる
  - Hosted Zone：$0.50/月（最初の25ゾーン）
  - DNSクエリ：$0.40/100万クエリ（最初の10億クエリ/月）
  - Route 53ヘルスチェックを使う場合はその料金も
- 料金ページの表が取得できず、上記は公式の料金ページ内の計算例から読み取った値。最新は公式ページを確認すること

---

## クォータ（Service Quotas）
- https://docs.aws.amazon.com/cloud-map/latest/dg/cloud-map-limits.html

| 項目 | デフォルト | 引き上げ |
| --- | --- | --- |
| Namespaces per Region | 50 | 可 |
| Instances per namespace | 2,000 | 可 |
| Instances per service | 1,000 | **不可** |
| Custom attributes per instance | 30 | 不可 |
| `DiscoverInstances` steady rate (per account) | 1,000 req/s | 可 |
| `DiscoverInstances` burst rate (per account) | 2,000 req/s | 可 |
| `DiscoverInstancesRevision` rate (per account) | 3,000 req/s | 可 |

- DNS namespaceでInstance数の上限を引き上げるには、**Route 53の「Hosted Zoneあたりのレコード数」の引き上げ**が必要で、追加料金がかかる
- `DiscoverInstances`の**スロットリングはトークンバケット方式**（リージョン・アカウント単位で全エンドポイント共通）
  - バケットサイズ2,000、リフィルレート1,000/秒。超えると`RequestLimitExceeded`
  - コンソールやCLI、サードパーティからの呼び出しも全てカウントされる
  - 対策：指数バックオフでリトライ、結果のキャッシュ、`DiscoverInstancesRevision`で変更検知してから取得する
- 引き上げはService Quotasから、またはAWS Supportに依頼

---

## ネットワーク/エンドポイント
- 管理系API: `servicediscovery.<region>.amazonaws.com`、データプレーン（`DiscoverInstances`）: `data-servicediscovery.<region>.amazonaws.com`
- **VPCエンドポイント（PrivateLink）を使う場合は2種類ある**

| 用途 | サービス名 |
| --- | --- |
| 管理系API | `com.amazonaws.<region>.servicediscovery`（FIPS: `...servicediscovery-fips`） |
| `DiscoverInstances` | `com.amazonaws.<region>.data-servicediscovery`（FIPS: `...data-servicediscovery-fips`） |

  - **`DiscoverInstances`は管理系エンドポイントでは使えない**ので、データプレーン用が別途必要
  - データプレーン用エンドポイントのリージョナル/ゾーナルDNS名で`DiscoverInstances`を呼ぶ場合は、**SDK/CLIのホストプレフィックス注入(host prefix injection)を無効化**する必要がある
  - Private DNSを有効にすると、デフォルトのリージョナルDNS名（`servicediscovery.us-east-1.amazonaws.com`）でそのまま呼べる
  - VPCエンドポイントはクロスリージョンのリクエストに非対応。Security Groupは443のインバウンドを許可する
  - ECSのTaskがawslogsでログを送る場合、IGWがないVPCではCloudWatch Logs用のVPCエンドポイントも必要

---

## IAM
- https://docs.aws.amazon.com/cloud-map/latest/dg/security-iam-awsmanpol.html
- AWS管理ポリシー

| ポリシー名 | 内容 |
| --- | --- |
| `AWSCloudMapDiscoverInstanceAccess` | Discovery API（`DiscoverInstances`等）へのアクセス |
| `AWSCloudMapReadOnlyAccess` | 全Cloud Mapアクションの読み取り専用 |
| `AWSCloudMapRegisterInstanceAccess` | Namespace/Serviceの読み取り + Instanceの登録/解除 |
| `AWSCloudMapFullAccess` | 全Cloud Mapアクション |

- ECSのService Discovery/Service Connectでの登録/解除は、ユーザーではなく**`AWSServiceRoleForECS`**が行う（詳細は上の「AWSServiceRoleForECS」参照）
- アプリが`DiscoverInstances`を使う場合は、アプリ側のIAMロールに`servicediscovery:DiscoverInstances`が必要（SDKも必要）

---

## 注意点/ハマりどころ
- **削除の順番**：Instance（登録解除）→ Service → Namespace。Instanceが残っているServiceや、Serviceが残っているNamespaceは削除できない
  - ECS Service Discoveryを使っている場合、**ECSが作ったCloud Mapリソースは自動では削除されず手動で掃除が必要**
- Cloud Mapが作成したRoute 53のHosted Zone/レコードは**手動で変更しない**。Service Connectが作ったCloud Mapリソースも手動で変更しない
- DNSのTTLキャッシュにより、**Task入れ替え直後に古いIPへ接続して失敗**することがある。TTLを短くする + アプリ側でリトライ
  - Javaなど言語/ランタイムによってはDNSの結果を独自にキャッシュするので注意（JVMの`networkaddress.cache.ttl`など）
- Private DNS namespaceのドメインは、VPC内でしか名前解決できない。**他VPC/オンプレから引きたい場合は、VPCの関連付け、もしくはRoute 53 Resolver (Inbound/Outbound)** が必要
- Route 53ヘルスチェックは**Private DNS namespaceでは使えない**。VPC内のリソースはカスタムヘルスチェックを使う
- カスタムヘルスチェックは**既存のServiceに後から追加/変更できない**（最初から設定しておく）
- DNSレコードの種類の変更も、Serviceの作り直しが必要
- 全インスタンスが異常になってもDNSでは応答が返り続ける（フェイルオープン）。`DiscoverInstances`のデフォルトは空リストになる
