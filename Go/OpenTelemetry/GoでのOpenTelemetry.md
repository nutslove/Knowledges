# OpenTelemetry Go SDKの使い方
- まず`otlptracehttp.New`でexporter(トレースの送り先)を設定し、接続を確立する  
  ```go
  import (
 	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"   
  )

  func main() {
	// OTLPエクスポーターの設定 (Start establishes a connection to the receiving endpoint.)
	// 第１引数がcontextで第２引数がoptions(2つ目のパラメータからスライスとなる)
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("<traceを受け付けるツール(e.g. jaeger, otel collector)のアドレス>:<traceを受け付けるポート(e.g. 4318)>"),
		otlptracehttp.WithInsecure(), // TLSを無効にする場合に指定
	)
  }
  ```
- `NewTracerProvider`で作成したexporterとSevice名などを渡してTraceProviderを設定する

## `NewTracerProvider`について
OpenTelemetry Go SDKの`trace.NewTracerProvider`は、トレースを生成および管理するためのトレーサープロバイダーを作成するための機能です。これはOpenTelemetryのトレース機能を使用するための基本的なコンポーネントの一つです。以下にその詳細を説明します。

### `trace.NewTracerProvider`の役割

1. **トレーサープロバイダーの作成**:
   - `trace.NewTracerProvider`は、新しいトレーサープロバイダーを作成します。このプロバイダーは、アプリケーション全体でトレースを収集するために使用される複数のトレーサーを生成します。

2. **トレーサーの取得**:
   - トレーサープロバイダーを使用して、特定の名前空間やバージョンに関連付けられたトレーサーを取得します。これにより、アプリケーションの異なる部分で異なるトレーサーを使用できます。

3. **トレースデータの収集とエクスポート**:
   - トレーサープロバイダーは、収集されたトレースデータをエクスポートするためのエクスポータを設定できます。これにより、トレースデータを外部のシステム（例: Jaeger、Zipkin、Prometheusなど）に送信することが可能です。

### 基本的な使用方法

以下は、`trace.NewTracerProvider`を使用してトレーサープロバイダーを作成し、トレーサーを取得する基本的な例です。

```go
package main

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/trace/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
    // Jaegerエクスポータを作成
    exp, err := jaeger.NewRawExporter(
        jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
    )
    if err != nil {
        panic(err)
    }

    // トレーサープロバイダーを作成
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
    )

    // グローバルなトレーサープロバイダーを設定
    otel.SetTracerProvider(tp)

    // トレーサーを取得
    tracer := otel.Tracer("example.com/trace")

    // トレースの開始
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()

    // ここにトレースしたいコードを追加
}
```

### トレーサープロバイダーの構成オプション

- **`trace.WithBatcher`**:
  - トレースをバッチでエクスポートするエクスポータを指定します。バッチエクスポータは効率的にトレースデータを収集し、一定の間隔でまとめて送信します。

- **`trace.WithSimpleSpanProcessor`**:
  - シンプルなスパンプロセッサを使用します。これは、各スパンを直ちにエクスポートするシンプルな方法です。

- **`trace.WithResource`**:
  - トレースデータに付加情報（例: サービス名、バージョンなど）を追加するためのリソースを指定します。

`trace.NewTracerProvider`は、OpenTelemetry Go SDKのトレース機能を最大限に活用するための基本的なスタートポイントであり、アプリケーションのトレースを効果的に管理・エクスポートするために重要な役割を果たします。

## `TracerProvider`について
`trace.NewTracerProvider`の戻り値である`TracerProvider`は、OpenTelemetryのGo SDKにおける重要なコンポーネントで、トレーシングの中心的な役割を果たします。具体的には、`TracerProvider`は以下のような機能を持っています。

### `TracerProvider`の機能と役割

1. **トレーサーの生成**:
   - `TracerProvider`は、アプリケーション全体で使用するトレーサーを生成します。トレーサーは、特定の操作やコンテキストに関連するスパン（トレースの一部）を作成するために使用されます。

2. **トレースの管理**:
   - `TracerProvider`は、生成された全てのトレーサーおよびそれらが生成するスパンを管理します。これには、トレースデータの収集、処理、およびエクスポートが含まれます。

3. **スパンプロセッサの設定**:
   - `TracerProvider`は、スパンプロセッサを使用して収集したスパンデータを処理します。スパンプロセッサには、バッチ処理やシンプルなスパン処理など、さまざまな方式があります。

4. **エクスポータの設定**:
   - トレースデータを外部のシステムにエクスポートするためのエクスポータを設定します。これにより、収集されたトレースデータをJaegerやZipkin、Prometheusなどのシステムに送信できます。

### `TracerProvider`の構成

`trace.NewTracerProvider`関数を使用して`TracerProvider`を作成する際に、以下のようにさまざまなオプションを指定することができます。

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/trace/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func main() {
    // Jaegerエクスポータを作成
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"))
    if err != nil {
        panic(err)
    }

    // リソースを定義
    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String("example-service"),
        semconv.ServiceVersionKey.String("v0.1.0"),
    )

    // トレーサープロバイダーを作成
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),   // バッチプロセッサを使用
        trace.WithResource(res),  // リソースを設定
    )

    // グローバルなトレーサープロバイダーを設定
    otel.SetTracerProvider(tp)

    // トレーサーを取得
    tracer := otel.Tracer("example.com/trace")

    // トレースの開始
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()

    // ここにトレースしたいコードを追加
}
```

### `TracerProvider`の主なメソッド

- **`Tracer(name string, opts ...trace.TracerOption) trace.Tracer`**:
  - 指定された名前とオプションでトレーサーを取得します。トレーサーは、スパンの作成やトレースの開始に使用されます。

- **`Shutdown(ctx context.Context) error`**:
  - トレーサープロバイダーをシャットダウンし、すべてのスパンをエクスポートします。アプリケーション終了時に呼び出して、未送信のトレースデータを確実にエクスポートするために使用します。

### まとめ

`TracerProvider`は、OpenTelemetryのGo SDKでトレースデータを収集、処理、エクスポートするための中心的なコンポーネントです。`trace.NewTracerProvider`関数を使用して作成され、アプリケーション全体でトレーシング機能を統合するために使用されます。適切なスパンプロセッサやエクスポータを設定することで、効率的にトレースデータを管理およびエクスポートできます。