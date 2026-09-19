# Cloud Runについて

最終更新: 2026-09-20（Web調査ベース）

## 1. 概要

Cloud Runは、コンテナ・コード・関数をGoogleのスケーラブルなインフラ上でサーバー管理なしに実行できる**フルマネージドのサーバーレスプラットフォーム**。もともとはHTTPリクエスト駆動のコンテナ実行サービス（旧称のイメージ）だったが、2025〜2026年にかけて「Worker Pools」「Instances」などが追加され、リクエスト駆動以外のワークロードもカバーする総合コンピュートプラットフォームへと拡張されている。

## 2. 4つの実行形態（コンポーネント）

| コンポーネント | 用途 | 特徴 |
|---|---|---|
| **Services** | HTTPリクエストを処理する常時稼働型エンドポイント | `*.run.app`のHTTPSエンドポイント自動付与。TLS/WebSocket/HTTP2/gRPC対応。リクエストベースの高速オートスケール（最大1000インスタンス）。デフォルトでゼロスケール |
| **Jobs** | 完了後に終了するバッチ・タスク | 単発実行 or 「array job」で複数タスクを並列処理。CLI実行・スケジュール実行・Workflows内実行に対応 |
| **Worker Pools**（2026年4月GA） | HTTPリクエストに依存しない継続的バックグラウンド処理 | Pub/Sub・Kafka・RabbitMQなどのプル型ワークロード向け。**自動スケールしない**（手動スケールが基本、CREMAで外部メトリクスベースの自動スケールも可）。ロードバランサ型のURLを持たない。GPU（NVIDIA Blackwell RTX PRO 6000等）対応 |
| **Instances**（プレビュー） | 単一の永続的ランタイムが必要なワークロード | 個別に管理・アドレス指定可能な単一インスタンス。起動約20秒以内。長時間稼働のAIエージェントや開発環境のデバッグ用途 |

※ **Functions**（旧Cloud Functions）はServicesの一部という位置づけで、Pub/Subメッセージ、Cloud Storage変更、FirebaseイベントなどイベントドリブンのワークロードをServiceとして実行する形。

### 2.1 Worker PoolsとInstancesの違い

一見似ているが、違いの核心は **「単一の個別アドレス指定可能なインスタンス」か「複数インスタンスのプール」か** という点。

| 観点 | Instances | Worker Pools |
|---|---|---|
| インスタンス数 | 常に1つ（シングルトン） | 複数（プール） |
| アドレス指定 | 可能（個別に名指しできる） | 不可（群として扱う） |
| スケーリング概念 | なし（複製しない） | 手動スケール（インスタンス数を増減、CREMAで自動化も可） |
| 典型用途 | 長時間稼働の単一プロセス（AIエージェントセッション、デバッグ環境） | キュー消費・分散バッチ処理（Pub/Sub, Kafka） |
| ロールアウト方式 | ― | トラフィック分割ではなく「インスタンス数の分割」で行う（例: 4インスタンス中1つを新リビジョン、3つを安定版に） |
| ステータス | プレビュー | GA（2026年4月） |

- **Instances**: 個別にアドレス指定・管理できる、安定した継続稼働ランタイムが必要なワークロード向け。Servicesのような複数インスタンスへの水平スケーリングは行わない（そもそも1個しか存在しない）
- **Worker Pools**: プル型（pull-based）の継続的バックグラウンド処理を複数インスタンスで水平分散して捌く。個々のインスタンスに意味はなく「群」として仕事をこなす。ロードバランサ型のURL・エンドポイントを持たず、外部からの直接アドレス指定は想定されていない

→「同じものを増やして分散処理させたい」ならWorker Pools、「唯一無二のプロセスを個別に触りたい」ならInstances、という住み分け。

#### 「アドレス指定可能（addressable）」とは

特定の1つのインスタンスを名指しして直接アクセスできるかという意味。ネットワーク上の宛先（IPアドレスやホスト名のようなもの）が個別に割り当てられていて、「このインスタンスに繋ぐ」と指定できる状態を指す。

