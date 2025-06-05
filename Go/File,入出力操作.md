## `io.Writer`と`io.Reader`
- Goの標準ライブラリioパッケージで定義されている重要なインターフェースであり、データの読み書きを抽象化し、様々な入出力操作を統一的に扱うことができる
### `io.Writer`
- `Write(p []byte) (n int, err error)`メソッドを持つ。
  - byteスライス`p` (書き込むデータを一時的に入れておくバッファ) をデータの宛先に書き込むことを表す。
  - 書き込まれたバイト数`n`と、エラーが発生した場合は`err`を返す。
- ファイル、ネットワーク接続、バッファなど、様々な出力先を抽象化する。
- `os.Stdout`、`http.ResponseWriter`、`bufio.Writer`など、多くの型が`io.Writer`を実装している。

### `io.Reader`
- `Read(p []byte) (n int, err error)`メソッドを持つ。
  - データをbyteスライス`p` (読み込んだデータを一時的に入れておくバッファ) に読み込むことを表す。
  - 読み込まれたバイト数`n`と、エラーが発生した場合や読み込みが終了した場合は`err`を返す。
- ファイル、ネットワーク接続、バッファなど、様々な入力元を抽象化する。
- `os.Stdin`、`http.Request.Body`、`bufio.Reader`など、多くの型が`io.Reader`を実装している。

## `io.Copy`関数
- ある`io.Reader`からデータを読み取り、`io.Writer`に書き込むことができる関数
- `io.Copy`の基本的な使用法  
  ```go
  func Copy(dst io.Writer, src io.Reader) (written int64, err error)
  ```
  - `dst`：データの書き込み先となる`io.Writer`
  - `src`：データの読み取り元となる`io.Reader`
  - srcからデータを読み取り、dstに書き込む。コピーされたバイト数と、発生したエラー（ある場合）が返される。
- 例  
  ```go
  package main

  import (
      "io"
      "os"
  )

  func main() {
      // コピー元のファイルを開く
      src, err := os.Open("source.txt")
      if err != nil {
          panic(err)
      }
      defer src.Close()

      // コピー先のファイルを作成
      dst, err := os.Create("destination.txt")
      if err != nil {
          panic(err)
      }
      defer dst.Close()

      // コピーを実行
      _, err = io.Copy(dst, src)
      if err != nil {
          panic(err)
      }
  }
  ```
### `io.Copy`の注意点
- **`io.Copy`関数はファイルに書き込んだ後、ファイルポインタがファイルの終端になる。なので、その状態で`bufio.NewScanner`でファイルを読み込むと何も読み込まれない。そのため、ファイルを読み込む前に`Seek`メソッドでファイルポインタを最初に戻す必要がある。**  
  ```go
  // 解凍したデータを出力ファイルにコピー
  _, err = io.Copy(outputFile, gzipReader)
  if err != nil {
      log.Printf("Could not copy data to output file: %v\n", err)
      return
  }
  fmt.Println("Decompressed file:", outputFile.Name())

  // ファイルポインタを先頭に戻す
  _, err = outputFile.Seek(0, 0)
  if err != nil {
      log.Printf("Could not reset file pointer: %v\n", err)
      return
  }

  scanner := bufio.NewScanner(outputFile) // ファイルを行ごとに読み込む
  ```

#### `Seek`メソッド
- ファイルポインタを移動させる
- `Seek`メソッドは次の2つの引数を取る
  1. **第1引数（offset）**: ファイル内の位置をバイト単位で指定。これは整数値（通常は`int64`型）。
  2. **第2引数（whence）**: 基準点を指定する定数で、以下の3つの値のいずれかを使用
     - `0` または `io.SeekStart`: ファイルの先頭からのオフセット
     - `1` または `io.SeekCurrent`: 現在の位置からのオフセット
     - `2` または `io.SeekEnd`: ファイルの末尾からのオフセット
- 例  
  ```go
  // ファイルの先頭に移動
  file.Seek(0, 0)  // または file.Seek(0, io.SeekStart)

  // 現在位置から10バイト進む
  file.Seek(10, 1)  // または file.Seek(10, io.SeekCurrent)

  // ファイルの末尾から5バイト戻る
  file.Seek(-5, 2)  // または file.Seek(-5, io.SeekEnd)
  ```

