## Mutex
- Mutual Exclusion(排他制御)の略
- Race Conditionを防ぐために用いられる
- go run実行時`-race`オプションをつけることでRace Conditionを検出することができる
  - ex) `go run -race main.go`
- 使い方
  - Goroutine間で共有する変数を`sync.Mutex.Lock()`と`sync.Mutex.Unlock()`で囲むだけ
    ~~~go
    package main

    import (
        "fmt"
        "sync"
    )

    var msg string
    var wg sync.WaitGroup

    func updateMessage(s string, m *sync.Mutex) {
        defer wg.Done()

        m.Lock()
        msg = s
        m.Unlock()
    }

    func main() {
        msg = "Hello, world!"

        var mutex sync.Mutex

        wg.Add(2)
        go updateMessage("Hello, universe!", &mutex)
        go updateMessage("Hello, cosmos!", &mutex)
        wg.Wait()

        fmt.Println(msg)
    }
    ~~~

---

## **Race Condition**とは  
- 競合状態のこと  
  > A race condition or race hazard is an undesirable condition of an electronics, software, or other system where the system's substantive behavior is dependent on the sequence or timing of other uncontrollable events. It becomes a bug when one or more of the possible behaviors is undesirable.
  > A race condition occurs when two Goroutine access a shared variable at the same time. 
- **Race Conditionは２つ以上のgo routineが同じもの(e.g. variable,struct,・・・)に対して更新処理を行う時に発生する。参照のみの時は発生しない**
- 参考URL
  - https://pkg.go.dev/sync#Mutex
  - https://learn.microsoft.com/en-us/troubleshoot/developer/visualstudio/visual-basic/language-compilers/race-conditions-deadlocks
  - https://stackoverflow.com/questions/34510/what-is-a-race-condition

---

## グローバル(共通)の`mutex`、インスタンスごとの`mutex`
`mutex`（`sync.Mutex` や `sync.RWMutex`）の使い方は、**どの範囲で共有データを保護する必要があるか** による。

### 1. **インスタンスごとに `mutex` を用意するべきケース**
各インスタンスが**独立したデータ**を持っていて、それぞれのデータの整合性を保証する必要がある場合は、インスタンスごとに `mutex` を用意するのが適切。

**例: 各インスタンスが個別のカウンタを持つ場合**
```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```
この場合、各 `Counter` インスタンスが独立した `mutex` を持つため、異なる `Counter` インスタンス間での競合は発生しない。

---

### 2. **複数のインスタンスで `mutex` を共有すべきケース**
異なるインスタンスが**同じ共有リソース**（例えば、グローバル変数や共有データ構造）を扱う場合は、1つの `mutex` を複数のインスタンスで共有する必要がある。

**例: 複数のインスタンスが共通のマップを更新する場合**
```go
type SharedCounter struct {
    mu    *sync.Mutex
    value int
}

func (sc *SharedCounter) Increment() {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.value++
}

// 共有の mutex を作る
sharedMutex := &sync.Mutex{}

counter1 := &SharedCounter{mu: sharedMutex}
counter2 := &SharedCounter{mu: sharedMutex}

// どちらのインスタンスも sharedMutex を使用する
```
この場合、異なる `SharedCounter` インスタンスでも `mutex` を共有しているため、データの整合性が保たれる。

