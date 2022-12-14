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
- main関数の中で`return`でプログラムが終了する

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

### `...`について（Lexical elementsと呼ぶらしい）
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
    - `...<型>`が複数の引数の中で最後にある場合、呼び出す側は`<型>...`がなくても良い（その場合`...<型>`にはnilが連携される）
      https://go.dev/ref/spec
      > Passing arguments to ... parameters
      > If f is variadic with a final parameter p of type ...T, then within f the type of p is equivalent to type []T. If f is invoked with no actual arguments for p, the value passed to p is nil. Otherwise, the value passed is a new slice of type []T with a new underlying array whose successive elements are the actual arguments, which all must be assignable to T. The length and capacity of the slice is therefore the number of arguments bound to p and may differ for each call site.
      > 
      > Given the function and calls
      > 
      > func Greeting(prefix string, who ...string)
      > Greeting("nobody")
      > Greeting("hello:", "Joe", "Anna", "Eileen")
      > within Greeting, who will have the value nil in the first call, and []string{"Joe", "Anna", "Eileen"} in the second.
      > 
      >If the final argument is assignable to a slice type []T and is followed by ..., it is passed unchanged as the value for a ...T parameter. In this case no new slice is created.
      > 
      > Given the slice s and call
      > 
      > s := []string{"James", "Jasmine"}
      > Greeting("goodbye:", s...)
      > within Greeting, who will have the same value as s with the same underlying array.

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

