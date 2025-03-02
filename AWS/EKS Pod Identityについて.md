- https://techblog.ap-com.co.jp/entry/eks-pod-identity-deep-dive
- https://docs.aws.amazon.com/ja_jp/eks/latest/userguide/pod-id-association.html

# *Pod Identity* vs *IRSA (IAM roles for service accounts)*
- PodにIAM権限を付与する仕組みとしてPod IdentityとIRSAがある
- AWSはPod Identityの利用を推奨している
  > ### EKS Pod Identities compared to IRSA
  > Both EKS Pod Identities and IRSA are preferred ways to deliver temporary AWS credentials to your EKS pods. Unless you have specific usecases for IRSA, we recommend you use EKS Pod Identities when using EKS. This table helps compare the two features.

# Pod Identityの使い方
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
- 作成したIAM RoleにPodに必要なIAM Policy（e.g. S3FullAccess）をアタッチする

## Service Accountの作成
- IAM Roleとマッピングするk8sのService Accountを作成する  
  ```yaml
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: thanos-service-account
    namespace: monitoring
  ```

## IAM RoleとService Accountを関連付ける
- Terraformの場合  
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