- **Services**: `*.run.app`という1つのURLの裏に複数インスタンスが隠れており、ロードバランサがその時空いているどれか1つに振り分ける。どのインスタンスが応答したかは分からないし選べない（個別にアドレス指定できない＝匿名的・使い捨て）
- **Instances**: インスタンスが1つしかなく、それに対して固有の宛先を割り当てられる。「毎回同じそのインスタンスに接続する」ことが保証される。長時間動いているAIエージェントのセッションに途中から再接続して続きの状態にアクセスしたい、といった用途で必要になる
- **Worker Pools**: そもそも外部から接続されるURLやエンドポイント自体を持たない。各ワーカーはPub/Sub等のキューから自発的に仕事を取りに行く（プル型）だけなので、「特定のワーカーを名指しする」という概念自体が存在しない

たとえるなら、Services＝コールセンターの代表電話番号（誰が出るか分からない）、Instances＝担当者への直通の内線番号（毎回同じ人につながる）、Worker Pools＝倉庫で待機している複数の作業員（外から呼び出すのではなく、彼らが自分で仕事を取りに行く）。「ステートを保持した特定のプロセスに、後から確実に戻ってこられるか」がアドレス指定可能性のポイント。

## 3. 主な機能・最新動向（2025〜2026）

- **GPU対応の展開順序**: Services（2025年6月GA、NVIDIA L4、クォータ申請不要）→ Jobs（2025年10月GA）→ Worker Pools（2026年4月GA、NVIDIA RTX PRO 6000 Blackwell対応）という順に段階的に拡大。いずれもスケールtoゼロが可能で、秒単位課金
  - L4 GPU: 24GB VRAM、最低4 vCPU・16GiBメモリが必須。中規模モデル向け。提供リージョン: asia-southeast1, asia-south1, europe-west1, europe-west4, us-central1, us-east4など
  - Blackwell GPU（Worker Pools向け）: 96GB VRAM、1インスタンスあたり最大44 vCPU・176GiBメモリ。GPU利用にはCPU 20以上・メモリ80GiB以上が必須。起動時間は約5秒（ドライバプリインストール済み）。us-central1・europe-west4でオンデマンド提供（asia-south2・asia-southeast1は限定提供）
  - **Zonal Redundancy（ゾーン冗長性、Worker Pools向け）**: デフォルトでON。ゾーン障害時に他ゾーンへGPUキャパシティを予約確保（コスト増）。OFFにするとベストエフォート failover（低コスト）
  - 実運用例: Estee Lauder Companiesが自社コンテナ運用の代わりにバッチAI処理をWorker Poolsで実行
- **Direct VPC Egress/Ingress**: Serverless VPC Accessコネクタ不要でVPCに接続可能。ネットワークコストもゼロスケール。Worker PoolsのDirect VPC Ingressは2026-02-05追加。Shared VPC・Private NAT（GA, 2025-10-21）・VPC Flow Logs（GA, 2025-10-20）・デュアルスタックIPv6（GA, 2025-11-06）にも対応
- **Cloud Service Mesh連携**: 構成可能だが、コールドスタート遅延が増加する点に注意
- **ジョブのタスクタイムアウト**: 最大168時間（7日）までGA（2025-11-11）。GPU使用時は最大1時間
- **ランタイム追加**: PHP 8.5（プレビュー, 2026-02）、.NET 10（プレビュー, 2026-01-27）、Java 25（プレビュー, 2025-10-31）
- **ビルド強化**: GPU有効なソースデプロイ時、Cloud Buildのe2-highcpu-8マシンタイプがデフォルトに（2025-10-30）

## 4. 実行環境の世代（gen1 / gen2）

公式ドキュメントで明記されている選択指針：

