## HTTP API vs REST API
- https://docs.aws.amazon.com/ja_jp/apigateway/latest/developerguide/http-api-vs-rest.html

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