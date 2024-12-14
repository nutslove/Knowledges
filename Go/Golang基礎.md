## `GOROOT`の下のディレクトリ
   | パス | 内容 |
   | --- | --- |
   | /bin | コンパイルされたものが格納される |
   | /pkg | プロジェクト内で再利用可能なパッケージ(コード)を格納 |
   | /src | `GOROOT`の下の`/src`ディレクトリは標準パッケージのソースコードが格納されている。 |

### 環境変数
- `GOPATH`
  - points to your go workspace
- `GOROOT`
  - points to your binary installation of Go

## Goプロジェクトの構造
- 以下githubリポジトリ参照！
  - **https://github.com/golang-standards/project-layout?tab=readme-ov-file**

### 現代的なGoプロジェクトの構造
- 一例
  ```
  project/
  ├── cmd/
  │   ├── app1/
  │   │   └── main.go
  │   └── app2/
  │       └── main.go
  ├── pkg/
  │   ├── logger/
  │   │   └── logger.go
  │   └── config/
  │       └── config.go
  ├── internal/
  │   ├── service/
  │   │   └── service.go
  │   └── repository/
  │       └── repository.go
  ├── go.mod
  ├── go.sum
  └── README.md
  ```

- 小規模の場合は`cmd/`ディレクトリがない場合もある  
  ```
  project/
  ├── main.go
  ├── internal/
  │   ├── service.go
  │   └── repository.go
  ├── pkg/
  │   └── utils.go
  ├── go.mod
  ├── go.sum
  └── README.md
  ```

- いくつかのOSSのGithubリポジトリを見た感じだと、**`pkg/`ディレクトリにビジネスロジックを書いているところが多そう**

#### `cmd/`ディレクトリ
- **役割**
  - アプリケーションのエントリーポイントを格納
- **特徴**
  - 各サブディレクトリが個別のバイナリを生成する
  - 通常、エントリーポイントとなる`main.go`のみを配置
  - ビジネスロジックは直接記述せず、内部パッケージを呼び出す形にする
- コード例 `cmd/app1/main.go`  
  ```go
  package main

  import "project/internal/service"

  func main() {
      service.Start()
  }
  ```

#### `pkg/`ディレクトリ
- **役割**
  - 再利用可能なライブラリや汎用的なコードを格納
- **特徴**
  - 外部プロジェクトからインポート可能
  - 他プロジェクトで再利用するユーティリティや汎用ライブラリを含む
- コード例 `pkg/logger/logger.go`  
  ```go
  package logger

  import "log"

  func LogInfo(message string) {
      log.Println("[INFO]", message)
  }
  ```

#### `internal/`ディレクトリ
- **役割**
  - アプリケーション内部でのみ利用するコードを格納
- **特徴**
  - `internal`ディレクトリ配下のパッケージは、Goのルールにより同一モジュール外からインポートできない
- コード例 `internal/service/service.go`  
  ```go
  package service

  import "fmt"

  func Start() {
      fmt.Println("Service started!")
  }
  ```

## Go Module
- Moduleはプロジェクト構成単位（1つのモジュールが1つのプロジェクト）
- Moduleは、関連するパッケージの集合であり、プロジェクト全体を一つの単位として扱う。これにより、バージョン管理や依存関係の解決が容易になる。
- **`go mod tiny`** コマンド
  - 依存関係を整理し、`go.mod`と`go.sum`ファイルを更新する
  - **`go.mod`の中には定義されているけど実際コードでは未使用のモジュールを`go.mod`から削除し、`go.mod`の中には定義されてないけどコードでは使っている必要なモジュールを`go.mod`に追加**
  - **依存関係のモジュールのソースコードのダウンロード（削除）もする**
  - `go.mod`ファイルと`go.sum`ファイルを、実際のソースコードと同期させる
  - これにより、プロジェクトの依存関係が正確に管理される
- `go build`は指定されたパッケージとそのすべての依存パッケージがコンパイルされ、静的にリンクされた単一の実行可能ファイルが生成される。この実行可能ファイルには、`import`しているすべてのパッケージのバイナリコードが含まれている。  
  Goは静的リンクを使用しているため、生成された実行可能ファイルは、実行に必要なすべてのコードを含んでいる。つまり、実行可能ファイルを別のシステムに移動しても、依存パッケージを別途インストールする必要がない。
  - `-o <バイナリ名>`でファイル名とは異なるバイナリファイルを生成できる
    - e.g. `go build -o logaas main.go`
### Go Module作成(初期化)
- `go mod init <Module名>`で初期化  
  - `go.mod`が作成される（これによって、現在のディレクトリがGoモジュールのルートディレクトリであることが示される）
- **`go.mod`ファイルは**
  - Goモジュールの依存関係管理ファイル
  - module名と使用しているGoのバージョン、使用している(依存関係の)モジュールが記載されている

## パッケージ関連
- Go v1.16までは`go get`は、パッケージをダウンロードした後に`go install`を実行してダウンロードしたパッケージのコンパイルまでしていた。
  - コンパイルされたパッケージ(バイナリ)は`$GOBIN`または`$GOPATH/bin`または`$HOME/go/bin`配下に配置される
- 現在(Go v1.17以降)は`go get`は`go.mod`の依存関係の調整にだけ使われて、パッケージのインストール(コンパイル＋`$GOBIN`または`$GOPATH/bin`への配置)には`go install`を使う
  - v1.16まで`go get`１つになっていた依存関係の管理とバイナリのビルド(コンパイル)が明確に分離された
  - `go get`でパッケージをダウンロードすると`go.mod`ファイルに依存関係が追加される  
  > Get resolves its command-line arguments to packages at specific module versions,
updates go.mod to require those versions, and downloads source code into the module cache.
- `go get`コマンドを使用してパッケージをインストールする際、Goはデフォルトで`GOPATH`と`GOROOT`環境変数を使用する。`GOROOT`はGo言語自体がインストールされているディレクトリを指し、`GOPATH`はワークスペース（プロジェクトのビルド、パッケージのインストールなどを行うディレクトリ）を指す。
- すべてのパッケージの最新化は`go get -u`後、`go mod tidy`を実行
  ~~~
  # Update the all packages in the current directory.
  go get -u
  # prune `go.sum` and organize by removing the unncessary checksums
  # add missing and remove unused modules
  go mod tidy
  ~~~
- DockerコンテナでGolangを使用する際に、`go.mod`と`go.sum`ファイルを作成しておき、Dockerfileの中で`go mod download`で一括ダウンロードするのは一般的なアプローチ
  - Dockerfile例  
    ~~~
    FROM golang:1.21-alpine

    WORKDIR /app

    COPY go.mod ./
    COPY go.sum ./

    ## Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
    RUN go mod download

    COPY *.go ./
    COPY ./templates ./templates
    COPY ./static ./static

    RUN go build -o /web_server

    EXPOSE 8080

    CMD [ "/web_server" ]
    ~~~
    - 以下Chat-GPTからの回答
      > DockerコンテナでGo言語を使用する際に、`go.mod`と`go.sum`ファイルを作成しておき、Dockerfileの中で`go mod download`コマンドを実行するのは一般的なアプローチです。この方法にはいくつかの利点があります：
      >
      > 1. **依存関係の明確化**:
      >     - `go.mod`はプロジェクトの依存関係を定義し、`go.sum`はそれらの依存関係の整合性を保証します。
      >     - これにより、Dockerコンテナ内でアプリケーションが正しくビルドされることを保証できます。
      >
      > 2. **キャッシュの最適化**:
      >     - `go mod download`をDockerfileで実行することで、依存関係をダウンロードし、それらをDockerのレイヤーとしてキャッシュします。
      >     - これにより、依存関係に変更がない限り、再ビルド時にDockerはキャッシュされたレイヤーを再利用でき、ビルド時間が短縮されます。
      >
      > 3. **再現性の向上**:
      >     - `go.mod`と`go.sum`ファイルを使用することで、どの環境でも同じバージョンの依存関係が使用されることが保証されます。
      >     - これにより、異なる開発環境やCI/CDパイプラインでのビルドの一貫性と再現性が向上します。
      > 
      > 一般的なDockerfileでは、次のようなステップでGoアプリケーションがビルドされます：
      >
      > 1. Goのベースイメージを指定する。
      > 2. ワークディレクトリを設定する。
      > 3. `go.mod`と`go.sum`ファイルをコンテナにコピーする。
      > 4. `go mod download`を実行して依存関係をプリフェッチし、キャッシュする。
      > 5. 残りのソースコードをコンテナにコピーする。
      > 6. `go build`コマンドを使用してアプリケーションをビルドする。
      >
      > これにより、Dockerコンテナ内でのGoアプリケーションのビルドが効率的かつ一貫性を持って行われます。

### `go get`と`go install`について
- **`go get`**
  - `go get`コマンドは、指定したパッケージのソースコードをインターネット上からダウンロードし、ローカルの作業環境に配置する。このコマンドは、指定したパッケージだけでなく、その依存関係にあるパッケージも一緒にダウンロードする。ダウンロードしたパッケージは、`GOPATH`環境変数で指定されたディレクトリの`src`フォルダ内に配置される。
- **`go install`**
  - `go install`コマンドは、ソースコードをコンパイルして実行可能なバイナリファイルを生成し、それを`GOPATH`環境変数で指定されたディレクトリの`bin`フォルダ内に配置する。このコマンドは、開発中のプロジェクトや依存するパッケージをビルドして、すぐに実行可能な状態にする。

### `go.sum`ファイル
- プロジェクトの依存関係として使用される各パッケージの特定のバージョンに対するchecksum（ハッシュ値）が記録されていて、パッケージの内容が変更されていないことを確認するために使用され、依存関係の中で意図しない変更や悪意のある変更がないかを検証することができる
- `go mod`コマンド（特に`go mod tidy`や`go get`など）を使用する際に自動的に生成または更新される。このファイルは通常、ソースコード管理システム（例えばGit）にコミットされるべき。これにより、プロジェクトをクローンまたはダウンロードするすべての開発者が、同じ依存関係を使用してプロジェクトをビルドできるようになる。

### Windowsで`go get`時、"Access is denied"が出た時の対応
- 事象
  - WindowsでGoをインストールするとCドライブの`Program Files`フォルダ内にインストールされ、  
    その後、一般ユーザで`go get`でパッケージをダウンロードしようとすると、  
    `Program Files`フォルダにアクセス権限がなくて`Access is denied`エラーが出る
- 解決方法
  - ユーザの環境変数に`GOPATH`変数を追加し、値に`C:\Users\<ユーザ名>`など、一般ユーザもアクセス可能なパスを指定して保存する（設定後再ログイン or 再起動が必要）


## その他Goについて色々
- GoはClassがない。ただ、Struct(構造体)で同じようなことができる
- Goはtry catch(except)ではなく、errorというエラー専用型(interface)がある
- Goにwhileはない
- gofmtコマンドを使うとgoのフォーマットに変換してくれる  
`gofmt -w <対象goファイル>`
- `main()`関数以外はmainもしくはinit関数内で明示的に呼び出す必要がある
- `init()`関数が`main()`関数より先に実行される
- GoにもGCが存在する
  - ただ、fileなど明示的に開放しなければいけないものもある
  - 明示的に開放する必要があるものはdeferを使って開放する
- Goは大文字/小文字、全角/半角は別の文字として扱われる
- Goは宣言された変数は必ず利用されている必要がある  
  (変数宣言だけしといて使わないとエラーになる)
- Goは型変換をConversionという（他の言語ではCastingというらしい）  
  https://go.dev/ref/spec#Conversions  
  https://go.dev/doc/effective_go#conversions
- Goも本当は構文の最後に`;`が付くけど、コンパイラーがコンパイル時に付けてくれるので人が意識する(付ける)必要はない。ただ、for文やif文など1行に複数の構文を書く場合は明示的に`;`を付ける必要がある
- main関数の中で`return`でプログラムが終了する

## Goインストール（Linux）
- `wget https://dl.google.com/go/go1.18.4.linux-amd64.tar.gz`
- `rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.4.linux-amd64.tar.gz`
- `export PATH=$PATH:/usr/local/go/bin`
- `go version`

# Goの文法
## 変数の書き方
1. varを使った定義
   - 関数の外側でも定義可  
      → その場合コード(パッケージ)内のすべての関数で参照可  
      → 関数内で定義した場合は省略型`:=`と同様にその関数内でだけ使用可能  
     - 基本
         ~~~go
         var 変数 型
         変数 = 値
         
         var 変数 型 = 値  
         ~~~
      - 同じ型の変数を複数作成する場合
         ~~~go
         var 変数1,変数2[,変数3,・・・] 型
         変数1,変数2[,変数3,・・・] = 値1,値2[,値3,・・・]

         var 変数1,変数2[,変数3,・・・] 型 = 値1,値2[,値3,・・・]
         ~~~
      - 型の宣言と同時に値を代入する場合は型の指定を省略することができる
         (代入される値から型が推論される)
         ~~~go
         var 変数 = 値
         ~~~

2. 省略型`:=`での定義
   - 関数の中でのみ定義可
   - 型の指定は不要(自動的に型が設定される)
      ~~~go
      変数 := 値
      ~~~
    - 省略型に整数/少数の値を代入する場合、自動的に`int`/`float64`型になる

## `...` (可変長引数) について（Lexical elementsと呼ぶらしい）
- **可変長引数**という
- https://go.dev/ref/spec#Operators_and_punctuation
- `...<型>`
  - スライス (つまり`[]<型>`と同じ) を作成
  - 例
    ~~~go
    func main() {
	    foo(1, 2, 3, 4, 5, 6)
    }

    func foo(x ...int) {
	    fmt.Println(x) ---------→ [1 2 3 4 5 6]と出力される
	    fmt.Printf("%T\n", x) --→ []intと出力される
    }
    ~~~
- `<型>...`
  - スライスから1つずつ展開する
  - 例えば`a := []int{1,2,3}`は`a[0]`,`a[1]`,`a[2]`に展開される
  - 例
    ~~~go
    func main() {
      xi := []int{1,2,3,4,5}
      foo(xi...) -------------→ スライス(xi)を展開して関数に渡す
    }
    func foo(xi ...int) { ----→ 展開されたものを再度スライスにする
      fmt.Println(xi) --------→ [1 2 3 4 5]が出力される
      fmt.Printf("%T\n", xi) -→ []intが出力される
    }
    ~~~
    - 以下のように引数なしでもできる
      ~~~go
      func main() {
        foo() --------→ この場合xiに連携された値はnil
      }
      func foo(xi ...int) {
        fmt.Println(xi) --------→ []が出力される
        fmt.Printf("%T\n", xi) -→ []intが出力される
      }
      ~~~

### 可変長引数の注意点
1. 引数が複数ある場合、可変長引数は最後の引数である必要がある
    - OK  
      ```go
      func push(a []string, v ...string) {}
      ```
    - NG  
      ```go
      func push(v ...string, a []string) {}
      ```
2. 可変長引数には同一の型の値のみ使用できる

## 戻り値を`_`で捨てる
- 戻り値などで定義は必要だけど使わない変数は`_`で捨てる  
  - 例
    ~~~go
    変数, _ = strconv.Atoi(string)
    ~~~
