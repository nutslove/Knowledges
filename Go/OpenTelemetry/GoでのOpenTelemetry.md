# OpenTelemetry Go SDKの使い方
1. `otlptracehttp.New`でexporter(トレースの送り先)を設定し、接続を確立する  
2. `NewTracerProvider`で作成したexporterとSevice名などを渡してTraceProviderを設定する
3. `otel.SetTracerProvider`でアプリ全体でTracerProviderを使用するようにする
4. `otel.SetTextMapPropagator`でコンテキストのフォーマットを設定
5. `otel.Tracer`でtracerを取得
6. `tracer.Start`でspanを開始
   - `tracer.Start`の第2引数がspanのタイトルとなる  
     ![](../image/span_title.jpg)
7. `span.SetAttributes`でspanにattribute(付加情報)を追加

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

   // Tracerの設定
   tp := trace.NewTracerProvider(
      trace.WithBatcher(exporter),
      trace.WithResource(resource.NewWithAttributes(
         semconv.SchemaURL,                     // SchemaURL is the schema URL used to generate the trace ID. Must be set to an absolute URL.
         semconv.ServiceNameKey.String("HAM3"), // ServiceNameKey is the key used to identify the service name in a Resource.
      )),
   )
   otel.SetTracerProvider(tp)
   otel.SetTextMapPropagator(propagation.TraceContext{})

   tr := otel.Tracer("ham3")                      // spanのotel.library.name semantic conventionsに入る値
   ctx, span := tr.Start(ctx, "somethins started") // (新しい)spanの開始
   defer span.End()                               // spanの終了

   // Add attributes to the span
   span.SetAttributes(
      attribute.String("http.method", c.Request.Method),
      attribute.String("http.path", c.Request.URL.Path),
      attribute.String("http.host", c.Request.Host),
      attribute.Int("http.status_code", statusCode),
      attribute.String("http.user_agent", c.Request.UserAgent()),
      attribute.String("http.remote_addr", c.Request.RemoteAddr),
   )
}
```

## ■ `NewTracerProvider`について
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

## ■ `TracerProvider`について
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

## ■ `otel.SetTracerProvider`について
`otel.SetTracerProvider`は、OpenTelemetryのグローバルトレーサープロバイダーを設定するための関数です。これにより、アプリケーション全体で一貫して同じトレーサープロバイダーが使用されるようになります。以下に詳細を説明します。

### `otel.SetTracerProvider`とは何か？

`otel.SetTracerProvider`は、OpenTelemetryのグローバルトレーサープロバイダーを設定するための関数です。グローバルトレーサープロバイダーは、アプリケーションのどこからでもアクセスできるように設定されるため、各コンポーネントが個別にトレーサープロバイダーを設定する必要がなくなります。

### なぜ`otel.SetTracerProvider`が必要か？

1. **一貫性のあるトレーシング**:
   - アプリケーション全体で同じトレーサープロバイダーを使用することで、トレースが一貫して収集されます。異なるコンポーネント間でのトレースの結合や分析が容易になります。

2. **コードのシンプル化**:
   - 各コンポーネントが独自にトレーサープロバイダーを設定する必要がなくなるため、コードがシンプルになります。全体の設定を一箇所で管理できるようになります。

3. **中央管理**:
   - トレーサープロバイダーの設定や変更が一箇所で行われるため、管理が容易になります。例えば、エクスポータの変更やスパンプロセッサの設定変更が一元的に行えます。

### `otel.SetTracerProvider`の使用方法

`trace.NewTracerProvider`でトレーサープロバイダーを作成した後、それをグローバルトレーサープロバイダーとして設定するために`otel.SetTracerProvider`を使用します。以下はその例です。

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
    otel.SetTracerProvider(tp)  // ここでグローバルトレーサープロバイダーを設定

    // トレーサーを取得
    tracer := otel.Tracer("example.com/trace")

    // トレースの開始
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()

    // ここにトレースしたいコードを追加
}
```

### `otel.SetTracerProvider`は必須か？

`otel.SetTracerProvider`は必須ではありませんが、次のような場合に非常に有用です。

- **アプリケーション全体で一貫したトレーシングを行いたい場合**:
  - 各コンポーネントが統一されたトレーサープロバイダーを使用するため、トレースの管理が容易になります。

- **中央での設定管理を行いたい場合**:
  - 一箇所でトレーサープロバイダーの設定やエクスポータの変更を管理できます。

一方、特定のコンポーネントのみでトレーシングを行いたい場合や、複数の異なるトレーサープロバイダーを使用する場合には、各コンポーネントで個別にトレーサープロバイダーを設定することも可能です。その場合、`otel.SetTracerProvider`は不要です。

まとめると、`otel.SetTracerProvider`はアプリケーション全体で一貫したトレーシングを実現し、管理を簡素化するために推奨される方法です。しかし、特定の要件に応じて使用するかどうかを決定することができます。

## ■ `otel.SetTextMapPropagator`について
`otel.SetTextMapPropagator`は、OpenTelemetryのコンテキストプロパゲーションを設定するための関数です。コンテキストプロパゲーションは、分散システムにおいてトレースコンテキストを伝播させるためのメカニズムです。これにより、異なるサービス間でトレースコンテキストを共有し、一貫したトレーシングが可能になります。

