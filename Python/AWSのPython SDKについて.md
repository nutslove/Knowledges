## `boto3.client`と`boto3.Session().client`の違いについて
- ChatGPTからの回答
> `boto3.client` and `boto3.Session().client` are two different ways of creating a client object to interact with AWS services. Let's look at them more closely:
> 
> 1. `boto3.client(service_name, region_name=None, api_version=None, use_ssl=True, verify=None, endpoint_url=None, aws_access_key_id=None, aws_secret_access_key=None, aws_session_token=None, config=None)`: 
This function returns a low-level, service-specific client object. You can create a client for any AWS service using the service name, like 's3' for S3 service. It will use the default AWS credentials found by `boto3` (using the AWS CLI, environment variables, IAM roles, etc.), but you can also specify AWS access keys manually.
> 
> 2. `boto3.Session().client(service_name, region_name=None, api_version=None, use_ssl=True, verify=None, endpoint_url=None, aws_access_key_id=None, aws_secret_access_key=None, aws_session_token=None, config=None)`: 
This function is essentially the same as `boto3.client`, but it is created from a session object. A session in `boto3` allows you to manage state, like the AWS region or credentials. By creating a session object, you can have fine-grained control over how your application interacts with AWS. You can customize sessions using different profiles (as set in your AWS configuration file) or manually specify credentials, among other things. 
> 
> In general, if you need to manage state or use different profiles for different parts of your application, using `boto3.Session().client` might be the better option. If your application only uses a single set of AWS credentials, `boto3.client` should be sufficient.
> 
> Remember that both `boto3.client` and `boto3.Session().client` do not establish the actual connections to AWS services when they are called. They just return client objects that are capable of making requests to AWS services. The actual connections are established when you make a request using these client objects.

## Podで複数のEC2にアタッチされているIAM Roleを利用して、configファイルで複数のprofileを切り替えながら処理する方法
- IAM Roleに信頼関係を結んだ (Assume Roleできる) IAM Policyがアタッチされている必要がある
- configファイルを`ConfigMap`として設定し、`os.environ['AWS_PROFILE']`でprofileを切り替えながら処理
- Podのマニフェストファイル
  ~~~yaml
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: django
    namespace: django
    labels:
      name: django
  spec:
    replicas: 1
    selector:
      matchLabels:
        name: django
    template:
      metadata:
        labels:
          name: django
          version: 20230527-5
      spec:
        containers:
          - name: django
            image: somestring.dkr.ecr.ap-northeast-1.amazonaws.com/ecr:django
            volumeMounts:
            - name: aws-config
              mountPath: /tmp
            env:
              - name: HTTP_PROXY
                value: http://**.**.**.**
              - name: HTTPS_PROXY
                value: http://**.**.**.**
              - name: NO_PROXY
                value: '127.0.0.1,localhost,169.254.169.254'
              - name: ENV
                value: stg
              - name: AWS_CONFIG_FILE
                value: "/tmp/config"
            imagePullPolicy: Always
        volumes:
        - name: aws-config
          configMap:
            name: aws-configfile
  ---
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: aws-configfile
    namespace: django
  data:
    config: |
      [profile TAG_HEUER]
      role_arn = arn:aws:iam::901234567890:role/role-01
      credential_source = Ec2InstanceMetadata
      region = ap-northeast-1
      [profile ROLEX]
      role_arn = arn:aws:iam::012345678901:role/role-02
      credential_source = Ec2InstanceMetadata
      region = ap-northeast-1
      [profile OMEGA]
      role_arn = arn:aws:iam::123456789012:role/role-03
      credential_source = Ec2InstanceMetadata
      region = ap-northeast-1
  ~~~
- pythonでROLEX profileのIAMユーザのパスワードを更新
  ~~~python
  import boto3
  import os

  os.environ['AWS_PROFILE'] = "ROLEX"

  session = boto3.Session()
  client = session.client('iam')
  client.update_login_profile(UserName="ME", Password="PASSWORD", PasswordResetRequired=True)
  ~~~