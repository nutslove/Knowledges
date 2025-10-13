# 各言語の自動計装の仕組み

## ■ Java（JVM言語）の場合
Javaは**バイトコード**にコンパイルされ、**JVM（Java仮想マシン）** 上で実行される：

1. **バイトコード変換**：ソースコード → バイトコード → JVMで実行

> [!NOTE]  
> ### バイトコードとは
> - Javaプログラムがコンパイルされる中間形式
> - JVMによって実行される
> - バイトコードはプラットフォーム（OSやハードウェア）に依存しない
> - バイトコード ≠ 機械語
> - JVMが `.class` のバイトコードを解釈実行するか、JITコンパイラ（Just-In-Time Compiler） が機械語に変換してCPUで実行する
> ```scss
> ソースコード (.java)
>    ↓ javac（コンパイル）
> バイトコード (.class)
>    ↓ JVMが解釈またはJITコンパイル
> 機械語（CPUが実行）
> ```

2. **動的バイトコード操作**：JVMには実行時にバイトコードを変更する機能がある
3. **Java Agent**：JVMの `-javaagent` オプションを使って、Class loading時にバイトコードを自動的に書き換え
4. **ASM/ByteBuddy**：これらのライブラリを使ってメソッドの開始・終了にトレース用のコードを自動挿入

例：`public void myMethod()` が実行時に以下のように変換される
```java
// 元のUserService.class（バイトコード）
class UserService {
    public void createUser(String name) {
        System.out.println("Creating user: " + name);
    }
}

// ↓ クラスローディング時にJava Agentが自動変換 ↓
class UserService {
    public void createUser(String name) {
        // 自動挿入：トレース開始
        NewRelicAgent.getTransaction().startSegment("UserService.createUser");
        try {
            // 元のコード
            System.out.println("Creating user: " + name);
        } finally {
            // 自動挿入：トレース終了
            NewRelicAgent.getTransaction().endSegment();
        }
    }
}
```

---

## ■ Python（インタープリター言語）の場合
Pythonは**インタープリター**で実行される：

1. **動的実行**：コードは実行時に解釈される
2. **関数の動的置換**：Pythonでは実行時に関数を別の関数で置き換え可能
3. **モンキーパッチング**：ライブラリの関数を自動的にトレース機能付きの関数で置換
4. **import hook**：モジュールのimport時に自動的にパッチを適用

例：
```python
# 元のコード
import requests
requests.get("http://example.com")

# 自動計装により内部的に以下のように変換
def instrumented_get(*args, **kwargs):
    span = tracer.start_span("http_request")
    try:
        return original_get(*args, **kwargs)
    finally:
        span.finish()

requests.get = instrumented_get  # 関数を置換
```

---

## ■ Node.jsの場合
Node.jsは**JavaScriptエンジン（V8など）**で実行される：

1. **動的実行**：コードは実行時に解釈される
2. **関数の動的置換**：Node.jsでは実行時に関数を別の関数で置き換え可能
3. **モンキーパッチング**：ライブラリの関数を自動的にトレース機能付きの関数で置換
   - OpenTelemetryは主要なライブラリ（`http`, `mysql` など）の関数をラップしてトレース処理を挿入する。
4. **requireフック**：モジュールの読み込み時に自動的にパッチを適用

例：
```javascript
// 元のコード
const http = require('http');
http.get("http://example.com", (res) => {
    console.log("Response received");
});

# 自動計装により内部的に以下のように変換
const http = require('http');
http.get = (originalGet => {
    return function(...args) {
        const span = tracer.start_span("http_request");
        try {
            return originalGet.apply(this, args);
        } finally {
            span.finish();
        }
    };
})(http.get);
```

---

# Goで自動計装ができない理由
## **コンパイル言語の制約**
1. **静的コンパイル**：Goはソースコードからマシンコードに直接コンパイルされる
2. **実行時変更不可**：一度コンパイルされたバイナリは実行時に変更できない
3. **仮想マシンなし**：JVMのような中間層がないため、実行時のコード操作ができない
4. **静的リンク**：すべての依存関係がバイナリに静的にリンクされる

## **技術的な違い**

| 言語 | 実行方式 | 自動計装方法 | 変更タイミング |
|------|----------|--------------|----------------|
| Java | JVM上でバイトコード実行 | バイトコード変換 | Class load時 |
| Python | インタープリター実行 | 関数置換/パッチング | import/実行時 |
| Go | ネイティブバイナリ実行 | **不可能** | コンパイル時のみ |

## **Goが試行している解決策**
- **eBPF**
