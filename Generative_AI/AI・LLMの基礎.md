## ベクトル（Vector）
- 数値の配列やリストのこと
- 空間内の点や方向を表すために使用される数学的な概念
- ベクトルを通じて、複雑なデータや概念を数学的(定量的)に扱うことができる

### AIにおけるベクトルの使用例
1. **特徴ベクトル**
   - データの特徴を数値化したもの。たとえば、画像を表す際には、画像の各ピクセルの色情報がベクトルとして表される。

2. **単語ベクトル**
   - 自然言語処理において、単語や文章をベクトルとして表すことがある。この技術により、単語間の意味的な関係を計算できるようになる。
   - 例えば、文書をベクトル化する場合、各単語の出現回数などを要素とするベクトルを作成する。2つの文書が意味的に近いかどうかは、対応するベクトルの角度や距離を比較することで計算できる。

3. **状態ベクトル**
   - 機械学習のモデルが、問題を解くためのある時点での「状態」を数値で表したもの。

4. **重みベクトル**
   - 機械学習のアルゴリズムで、入力データに対してどの程度の重みを付けるかを示す数値のリスト。

## Token
- LLMモデルは(入力/出力)テキストをトークンという単位で分割して扱う
- 英語より日本語の方が１文字に必要なトークン数が多いといわれている
- OpenAIの場合、`tiktoken`というPythonのパッケージを使って入力/出力のトークン数を確認できる
  - https://github.com/openai/tiktoken