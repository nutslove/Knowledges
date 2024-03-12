## Benchmark / Test 共通
- `import testing`でtestingパッケージのimportが必要
- テストしたいgoファイル名に`_test.go`をつける
  - ex) `main.go`の場合`main_test.go`
- main関数はテストされない
- 戻り値を持たない
- テストファイル(`_test.go`)は、テスト対象のパッケージと同じディレクトリに配置

## Benchmarkの基本
- 性能測定時使う
- Benchmark関数名は必ず`Benchmark`から始まる必要がある。
- `Benchmark`の次には`_`や大文字が来れる
  - OK例: `Benchmark_someFunction`、`BenchmarkSomeFunction`
  - NG例: `BenchmarksomeFunction`
- Benchmark関数は`*testing.B`引数のみを受け付ける

## `testing`の基本
- https://pkg.go.dev/testing
  > Package testing provides support for automated testing of Go packages.
- テスト関数はテストする関数の前に`Test`(もしくは`Test_`)を付ける
  - 例えば`main.go`内の`SomeCheck`関数をテストする場合、`main_test.go`ファイル内に`Test_SomeCheck`関数を作成する
- テスト関数は１つの`*testing.T`引数のみを受け付ける
- テスト関数内で、`t.Errorf`や`t.Fatalf`を使用してエラーを報告する
- テストコードは`go test`コマンドで実行
  - コマンドを実行したディレクトリ内のすべての`*_test.go`ファイルを検索し、その中に定義されているすべてのテストを実行する
  - `go test -run <実行したいテスト関数名>`コマンドで特定のテスト関数のみを実行することもできる
    - 例えば`Test_MyFunction`というテスト関数のみをテストしたい場合は`go test -run Test_MyFunction`を実行
  - `go test ./<対象Package名>`で特定のパッケージのみをテストすることもできる
  - `-v`オプション(`go test -v`)でテストの詳細なログを確認できる
### 例
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
### テストカバレッジ
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
