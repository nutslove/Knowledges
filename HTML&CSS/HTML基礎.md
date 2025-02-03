# タグと属性
- `<>`で囲まれているのが**タグ**
  - 例: `<a href="https://www.yahoo.co.hp">ヤフー</a>`、
- `<>`の中で書いてあるのが**属性**（上の例だと`href=""`が属性）

## タグ
#### ・`div`


#### ・`span`

#### ・`form`
- データを選択したり入力するためのタグ
- `placehold="<文字列>"`は

#### ・`meta`
- https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meta  
  > The `<meta>` HTML element represents metadata that cannot be represented by other HTML meta-related elements, like `<base>`, `<link>`, `<script>`, `<style>` or `<title>`.
- 例  
  ```html
  <!DOCTYPE html>
  <html lang="ja">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
  </head>
  ```
  - `name=viewport`  
    → ビューポート（ユーザーのデバイスの表示領域）に関する設定であることを示
  - `content="width=device-width"`  
    → ページの幅をデバイスの幅に合わせることを指定。例えば、スマートフォンやタブレットで表示する際に、ページがデバイスの画面幅に収まるように調整される。
  - `initial-scale=1.0`  
    → ページが初めてロードされたときのズームレベルを指定。`1.0`はデフォルトのズームレベルで、通常は100%ズームを意味