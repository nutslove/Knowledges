## 比較演算子
- 等しいか等しくないかを確認する時は基本`===`と`!==`を使うこと
  - `==`と`!=`は型の比較まではしてくれないため
- `==`と`===`、`!=`と`!==`の違い
  - `==`と`!=`
    - 値の比較だけして、型の比較はしない
  - `===`と`!==`
    - 値の比較だけではなく、型の比較もしてくれる

## `document.getElementById(<ID名>)`について
- `document.getElementById('id')`は、HTMLドキュメント内の指定した`id`属性を持つ要素を1つ取得するメソッド
- 取得した要素に対して、内容の変更・スタイルの変更などJavaScriptによる操作ができる
- `id`は文書内で一意である前提のため、常に単一の要素（見つからなければ`null`）を返す
- 例：以下の`<p>`要素があるとする
  ```html
  <p id="example">これはテストです。</p>
  ```
  この要素のテキストを変更する
  ```javascript
  // 'example'というIDを持つ要素を取得
  const element = document.getElementById('example');

  // 要素のテキスト内容を変更
  element.textContent = '内容が変更されました！';
  ```
  実行すると「これはテストです。」が「内容が変更されました！」に変わる
- 関連：CSSセレクタで柔軟に要素を取得したい場合は`document.querySelector('#example')`（最初の1つ）や`document.querySelectorAll('.item')`（複数）も使える

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
      let query = document.getElementsByName("query")[0].value;
      if (query == null || query.trim() == "") {
        alert("検索語を入力してください");
        return false;
      }
      return true;
    }
    </script>
    ```


## `innerHTML`・`innerText`・`textContent`について
- 要素の内容を取得・設定するプロパティ。HTMLタグの扱いが異なる
- **`innerHTML`**：HTMLマークアップ（タグ）も含めた内容を取得・設定する
  - 例：内容が`<b>bold</b>`なら`innerHTML`は`"<b>bold</b>"`を返す
  - **⚠️ セキュリティ注意**：ユーザー入力など信頼できない文字列を`innerHTML`に代入するとXSS（クロスサイトスクリプティング）の脆弱性になる。テキストを入れるだけなら`textContent`を使う
- **`innerText`**：ブラウザ上で「見える」テキストだけを取得・設定する
  - CSSで非表示（`display:none`など）の部分は含まれない。表示状態を考慮するため、取得時にレイアウト計算が走り比較的重い
- **`textContent`**：要素内の全テキストを取得・設定する（タグは除くが、非表示要素のテキストも含む）
  - 表示状態を考慮しないので`innerText`より高速。単純にテキストを入れ替える用途ではこれが基本
- 例：次のHTML要素があるとする
  ```html
  <p id="example">これは<b>テスト</b>です。</p>
  ```
  ```javascript
  const element = document.getElementById('example');

  console.log(element.innerHTML);    // "これは<b>テスト</b>です。"（タグ込み）
  console.log(element.innerText);    // "これはテストです。"（見えるテキスト）
  console.log(element.textContent);  // "これはテストです。"（全テキスト）
  ```

## `window`と各メソッドについて
- `window`はブラウザウィンドウそのものを表すグローバルオブジェクト。ページ内のすべての要素（DOM要素・関数・変数など）は`window`の一部として存在する
- グローバルスコープで宣言したものは`window`のプロパティになるため、`window.`は省略できることが多い（`window.alert()` ≒ `alert()`）
- 主なメソッド・プロパティ

  1. **`window.onload`**：ページの全コンテンツ（画像・スクリプトなど）が読み込まれてから実行されるイベントハンドラ
     ```javascript
     window.onload = function() {
       alert("全てのコンテンツが読み込まれました！");
     };
     ```
     - ※現在はDOMの構築完了だけを待つ`DOMContentLoaded`（`document.addEventListener('DOMContentLoaded', ...)`）を使うことも多い。画像等の読み込みを待たない分、実行タイミングが早い

  2. **`window.location.href`**：現在のページのURLを取得・設定するプロパティ
     ```javascript
     console.log(window.location.href);  // 現在のページのURLを表示
     window.location.href = 'https://www.example.com';  // 指定URLへ遷移（リダイレクト）
     ```

  3. **`window.URL.createObjectURL`**：`Blob`/`File`オブジェクトを参照する一時URLを生成する。ダウンロードリンクや画像・動画のソースなどに使える
     ```javascript
     const blob = new Blob(["Hello, world!"], { type: 'text/plain' });
     const url = window.URL.createObjectURL(blob);
     console.log(url);  // "blob:https://example.com/xxxxxxxx-...."
     ```
     - ※不要になったら`URL.revokeObjectURL(url)`で解放しないとメモリリークになる

  4. **`window.addEventListener`**：指定イベント発生時に実行する関数（イベントハンドラ）を登録する
     ```javascript
     window.addEventListener('resize', function() {
       console.log('ウィンドウがリサイズされました！');
     });
     ```
     - `onload`のような`on〇〇`プロパティへの代入と違い、**同じイベントに複数のハンドラを登録できる**ため、こちらが推奨

- **⚠️ 廃止された`window.navigator.msSaveBlob`について**
  - かつてInternet Explorer / 旧Edge（EdgeHTML版）で`Blob`/`File`をローカル保存するために提供されていた非標準メソッド
  - **現在のChromium版Edgeを含む最新ブラウザでは削除されており使えない（`undefined`）**。使ってはいけない
  - 代替：`URL.createObjectURL()`で生成したURLを`<a download>`要素に設定してプログラムからクリックする方法が標準的
    ```javascript
    const blob = new Blob(["Hello"], { type: 'text/plain' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'hello.txt';
    a.click();
    URL.revokeObjectURL(a.href);
    ```

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

---

# JavaScript基礎文法まとめ

## 変数宣言（`var`・`let`・`const`）
- **`const`を基本とし、再代入が必要な場合のみ`let`を使う。`var`は原則使わない**
- 3つの違い

  | 宣言 | 再代入 | 再宣言 | スコープ | 巻き上げ(hoisting) |
  |------|--------|--------|----------|----------------------|
  | `var`   | 可 | 可 | 関数スコープ | される（`undefined`で初期化） |
  | `let`   | 可 | 不可 | ブロックスコープ | されるがTDZ（後述）で参照不可 |
  | `const` | 不可 | 不可 | ブロックスコープ | されるがTDZで参照不可 |

- **関数スコープ vs ブロックスコープ**
  - `var`は関数単位でしか変数を閉じ込められない（`if`や`for`の`{}`を無視する）
  - `let`/`const`は`{}`（ブロック）単位で変数が有効
  ```javascript
  if (true) {
    var a = 1;
    let b = 2;
  }
  console.log(a); // 1（ブロック外でも参照できてしまう）
  console.log(b); // ReferenceError（ブロック外では参照不可）
  ```
- **`const`は宣言と同時に必ず初期値を代入しないといけない**
  - 再代入できない変数なので、後から値を入れるチャンスが無い → 宣言時に値を確定させないと構文エラーになる
  ```javascript
  const a = 1;   // OK
  // const b;    // SyntaxError: Missing initializer in const declaration
  // b = 2;

  // let / var は初期値なしで宣言でき、後から代入できる
  let x;         // OK（この時点では undefined）
  x = 10;        // 後から代入OK
  ```
  - 「宣言時点では値が決まらないが、一度決めたら変えたくない」場合は、条件を式にして`const`に入れる
  ```javascript
  // letで書く
  let status1;
  if (score >= 60) {
    status1 = "合格";
  } else {
    status1 = "不合格";
  }

  // constで書く（三項演算子を使う）
  const status2 = score >= 60 ? "合格" : "不合格";
  ```
- **`const`は「再代入禁止」であって「不変(immutable)」ではない**
  - オブジェクトや配列の中身は変更できる
  ```javascript
  const arr = [1, 2];
  arr.push(3);      // OK（中身の変更は可能）
  // arr = [4, 5];  // Error（再代入は不可）
  ```
- **TDZ（Temporal Dead Zone / 一時的なデッドゾーン）**
  - `let`/`const`は宣言前に参照するとエラーになる（`var`は`undefined`になるだけ）
  ```javascript
  console.log(x); // ReferenceError（TDZ）
  let x = 10;
  ```

## データ型
- **プリミティブ型（7種類）**：`string` / `number` / `bigint` / `boolean` / `undefined` / `null` / `symbol`
  - プリミティブは**値そのものが渡される（値渡し）**
- **オブジェクト型**：`object`（配列・関数・`Object`など全て）
  - オブジェクトは**参照が渡される（参照渡し）**
- `typeof`で型を確認できる
  ```javascript
  typeof "hello"   // "string"
  typeof 123       // "number"
  typeof true      // "boolean"
  typeof undefined // "undefined"
  typeof null      // "object" ← JSの有名なバグ。nullだがobjectと返る
  typeof {}        // "object"
  typeof []        // "object" ← 配列もobject。Array.isArray()で判定する
  typeof function(){} // "function"
  ```
- **`undefined`と`null`の違い**
  - `undefined`：値が代入されていない（システムが自動的に設定する）
  - `null`：意図的に「空・なし」を表す（開発者が明示的に設定する）

## 文字列（String）
- シングルクォート`'`、ダブルクォート`"`、テンプレートリテラル`` ` ``のいずれも使える
- **テンプレートリテラル（バッククォート）**
  - `${}`で変数や式を埋め込める。複数行もそのまま書ける
  ```javascript
  const name = "太郎";
  const age = 25;
  const msg = `${name}さんは${age}歳です。
  改行もそのまま書けます。`;
  ```
- よく使う文字列メソッド
  ```javascript
  "Hello".length            // 5
  "Hello".toUpperCase()     // "HELLO"
  "Hello".toLowerCase()     // "hello"
  "  hi  ".trim()           // "hi"（前後の空白削除）
  "a,b,c".split(",")        // ["a", "b", "c"]
  "Hello".includes("ell")   // true
  "Hello".startsWith("He")  // true
  "Hello".endsWith("lo")    // true
  "Hello".replace("l", "L") // "HeLlo"（最初の1つだけ）
  "Hello".replaceAll("l","L")// "HeLLo"（全て）
  "Hello".indexOf("l")      // 2（見つからなければ-1）
  "Hello".slice(1, 3)       // "el"
  "Hello".charAt(0)         // "H"
  "5".padStart(3, "0")      // "005"
  ```

## 数値（Number）とMathオブジェクト
```javascript
Number("123")       // 123（文字列→数値変換）
parseInt("123px")   // 123（整数として解釈）
parseFloat("1.5em") // 1.5
(3.14159).toFixed(2)// "3.14"（小数点以下2桁、文字列で返る）
Number.isNaN(NaN)   // true
Number.isInteger(5) // true

Math.floor(3.7)  // 3（切り捨て）
Math.ceil(3.2)   // 4（切り上げ）
Math.round(3.5)  // 4（四捨五入）
Math.max(1,2,3)  // 3
Math.min(1,2,3)  // 1
Math.abs(-5)     // 5
Math.random()    // 0以上1未満の乱数
Math.pow(2, 3)   // 8（2の3乗）。 2 ** 3 でも可
Math.sqrt(9)     // 3
```
- **`NaN`（Not a Number）**：数値変換に失敗した時などに返る特殊な値。`NaN === NaN`は`false`なので判定は`Number.isNaN()`を使う

## 演算子
- **算術演算子**：`+` `-` `*` `/` `%`（剰余） `**`（べき乗）
- **論理演算子**：`&&`（AND） `||`（OR） `!`（NOT）
- **Nullish coalescing（`??`）**
  - **左辺が`null`または`undefined`の時だけ右辺を返す**（`||`との違いに注意）
  ```javascript
  0 || "default"   // "default"（0はfalsyなので右辺）
  0 ?? "default"   // 0（0はnull/undefinedではないので左辺）
  null ?? "default"// "default"
  ```
- **Optional chaining（`?.`）**
  - プロパティが`null`/`undefined`の場合にエラーにせず`undefined`を返す
  ```javascript
  const user = { profile: { name: "太郎" } };
  user.profile?.name       // "太郎"
  user.account?.id         // undefined（エラーにならない）
  user.getName?.()         // メソッドが無ければundefined
  ```
- **三項演算子**：`条件 ? 真の時 : 偽の時`
  ```javascript
  const result = age >= 20 ? "成人" : "未成年";
  ```
- **falsyな値**（`if`などで`false`扱いになる値）：`false` / `0` / `""`（空文字） / `null` / `undefined` / `NaN`
  - これ以外は全て`truthy`（`"0"`、`[]`、`{}`もtruthy）

## 条件分岐
```javascript
// if / else if / else
if (score >= 80) {
  console.log("A");
} else if (score >= 60) {
  console.log("B");
} else {
  console.log("C");
}

// switch（breakを忘れると次のcaseに流れる=フォールスルー）
switch (fruit) {
  case "apple":
    console.log("りんご");
    break;
  case "banana":
    console.log("バナナ");
    break;
  default:
    console.log("その他");
}
```

## 繰り返し（ループ）
```javascript
// for
for (let i = 0; i < 5; i++) { console.log(i); }

// for...of（配列などの「値」を回す）
for (const item of ["a", "b", "c"]) { console.log(item); }

// for...in（オブジェクトの「キー」を回す。配列には非推奨）
for (const key in {x:1, y:2}) { console.log(key); } // "x", "y"

// while
let n = 0;
while (n < 3) { n++; }

// do...while（最低1回は実行される）
do { console.log("once"); } while (false);
```
- `break`：ループを抜ける / `continue`：その回だけスキップして次へ

## 関数
- **関数宣言（function declaration）**：巻き上げされるので定義前に呼び出せる
  ```javascript
  function add(a, b) {
    return a + b;
  }
  ```
- **関数式（function expression）**：変数に代入。巻き上げされない
  ```javascript
  const add = function(a, b) {
    return a + b;
  };
  ```
- **アロー関数（arrow function）**：短く書ける。`this`を束縛しない（後述）
  ```javascript
  const add = (a, b) => a + b;      // 1行なら{}とreturn省略可
  const square = x => x * x;         // 引数1つなら()も省略可
  const greet = () => { console.log("hi"); }; // 引数なしは()必須
  const makeObj = () => ({ a: 1 });  // オブジェクトを返す時は()で囲む
  ```
- **デフォルト引数**
  ```javascript
  function greet(name = "ゲスト") { return `こんにちは、${name}`; }
  greet();       // "こんにちは、ゲスト"
  ```
- **残余引数（rest parameters）**：可変長の引数を配列で受け取る
  ```javascript
  function sum(...nums) { return nums.reduce((a, b) => a + b, 0); }
  sum(1, 2, 3, 4); // 10
  ```

## スコープ・クロージャ・巻き上げ
- **クロージャ（closure）**：関数が定義された時のスコープの変数を、関数の外に出た後も覚えている仕組み
  ```javascript
  function counter() {
    let count = 0;
    return function() {
      count++;
      return count;
    };
  }
  const c = counter();
  c(); // 1
  c(); // 2（countが保持され続けている）
  ```
- **巻き上げ（hoisting）とは**
  - JavaScriptは、コードを実行する前にそのスコープ内の**宣言だけを先に読み取る**という動きをする。その結果、**変数や関数が「宣言より前の行」でも認識されている**ように見える挙動を巻き上げ（hoisting）と呼ぶ
  - イメージ：「宣言部分がスコープの先頭に自動で持ち上げられる」と考えると分かりやすい（実際にコードが移動するわけではない）
  - **重要なのは「宣言」だけが巻き上げられ、「代入（値）」は元の行のまま**という点

  - **① `function`宣言：中身ごと巻き上げられる → 定義より前で呼べる**
    ```javascript
    greet();  // "こんにちは" ← 定義より前なのに呼べる

    function greet() {
      console.log("こんにちは");
    }
    ```

  - **② `var`：宣言だけ巻き上げられ、値は`undefined`になる**
    - 「宣言前でもエラーにはならないが、値はまだ入っていない」状態になり、バグの原因になりやすい
    ```javascript
    console.log(x);  // undefined ← エラーにはならないが値は未定義
    var x = 10;
    console.log(x);  // 10

    // JavaScriptが内部的にこう解釈しているイメージ ↓
    // var x;           ← 宣言だけ先頭に持ち上げられる
    // console.log(x);  → undefined
    // x = 10;          ← 代入は元の位置のまま
    ```

  - **③ `let`/`const`：巻き上げはされるが、宣言前に使うとエラー（TDZ）**
    - `var`と違い`undefined`にはならず、`ReferenceError`になる。この「宣言前は触れない区間」がTDZ（前述）
    ```javascript
    console.log(y);  // ReferenceError ← varと違いエラーになる
    let y = 20;
    ```
    - この挙動により`let`/`const`は「宣言前にうっかり使う」バグを防いでくれる。`var`を避けるべき理由の一つ

  - **関数式・アロー関数は巻き上げされない**（中身が巻き上げられるのは`function`宣言だけ）
    ```javascript
    add(1, 2);                    // Error（addはまだ undefined / TDZ）
    const add = (a, b) => a + b;
    ```

## `this`
- **`this`は「関数がどう呼ばれたか」で決まる**（定義場所ではない）
  - 通常の関数：呼び出し方によって変わる（メソッド呼び出しならそのオブジェクト、単独呼び出しなら`undefined`(strict)またはグローバル）
  - **アロー関数：自身の`this`を持たず、外側のスコープの`this`を引き継ぐ**
  ```javascript
  const obj = {
    name: "太郎",
    normalFn: function() { return this.name; }, // "太郎"
    arrowFn: () => this.name,                    // 外側のthis（objではない）
  };
  ```
- コールバック内で`this`を維持したい場合はアロー関数が便利

## オブジェクト
```javascript
const user = {
  name: "太郎",
  age: 25,
  greet() { return `${this.name}です`; }, // メソッドの短縮記法
};

user.name          // "太郎"（ドット記法）
user["age"]        // 25（ブラケット記法。動的なキーに使う）
user.email = "x@y" // プロパティ追加
delete user.age    // プロパティ削除

// よく使うObjectメソッド
Object.keys(user)    // ["name", "greet", "email"]（キーの配列）
Object.values(user)  // 値の配列
Object.entries(user) // [["name","太郎"], ...] キーと値のペア配列
Object.assign({}, user)      // 浅いコピー
{ ...user, age: 30 }         // スプレッドでコピー＋上書き
```
- **プロパティの短縮記法**：変数名とキー名が同じなら省略できる
  ```javascript
  const name = "太郎";
  const obj = { name }; // { name: "太郎" } と同じ
  ```

## 配列（Array）
```javascript
const arr = [1, 2, 3, 4, 5];
arr.length          // 5
arr.push(6)         // 末尾に追加 → [1,2,3,4,5,6]
arr.pop()           // 末尾を削除して返す
arr.unshift(0)      // 先頭に追加
arr.shift()         // 先頭を削除して返す
arr.slice(1, 3)     // [2,3] 元配列は変更しない（非破壊）
arr.splice(1, 2)    // index1から2個削除（破壊的）
arr.indexOf(3)      // 2
arr.includes(3)     // true
arr.concat([6,7])   // 結合
arr.join("-")       // "1-2-3-4-5"（文字列化）
arr.reverse()       // 反転（破壊的）
[3,1,2].sort()      // [1,2,3]（デフォルトは文字列比較に注意）
[3,1,2].sort((a,b) => a - b) // 数値の昇順
```
- **高階関数（配列の反復メソッド）** ※`for`より頻用
  ```javascript
  // map：各要素を変換して新しい配列を返す
  [1,2,3].map(x => x * 2)          // [2,4,6]

  // filter：条件に合う要素だけの新しい配列
  [1,2,3,4].filter(x => x % 2===0) // [2,4]

  // reduce：畳み込み（合計など）。第2引数は初期値
  [1,2,3,4].reduce((acc, x) => acc + x, 0) // 10

  // forEach：各要素に処理（戻り値なし）
  [1,2,3].forEach(x => console.log(x));

  // find：条件に合う最初の「要素」を返す
  [1,2,3].find(x => x > 1)         // 2

  // findIndex：条件に合う最初のindex
  [1,2,3].findIndex(x => x > 1)    // 1

  // some：1つでも条件を満たせばtrue
  [1,2,3].some(x => x > 2)         // true

  // every：全て条件を満たせばtrue
  [1,2,3].every(x => x > 0)        // true
  ```
- `map`/`filter`/`slice`などは**非破壊**（新しい配列を返す）、`push`/`splice`/`sort`/`reverse`などは**破壊的**（元を変更）

## 分割代入（destructuring）
```javascript
// 配列の分割代入
const [a, b, c] = [1, 2, 3]; // a=1, b=2, c=3
const [first, ...rest] = [1, 2, 3, 4]; // first=1, rest=[2,3,4]

// オブジェクトの分割代入
const { name, age } = { name: "太郎", age: 25 };
const { name: n } = user; // 別名を付ける（n=user.name）
const { city = "東京" } = user; // デフォルト値

// 関数の引数で分割代入
function greet({ name, age }) { return `${name}(${age})`; }
```

## スプレッド構文（`...`）
```javascript
// 配列のコピー・結合
const arr2 = [...arr];          // 浅いコピー
const merged = [...a, ...b];    // 結合

// オブジェクトのコピー・マージ
const obj2 = { ...obj };
const merged2 = { ...obj1, ...obj2 }; // 後ろが優先で上書き

// 関数呼び出しで配列を展開
Math.max(...[1, 2, 3]);         // 3
```
- ※スプレッドは**浅いコピー（shallow copy）**。ネストしたオブジェクト/配列は参照が共有される。深いコピーは`structuredClone(obj)`を使う

## 例外処理（try / catch / finally）
```javascript
try {
  throw new Error("エラー発生");
} catch (error) {
  console.error(error.message); // "エラー発生"
} finally {
  console.log("必ず実行される");
}
```

## 非同期処理
### Promise
- 非同期処理の結果（成功/失敗）を表すオブジェクト
```javascript
const promise = new Promise((resolve, reject) => {
  if (成功) resolve(値);
  else reject(new Error("失敗"));
});

promise
  .then(value => console.log(value))  // 成功時
  .catch(err => console.error(err))   // 失敗時
  .finally(() => console.log("完了")); // 成否問わず
```
- `Promise.all([...])`：全て成功で解決（1つでも失敗なら即reject）
- `Promise.race([...])`：最初に決着したもの
- `Promise.allSettled([...])`：全ての結果（成否問わず）を待つ

### async / await
- Promiseを同期的な見た目で書ける（`await`はPromiseの解決を待つ）
- `await`は`async`関数の中でのみ使える
```javascript
async function fetchData() {
  try {
    const res = await fetch("https://api.example.com/data");
    const data = await res.json();
    return data;
  } catch (error) {
    console.error(error);
  }
}
```

## モジュール（import / export）
```javascript
// export（named export：複数可）
export const PI = 3.14;
export function add(a, b) { return a + b; }

// export（default export：1ファイル1つ）
export default function main() {}

// import
import main, { PI, add } from "./module.js";
import * as utils from "./module.js"; // 全部まとめて
```

## クラス（class）
```javascript
class Animal {
  constructor(name) {
    this.name = name;       // インスタンスプロパティ
  }
  speak() {                 // メソッド
    return `${this.name}が鳴く`;
  }
  static create(name) {     // 静的メソッド（インスタンス不要）
    return new Animal(name);
  }
}

// 継承
class Dog extends Animal {
  constructor(name) {
    super(name);            // 親のconstructorを呼ぶ
  }
  speak() {                 // オーバーライド
    return `${this.name}がワンと鳴く`;
  }
}

const dog = new Dog("ポチ");
dog.speak(); // "ポチがワンと鳴く"
```

## 等価比較の補足（`==` と `===`）
- 冒頭の「比較演算子」参照。**基本は`===`/`!==`（厳密等価）を使う**
- `==`は暗黙の型変換が行われ直感に反する結果になりやすい
  ```javascript
  0 == ""        // true（型変換される）
  0 == "0"       // true
  null == undefined // true
  1 == "1"       // true
  1 === "1"      // false（型が違う）
  NaN === NaN    // false（NaNは何とも等しくない）
  ```

## `===`でもオブジェクトは参照比較になる点に注意
```javascript
{} === {}            // false（別々のオブジェクト）
[1] === [1]          // false
const a = {}; const b = a;
a === b              // true（同じ参照）
```