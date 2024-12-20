# 目次
<!-- TOC -->

- [目次](#%E7%9B%AE%E6%AC%A1)
- [Benchmark / Testtesting 共通](#benchmark--testtesting-%E5%85%B1%E9%80%9A)
- [Benchmark](#benchmark)
- [testing](#testing)
  - [例](#%E4%BE%8B)
  - [table testsの他の例](#table-tests%E3%81%AE%E4%BB%96%E3%81%AE%E4%BE%8B)
  - [テストカバレッジ](#%E3%83%86%E3%82%B9%E3%83%88%E3%82%AB%E3%83%90%E3%83%AC%E3%83%83%E3%82%B8)
  - [t.Runメソッドによるサブテスト](#trun%E3%83%A1%E3%82%BD%E3%83%83%E3%83%89%E3%81%AB%E3%82%88%E3%82%8B%E3%82%B5%E3%83%96%E3%83%86%E3%82%B9%E3%83%88)
  - [テストの並列実行（Parallelメソッド）](#%E3%83%86%E3%82%B9%E3%83%88%E3%81%AE%E4%B8%A6%E5%88%97%E5%AE%9F%E8%A1%8Cparallel%E3%83%A1%E3%82%BD%E3%83%83%E3%83%89)
    - [Parallelメソッドの特性](#parallel%E3%83%A1%E3%82%BD%E3%83%83%E3%83%89%E3%81%AE%E7%89%B9%E6%80%A7)

<!-- /TOC -->
- 性能測定時使う
- Benchmark関数名は必ず`Benchmark`から始まる必要がある。
- `Benchmark`の次には`_`や大文字が来れる
  - OK例: `Benchmark_someFunction`、`BenchmarkSomeFunction`
  - NG例: `BenchmarksomeFunction`
- Benchmark関数は`*testing.B`引数のみを受け付ける

# `testing`
- https://pkg.go.dev/testing
  > Package testing provides support for automated testing of Go packages.
- テスト関数はテストする関数の前に`Test`(もしくは`Test_`)を付ける
  - 例えば`main.go`内の`SomeCheck`関数をテストする場合、`main_test.go`ファイル内に`Test_SomeCheck`関数を作成する
  - **`Test`の次に小文字は来れない。なので小文字で始まる関数をテストしたい場合は`Test`の次に`_`をつけてから小文字で始まる関数を指定すること！**
- テスト関数は１つの`*testing.T`引数のみを受け付ける
- テスト関数内で、`t.Errorf`や`t.Error`、`t.Fatalf`を使用してエラーを報告する
- テストコードは`go test`コマンドで実行
  - コマンドを実行したディレクトリ内のすべての`*_test.go`ファイルを検索し、その中に定義されているすべてのテストを実行する
  - `go test -run <実行したいテスト関数名>`コマンドで特定のテスト関数のみを実行することもできる
    - 例えば`Test_MyFunction`というテスト関数のみをテストしたい場合は`go test -run Test_MyFunction`を実行
  - `go test ./<対象Package名>`で特定のパッケージのみをテストすることもできる
  - `-v`オプション(`go test -v`)でテストの詳細なログを確認できる
## 例
- `calc.go`  
  ~~~go
  package calc

  func Add(a, b int) int {
      return a + b
  }

  func Subtract(a, b int) int {
      return a - b
  }
  ~~~
- `calc_test.go`  
  ~~~go
  package calc

  import "testing"

  func TestAdd(t *testing.T) {
      got := Add(2, 3)
      want := 5
      if got != want {
          t.Errorf("Add(2, 3) = %d, want %d", got, want)
      }
  }

  func TestSubtract(t *testing.T) {
      // 以下のようにstructにテストケースをまとめてテストするのを「table tests」という
      cases := []struct {
          a, b int
          want int
      }{
          {5, 3, 2},
          {10, 7, 3},
          {0, 0, 0},
      }

      for _, c := range cases {
          got := Subtract(c.a, c.b)
          if got != c.want {
              t.Errorf("Subtract(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
          }
      }
  }
  ~~~

## table testsの他の例
- main.go
  ```go
  package main

  import "fmt"

  func main() {
    n := 2

    _, msg := isPrime(n)
    fmt.Println(msg)
  }

  func isPrime(n int) (bool, string) {
    // 0 and 1 are not prime by definition
    if n == 0 || n == 1 {
      return false, fmt.Sprintf("%d is not prime, by definition!", n)
    }

    // negative numbers are not prime
    if n < 0 {
      return false, "Negative numbers are not prime, by definition!"
    }

    // use the modulus operator repeatedly to see if we have a prime number
    for i := 2; i <= n/2; i++ {
      if n%i == 0 {
        // not a prime number
        return false, fmt.Sprintf("%d is not a prime number because it is divisible by %d!", n, i)
      }
    }

    return true, fmt.Sprintf("%d is a prime number!", n)
  }
  ```
- main_test.go
  ```go
  package main

  import "testing"

  func Test_isPrime(t *testing.T) {
    primeTests := []struct {
      name string // test case名
      testNum int
      expected bool
      msg string
    }{
      {"prime", 7, true, "7 is a prime number!"},
      {"not prime", 8, false, "8 is not a prime number because it is divisible by 2!"},
    }

    for _, e := range primeTests {
      result, msg := isPrime(e.testNum)
      if e.expected && !result {
        t.Errorf("%s: expected true but got false", e.name)
      }

      if !e.expected && result {
        t.Errorf("%s: expected false but got true", e.name)
      }

      if e.msg != msg {
        t.Errorf("%s: expected %s but got %s", e.name, e.msg, msg)
      }
    }
  }
  ```

## テストカバレッジ
- テストカバレッジとは、テストスイート(複数のテストケースをまとめたもの)がテスト対象のコードをどれだけカバーしているかを示す指標
- `go test -cover`でテストカバレッジを確認できる
- 例えば、以下の`main.go`を以下の`main_test.go`を使って、`go test -cover`と実行すると、下記のように出力される。  
  - `main.go`  
    ~~~go
    package main

    func Add(a, b int) int {
        return a + b
    }

    func Multiply(a, b int) int {
        return a * b
    }

    func main() {
        // Some code here
    }
    ~~~
  - `main_test.go`  
    ~~~go
    package main

    import "testing"

    func TestAdd(t *testing.T) {
        got := Add(2, 3)
        want := 5
        if got != want {
            t.Errorf("Add(2, 3) = %d, want %d", got, want)
        }
    }
    ~~~
  - 出力  
    ~~~
    PASS --> テスト結果
    coverage: 50.0% of statements --> カバレッジ(テスト対象コードの50%がテストによってカバーされているということ)
    ok      example/package    0.013s --> 実行時間
    ~~~

## `t.Run()`メソッドによるサブテスト
- 1つのテスト関数内に`t.Run()`メソッドでサブテストを定義することができる
- 例１  
  ```go
  func TestExample(t *testing.T) {
      t.Run("ケース1", func(t *testing.T) {
          // テストケース1の内容
      })
      
      t.Run("ケース2", func(t *testing.T) {
          // テストケース2の内容
      })
  }
  ```
- 以下のように特定のサブテストのみを実行することもできる
  - `go test -run TestExample/ケース1  # "ケース1"のみ実行`

## テストの並列実行（`Parallel()`メソッド）
- 参考URL
  - https://engineering.mercari.com/blog/entry/how_to_use_t_parallel/
- `testing`パッケージを使ったテストコードの実行は、デフォルトでは**パッケージ内では逐次的に**、**パッケージごとは並列に実行される**
  - 例えば、aパッケージとbパッケージがあった場合、aパッケージ内のテストコードは逐次実行され、bパッケージ内のテストコードも逐次実行される。しかし、aパッケージとbパッケージのテストは並列に実行される。
- パッケージ内で並列に実行するために`Parallel()`メソッドを使用

### `Parallel()`メソッドの特性
- **`Parallel()`メソッドを設定しているテスト関数は、他の`Parallel()`メソッドを設定しているテスト関数とのみ並列に実行される**  
- **`t.Parallel()`メソッドの呼び出しは、一時停止してから再開する（一時停止した場合、`=== PAUSE`と表示され、処理が再開した場合、`=== CONT`と表示される）**  
- **`t.Parallel()メソッド`を呼び出していない（パッケージ内の）すべてのトップレベルのテスト関数が終了してから、`t.Parallel()`メソッドを呼び出しているトップレベルのテスト関数の処理が再開して並列に実行される**
- **`t.Run()`によるサブテスト関数内で`t.Parallel()`メソッドを呼び出している場合、その親のトップレベルのテスト関数が「終了して戻る」まで、サブテスト関数は`t.Parallel()`メソッドの呼び出しで一時停止する**
- 最大並列数は`-parallel`フラグで指定可能
  - デフォルトでは`GOMAXPROCS`の値が設定される
- 例１  
  - コード
    ```go
    package main

    import (
        "fmt"
        "testing"
    )

    func trace(name string) func() {
        fmt.Printf("%s entered\n", name)
        return func() {
            fmt.Printf("%s returned\n", name)
        }

    }

    func Test_Func1(t *testing.T) {
        defer trace("Test_Func1")()

        // ...
    }

    func Test_Func2(t *testing.T) {
        defer trace("Test_Func2")()
        t.Parallel()

        // ...
    }

    func Test_Func3(t *testing.T) {
        defer trace("Test_Func3")()

        // ...
    }

    func Test_Func4(t *testing.T) {
        defer trace("Test_Func4")()
        t.Parallel()

        // ...
    }

    func Test_Func5(t *testing.T) {
        defer trace("Test_Func5")()

        // ...
    }
    ```
  - 出力  
    ```
    === RUN   Test_Func1
    Test_Func1 entered
    Test_Func1 returned                <- 1 （完了）
    --- PASS: Test_Func1 (0.00s)
    === RUN   Test_Func2
    Test_Func2 entered
    === PAUSE Test_Func2               <- 2 (一時停止）
    === RUN   Test_Func3
    Test_Func3 entered
    Test_Func3 returned                <- 3 （完了）
    --- PASS: Test_Func3 (0.00s)
    === RUN   Test_Func4
    Test_Func4 entered
    === PAUSE Test_Func4               <- 4 (一時停止）
    === RUN   Test_Func5
    Test_Func5 entered
    Test_Func5 returned                <- 5 （完了）
    --- PASS: Test_Func5 (0.00s)
    === CONT  Test_Func2               <- 処理が再開
    Test_Func2 returned                <- 完了
    === CONT  Test_Func4               <- 処理が再開
    Test_Func4 returned                <- 完了
    --- PASS: Test_Func2 (0.00s)
    --- PASS: Test_Func4 (0.00s)
    PASS
    ```
- 例２（`t.Run()`によるサブテスト関数内で`t.Parallel()`メソッドを呼び出している場合）
  - コード  
    ```go
    package main

    import (
    	"fmt"
    	"testing"
    )

    func trace(name string) func() {
    	fmt.Printf("%s entered\n", name)
    	return func() {
    		fmt.Printf("%s returned\n", name)
    	}

    }

    func Test_Func1(t *testing.T) {
    	defer trace("Test_Func1")()

    	t.Run("Func1_Sub1", func(t *testing.T) {
    		defer trace("Func1_Sub1")()
    		t.Parallel()
    		// ...
    	})

    	t.Run("Func1_Sub2", func(t *testing.T) {
    		defer trace("Func1_Sub2")()
    		t.Parallel()
    		// ...
    	})

    	// ...
    }

    func Test_Func2(t *testing.T) {
    	defer trace("Test_Func2")()
    	t.Parallel()

    	// ...
    }

    func Test_Func3(t *testing.T) {
    	defer trace("Test_Func3")()

    	// ...
    }

    func Test_Func4(t *testing.T) {
    	defer trace("Test_Func4")()
    	t.Parallel()

    	// ...
    }

    func Test_Func5(t *testing.T) {
    	defer trace("Test_Func5")()

    	// ...
    }
    ```
  - 出力  
    ```
    === RUN   Test_Func1
    Test_Func1 entered
    === RUN   Test_Func1/Func1_Sub1
    Func1_Sub1 entered                          <- Func1_Sub1が開始
    === PAUSE Test_Func1/Func1_Sub1             <- Func1_Sub1が一時停止
    === RUN   Test_Func1/Func1_Sub2
    Func1_Sub2 entered                          <- Func1_Sub2が開始
    === PAUSE Test_Func1/Func1_Sub2             <- Func1_Sub2が一時停止
    Test_Func1 returned                         <- Test_Func1の呼び出し戻り（＊）
    === CONT  Test_Func1/Func1_Sub1             <- Func1_Sub1が再開
    Func1_Sub1 returned                         <- Func1_Sub1が完了
    === CONT  Test_Func1/Func1_Sub2             <- Func1_Sub2が再開
    Func1_Sub2 returned                         <- Func1_Sub2が完了
    --- PASS: Test_Func1 (0.00s)                <- Test_Func1の結果表示
        --- PASS: Test_Func1/Func1_Sub1 (0.00s)
        --- PASS: Test_Func1/Func1_Sub2 (0.00s)
    === RUN   Test_Func2                        <- ここまでTest_Func2は実行されない
    Test_Func2 entered
    === PAUSE Test_Func2
    === RUN   Test_Func3
    Test_Func3 entered
    Test_Func3 returned
    --- PASS: Test_Func3 (0.00s)
    === RUN   Test_Func4
    Test_Func4 entered
    === PAUSE Test_Func4
    === RUN   Test_Func5
    Test_Func5 entered
    Test_Func5 returned
    --- PASS: Test_Func5 (0.00s)
    === CONT  Test_Func2
    Test_Func2 returned
    === CONT  Test_Func4
    Test_Func4 returned
    --- PASS: Test_Func4 (0.00s)
    --- PASS: Test_Func2 (0.00s)
    PASS
    ```