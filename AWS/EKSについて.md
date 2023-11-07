- Terraformで指定するAMIリリースバージョンは以下のサイトから確認できる
  - https://github.com/awslabs/amazon-eks-ami/blob/master/CHANGELOG.md

## TerraformでクロスアカウントのEKSクラスターを作成した時、kubectlを打てるようにする方法
- EKSクラスターはデフォルトでは作成したIAMエンティティしかkubectlで操作できない
  - `kube-system`ネームスペースの`aws-auth`という`ConfigMap`にて`mapRoles`でEKSクラスター作成エンティティが設定されている
- `kubectl`がインストールされているサーバで以下コマンドを実行し新しいクラスターを登録する（`~/.kube/config`が更新される）
  ~~~
  aws eks update-kubeconfig --region ap-northeast-1 --name <EKSクラスター名>
  ~~~
- aws cliでEKSクラスターを構築した(EKSクラスター所有者)で一時的な認証情報を取得する
  ~~~
  aws sts assume-role --role-arn "arn:aws:iam::<EKSクラスターがあるAWSアカウントID>:role/<EKSクラスター所有者Role>" --role-session-name EKS-session
  ~~~
  **※`arn:aws:iam::<EKSクラスターがあるAWSアカウントID>:role/<EKSクラスター所有者Role>`はTerraformでST環境のEKSを作成する時に使用したIAMロール**
- 表示される「**AccessKeyId**」と「**SecretAccessKey**」と「**SessionToken**」の値を以下の環境変数として設定する  
※これでkube-apiserverへの認証が通るようになる
  ~~~
  export AWS_ACCESS_KEY_ID=<AccessKeyIdの値>
  export AWS_SECRET_ACCESS_KEY=<SecretAccessKeyの値>
  export AWS_SESSION_TOKEN=<SessionTokenの値>
  ~~~
- `kubectl edit cm -n kube-system aws-auth`コマンドで`mapRoles: |`の下に以下内容を追加
  ~~~yaml
  - groups:
    - system:masters
    rolearn: arn:aws:iam::<EKSクラスターがあるAWSアカウントID>:role/<kubectlを実行しているEC2にアタッチされているIAMロール名>
    username: masteruser
  ~~~
