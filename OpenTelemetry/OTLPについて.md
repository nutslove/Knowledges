# OTLP（OpenTelemetry Protocol）とは
- **https://opentelemetry.io/docs/specs/otlp/**
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
- OTLPはtransport(転送)プロトコルとして、HTTP/1.1とgRPCを選択できる  
  > This specification defines how OTLP is implemented over [gRPC](https://grpc.io/) and HTTP 1.1 transports and specifies [Protocol Buffers schema](https://protobuf.dev/overview/) that is used for the payloads.

---

## OTLP/gRPC
- gRPCの場合、HTTP/2 上で動作し、テレメトリーデータは gRPC のメッセージペイロードとして送信される。また、Protocol Buffers を使用してデータをシリアライズ（複雑なデータ構造などを一つの文字列やバイト列に変換すること）する。
- データ送信の流れ  
  OTLPデータ（＝OpenTelemetry用のProtobufメッセージ）が、gRPCメッセージとしてエンコードされ、  
  それがHTTP/2のフレームに分割されてTCPで送られる。

### gRPC over HTTP/2とは
- gRPCもHTTP/2もアプリケーション層(L7)のプロトコル
- gRPCはデフォルトで、Protocol Buffers(protobuf)を使用する
- gRPCはHTTP/2を前提として設計されている
- HTTP/2の特徴
  - ヘッダー圧縮（HPACK）
  - 多重化（Multiplexing）
    - 1つのTCPコネクション上で複数のリクエスト/レスポンスを同時にやり取りできる
    - これにより接続のオーバーヘッドを削減し、高速な通信を可能にする
  - 双方向ストリーミング（Bi-directional Streaming）
    - 単なる一方向のリクエスト/レスポンスだけでなく、クライアントからサーバーへ、またはその両方で、持続的なデータフローを確立できる。これは、長時間のリアルタイム通信（チャットなど）の基盤となる。
  - バイナリフレーミング（Binary Framing）
    - プロトコル全体をバイナリ形式で扱う
    - ペイロード（データ本体）もProtocol Buffersというバイナリシリアライズ形式を使用することで、従来のHTTP/1.1やJSONのようなテキストベースのプロトコルに比べて、パース（解析）が高速になり、データサイズも小さくなる
- gRPCは、Protocol Buffersで定義された形式に従ってバイナリデータにシリアライズ（直列化）され、HTTP/2のフレームにカプセル化されて送信される

## OTLP/HTTP
- HTTPの場合、テレメトリーデータは HTTP Request Bodyに含まれる。データは以下の２つの形式から選択できる
  > OTLP/HTTP uses Protobuf payloads encoded either in [binary format](https://opentelemetry.io/docs/specs/otlp/#binary-protobuf-encoding) or in [JSON format](https://opentelemetry.io/docs/specs/otlp/#json-protobuf-encoding). Regardless of the encoding the Protobuf schema of the messages is the same for OTLP/HTTP and OTLP/gRPC as [defined here](https://github.com/open-telemetry/opentelemetry-proto/tree/v1.8.0/opentelemetry/proto).
  - binary format
    - Protocol Buffers 形式でエンコード(バイナリ化)され、HeaderのContent-Type は`application/x-protobuf`となる。
  - JSON format
    - HeaderのContent-Type は`application/json`となる。
- POSTメソッドを使う
- HTTP/1.1だけではなく、HTTP/2でも利用可能。ただ、HTTP/2で接続確立ができない場合、HTTP/1.1にフォールバックされなければならない。  
  > OTLP/HTTP uses HTTP POST requests to send telemetry data from clients to servers. Implementations MAY use HTTP/1.1 or HTTP/2 transports. Implementations that use HTTP/2 transport SHOULD fallback to HTTP/1.1 transport if HTTP/2 connection cannot be established.

