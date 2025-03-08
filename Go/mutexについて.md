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

## **Race Condition**とは  
- 競合状態のこと  
  > A race condition or race hazard is an undesirable condition of an electronics, software, or other system where the system's substantive behavior is dependent on the sequence or timing of other uncontrollable events. It becomes a bug when one or more of the possible behaviors is undesirable.
  > A race condition occurs when two Goroutine access a shared variable at the same time. 
- **Race Conditionは２つ以上のgo routineが同じもの(e.g. variable,struct,・・・)に対して更新処理を行う時に発生する。参照のみの時は発生しない**
- 参考URL
  - https://pkg.go.dev/sync#Mutex
  - https://learn.microsoft.com/en-us/troubleshoot/developer/visualstudio/visual-basic/language-compilers/race-conditions-deadlocks
  - https://stackoverflow.com/questions/34510/what-is-a-race-condition