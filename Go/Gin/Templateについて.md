## Baseテンプレートの定義と使い分け
- https://gin-gonic.com/docs/examples/html-rendering/
- 参照先となるテンプレート(`*.tmpl`)にて`{{ define "<任意の名前>" }}`でベースを定義し、参照元のテンプレートにて`{{ template "<template名>" <参照先テンプレートファイル階層> }}`でインポートしたうえで追加の情報を上書きする
- 例（`index.tmpl`で`header.tmpl`をベースにし、独自の追加情報(以下の例だと"Hello")を上書き）
  - `header.tmpl`  
    ```html
    {{ define "header" }}
    <div class="logo" style="background-color: #D2C7AB; color: white;">
        <img src="/static/images/santa.png" alt="logo">
        <h1>TechCareer Talk</h1>
        <nav>
            <a href="/" style="color: white;">Documentation</a>
            <a href="/blog" style="color: white; margin-left: 10px;">Blog</a>
            <a href="/english" style="color: white; margin-left: 10px;">English</a>
            <input type="text" placeholder="Search this site..." style="margin-left: 10px;">
        </nav>
    </div>
    {{ end }}
    ```
  - `base.tmpl`  
    ```html
    {{ define "top" }}
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>TechCareer Talk</title>
      <link rel="icon" type="image/x-icon" href="/static/favicon.ico">
      <link rel="stylesheet" href="/static/css/style.css">
    </head>
    <body>
      <header>
        {{ template "header" . }}
      </header>
    {{ end }}
    {{ define "bottom" }}
    </body>
    </html>
    {{ end }}
    ```
  - `index.tmpl`  
    ```html
    {{ template "top" . }}
    <div class="container mt-5">
        Hello
    </div>
    {{ template "bottom" . }}
    ```

## template内の変数の定義と値の渡し方
- templateに値を渡すには、`gin.H`または`map[string]interface{}`を使う
- template内で渡された変数を使うには、`{{ .variableName }}`の形式を使う
- 例  
  ```go
  c.HTML(http.StatusOK, "template.html", gin.H{
    "title": "My Page",
    "user":  user,
  })
  ```

  ```html
  <!DOCTYPE html>
  <html>
  <head>
      <title>{{ .title }}</title>
  </head>
  <body>
      <h1>Hello, {{ .user.Name }}</h1>
  </body>
  </html>
  ```

## template内のif文
- https://it.noknow.info/ja/article/go/how-to-use-if-in-html-template
- フォーマット  
  ```html
  {{ if .condition }}
      <!-- 条件が真の場合 -->
  {{ else }}
      <!-- 条件が偽の場合 -->
  {{ end }}
  ```

Gin フレームワークでの Go テンプレートの使用方法について説明します。

1. 変数定義と値の渡し方:

テンプレートに値を渡すには、`gin.H` または `map[string]interface{}` を使用します。

```go
c.HTML(http.StatusOK, "template.html", gin.H{
    "title": "My Page",
    "user":  user,
})
```

テンプレート内では、これらの変数に次のようにアクセスできます：

```html
<h1>{{ .title }}</h1>
<p>Welcome, {{ .user.Name }}</p>
```

2. テンプレート内の if 文:

Go テンプレート言語では、以下のように if 文を使用できます：

```html
{{ if .condition }}
    <!-- 条件が真の場合 -->
{{ else if .condition }}
    <!-- 条件が真の場合 -->
{{ else }}
    <!-- 条件が偽の場合 -->
{{ end }}
```

### 比較演算子
- 等しい (==):
  ```html
  {{ if eq .value 5 }}Value is 5{{ end }}
  ```

- 等しくない (!=):
  ```html
  {{ if ne .value 5 }}Value is not 5{{ end }}
  ```

- より大きい (>):
  ```html
  {{ if gt .value 5 }}Value is greater than 5{{ end }}
  ```

- 以上 (>=):
  ```html
  {{ if ge .value 5 }}Value is greater than or equal to 5{{ end }}
  ```

- より小さい (<):
  ```html
  {{ if lt .value 5 }}Value is less than 5{{ end }}
  ```

- 以下 (<=):
  ```html
  {{ if le .value 5 }}Value is less than or equal to 5{{ end }}
  ```

### 論理演算子
- AND:
  ```html
  {{ if and .condition1 .condition2 }}Both conditions are true{{ end }}
  ```

- OR:
  ```html
  {{ if or .condition1 .condition2 }}At least one condition is true{{ end }}
  ```

- NOT:
  ```html
  {{ if not .condition }}Condition is false{{ end }}
  ```

## template内のrange文
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{ .title }}</title>
</head>
<body>
    <h1>Hello, {{ .name }}</h1>
    <ul>
    {{ range .items }}
        <li>{{ . }}</li>
    {{ end }}
    </ul>
</body>
</html>
```
### range文の注意事項
##### range は主にスライス、配列、マップ、チャネルなどの反復可能なオブジェクトに対して使用される。単一の整数値に対しては使用できない。
- NG例  
  - go側
    ```go
    c.HTML(http.StatusOK, "index.tpl", gin.H{
        "pageTotal":  3,
    })
    ```
  - template側
  ```tpl
  {{ range .pageTotal }}
  <span>{{ . }}</span>
  {{ end }}
  ```

- OK例  
  - go側
    ```go
    c.HTML(http.StatusOK, "index.tpl", gin.H{
        "pageTotal":  []int{1,2,3},
    })
    ```
  - template側
  ```tpl
  {{ range .pageTotal }}
  <span>{{ . }}</span>
  {{ end }}
  ```

##### テンプレート内での変数のスコープと参照方法
range内の`.`の参照先と親コンテキストを参照する方法

1. 親のコンテキスト  
親のコンテキストとは、テンプレートに最初に渡されたデータ全体を指す。Ginの場合、`gin.H{}`で渡された全てのデータがそれに当たる。

2. `.`（ドット）の意味  
テンプレート内で`.`は現在のコンテキストを表す。トップレベルでは、これは親のコンテキスト（渡された全データ）を指す。

3. `range`ループでのコンテキストの変更  
`range`ループに入ると、`.`の参照先が変更される。ループ内では、`.`は現在処理中の要素を指すようになる。

4. `$`を使った親コンテキストへのアクセス  
`$`は常に親のコンテキストを参照するための特別な変数。ループ内でも親のコンテキストにアクセスしたい場合に使用する。

例：

```go
// Goコード
c.HTML(http.StatusOK, "template.tpl", gin.H{
    "BoardType": "career",
    "Items": []string{"A", "B", "C"},
})
```

```html
<!-- テンプレート -->
<p>Board Type: {{ .BoardType }}</p>
<ul>
{{ range .Items }}
    <li>Item: {{ . }}, Board: {{ $.BoardType }}</li>
{{ end }}
</ul>
```

この例では：
- トップレベルの`{{ .BoardType }}`は"career"を出力する。
- `range`ループ内の`{{ . }}`は現在の項目（"A"、"B"、"C"）を参照する。
- `{{ $.BoardType }}`は親コンテキストの"BoardType"（"career"）を参照する。

出力結果：
```
Board Type: career
- Item: A, Board: career
- Item: B, Board: career
- Item: C, Board: career
```

`$`を使うことで、どんなに深いネスト(e.g. ループ)の中でも最上位のコンテキストにアクセスすることができる。