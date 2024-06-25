- https://golang.hateblo.jp/entry/golang-text-html-template

## `text/template`の基本的な使い方(文法)
Goの`text/template`パッケージでは、テンプレート内でif文を使用することができます。以下にif文の基本的な構文と使用例を示します：

基本的な構文：

```
{{if .Condition}} ... {{end}}
{{if .Condition}} ... {{else}} ... {{end}}
{{if .Condition1}} ... {{else if .Condition2}} ... {{else}} ... {{end}}
```

具体的な例を見てみましょう：

```go
package main

import (
    "log"
    "os"
    "text/template"
)

type Person struct {
    Name   string
    Age    int
    Active bool
}

const tmpl = `
Name: {{.Name}}
Age: {{.Age}}
{{if .Active}}
Status: Active
{{else}}
Status: Inactive
{{end}}
{{if ge .Age 18}}
{{if le .Age 60}}
 	Age group: Adult
{{else}}
	Age group: Senior
{{end}}
{{else}}
Age group: Minor
{{end}}
`

func main() {
    t, err := template.New("example").Parse(tmpl)
    if err != nil {
        log.Fatal(err)
    }

    person := Person{
        Name:   "Alice",
        Age:    30,
        Active: true,
    }

    err = t.Execute(os.Stdout, person)
    if err != nil {
        log.Fatal(err)
    }
}
```

このテンプレートでは、以下のようなif文を使用しています：

1. `.Active`の値に基づいて、ステータスを表示します。
2. `.Age`の値に基づいて、年齢グループを表示します。
   - `ge`は"greater than or equal to"（以上）を意味します。
   - `le`は"less than or equal to"（以下）を意味します。

テンプレート内では、以下のような比較演算子も使用できます：

- `eq`：等しい
- `ne`：等しくない
- `lt`：未満
- `le`：以下
- `gt`：より大きい
- `ge`：以上

また、論理演算子も使用できます：

- `and`：論理AND
- `or`：論理OR
- `not`：論理NOT

例：

```
{{if and .Condition1 .Condition2}} ... {{end}}
{{if or .Condition1 .Condition2}} ... {{end}}
{{if not .Condition}} ... {{end}}
```

これらの構文を使用することで、テンプレート内で条件に基づいて異なる出力を生成することができます。

## カスタム関数
- テンプレート内でカスタム関数を利用できる
- `template.Funcs()`メソッドを利用
- `<関数名> .<第１引数> .<第２引数>`のフォーマット
  - 以下の例だと`contain`が実際の関数(contains)と紐づいている関数名、`.FruitColors`が第１引数、`.Fruit`が第２引数
- 例  
  ```go
  package main

  import (
      "log"
      "os"
      "text/template"
  )

  type Data struct {
      Name        string
      Fruit       string
      FruitColors []string
  }

  const tmpl = `
  Name: {{.Name}}
  Fruit: {{.Fruit}}
  {{if contain .FruitColors .Fruit}}{{.Fruit}} はリストにあります。
  {{else}}{{.Fruit}} はリストにありません。
  {{end}}
  `

  func contains(slice []string, item string) bool {
      for _, s := range slice {
          if s == item {
              return true
          }
      }
      return false
  }

  func main() {
      funcMap := template.FuncMap{
          "contain": contains,
      }

      t, err := template.New("example").Funcs(funcMap).Parse(tmpl)
      if err != nil {
          log.Fatal(err)
      }

      data := Data{
          Name:        "Alice",
          Fruit:       "banana",
          FruitColors: []string{"apple", "banana", "cherry"},
      }

      err = t.Execute(os.Stdout, data)
      if err != nil {
          log.Fatal(err)
      }
  }
  ```

## `template.New`と`Execute`メソッドについて
`template.New`と`Execute`メソッドは、それぞれテンプレートの作成と実行という重要な役割を果たします。それぞれの詳細を説明します：

1. `template.New`:

`template.New`は新しいテンプレートを作成します。具体的には：

- 新しい`*template.Template`オブジェクトを生成します。
- このオブジェクトに名前を付けます（引数として与えられた文字列）。
- 生成されたテンプレートオブジェクトへのポインタを返します。

例：
```go
t := template.New("myTemplate")
```

この時点では、テンプレートの内容はまだ設定されていません。通常、この後に`Parse`メソッドを呼び出してテンプレートの内容を設定します。

2. `Execute`メソッド:

`Execute`メソッドは、パースされたテンプレートを実行し、結果を出力します。具体的には：

- テンプレートの内容を処理します。
- テンプレート内の変数やアクションを、提供されたデータ（通常は構造体）で置換します。
- 処理結果を指定された出力先（例：`io.Writer`インターフェースを実装したオブジェクト）に書き込みます。

例：
```go
err := t.Execute(os.Stdout, data)
```

`Execute`メソッドの動作：

1. テンプレート内の`{{...}}`で囲まれた部分を探します。
2. これらの部分を評価し、必要に応じてデータを挿入します。
3. 条件文（`if`）、ループ（`range`）などの制御構造を処理します。
4. 結果を指定された出力先に書き込みます。

全体の流れを簡単にまとめると：

1. `template.New`: テンプレートオブジェクトを作成
2. `Parse`: テンプレートの内容を設定
3. `Execute`: テンプレートを実行し、結果を出力

これらのステップにより、動的なテキスト生成が可能になります。テンプレートは一度作成・パースされれば、異なるデータで何度でも実行できます。

## テンプレートのコード内利用
- `Execute`メソッドの第１引数に変換されたテンプレートの格納先を指定
- `os.Stdout`を指定すると標準出力するだけ
- コード内で結果を保存して使うためには`bytes.Buffer`型に格納し、`String()`メソッドで文字列に変換する必要がある  
  ```go
	// Templateの結果を格納する変数を定義
	var buf bytes.Buffer
	err = t.Execute(&buf, OpenSearchTmpldata)
	if err != nil {
		return nil, err
	}

	// Templateの結果を文字列に変換
	opensearchYaml := buf.String()
  ```

## 改行について
- if文(endやelseなども含む)の部分も１行としてカウントされ、テンプレートに空行が入ってしまうため、if文などの条件分の前に`-`を付けて、空白（スペース、タブ、改行）を削除すること
  - `{{-`は、その直前の空白（スペース、タブ、改行）を削除、  
    `-}}`は、その直後の空白を削除
- `-`を付けない場合  
  ```go
  const tmpl = `
  Name: {{.Name}}
  Age: {{.Age}}
  {{if .Active}}
  Status: Active
  {{else}}
  Status: Inactive
  {{end}}
  {{if ge .Age 18}}
  {{if le .Age 60}}
    Age group: Adult
  {{else}}
    Age group: Senior
  {{end}}
  {{else}}
    Age group: Minor
  {{end}}
  `
  ```
  - 出力例  
    ```
    Name: Alice
    Age: 30

    Status: Active



      Age group: Adult
    ```
- `-`を付けた場合  
  ```go
  const tmpl = `
  Name: {{.Name}}
  Age: {{.Age}}
  {{- if .Active}}
  Status: Active
  {{- else}}
  Status: Inactive
  {{- end}}
  {{- if ge .Age 18}}
  {{- if le .Age 60}}
    Age group: Adult
  {{- else}}
    Age group: Senior
  {{- end}}
  {{- else}}
    Age group: Minor
  {{- end}}
  `
  ```  
  - 出力例  
    ```
    Name: Alice
    Age: 30
    Status: Active
      Age group: Adult
    ```