- **第1世代（gen1）**: gVisorベースのサンドボックス。コールドスタートが速いが、一部のシステムコールはエミュレーションとなり完全なLinux互換ではない。**トラフィックが急増（バースト）するサービス、多数のインスタンスへの高速スケールアウトが必要なサービス、コールドスタート時間に敏感なサービス**に向く
- **第2世代（gen2）**: microVMベースで完全なLinux互換を実現。持続的な高負荷時の性能は一般的にgen1より高く、ネットワーク性能も向上、NFS連携も可能。ただし**gen1よりコールドスタートが長い**（実測ではgen2は3秒を切ることはほぼないという報告もある）。**トラフィックが比較的安定していてコールドスタートの多少の遅さを許容できるサービス、CPU負荷の高い処理、高いネットワーク性能が必要なサービス**に向く。最低512MiBのメモリが必要（gen1はそれ未満も可）

## 5. 料金モデル

Servicesには2つの課金設定がある：

| 設定 | 説明 | 向くケース |
|---|---|---|
| **Request-based billing（デフォルト）** | リクエスト処理中のみCPUを割り当てて課金。起動・停止時にも課金 | トラフィックが不安定・突発的（スパイク型）なサービス |
| **Instance-based billing（旧"CPU always allocated"）**※Pre-GA | インスタンスのライフサイクル全体で課金（リクエスト単位の課金なし）。最低1分課金。CPUは25%安く、メモリは20%安い料金設定 | トラフィックが安定・緩やかに変動するサービス、バックグラウンド処理を伴うサービス |

- `min-instances`を1以上に設定した場合、Request-basedモードではアイドル時はメモリのみ課金（CPUは課金されない）。Instance-basedモードではアイドル時もCPU+メモリ両方が課金される
- Jobs・Worker Pools・Instancesはそれぞれ稼働時間に応じた課金
- ゼロスケール時は無課金。無料枠（Free Tier）も用意
- Google Cloud RecommenderがトラフィックパターンからRequest-based⇔Instance-based切り替えを提案してくれる
- ※Instance-based billingは執筆時点（2026年9月）で公式ドキュメント上「Pre-GA Offerings Terms」の対象と明記されており、正式なGA機能ではない点に注意（サポートが限定的な場合がある）
- **落とし穴**: Worker Poolsは自動スケールもゼロスケールもしないため、常時起動コストがかかる。ある分析では、最小構成のGPU Worker Poolsでも自前のHetznerサーバーの約4倍のコストになるとの試算あり（推奨スペックだとさらに差が拡大）

## 6. 主な制限事項

| 項目 | 上限 |
|---|---|
| メモリ | 最大32GiB（最小512MiB, gen2） |
| CPU | 最大8 vCPU（最小0.08 vCPU, 1未満は0.001刻み） |
| リクエストタイムアウト（Services） | デフォルト300秒、最大3600秒（1時間） |
| ジョブタスクタイムアウト | 最大168時間（7日）、GPU使用時は最大1時間 |
| ジョブの最大タスク数 | 10,000（1ジョブあたり）、リトライは最大10回 |
| 同時実行数（インスタンスあたり） | 最大1,000（デフォルトはvCPU数×80） |
| コンテナ起動タイムアウト | 4分 |
| 環境変数・コマンド引数 | 各1,000個まで（サービス/ジョブ単位） |
| ファイルシステム | コンテナのファイルシステムは使い捨て（disposable）。永続化にはCloud StorageやNFS連携が必要 |

**Cloud Runが適している条件**（公式ドキュメントが定義する4条件を全て満たす場合）:
1. HTTP/HTTP2/WebSocket/gRPCで配信されるリクエスト・ストリーム・イベントを処理する、または最後まで実行し切るアプリ
2. ローカルの永続ファイルシステムを必要としない（ローカルエフェメラル or ネットワークファイルシステムで動作可能）
3. アプリの複数インスタンスが同時実行される前提で設計されている
4. インスタンスあたりCPU 8個・メモリ32GiB以内に収まる

これらを満たさない場合はGKEが推奨され、サーバーサイドレンダリングのNext.js/AngularアプリにはFirebase App Hostingも選択肢として案内されている。

## 7. AWSサービスとの比較