- 変数の型を確認する方法  
  ~~~go
  fmt.Printf("%T\n", <確認したい変数>)
  ~~~
- stringの中でダブルクォーテーションを扱いたい場合、``で囲む
  ~~~go
  var a string = `She said "You are doing well"`
  var b string = `I said
  "yeahaaaaaaaaaaaaaa
  hoooooooooooooooooo"
  `
  ~~~

## Array(配列)
- Format
  1. `var <変数名>[配列数]<型>`
      ~~~go
      var x [5]int
      x[0] = 1
      x[1] = 2
      ~~~
  2. `<変数名> := [配列数]<型>{配列値}`
      ~~~go
      b := [5]int{1, 2, 3, 5, 7}
      ~~~
- `len(配列変数名)`で配列数を確認できる
- Arrayは固定長で`append`による要素の追加ができない
  > **Warning**  
  > GoではArrayの代わりにSlicesを使うことが推奨されている  
  > https://go.dev/doc/effective_go#arrays

## Slices（スライス）
- 参照URL
  - https://go.dev/tour/moretypes/11#:~:text=The%20capacity%20of%20a%20slice,provided%20it%20has%20sufficient%20capacity.
- スライスはArrayの参照であって、値の実体はArrayにある
- __length (長さ)__
  - The length of a slice is the number of elements it contains.
  - 現在スライスが持っている要素数
  - `len(スライス変数名)`でスライスの長さが確認できる
- __capacity (容量)__
  - The capacity of a slice is the number of elements in the underlying array, counting from the first element in the slice.
  - Arrayからスライスを作成した場合、元のArrayの要素数
  - 別にcapacityの数までしか要素を作成できない等の制約はなく、capacity数以上の要素を追加できる(capacity数以下に要素の削除もできる)
  - `cap(<スライス変数名>)`でスライスの容量が確認できる
- Arrayからではなく、最初からSliceを作成した場合はlength=capacityとなる
- Format
  1. `<変数> := []<型>{Values}`
      ~~~go
      x := []int{1, 2, 3, 5, 7}
      myslices := []int{} ------→ valueを入れずに作成することもできる
      ~~~
  2. `<変数> := make([]<型>, <長さ>, <容量>)`  
     `<変数> := make([]<型>, <長さ>)` → 容量を省略した場合は容量=長さ
      ~~~go
      x := make([]int, 5, 10)
      y := make([]int, 5)
      ~~~
  3. arrayからslicingしてスライスを作成。その場合容量(cap)はarrayの要素数になる
      ~~~go
      arr1 := [6]int{10, 11, 12, 13, 14,15}
      myslice := arr1[2:4]

      fmt.Printf("myslice = %v\n", myslice) ------→ [12 13]
      fmt.Printf("length = %d\n", len(myslice)) --→ 2
      fmt.Printf("capacity = %d\n", cap(myslice)) -→ 4
      ~~~
  4. `var <変数> []型`
      ~~~go
      var bytes []byte
      bytes = append(bytes, 64)
      ~~~
      ※byteは`uint8`型の別名。uint8型は8bit、つまり1バイト分(0~255)の表現が可能で、データを1バイトごとに分割して扱う。[参照ページ](https://qiita.com/s_zaq/items/a6d1df0d1bf79f5bcc2a)
- ArrayとSliceの違いについて  
  - https://qiita.com/seihmd/items/d9bc98a4f4f606ecaef7
  - https://qiita.com/tchnkmr/items/10071a53a8bce87b62a3
- Arrayと同様に`len(Slice名)`でSliceの長さを確認できる
- Sliceは`append`による要素の追加ができる  
  ~~~go
  slice := []int{1, 2, 3}
  slice = append(slice, 4)
  ~~~
- 要素の削除は``で行う
- rangeを使ってforでSliceのループ処理ができる
  ~~~go
  x := []int{10, 20, 30}
  for index, value := range x {
    fmt.Println(index, value)
  }
  ~~~
  → `0 10\n 1 20\n 2 30\n`が出力される
    - sliceでなくても`range`でindexとvalueを取得できる  
      下記は文字のindexと各文字のASCIIコードが表示される
      ~~~go
	    m := "Hello Lee!"
	    for k, v := range m {
		    fmt.Println("Key: ", k, "Value: ", v)
	    }
      ~~~
- Sliceを初期化する方法（By Chat-GPT）
  - リクエストの頻度が低く、次のリクエストが車で間が開くような場合は2.で１回メモリを解放した方が良い。
> Go言語でスライスを初期化する上記の3つの方法は、それぞれ異なる挙動と使用シナリオを持っています。それぞれの方法の違いを詳しく見てみましょう。
> 
> 1. **新しい空のスライスを割り当てる (`s = []Type{}`)**:
>    - これはスライス`S`を新しい空のスライスで完全に置き換えます。
>    - 新しいスライスは元のスライスとは異なるメモリ領域を指します。
>    - 元のスライスに対する参照は影響を受けず、元のスライスはガーベージコレクションの対象になります（他に参照がない場合）。
>    - 使用シナリオ: スライスの内容を完全にリセットし、新しいメモリ領域を割り当てたい場合。
>
> 2. **`nil`を割り当てる (`s = nil`)**:
>    - これによりスライスは`nil`に設定されます。
>    - `nil`スライスは長さと容量が0ですが、メモリを割り当てていません。
>    - 元のスライスのデータは残っていますが、`nil`に設定されたスライスからはアクセスできません。
>    - 使用シナリオ: スライスを完全に空にし、メモリを割り当てずにリソースを解放したい場合。
>
> 3. **長さを0にする (`s = s[:0]`)**:
>    - これはスライスの長さを0に設定しますが、元のスライスの容量（及び背後にある配列）は変更されません。
>    - スライスは空に見えますが、元のデータはメモリ上に残っており、容量を超えない限り追加の割り当てなしに再利用できます。
>    - 使用シナリオ: メモリを再利用しつつ、スライスを一時的に空にしたい場合。
>
> これらの違いを理解することで、アプリケーションの要件に応じて最適な初期化方法を選択できます。例えば、メモリの再利用を重視する場合は3番目の方法が適していますが、スライスを完全にリセットしたい場合は1番目または2番目の方法が良いでしょう。
> 
- Sliceの最後の要素を取得する方法
  - pythonみたいに`<Slice名>[-1]`では取得できない(エラーになる)
  - `<Slice名>[len(<Slice名>)-1]`というふうにSliceの長さから-1して最後の要素を取得する
    - 例：`fmt.Println("Metric:",*resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value)`

### スライスの重複排除
- `golang.org/x/exp/slices`ライブライで、`Sort()`関数でSortさせてから、`Compact()`関数を実行することで重複排除を行うことができる
- 例  
  ```go
  package main

  import (
  	"fmt"
  	"golang.org/x/exp/slices"
  )

  func main() {
  	strs := []string{"A", "B", "C", "A", "A", "D", "B", "E", "F"}
  	slices.Sort(strs)
  	unique := slices.Compact(strs)
  	fmt.Printf("%+v\n", unique)
    // [A B C D E F]
  }

  ```

## Map
- Format
  1. `var 変数 map[<Key型>]<Value型>`
  2. `変数 := make(map[<Key型>]<Value型>)`  
     - 空のMapが作成される
     - 初期化(定義と同時に値を代入)する場合は 3. の方法がもう少し性能が良いらしい
  3. `変数 := map[<Key型>]<Value型> { Key1: Value1, Key2: Value2, ・・・, }`  
    → 最後の要素の後にも`,`が必要
      ~~~go
      m := map[string]int {
        "Lee": 35,
        "Yamagiwa": 28,
      }
      ~~~
- MapのValue参照
  - `<Map変数名>[<Key名>]`  
    → `fmt.Println(m["Lee"])`
- Mapは該当のKeyが存在しなくてもエラーにならず、Zero Value(初期値)を返すので要注意!  
  Mapは実は2つの戻り値があって2つ目の戻り値(bool)でそのKeyが存在有無を判断する
  ~~~go
  if v, ok := m["Kim"]; ok {
    fmt.Println(v)
  }
  → "Kim"は存在しないので何も表示されない
  if v, ok := m["Lee"]; ok {
    fmt.Println(v)
  }
  → "Lee"は存在するので35が表示される
  ~~~
- 要素追加
  - `<Map変数名>[<追加するKey名>] = <追加するValue>`  
    → `m["Kim"] = 51`
- 要素削除
  - `delete(<Map変数名>[<削除するKey名>])`  
    → `delete(m, "Kim")`  
  - 削除するKeyが存在しなくてもエラーにならないので本当に削除されたか確認するためには2つ目の戻り値(bool)で確認してから削除する
    ~~~go
    if v, ok := m["Kim"]; ok {
      delete(m, "Kim")
    }
    ~~~
- rangeを使ってforでMapのループ処理ができる
  ~~~go
  for k, v := range m {
    fmt.Println("Key: ", k, "Value: ", v)
  }
  ~~~
  - **Mapのループで取り出される要素の順番はランダムなので要注意！(意図的に設計されたそう)**
- Mapにnilを代入することもできて、**nilのMapにデータを代入しようとするとプログラムがCrashする。**  
  **nilのMapにデータを入れる前に新しいMapを代入してからデータを入れること！**
  - NG
    ~~~go
    aMap := map[string]int{}
    aMap = nil
    aMap["key1"] = 1 --→ "panic: assignment to entry in nil map"エラーが出てプログラムがCrashする
    ~~~
  - OK
    ~~~go
    aMap := map[string]int{}
    aMap = nil
    if aMap == nil {
      aMap = map[string]int{}
    }
    aMap["key"] = 10 --→ 上でmapを代入したので正常にMapにデータを入れることができる
    ~~~

## 定数（const）
- 作成した後に値の変更ができない
  ~~~go
  const 定数 = 値
  const 定数 型 = 値
  ~~~
  - 複数を`()`内でまとめて定義することもできる
    ~~~go
    const (
      a = 42
      b = 42.78
      c = "James Bond"
    )
    ~~~
- 定義時に型を指定することもできる
  ~~~go
  const (
    a int = 42
    b float64 = 42.78
    c string = "James Bond"
  )
  ~~~
- 変数とは違い、型を指定しない場合、
  定義時に代入する値から型を推論するのではなく、使われる時に型を推論する

## 関数
- 関数もStringやintと同様にTypeの1つである  
  → 関数もreturnすることができる
- 定義方法にいくつかパターンがある 
  1. 引数や戻り値がないパターン
      ~~~go
      func 関数名() {
        ・・・処理・・・
      }
      ~~~
  2. 引数だけあるパターン
      ~~~go
      func 関数名(引数 引数の型) {
        ・・・処理・・・
      }
      ~~~
  3. 引数、戻り値があるパターン
      ~~~go
      func 関数名(引数 引数の型) 戻り値の型 { 
         ・・・処理・・・
         return 戻り値
      }
      ~~~
      - 例
        ~~~go
        func woofoo(s string) int {
          fmt.Println(s)
          return 4
        }
        ~~~
  4. 複数の引数、戻り値があるパターン
      ~~~go
      func 関数名(引数1 引数1の型, 引数2 引数2の型[,・・・]) (戻り値1の型, 戻り値2の型[,・・・]) {
         ・・・処理・・・
         return 戻り値1, 戻り値2[,・・・]
      }
      ~~~
      - 例
        ~~~go
        func test(s1 string, s2 string, s3 string) (int, bool, string) {
	      ・・・処理・・・
      	  return 4, true, "Wow!"
        }
        ~~~
      - すべての引数の型が同じの場合は型は最後に1回だけ書いても良い
        ~~~go
        func yeah(s1, s2, s3 string)
        ~~~
  5. 無名関数
      ~~~go
      func() {
        ・・・<処理>・・・
      }()
      ~~~
      - 例
        ~~~go
        func(x int) {
          fmt.Println("Age:", x)
        }(36)
        ~~~
  6. 関数を変数に代入して変数から呼び出すこともできる
       - 例
        ~~~go
        f := func() {
          fmt.Println("Yeah")
        }
        f()

        y := func(x int) {
          fmt.Println("Next year:",x)
        }
        y(2023)
        ~~~
  7. 関数から関数を返す
      - 例（func bar()の次の`func() int`が戻り値）
        ~~~go
        func main() {
          fmt.Println(bar()()) ------→ 1212と出力される

          x := bar()
          i := x()
          fmt.Println(i) ------→ 同様に1212と出力される

          x1 := bar()
          fmt.Println(x1()) ------→ 同様に1212と出力される
        }
        func bar() func() int {
          return func() int {
            return 1212
          }
        }
        ~~~
  8. Callback関数
      - Callback関数とは関数の引数としてfuncを引き渡すこと
      - 例
        ~~~go
        package main

        import (
          "fmt"
        )

        func main() {
          t := evenSum(sum, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}...)
          fmt.Println(t)
        }

        func sum(x ...int) int {
          n := 0
          for _, v := range x {
            n += v
          }
          return n
        }

        func evenSum(f func(x ...int) int, y ...int) int {
          var xi []int
          for _, v := range y {
            if v%2 == 0 {
              xi = append(xi, v)
            }
          }
          total := f(xi...)
          return total
        }
        ~~~
#### Signature（シグネチャ）
- 関数やメソッドの定義を指す用語。これには、**関数名**、**パラメータリスト（引数の型と名前）**、**戻り値の型**が含まれる。
- 例えば以下の関数のSignatureは`func Add(a int, b int) int`  
  ~~~go
  func Add(a int, b int) int {
      return a + b
  }
  ~~~

## Struct (構造体)
- 色んな型をひとまとめにしたもの
- 他の言語のClassのような感じで、1つのstructに対して (下のp1とp2のように) 何回でも変数宣言できる
- Format
  ~~~go
  type <Struct名> struct {
    <field1名> <型>
    <field2名> <型>
          ・
          ・
  }
  ~~~
  - 例
    ~~~go
    type person struct {
      first string
      last  string
      age   int
    }

    func main() {
      p1 := person {
        first: "James",
        last: "Bond",
        age: 28, -----→ 最後の要素の後ろにも,が要る
      }

      p2 := person {
        first: "Joonki",
        last: "Lee",
        age: 35,
      }

      fmt.Println(p1)  
	  fmt.Println(p1.first)  
	  fmt.Println(p1.age)
    }
    ~~~
  - 例２（Sliceがある場合）
    ~~~go
    type metrics struct {
      namespace  string
      metricname []string
    }

    var test = metrics {
      namespace:  "oci_compute",
      metricname: []string{
        "cpu",
        "memory",
        "diskio",
      },
    }
    ~~~

- field名を指定せずに値だけ代入することもできる。  
  ただし、構造体の定義が変更された場合 (例えば、フィールドが追加された場合や順序が変更された場合)、意図しないバグを引き起こす可能性があるため、明確さと将来の変更への対応を考慮して、フィールド名を指定して値を代入することが一般的に推奨される。
  ~~~go
  type Person struct {
      Name string
      Age  int
  }

  func main() {
      // フィールド名を省略して構造体に値を代入
      p := Person{"Alice", 30}
      fmt.Println(p)
  }
  ~~~
