# Go `embed` パッケージまとめ

## 概要

`embed` パッケージは **Go 1.16** で導入された機能で、ファイルやディレクトリをバイナリに埋め込める。コンパイル時にファイルの内容を変数に埋め込めるため、実行時のファイル読み込みが不要。
ビルド時にファイルの内容がバイナリに取り込まれるため、デプロイ時に外部ファイルが不要になる。

---

## 基本的な使い方

### インポート

```go
import _ "embed"
```

> ファイルを `[]byte` や `string` に埋め込むだけなら、パッケージを直接使わないため `_` でインポートする。  
> `embed.FS` を使う場合は `embed` を直接インポートする。

---

## 埋め込みの種類

### 1. `string` に埋め込む

```go
package main

import (
    _ "embed"
    "fmt"
)

//go:embed hello.txt
var content string

func main() {
    fmt.Println(content)
}
```

### 2. `[]byte` に埋め込む

```go
//go:embed image.png
var imgData []byte
```

バイナリファイル（画像・証明書など）に適している。

### 3. `embed.FS` に埋め込む（複数ファイル / ディレクトリ）

```go
package main

import (
    "embed"
    "fmt"
    "io/fs"
)

//go:embed static
var staticFiles embed.FS

func main() {
    data, _ := staticFiles.ReadFile("static/index.html")
    fmt.Println(string(data))

    // ファイル一覧
    fs.WalkDir(staticFiles, ".", func(path string, d fs.DirEntry, err error) error {
        fmt.Println(path)
        return nil
    })
}
```

---

## `//go:embed` ディレクティブのパターン

| パターン | 説明 |
|---|---|
| `//go:embed file.txt` | 単一ファイルを埋め込む |
| `//go:embed dir/` | ディレクトリ全体を埋め込む |
| `//go:embed dir/*.html` | glob でマッチしたファイルを埋め込む |
| `//go:embed file1.txt file2.txt` | 複数ファイルをスペース区切りで指定 |

```go
//go:embed templates/*.html configs/config.yaml
var files embed.FS
```

> ディレクトリを指定した場合、`_` や `.` で始まる隠しファイルは **デフォルトで除外** される。  
> 含めたい場合は `all:` プレフィックスを使う。

```go
//go:embed all:static
var staticFiles embed.FS
```

---

## `embed.FS` の主なメソッド

```go
// ファイル読み込み
data, err := files.ReadFile("path/to/file.txt")

// ファイル情報取得
info, err := files.Open("path/to/file.txt")

// ディレクトリエントリ一覧
entries, err := files.ReadDir("dir")
for _, e := range entries {
    fmt.Println(e.Name(), e.IsDir())
}
```

---

## `net/http` との連携

```go
package main

import (
    "embed"
    "net/http"
    "io/fs"
)

//go:embed static
var staticFiles embed.FS

func main() {
    subFS, _ := fs.Sub(staticFiles, "static")
    http.Handle("/", http.FileServer(http.FS(subFS)))
    http.ListenAndServe(":8080", nil)
}
```

> `fs.Sub()` でサブディレクトリを root にして `http.FileServer` に渡すのが定番パターン。

---

## `html/template` との連携

```go
//go:embed templates/*.html
var templateFiles embed.FS

tmpl, err := template.ParseFS(templateFiles, "templates/*.html")
```

---

## 注意点

| 項目 | 内容 |
|---|---|
| ビルド時解決 | 実行時にファイルシステムへのアクセスは発生しない |
| 読み取り専用 | `embed.FS` は読み取り専用、書き込み不可 |
| パスの基点 | パスはソースファイルがあるパッケージディレクトリからの相対パス |
| モジュール外不可 | モジュールルート外のファイルは埋め込めない |
| `..` 不可 | `//go:embed ../outside` のような親ディレクトリへの参照は不可 |
| バイナリサイズ | 埋め込みファイルが大きいとバイナリサイズが増加する |

---

## ユースケース

- **静的ファイルの配布**（HTML / CSS / JS をシングルバイナリに同梱）
- **設定ファイルのデフォルト値**（デフォルト config.yaml を埋め込む）
- **TLS 証明書・秘密鍵**（証明書をバイナリに含める）
- **マイグレーションSQL**（DBマイグレーションファイルの同梱）
- **テンプレート**（Go テンプレートファイルの埋め込み）

---

## まとめ

```
embed パッケージ
├── string / []byte  →  単一ファイルの埋め込みに最適
└── embed.FS         →  複数ファイル・ディレクトリ・glob に対応
                         ├── ReadFile() / ReadDir() / Open()
                         ├── io/fs.FS インターフェースを満たす
                         └── http.FileServer や template.ParseFS と相性◎
```