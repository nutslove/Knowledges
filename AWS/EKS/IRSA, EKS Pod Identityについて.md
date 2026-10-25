# *Pod Identity* vs *IRSA (IAM roles for service accounts)*
- https://speakerdeck.com/xblood/amazon-eks-pod-identityde-he-gabian-warunoka
- https://docs.aws.amazon.com/eks/latest/userguide/service-accounts.html
- PodにIAM権限を付与する仕組みとしてPod IdentityとIRSAがある
- AWSはPod Identityの利用を推奨している
  > AWS recommends using EKS Pod Identities to grant access to AWS resources to your pods whenever possible.
- IRSAはEKS以外（EKS Anywhere、ROSA、EC2上の自前k8sなど）でも使える汎用的な仕組みで、EKSのAPIに依存しない。一方Pod IdentityはEKS専用で、EKS APIだけで設定が完結する

| 項目 | Pod Identity | IRSA |
|---|---|---|
| 仕組み | EKS Pod Identity Agent（DaemonSet）が、EKS Auth APIで取得した認証情報をSDKに配布する（Container credential provider） | webhookが注入したService Accountトークン（OIDC）をSDKが`sts:AssumeRoleWithWebIdentity`でSTSに渡して認証情報を得る |
| 事前準備 | `eks-pod-identity-agent`アドオンのインストール（EKS Auto Modeでは不要） | クラスターごとにIAM OIDC Providerの作成 |
| IAM Roleの信頼ポリシー | `pods.eks.amazonaws.com`を許可（クラスターに依存しない。一度設定すれば使い回せる） | OIDC Providerを`Federated`で指定し、`sub`/`aud`条件でnamespace/SAを絞る（クラスターごとに更新が必要） |
| Service Accountへの設定 | 不要（EKS APIのPod Identity Associationで関連付け。k8s側にassociationの情報は残らない） | `eks.amazonaws.com/role-arn`アノテーションが必要 |
| 複数クラスターでのRole再利用 | 同じRoleをそのまま使い回せる | クラスターごとに信頼ポリシーの更新が必要 |
| セッションタグ（ABAC） | 対応（cluster名、namespace、SA名などが自動でタグ付けされる） | 非対応（Roleを共有するとPodは全て同じ権限を持つ） |
| クラスター数の上限 | OIDC Providerが不要なので関係なし | IAMのOIDC Providerは1アカウントあたりデフォルト100個まで |
| Roleあたりの信頼関係数 | 信頼ポリシーに個別のSAを書かないので関係なし | 信頼ポリシーの長さ制限（デフォルト2048文字）により、1ポリシーに書けるのは4個程度（上限緩和しても最大8個程度） |
| STS APIクォータ | EKS側がAssumeRoleするので、PodからSTSを直接呼ばない | SDKがPodごとにSTSを呼ぶ。実装次第で不要な呼び出しが増え、スロットリングされる可能性がある |
| クロスアカウント | `--target-role-arn`でロールチェーンをEKSが自動実行 | 相手アカウントにOIDC Providerを作る、または自分でロールチェーンを設定する |
| 対応環境 | Linux EC2ノードのみ。Fargate、Windows EC2、Outposts、EKS Anywhere、EC2上の自前k8sは**非対応** | EKS（EC2、Fargate）、EKS Anywhere、ROSA、EC2上の自前k8sなど（Outpostsのローカルクラスターは非対応） |
| EKSバージョン | 全サポートバージョン（1.28はプラットフォームバージョン`eks.4`以降） | 全サポートバージョン |

# Pod Identity
- https://techblog.ap-com.co.jp/entry/eks-pod-identity-deep-dive
- https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/pod-id-association.html
- https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html

## 仕組み
- Pod Identity Agent（DaemonSet）と、追加された**EKS Auth API**で認証情報を配布する
- Podが作成される時、関連付け（association）のあるService Accountを使っているPodに以下が自動で追加される  
  ```yaml
  env:
  - name: AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE
    value: "/var/run/secrets/pods.eks.amazonaws.com/serviceaccount/eks-pod-identity-token"
  - name: AWS_CONTAINER_CREDENTIALS_FULL_URI
    value: "http://169.254.170.23/v1/credentials"
  volumes:
  - name: eks-pod-identity-token
    projected:
      sources:
      - serviceAccountToken:
          audience: pods.eks.amazonaws.com
          expirationSeconds: 86400 # 24 hours
          path: eks-pod-identity-token
  ```
