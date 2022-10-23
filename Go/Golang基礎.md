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

  ### 環境変数
  - GOPATH
    - points to your go workspace
  - GOROOT
    - points to your binary installation of Go

## Go Module
- ModuleはGoパッケージ管理の手助けをしてくれるもの
  - パッケージのバージョンを固定したり、常に最新バージョンを使うように設定したりすることができる
   #### Go Module作成(初期化)
   - `go mod init <Module名>`で初期化  
     → `go.mod`が作成される
     - `go.mod`とは
       - Goモジュールのパスを書いておくファイル

### パッケージ関連
- 初期のGoでは`go get`でパッケージをビルド/インストールしていたが、  
  現在は`go get`は`go.mod`の依存関係の調整にだけ使われる。  
  現在は`go install`でパッケージのビルド/インストールを行う

## その他Goについて色々
- GoはClassがない（Goはオブジェクト指向言語ではない）
- Goはtry catch(except)ではなく、errorというエラー専用型(interface)がある
- Goにwhileはない
- gofmtコマンドを使うとgoのフォーマットに変換してくれる  
`gofmt -w <対象goファイル>`
- main()関数以外はmainもしくはinit関数内で明示的に呼び出す必要がある
- init()関数がmain()関数より先に実行される
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

### Goインストール（Linux）
- `wget https://dl.google.com/go/go1.18.4.linux-amd64.tar.gz`
- `rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.4.linux-amd64.tar.gz`
- `export PATH=$PATH:/usr/local/go/bin`
- `go version`

## Goの文法
### 変数の書き方
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

### 戻り値を`_`で捨てる
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

### Array(配列)
- Format
  1. `var <変数名>[配列数]<型>`
      ~~~go
      var x [5]int
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

### Slices
- Format
  1. `<変数> := []<型>{Values}`
      ~~~go
      x := []int{1, 2, 3, 5, 7}
      ~~~
  2. `<変数> := make([]<型>{<長さ>, <容量>})`
      ~~~go
      x := make([]int{5, 10})
      ~~~
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

### Map
- Format
  1. `var 変数 map[<Key型>]<Value型>`
  2. `変数 := map[<Key型>]<Value型> { Key1: Value1, Key2: Value2, ・・・, }`  
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

### 定数（const）
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

### 各型について
- 

### 演算子
- 

### 各型のデフォルト値(Zero Value)
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

### Goは独自のTypeを作成することができる
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

### 関数
- いくつかのパターンがある 
  1. 関数名、引数、戻り値を定義
      ~~~go
      func 関数名 (引数) 戻り値 { 
         ・・・処理・・・
      }
      ~~~
  2. 

### if文
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

### Conditional logic operators
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

### switch文
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

### for文
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

### break & continueについて
- continueの下は実行されない
- breakはfor文から抜ける
  - 例（2から2の倍数だけ100まで出力されて最後にdoneが出力される）
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

### Struct (構造体)
- 色んな型を値をひとまとめにしたもの
- 他の言語のClassのような感じで、1つのstructに対して (下のp1とp2のように) 何回でも変数宣言できる
- Format
  ~~~go
  type <type名> struct {
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
- __Embedded structs__
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

- __Anonymous structs__
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

### ポインタ
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

### goto


### make

### パッケージ(import)
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

### 文字列と数値の型変換
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