### 7.1 該当するAWSサービス

Cloud Runは単一のAWSサービスに1対1で対応するわけではなく、**AWS Fargate（ECS/EKS）とAWS Lambdaの中間**に位置する存在として説明されることが多い。コンポーネント単位で見ると対応関係がより明確になる。

> **⚠️ 重要な最新動向（2026年）**: 従来最も近いサービスとされてきた**AWS App Runnerは2026年4月30日付で新規顧客の受付を終了し、事実上のメンテナンスモード入り**した（既存顧客は利用継続可・新規サービス作成も可能だが、AWSは「新機能は追加しない」と明言）。背景には2025年11月に登場した**Amazon ECS Express Mode**との機能重複があり、AWSは移行先としてECS Express Modeを公式に推奨している。そのため現時点でCloud Run Servicesに最も近い「今後も進化が続く」AWSサービスは**ECS Express Mode**と考えるのが実態に近い。

| Cloud Runコンポーネント | 対応するAWSサービス |
|---|---|
| Services（リクエスト駆動コンテナ） | **Amazon ECS Express Mode**（2025年11月登場、現在最も近い）。旧来は**AWS App Runner**が近いとされてきたが2026年4月にメンテナンスモード入り。いずれも該当しない場合は**AWS Fargate（ECS通常モード）** |
| Jobs（バッチ・単発タスク） | **AWS Batch**、**ECS/Fargateのスケジュールタスク** |
| Worker Pools（常駐プル型処理） | **ECS/Fargateサービス**（常時稼働ワーカー）、**EC2 + Auto Scaling** |
| Functions（イベント駆動） | **AWS Lambda** |
| Instances（永続単一インスタンス） | 直接対応するマネージドサービスはなし。強いて言えば**軽量EC2インスタンス**（個別アドレス指定・単一インスタンス管理という点で近い） |

### 7.2 実行モデル・同時実行の比較

| 観点 | AWS Lambda | AWS App Runner（メンテナンスモード） | Amazon ECS Express Mode | AWS Fargate（ECS通常モード） | Google Cloud Run |
|---|---|---|---|---|---|
| 実行単位 | 関数（1リクエスト=1実行） | コンテナ | コンテナ（Fargate上で自動実行） | コンテナ（タスク） | コンテナ |
| デプロイの手間 | 関数コードのみ | Gitリポジトリ指定のみ | イメージ＋IAMロール2つのみ、1 API呼び出しでデプロイ（3〜5分） | ALB・ターゲットグループ・クラスタ等を自前構築 | ソース or イメージ指定のみ |
| ビルトインLB | 不要（内部で処理） | あり | あり（自動プロビジョニング） | **なし**（自前構築） | **あり**（内蔵、追加構成不要） |
| インスタンスあたり同時リクエスト数 | 1（多数のインスタンスを並列起動） | 設定可能（例: 100まで） | ECSの設定に準拠 | 制限なし（自前管理） | 設定可能（最大1,000、デフォルトはvCPU×80） |
| スケール速度 | 1秒未満で新規インスタンス起動 | 1分未満でスケール | 自動スケーリング設定に準拠 | メトリクスベースでスケール（自前設定） | 高速（リクエストベースで自動、最大1,000インスタンス） |
| ゼロスケール | ○（デフォルト） | △（アイドル時割引価格あり） | ×（最小タスク数1以上が前提） | ×（自分で管理しない限り常時稼働） | ○（デフォルト、min-instancesで無効化可） |
| 最大実行時間 | 15分 | 実質無制限（常時起動） | 実質無制限（常時起動） | 無制限（常時起動） | Services: 最大1時間 / Jobs: 最大7日 |
| GPU対応 | **非対応**（GPUリソースタイプ自体が存在しない。GPUが必要ならECS/Batch/SageMaker等と組み合わせる） | なし | 非対応（Fargate基盤のため） | **Fargate自体は非対応**（AWS公式FAQも明記。GPUが必要な場合はEC2起動タイプか、新設の「ECS Managed Instances」を使う必要がある＝Fargateとは別物） | あり（Services: 2025年6月GA、Jobs: 2025年10月GA、Worker Pools: 2026年4月GA。NVIDIA L4・Blackwellに対応、いずれもスケールtoゼロ可能） |
| 今後の展望 | 継続開発中 | **新機能追加なし（2026年4月〜）**、既存顧客のみ利用継続可 | 2025年11月登場、活発に開発中 | 継続開発中 | 継続開発中 |

