## `bytes.Buffer`
- https://pkg.go.dev/bytes#Buffer  
  > A Buffer is a variable-sized buffer of bytes with [Buffer.Read](https://pkg.go.dev/bytes#Buffer.Read) and [Buffer.Write](https://pkg.go.dev/bytes#Buffer.Write) methods. The zero value for Buffer is an empty buffer ready to use.
- メモリ上にデータを蓄積し、文字列やバイナリデータを効率的に扱うことができる
- 内部的に可変長の`[]byte`をメモリ上で管理し、データを追加するたびに自動的に拡張される
- `bytes.Buffer` は `io.Writer` インターフェースを実装しているため、 `fmt.Fprint` や `csv.NewWriter` などの出力先として利用可能

### コード例
```go
func generateCSVString(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("データが空です")
	}

	var buffer bytes.Buffer // メモリ上のバッファを作成
	writer := csv.NewWriter(&buffer) // バッファにCSVを書き込むためのWriterを作成 (`csv.NewWriter`の出力先として指定)

	headers := getHeaders(data)
	writer.Write(headers) // buffer にデータが蓄積される（データ追加）

	for _, entry := range data {
		row := make([]string, len(headers))
		for i, header := range headers {
			if val, exists := entry[header]; exists {
				row[i] = formatValue(val)
			} else {
				row[i] = ""
			}
		}
		writer.Write(row) // bufferにデータ追加
	}

	writer.Flush() // csv.Writer は内部的にバッファリングを行うため、Flush() を呼ぶことで buffer に確実に書き込む

	return buffer.String(), nil
  // buffer.String() を呼び出すと bytes.Buffer 内のデータを文字列として取得できる
}
```