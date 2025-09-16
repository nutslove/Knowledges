## HTTP API vs REST API
- https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/http-api-vs-rest.html  
  > REST API と HTTP API は、いずれも RESTful API 製品です。REST API は HTTP API よりも多くの機能をサポートしていますが、HTTP API は低価格で提供できるように最小限の機能で設計されています。API キー、クライアントごとのスロットリング、リクエストの検証、AWS WAF の統合、プライベート API エンドポイントなどの機能が必要な場合は、REST API を選択します。REST API に含まれる機能が必要ない場合は、HTTP API を選択します。

### REST API
#### ルートの設定
- https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/getting-started-rest-new-console.html
- HTTP APIは「Routes」タブがあって、そこでURLパスとMethodを設定するが、REST APIは、「リソース」タブで該当URLパスを選択後、「リソースを作成」から記入する「リソース名」がURLパスになる
  - 例えば、`/`を選択し、「リソースを作成」ボタンを押し、「リソース名」に`items`と記入して「リソースの作成」を押下すると、`https://{rest-api-id}.execute-api.{region}.amazonaws.com/{stage}/items`がエンドポイントになる
  - リソース作成後「APIをデプロイ」ボタンを押して、stageを選択してAPIをデプロイする必要がある

#### APIタイプ
##### REST API プライベート(Private)
> [!CAUTION]  
> REST APIをPrivate（VPC内からのみ利用可能）に設定した場合、VPCエンドポイントを作成してそれ経由でアクセスする必要がある
- https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/apigateway-private-api-test-invoke-url.html  
  > プライベート API は VPC 内から VPC エンドポイントを使用してのみ呼び出すことができます。プライベート API には、特定の VPC や VPC エンドポイントに API の呼び出しを許可するリソースポリシーが必要です。
- `com.amazonaws.{region}.execute-api` VPCエンドポイント作成後、**`https://{rest-api-id}-{vpce-id}.execute-api.{region}.amazonaws.com/{stage}`でアクセスできる**
  - https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/apigateway-private-api-test-invoke-url.html
- **「リソースポリシー」タブでアクセスコントロールを設定できる**
  - 特定のVPCエンドポイントからのアクセスのみを許可するポリシーの例  
    ```json
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": "*",
          "Action": "execute-api:Invoke",
          "Resource": "arn:aws:execute-api:ap-northeast-1:{アカウントID}:{REST APIのAPI ID}/*",
          "Condition": {
            "StringEquals": {
              "aws:sourceVpce": "vpce-{VPCエンドポイントID}"
            }
          }
        }
      ]
    }
    ```

#### Lambdaプロキシ統合
- https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/getting-started-rest-new-console.html
- Lambda関数をAPI Gatewayのバックエンドとして使用する場合、Lambdaプロキシ統合を使用することが推奨される
  - https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/api-gateway-api-integration-types.html  
    > Lambda プロキシ統合は、1 つの Lambda 関数との合理化された統合設定をサポートしています。設定はシンプルで、既存の設定を破棄することなくバックエンドで拡張できます。このような理由から、Lambda 関数との統合を強くお勧めします。
- Lambda プロキシ統合（Lambda Proxy Integration）は、API Gateway が受け取ったリクエスト全体（HTTP メソッド、パス、クエリパラメータ、ヘッダ、ボディなど）をそのまま JSON 形式で Lambda 関数に渡し、Lambda の戻り値を API Gateway が HTTP レスポンスに変換する方式
- **REST APIはLambdaプロキシ統合を明示的に有効にする必要があるが、Lambda との統合は 常に「プロキシ統合」扱いで、非プロキシ統合という選択肢はなく、API Gateway が受けたリクエストをそのまま Lambda に渡す。** らしい

---

## API Gatewayの構成
- https://docs.aws.amazon.com/ja_jp/lambda/latest/dg/services-apigateway.html#apigateway-proxy
- API Gatewayは**Stage**、**Resource**、**Method**で構成されている
- **Stage**と**Resource**でエンドポイントのパスが決まる
  - `https://<API Gatewayエンドポイント>/<Stage名>/<Resource名>`
#### Stage
- API Deploymentのための倫理的な概念
- APIの異なるバージョンや環境（開発、テスト、本番など）を管理するために使われる

#### Resources
- API内の個々のエンドポイント

## API Gatewayのクォーター
- https://docs.aws.amazon.com/apigateway/latest/developerguide/limits.html

## API GatewayのバックエンドとしてLambdaの紐づけ
1. API Gateway側から紐づけ
   - API Gateway(REST APIまたはREST API)作成後、「Create Resource」でエンドポイントのパスを作成。その後、作成したResourceに対して「Create method」でMethod typeを選択し、「Integration type」でLambda functionを選んで該当のFunctionを選択してMethodを作成する（事前にLambdaを作成しておく必要がある）
   - **作成したResourceに対して「Deploy API」でAPIをデプロイする必要がある。** その時にstageの選択が求められるので、ない場合は「_\*New stage\*_」でstageを一緒に作成する
   - APIをデプロイした後、**Stage**タブにて該当ResourceのMethod(e.g. POST)をクリックすると**Invoke URL**が表示される。これがエンドポイントなので、このエンドポイントに対してAPIを叩く
2. Lambda側から紐づけ
   - API Gateway側でstageが作成されている必要がある
   - Lambda関数名のパスで、GET/POST/PUT/DELETEなど、すべてのMethodのResourceが自動的に作成される

## API GatewayにWAF適用
> [!IMPORTANT]  
> API GatewayのWAF適用はREST APIでのみ可能。HTTP APIでは不可

- https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/apigateway-control-access-aws-waf.html
- まずWAFでACL/ruleを作成する必要がある

#### 送信元IPアドレスで制御
##### WAF側
- 「IP set」を作成
- 「Web ACLs」タブにて「Create web ACL」を押下
  - 「Associated AWS resources」で「Add AWS resources」を押下
  - 作成したAPI Gatewayを選択（REST APIのみ）
  - 「Rules」で「Add rules」→「Add my own rules and rule groups」を押下
  - 必要な項目を記入/選択し、Web ACLを作成
##### API Gateway側
- 「Stages」タブで該当Stageをクリックして「Stage details」の「Web ACL」にACLが関連付けされていることを確認する