### 7.3 料金モデルの違い

- **Lambda**: 実行時間（ミリ秒単位）の従量課金、未実行時は無課金
- **App Runner**: 秒単位、CPU/メモリベース。リクエスト数に関わらず起動中は同額。アイドル時は割引価格
- **Fargate**: 秒単位、タスクのCPU/メモリベースの固定課金（起動中は常に課金）
- **Cloud Run**: Request-based（リクエスト処理時のみ課金、ゼロスケール可）とInstance-based（インスタンス生存期間全体で課金）を選択可能という点がAWS勢にはない柔軟性

### 7.4 ネットワーキングの違い

Cloud Runは**ロードバランサが標準内蔵**されており追加構成不要な点が大きな差別化ポイント。AWS Fargate（ECS通常モード）ではロードバランサ（ALB）、ターゲットグループ、クラスタを自前で構築する必要があり、構成の手間とコストが増える。Amazon ECS Express ModeはこのギャップをAWS側が埋める形で登場した機能で、ALB・Fargateサービス・自動スケーリング設定を自動プロビジョニングし、Cloud Runに近いマネージド度を提供する。

### 7.5 総括

- **「Gitプッシュ／イメージ指定だけでコンテナを公開したい」「LBもTLSも全部お任せしたい」** という要件では、Cloud Runに最も近いのは現時点で**Amazon ECS Express Mode**（2025年11月登場）。旧来この位置づけだった**AWS App Runnerは2026年4月末で新規顧客受付を終了しメンテナンスモード入り**しており、AWS自身がExpress Modeへの移行を推奨している
- **「本格的なコンテナオーケストレーション、柔軟なネットワーク制御が必要」**という要件ではAWS Fargate（ECS通常モード/EKS）が対応する。Cloud Run側の対応物はGKEになる
- **「イベント駆動の軽量関数」**を求めるならLambdaが直接の対抗馬（Cloud Run Functionsも同等）
- Google自身の設計思想は「Fargateの柔軟性とLambdaの運用の手軽さ・スケール性の中間」を狙ったものという評もある
- AWS側のPaaS層（App Runner）が縮小し、GCP側（Cloud Run）はWorker Pools・Instances・GPU対応と機能拡張を続けている点で、2026年時点では両社の投資姿勢に明確な差が見える

## 8. Cloud Runが向いているワークロード

- **ステートレスなHTTP/gRPC API・Webアプリケーション**（オンラインショッピング、SNS通知サービス、モバイルバックエンドAPIなど）
- **Webhook受信エンドポイント**（決済サービスからの通知受け口など）
- **トラフィックが変動する、または断続的なサービス**（アクセスが波のように増減するAPI）
- **定期・スケジュールバッチ処理**（Jobs機能を利用）
- **コンテナさえあれば動く軽量マイクロサービス**（Kubernetesの学習コストをかけたくないチーム）
- **プロトタイプ〜中規模プロダクションサービス**（迅速なデプロイ、Git連携によるCI/CD構築のしやすさ）
- **GPUを使うがKubernetesクラスタは持ちたくないAI推論・バッチAIパイプライン**（Worker Pools + GPU、embedding生成、画像処理、LLM推論など）
- **キュー消費型のバックグラウンドワーカー**（Pub/Sub・Kafka・Redisタスクキューを継続的にpollするワーカー。Worker Poolsが該当）

## 9. Cloud Runが向いていないワークロード

