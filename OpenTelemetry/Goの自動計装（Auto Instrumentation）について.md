- 手動計装については、`Go/Opentelemetry/OpenTelemetry Go SDK.md`を参照

## 概要
- https://github.com/open-telemetry/opentelemetry-go-instrumentation
  - eBPFを使った自動計装
- https://opentelemetry.io/docs/zero-code/go/autosdk/
  - まだ限られたパッケージしかサポートしてないため、コード内でカスタムspanを作成したくなるかもしれない。
  - それで、eBPFベースの自動計装と手動計装を組み合わせて使うことも可能  
    > In this example, the eBPF framework automatically instruments incoming HTTP requests, then links the manual span to the same trace instrumented from the HTTP library. Note that there is no TracerProvider initialized in this sample. The Auto SDK registers its own TracerProvider that is crucial to enabling the SDK.
    > 
    > Essentially, there is nothing you need to do to enable the Auto SDK except create manual spans in an application instrumented by a Go zero-code agent. As long as you don’t manually register a global TracerProvider, the Auto SDK will automatically be enabled.
    ```go
    package main

    import (
    	"log"
    	"net/http"

    	"go.opentelemetry.io/otel"
    )

    func main() {
    	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    		// Get tracer
    		tracer := otel.Tracer("example-server")

    		// Start a manual span
    		_, span := tracer.Start(r.Context(), "manual-span")
    		defer span.End()

    		// Add an attribute for demonstration
    		span.SetAttributes()
    		span.AddEvent("Request handled")
    	})

    	log.Println("Server running at :8080")
    	log.Fatal(http.ListenAndServe(":8080", nil))
    }
    ```

> [!CAUTION]  
> Manually setting a global TracerProvider will conflict with the Auto SDK and prevent manual spans from properly correlating with eBPF-based spans. If you are creating manual spans in a Go application that is also instrumented by eBPF, do not initialize your own global TracerProvider.

> [!NOTE]  
> 2025/10現時点では、メトリクスとログの自動計装には対応していない。

### Getting Started
- https://github.com/open-telemetry/opentelemetry-go-instrumentation/blob/main/docs/getting-started.md

### Configuration
- https://github.com/open-telemetry/opentelemetry-go-instrumentation/blob/main/docs/configuration.md

### How it works
- https://github.com/open-telemetry/opentelemetry-go-instrumentation/blob/main/docs/how-it-works.md