### Slices
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
  2. `<変数> := make([]<型>{<長さ>, <容量>})`  
     `<変数> := make([]<型>{<長さ>)` → 容量を省略した場合は容量=長さ
      ~~~go
      x := make([]int{5, 10})
      y := make([]int{5})
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
      ※byteは`uint8`型の別名。uint8型は8bit、つまり1バイト分の表現が可能で、データを1バイトごとに分割して扱う。[参照ページ](https://qiita.com/s_zaq/items/a6d1df0d1bf79f5bcc2a)
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
- Sliceの最後の要素を取得する方法
  - pythonみたいに`<Slice名>[-1]`では取得できない(エラーになる)
  - `<Slice名>[len(<Slice名>)-1]`というふうにSliceの長さから-1して最後の要素を取得する
    - 例：`fmt.Println("Metric:",*resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value)`

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

### 関数
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

### Struct (構造体)
- 色んな型をひとまとめにしたもの
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

### Methods
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

### Interfaces
- InterfaceはMethod(s)を持つ(Methodのラッピング？)
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

  func DisplayArea(f Figure) {
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
- 参考URL
  - https://go.dev/play/p/rZH2Efbpot
  - https://dev-yakuza.posstree.com/golang/interface/

### CallBack
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

### Closure
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

### Recursion
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

### Pointer(ポインタ)
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
- **ポインタの利用シーン**
  1. big chunk of dataを受け渡ししたい場合
  2. 特定のメモリアドレスにある値を変更したい場合

### Goroutine
- Goroutineは並列処理を保証するのではなく、並列処理を実行できる環境の場合のみ並列処理をする
  - 例えばcpuコアが1つしかないコンピューターではGoroutineを使っても、並列(parallel)ではなく、並行(concurrent)処理になる
- 関数の前に`go`をつけるとGoroutineになる
  - 例：`go foo()`
- **GoRoutine処理は順番が保証されない**(毎回順番が異なる)。  
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
- __WaitGroup__
  - Goroutineで実行した処理はデフォルトでは待ってもらえず、main関数が終了すればGoroutine処理が終わってなくてもプログラムは終了してしまう
  - ProcessがkillされるとすべてのGoroutineもcancelされる
    - Goroutineのleakを防ぐため（Goroutineはmemoryなどリソースを消費する）
  - Goroutine処理が終わるまで待ってもらうためのものが`WaitGroup`
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
    - OK例
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

### Context
- (一連の)Deadline(`WithDeadline`)やTimeout(`WithTimeout`)を設定してそれが過ぎたら親Goroutineとそこ派生したすべてのGoroutineをcancelしてリソースleakを防いだり、`WithCancel`でGoroutineをcancelするタイミングを制御したり、`WithValue`で1つのrequestで関連する複数のGoroutine間で値をpassしたりするために使われる
  > In Go servers, each incoming request is handled in its own goroutine. Request handlers often start additional goroutines to access backends such as databases and RPC services. The set of goroutines working on a request typically needs access to request-specific values such as the identity of the end user, authorization tokens, and the request’s deadline. When a request is canceled or times out, all the goroutines working on that request should exit quickly so the system can reclaim any resources they are using.
  > 
  > At Google, we developed a context package that makes it easy to **pass request-scoped values, cancellation signals, and deadlines across API boundaries to all the goroutines involved in handling a request.**
- Contextがcancelされると、そのGoroutineから派生したすべてのGoroutineがcancelされる
  > When a Context is canceled, all Contexts derived from it are also canceled.
- 以下4つのfunctionがある
  1. `WithCancel` -> [例](https://pkg.go.dev/context#example-WithCancel)
  2. `WithDeadline` -> [例](https://pkg.go.dev/context#example-WithDeadline)
  3. `WithTimeout` -> [例](https://pkg.go.dev/context#example-WithTimeout)
  4. `WithValue` -> [例](https://pkg.go.dev/context#example-WithValue)
- `Done()`は`select`文の中で使う必要がある
  ~~~go
  package main

  import (
      "context"
      "fmt"
      "runtime"
      "time"
  )

  func main() {
      ctx, cancel := context.WithCancel(context.Background())

      fmt.Println("error check 1:", ctx.Err())
      fmt.Println("num gortins 1:", runtime.NumGoroutine())

      go func() {
          n := 0
          for {
              select {
              case <-ctx.Done():
                  return // go func()を抜ける
              default:
                  n++
                  time.Sleep(time.Millisecond * 200)
                  fmt.Println("working", n)
              }
          }
      }()

      time.Sleep(time.Second * 2)
      fmt.Println("error check 2:", ctx.Err())
      fmt.Println("num gortins 2:", runtime.NumGoroutine())

      fmt.Println("about to cancel context")
      cancel()
      fmt.Println("cancelled context")

      time.Sleep(time.Second * 2)
      fmt.Println("error check 3:", ctx.Err())
      fmt.Println("num gortins 3:", runtime.NumGoroutine())
  }
  ~~~
- 参考URL
  - **https://go.dev/blog/context**
  - **https://peter.bourgon.org/blog/2016/07/11/context.html**
  - https://pkg.go.dev/context#pkg-overview
  - https://zenn.dev/hsaki/books/golang-context/viewer/definition

### Channels
- Channels are the pipes that connect concurrent goroutines. You can send values into channels from one goroutine and receive those values into another goroutine.
- Dataを送受信できる空間
- ChannelもType
- **Channelは値を入れた後に遮断される**
  - なので**goroutine**で**channelに値を入れるのとchannelから値を取り出すのを同時に**するようにコードを書く必要がある
  - または**buffer**を使ってchannelに値が残れるようにする
    - 定義したbufferの数より多くの数の値をchannelに入れようとするとエラーになる  
      → 定義した数の分がbufferに入ってきたらchannelは遮断される
- `make(chan int)`でChannelを作成する  
  → intを入れるchannel
  - buffer channelを作る場合は`make(chan int, <buffer数>)`
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

### select
- Select statements pull the value from whatever channel has a value ready to be pulled.
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
  ~~~
- `for`と組合せて使うことで
- 参考URL
  - https://www.spinute.org/go-by-example/select.html
  - https://go-tour-jp.appspot.com/concurrency/6
  - https://www.ardanlabs.com/blog/2013/11/label-breaks-in-go.html

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



### make


### goto



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

### Label付きfor break
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

### deferについて
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