## **`time.Ticker` とは？**
`time.Ticker` は、指定した間隔ごとに **チャンネル (`C`) に時刻を送信** するタイマー。  
Go で **定期的に処理を実行する** 場合に使われる。  

---

## **`time.Ticker` の基本的な使い方**
```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(2 * time.Second) // 2秒ごとに実行
	defer ticker.Stop() // プログラム終了時にリソースを解放

	for i := 0; i < 5; i++ {
		<-ticker.C // チャンネルから時刻を受信（ブロッキング）
		fmt.Println("Tick at", time.Now())
	}
}
```

### **ポイント**
- `time.NewTicker(間隔)` で `Ticker` を作成。
- `ticker.C` という **チャネル** から **定期的に時刻を受信** できる。
- `ticker.Stop()` を呼ぶと **Ticker の停止** & **リソース解放**。

---

## **`Ticker` を使った定期実行**
`time.Ticker` を **goroutine** で使うと、一定間隔でタスクを実行できる。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second) // 1秒ごと
	defer ticker.Stop()

	go func() {
		for tick := range ticker.C {
			fmt.Println("Tick at", tick)
		}
	}()

	// 10秒後にプログラム終了
	time.Sleep(10 * time.Second)
	fmt.Println("Ticker stopped")
}
```

### **ポイント**
- `range ticker.C` で、**チャネルが閉じるまで定期的に受信** できる。
- `time.Sleep(10 * time.Second)` で 10 秒間動作。
- `ticker.Stop()` をしないと、**ゴルーチンが無駄に動き続ける** ので注意。

---

## **`Ticker` の停止**
`Ticker` は `Stop()` で停止できる。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(500 * time.Millisecond) // 500msごと
	defer ticker.Stop()

	go func() {
		for tick := range ticker.C {
			fmt.Println("Tick at", tick)
		}
	}()

	time.Sleep(2 * time.Second) // 2秒動作
	ticker.Stop()               // 停止
	fmt.Println("Ticker stopped")

	time.Sleep(1 * time.Second) // 追加の1秒で確認
}
```

### **動作の流れ**
1. `500ms` ごとにログを出力。
2. **2秒後に `Stop()` で停止** し、`ticker.C` の受信が止まる。
3. **3秒目にはもう `Tick` しない**（ゴルーチンが止まる）。

---

## **`Ticker` を `select` で使う**
`select` を使うと **複数のチャンネルを監視** できる。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second)
	stop := make(chan bool) // 停止用のチャンネル

	go func() {
		for {
			select {
			case tick := <-ticker.C:
				fmt.Println("Tick at", tick)
			case <-stop:
				fmt.Println("Received stop signal")
				ticker.Stop()
				return
			}
		}
	}()

	// 5秒後に停止信号を送る
	time.Sleep(5 * time.Second)
	stop <- true

	time.Sleep(1 * time.Second) // 追加の1秒待って確認
}
```

### **ポイント**
- `select` を使うと **`ticker.C` だけでなく、停止用のチャンネル (`stop`) も監視** できる。
- `stop <- true` で停止すると、**ゴルーチンが終了** する。

---

## **`time.Ticker` vs `time.After`**
| 特徴 | `time.Ticker` | `time.After` |
|------|--------------|--------------|
| **繰り返し** | ○ (定期実行) | ✗ (1回だけ) |
| **チャンネル** | `C` を `for` で監視 | `<-time.After()` を `select` で待つ |
| **リソース管理** | `Stop()` しないとゴルーチンが残る | `time.After()` は自動解放 |

### **`time.After` の例**
```go
time.Sleep(5 * time.Second) // 5秒待つのと同じ
<-time.After(5 * time.Second)
```
`time.After()` は **一度だけ** 時間が来たら値を送る。

---

## **実用例**
### **1. 一定間隔でAPIリクエスト**
```go
package main

import (
	"fmt"
	"time"
)

func requestAPI() {
	fmt.Println("Requesting API at", time.Now())
}

func main() {
	ticker := time.NewTicker(10 * time.Second) // 10秒ごと
	defer ticker.Stop()

	for tick := range ticker.C {
		fmt.Println("Tick at", tick)
		requestAPI()
	}
}
```
