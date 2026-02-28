# Ginのモード

Ginには3つのモードがあり、環境変数 `GIN_MODE` または `gin.SetMode()` で切り替えられる。

## 各モードの概要

| モード | 定数 | 内部コード | `IsDebugging()` |
|---|---|---|---|
| `gin.DebugMode` | `"debug"` | 0 | `true` |
| `gin.ReleaseMode` | `"release"` | 1 | `false` |
| `gin.TestMode` | `"test"` | 2 | `false` |

デフォルトは `DebugMode`。

## モードが影響するもの

唯一の実質的な違いは `debugPrint()` の出力有無。

```go
func debugPrint(format string, values ...any) {
    if !IsDebugging() {
        return
    }

    if DebugPrintFunc != nil {
        DebugPrintFunc(format, values...)
        return
    }

    if !strings.HasSuffix(format, "\n") {
        format += "\n"
    }
    fmt.Fprintf(DefaultWriter, "[GIN-debug] "+format, values...)
}
```

`IsDebugging()` は `ginMode == debugCode`（= 0）のときだけ `true` を返すため、DebugMode以外では早期リターンされる。結果として**DebugModeでのみ**以下のようなログが出力される。

```
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
[GIN-debug] GET /users/:id --> main.GetUser (3 handlers)
[GIN-debug] POST /users   --> main.CreateUser (3 handlers)
```

ReleaseModeとTestModeではこれらが全て抑制される。

なお、`DebugPrintFunc` を設定することで、デバッグログの出力先をカスタマイズすることも可能。

## ReleaseMode と TestMode の違い

gin内部の動作上の差はない。`gin.Mode()` が返す文字列が `"release"` か `"test"` かだけの違い。アプリケーションコード側でモードに応じた分岐をしたい場合に使い分けられる。

```go
if gin.Mode() == gin.TestMode {
    // テスト固有の処理
}
```

## 設定方法

```go
// コードで設定
gin.SetMode(gin.ReleaseMode)

// 環境変数で設定
// GIN_MODE=release go run main.go
```