# Interceptorとは
- メソッドの前後に処理を行うための仕組み
- メソッドがリクエストを受け取る前、レスポンスを返した後など、任意のタイミングで任意の処理を割り込ませることができる
- 認証やロギング、バリデーションチェックなど、複数のRPCで共通して行いたい処理で使用
- gRPCではUnary用と、Stream用のInspectorが用意されている
- Server側、Client側のどちらにも対応
  - Server側
    - `UnaryServerInterceptor`
    - `StreamServerInterceptor`
  - Client側
    - `UnaryClientInterceptor`
    - `StreamClientInterceptor`