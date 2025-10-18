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

---

## ■ Python（インタープリター言語）の場合
Pythonは**インタープリター**で実行される：

1. **動的実行**：コードは実行時に解釈される
2. **関数の動的置換**：Pythonでは実行時に関数を別の関数で置き換え可能
3. **モンキーパッチング**：ライブラリの関数を自動的にトレース機能付きの関数で置換
4. **sitecustomizeでPython起動時にモンキーパッチを適用**

---

## ■ Node.jsの場合
Node.jsは**JavaScriptエンジン（V8など）**で実行される：

1. **JITコンパイルと動的実行**
   - コードは最初にバイトコードにコンパイルされ、頻繁に実行されるコード（ホットパス）は最適化された機械語にコンパイルされる
   - 実行時の型変更やプロパティ追加など、JavaScriptの動的な性質は保持される
   - この動的性により、実行時に関数の置き換えが可能
2. **関数の動的置換（Shimmerパターン）**
   - JavaScriptでは実行時にオブジェクトのプロパティ（関数を含む）を別の関数で置き換え可能
   - OpenTelemetryは元の関数を保存し、トレース機能付きのラッパー関数で置き換える
   - ラッパー関数内で：①トレース開始（span作成） → ②元の関数を呼び出し → ③トレース終了
3. **自動インストルメンテーション（モンキーパッチング）**
   - OpenTelemetryは主要なライブラリ（http, mysql, express など）の関数を自動的にラップ
   - 例：mysqlのquery()関数をトレース機能付きの関数で置き換え
   - アプリケーションコードを変更せずにトレースを収集可能
4. **モジュールロードフック（InstrumentationNodeModuleDefinition）**
   - Node.jsのrequire()メカニズムにフックし、特定のモジュールがロードされる際に自動的にパッチを適用
   - 対象モジュール名とバージョン範囲を指定して、ロード時に関数を置き換え
   - パッチの適用と解除（アンラップ）の両方をサポート

例：
```javascript
// 元のコード
const mysql = require('mysql');
const connection = mysql.createConnection({ /* config */ });
connection.query('SELECT * FROM users', (err, results) => {
   console.log(results);
});

// 自動計装により内部的に以下のように変換される（簡略版）
const mysql = require('mysql');

// createConnection関数をラップ
mysql.createConnection = (originalCreateConnection => {
   return function(...args) {
       const connection = originalCreateConnection.apply(this, args);

       // query関数をラップ
       connection.query = (originalQuery => {
           return function(sql, values, callback) {
               // spanを開始
               const span = tracer.startSpan('mysql.query', {
                   kind: SpanKind.CLIENT,
                   attributes: {
                       'db.system': 'mysql',
                       'db.statement': sql
                   }
               });

               // コールバックをラップ（非同期処理対応）
               const wrappedCallback = function(err, results, fields) {
                   if (err) {
                       span.setStatus({
                           code: SpanStatusCode.ERROR,
                           message: err.message
                       });
                   }
                   span.end();  // 非同期処理完了後にspanを終了

                   // 元のコールバックを呼び出す
                   if (callback) callback(err, results, fields);
               };

               // 元のquery関数を実行
               return originalQuery.call(this, sql, values, wrappedCallback);
           };
       })(connection.query);

       return connection;
   };
})(mysql.createConnection);
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