## `os.Open` と `os.Create`、`os.Remove`、`os.RemoveAll`
- ファイルの新規作成は`os.Create`、既存のファイルを開くときは`os.Open`  
- `os.Remove`はファイル削除  
  ```go
  import "os"

  // ファイルを削除
  err := os.Remove("filename.txt")
  if err != nil {
      // エラー処理
      fmt.Println("ファイル削除エラー:", err)
  }
  ```
- ディレクトリの再帰的削除は`os.RemoveAll()`  
  ```go
  // ディレクトリとその中身を全て削除
  err := os.RemoveAll("directory")
  if err != nil {
      fmt.Println("削除エラー:", err)
  }
  ```

## `bufio.Scanner`
- **方法**: この方法では `bufio.Scanner` を使ってファイルの内容を行単位で読み込みます。一度に一行ずつ読み込むため、メモリ使用量が少なく済みます。
- **利点**: 大きなファイルを扱う場合、一度に少量のデータだけをメモリに読み込むため、メモリ効率が良いです。
- **欠点**: 行単位での読み込みなので、一つの行が非常に長い場合や、検索する文字列が行の境界にまたがる場合、正確な検索ができないことがあります。
```go
package main

import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    // ファイルを開く
    file, err := os.Open("sample.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer file.Close()

    // Scannerを作成
    scanner := bufio.NewScanner(file)

    // 1行ずつ読み込む
    for scanner.Scan() {
        line := scanner.Text() // 読み込んだ行はscanner.Text()で取得
        fmt.Println(line)
    }

    // エラーチェック(スキャン中にエラーが発生したかどうかをチェック。エラーがある場合は、エラーメッセージを表示)
    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading from file:", err)
    }
}
```

## `io.ReadAll`
- 与えられた`io.Reader`インターフェイスからデータを読み取り、読み取ったデータをbyteスライスとして返す。
- **方法**: この方法では `io.ReadAll` を使ってファイルの内容を一度に全て読み込みます。これにより、ファイル全体を一つの大きな文字列として扱えます。
- **利点**: 単純で理解しやすい。ファイル全体を一度に読み込むため、文字列が行の境界にまたがっていても正確に検索できます。
- **欠点**: ファイルサイズが大きい場合、その全てをメモリに読み込む必要があるため、メモリ使用量が多くなります。非常に大きなファイルではメモリ不足を引き起こす可能性があります。
```go
package main

import (
    "io"
    "os"
    "log"
)

func main() {
    // 標準入力からデータを読み取る例
    data, err := io.ReadAll(os.Stdin)
    if err != nil {
        log.Fatalf("Error reading data: %v", err)
    }

    // 読み取ったデータを出力
    println(string(data))
}
```

### どちらがより良いか
- **ファイルサイズが小さいまたは中程度の場合**: `io.ReadAll` を使う方法が簡単で効果的です。ファイル全体を一度に読み込むことで、検索が簡単になります。
- **ファイルサイズが大きい場合**: `bufio.Scanner` を使う方法が適しています。メモリ効率が良く、大きなファイルでも扱いやすいです。ただし、文字列が行の境界にまたがる場合は注意が必要です。

最終的には、ファイルのサイズと検索する文字列の性質に応じて、どちらの方法を選ぶかを決めることになります。また、パフォーマンスとメモリ使用量のバランスを考慮することも重要です。

## `os`パケットと`ioutil`パッケージ
- **ioutilパッケージはgo1.16以降、非推奨になった。なので代わりに`os`パッケージを使うこと**
  - `ioutil.ReadFile`は **`os.ReadFile`** に、`ioutil.WriteFile`は **`os.WriteFile`** に置き換えられた。

## ディレクトリ(directory)存在確認方法
```go
package main

import (
    "fmt"
    "os"
)

func checkDirectory(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil // 存在しない
        }
        return false, err // その他のエラー
    }
    
    if !info.IsDir() {
        return false, fmt.Errorf("'%s' はディレクトリではありません", path)
    }
    
    return true, nil
}

func main() {
    dirPath := "./example"
    
    exists, err := checkDirectory(dirPath)
    if err != nil {
        fmt.Printf("エラー: %v\n", err)
        return
    }
    
    if exists {
        fmt.Printf("ディレクトリ '%s' は存在します\n", dirPath)
    } else {
        fmt.Printf("ディレクトリ '%s' は存在しません\n", dirPath)
    }
}
```