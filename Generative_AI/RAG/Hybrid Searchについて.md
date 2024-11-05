- 参考URL
  - https://python.langchain.com/docs/how_to/hybrid/
  - https://dalab.jp/archives/journal/hybrid-search/
  - https://qiita.com/isanakamishiro2/items/4eb175bb2bc80d7225cb
  - https://qiita.com/jw-automation/items/045917be7b558509fdf2#1-%E3%83%8F%E3%82%A4%E3%83%96%E3%83%AA%E3%83%83%E3%83%89%E3%82%B5%E3%83%BC%E3%83%81

## セマンティック検索（Semantic Search）とは
- クエリの単語に文字通り一致する内容ではなく、クエリの**意味**に一致する内容を返す検索方式
  - 語句の意味を解釈する検索エンジン技術のこと
- 参考URL
  - **https://www.elastic.co/jp/what-is/semantic-search**
  - https://boramorka.github.io/LLM-Book/CHAPTER-2/2.5%20Semantic%20Search.%20Advanced%20Retrieval%20Strategies/
  - https://aws.amazon.com/jp/blogs/news/knowledge-bases-for-amazon-bedrock-now-supports-hybrid-search/
  - https://aws.amazon.com/jp/blogs/news/amazon-opensearch-services-vector-database-capabilities-explained/
- **そもそも「ベクトル検索」が「セマンティック検索」の一種としてみなすことができる**
  - https://www.elastic.co/jp/what-is/vector-search  
    > ベクトル検索は、セマンティック検索や類似性検索に威力を発揮します。意味とコンテキストが埋め込み内に取り込まれるため、ベクトル検索ではキーワードの完全一致を必要とせずにユーザーの意味する内容を検索できます。テキストデータ（ドキュメント）、画像、音声の処理が可能です。クエリに類似または関連する製品を簡単かつ迅速に見つけることができます。
  - https://www.elastic.co/jp/what-is/semantic-search  
    > セマンティック検索は [ベクトル検索](https://www.elastic.co/jp/what-is/vector-search) を活用しており、コンテクストや検索意図の関連性に基づいて、コンテンツをランク付けして供給することができます。ベクトル検索は、検索可能な情報の詳細を関連する用語や項目のフィールド、あるいは複数のベクトルにエンコードし、次に各ベクトルを比較してどれが最もよく似ているかを判定します。
    > 
    > ベクトル検索対応のセマンティック検索では、クエリパイプラインの両端で同時に作動して結果を出します。クエリが発せられたら、検索エンジンはそのクエリを埋め込みに変換（ベクトル化）します。すなわちデータと関連するコンテクストの数値表現に変換するのです。この値はベクトルで保管されます。次に [kNNアルゴリズム（またはk近傍法）](https://www.elastic.co/jp/what-is/knn) を使って、既存文書（セマンティック検索が関与するテキスト）のベクトルを、クエリのベクトルと照合します。次にセマンティック検索は結果を生成し、概念的な関連性に基づいてその結果をランク付けします。
    > 
    > １. クエリが発せられたら、検索エンジンはそのクエリを埋め込みに変換（ベクトル化）します。すなわちデータと関連するコンテクストの数値表現に変換するのです。この値はベクトルで保管されます。
    >
    > ２. 次にkNNアルゴリズム（またはk近傍法）を使って、既存文書（セマンティック検索が関与するテキスト）のベクトルを、クエリのベクトルと照合します。
    >
    > ３. 次にセマンティック検索は結果を生成し、概念的な関連性に基づいてその結果をランク付けします。

## Hybrid Searchとは
- 基本的にはベクトル類似度により、ベクトルストアから検索（ベクトル検索）が行われるけど、ベクトル類似度検索と他の検索手法（全文検索、BM25、キーワード検索、セマンティック検索など）を組み合わせて、より高度な検索を行う方式を**Hybrid Search**という
  - ハイブリッド検索の方法はベクトルストアによって異なる
- 複数の検索方式によって取得したデータを組み合わせて最も適合性の高い結果を返す
- AWSのKnowledge BasesでのHybrid Searchは「キーワード検索」と「セマンティック検索」の組み合わせ
  - https://aws.amazon.com/jp/blogs/news/knowledge-bases-for-amazon-bedrock-now-supports-hybrid-search/