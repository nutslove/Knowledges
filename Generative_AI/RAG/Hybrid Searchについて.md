- 参考URL
  - https://python.langchain.com/docs/how_to/hybrid/
  - https://dalab.jp/archives/journal/hybrid-search/
  - https://qiita.com/isanakamishiro2/items/4eb175bb2bc80d7225cb

## セマンティック検索（Semantic Search）とは
- クエリの単語に文字通り一致する内容ではなく、クエリの**意味**に一致する内容を返す検索方式
  - 語句の意味を解釈する検索エンジン技術のこと
- 参考URL
  - https://www.elastic.co/jp/what-is/semantic-search
  - https://boramorka.github.io/LLM-Book/CHAPTER-2/2.5%20Semantic%20Search.%20Advanced%20Retrieval%20Strategies/
  - https://aws.amazon.com/jp/blogs/news/knowledge-bases-for-amazon-bedrock-now-supports-hybrid-search/
  - https://aws.amazon.com/jp/blogs/news/amazon-opensearch-services-vector-database-capabilities-explained/

## Hybrid Searchとは
- 基本的にはベクトル類似度により、ベクトルストアから検索（ベクトル検索）が行われるけど、ベクトル類似度検索と他の検索手法（全文検索、BM25、キーワード検索、セマンティック検索など）を組み合わせて、より高度な検索を行う方式を**Hybrid Search**という
  - ハイブリッド検索の方法はベクトルストアによって異なる
- 複数の検索方式によって取得したデータを組み合わせて最も適合性の高い結果を返す
- AWSのKnowledge BasesでのHybrid Searchは「キーワード検索」と「セマンティック検索」の組み合わせ
  - https://aws.amazon.com/jp/blogs/news/knowledge-bases-for-amazon-bedrock-now-supports-hybrid-search/