- field名を指定して値を代入する時は、fieldの定義順序と異なる順序で値を設定しても問題ない
  - 例  
    ~~~go
    type Person struct {
        Name string
        Age  int
    }

    func main() {
        // フィールド名を指定しているので、代入する順序は自由
        p := Person{Age: 30, Name: "Alice"}
        fmt.Println(p)
    }
    ~~~

### Embedded structs（構造体の埋め込み）
- 他の言語のClassの継承みたいな感じ
- 既存のstructの中のfieldを継承し、追加のfieldを追加して使う
- Format
  ~~~go
  type <Struct名> struct {
    <継承するStruct名>
    <追加field1名> <型>
    <追加field2名> <型>
            ・
            ・
  }
  <変数> := <Struct名> {
    <継承したStruct名>: <継承したStruct名> {
        <継承したStruct名の中のfield1>: <値>,
        <継承したStruct名の中のfield2>: <値>,
                      ・
                      ・
    },
    <追加field1名>: <値>,
    <追加field2名>: <値>,
            ・
            ・
  }
  ~~~
- 例
  ~~~go
  type person struct {
    name string
    sex string
    age int
  }
  type killer struct {
    person
    pay int
    country string
  }
  agent := killer {
    person: person {
      name: "Anonymous",
      sex: "Unknown",
      age: 100,
    },
    pay: 500000,
    country: "USA",
  }
  fmt.Println(agent.name, agent.sex, agent.age, agent.pay, agent.country)
  -→ agent.person.nameのようにpersonを入れなくて良い 
  ~~~
- 埋め込んだ型のメソッドセットを埋め込み元でも使用できる
- Goでは、埋め込み型のメソッドよりも、埋め込み元で直接定義されたメソッドが優先される

### Anonymous structs
- `type <struct名>`でstructを宣言せず、1回限りの (1つの変数だけで使える) struct
  ~~~go
  p1 := struct {
      name string
      sex string
      age int
  }{
      name: "Joonki Lee",
      sex: "male",
      age: 35,
  }
  ~~~

## Methods（メソッド）
- *a method is just a function with a receiver argument.*
  - つまり、MethodはStructをReceiver引数として持つ関数
- A method is nothing more than a FUNC attached to a TYPE
- Method(関数)は、特別なreceiver引数(Struct)を取る
- receiverはfuncキーワードとMethod名の間に自身の引数リストで表現
  - `func (<receiver名> <Struct名>) Method名([引数]) [戻り値の型] { ・・・処理・・・ }`
  - `receiver名`をMethod内でtypeの値を扱える
- 呼び出す側はMethodのreseiverのtypeの値を含む変数を使って`<変数名>.<Method関数名>()`で呼び出す
- 例
  ~~~go
  type person struct {
  	firstname string
  	lastname  string
  }

  type secretAgent struct {
  	person
  	ltk bool
  }

  func (s secretAgent) speak() { ----------------→ これがMethod
  	fmt.Println("I am", s.firstname, s.lastname)
  	fmt.Println(s)
  }

  func main() {
  	sa1 := secretAgent{
  		person: person{
  			"James",
  			"Bond",
  		},
  		ltk: true,
  	}

  	sa2 := secretAgent{
  		person: person{
  			"Miss",
  			"Moneypenny",
  		},
  		ltk: true,
  	}

  	sa1.speak() ---→ "I am James Bond"と出力される
  	sa2.speak() ---→ "I am Miss Moneypenny"と出力される
  }
  ~~~

## Interfaces
- InterfaceはMethodの集合であり、Typeである
- Interfaceは具現しなければいけないMethodの集合を表した抽象Type
- あるData TypeがInterfaceを満たすためには、そのInterfaceが求めるすべてのMethodを具現しなければいけない
- Interfaceを通じて動作を定義できる
  - 下記例の`GetArea() int`や`speak()`
- InterfaceのTypeはInterfaceに指定したMethodを持つStructのTypeになれる（Interfaceは値が1つ以上のTypeになり得るようにする）
  - Interfaceに指定したMethodを持つStructが複数ある場合はそのすべてのStructのTypeになれる
  - 例えば例１の場合、`Figure`は`Circle`Typeにも`Square`Typeにもなり得る
- Structがある → そのStructをReceiver引数として持つMethodがある → そのMethodを持つInterfaceがある
- Format
  ~~~go
  type <Interface名> interface {
    <Method名([引数])> <Methodでreturnされる型>
  }
  ~~~
- 例１
  ~~~go
  type Circle struct {
	  Radius int
  }

  func (c Circle) GetArea() int {
	  return 3 * c.Radius * c.Radius
  }

  type Square struct {
	  Height int
  }

  func (s Square) GetArea() int {
	  return s.Height * s.Height
  }

  type Figure interface {
	  GetArea() int
  }

  func DisplayArea(f Figure) { ---> Figureのf変数(引数)にCircle structまたはSquare structの値が入る
	  fmt.Printf("%T\n", f)
	  fmt.Printf("面積は%vです\n", f.GetArea())
  }

  func main() {
	  circle := Circle{Radius: 2}
	  DisplayArea(circle) ------------→ 3*2*2で"面積は12です"と出力される

	  square := Square{Height: 3}
	  DisplayArea(square) ------------→ 3*3で"面積は9です"と出力される
  }
  ~~~
- 例２  
  ~~~go
  type Stringer interface {
    String() string
  }

  type Student struct {
    Name string
    Age int
  }

  func (s Student) String() string {
    return fmt.Sprintf("Hey! I am %d years old and my name is %s", s.Age, s.Name) // fmt.Sprintfはターミナルに出力せず、出力値を変数に保存したりする際に利用
  }

  func main() {
    student := Student{
      "Lee",
      32
    }

    stringer := Stringer(student)
    // 以下のようにinterface型の変数を定義し、structインスタンスを代入することもできるが、上記の書き方がより多く使われるらしい
    // var stringer Stringer
    // stringer = student

    fmt.Printf("%s\n", stringer.String())
  }
  ~~~
- 例３
  ~~~go
  type person struct {
	  first string
	  last  string
  }

  type secretAgent struct {
	  person
	  ltk bool
  }

  func (s secretAgent) speak() {
	  fmt.Println("I am", s.first, s.last, " - the secretAgent speak")
  }

  func (p person) speak() {
	  fmt.Println("I am", p.first, p.last, " - the person speak")
  }

  type human interface {
	  speak()
  }

  func bar(h human) {
    // human Interfaceが指定しているspeak() Methodを持つStructがpersonとsecretAgent、２つあるのでh(human)のtypeはpersonとsecretAgent両方になり得る
	  switch h.(type) {
	  case person:
      fmt.Printf("%T\n",h) -----------→ "main.person"と表示される
      fmt.Printf("%T\n",h.(person)) --→ "main.person"と表示される
		  fmt.Println("I was passed into bar. I am person", h.(person).first)
	  case secretAgent:
		  fmt.Println("I was passed into bar. I am secretAgent", h.(secretAgent).first)
	  }
  }

  func main() {
	sa1 := secretAgent{
		person: person{
			"James",
			"Bond",
		},
		ltk: true,
	}

	sa2 := secretAgent{
		person: person{
			"Miss",
			"Moneypenny",
		},
		ltk: true,
	}

	p1 := person{
		first: "Dr.",
		last:  "Yes",
	}

	bar(sa1)
	bar(sa2)
	bar(p1)
  }
  ~~~
