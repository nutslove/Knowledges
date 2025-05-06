- 参考URL
  - https://zenn.dev/hsaki/books/golang-concurrency/viewer/gointernal

## GoランタイムによるGoroutineスケジューリング
- OS(カーネル) ThreadとGo Thread(Goroutine)はN対Mの関係で、1つ(もしくは複数)のOS Threadに複数のGo Thread(Goroutine)をマッピングする
- OS ThreadとGoroutineのマッピングおよびOS ThreadにどのGoroutineを割り当てるかを管理するのがGoランタイム
- 普通はOSがOS Threadに割り当てるプロセスのスケジューリングを行い、Context Switchingが発生するけど、GoはGoランタイムがGoroutineのスケジューリングを管理するため、OS(Thread)レベルでのContext Switchingを最小限に抑え、OS Thread数より多い(e.g. 数百~数千)Goroutineを実行することができる
  - Go Thread(Goroutine)は軽量で、Goroutineの切り替えはOSレベルのContext Switchingより遥かに高速でオーバヘッドが少ない

## IPC（inter-process communication）、ITC（inter-thread communication）の方法
- IPC（プロセス間通信）やITC（スレッド/Goroutine間通信）には、主に**メモリ共有**型と**メッセージパッシング**型の2つがある
- Goにおいては、channelはメッセージパッシングの手段としてGoroutine間通信を行う
- `sync.Mutex`, `atomic`がメモリ共有型の通信手段
- `WaitGroup`は通信ではなく、Goroutineの終了を待つための同期ツール（協調手段）である