## Go Workspace
- Workspaceには以下の3つのディレクトリが必要
  - /bin
  - /pkg
  - /src  
<br>
   | パス | 内容 |
   | --- | --- |
   | /bin | コンパイルされたものが格納される |
   | /pkg | プログラムから呼び出されるライブラリ(?) |
   | /src | ?? |

  #### 環境変数
  - GOPATH
    - points to your go workspace
  - GOROOT
    - points to your binary installation of Go

## Go Module
- 

#### その他Goについて色々
- GoはClassがない（Goはオブジェクト指向言語ではない）<br><br>
- Goはtry catch(except)ではなく、errorというエラー専用型(interface)がある<br><br>
- gofmtコマンドを使うとgoのフォーマットに変換してくれる
`gofmt -w <対象goファイル>`<br><br>
- main()関数以外はmainもしくはinit関数内で明示的に呼び出す必要がある<br><br>
- init()関数がmain()関数より先に実行される<br><br>
- GoにもGCが存在する
  - ただ、fileなど明示的に開放しなければいけないものもある
  - 明示的に開放する必要があるものはdeferを使って開放する<br><br>
- Goは大文字/小文字、全角/半角は別の文字として扱われる<br><br>
- Goは宣言された変数は必ず利用されている必要がある
  (変数宣言だけしといて使わないとエラーになる)

#### Goインストール（Linux）
- `wget https://dl.google.com/go/go1.18.4.linux-amd64.tar.gz`
- `rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.4.linux-amd64.tar.gz`
- `export PATH=$PATH:/usr/local/go/bin`
- `go version`

#### 変数の書き方
1. varを使った定義
   - 関数の外側でも定義可
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

#### 定数（const）
- 作成した後に値の変更ができない
  ~~~go
  const 定数 = 値
  const 定数 型 = 値
  ~~~
- 変数とは違い、型を指定しない場合、
  定義時に代入する値から型を推論するのではなく、使われる時に型を推論する

#### 各変数のデフォルト値
- int -> 0
- string -> empty string (i.e "")

#### スライス

#### map

#### 関数
- いくつかのパターンがある 
  1. 関数名、引数、戻り値を定義
      ~~~go
      func 関数名 (引数) 戻り値 { 
         ・・・処理・・・
      }
      ~~~
  2. 

#### if文
~~~go
if 条件式 {
 　・・・処理・・・
} else if  条件式  {
 　・・・処理・・・
} else {
 　・・・処理・・・
}
~~~

#### switch文

#### ポインタ
- 値が入るメモリのアドレス
- `*int`がポインタ型変数
- メモリアドレスの指定には変数の前に`&`を指定
- メモリアドレスに格納されている値を操作する場合はポインタ型変数の前に`*`を付ける
~~~go
var n int = 100
var p *int = &n  → ポインタ型変数Pに変数nが格納されているメモリアドレスを格納
fmt.Println(p)   → "0xc00007c008"等の変数nが格納されているメモリアドレスが表示される
fmt.Println(*p)  → メモリアドレス(p)に格納されている値 100 が表示される
*p = 300         → メモリアドレス(p)に格納されている値を100 → 300 に変更
fmt.Println(*p)  → メモリアドレス(p)に格納されている値 300 が表示される
~~~

#### パッケージ(import)
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

#### 文字列と数値の型変換
- 「strconv」というパッケージを使って型変換を行う
- 1目の変数には変換後の型の値が渡されて、2つ目の変数(err)には型変換に失敗した時、
 「error」型のエラー情報が渡される（正常に型変換された場合は「nil」が渡される）
  1. 文字列(Ascii) → 数値(Int)
   ~~~go
   変数, err = strconv.Atoi(string)
   ~~~
  2. 数値(Int) → 文字列(Ascii)
   ~~~go
   変数, err = strconv.Itoa(int)
   ~~~