### `otel.SetTextMapPropagator`とは？

`otel.SetTextMapPropagator`は、グローバルなテキストマッププロパゲータを設定するために使用されます。プロパゲータは、トレースコンテキスト（スパンコンテキストやバゲージなど）をHTTPヘッダやメッセージのプロパティとしてエンコードおよびデコードする役割を担います。
伝播されるコンテキストのフォーマットを決定し、アプリケーション全体でそのフォーマットを使用するように設定するためのもの。

### プロパゲータの種類
OpenTelemetryにはいくつかの標準的なプロパゲータがあります。以下は一般的なプロパゲータの例です。

1. **W3C Trace Context (`tracecontext`)**:
   - W3Cによって標準化されたトレースコンテキストのフォーマットです。
   
2. **Baggage (`baggage`)**:
   - 複数のサービス間でカスタムのキーと値のペアを伝播させるために使用されます。

3. **Composite Propagator**:
   - 複数のプロパゲータを組み合わせて使用することができます。

### `otel.SetTextMapPropagator`の使用方法

以下に、OpenTelemetryでプロパゲータを設定する基本的な例を示します。

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/trace"
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

    // W3C Trace ContextとBaggageのプロパゲータを設定
    propagator := propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},  // W3C Trace Context
        propagation.Baggage{},       // Baggage
    )

    // グローバルなプロパゲータを設定
    otel.SetTextMapPropagator(propagator)

    // トレーサーを取得
    tracer := otel.Tracer("example.com/trace")

    // トレースの開始
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()

    // ここにトレースしたいコードを追加
}
```

### プロパゲータを使用する利点

1. **一貫したコンテキストの伝播**:
   - サービス間でトレースコンテキストを確実に伝播させることで、分散トレーシングが容易になります。

2. **相互運用性の確保**:
   - 標準化されたプロパゲータ（例: W3C Trace Context）を使用することで、異なるトレーシングシステム間での相互運用性が確保されます。

3. **カスタムデータの伝播**:
   - Baggageプロパゲータを使用することで、トレースコンテキストにカスタムデータを含めて伝播させることができます。

### まとめ

`otel.SetTextMapPropagator`は、OpenTelemetryでコンテキストプロパゲーションを設定するために使用される重要な関数です。適切なプロパゲータを設定することで、サービス間でのトレースコンテキストの伝播が確実に行われ、一貫したトレーシングが可能になります。グローバルなトレーサープロバイダーと共に使用することで、アプリケーション全体で統一されたトレーシング環境を構築できます。
`otel.SetTextMapPropagator`を使用することで、アプリケーション全体で一貫したコンテキスト伝播フォーマットを設定できます。これにより、異なるサービス間でのトレースの統合が容易になり、分散システムにおけるトレーシングの一貫性と信頼性が向上します。

## ■ `otel.Tracer`について
`otel.Tracer`は、OpenTelemetryのトレースAPIを使用してトレーサーを取得するための関数です。トレーサー（Tracer）は、アプリケーション内でスパン（Span）を作成し、トレース（Trace）を生成するために使用されます。

### `otel.Tracer`とは？

`otel.Tracer`は、指定された名前とオプションのバージョンを持つトレーサーを返します。このトレーサーは、スパンを開始し、それらのスパンを終了するためのメソッドを提供します。トレーサーは、トレースプロバイダーによって管理されるインスタンスであり、トレースの開始点となります。

### トレーサーの具体的な役割

1. **スパンの作成**:
   - トレーサーは、アプリケーション内でスパンを作成するためのメソッドを提供します。スパンは、トレースの一部を構成する個々の操作やイベントを表します。

2. **コンテキスト管理**:
   - トレーサーは、スパンのコンテキストを管理し、スパン間の関係性を追跡します。これにより、分散システム内でのリクエストフローを理解するのが容易になります。

3. **メタデータの付加**:
   - トレーサーは、スパンに属性、イベント、およびステータスなどのメタデータを追加するためのメソッドを提供します。

### `otel.Tracer`の使用例

以下は、Goアプリケーションでトレーサーを取得し、スパンを作成してトレースする例です。

```go
import (
    "context"
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
    ctx := context.Background()
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()

    // スパン内で行う操作
    doWork(ctx)

    // トレーサーを使用した別のスパンの開始
    ctx, span2 := tracer.Start(ctx, "another operation")
    defer span2.End()

    // 別のスパン内で行う操作
    doMoreWork(ctx)
}

func doWork(ctx context.Context) {
    // ここでトレースする操作を実行
}

func doMoreWork(ctx context.Context) {
    // ここでトレースする別の操作を実行
}
```

### まとめ

`otel.Tracer`は、トレースAPIを使用してトレーサーを取得するための関数であり、トレーサーはスパンを作成し、トレースを生成するためのインスタンスです。トレーサーを使用することで、アプリケーション内の操作やイベントを詳細にトレースし、分散システムにおけるパフォーマンスの分析や問題の診断を行うことができます。