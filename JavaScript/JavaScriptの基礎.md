## 比較演算子
- 等しいか等しくないかを確認する時は基本`===`と`!==`を使うこと
  - `==`と`!=`は型の比較まではしてくれないため
- `==`と`===`、`!=`と`!==`の違い
  - `==`と`!=`
    - 値の比較だけして、型の比較はしない
  - `===`と`!==`
    - 値の比較だけではなく、型の比較もしてくれる

## `document.getElementById(<ID名>)`について
- 以下Chat-GPTからの回答
> **document.getElementById**はJavaScriptで提供されるメソッドの一つで、HTMLドキュメント内の特定のID属性を持つ要素を探して返すことができます。これにより、特定のHTML要素に対してJavaScriptを用いた操作（例えば、その要素の内容を変更する、スタイルを変えるなど）を行うことが可能になります。
> 
> 使い方は非常にシンプルで、**document.getElementById('id')** という形で使用します。ここで、**'id'** は探す要素のIDを表します。
> 
> 以下に具体的な例を示します。HTMLファイルに以下のような要素があるとします：
>~~~html
> <p id="example">これはテストです。</p>
>~~~
> この`<p>`タグの内容をJavaScriptを使用して変更することができます。以下にその方法を示します：
>
>~~~javascript
> // 'example'というIDを持つ要素を取得
> var element = document.getElementById('example');
> 
> // 要素のテキスト内容を変更
> element.textContent = '内容が変更されました！';
>~~~
> このコードを実行すると、元々「これはテストです。」と表示されていた部分が「内容が変更されました！」に変わります。これは、**document.getElementById**が`<p id="example"\>`要素を取得し、その`textContent`プロパティを変更したためです。

## `document.getElementsByName(<nameの値>)`について
- HTML内の特定の`name`属性を持つ要素を取得し、操作する
- **`<form>`タグでフォーム送信時にjavascriptの特定の関数を実行できる`onsubmit`イベント属性を使った例**
  - 検索ボタンを押した時に`validateSearchForm()`関数が実行され、何も入力されてなかったりスペースだけ入力されているとアラートを返す
  - **`onsubmit`で指定した関数が`false`を返すとフォームの送信処理は中止され、`true`を返すか何も返さない場合はフォーム送信処理が続行される**
    ```html
                          ・
                          ・
    <form action="/search" method="get" onsubmit="return validateSearchForm()">
      <input type="text" name="query" placeholder="Search post...">
      <button type="submit">検索</button>
    </form>
                          ・
                          ・
    <script>
    function validateSearchForm() {
      var query = document.getElementsByName("query")[0].value;
      if (query == null || query.trim() == "") {
        alert("検索語を入力してください");
        return false;
      }
      return true;
    }
    </script>
    ```


## `document.getElementById(<ID名>).innerText`と`document.getElementById(<ID名>).innerHTML`について
- 以下Chat-GPTからの回答
> JavaScriptにおいて、`document.getElementById(<ID名>).innerHTML`と`document.getElementById(<ID名>).innerText`は特定のHTML要素の内容を取得または設定するために使用されます。これら二つのプロパティの主な違いは、HTMLタグをどのように扱うかです。
> 
> - `innerHTML`：これは要素のHTML内容を表します。つまり、要素内のHTMLマークアップ（タグなど）も含めた内容を取得または設定します。たとえば、要素の内容が`<b>bold</b>`である場合、`innerHTML`は`<b>bold</b>`を返します。
> 
> - `innerText`：これは要素の「見える」テキスト内容を表します。つまり、HTMLマークアップは無視され、人間がブラウザ上で見るテキストだけが取得または設定されます。上記の例で言えば、`innerText`は`bold`を返します。
> 
> 以下に具体的な例を示します。まず、次のHTML要素があるとします。
>
>```html
><p id="example">これは<b>テスト</b>です。</p>
>```
>
>次に、この要素の内容を`innerHTML`と`innerText`で取得し、その結果をコンソールに表示するJavaScriptコードを見てみましょう。
>
>```javascript
>var element = document.getElementById('example');
>
>console.log(element.innerHTML);  // "これは<b>テスト</b>です。"
>console.log(element.innerText);  // "これはテストです。"
>```
>
>このように、`innerHTML`はHTMLマークアップを含んだ内容を返し、`innerText`は見えるテキスト内容だけを返します。

## `window`と各メソッドについて
- 以下Chat-GPTからの回答
> JavaScriptにおける「window」は、ブラウザウィンドウそのものを表すグローバルオブジェクトであり、ウェブページのすべての要素（DOM要素、JavaScript関数、変数など）はこのオブジェクトの一部として存在します。つまり、`window`はウェブページ全体のコンテクストを提供し、そのページで利用可能な機能や情報へのアクセスを提供します。
>
> 次に、各メソッドとプロパティについて説明します。
>
> 1. `window.onload`：`window.onload`はイベントハンドラで、ページの全てのコンテンツ（画像やスクリプトファイルなど）が完全に読み込まれてから実行されます。以下にその使用例を示します：
>
>    ```javascript
>    window.onload = function() {
>      alert("全てのコンテンツが読み込まれました！");
>    };
>    ```
>
> 2. `window.location.href`：`window.location.href`は現在のページのURLを取得または設定するプロパティです。以下にその使用例を示します：
>
>    ```javascript
>    console.log(window.location.href);  // 現在のページのURLを表示
>
>    window.location.href = 'https://www.example.com';  // ページをwww.example.comにリダイレクト
>    ```
>
> 3. `window.navigator.msSaveBlob`：この関数はMicrosoft Internet Explorer及びEdgeの特定のバージョンで提供され、BlobまたはFileオブジェクトをローカルファイルとして保存するために使用されます。GPT-4の知識のカットオフ（2021年9月）時点では、このメソッドは非標準であり、非推奨となっています。
>
> 4. `window.URL.createObjectURL`：この関数はBlobオブジェクトを引数にとり、そのオブジェクトを参照するためのURLを作成します。このURLは、たとえば、ダウンロード可能なファイルのリンクや、画像やビデオのソースとして使用できます。以下にその使用例を示します：
>
>    ```javascript
>    var blob = new Blob(["Hello, world!"], { type: 'text/plain' });
>    var url = window.URL.createObjectURL(blob);
>
>    console.log(url);  // "blob:https://example.com/d41d8cd98f00b204e9800998ecf8427e"
>    ```
>
> 5. `window.addEventListener`：`window.addEventListener`は指定したイベントが発生したときに実行される関数（イベントハンドラ）を登録するメソッドです。以下にその使用例を示します：
>
>    ```javascript
>    window.addEventListener('resize', function() {
>      console.log('ウィンドウがリサイズされました！');
>    });
>    ```
> 上記のコードは、ウィンドウがリサイズされるたびにメッセージをコンソールに表示します。

## 配列と辞書について
### 配列
- 存在しないindexを参照してもエラーにはならず`undefined`が返ってくる

### 辞書
- 存在しないkeyを参照してもエラーにはならず`undefined`が返ってくる

## `confirm()`メソッド
- `confirm()`メソッドを使って"OK"と"キャンセル"ボタンを持つダイアログボックスを表示できて、それぞれのボタンを押した時の処理も実装可能
- 例  
  ```javascript
  let result = confirm("本当に続行しますか？");

  if (result) {
      // OKボタンが押された時の処理
      console.log("OKが押されました。処理を続行します。");
  } else {
      // キャンセルボタンが押された時の処理
      console.log("キャンセルされました。処理を中止します。");
  }
  ```