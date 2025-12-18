# Gateway
- MCPサーバを公開できるサービス
- LambdaやREST APIタイプのAPI GatewayなどをMCPサーバとして公開することができる

## API Gatewayの場合
> [!CAUTION]  
> ### API Gateway時の制約事項
> - 参考URL: https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-target-api-gateway.html
> - **Your API must be in the same account as your AgentCore Gateway.** 
> - **Your API must be in the same Region as your AgentCore Gateway.**
> - **Your API must be an API Gateway REST API. We do not support API Gateway HTTP APIs or WebSocket APIs.**
> - **Your API must be configured with a public endpoint type. Private endpoints are not supported. To create a Gateway Target that can access resources in your VPC, you should use a public endpoint and an API Gateway private integration.**
> - **If your REST API has a method that uses `AWS_IAM` authorization and requires an API key, AgentCore Gateway won't support this method. It will be excluded from processing.**
> - **If your API uses a proxy resources, such as `/pets/{proxy+}`, AgentCore Gateway won't support this method.**
> - **To set up your API Gateway Target, AgentCore Gateway calls API Gateway's GetExport API on your behalf to get an OpenAPI 3.0 formatted export of your REST API Definition. For more details on this and how it might affect your Target configuration, see API Gateway Export.**