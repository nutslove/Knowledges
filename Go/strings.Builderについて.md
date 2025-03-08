## **`strings.Builder` とは？**
`strings.Builder` は **文字列を効率的に扱うための構造体（文字列の連結に特化したバッファ）** 。  
パフォーマンスを向上させるために設計されており、`+` 演算子や `fmt.Sprintf` よりも効率的に文字列を連結できます。

---

## **`strings.Builder` のメリット**
1. **メモリ効率が良い**  
   - `strings.Builder` は **連結時に新しい文字列を作成せず、内部バッファを拡張** しながらデータを蓄積する。
   - `+` 演算子のように毎回新しい文字列を生成しないため、**GC の負担が減る**。

2. **高速**  
   - `bytes.Buffer` のようなスライス管理を行いながらも、**バイトスライスを文字列に変換するオーバーヘッドがない**。
   - `bytes.Buffer` の場合 `b.Bytes()` を `string(b.Bytes())` に変換するとコピーが発生するが、`Builder` は `String()` の呼び出しが軽量。

---

## **基本的な使い方**
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder

	// 文字列の追加
	sb.WriteString("Hello, ")
	sb.WriteString("Golang!")
	
	// 結果を取得
	result := sb.String()
	fmt.Println(result) // Hello, Golang!
}
```
---

## **`strings.Builder` の主なメソッド**
| メソッド                  | 説明 |
|--------------------------|-------------------------------------------|
| `Write(p []byte) (int, error)` | `[]byte` をバッファに追加 |
| `WriteString(s string) (int, error)` | 文字列をバッファに追加 |
| `WriteByte(b byte) error` | 1バイト追加 |
| `WriteRune(r rune) (int, error)` | Unicode 文字 (rune) を追加 |
| `String() string` | 連結された文字列を取得 |
| `Len() int` | 現在の長さを取得 |
| `Reset()` | バッファをクリア (メモリは解放されない) |

---

## **実践的なサンプル**
### **1. `WriteRune` を使った Unicode 文字の連結**
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder
	sb.WriteRune('世')
	sb.WriteRune('界')

	fmt.Println(sb.String()) // 世界
}
```

### **2. `Write` を使ったバイトスライスの連結**
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder
	sb.Write([]byte("こんにちは"))

	fmt.Println(sb.String()) // こんにちは
}
```

### **3. `Reset()` を使ったリセット**
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder
	sb.WriteString("リセット前")
	fmt.Println(sb.String()) // リセット前

	sb.Reset()
	sb.WriteString("リセット後")
	fmt.Println(sb.String()) // リセット後
}
```

---

## **`strings.Builder` vs `bytes.Buffer`**
| 特徴                | `strings.Builder` | `bytes.Buffer` |
|--------------------|------------------|---------------|
| 文字列専用か       | **Yes** (string に最適化) | No (`[]byte` を扱う) |
| パフォーマンス      | **高速 (文字列処理に特化)** | バイト配列処理向け |
| `String()` の負荷  | **軽い (コピーなし)** | **重い (コピー発生)** |
| スレッドセーフ性    | **なし (速い)** | **なし (速い)** |
| 用途               | 文字列の結合・生成 | バイナリデータ処理 |

### **`bytes.Buffer` を使った場合の `String()` のコスト**
```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	var buf bytes.Buffer
	buf.WriteString("Hello, ")
	buf.WriteString("World!")

	// string(buf.Bytes()) でコピー発生
	result := string(buf.Bytes())
	fmt.Println(result)
}
```
- `bytes.Buffer` の場合、`Bytes()` で `[]byte` を取得し、`string(buf.Bytes())` で **コピーが発生** するため、`strings.Builder` よりも **メモリ効率が悪い**。