- **永続的なローカルファイルシステムが必要なアプリ**（コンテナのファイルシステムは使い捨てのため、ステートフルなワークロードには不向き。NFS/Cloud Storage連携で代替は可能だが根本的にファイルシステム永続を前提とする設計には不向き）
- **リクエストを伴わないバックグラウンドアクティビティが中心の処理**（リクエストを処理していないインスタンスのCPU使用率は実質無視されるため、CPU使用率ベースのオートスケールが機能しにくい。Worker Poolsで一部緩和されるが、自動スケールしない制約がある）
- **8 vCPU・32GiBメモリを超える大規模リソースを必要とするワークロード**（GKEなど別サービスへの移行が必要）
- **常時稼働かつ高負荷・安定トラフィックで、コスト最適化を厳密に行いたい大規模システム**（Request-basedは向かず、Instance-basedにしても専有インスタンス運用よりコスト高になりやすい。GPU Worker Poolsは自動スケール・ゼロスケールがないため、常時起動コストが自前サーバーの数倍になるケースも報告あり）
- **細かいネットワーク制御・サービスメッシュ・カスタムルーティングを多用する複雑なマイクロサービス基盤**（可能だがコールドスタート増加などのトレードオフがあり、GKE + Istio等のほうが素直な場合が多い）
- **起動遅延（コールドスタート）が許容できないレイテンシシビアな用途**（gen2利用時は特に顕著。min-instancesで緩和は可能だがコストとのトレードオフ）
- **コスト予測が重要な予算管理下でのプロジェクト**（リクエストベース課金は伸縮性が高い反面、トラフィックパターン次第で見積もりが難しくなりがち）

## 10. デプロイ手順（Dockerfile / ソースコードから）

Cloud Runへのデプロイには大きく分けて「ソースからデプロイ（ビルドをGoogle側に任せる）」と「イメージを指定してデプロイ（ビルドは自前で行う）」の2パターンがある。**KubernetesでよくあるDeployment/Service用の`service.yaml`は使用しない**（Cloud Run独自の宣言的YAML形式は別途存在するが、一般的なコンテナチュートリアルのservice.yamlとは無関係）。

### 10.1 パターンA: ソースコードから直接デプロイ（`--source`）

```bash
gcloud run deploy SERVICE_NAME --source .
```

- `--image`も`--source`も指定しない場合、デフォルトで`--source`扱いになる
- **Dockerfileがある場合**: そのDockerfileの内容に従ってビルドされる
- **Dockerfileがない場合**: Google Cloud buildpacksが言語（Go, Node.js, Python, Java/Kotlin/Groovy/Scala, .NET, Ruby, PHP）を自動検出し、依存関係を解決してGoogle管理の安全なベースイメージから自動でコンテナを生成する
- 裏側ではCloud Buildが動き、Artifact Registryの`cloud-run-source-deploy`リポジトリにイメージが自動保存される
- 初回はAPI有効化などのプロンプトに`y`で応答する必要あり
- デプロイ完了後、そのリビジョンにトラフィック100%が割り当てられる

**ビルドを伴わないデプロイ（プレビュー、`--no-build`）**: 事前ビルド済みのアーティファクトをCloud Buildを経由せず直接デプロイし高速化する方式。ソースアーカイブは250MiB以下、x86互換のバイナリ/スクリプトである必要があり、依存関係は事前にローカルでvendor化しておく必要がある。

```bash
gcloud beta run deploy SERVICE_NAME \
  --source APPLICATION_PATH \
  --no-build \
  --base-image=nodejs24 \
  --command=COMMAND \
  --args=ARG
```

### 10.2 パターンB: イメージをビルドしてから指定デプロイ（`--image`）

本番運用ではこちらが主流。ビルド工程を完全に自前でコントロールできる（マルチステージビルド、カスタムベースイメージ、ビルド時シークレットなど）。

```bash
# 1. イメージをビルド（ローカル docker build / Cloud Build どちらでも可）
gcloud builds submit --tag us-docker.pkg.dev/PROJECT_ID/REPO/IMAGE:v1.0.0

# 2. Artifact Registry上のイメージを指定してデプロイ
gcloud run deploy SERVICE_NAME \
  --image us-docker.pkg.dev/PROJECT_ID/REPO/IMAGE:v1.0.0 \
  --region us-central1
```

