- 参考URL
  - https://zenn.dev/hsaki/books/golang-concurrency/viewer/gointernal

## GoランタイムによるGoroutineスケジューリング
- OS(カーネル) ThreadとGo Thread(Goroutine)はN対Mの関係で、1つ(もしくは少数)のOS Threadに複数のGo Thread(Goroutine)をマッピングする
- OS ThreadとGoroutineのマッピングおよびOS ThreadにどのGoroutineを割り当てるかを管理するのがGoランタイム
- 普通はOSがOS Threadに割り当てるプロセスのスケジューリングを行い、Context Switchingが発生するけど、GoはGoランタイムがGoroutineのスケジューリングを管理するため、OS(Thread)レベルでのContext Switchingを最小限に抑え、OS Thread数より多い(e.g. 数百~数千)Goroutineを実行することができる
  - Go Thread(Goroutine)は軽量で、Goroutineの切り替えはOSレベルのContext Switchingより遥かに高速でオーバヘッドが少ない