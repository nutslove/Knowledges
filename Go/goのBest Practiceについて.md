## `init`関数内で`panic`を引き起こす可能性のある関数を呼ぶのは避けるべき
### 理由
1. **回復不可能なエラー**
   - `init`関数で`panic`が発生すると、プログラム全体が起動時に終了する
   - `recover()`で捕捉することも困難

2. **デバッグの困難さ**
   - 初期化段階でのエラーは原因特定が難しい
   - スタックトレースが分かりにくい場合がある

3. **テストの問題**
   - パッケージをインポートするだけでテストが失敗する可能性
   - モックやスタブが困難

**良くない例：**
```go
func init() {
    db, err := sql.Open("mysql", "invalid-dsn")
    if err != nil {
        panic(err) // アプリケーション起動時に必ず失敗
    }
}
```

**推奨される方法：**
```go
var db *sql.DB

func InitDB() error {
    var err error
    db, err = sql.Open("mysql", dsn)
    return err
}

func main() {
    if err := InitDB(); err != nil {
        log.Fatal(err) // エラーハンドリングが明確
    }
}
```
- `init`関数は設定の読み込みや軽量な初期化処理に留め、失敗する可能性のある重い処理は明示的な初期化関数で行うのがベストプラクティス