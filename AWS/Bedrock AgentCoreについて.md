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

- OpenAPI 3.0仕様に準拠している必要がある
  - https://blog.serverworks.co.jp/api-gateway-mcp-policy
  - 例  
    ```json
    {
      "openapi": "3.0.1",
      "info": {
        "title": "dev-rca-apigw",
        "description": "API Gateway for RCA Tools",
        "version": "2025-12-17T19:16:58Z"
      },
      "servers": [
        {
          "url": "https://aaaaaa.execute-api.ap-northeast-1.amazonaws.com/{basePath}",
          "variables": {
            "basePath": {
              "default": "tools"
            }
          }
        }
      ],
      "paths": {
        "/newrelic/api/v1/nrql": {
          "post": {
            "operationId": "run_nrql_metric_log",
            "summary": "Run NRQL queries for metrics and logs",
            "requestBody": {
              "required": true,
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "required": ["account_id", "queries"],
                    "properties": {
                      "account_id": {
                        "type": "string",
                        "description": "NewRelic Account ID"
                      },
                      "queries": {
                        "type": "array",
                        "items": {
                          "type": "string"
                        },
                        "description": "List of NRQL queries to execute"
                      }
                    }
                  }
                }
              }
            },
            "responses": {
              "200": {
                "description": "Success",
                "content": {
                  "application/json": {
                    "schema": {
                      "type": "object",
                      "properties": {
                        "data": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "/newrelic/api/v1/nrql/trace": {
          "post": {
            "operationId": "run_nrql_trace",
            "summary": "Query distributed trace by trace ID",
            "requestBody": {
              "required": true,
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "required": ["trace_id"],
                    "properties": {
                      "trace_id": {
                        "type": "string",
                        "description": "Trace ID to search for"
                      }
                    }
                  }
                }
              }
            },
            "responses": {
              "200": {
                "description": "Success",
                "content": {
                  "application/json": {
                    "schema": {
                      "type": "object",
                      "properties": {
                        "trace": {
                          "type": "object"
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "/newrelic/api/v1/alerts/lists": {
          "post": {
            "operationId": "list_alerts",
            "summary": "List active alerts from NewRelic",
            "requestBody": {
              "required": true,
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "required": ["account_id"],
                    "properties": {
                      "account_id": {
                        "type": "string",
                        "description": "NewRelic Account ID"
                      }
                    }
                  }
                }
              }
            },
            "responses": {
              "200": {
                "description": "Success",
                "content": {
                  "application/json": {
                    "schema": {
                      "type": "object",
                      "properties": {
                        "data": {
                          "type": "string"
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
    ```
- API Gatewayで現在の形式を確認したい場合は、「**Stages**」タブの「**Stage actions**」→「**Export**」でOpenAPI 3.0形式でエクスポートできる
- ダウンロードしたOpenAPI定義を修正して、「**Resource**」タブの「**API actions**」→「**Import API**」でインポートすることで、API GatewayのAPI定義を更新できる

> [!CAUTION]  
> - 「import mode」を「Overwrite」にすると既存の設定がすべて上書きされるため注意すること
>   - 一部だけ更新したい場合は「Merge」を選択すること

- または、API Gatewayの「**Models**」タブで「**Create model**」にて、「**Content type**」を `application/json` とし、「**Model schema**」に以下のようなJSONスキーマを設定することでも定義できる  
  ```json
  {
    "required" : [ "account_id", "queries" ],
    "type" : "object",
    "properties" : {
      "account_id" : {
        "type" : "string",
        "description" : "NewRelic Account ID"
      },
      "queries" : {
        "type" : "array",
        "description" : "List of NRQL queries to execute",
        "items" : {
          "type" : "string"
        }
      }
    }
  }
  ```
  その後、「**Resource**」タブで、対象のメソッドを選択し、「**Method Response**」で `HTTP status code` `200` を選択、「**Response body**」で先ほど作成したモデルを紐付けることで、API GatewayのAPI定義を更新できる