この場合Cloud Run側ではビルドは一切発生しない。

### 10.3 CI/CDパイプライン化（`cloudbuild.yaml`）

継続的デプロイを組む場合は`cloudbuild.yaml`にビルド〜プッシュ〜デプロイの手順を定義し、Cloud Buildのトリガー（GitHub連携等）に紐付けるのが一般的。

```yaml
steps:
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'us-docker.pkg.dev/$PROJECT_ID/repo/image:$COMMIT_SHA', '.']
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'us-docker.pkg.dev/$PROJECT_ID/repo/image:$COMMIT_SHA']
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: gcloud
    args:
      - run
      - deploy
      - SERVICE_NAME
      - --image=us-docker.pkg.dev/$PROJECT_ID/repo/image:$COMMIT_SHA
      - --region=us-central1
```

Cloud BuildのトリガーにこのYAMLファイルを指定しておけば、ソースリポジトリへのpushをきっかけに自動でビルド・デプロイされる。

## 11. Docker Composeからのデプロイ（`gcloud run compose up`）

2026年3月にGA。既存の`compose.yaml`をほぼそのまま使い、`docker compose up`感覚でCloud Runにデプロイできる機能。

```bash
gcloud run compose up compose.yaml
```

- ソースからビルドする場合は`compose.yaml`に`build`フィールドを指定
- 既成イメージを使う場合は`image`フィールドを指定
- リモートのイメージ/リポジトリが変わったがローカルソースは変わっていない場合など、再ビルドを強制したい時は`--build`フラグを付与（`--build`と`--no-build`は併用不可）
- 古いgcloudバージョンではコマンド自体が存在しないため`gcloud components update`が必要

### 11.1 compose.yamlフィールドの変換ルール（主なもの）

| Composeフィールド | Cloud Runへの変換内容 |
|---|---|
| `services` | 各サービスが個々のコンテナに対応 |
| `build` | ソースからビルドし、Artifact Registryの`cloud-run-source-deploy`リポジトリへ自動プッシュ |
| `image` | Docker HubまたはArtifact Registry上の既成イメージをそのままデプロイ |
| `ports` | どのコンテナが外部からのingressを受けるかを決定 |
| `expose` | 外部非公開だが、サービス間通信用にポートを確保 |
| `depends_on` | コンテナの起動順序を定義 |
| `cpus` | CPU/メモリ割り当てのヒントとして使用 |
| `environment` | 環境変数としてコンテナに渡す |
| `command` / `entrypoint` | デフォルトコマンドを上書き |
| `secrets` | Secret Managerへ自動プロビジョニング（`SERVICE_NAME-REGION-SECRET_NAME`の命名規則） |
| `configs` / `volumes` | Cloud Storageバケットへ自動プロビジョニング（部分対応） |
| `x-google-cloudrun:ingress-container`（独自拡張） | 該当コンテナを外部トラフィックの受け口に指定 |
| `x-google-cloudrun:volume-type: in-memory`（独自拡張） | ボリュームをCloud Storageではなくインメモリにする |

同一インスタンス内の複数コンテナ間は、Cloud Runが各コンテナの`/etc/hosts`にエントリを追加することでサービス名による名前解決・相互通信を実現している。デプロイ時にはサービスIDに対して、Cloud Storageバケット用の`roles/storage.objectUser`、Secret Manager用の`roles/secretmanager.secretAccessor`などの必要なIAMロールが自動付与される。

### 11.2 GPU/AIワークロード対応

Docker Model Runnerの`models:`トップレベルキーを使うと、`gcloud run compose`がGPUバックエンドのCloud Runサービスに自動変換してくれる。デフォルトではNVIDIA L4（24GB VRAM、最も低コストな推論用GPU）が使われ、7B〜14BパラメータモデルのFP16/INT8/FP8推論に最適化されている。