- Podが動くノード上のPod Identity Agentが、EKS Auth APIの`AssumeRoleForPodIdentity`で一時的な認証情報を取得し、SDKに渡す
- SDKはこの環境変数を見て**Container credential provider**経由でAgentから認証情報を取得する
  - **デフォルトのクレデンシャルチェーンでそれより前にある認証情報（環境変数のアクセスキーなど）が見つかると、そちらが優先される**。Pod Identityへ移行する時は先にassociationを作り、その後で古い認証情報を外せば安全
- 各Podではなくノード単位でAgentが認証情報を扱うので、STSへの負荷はノードあたり1回に抑えられる

## Pod Identity Agentの導入
- **EKS Auto Modeの場合は組み込み済みなので不要**
- アドオンとしてインストールする（クラスターごとに1回）  
  ```shell
  aws eks create-addon --cluster-name <クラスター名> --addon-name eks-pod-identity-agent
  ```
  - `service-account-role-arn`（IRSA用の設定）は使わない。AgentにはノードのIAM Roleで権限を与える
- インストール確認  
  ```shell
  kubectl get pods -n kube-system | grep 'eks-pod-identity-agent'
  ```
- 前提条件
  - **ノードのIAM Roleに`eks-auth:AssumeRoleForPodIdentity`の権限が必要**。AWS管理ポリシーの`AmazonEKSWorkerNodePolicy`に含まれている  
    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": ["eks-auth:AssumeRoleForPodIdentity"],
                "Resource": "*"
            }
        ]
    }
    ```
  - ノードがEKS Auth APIに到達できること。**プライベートサブネットのノードの場合はEKS Auth API用のAWS PrivateLinkインターフェースエンドポイントが必要**
  - ノードがECRからAgentのイメージをpullできること
- Agentの仕様
  - `hostNetwork`を使い、リンクローカルアドレス`169.254.170.23`（IPv4）と`[fd00:ec2::23]`（IPv6）のポート`80`と`2703`を使う
  - **IPv6を無効にしているノードではAgentが起動しない**。その場合はIPv4だけで待ち受けるよう設定する  
    ```shell
    aws eks update-addon --cluster-name <クラスター名> --addon-name eks-pod-identity-agent \
        --resolve-conflicts PRESERVE --configuration-values '{"agent":{"additionalArgs": { "-b": "169.254.170.23"}}}'
    ```
  - Podでプロキシを使っている場合は、`169.254.170.23`と`[fd00:ec2::23]`を`NO_PROXY`に追加する
  - Security Groups for PodsをVPC CNIで使う場合は、`ENABLE_POD_ENI=true`、`POD_SECURITY_GROUP_ENFORCING_MODE=standard`が必要

## Pod Identityの使い方
## IAM Role作成
- 以下のように`Principal.Service`に`"pods.eks.amazonaws.com"`を、`Action`に`"sts:AssumeRole"`と`"sts:TagSession"`を与えたIAM Roleを作成する  
  ```json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Sid": "AllowEksAuthToAssumeRoleForPodIdentity",
              "Effect": "Allow",
              "Principal": {
                  "Service": "pods.eks.amazonaws.com"
              },
              "Action": [
                  "sts:AssumeRole",
                  "sts:TagSession"
              ]
          }
      ]
  }
  ```
  - `sts:AssumeRole`: Podに認証情報を渡す前にEKS Pod IdentityがRoleをAssumeするために必要
  - `sts:TagSession`: セッションタグを付けるために必要。このタグを信頼ポリシーの条件キーで使うと、使えるService Account、namespace、クラスターを絞れる
- 作成したIAM RoleにPodに必要なIAM Policy（e.g. S3FullAccess）をアタッチする

## Service Accountの作成
- IAM Roleとマッピングするk8sのService Accountを作成する（**IRSAと違い、アノテーションは不要**）  
  ```yaml
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: thanos-service-account
    namespace: monitoring
  ```

## IAM RoleとService Accountを関連付ける
- 関連付けを作る側のIAM principalに`iam:PassRole`が必要
- 1つのService Accountに関連付けられるRoleは、クラスターと同じアカウントの1つだけ（変更したい場合はassociationを編集する）
- Terraformの場合（https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_pod_identity_association）  
  ```
  resource "aws_eks_pod_identity_association" "s3" {
      cluster_name    = aws_eks_cluster.platform_cluster_auto_mode.name
      namespace       = "monitoring"
      service_account = "thanos-service-account"
      role_arn        = var.s3_iam_role_for_pod_arn
  }
  ```

- AWS CLIの場合  
  ```shell
  aws eks create-pod-identity-association --cluster-name my-cluster --role-arn arn:aws:iam::111122223333:role/my-role --namespace default --service-account my-service-account
  ```
- namespaceとService Accountは、クラスター内にまだ存在しない名前でも指定できる（associationが機能するには、namespace、Service Account、それを使うワークロードを作っておく必要がある）
- associationは結果整合で、API成功後に反映されるまで数秒かかることがある。アプリの重要なコードパスの中でassociationを作成・更新しない

## 関連づけたPod identity association確認方法
- Pod identity association 一覧確認  
  ```shell
  aws eks list-pod-identity-associations --cluster-name <EKSクラスター名>
  ```
- associationごとの付いているIAM Roleなどの確認  
  ```shell
  aws eks describe-pod-identity-association --cluster-name <クラスター名> --association-id <アソシエーションID>
  ```

## Podへの反映
- 環境変数とトークンのvolumeは**Pod作成時に注入される**ので、association作成前から動いているPodには反映されない。作成後にPodを作り直す必要がある（`kubectl rollout restart deployment/<名前>`など）
  - これはAWSのブログ・GitHub issueの記述に基づく内容で、公式ユーザーガイドには明記されていない
- すでにassociationが効いているPodに対してRoleを変更した場合は、Agentが認証情報を自動で更新するので、Podの再作成は不要（即時反映したいなら再作成）
- Podで`serviceAccountName`にそのService Accountを指定する
- 注入されているか確認  
  ```shell
  kubectl describe pod <pod名> | grep AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE:
  ```
- ワークロードがそのService Accountを使っているか確認  
  ```shell
  kubectl describe deployment <deployment名> | grep "Service Account"
  ```

## SDKの最低バージョン
- Container credential providerにPod Identityが対応したバージョン以降のSDKが必要。かつ、認証情報のプロバイダーを明示せず**デフォルトのクレデンシャルチェーン**を使う（または、container credential providerを指定する）こと
  - Java v2: 2.21.30 / Java v1: 1.12.746
  - Go v1: v1.47.11 / Go v2: release-2023-11-14
  - Python: boto3 1.34.41、botocore 1.34.41
  - AWS CLI: v1は1.30.0、v2は2.15.0
  - JavaScript: v2は2.1550.0、v3はv3.458.0
  - Kotlin: v1.0.1 / Ruby: 3.188.0 / Rust: release-2024-03-13
  - C++: 1.11.263 / .NET: 3.7.734.0 / PowerShell: 4.1.502 / PHP: 3.289.0
- 最新は https://docs.aws.amazon.com/eks/latest/userguide/pod-id-minimum-sdk.html を参照

## セッションタグ（ABAC）
- Pod IdentityがRoleをAssumeする時、以下のセッションタグが自動で付与される（何も設定しなくてよい）
  - `eks-cluster-arn`
  - `eks-cluster-name`
  - `kubernetes-namespace`
  - `kubernetes-service-account`
  - `kubernetes-pod-name`
  - `kubernetes-pod-uid`
- IAM PolicyやS3バケットポリシーなどで`${aws:PrincipalTag/kubernetes-namespace}`のように参照できる。1つのRoleを複数のService Accountで共有しつつ、タグで権限を分けられる  
  ```json
  {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:GetObjectTagging"],
      "Resource": "*",
      "Condition": {
          "StringEquals": {
              "s3:ExistingObjectTag/eks-cluster-name": "${aws:PrincipalTag/eks-cluster-name}"
          }
      }
  }
  ```
- これらのタグは**transitive**で、Podの中で別アカウントのRoleにAssumeRoleした時にも引き継がれる
- 独自のカスタムタグをPod Identityから追加することはできない。IAM Roleに付けたタグは`${aws:PrincipalTag/<キー>}`で参照できる。同じキーが競合した場合はPod Identityが付けたセッションタグが優先される
- `PackedPolicyTooLarge`エラーが出る場合は、associationの`--disable-session-tags`（コンソールでは**Disable session tags**）でセッションタグを無効にするとサイズを減らせる

## セッションポリシー
- associationに`--policy`でIAMポリシー（JSON文字列）を指定すると、Roleにアタッチされたポリシーとの**積集合**に権限を絞れる。1つのRoleを複数アプリで共有しつつ最小権限にしたい時に使う
- **`disableSessionTags`を`true`にする必要がある**（セッションタグと併用できない）
- `targetRoleArn`と併用した場合、ポリシーの制限が適用されるのはターゲットRole側の権限だけ（最初のRoleには適用されない）

## クロスアカウントアクセス
- 2025年6月に追加された`targetRoleArn`（CLIでは`--target-role-arn`）を使うと、EKSが2段階のAssumeRole（ロールチェーン）を自動で行い、別アカウントのRoleの認証情報をPodに渡す
  - 1段階目: associationの`roleArn`（**クラスターと同じアカウント**のRole）をAssumeする
  - 2段階目: その認証情報で`targetRoleArn`のRole（同じまたは別アカウント）をAssumeする
- 必要なIAMの設定
  - `roleArn`側のRoleに、ターゲットRoleに対する`sts:AssumeRole`と`sts:TagSession`を許可する権限ポリシーを付ける
  - ターゲットRole側は、`roleArn`のRoleを信頼する信頼ポリシーを設定する（通常のクロスアカウントのロールチェーンと同じ）
  ```shell
  aws eks create-pod-identity-association --cluster-name my-cluster --namespace default --service-account my-service-account \
      --role-arn arn:aws:iam::111122223333:role/my-role \
      --target-role-arn arn:aws:iam::444455556666:role/target-role
  ```
- https://aws.amazon.com/blogs/containers/amazon-eks-pod-identity-streamlines-cross-account-access

## 制約・注意点
- 1クラスターあたり最大5,000個のassociation
- 対応するのは**Linux EC2ノードのPodだけ**。Fargate（LinuxもWindowsも）、Windows EC2ノード、Outposts、EKS Anywhere、EC2上の自前k8sでは使えない
  - Fargateで使いたい場合はIRSAを使う
- IMDSへのアクセスを制限していないと、PodのコンテナはノードのIAM Role（と、場合によっては同一ノード上の他PodのRoleの認証情報）にもアクセスできてしまう。Pod間の認証情報を分離したいなら、IMDSを制限する
  - `hostNetwork: true`のPodは常にIMDSにアクセスできる。ただしSDKとCLIはPod Identityが有効ならそちらの認証情報を使う
- コンテナはセキュリティの境界ではない。同じノードのPodはカーネルを共有する

# IRSA
- https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/iam-roles-for-service-accounts.html
- https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html

## 仕組み
- EKSクラスターごとにOIDC Issuer URLがあり、EKSがOIDCディスカバリーエンドポイント（署名鍵の公開）をホストしている。IAMにこのURLを**IDプロバイダー（OIDC Provider）**として登録しておくと、IAMがKubernetesが発行したService Accountトークン（JWT）を検証できる
- Podは`ProjectedServiceAccountToken`（audience・有効期限付きのOIDCトークン）を使い、SDKがそれを`sts:AssumeRoleWithWebIdentity`でSTSに渡して一時的な認証情報を得る
- Pod作成時に**Amazon EKS Pod Identity Webhook**（mutating admission webhook）が、`eks.amazonaws.com/role-arn`アノテーション付きのService Accountを使うPodに以下を自動で注入する
  - 環境変数`AWS_ROLE_ARN`、`AWS_WEB_IDENTITY_TOKEN_FILE`（`/var/run/secrets/eks.amazonaws.com/serviceaccount/token`）
  - Service Accountトークンをprojected volumeとしてマウント
  - ※名前が紛らわしいが、これはIRSA用のwebhookで、EKS Pod Identityとは別物
- トークンはkubeletが取得・保存し、トークンの有効期間の80%を超えるか24時間より古くなると更新される
- ワークロード側でwebhookを使わず、同じ環境変数とトークンファイルのマウントを手動で設定することも可能
- 署名鍵のペアは7日ごとにローテーションされる。外部のOIDCクライアントが署名鍵を検証に使う場合は、公開鍵が失効する前に更新する必要がある

## IAM OIDC Providerの作成
- クラスターごとに1回だけ必要
- OIDC Issuer URLの確認  
  ```shell
  aws eks describe-cluster --name <クラスター名> --query "cluster.identity.oidc.issuer" --output text
  ```
- 既にIAM OIDC Providerがあるかの確認（issuer IDで絞る）  
  ```shell
  cluster_name=<クラスター名>
  oidc_id=$(aws eks describe-cluster --name $cluster_name --query "cluster.identity.oidc.issuer" --output text | cut -d '/' -f 5)
  aws iam list-open-id-connect-providers | grep $oidc_id | cut -d "/" -f4
  ```
  - 出力があれば作成済み
- 作成（eksctlの場合）  
  ```shell
  eksctl utils associate-iam-oidc-provider --cluster $cluster_name --approve
  ```
- コンソールの場合: IAMの「IDプロバイダ」で「プロバイダを追加」→ タイプ`OpenID Connect`、プロバイダのURLにOIDC Issuer URL、対象者（Audience）に`sts.amazonaws.com`を指定する
- IAMのOIDC Providerは**1アカウントあたりデフォルト100個まで**。1アカウントに100を超えるEKSクラスターをIRSAで運用する場合は上限に当たる
- クラスターのVPCにアウトバウンドのインターネットアクセスがない場合
  - OIDCのエンドポイントに到達できず、`eksctl`などの操作が`NXDOMAIN`で失敗することがある
  - その場合は`com.amazonaws.<region>.oidc-eks`のVPCインターフェースエンドポイント（プライベートDNS有効）を作るか、VPCの外（CloudShellなど）で実行する

## IAM Role作成
- `Principal.Federated`にOIDC ProviderのARNを、`Action`に`sts:AssumeRoleWithWebIdentity`を指定し、`Condition`で対象のnamespace/Service Accountを絞ったIAM Roleを作成する  
  ```json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Effect": "Allow",
              "Principal": {
                  "Federated": "arn:aws:iam::111122223333:oidc-provider/oidc.eks.ap-northeast-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
              },
              "Action": "sts:AssumeRoleWithWebIdentity",
              "Condition": {
                  "StringEquals": {
                      "oidc.eks.ap-northeast-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:aud": "sts.amazonaws.com",
                      "oidc.eks.ap-northeast-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub": "system:serviceaccount:monitoring:thanos-service-account"
                  }
              }
          }
      ]
  }
  ```
  - `sub`の条件を付けないと、同じクラスター内の任意のService AccountがこのRoleをAssumeできてしまうので必ず指定する
  - `sub`の形式は`system:serviceaccount:<namespace>:<Service Account名>`
  - namespace内の全Service Accountに許可したい場合は`StringEquals`を`StringLike`に変え、SA名を`*`にする（e.g. `system:serviceaccount:monitoring:*`）
  - `StringEquals`/`StringLike`に複数エントリーを書けば、複数のSAやnamespaceを許可できる
- 作成したIAM RoleにPodに必要なIAM Policy（e.g. S3FullAccess）をアタッチする
- eksctlなら、Roleの作成・ポリシーのアタッチ・Service Accountの作成とアノテーションの付与を一括で行える  
  ```shell
  eksctl create iamserviceaccount --name my-service-account --namespace default --cluster my-cluster --role-name my-role \
      --attach-policy-arn arn:aws:iam::111122223333:policy/my-policy --approve
  ```
  - RoleやService Accountが既に存在すると失敗することがある。その場合のオプションは`eksctl create iamserviceaccount --help`で確認

## Service Accountの作成
- `eks.amazonaws.com/role-arn`アノテーションでIAM Roleを指定したService Accountを作成する  
  ```yaml
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: thanos-service-account
    namespace: monitoring
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::111122223333:role/my-role
  ```
- 既存のService Accountに付ける場合  
  ```shell
  kubectl annotate serviceaccount -n monitoring thanos-service-account eks.amazonaws.com/role-arn=arn:aws:iam::111122223333:role/my-role
  ```
- PodのspecでこのService Accountを指定する（`serviceAccountName: thanos-service-account`）
- **webhookは起動済みのPodには適用されない**ので、アノテーションを付けた後は、そのService Accountを使うPodを作り直す必要がある

## STSエンドポイントの設定
- AWSは、グローバルエンドポイントではなく**リージョナルSTSエンドポイントの利用を推奨**している（レイテンシーの低減、冗長性、セッショントークンの有効性の向上）
- Service Accountのアノテーションで切り替える（変更後はPodの作り直しが必要）  
  ```shell
  # リージョナルにする
  kubectl annotate serviceaccount -n <namespace> <SA名> eks.amazonaws.com/sts-regional-endpoints=true
  # グローバルにする
  kubectl annotate serviceaccount -n <namespace> <SA名> eks.amazonaws.com/sts-regional-endpoints=false
  ```
- 設定されている値は`AWS_STS_REGIONAL_ENDPOINTS`環境変数で確認できる
- VPCにインターネット出口がない環境でIRSAを使う場合は、STSのVPCインターフェースエンドポイント（`com.amazonaws.<region>.sts`）を作り、リージョナルエンドポイントを使うようにする（グローバルエンドポイント`sts.amazonaws.com`はVPCエンドポイントを経由しない）

## 動作確認方法
- Podに環境変数とトークンが注入されているか確認  
  ```shell
  kubectl describe pod <pod名> | grep AWS_ROLE_ARN:
  kubectl describe pod <pod名> | grep AWS_WEB_IDENTITY_TOKEN_FILE:
  ```
  - `AWS_ROLE_ARN`が、Service Accountのアノテーションに付けたRoleのARNと一致していること
- Service Accountにアノテーションが付いているか確認  
  ```shell
  kubectl describe serviceaccount <SA名> -n <namespace>
  ```
- IAM Roleの信頼ポリシーの確認  
  ```shell
  aws iam get-role --role-name my-role --query Role.AssumeRolePolicyDocument
  ```
- Pod内で実際にAssumeされているRoleの確認（コンテナにAWS CLIが入っている場合のみ）  
  ```shell
  kubectl exec -n <namespace> <pod名> -- aws sts get-caller-identity
  ```

## SDKの最低バージョン
- web identityトークンファイルで`AssumeRoleWithWebIdentity`できるバージョン以降のSDKが必要（デフォルトのクレデンシャルチェーンに含まれている）
  - Java v2: 2.10.11 / Java v1: 1.12.782（Javaは**`sts`モジュールをclasspathに含める必要がある**）
  - Go v1: 1.23.13 / Go v2: 全バージョン対応
  - Python: boto3 1.9.220、botocore 1.12.200
  - AWS CLI: 1.16.232
  - Node: v2は2.525.0、v3は3.27.0
  - Ruby: 3.58.0 / C++: 1.7.174 / PHP: 3.110.7
  - .NET: 3.3.659.1（**`AWSSDK.SecurityToken`も必要**）
- 最新は https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts-minimum-sdk.html を参照

## クロスアカウントアクセス
- https://docs.aws.amazon.com/eks/latest/userguide/cross-account-access.html
- アカウントAのクラスターのPodが、アカウントBのRoleの権限を使いたい場合
- **方法1: アカウントBにアカウントAのクラスターのOIDC Providerを作る**
  - アカウントBがアカウントAのOIDC Issuer URLでOIDC ProviderとIAM Roleを作り、アカウントAのService AccountのアノテーションにアカウントBのRoleを指定する
  - 簡単だが、アカウントBがOIDC Providerを作成・管理する必要がある
- **方法2: ロールチェーン（`AssumeRole`を2回）**
  - アカウントBのRoleの信頼ポリシーでアカウントAのRole（IRSAのRole）を許可（最小権限にするため`root`ではなく具体的なRoleのARNを指定する）
  - アカウントAのIRSA Roleの権限ポリシーに、アカウントBのRoleへの`sts:AssumeRole`を追加
  - Pod側のAWS CLIプロファイル（`~/.aws/config`）で、`web_identity_token_file`のプロファイルを`source_profile`にしたプロファイルを作る  
    ```
    [profile account_b_role]
    source_profile = account_a_role
    role_arn=arn:aws:iam::444455556666:role/account-b-role

    [profile account_a_role]
    web_identity_token_file = /var/run/secrets/eks.amazonaws.com/serviceaccount/token
    role_arn=arn:aws:iam::111122223333:role/account-a-role
    ```
  - OIDC Providerの管理がアカウントAに残る

## 制約・注意点
- **Fargateで使える**（FargateのPodはIMDSが使えないため、IAMの認証情報が必要な場合はIRSAを使うよう公式が案内している）
- EKS on Outpostsの**ローカルクラスターでは使えない**
- Roleの信頼ポリシーにService Accountごとの条件を書くため、ポリシー長（デフォルト2048文字）の制限で1つのRoleに紐づけられるSAの数は限られる
- SDKが各Podから直接STSを呼ぶので、コードの作りによってはSTSのAPIクォータでスロットリングされることがある
- セッションタグは使えない。Roleをクラスター間で再利用できるが、**そのRoleを使うPodは全て同じ権限を持つ**
- IMDSへのアクセスを制限していないと、PodはノードのIAM Roleの認証情報にもアクセスできてしまう（Pod Identityと同じ）
- コンテナはセキュリティの境界ではない
