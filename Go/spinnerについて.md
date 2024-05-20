- https://github.com/briandowns/spinner
- 処理中のspinnerのためのpackage
- 複数のspinnerの種類があり、`CharSets`で指定できる

## 使い方
- `New`関数で、第1引数に使うspinnerの種類を、第2引数にspinnerが回る速度を指定してインスタンスを初期化
- `Prefix`にspinnerの前に出力する文字列を入れる
- `Color`メソッドでspinnerの色を指定することもできる
- `Start()`と`Stop()`の間に処理を書く
```go
package main

import (
	"github.com/briandowns/spinner"
    "github.com/fatih/color"
	"time"
)

func main() {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)  // Build our new spinner
    s.Prefix = color.New(color.FgGreen).Sprint("処理中...")
	s.Color("fgHiGreen")
	s.Start()                                                    // Start the spinner
	time.Sleep(4 * time.Second)                                  // Run for some time to simulate work
	s.Stop()
}
```