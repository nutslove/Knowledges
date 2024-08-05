##### `margin: 0 auto;`
- 以下と同じ意味で、固定幅の要素を水平方向に中央揃えする。  
  ```css
  margin-top: 0;
  margin-bottom: 0;
  margin-left: auto;
  margin-right: auto;
  ```

##### `transition`について
- `transition`はCSSのプロパティで、要素のスタイルが変化する際のアニメーション効果を制御する
- `transition`の基本的な構文  
  ```css
  transition: [property] [duration] [timing-function] [delay];
  ```
  - `property`: アニメーション化するプロパティ（例：width, color, all）
  - `duration`: アニメーションの持続時間（例：0.3s, 300ms）
  - `timing-function`: アニメーションの速度曲線（例：ease, linear, ease-in, ease-out）
  - `delay`: アニメーション開始までの遅延時間（オプション）
- 例 (要素にマウスを乗せると背景色が青から赤に0.3秒かけて滑らかに変化する)  
  ```css
  .element {
      background-color: blue;
      transition: all 0.3s ease;
  }

  .element:hover {
      background-color: red;
  }
  ```

##### `transform`について
- 要素の形状、位置、サイズを変更するためのプロパティ
- `:hover`と一緒に使われる(ことが多い)
- 主な変換関数
  - `translate()`: 要素を水平・垂直方向に移動
    - 例  
      ```css
      .sidebar a:hover {
          background-color: #DCDCDC;
          transform: translateX(2px) translateY(2px); /* 右下に少し動く */
      }
      ```
  - `scale()`: 要素のサイズを変更
    - 例  
      ```css
      .element:hover {
        transform: scale(1.1);  /* マウスオーバー時に10%拡大 */
      }
      ```
  - `rotate()`: 要素を回転
    - 例  
      ```css
      .element {
        transform: rotate(90deg);
      }
      ```
  - `skew()`: 要素を傾斜
    - 例  
      ```css
      .element {
        transform: skew(15deg, 15deg);
      }
      ```

##### `:hover`について
- ユーザーがマウスポインタを要素の上に置いた時（ホバー状態）にスタイルを適用するためのセレクタ
- 構文  
  ```css
  selector:hover {
    /* ホバー時のスタイル */
  }
  ```
- 例(1)  
  ```css
  a:hover {
    color: red;
    text-decoration: underline;
  }
  ```
- 例(2)  
  ```css
  .element:hover {
    transform: scale(1.1);  /* マウスオーバー時に10%拡大 */
  }
  ```

##### `table-layout: fixed;`
- テーブルのレイアウトアルゴリズムを固定モードに設定する。
- 最初の行のセル幅に基づいてカラム幅を決定し、その後のコンテンツの長さに関わらず幅を維持する。
- テーブルのレンダリング速度が向上し、幅の予測が容易になる。

##### `white-space: nowrap;`
- テキストの折り返しを防ぎ、一行で表示させる。
- スペースや改行が入っていても、テキストは横に伸び続ける。

##### `overflow: hidden;`
- 要素の境界を超えるコンテンツを非表示にする。
- コンテンツが要素のボックスに収まらない場合、はみ出た部分が切り取られる。

##### `text-overflow: ellipsis;`
- テキストが要素の幅を超える場合、省略記号（...）で表示する。
- `overflow: hidden;` と組み合わせて使用すると効果的

##### `border-collapse: collapse;`
- テーブルのセル（th, td）と外枠（table）の境界線を結合し、隣接するセル間の境界線を1本にまとめる
- デフォルトは `border-collapse: separate;`であり、`separate` ではセル間に小さな隙間ができ、境界線が二重に表示されることがあるが、`border-collapse: collapse;`でセル間の余分なスペースを除去し、よりコンパクトな見た目になる。
