# OTLP（OpenTelemetry Protocol）とは
- https://opentelemetry.io/docs/specs/otlp/
> The OpenTelemetry Protocol (OTLP) specification describes the encoding, transport, and delivery mechanism of telemetry data between telemetry sources, intermediate nodes such as collectors and telemetry backends.
>
> OTLP is a general-purpose telemetry data delivery protocol designed in the scope of the OpenTelemetry project.

> OTLP defines the encoding of telemetry data and the protocol used to exchange data between the client and the server.
> 
> This specification defines how OTLP is implemented over [gRPC](https://grpc.io/) and HTTP 1.1 transports and specifies [Protocol Buffers schema](https://protobuf.dev/overview/) that is used for the payloads.
>
> OTLP is a request/response style protocol: the clients send requests, and the server replies with corresponding responses. 

- otlpはOpenTelemetryでtelemetryデータ(metric/log/trace)のencoding、データ収集元(e.g. アプリ、サーバ)とCollectorやバックエンド(e.g. Tempo、Prometheus)とでデータやり取りする際のprotocol (HTTP/1.1、gRPC(with HTTP/2)) などを定義した仕様
- 参考URL
  - https://github.com/open-telemetry/opentelemetry-proto/tree/main/docs

---

# OTLPベースプロトコル
- OTLPはベースプロトコルとして、HTTP/1.1とgRPCを選択できる  
  > This specification defines how OTLP is implemented over [gRPC](https://grpc.io/) and HTTP 1.1 transports and specifies [Protocol Buffers schema](https://protobuf.dev/overview/) that is used for the payloads.
- HTTP/1.1の場合、テレメトリーデータは HTTP Request Bodyに含まれる。また、データは Protocol Buffers 形式でエンコードされ、Content-Type は`application/x-protobuf`となる。
- gRPCの場合、HTTP/2 上で動作し、テレメトリーデータは gRPC のメッセージペイロードとして送信される。また、Protocol Buffers を使用してデータをシリアライズ（複雑なデータ構造などを一つの文字列やバイト列に変換すること）する。