### ２つの違いを確認できる例
```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// グローバルミューテックス方式で使用されるミューテックス
var globalMutex sync.Mutex

// ログバックエンドのインターフェース
type LogBackend interface {
	Name() string
	ProcessLog(log string) 
}

// 2種類のモックバックエンド
type SlowBackend struct{}
type FastBackend struct{}

func (s *SlowBackend) Name() string { return "Slow Backend" }
func (f *FastBackend) Name() string { return "Fast Backend" }

// SlowBackendは処理に時間がかかる
func (s *SlowBackend) ProcessLog(log string) {
	fmt.Printf("開始: %s でログ '%s' を処理中...\n", s.Name(), log)
	time.Sleep(2 * time.Second) // 遅いバックエンドをシミュレート
	fmt.Printf("完了: %s でログ処理\n", s.Name())
}

// FastBackendは高速に処理する
func (f *FastBackend) ProcessLog(log string) {
	fmt.Printf("開始: %s でログ '%s' を処理中...\n", f.Name(), log)
	time.Sleep(200 * time.Millisecond) // 速いバックエンドをシミュレート
	fmt.Printf("完了: %s でログ処理\n", f.Name())
}

// 方式1: グローバルミューテックスを使用するLogBuffer
type GlobalMutexLogBuffer struct {
	backend LogBackend
}

func (g *GlobalMutexLogBuffer) WriteLog(log string) {
	// グローバルミューテックスを使用
	globalMutex.Lock()
	defer globalMutex.Unlock()
	
	fmt.Printf("[%s] ログを処理開始\n", g.backend.Name())
	g.backend.ProcessLog(log)
	fmt.Printf("[%s] ログを処理終了\n", g.backend.Name())
}

// 方式2: インスタンス固有のミューテックスを使用するLogBuffer
type InstanceMutexLogBuffer struct {
	mutex   sync.Mutex
	backend LogBackend
}

func (i *InstanceMutexLogBuffer) WriteLog(log string) {
	// インスタンス固有のミューテックスを使用
	i.mutex.Lock()
	defer i.mutex.Unlock()
	
	fmt.Printf("[%s] ログを処理開始\n", i.backend.Name())
	i.backend.ProcessLog(log)
	fmt.Printf("[%s] ログを処理終了\n", i.backend.Name())
}

func main() {
	// 2つの異なるバックエンド
	slowBackend := &SlowBackend{}
	fastBackend := &FastBackend{}
	
	fmt.Println("==== グローバルミューテックス方式のテスト ====")
	// グローバルミューテックスを使用する2つのLogBuffer
	slowBufferGlobal := &GlobalMutexLogBuffer{backend: slowBackend}
	fastBufferGlobal := &GlobalMutexLogBuffer{backend: fastBackend}
	
	// 並行実行
	var wg sync.WaitGroup
	wg.Add(2)
	
	start := time.Now()
	
	go func() {
		defer wg.Done()
		slowBufferGlobal.WriteLog("重要なイベント")
	}()
	
	go func() {
		defer wg.Done()
		fastBufferGlobal.WriteLog("軽微な通知")
	}()
	
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("グローバルミューテックス方式の所要時間: %v\n\n", elapsed)
	
	// 少し待機して出力を分ける
	time.Sleep(1 * time.Second)
	
	fmt.Println("==== インスタンス固有ミューテックス方式のテスト ====")
	// インスタンス固有のミューテックスを使用する2つのLogBuffer
	slowBufferInstance := &InstanceMutexLogBuffer{backend: slowBackend}
	fastBufferInstance := &InstanceMutexLogBuffer{backend: fastBackend}
	
	// 並行実行
	wg.Add(2)
	
	start = time.Now()
	
	go func() {
		defer wg.Done()
		slowBufferInstance.WriteLog("重要なイベント")
	}()
	
	go func() {
		defer wg.Done()
		fastBufferInstance.WriteLog("軽微な通知")
	}()
	
	wg.Wait()
	elapsed = time.Since(start)
	fmt.Printf("インスタンス固有ミューテックス方式の所要時間: %v\n", elapsed)
}
```

```shell
==== グローバルミューテックス方式のテスト ====
[Fast Backend] ログを処理開始
開始: Fast Backend でログ '軽微な通知' を処理中...
完了: Fast Backend でログ処理
[Fast Backend] ログを処理終了
[Slow Backend] ログを処理開始
開始: Slow Backend でログ '重要なイベント' を処理中...
完了: Slow Backend でログ処理
[Slow Backend] ログを処理終了
グローバルミューテックス方式の所要時間: 2.20147626s

==== インスタンス固有ミューテックス方式のテスト ====
[Fast Backend] ログを処理開始
開始: Fast Backend でログ '軽微な通知' を処理中...
[Slow Backend] ログを処理開始
開始: Slow Backend でログ '重要なイベント' を処理中...
完了: Fast Backend でログ処理
[Fast Backend] ログを処理終了
完了: Slow Backend でログ処理
[Slow Backend] ログを処理終了
インスタンス固有ミューテックス方式の所要時間: 2.000275883s
```
#### ２つの処理の違い
1. **グローバルミューテックス方式:**
  - 最初にSlowBackendの処理が完了するまで、FastBackendの処理は開始されない
  - 両方のバッファが同じミューテックスを共有しているため、一方がロックしている間は他方は待機状態になる
  - 合計所要時間は約2.2秒（2秒 + 0.2秒）になる

2. **インスタンス固有ミューテックス方式:**
- SlowBackendとFastBackendの処理が並行して行われる
- それぞれのバッファが独自のミューテックスを持っているため、互いに独立して動作できる
- 合計所要時間は約2秒（遅い方に合わせる）になる

---

### まとめ
- **インスタンスごとのデータを保護する場合** → **インスタンスごとに `mutex` を持つ**
- **複数のインスタンスが同じリソースを共有する場合** → **1つの `mutex` を共有する**
