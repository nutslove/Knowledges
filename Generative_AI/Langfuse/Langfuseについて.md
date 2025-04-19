# Install（Self Hosting）
## Docker Compose
- https://langfuse.com/self-hosting/docker-compose

## Helm（k8s）
- https://github.com/langfuse/langfuse-k8s
- `langfuse.nextauth.secret.value`には`openssl rand -hex 32`で生成された値を入れる  
  ```shell
  helm install langfuse langfuse/langfuse -n langfuse -f values.yaml
  ```

# LangGraphとLangfuseの連携
- https://langfuse.com/docs/integrations/langchain/example-python-langgraph

# Trace IDの設定
- UUIDを使って利用者側でTraceIDを指定することもできる
  - これを使えば連続しない処理も同じTraceIDを指定することでトレースを連結させることができる
- https://langfuse.com/docs/tracing-features/trace-ids