- 例４ ([sortパッケージ(interface)](https://pkg.go.dev/sort#Interface))
  - `sort.Sort`でsortするためには上のURLのInterfaceで定義している`Len() int`、`Less(i, j int) bool`、`Swap(i, j int)` Methodを具現する必要がある
  ~~~go
  package main

  import (
  	"fmt"
  	"sort"
  )

  type S1 struct {
  	F1 int
  	F2 string
  	F3 int
  }

  // We want to sort S2 records based on the value of F3.F1
  // Which is S1.F1 as F3 is an S1 structure
  type S2 struct {
  	F1 int
  	F2 string
  	F3 S1
  }

  type S2slice []S2

  // Implementing sort.Interface for S2slice
  func (a S2slice) Len() int {
  	return len(a)
  }

  // What field to use for comparing
  func (a S2slice) Less(i, j int) bool {
  	return a[i].F3.F1 < a[j].F3.F1
  }

  func (a S2slice) Swap(i, j int) {
  	a[i], a[j] = a[j], a[i]
  }

  func main() {
  	data := []S2{
  		S2{1, "One", S1{1, "S1_1", 10}},
  		S2{2, "Two", S1{2, "S1_1", 20}},
  		S2{-1, "Two", S1{-1, "S1_1", -20}},
  	}
  	fmt.Println("Before:", data)
  	sort.Sort(S2slice(data))
  	fmt.Println("After:", data)

  	// Reverse sorting works automatically
  	sort.Sort(sort.Reverse(S2slice(data)))
  	fmt.Println("Reverse:", data)
  }
  ~~~
- 参考URL
  - https://go.dev/play/p/rZH2Efbpot
  - https://dev-yakuza.posstree.com/golang/interface/

### `map[string]interface{}`（=`map[string]any{}`）について
- **そもそも`interface{}`は、empty interfaceでどんな型の値でも格納できるもの**
- goのv1.18から`any`というのが追加されたけど、これは`interface{}`のalias
- `map[string]interface{}`からのデータ抽出の例
  - valueに更に`map[string]interface{}`が設定されている場合、  
    `変数[key名].(map[string]interface{})[key名]`のように下の階層のvalueにアクセスするためには`.(map[string]interface{})`が必要  
    ```go
    var Flavors = map[string]interface{}{
      "m1.tiny": map[string]interface{}{
        "requests": map[string]interface{}{
          "cpu":    "125m",
          "memory": "640Mi",
        },
        "limits": map[string]interface{}{
          "cpu":    "500m",
          "memory": "1Gi",
        },
        "jvm_heap": "512M",
        "jvm_perm": "128M",
      },
    }

    flavor := Flavors["m1.tiny"].(map[string]interface{})
    requests := flavor["requests"].(map[string]interface{})
    limits := flavor["limits"].(map[string]interface{})
    requests_cpu := requests["cpu"]
    requests_memory := requests["memory"]
    limits_cpu := limits["cpu"]
    limits_memory := limits["memory"]
    jvm_heap := flavor["jvm_heap"]
    jvm_perm := flavor["jvm_perm"]
    fmt.Println("flavor:", flavor) // "flavor: map[jvm_heap:512M jvm_perm:128M limits:map[cpu:500m memory:1Gi] requests:map[cpu:125m memory:640Mi]]"
    fmt.Println("requests:", requests) // "requests: map[cpu:125m memory:640Mi]"
    fmt.Println("requests_cpu:", requests_cpu) // "requests_cpu: 125m"
    fmt.Println("requests_memory:", requests_memory) // "requests_memory: 640Mi"
    fmt.Println("limits_cpu:", limits_cpu) // "limits_cpu: 500m"
    fmt.Println("limits_memory:", limits_memory) // "limits_memory: 1Gi"
    fmt.Println("jvm_heap:", jvm_heap) // "jvm_heap: 512M"
    fmt.Println("jvm_perm:", jvm_perm) // "jvm_perm: 128M"
    ```

### Type assertions
- `interface{}`型の変数に割り当てた値は、実行(ランタイム)時にその値の実際の型(e.g. string、int)に変換して使う必要があり、その型変換機能をType assertionsという
- 書き方
  - `<interface{}型変数>.(変換したい型)`
- ２つ目の戻り値のための変数を用意すると`interface{}`型に格納された値が変換したい型に一致すれば`true`が、一致しなければ`false`が返ってくる。  
  １つ目の戻り値には変換したい型のzero valueが返ってくる。  
  ２つ目の戻り値のための変数を用意してない場合は、`interface{}`型に格納された値が変換したい型に一致しないとpanicになる。
- `switch`文にて`<interface{}型変数>.(type)`で`interface{}`内の型による分岐処理を実装できる
  - `<interface{}型変数>.(type)`は現在のデータ型を返す
```go
func main() {
	var i interface{} = "hello"

	s := i.(string)
	fmt.Println(s) // "hello"

	s, ok := i.(string)
	fmt.Println(s, ok) // "hello true"

	f, ok := i.(float64)
	fmt.Println(f, ok) // "0 false"

	f = i.(float64) // panic
	fmt.Println(f)

  // 型チェック
  if v, ok := i.(string); ok {
    fmt.Println(v) // "hello"
  }
  
}

// swhitch文と`<interface{}型変数>.(type)`で型チェック
func checkType(arg interface{}) {
  switch arg.(type) {
  case bool:
    fmt.Println("This is a bool", arg)
  case int, int8, int16, int32, int64:
    fmt.Println("This is a int", arg)
  case float64:
    fmt.Println("This is a float", arg)
  case string:
    fmt.Println("This is a string", arg)
  case nil:
    fmt.Println("This is a nil", arg)
  default:
    fmt.Println("Unknown Type", arg)
  }
}
```

### `reflect.TypeOf()`による`interface{}`に割り当てられた値の型確認
- `reflect`パッケージの`TypeOf()`メソッドで、`interface{}`型に格納された値の実際の型を確認できる
  ```go
  var a interface{} = 15
  b := a
  c := a.(int)

  fmt.Println("a type:", reflect.TypeOf(a)) // int
  fmt.Println("b type:", reflect.TypeOf(b)) // int
  fmt.Println("c type:", reflect.TypeOf(c)) // int
  ```
- もちろん普通の型確認にも使える
  ```go
  import (
      "fmt"
      "reflect"
  )

  func main() {
      var x int = 10
      var y string = "hello"
      var z float64 = 3.14

      fmt.Println(reflect.TypeOf(x)) // "int"
      fmt.Println(reflect.TypeOf(y)) // "string"
      fmt.Println(reflect.TypeOf(z)) // "float64"
  }
  ```

## CallBack
- 引数として関数を引き渡すこと
- 例
  ~~~go
  package main

  import (
	"fmt"
  )

  func main() {
	ii := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	s := sum(ii...)
	fmt.Println("all numbers", s)

	s2 := even(sum, ii...)
	fmt.Println("even numbers", s2)
  }

  func sum(xi ...int) int {
	total := 0
	for _, v := range xi {
		total += v
	}
	return total
  }

  func even(f func(xi ...int) int, vi ...int) int {
	var yi []int
	for _, v := range vi {
		if v%2 == 0 {
			yi = append(yi, v)
		}
	}
	return f(yi...)
  }
  ~~~

## Closure
- one scope enclosing other scopes
  - variables declared in the outer scope are accessible in inner scopes
- ポイントとしては、**関数の戻り値に関数を指定する**ことと**戻り値の関数は無名関数**である
- 戻り値の関数が格納されている変数を使い続ける限り、変数を初期化せずに値を保持しておきたい時に使う
- 例
  ~~~go
  func main() {
  	a := incrementor()
	  b := incrementor()
  	fmt.Println(a()) ---> 1
	  fmt.Println(a()) ---> 2
  	fmt.Println(a()) ---> 3
	  fmt.Println(b()) ---> 1
  	fmt.Println(b()) ---> 2
	  fmt.Println(b()) ---> 3
  }

  // incrementor()関数がClosure
  func incrementor() func() int {
	  var x int
	  return func() int {
		  x++
		  return x
	  }
  }
  ~~~
- 参考URL
  - https://golangstart.com/go_closure/
  - https://go.dev/tour/moretypes/25
  - https://go-tour-jp.appspot.com/moretypes/25

## Recursion
- 関数が自分自身を呼び出すこと
- **RecursionでできることはLoopでもできる**
- 例
  ~~~go
  func main() {
  	n := factorial(4) ---> 24(4 * 3 * 2 * 1)
	  fmt.Println(n)
  }

  func factorial(n int) int {
	  if n == 0 {
  		return 1
	  }
	  return n * factorial(n-1)
  }
  ~~~
  - 同じことをfor文(loop)でやる方法
    ~~~go
    func main() {
	    n2 := loop1(4)
    	fmt.Println(n2) ---> 24

      n3 := loop2(4)
      fmt.Println(n3) ---> 24
    }

    func loop1(n int) int {
    	total := 1
    	for ; n > 0; n-- {
		    total *= n
    	}
	    return total
    }

    func loop2(n int) int {
    	x := n
    	for i := 1; i < n; i++ {
	    	x *= (n - i)
  	  }
	    return x
    }
    ~~~

## Pointer(ポインタ)
- > All values are stored in memory. Every location in memory has an address. A **pointer is a memory address**.
- 値が保管されたメモリのアドレスを指しているもの
- 変数(の値)が入るメモリのアドレスを保管するType
- ポインタ型変数には値を直接保管(代入)することはできない  
  → 値が保管されている変数のメモリアドレスを代入する
- ポインタ型変数は`var 変数名 *型`で定義
  - 例： `var x *string`
- メモリアドレスの指定には変数の前に`&`を指定（値が入っているメモリアドレスが知りたい場合、変数の前に`&`をつける）  
  - `&`は`ampersand(アンパサンド/엠퍼센드)`と読むらしい
- メモリアドレスに格納されている値を操作する場合はポインタ型変数の前に`*`を付ける
- `*&<変数>`で直接メモリにある値を指定することもできる
- `int`と`*int`は(stringとintのように)完全に違うタイプである
~~~go
var n int = 100
fmt.Println(&n)  → "0xc00007c008"等の変数nが格納されているメモリアドレスが表示される
var p *int = &n  → ポインタ型変数pに変数nが格納されているメモリアドレスを格納  
  →「p := &n」にすることもできる
fmt.Println(p)   → "0xc00007c008"等の変数nが格納されているメモリアドレスが表示される
fmt.Println(*p)  → メモリアドレス(p)に格納されている値 100 が表示される
*p = 300         → メモリアドレス(p)に格納されている値を100 → 300 に変更
fmt.Println(*p)  → メモリアドレス(p)に格納されている値 300 が表示される
x := 41
fmt.Println(*&x) → 41が表示される
~~~
~~~go
package main

import (
	"fmt"
)

type ar2x2 [2][2]int

func (a *ar2x2) Test() { ---> Structをpointerとして受け取る
	a[0] = [2]int{5, 6} 
	a[1] = [2]int{7, 8} --> この２行でメモリ上の値を直接書き換えているので、returnする必要がない
}

func main() {

	fmt.Println(ar2x2{}) ---------> zero値の"[[0 0] [0 0]]"が出力される
	a2 := ar2x2{{1, 2}, {3, 4}}
	fmt.Println("a2: ", a2) ------> 上で代入した"[1 2][3 4]"が出力
	a2.Test()
	fmt.Println("a2: ", a2) ------> Test methodで代入した"[5 6][7 8]"が出力
}
~~~

- **関数の引数がポインタ型のとき、その関数内では値の更新するために頭に`*`をつけなくて良い**  
  ```go
  type Foo struct {
      value int
  }

  func PassStruct(foo Foo) {
      foo.value = 1
  }

  func PassStructPointer(foo *Foo) {
      foo.value = 1 ------------------> *foo.value = 1ではない！
  }

  func main() {
      var foo Foo

      fmt.Printf("before PassStruct: %v\n", foo.value) -> 0 
      PassStruct(foo)
      fmt.Printf("after PassStruct: %v\n", foo.value) -> 0

      fmt.Printf("before PassStructPointer: %v\n", foo.value) -> 0
      PassStructPointer(&foo)
      fmt.Printf("after PassStructPointer: %v\n", foo.value) -> 1
  }
  ```

- **ポインタの利用シーン**
  1. big chunk of dataを受け渡ししたい場合
  2. 特定のメモリアドレスにある値を変更したい場合

### Method(メソッド)のReceiverがPointer型の場合
- **Methodを呼び出す時`&`を付けなくても、また、Method内でReceiverの値を参照/更新する時`*`を付けなくてもGoコンパイラが自動的に変換してくれる**
  - Methodの呼び出す例  
    例えば、以下のような構造体とメソッドがあるとします。
    ```go
    type MyStruct struct {
        Field int
    }

    func (s *MyStruct) SetField(value int) {
        s.Field = value
    }
    ```
    このメソッドを呼び出すには、以下のどちらの方法でも動作します。
    ```go
    s := MyStruct{}
    s.SetField(5) // 自動的に &s に変換されます。

    p := &MyStruct{}
    p.SetField(5) // すでにポインタなので変換は不要です。
    ```
    上記の `s.SetField(5)` では、`s` は値ですが、Goは自動的にポインタ `&s` に変換して、`SetField` メソッドを呼び出します。この機能により、ポインタを明示的に使用することなく、メソッドを簡単に呼び出すことができます。
  - Method内でReceiverの値を参照/更新する例  
    Goでは、Pointer Receiverを使用してMethodを定義すると、そのMethod内ではReceiverの実際の値に自動的にアクセスすることができます。したがって、Method内でReceiverのフィールドにアクセスする際には、特に`*`を付ける必要はありません。

    例えば、以下のコードの中:
    ```go
    func (shop *BarberShop) addBarber(barber string) {
        shop.NumberOfBarbers++
        // ...
        if len(shop.ClientsChan) == 0 {
            // ...
        }
    }
    ```
    ここで`shop`は`*BarberShop`型（`BarberShop`のポインタ）ですが、`shop.NumberOfBarbers`や`shop.ClientsChan`のように、直接そのフィールドにアクセスしています。この場合、Goは自動的にポインタをデリファレンス（参照している実際の値にアクセス）します。

    実際、以下の二つのコードスニペットは同等です：
    ```go
    func (shop *BarberShop) addBarber(barber string) {
        shop.NumberOfBarbers++
        // ...
    }
    ```
    と
    ```go
    func (shop *BarberShop) addBarber(barber string) {
        (*shop).NumberOfBarbers++
        // ...
    }
    ```
    しかし、通常は前者の方法が使用されます、なぜならそれはより簡潔で読みやすいからです。
- **これはMethod(メソッド)にのみ適用される話で、普通の関数では以下のように`&`を付けて関数を呼び出して値を連携し、`*`を付けてメモリアドレス内の値を変える必要がある**
  ```go
  func increment(x *int) {
      *x = *x + 1 // ポインタxの指す値にアクセスして、1を加える。
  }

  func main() {
      a := 5
      increment(&a)
      fmt.Println(a) // 出力: 6
  }
  ```

### 参照型（reference type）と値型（value type）について
- 参照型は関数に渡される際に参照（つまり、メモリ上のアドレス）が渡される。そのため、関数内でこれらの型の値を変更すると、元の値も変更される。
  - 関数から明示的に変更されたマップを(`return`で)返す必要はない
- 整数や文字列などの基本型は値型（value type）として扱われ、関数に渡す際にコピーが作成される。これらの型を関数内で変更しても、元の変数は変更されない。
```go
func modifyMap(m map[string]string) {
    m["new"] = "value"
}

func modifyString(s string) {
    s = "modified"  // この変更は呼び出し元には影響しない
}

func main() {
    myMap := make(map[string]string)
    myString := "original"

    modifyMap(myMap)
    modifyString(myString)

    fmt.Println(myMap)    // map[new:value] が出力される
    fmt.Println(myString) // "original" が出力される（変更されていない）
}
```
- 参照型（reference type）
  - マップ（Map）
  - スライス（Slice）
  - チャネル（Channel）
  - 関数（Function）
  - インターフェース（Interface）
  - ポインタ（Pointer）
- 値型（value type）
  - 整数型（int, int64など）
  - 浮動小数点型（float32, float64）
  - 論理型（bool）
  - 文字列型（string）
  - 配列型（固定長）

## Goroutine
- GoroutineはGoで実行できる一番小さい単位
  - main関数も1つのGoroutine
- Goroutineは並列処理を保証するのではなく、並列処理を実行できる環境の場合のみ並列処理をする
  - 例えばcpuコアが1つしかないコンピューターではGoroutineを使っても、並列(parallel)ではなく、並行(concurrent)処理になる
- 関数の前に`go`をつけるとGoroutineになる
  - 例：`go foo()`
- 1つのgoroutineは1つのスレッドとして動作する。そして、GoランタイムはこれらのgoroutineをOSスレッドにマッピングする（これにより、goroutineは非常に軽量となっている）。1つのgoroutine内で、関数やメソッドはシーケンシャルに（つまり1つずつ順番に）実行される。
- (Goroutineが複数ある場合)**GoRoutine処理は順番が保証されない**(毎回順番が異なる)。  
  **処理の順番はGo Schedulerによって決まる。**  
  例えば以下の例では"alpha"→"beta"→"delta"→"gamma"→・・・順ではなく、実行のたびに異なるRandom順で出力される
  ~~~go
  package main

  import (
	  "fmt"
	  "sync"
  )

  func printSomething(s string, wg *sync.WaitGroup) {
	  defer wg.Done() // decrement wg by one after this function completes

	  fmt.Println(s)
  }

  func main() {
	  // create a variable of type sync.WaitGroup
	  var wg sync.WaitGroup

	  // this slice consists of the words we want to print using a goroutine
  	words := []string{
	  	"alpha",
  		"beta",
		  "delta",
	  	"gamma",
  		"pi",
		  "zeta",
	  	"eta",
  		"theta",
		  "epsilon",
	  }

  	// we add the length of our slice (9) to the waitgroup
	  wg.Add(len(words))

  	for i, x := range words {
	  	// call printSomething as a goroutine, and hand it a pointer to our
		  // waitgroup, since you never want to copy a waitgroup after it has
		  // been created, or bad things happen...
		  go printSomething(fmt.Sprintf("%d: %s", i, x), &wg)
	  }

	  // our program will pause at this point, until wg is 0
	  wg.Wait()

	  // we have to add one to wg or we'll get an error when we call
	  // printSomething again, since wg is already at 0
	  wg.Add(1)
	  printSomething("This is the second thing to be printed!", &wg)
  }
  ~~~
- 複数のGoroutineが実行されている場合、あるGoroutineを抜ける(終了する)には`return`を使えば良い  
  ※`pizzeria`関数の中の`return`
  ~~~go
  package main

  import (
  	"fmt"
  	"math/rand"
  	"time"

  	"github.com/fatih/color"
  )

  const NumberOfPizzas = 10

  var pizzasMade, pizzasFailed, total int

  // Producer is a type for structs that holds two channels: one for pizzas, with all
  // information for a given pizza order including whether it was made
  // successfully, and another to handle end of processing (when we quit the channel)
  type Producer struct {
  	data chan PizzaOrder
  	quit chan chan error
  }

  // PizzaOrder is a type for structs that describes a given pizza order. It has the order
  // number, a message indicating what happened to the order, and a boolean
  // indicating if the order was successfully completed.
  type PizzaOrder struct {
  	pizzaNumber int
  	message     string
  	success     bool
  }

  // Close is simply a method of closing the channel when we are done with it (i.e.
  // something is pushed to the quit channel)
  func (p *Producer) Close() error {
  	ch := make(chan error)
  	p.quit <- ch
  	return <-ch
  }

  // makePizza attempts to make a pizza. We generate a random number from 1-12,
  // and put in two cases where we can't make the pizza in time. Otherwise,
  // we make the pizza without issue. To make things interesting, each pizza
  // will take a different length of time to produce (some pizzas are harder than others).
  func makePizza(pizzaNumber int) *PizzaOrder {
  	pizzaNumber++
  	if pizzaNumber <= NumberOfPizzas {
  		delay := rand.Intn(5) + 1
  		fmt.Printf("Received order #%d!\n", pizzaNumber)

  		rnd := rand.Intn(12) + 1
  		msg := ""
  		success := false

  		if rnd < 5 {
  			pizzasFailed++
  		} else {
  			pizzasMade++
  		}
  		total++

  		fmt.Printf("Making pizza #%d. It will take %d seconds....\n", pizzaNumber, delay)
  		// delay for a bit
  		time.Sleep(time.Duration(delay) * time.Second)

  		if rnd <=2 {
  			msg = fmt.Sprintf("*** We ran out of ingredients for pizza #%d!", pizzaNumber)
  		} else if rnd <= 4 {
  			msg = fmt.Sprintf("*** The cook quit while making pizza #%d!", pizzaNumber)
  		} else {
  			success = true
  			msg = fmt.Sprintf("Pizza order #%d is ready!", pizzaNumber)
  		}

  		p := PizzaOrder{
  			pizzaNumber: pizzaNumber,
  			message: msg,
  			success: success,
  		}

  		return &p

  	}

  	return &PizzaOrder{
  		pizzaNumber: pizzaNumber,
  	}
  }

  // pizzeria is a goroutine that runs in the background and
  // calls makePizza to try to make one order each time it iterates through
  // the for loop. It executes until it receives something on the quit
  // channel. The quit channel does not receive anything until the consumer
  // sends it (when the number of orders is greater than or equal to the
  // constant NumberOfPizzas).
  func pizzeria(pizzaMaker *Producer) {
  	// keep track of which pizza we are making
  	var i = 0

  	// this loop will continue to execute, trying to make pizzas,
  	// until the quit channel receives something.
  	for {
  		currentPizza := makePizza(i)
  		if currentPizza != nil {
  			i = currentPizza.pizzaNumber
  			select {
  			// we tried to make a pizza (we send something to the data channel -- a chan PizzaOrder)
  			case pizzaMaker.data <- *currentPizza:

  			// we want to quit, so send pizzMaker.quit to the quitChan (a chan error)
  			case quitChan := <-pizzaMaker.quit:
  				// close channels
  				close(pizzaMaker.data)
  				close(quitChan)
  				return
  			}
  		}
  	}
  }

  func main() {
  	// seed the random number generator
  	rand.Seed(time.Now().UnixNano())

  	// print out a message
  	color.Cyan("The Pizzeria is open for business!")
  	color.Cyan("----------------------------------")

  	// create a producer
  	pizzaJob := &Producer{
  		data: make(chan PizzaOrder),
  		quit: make(chan chan error),
  	}

  	// run the producer in the background
  	go pizzeria(pizzaJob)

  	// create and run consumer
  	for i := range pizzaJob.data {
  		if i.pizzaNumber <= NumberOfPizzas {
  			if i.success {
  				color.Green(i.message)
  				color.Green("Order #%d is out for delivery!", i.pizzaNumber)
  			} else {
  				color.Red(i.message)
  				color.Red("The customer is really mad!")
  			}
  		} else {
  			color.Cyan("Done making pizzas...")
  			err := pizzaJob.Close()
  			if err != nil {
  				color.Red("*** Error closing channel!", err)
  			}
  		}
  	}

  	// print out the ending message
  	color.Cyan("-----------------")
  	color.Cyan("Done for the day.")

  	color.Cyan("We made %d pizzas, but failed to make %d, with %d attempts in total.", pizzasMade, pizzasFailed, total)

  	switch {
  	case pizzasFailed > 9:
  		color.Red("It was an awful day...")
  	case pizzasFailed >= 6:
  		color.Red("It was not a very good day...")
  	case pizzasFailed >= 4:
  		color.Yellow("It was an okay day....")
  	case pizzasFailed >= 2:
  		color.Yellow("It was a pretty good day!")
  	default:
  		color.Green("It was a great day!")
  	}
  }
  ~~~

### 実行する最大Goroutine数の制御
- `semaphore`パッケージを使って実行可能な最大Goroutine数を制御できる
- `golang.org/x/sync/semaphore`をimportして使う
- https://github.com/golang/sync/blob/master/semaphore/semaphore.go
- https://github.com/golang/sync/tree/master
- サンプルコード
  ~~~go
  package main

  import (
  	"context"
  	"fmt"
  	"math/rand"
  	"strconv"
  	"sync"
  	"time"

  	"github.com/aws/aws-sdk-go/aws"
  	"github.com/aws/aws-sdk-go/aws/session"
  	"github.com/aws/aws-sdk-go/service/sqs"
  	"golang.org/x/sync/semaphore"
  )

  const (
  	QueueURL                = "https://sqs.ap-northeast-1.amazonaws.com/1234567890/test.fifo"
  	MaxConcurrentGoroutines = 2 // 同時に実行されるgoroutineの最大数
  )

  var wg sync.WaitGroup
  var sem = semaphore.NewWeighted(MaxConcurrentGoroutines) // セマフォを初期化

  func sendmsg(i int, svc *sqs.SQS) {
  	defer wg.Done()
  	// セマフォを取得
  	if err := sem.Acquire(context.Background(), 1); err != nil { // 指定した同時実行数制限semから1つ実行権限を取得。上限に達していて取得できない場合は、取得でき次第、実行を開始
  		fmt.Println("Failed to acquire semaphore:", err)
  		return
  	}
  	defer sem.Release(1) // goroutineが完了したらリリース

  	n := rand.Intn(100000000000000)

  	MessageDedupId := strconv.Itoa(n)
  	messageBody := "Hello, SQS from Go! " + strconv.Itoa(i)

  	// メッセージの送信
  	sendMsgInput := &sqs.SendMessageInput{
  		MessageBody: aws.String(messageBody),
  		QueueUrl:    aws.String(QueueURL),
  		// FIFOキューを使用している場合、MessageGroupIdが必要
  		MessageGroupId: aws.String("SQS_TEST"), // 通常、同じ処理を行うメッセージに共通の値を設定
  		// MessageDeduplicationIdはオプション
  		MessageDeduplicationId: aws.String(MessageDedupId),
  		// MessageDeduplicationId は、可能な限りユニークな値を提供することが重要（ただし、必須ではない）。
  		// これは、同一の MessageDeduplicationId を持つメッセージが重複排除期間内に複数回送信された場合、後続のメッセージが受け入れられないことを意味する。(MessageDeduplicationIdが重複するメッセージはreceive側で受信しない)
  	}
  	_, err := svc.SendMessage(sendMsgInput)
  	if err != nil {
  		fmt.Println("Error:", err)
  		return
  	}
  	fmt.Println("Message sent:  Hello, SQS from Go!", i)
  }

  func main() {
  	sess := session.Must(session.NewSession(&aws.Config{
  		Region: aws.String("ap-northeast-1"),
  	}))

  	svc := sqs.New(sess)

  	// 乱数生成器を初期化。これは一度だけ実行する必要がある。
  	rand.Seed(time.Now().UnixNano())

  	count := 10

  	for i := 1; i <= count; i++ {
  		wg.Add(1)
  		go sendmsg(i, svc)
  	}
  	wg.Wait()
  }
  ~~~

## WaitGroup
- Goroutineで実行した処理はデフォルトでは待ってもらえず、main関数が終了すればGoroutine処理が終わってなくてもプログラムは終了してしまう
- ProcessがkillされるとすべてのGoroutineもcancelされる
  - Goroutineのleakを防ぐため（Goroutineはmemoryなどリソースを消費する）
- Goroutine処理が終わるまで待ってもらうためのものが`WaitGroup`
- **`sync.WaitGroup.Wait()`は他のすべてのGoroutineが完了するまで待つ**
- 例えば以下のコードではfooは出力されず、barだけ出力されて終了する
    ~~~go
    package main

    import (
      "fmt"
    )

    func main() {
      go foo()
      bar()
    }

    func foo() {
      for i := 0; i < 10; i++ {
        fmt.Println("foo:", i)
      }
    }

    func bar() {
      for i := 0; i < 10; i++ {
        fmt.Println("bar:", i)
      }
    }    
    ~~~
  - WaitGroupで以下を定義して明示的にGoroutine処理完了を待ってもらう
    - `sync.WaitGroup.Add(待たすGoroutine数)`
    - `sync.WaitGroup.Wait()`
    - `sync.WaitGroup.Done()`
      ~~~go
      package main

      import (
        "fmt"
        "sync"
      )

      var wg sync.WaitGroup

      func main() {
        wg.Add(1)

        go foo()
        bar()
        wg.Wait()
      }

      func foo() {
        for i := 0; i < 10; i++ {
          fmt.Println("foo:", i)
        }
        wg.Done()
      }

      func bar() {
        for i := 0; i < 10; i++ {
          fmt.Println("bar:", i)
        }
      }
      ~~~
  - **`Add()`内の待機するGoroutine数に関係なく、`Add()`と`Wait()`は１対１である必要がある。**  
    - NG例  
      `fatal error: all goroutines are asleep - deadlock!`エラーが出る
      ~~~go
      package main

      import (
          "fmt"
          "sync"
      )

      var msg string

      func updateMessage(s string, wg *sync.WaitGroup) {
          msg = s
          defer wg.Done()
      }

      func printMessage() {
          fmt.Println(msg)
      }

      func main() {

          msg = "Hello, world!"
          var wg sync.WaitGroup

          wg.Add(3)
          go updateMessage("Hello, universe!", &wg)
          wg.Wait()
          printMessage()

          go updateMessage("Hello, cosmos!", &wg)
          wg.Wait()
          printMessage()

          go updateMessage("Hello, world!", &wg)
          wg.Wait()
          printMessage()
      }          
      ~~~
    - OK例（1）
      ~~~go
      package main

      import (
          "fmt"
          "sync"
      )

      var msg string
      var wg sync.WaitGroup

      func updateMessage(s string) {
          defer wg.Done()
          msg = s
          fmt.Println(msg)
      }

      func main() {
          msg = "Hello, world!"

          wg.Add(2)
          go updateMessage("Hello, universe!")
          go updateMessage("Hello, cosmos!")
          wg.Wait()
      }      
      ~~~
    - OK例（2）
      ~~~go
      package main

      import (
          "fmt"
          "sync"
      )

      var msg string

      func updateMessage(s string, wg *sync.WaitGroup) {
          msg = s
          defer wg.Done()
      }

      func printMessage() {
          fmt.Println(msg)
      }

      func main() {

          msg = "Hello, world!"
          var wg sync.WaitGroup

          wg.Add(1)
          go updateMessage("Hello, universe!", &wg)
          wg.Wait()
          printMessage()

          wg.Add(1)
          go updateMessage("Hello, cosmos!", &wg)
          wg.Wait()
          printMessage()

          wg.Add(1)
          go updateMessage("Hello, world!", &wg)
          wg.Wait()
          printMessage()
      }      
      ~~~
- 処理ごとに複数のWaitGroupを作成することも可能
  - 以下の例ではphilosophers数の分 *"xx is seated at the table."* が出力されるまで`for i := hunger; i > 0; i-- {}`の部分には進まない
    ~~~go
    package main

    import (
        "fmt"
        "sync"
        "time"
    )

    // The Dining Philosophers problem is well known in computer science circles.
    // Five philosophers, numbered from 0 through 4, live in a house where the
    // table is laid for them; each philosopher has their own place at the table.
    // Their only difficulty – besides those of philosophy – is that the dish
    // served is a very difficult kind of spaghetti which has to be eaten with
    // two forks. There are two forks next to each plate, so that presents no
    // difficulty. As a consequence, however, this means that no two neighbours
    // may be eating simultaneously, since there are five philosophers and five forks.
    //
    // This is a simple implementation of Dijkstra's solution to the "Dining
    // Philosophers" dilemma.

    // Philosopher is a struct which stores information about a philosopher.
    type Philosopher struct {
        name      string
        rightFork int
        leftFork  int
    }

    // philosophers is list of all philosophers.
    var philosophers = []Philosopher{
        {name: "Plato", leftFork: 4, rightFork: 0},
        {name: "Socrates", leftFork: 0, rightFork: 1},
        {name: "Aristotle", leftFork: 1, rightFork: 2},
        {name: "Pascal", leftFork: 2, rightFork: 3},
        {name: "Locke", leftFork: 3, rightFork: 4},
    }

    // Define a few variables.
    var hunger = 3                  // how many times a philosopher eats
    var eatTime = 1 * time.Second   // how long it takes to eatTime
    var thinkTime = 3 * time.Second // how long a philosopher thinks
    var sleepTime = 1 * time.Second // how long to wait when printing things out

    func main() {
        // print out a welcome message
        fmt.Println("Dining Philosophers Problem")
        fmt.Println("---------------------------")
        fmt.Println("The table is empty.")

        // start the meal
        dine()

        // print out finished message
        fmt.Println("The table is empty.")

    }

    func dine() {
        eatTime = 0 * time.Second
        sleepTime = 0 * time.Second
        thinkTime = 0 * time.Second

        // wg is the WaitGroup that keeps track of how many philosophers are still at the table. When
        // it reaches zero, everyone is finished eating and has left. We add 5 (the number of philosophers) to this
        // wait group.
        wg := &sync.WaitGroup{}
        wg.Add(len(philosophers))

        // We want everyone to be seated before they start eating, so create a WaitGroup for that, and set it to 5.
        seated := &sync.WaitGroup{}
        seated.Add(len(philosophers))

        // forks is a map of all 5 forks. Forks are assigned using the fields leftFork and rightFork in the Philosopher
        // type. Each fork, then, can be found using the index (an integer), and each fork has a unique mutex.
        var forks = make(map[int]*sync.Mutex)
        for i := 0; i < len(philosophers); i++ {
            forks[i] = &sync.Mutex{}
        }

        // Start the meal by iterating through our slice of Philosophers.
        for i := 0; i < len(philosophers); i++ {
            // fire off a goroutine for the current philosopher
            go diningProblem(philosophers[i], wg, forks, seated)
        }

        // Wait for the philosophers to finish. This blocks until the wait group is 0.
        wg.Wait()
    }

    // diningProblem is the function fired off as a goroutine for each of our philosophers. It takes one
    // philosopher, our WaitGroup to determine when everyone is done, a map containing the mutexes for every
    // fork on the table, and a WaitGroup used to pause execution of every instance of this goroutine
    // until everyone is seated at the table.
    func diningProblem(philosopher Philosopher, wg *sync.WaitGroup, forks map[int]*sync.Mutex, seated *sync.WaitGroup) {
        defer wg.Done()

        // seat the philosopher at the table
        fmt.Printf("%s is seated at the table.\n", philosopher.name)
        
        // Decrement the seated WaitGroup by one.
        seated.Done()

        // Wait until everyone is seated.
        seated.Wait()

        // Have this philosopher eatTime and thinkTime "hunger" times (3).
        for i := hunger; i > 0; i-- {
            // Get a lock on the left and right forks. We have to choose the lower numbered fork first in order
            // to avoid a logical race condition, which is not detected by the -race flag in tests; if we don't do this,
            // we have the potential for a deadlock, since two philosophers will wait endlessly for the same fork.
            // Note that the goroutine will block (pause) until it gets a lock on both the right and left forks.
            if philosopher.leftFork > philosopher.rightFork {
                forks[philosopher.rightFork].Lock()
                fmt.Printf("\t%s takes the right fork.\n", philosopher.name)
                forks[philosopher.leftFork].Lock()
                fmt.Printf("\t%s takes the left fork.\n", philosopher.name)
            } else {
                forks[philosopher.leftFork].Lock()
                fmt.Printf("\t%s takes the left fork.\n", philosopher.name)
                forks[philosopher.rightFork].Lock()
                fmt.Printf("\t%s takes the right fork.\n", philosopher.name)
            }
            
            // By the time we get to this line, the philosopher has a lock (mutex) on both forks.
            fmt.Printf("\t%s has both forks and is eating.\n", philosopher.name)
            time.Sleep(eatTime)

            // The philosopher starts to think, but does not drop the forks yet.
            fmt.Printf("\t%s is thinking.\n", philosopher.name)
            time.Sleep(thinkTime)

            // Unlock the mutexes for both forks.
            forks[philosopher.leftFork].Unlock()
            forks[philosopher.rightFork].Unlock()

            fmt.Printf("\t%s put down the forks.\n", philosopher.name)
        }

        // The philosopher has finished eating, so print out a message.
        fmt.Println(philosopher.name, "is satisified.")
        fmt.Println(philosopher.name, "left the table.")
    }
    ~~~

#### ■ WaitGroup使用上注意点
- **`sync.WaitGroup.Wait()`は他のすべてのGoroutineが完了するまで待つので、`sync.WaitGroup.Wait()`は独自のGoroutineで定義しないといけない。**  
  例えば以下のコードで`wg.Wait()`をgoroutineにしないと`wg.Wait()`は自身が実行されているmain goroutineの終了も待つことになるので、いつまで経っても`wg.Wait()`以降が実行されない。
  ```go
  package main

  import (
  	"fmt"
  	"math/rand"
  	"strconv"
  	"sync"
  	"time"

  	"github.com/aws/aws-sdk-go/aws"
  	"github.com/aws/aws-sdk-go/aws/session"
  	"github.com/aws/aws-sdk-go/service/sqs"
  )

  const (
  	QueueURL = "https://sqs.ap-northeast-1.amazonaws.com/1234567890/test.fifo"
  )

  func sendmsg(i int, c chan int, svc *sqs.SQS, wg *sync.WaitGroup) {
  	rand.Seed(time.Now().UnixNano())

  	n := rand.Intn(1000000000000)

  	MessageDedupId := strconv.Itoa(n)
  	messageBody := "Hello, SQS from Go! " + strconv.Itoa(i)

  	sendMsgInput := &sqs.SendMessageInput{
  		MessageBody: aws.String(messageBody),
  		QueueUrl:    aws.String(QueueURL),
  		MessageGroupId: aws.String("SQS_TEST"),
  		MessageDeduplicationId: aws.String(MessageDedupId),
  	}
  	_, err := svc.SendMessage(sendMsgInput)
  	if err != nil {
  		fmt.Println("Error:", err)
  		return
  	}
  	c <- i
  	wg.Done()
  }

  func main() {
  	c := make(chan int)

  	sess := session.Must(session.NewSession(&aws.Config{
  		Region: aws.String("ap-northeast-1"),
  	}))
  	svc := sqs.New(sess)

  	var wg sync.WaitGroup

  	for i := 1; i <= 10; i++ {
  		wg.Add(1)
  		go sendmsg(i, c, svc, &wg)
  	}

  	go func() {
  		wg.Wait()
  		close(c)
  	}()

  	for i := range c {
  		fmt.Println("Message sent: ", i)
  	}
  }
  ```
- **上記で言ったように`sync.WaitGroup.Wait()`は他のすべてのGoroutineが完了するまで待つので、main関数から派生されるgoroutineの中でさらにgoroutineを生成する場合、main関数から派生されたgoroutineがある関数では`sync.WaitGroup.Wait()`を定義してはいけない**
  ~~~go
  func main() {
  	// DB接続
  	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
  	if err != nil {
  		fmt.Println("failed to connect database")
  		panic(err)
  	}

  	db_users := []string{"dbuser1", "dbuser2", "dbuser4", "dbuser5", "dbuser7"}
  	wg.Add(2)
  	go DbuserExistCheck(DB, db_users...)

  	iam_users := []string{"iamuser1", "iamuser2", "iamuser3", "iamuser4", "iamuser5"}
  	go AwsIamUserExistCheck(iam_users...)
  	wg.Wait() // wait until all goroutines are finished (including goroutines that are not created in this function)

  	fmt.Println("no_exist_db_user:", no_exist_db_user)
  	fmt.Println("no_exist_iam_user:", no_exist_iam_user)
  }

  func DbuserExistCheck(db *gorm.DB, db_user ...string) {
  	defer wg.Done()

  	// セッションを利用するために、DBに接続する
  	sqldb, err := db.DB()
  	if err != nil {
  		fmt.Println("failed to connect DB")
  		panic(err)
  	}
  	// DBに対する処理が終わったら、DB接続を解除する
  	defer sqldb.Close()

  	for _, v := range db_user {
  		db.Where("dbuser = ?", v).Find(&Dbuserpassword{}).Count(&no_db_user_count)
  		if no_db_user_count == 0 {
  			no_exist_db_user = append(no_exist_db_user, v)
  			fmt.Printf("%v is not exist\n", v)
  		} else {
  			continue
  		}
  	}
  }

  func AwsIamUserExistCheck(iam_user ...string) {
  	defer wg.Done()

  	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))

  	if err != nil {
  		fmt.Println("failed to load config")
  		panic(err)
  	}

  	svc := iam.NewFromConfig(cfg)

  	// IAMユーザの存在チェック
  	wg.Add(len(iam_user))
  	for _, v := range iam_user {
  		go func(v string) {
  			defer wg.Done()

  			input := &iam.GetUserInput{
  				UserName: aws.String(v),
  			}

  			_, err := svc.GetUser(context.TODO(), input)
  			if err != nil {
  				var nsk *types.NoSuchEntityException
  				if errors.As(err, &nsk) {
  					fmt.Println("NoSuchEntityException")
  					mu.Lock()
  					no_exist_iam_user = append(no_exist_iam_user, v)
  					mu.Unlock()
  				}
  				var apiErr smithy.APIError
  				if errors.As(err, &apiErr) {
  					fmt.Println("StatusCode:", apiErr.ErrorCode(), ", Msg:", apiErr.ErrorMessage())
  				}
  			}
  		}(v)
  	}
  	// wg.Wait()
  	// wg.Wait()はすべてのgoroutineが完了するまで待つため、ここにwg.Wait()があるとmain関数内のgoroutineが完了するまで待つことになり、
  	// main関数内のwg.Wait()とここにあるwg.Wait()の2つのwg.Wait()がお互いを待ち合うことになり、デッドロックが発生する。
  }
  ~~~

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
- **Race Condition**とは  
  → 競合状態
  > A race condition or race hazard is an undesirable condition of an electronics, software, or other system where the system's substantive behavior is dependent on the sequence or timing of other uncontrollable events. It becomes a bug when one or more of the possible behaviors is undesirable.
  > A race condition occurs when two Goroutine access a shared variable at the same time. 
  - **Race Conditionは２つ以上のgo routineが同じもの(e.g. variable,struct,・・・)に対して更新処理を行う時に発生する。参照のみの時は発生しない**
