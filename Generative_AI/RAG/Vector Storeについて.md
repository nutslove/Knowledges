## Vector Storeから質問と関連するデータを検索するアルゴリズム
- https://opensearch.org/docs/latest/search-plugins/knn/index/
#### Approximate k-NN (k-Nearest Neighbors)
- 機械学習やデータベースの分野で使われるアルゴリズム。このアルゴリズムは、与えられたデータポイントに最も近い「k」個のデータポイントを見つけ出すことを目的としている。近似 k-NNは、完全な精度を犠牲にして計算の効率を向上させる方法で、大規模なデータセットでの使用に特に適している。
- デフォルトでは「**Approximate k-NN (k-Nearest Neighbors)**」アルゴリズムが使われる（OpenSearchのみ？？）
  - https://python.langchain.com/docs/integrations/vectorstores/opensearch  
    > similarity_search by default performs the Approximate k-NN Search which uses one of the several algorithms like lucene, nmslib, faiss recommended for large datasets. To perform brute force search we have other search methods known as Script Scoring and Painless Scripting. Check this for more details.

#### Script Score k-NN



#### Painless extensions



## Vector StoreとしてOpenSearch
- https://opensearch.org/platform/search/vector-database.html
- LangChainフォルダの「Retrievalについて.md」参照