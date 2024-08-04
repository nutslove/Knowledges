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