### 11.3 制限事項・注意点

- Composeでデプロイされた複数コンテナは、**単一のCloud Runサービス**（マルチコンテナ）としてまとめてデプロイされる（Composeの各serviceが個別のCloud Runサービスになるわけではない）
- サポートされるのはCloud Run機能の**サブセットのみ**（全機能が使えるわけではない）
- Composeで作成したサービスはデフォルトで**最大インスタンス数が1**に制限される
- `docker compose`の`compose.override.yml`自動マージのような機能は**非対応**。`gcloud run compose`は単一の`COMPOSE_FILE`引数のみ受け付ける（`docker compose --file`のような複数ファイル指定・マージは不可）
- イメージの取得元レジストリに制限あり（例: GHCR＝GitHub Container Registryは非対応）
- 本番運用における包括的なIaC戦略の代替にはならないとされ、恒久的な運用にはTerraformなどの利用が推奨されている。一部レビューでは「本番投入というより開発中の簡易確認向け」という評価もある
- GPU（`models:`）を使うサービスを含む場合、そのスタック全体がGPU対応リージョンにしかデプロイできない

## 参考リンク

- [What is Cloud Run（公式）](https://docs.cloud.google.com/run/docs/overview/what-is-cloud-run)
- [Cloud Run fit-for-run（適合条件, 公式・日本語）](https://docs.cloud.google.com/run/docs/fit-for-run?hl=ja)
- [Cloud Run pricing（公式）](https://cloud.google.com/run/pricing)
- [Billing settings for services（公式）](https://docs.cloud.google.com/run/docs/configuring/billing-settings)
- [GPU support for worker pools（公式）](https://docs.cloud.google.com/run/docs/configuring/workerpools/gpu)
- [Cloud Run release notes（公式）](https://docs.cloud.google.com/run/docs/release-notes)
- [Cloud Run Worker Pools Hit GA With Blackwell GPUs（bex.co, 2026-07）](https://bex.co/blog/2026/07/12/cloud-run-worker-pools-blackwell-gpu-vs-hetzner)
- [Google Cloud Run vs AWS ECS Fargate（Medium）](https://medium.com/@o.hanhaliuk/google-cloud-run-vs-aws-ecs-fargate-2bcc49f0dd46)
- [AWS App Runner alternatives（Northflank, 2026）](https://northflank.com/blog/aws-app-runner-alternatives)
- [Choosing the Right AWS Compute Service（Thoughtful Architect）](https://www.thoughtfularchitect.dev/posts/aws-compute-comparison)
- [Deploy services from source code（公式）](https://docs.cloud.google.com/run/docs/deploying-source-code)
- [Deploy services using Compose（公式）](https://docs.cloud.google.com/run/docs/deploy-run-compose)
- [gcloud run compose upでGPUスタックをデプロイ（Medium）](https://medium.com/google-cloud/gcloud-run-compose-up-deploy-a-multi-service-gpu-stack-to-cloud-run-from-docker-compose-77d650b39972)
- [This is Cloud Run: Nine Ways to Deploy（Medium）](https://medium.com/google-developer-experts/this-is-cloud-run-nine-ways-to-deploy-and-when-to-use-each-72661f7bb6db)
- [AWS App Runner availability change（AWS公式・日本語）](https://docs.aws.amazon.com/ja_jp/apprunner/latest/dg/apprunner-availability-change.html)
- [AWS Ends WorkMail and Moves App Runner to Maintenance Mode（InfoQ, 2026-04）](https://www.infoq.com/news/2026/04/aws-deprecates-workmail-apprunne/)
- [The End of AWS App Runner: What It Means for Your Apps（Encore）](https://encore.dev/articles/end-of-app-runner)
- [Cloud Run execution environments（公式・gen1/gen2選択指針）](https://docs.cloud.google.com/run/docs/configuring/execution-environments)
- [GPU support for services（公式）](https://docs.cloud.google.com/run/docs/configuring/services/gpu)