- 参考URL
  - https://pkg.go.dev/sync#Mutex
  - https://learn.microsoft.com/en-us/troubleshoot/developer/visualstudio/visual-basic/language-compilers/race-conditions-deadlocks
  - https://stackoverflow.com/questions/34510/what-is-a-race-condition

## Channels
- Channels are the pipes that connect concurrent goroutines. You can send values into channels from one goroutine and receive those values into another goroutine.
- ChannelsはGoroutine間でデータを共有(送受信)する方法/仕組み
- Dataを送受信できる空間(Pipe)
  - ブロッキング付きのキューみたいなもの
- Channelは使った後に必ず`close(<channel名>)`で閉じなければならない  
  → closeしないとResource Leakが発生する恐れがある
  > Once you're done with a channel, you must close it !
- **閉じたChannelから値を取り出すと、そのChannelの型のZero値が取得される**
- Channelにはchannelから受け取ることができるoptionalな2つ目のparameter(`boolean`typeで通常受け取る変数名は`ok`とする)がある  
  **Channelがcloseされてemptyの場合は`False`が、別のGoroutineから送られた値をChannelから受け取った場合は`True`が入る**
  - **https://stackoverflow.com/questions/10437015/does-a-channel-return-two-values**
    > The boolean variable ok returned by a [receive operator](https://go.dev/ref/spec#Receive_operator) indicates whether the received value was sent on the channel (true) or is a zero value returned because the channel is closed and empty (false).
    > 
    > The for loop terminates when some other part of the Go program closes the fromServer or the fromUser channel. In that case one of the case statements will set ok to true. So if the user closes the connection or the remote server closes the connection, the program will terminate.
    - what does empty mean?
      > If a channel is closed but it is still containing some items it is possible to receive from it and ok is true. But it is impossible to write to a closed channel (this is a definition of "closed channel" in fact). When a channel had been closed by producer goroutine and drained by consumer than ok is false. Empty and closed, just like it said.
    ~~~go
    for self.isRunning {

      select {
      case serverData, ok = <-fromServer:   // What's the meaning of the second value(ok)?
          if ok {
              self.onServer(serverData)
          } else {
              self.isRunning = false
          }

      case userInput, ok = <-fromUser:
          if ok {
              self.onUser(userInput)
          } else {
              self.isRunning = false
          }
      }
    }
    ~~~
  - https://go.dev/ref/spec#Receive_operator
- Channelは値が入るまで待つ(後続処理を実行しない)ので、WaitGroupが不要
  ~~~go
  package main

  import "fmt"

  func goroutine1(s []int, c chan int) {
      sum := 0
      for _, v := range s {
          sum += v
      }
      c <- sum
  }

  func main() {
      s := []int{1, 2, 3, 4, 5}
      c := make(chan int)

      go goroutine1(s, c)
      x := <-c ==============> ここでPrintlnに進まずにxに値が入るまで待つ
      fmt.Println(x)
  }
  ~~~
- 1つのChannelを複数のGoroutineで共有することもできる
  ~~~go
    package main

    import "fmt"

    func goroutine1(s []int, c chan int) {
        sum := 0
        for _, v := range s {
            sum += v
        }
        c <- sum
    }

    func goroutine2(t []int, c chan int) {
        sum := 0
        for _, v := range t {
            sum += v
        }
        c <- sum
    }

    func main() {
        s := []int{1, 2, 3, 4, 5}
        t := []int{6, 7, 8, 9, 10}

        c := make(chan int)

        go goroutine1(s, c)
        go goroutine2(t, c)

        x := <-c
        fmt.Println(x)

        y := <-c
        fmt.Println(y)
    }
  ~~~
- ChannelもType
- **Channelは値を入れた後に遮断(Blocking)される**
  - **BlockingされたChannelに値を入れようとするとpanicを起こす**
  - なので**goroutine**で**channelに値を入れるのとchannelから値を取り出すのを並行で**実施するようにコードを書く必要がある
  - **unbuffered Channelに値を入れた後、取り出さずにまた値を入れようとすると`fatal error: all goroutines are asleep - deadlock!`エラーが発生する**
  - または**buffer**を使ってchannelに値が残れるようにする
    - 定義したbufferの数より多くの数の値をchannelに入れようとするとエラーになる  
      → 定義した数の分がbufferに入ってきたらchannelは遮断される
- `make(chan <Channelに入るデータの型>)`でChannelを初期化(作成)する  
  e.g. `make(chan int)` → int型データを入れるchannel
  - buffer channelを作る場合は`make(chan <Channelに入るデータの型>, <buffer数>)`
- **unbuffered Channelは１つのgoroutineの中では使えないけど、buffered Channelは１つのgoroutineの中で使える**
- OK例（goroutineでchannelへの格納とchannelからの取り出しを同時にする例）
  ~~~go
  func main() {
	  c := make(chan int)
      go func() {
        c <- 42
      }()
	  fmt.Println(<-c)
  }
  ~~~
- OK例（bufferを使う例）
  ~~~go
  func main() {
	  c := make(chan int, 1)
      c <- 42
	  fmt.Println(<-c)
  }
  ~~~
- OK例（bufferを使う例２）
  ~~~go
  func main() {
	  c := make(chan int, 2)
      c <- 42
      c <- 43
	  fmt.Println(<-c) ---> 42と出力される
	  fmt.Println(<-c) ---> 43と出力される
  }
  ~~~

- NG例  
  → `all goroutines are asleep - deadlock!`とエラーになる  
    - 以下Chat-GPTからの回答
      > Goのチャネルはデフォルトで同期的（unbuffered）です。つまり、あるゴルーチンがチャネルにデータを送信すると、そのデータが別のゴルーチンによって受信されるまで送信ゴルーチンはブロックされます。同様に、ゴルーチンがチャネルからデータを受信しようとすると、別のゴルーチンがデータを送信するまで受信ゴルーチンはブロックされます。
      > 
      > このコードの問題は、c <- 42によってデータを送信しようとするゴルーチンがありますが、その時点でデータを受信しようとする別のゴルーチンがないため、送信ゴルーチンが永遠にブロックされる点にあります。
  ~~~go
  func main() {
	  c := make(chan int)
      c <- 42
	  fmt.Println(<-c)
  }
  ~~~
- NG例（定義したbuffer数より多くの値を入れる）  
  → `all goroutines are asleep - deadlock!`とエラーになる
  ~~~go
  func main() {
	  c := make(chan int, 1)
      c <- 42
      c <- 43
	  fmt.Println(<-c)
  }
  ~~~
  - ただ、Channelから値を取り出してから入れればエラーにならない
    ~~~go
    package main

    import "fmt"

    func main() {
        c := make(chan int, 1)
        c <- 42
        fmt.Println(len(c)) ==> 1
        x := <-c

        fmt.Println(x) =======> 42
        fmt.Println(len(c)) ==> 0
        c <- 77
        fmt.Println(len(c)) ==> 1
        fmt.Println(<-c) =====> 77
        fmt.Println(len(c)) ==> 0
    }    
    ~~~

  #### 単方向(受信だけ or 送信だけ)のChannelも作成できる
  - 例
    ~~~go
    func main() {
    	c := make(chan int)
    	cr := make(<-chan int) // receive (Channelから値を取り出す)
    	cs := make(chan<- int) // send (Channelに値を入れる)

    	fmt.Println("-----")
    	fmt.Printf("%T\n", c) ------> "chan int"と出力される
    	fmt.Printf("%T\n", cr) -----> "<-chan int"と出力される
    	fmt.Printf("%T\n", cs) -----> "chan<- int"と出力される
    }    
    ~~~
  - NG例（送信用Channelに対して受信しようとした場合）  
    → `invalid operation: cannot receive from send-only channel cs (variable of type chan<- int)`エラーが出る
    ~~~go
    func main() {
      cs := make(chan<- int)

      go func() {
        cs <- 42
      }()
      fmt.Println(<-cs) ---> ここがNG(取り出そうとしている)

      fmt.Printf("------\n")
      fmt.Printf("cs\t%T\n", cs)
    }
    ~~~
  - NG例（受信用Channelに対して送信しようとした場合）  
    → `invalid operation: cannot send to receive-only channel cr (variable of type <-chan int)`エラーが出る
    ~~~go
    func main() {
      cr := make(<-chan int)

      go func() {
        cr <- 42 -------> ここがNG(値を入れようとしている)
      }()
      fmt.Println(<-cr)

      fmt.Printf("------\n")
      fmt.Printf("cr\t%T\n", cr)
    }
    ~~~

  #### Channelの`for`&`range`によるLoopと`close`について
  - Goroutineの中で1つのChannelに複数の値を入れる場合、Channelから受け取る処理も複数行う必要がある。  
    例えば以下のような例では`fmt.Println(x)`で最初にChannelに入れた値"1"しか出力されない
    ~~~go
    package main

    import "fmt"

    func goroutine1(s []int, c chan int) {
        sum := 0
        for _, v := range s {
            sum += v
            c <- sum
        }
    }

    func main() {
        s := []int{1, 2, 3, 4, 5}
        c := make(chan int)

        go goroutine1(s, c)
        x := <-c
        fmt.Println(x) ==> 1
    }
    ~~~
  - Channel内のすべての値を引き出すためにfor文を使わずにやる場合は値の数の分処理が増える
    ~~~go
      　　・
      　　・
      　　・
    
    func main() {
        s := []int{1, 2, 3, 4, 5}
        c := make(chan int)

        go goroutine1(s, c)
        x := <-c
        x2 := <-c
        x3 := <-c
        x4 := <-c
        x5 := <-c
        fmt.Println(x) ===> 1
        fmt.Println(x2) ==> 3
        fmt.Println(x3) ==> 6
        fmt.Println(x4) ==> 10
        fmt.Println(x5) ==> 15
    }
    ~~~
  - そこでfor文を使ってChannel内の値の数の分、処理を回すことができる
    > **Note**  
    > これでうまくいくように見えるがエラーになる
    ~~~go
    package main

    import "fmt"

    func goroutine1(s []int, c chan int) {
        sum := 0
        for _, v := range s {
            sum += v
            c <- sum
        }
    }

    func main() {
        s := []int{1, 2, 3, 4, 5}
        c := make(chan int)

        go goroutine1(s, c)

        for ch := range c {
            fmt.Println(ch)
        }
    }
    ~~~
  - 上記では`goroutine1`関数は5までchannelに値を入れた後完了するが、main関数内のfor文はChannelから最後(5番目)の値を取り出した後もChannelがcloseされてないため、channelに新しい値が入ってくることを待っている。しかし、`goroutine1`関数は終了しており、新しい値が入ってくることはないため、main関数内のfor文は永遠に待ち続ける。これはDeadlockの状態と言えるので以下のように`all goroutines are asleep - deadlock!`エラーが出る  
    ~~~
    1
    3
    6
    10
    15
    fatal error: all goroutines are asleep - deadlock!

    goroutine 1 [chan receive]:
    main.main()
        /opt/go/concurrency/test2.go:20 +0x125
    exit status 2
    ~~~
    これを防ぐためにChannelにすべての値を入れた後に明示的に`close(<Channel名>)`でChannelをCloseする必要がある
    ~~~go
    package main

    import "fmt"

    func goroutine1(s []int, c chan int) {
        sum := 0
        for _, v := range s {
            sum += v
            c <- sum
        }
        close(c) ========> ここ！
    }

    func main() {
        s := []int{1, 2, 3, 4, 5}
        c := make(chan int)

        go goroutine1(s, c)

        for ch := range c {
            fmt.Println(ch)
        }
    }
    ~~~
  - 以下Chat-GPTからの回答
    > for文は、channelから値が利用可能になるまで待機します。この期間、forループはブロックされ、新しい値がchannelに送信されるまで進行しません。
    > channelがcloseされると、forループは終了します。closeされたchannelからの読み取りは常に可能で、それ以降の読み取りではゼロ値（型に応じたゼロ値）が返されます。
  - **for文はchannelから値が利用可能になるまで待ち続けるため、channelをcloseしない場合、forより下にあるコードは実行されない**

  #### `chan error`について
  - `error`型のzero valueは`nil`  
    なのでcloseされた`error`型のchannelから値を取り出そうとすると`nil`が返ってくる  
    ~~~go
    func main() {
    	ch := make(chan error)
    	close(ch)

    	val, ok := <-ch
    	fmt.Println(val, ok) // 出力: <nil> false
    }
    ~~~

  #### `time.After()`を使ったtimeout設定
  - `time.After(d)`は指定した期間`d`が経過すると、現在の時刻を送信する新しいchannelを返す。
  - `select`と組み合わせて使うことで複数のchannel操作を同時に待機する際に便利  
    ~~~go
    select {
    case <-time.After(1 * time.Second):
        fmt.Println("Timed out")
    case msg := <-ch:
        fmt.Println("Received message:", msg)
    }
    ~~~

## select
- selectはChannelでしか使えない。文法は`switch`とほぼ一緒。
- *Select statements pull the value from whatever channel has a value ready to be pulled.*
- channelは通常値が入っていなければ受信をブロックするが、select文はブロックしないで処理する時に利用
- selectを使うと複数のChannelからの受信を待てる
- 例（"received one","received two"）
  ~~~go
  package main

  import (
      "fmt"
      "time"
  )

  func main() {
      c1 := make(chan string)
      c2 := make(chan string)

      go func() {
          time.Sleep(1 * time.Second)
          c1 <- "one"
      }()
      go func() {
          time.Sleep(2 * time.Second)
          c2 <- "two"
      }()

      for i := 0; i < 2; i++ {
          select {
          case msg1 := <-c1:
              fmt.Println("received", msg1)
          case msg2 := <-c2:
              fmt.Println("received", msg2)
          }
      }
  }
- **1つのChannelに対して複数のcaseで待ち受ける時、ランダムでcaseが選ばれる**  
  例えば、以下例の場合、`channel1`に値を送信したら`case s1`か`case s2`のどっちかがランダムに選ばれる
  ~~~go
	for {
		select {
		// because we have multiple cases listening to
		// the same channels, random ones are selected
		case s1 := <-channel1:
			fmt.Println("Case one:", s1)
		case s2 := <-channel1:
			fmt.Println("Case two:", s2)
		case s3 := <-channel2:
			fmt.Println("Case three:", s3)
		case s4 := <-channel2:
			fmt.Println("Case four:", s4)
			// default:
			// avoiding deadlock
		}
	}
  ~~~
- `select`文内の`case`ブロックの中に`return`がある(実行される)と、対象`case`ブロックが指定したchannelから値を受信or送信した後に関数全体が終了する。    
  - 以下の例の場合、１か２が出力された後に *"This line will not executed."* は出力されずmain関数が終了する  
    ※**goroutineの実行順序とselectが複数のchannelのうちどこから使うかはランダム(Go runtimeによって決まる)**  
    ~~~go
    package main

    import "fmt"

    func main() {
    	ch1 := make(chan int)
    	ch2 := make(chan int)

    	go func() {
    		ch1 <- 1
    	}()

    	go func() {
    		ch2 <- 2
    	}()

    	select {
    	case i := <-ch1:
    		fmt.Println(i)
    		return
    	case i := <-ch2:
    		fmt.Println(i)
    		return
    	}

    	fmt.Println("This line will not be executed.")
    }
    ~~~
- `case`ブロックに`return`がないと`select`はchannelからの送信/受信を待ち続けるが、`time.After()`を使えば特定時間後`select`を終了させることができる。  
  以下の例では *"Timeout"* だけ出力されてプログラムが終了する。
  ~~~go
  package main

  import (
  	"fmt"
  	"time"
  )

  func main() {
  	c := make(chan int)
  	
  	go func() {
  		time.Sleep(2 * time.Second)
  		c <- 42
  	}()
  	
  	select {
  	case val := <-c:
  		fmt.Println("Received:", val)
  		return
  	case <-time.After(1 * time.Second):
  		fmt.Println("Timeout")
  		return
  	}
  }
  ~~~
- `for`と組合せて使うことで(明示的に終了させるまでは)プログラムが終了せず、Channelに値が入ることを待つことができる
- 参考URL
  - https://www.spinute.org/go-by-example/select.html
  - https://go-tour-jp.appspot.com/concurrency/6
  - https://www.ardanlabs.com/blog/2013/11/label-breaks-in-go.html

## パッケージ(import)
- Format
  1. 1つずつ個別にimport
      ~~~go
      import "fmt"
      import "os"
      import "time"
      ~~~
  2. まとめてimport
      ~~~go
      import (
        "fmt"
        "os"
        "time"
      )
      ~~~
  3. alias(別名)でimport
      ~~~go
      import (
        f "fmt" -→ f.Println() になる 
        "os"
        t "time" -→ t.Sleep() になる
      )
      ~~~
- importパッケージ名はファイル名ではなく、import対象ファイルの`package`名
  - `input.go` (importされる側)
      ~~~go
      package hello ⇒ ここの名前がimport時に使われる

      import (
         "bufio"
         "fmt"
         "os"
      )

      func Input(msg string) string {
         canner := bufio.NewScanner(os.Stdin)
         fmt.Print(msg + ": ")
         scanner.Scan()
         return scanner.Text()
      }
      ~~~
   - `hello.go` (importする側)
      ~~~go
      package main

      import (
         "fmt"
         "hello" ⇒ ファイル名のinputではなく、packageで指定されたhello
      )

      func main() {
         name := hello.Input("type your name")
         fmt.Println("Hello, " + name + "!!")
      }
      ~~~

## `new()` Keyword
- 指定されたデータType(型)のZero Valueのインスタンスをメモリ空間を割り当て、割り当てられたメモリのPointerを返す
- ChannelとMap以外のすべてのデータTypeに使える

## `make()`関数
- 組み込みの関数であり、スライス（slice）、マップ（map）、チャネル（channel）を作成するために使用される
- `make`関数を使用することで、これらのデータ構造を適切に初期化し、メモリを割り当てることができる。`make`を使用せずに宣言すると`nil`の値が割り当てられ、使用前に初期化する必要がある。  
  また、`make`はこれらのデータ構造に特化した関数であり、他の型の変数を作成するためには使用できない。他の型の変数を作成する場合は、`var`による宣言や`:=`を使用した短い宣言などを使用する。
  - 例  
    ```go
    func main() {
      map1 := make(map[string]int)
      map1["key1"] = 1
      fmt.Println("map1:", map1) // "map1: map[key1:1]" が出力

      var map2 map[string]int
      map2["key2"] = 2 // "panic: assignment to entry in nil map" とPanicになる
      fmt.Println("map2:", map2)
    }
    ```

#### スライス（slice）の作成
```go
slice := make([]int, length, capacity)
```
- `length`は初期長を指定
- `capacity`は容量（オプション）を指定。省略した場合はlengthと同じ値になる。

#### マップ（map）の作成
```go
m := make(map[keyType]valueType, capacity)
```
- `keyType`はマップのキーの型を指定
- `valueType`はマップの値の型を指定
- `capacity`は容量（オプション）を指定。省略した場合はデフォルトの小さな容量で初期化される
  - mapのサイズが大きくなるにつれて自動的に拡張
  - 小さなmapや、サイズが不明な場合に適している
  - 指定した場合は、指定した容量に基づいて内部メモリが事前に割り当てられる。パフォーマンスの最適化に役立つ。特に大きなmapを作成する場合に有効。

#### チャネル（channel）の作成
```go
ch := make(chan elementType, bufferSize)
```
- `elementType`はチャネルを通して送受信される要素の型を指定
- `bufferSize`はチャネルのバッファサイズ（オプション）を指定。省略するとバッファなしのチャネルが作成される。

## 各型のデフォルト値(Zero Value)
- int  
  → 0
- string  
  → "" (empty string)
- bool  
  → false
- floats  
  → 0.0
- その他  
  → nil

## Goは独自のTypeを作成することができる
- Format
  ~~~go
  type <Type名> <実際の型>
  ~~~
  → この「実際の型」を`Underlying Type`という
- 例
  ~~~go
  type my_type string
  ~~~
- 独自Typeを使って変数宣言
  ~~~go
  var x my_type
  ~~~
- 独自Typeの型は`main.<Type名>`
  ~~~go
  var x my_type
  fmt.Printf("%T\n", x)  → main.my_typeと出力される
  ~~~
- 独自Typeの`Underlying Type`の変数に直接代入はできない（Conversionが必要）
  - NG
    ~~~go
    type my_type int
    var x my_type = 10
    var y int
    y = x
    ~~~
  - OK
    ~~~go
    type my_type int
    var x my_type = 10
    var y int
    y = int(x)
    ~~~

## if文
~~~go
if 条件式 {
 　・・・処理・・・
} else if  条件式  {
 　・・・処理・・・
} else {
 　・・・処理・・・
}
~~~
- if文の中で変数を初期化して使うことができる
  - `if <変数初期化>; <条件> {}`
  - if文の中で定義(初期化)して変数はif文の中でしか使えない
  - 例
    ~~~go
    if num := 9; num < 0 {
        fmt.Println(num, "is negative")
    } else if num < 10 {
        fmt.Println(num, "has 1 digit")
    } else {
        fmt.Println(num, "has multiple digits")
    }
    // fmt.Println(num) → num is undefinedとエラーになる
    ~~~ 

## Conditional logic operators
- `&&`
  - 複数の条件を**AND**で比較
  - 例
    ~~~go
    func main() {
      fmt.Println(true && true) // → true
      fmt.Println(true && false) // → false
      fmt.Println(!true)
    }
    ~~~
- `||`
  - 複数の条件を**OR**で比較
  - 例
    ~~~go
    func main() {
      fmt.Println(true || true) // → true
      fmt.Println(true || false)// → true
      fmt.Println(!true)
    }
    ~~~
- `!`
  - 条件を**否定**
  - 例
    ~~~go
    func main() {
      if !false {
  		  fmt.Println("true, printed")
	    }

      if !true {
	  	  fmt.Println("false, Not printed")
      }

  	  if (2 == 2) {
	      fmt.Println("true, printed")
      }	

  	  if !(2 == 2) {
	      fmt.Println("false, Not printed")
      }	

  	  if !(2 != 2) {
	      fmt.Println("true, printed")
      }	
    }
    ~~~

## switch文
- caseの中でtrueとなる条件が複数ある場合、デフォルトでは上にある条件だけが実行される  
  (trueとなったらその下は判定せず抜ける)  
  ただ`fallthrough`でtrueとなる条件の下の条件も判定するようにすることができる  

   > **Warning**  
   > 基本`fallthrough`は使わないこと！

- `default`でtrueとなる(一致する)caseがない場合のみ実行する処理を定義することができる
  - `default`が実行されない例
    ~~~go
    func main() {
        switch {
        case (2 == 2):
            fmt.Println("this should print")
        default:
            fmt.Println("this is default")
        }
    }
    ~~~
  - `default`が実行される例
    ~~~go
    func main() {
        switch {
        case false:
            fmt.Println("this should not print")
        case (2 == 4):
            fmt.Println("this should not print2")
        default:
            fmt.Println("this is default")
        }
    }
    ~~~
- `case`で`,`区切りはORを意味する
  - 以下の例では"miss money or bond or dr no"が出力される
    ~~~go
    func main() {
      n := "Bond"
      switch n {
      case "Moneypenny", "Bond", "Do No":
        fmt.Println("miss money or bond or dr no")
      case "M":
        fmt.Println("m")
      default:
        fmt.Println("this is default")
      }
    }
    ~~~
- 例文  ([参考URL](https://gobyexample.com/switch))
  ~~~go
  func main() {

      i := 2
      fmt.Print("Write ", i, " as ")
      switch i {
      case 1:
          fmt.Println("one")
      case 2:
          fmt.Println("two")
      case 3:
          fmt.Println("three")
      }

      switch time.Now().Weekday() {
      case time.Saturday, time.Sunday: // ","区切り = or
          fmt.Println("It's the weekend")
      default:
          fmt.Println("It's a weekday")
      }

      t := time.Now()
      switch {
      case t.Hour() < 12:
          fmt.Println("It's before noon")
      default:
          fmt.Println("It's after noon")
      }

      whatAmI := func(i interface{}) {
          switch t := i.(type) {
          case bool:
              fmt.Println("I'm a bool")
          case int:
              fmt.Println("I'm an int")
          default:
              fmt.Printf("Don't know type %T\n", t)
          }
      }
      whatAmI(true)
      whatAmI(1)
      whatAmI("hey")
  }
  ~~~

## for文
- 3つの書き方がある
  1. `for init; condition; post { }`
      - 例
        ~~~go
	      for i := 0; i <= 10; i++ {
		      fmt.Println(i)
      	}
        ~~~
  2. `for condition { }`
      - `while <condition>`のような感じ
      - 例
        ~~~go
        func main() {
	        x := 1
	        for x < 10 {
		        fmt.Println(x)
		        x++
	        }
	        fmt.Println("done.")
        }
        ~~~
  3. `for { }`
      - `while true`のような感じ
      - 例
        ~~~go
        func main() {
	        x := 1
	        for {
		        if x > 9 {
			        break
		        }
		        fmt.Println(x)
		        x++
	        }
	        fmt.Println("done.")
        }
        ~~~

## break & continueについて
- continueはループ処理の先頭に戻る（continueの下は実行されない）
- breakはfor文から抜ける
  - 例（2から2の倍数だけ100まで出力されて最後にdoneが出力される）  
    → ２の倍数じゃない数字は`if x%2 != 0 { continue }`でループの先頭に戻るので出力されない
    ~~~go
    func main() {
	    x := 1
	    for {
		    x++
		    if x > 100 {
			    break
		    }

		    if x%2 != 0 {
			    continue
		    }

		    fmt.Println(x)
		    }

	    fmt.Println("done.")
    }
    ~~~
- 参考URL
  - https://itsakura.com/go-for

## Label付きfor break
- 多重for文でどのfor文をbreakするか指定することができる
- Label名は任意
- `for -> select -> case`文の時も使う
  - https://www.ardanlabs.com/blog/2013/11/label-breaks-in-go.html
- 例えば以下のような多重for文はbreakが内側のfor文にあるため無限ループになる
  ~~~go
  func main() {
    for {
        for {
          fmt.Println("Start")
          break
        }
        fmt.Println("End")
    }
  }
  ~~~
- for文の前にfor文のLabelを付けてbreakの次に抜けるfor文のLabelを指定する  
  → 下記の例だと"End"は出力されず、"Start"だけ出力されて終わる
  ~~~go
  func main() {
  LeeLoop:
    for {
      for {
        fmt.Println("Start")
        break LeeLoop
      }
      fmt.Println("End")
    }
  }
  ~~~

## deferについて
- deferは関数内の記述場所に関係なく、関数内のすべての処理か完了して関数が終了する直前に実行される
- ファイルをcloseする処理などで良く使われる
- 例  
  → foo()が上にあるけどbar()が先に実行されてbar → fooの順に出力される
  ~~~go
  func main() {
	  defer foo()
	  bar()
  }

  func foo() {
	  fmt.Println("foo")
  }

  func bar() {
	  fmt.Println("bar")
  }
  ~~~

## ブロックスコープ
- ブロックスコープは、その中で宣言された変数がそのブロック内でのみ有効となるスコープを作成
- そのブロック内でのみ使用する一時的な変数を作成したり、同じ名前の変数を別のスコープで再利用することが可能になる
```go
x := 10
{
    x := 20
    fmt.Println(x)  // 20が出力される
}
fmt.Println(x)  // 10が出力される
```

## 型変換（Convert）
#### 文字列と数値の型変換
- `strconv`というパッケージを使って型変換を行う
- 1目の変数には変換後の型の値が渡されて、2つ目の変数(err)には型変換に失敗した時、
 「error」型のエラー情報が渡される（正常に型変換された場合は「nil」が渡される）
  1. 文字列(Ascii) → 数値(Int)
   ~~~go
   変数, err = strconv.Atoi(string)
   ~~~
  2. 数値(Int) → 文字列(Ascii)  
     ※**数値(Int) → 文字列(Ascii)の変換は常に成功するため、エラー値(第２戻り値)はない**
   ~~~go
   変数 = strconv.Itoa(int)
   ~~~

#### byteスライスから文字列に変換
- `string()`を使って返還
- 例
  ~~~go
  r := []rune{'h', 'e', 'l', 'l', 'o'}
  s := string(r)
  // s は "hello"
  ~~~

#### intからfloat64に変換
- `float64(<int>)`するだけ
- 例  
  ```go
  package main

  import "fmt"

  func main() {
      i := 5
      f := float64(i)
      fmt.Printf("f is %f\n", f)
  }
  ```

#### float64からintに変換
- `int(<float64>)`するだけ
- 例
  ```go
  package main
  import "fmt"
  func main() {
    var x float64 = 5.7
    var y int = int(x)
    fmt.Println(y)  // outputs "5"
  }
  ```

## `fmt.Sprintf`と`fmt.Printf`の違い
- `fmt.Sprintf`は文字列を生成してそれを返すのに対し、`fmt.Printf`は文字列を生成してそれをコンソールに出力する。さらに、`fmt.Printf`は生成した文字列を返さない。
### `fmt.Sprintf`
- `fmt.Sprintf`関数は、フォーマット指定子を使用して変数を文字列化し、その結果を返すが、実際には何も出力しない。
- 例
  ~~~go
  i := 42
  s := fmt.Sprintf("数値は %d です", i)
  // s には "数値は 42 です" という文字列が格納される
  ~~~

### `fmt.Printf`
- `fmt.Printf`関数は、フォーマット指定子を使用して変数を文字列化し、その結果を標準出力（通常はコンソール）に出力する。ただし、この関数は返り値を提供しない（正確には、出力されたバイト数とエラーを返すが、これは通常は無視される）。
- 例
  ~~~go
  i := 42
  fmt.Printf("数値は %d です", i)
  // コンソールに "数値は 42 です" と出力される
  ~~~
#### `fmt.Printf`で使用できる`%`フォーマット指定子
1. 一般的な指定子:
   - `%v`: デフォルトフォーマットで値を出力
   - `%+v`: 構造体のフィールド名も含めて出力
   - `%#v`: Go言語の構文に近い形で値を出力

2. 整数:
   - `%d`: 10進数
   - `%b`: 2進数
   - `%o`: 8進数
   - `%x`, `%X`: 16進数 (小文字/大文字)

3. 浮動小数点数:
   - `%f`: 小数点表記
   - `%e`, `%E`: 指数表記
   - `%g`, `%G`: `%e` と `%f` のうち、より短い方を選択

4. 文字列:
   - `%s`: 文字列
   - `%q`: クォートされた文字列

5. 文字:
   - `%c`: 文字（Unicode コードポイント）

6. ポインタ:
   - `%p`: ポインタのアドレス

7. 真偽値:
   - `%t`: true または false

8. 幅と精度の指定:
   - `%5d`: 最小幅5桁で整数を右寄せ
   - `%-5d`: 最小幅5桁で整数を左寄せ
   - `%.2f`: 小数点以下2桁で浮動小数点数を表示
   - `%9.2f`: 最小幅9桁、小数点以下2桁で浮動小数点数を表示

9. その他:
   - `%%`: パーセント記号自体を出力

- 使用例:
  ```go
  package main

  import "fmt"

  func main() {
      i := 15
      f := 3.14159
      s := "Hello"
      b := true

      fmt.Printf("整数: %d, %b, %o, %x\n", i, i, i, i)
      fmt.Printf("浮動小数点: %f, %e, %g\n", f, f, f)
      fmt.Printf("文字列: %s, %q\n", s, s)
      fmt.Printf("真偽値: %t\n", b)
      fmt.Printf("幅指定: |%5d|%-5d|\n", i, i)
      fmt.Printf("精度指定: %.2f\n", f)
      fmt.Printf("パーセント記号: %%\n")
  }
  ```

- 出力:
  ```
  整数: 15, 1111, 17, f
  浮動小数点: 3.141590, 3.141590e+00, 3.14159
  文字列: Hello, "Hello"
  真偽値: true
  幅指定: |   15|15   |
  精度指定: 3.14
  パーセント記号: %
  ```

##### `%v`について
- 与えられた値の型に応じて適切な出力形式を自動的に選択する

1. 基本型:
   - 整数: 10進数で出力 (`%d`と同等)
   - 浮動小数点数: 小数点表記で出力 (`%g`と同等)
   - 文字列: クォートなしで出力 (`%s`と同等)
   - ブール値: `true`または`false`で出力 (`%t`と同等)

2. 配列、スライス:
   - 要素をカンマで区切り、括弧`[]`で囲んで出力

3. マップ:
   - キーと値のペアをコロンで区切り、括弧`map[]`で囲んで出力

4. 構造体:
   - フィールド名を省略し、値のみを中括弧`{}`で囲んで出力

5. ポインタ:
   - アドレスではなく、ポインタが指す値を出力

6. チャネル、関数、複素数:
   - 型に応じた適切な表現で出力

- 例：
  ```go
  package main

  import "fmt"

  type Person struct {
      Name string
      Age  int
  }

  func main() {
      i := 42
      f := 3.14
      s := "hello"
      b := true
      arr := [3]int{1, 2, 3}
      slc := []string{"a", "b", "c"}
      m := map[string]int{"one": 1, "two": 2}
      p := Person{Name: "Alice", Age: 30}
      ptr := &i

      fmt.Printf("整数: %v\n", i)
      fmt.Printf("浮動小数点: %v\n", f)
      fmt.Printf("文字列: %v\n", s)
      fmt.Printf("ブール: %v\n", b)
      fmt.Printf("配列: %v\n", arr)
      fmt.Printf("スライス: %v\n", slc)
      fmt.Printf("マップ: %v\n", m)
      fmt.Printf("構造体: %v\n", p)
      fmt.Printf("ポインタ: %v\n", ptr)
  }
  ```

- 出力:
  ```
  整数: 42
  浮動小数点: 3.14
  文字列: hello
  ブール: true
  配列: [1 2 3]
  スライス: [a b c]
  マップ: map[one:1 two:2]
  構造体: {Alice 30}
  ポインタ: 42
  ```

## init関数
- mainパッケージでimportしたPackageにinit関数がある場合、mainパッケージ内のinit関数よりPackage内のinit関数が先に実行される
  - importされるタイミングで実行される
- 1つのコード内に複数のinit関数を定義することは一応できる。(実際に複数のinit関数を定義することはないだろう)
- 下記例の場合、"Hello from somepackage"が先に出力されて、その後"init in main package"が出力される
  - `main.go`  
    ```go
    package main

    import (
      "somepackage"
      "fmt"
    )

    func init() {
      fmt.Println("init in main package")
    }

    func main() {
        ・
        ・
    }
    ```
  - `somepackage.go`  
    ```go
    package somepackage

    import (
      "fmt"
    )

    func init() {
      fmt.Println("Hello from somepackage")
    }
    ```

## 文字列操作
### 文字列検索
- `strings`パッケージの`Contains`関数である文字列(変数)の中に特定の文字列が含まれているか確認することができる
  - `Contains`関数の戻り値は`bool`型で含まれているときは`true`、含まれてないときは`false`が返される
- 例  
  ```go
  import (
    "fmt"
    "strings"
  )

  func main() {
    fmt.Println(strings.Contains("seafood", "foo"))
    fmt.Println(strings.Contains("seafood", "bar"))
    fmt.Println(strings.Contains("seafood", ""))
    fmt.Println(strings.Contains("", ""))
  }

  // Output:
  // true
  // false
  // true
  // true
  ```

### 文字列の分割
- `strings`パッケージの`Split`関数で文字列を特定の区切り文字でスライスに分割して格納することができる
  - `Split`関数の第１引数に分割対象の文字列、第２引数に区切り文字を指定
```go
import (
    "fmt"
    "strings"
)

func main() {
    str := "a b"
    result := strings.Split(str, " ")
    
    fmt.Println(result) // 出力: [a b]
    
    // 個別の要素にアクセスする場合
    fmt.Println(result[0]) // 出力: a
    fmt.Println(result[1]) // 出力: b
}
```

### 文字列の連結
- `+`で連結できる
  ```go
  func main() {
    string1 := "something"
    string2 := "cool"
    combine := string1 + string2
    fmt.Println(combine) // "somethingcool"
  }
  ```

### 小文字 ⇔ 大文字変換
1. `strings.ToUpper()` と `strings.ToLower()` 関数を使う方法  
  - 引数と戻り値ともに`string`型  
  ```go
  import "strings"

  str := "Hello, World!"
  upper := strings.ToUpper(str) // "HELLO, WORLD!"
  lower := strings.ToLower(str) // "hello, world!"
  ```

2. `unicode.ToLower()`と`unicode.ToUpper()`関数を使う方法  
  - 引数と戻り値ともに`rune`型
    - `rune`型
       - Goの組み込み型の一つで、Unicode文字（コードポイント）を表す
       - int32型のエイリアス
       - 1つのUnicodeコードポイントを表現  
         ```go
          var r rune = 'A'
          fmt.Printf("%c %d %T\n", r, r, r)
          // 出力: A 65 int32
         ```
       - 文字列に対してrange文を使うと、自動的にrune型の値が得らる  
         ```go
          for _, r := range "Hello" {
              fmt.Printf("%c ", r)
          }
          // 出力: H e l l o
         ```
         ```go
         func solution(str1 string, str2 string) string {
             var sumstr string
             for i,_ := range str1 {
                 sumstr += string(str1[i]) + string(str2[i])         
             }
             return sumstr
         }
         ```
  ```go
  func main() {
    var result string
    s1 := "Hello"
    for _, r := range s1 {
        if unicode.IsUpper(r) {
            result = result + string(unicode.ToLower(r))
        } else {
            result = result + string(unicode.ToUpper(r))
        }
    }
    fmt.Println(result) // hELLO
  }
  ```

### 文字が大文字か小文字かを確認
#### `unicode.IsUpper()`、`unicode.IsLower()`関数で1文字ずつ確認  
```go
func main() {
  s1 := "HellO"
  for _, r := range s1 {
      if unicode.IsUpper(r) {
          fmt.Println("It is Upper")
      } else if unicode.IsLower(r) {
          fmt.Println("It is Lower")
      }
  }
}
```
#### `regexp` パッケージを使用して、正規表現で確認
```go
import (
    "fmt"
    "regexp"
)

func main() {
    text := "The quick brown fox jumps over the lazy dog"
    pattern := `[A-Z]`

    matched, _ := regexp.MatchString(pattern, text)
    fmt.Println(matched) // true（最初のT）
}
```

### ある文字列をN回繰り返してくっつける
- `strings`パッケージの`strings.Repeat`を使用  
  ```go
  import "strings"

  repeatedString := strings.Repeat("元の文字列", N)
  ```

### 特殊文字をそのまま出力させる
- `` で囲むと特殊文字を解釈せずそのまま出力してくれる  
  ```go
  func main() {
      fmt.Println(`!@#$%^&*(\'"<>?:;`)   
  }
  ```

## `error`型の戻り値を返す方法
1. `errors.New()`を使う  
   ```go
   import "errors"

   func someFunction() error {
       // エラーが発生した場合
       return errors.New("エラーメッセージ")
   }
   ```
2. `fmt.Errorf()`を使う  
   ```go
   import "fmt"

   func someFunction() error {
       // エラーが発生した場合
       return fmt.Errorf("エラーメッセージ: %v", someValue)
   }
   ```

## 標準入力を受け付けて処理
### `fmt.Scan`を使う方法
```go
package main

import "fmt"

func main() {
    var input string
    fmt.Scan(&input)
    fmt.Printf("入力された文字列: %s\n", input)
}
```

### `bufio.Scanner`を使う方法
```go
package main

import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("文字列を入力してください: ")
    scanner.Scan()
    input := scanner.Text()
    fmt.Printf("入力された文字列: %s\n", input)
}
```

### `bufio.Scanner`を使う方法（Ctrl+Dや特定の文字入力まで入力を受け付け続ける）
```go
package main

import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("複数行の入力を受け付けます。終了するには Ctrl+D を押してください:")
    for scanner.Scan() {
        line := scanner.Text()
        if line == "x" || line == "exit" {
            break
      	}
        fmt.Printf("入力された行: %s\n", line)
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "読み込みエラー:", err)
    }
}
```