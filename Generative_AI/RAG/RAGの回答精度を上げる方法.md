#### 参考URL
- **https://aws.amazon.com/jp/blogs/news/a-practical-guide-to-improve-rag-systems-with-advanced-rag-on-aws/** （★）
- https://speakerdeck.com/naoki_0531/amazon-bedrock-amazon-aurorawozu-mihe-wasetaragdehui-da-jing-du-noxiang-shang-niqu-rizu-ndemita
- https://fintan.jp/page/10301/
- https://qiita.com/DeepMata/items/3c27394e4475b1b7e7ff

### ■ Semantic search（セマンティック検索）とは
- https://www.softbank.jp/biz/solutions/generative-ai/ai-glossary/semantic-search/
- 検索エンジンが単純なキーワードの一致検索だけではなく、ユーザの意図やクエリの意味を理解して、関連性の高い情報を検索するための技術
- その代表的な実装が **「ベクトル検索」** で、**テキストを埋め込みベクトルに変換し、コサイン類似度や内積などで近いものを探す。**

  #### Semantic search（セマンティック検索）の特徴
  1. 意味の理解
     - キーワードの単純な一致だけでなく、文脈や文の意味を理解します。これにより、ユーザーが意図した情報をより正確に把握することができます。
  2. 関連性の高い結果
     - 関連性の高い情報を提供するため、より適切な検索結果を表示することができます。キーワードの一致だけではなく、関連するトピックや意味的に類似した情報も考慮されます。
  3. 構造化データの利用
     - セマンティック検索では、構造化データ（例: RDF、OWL）やオントロジー（知識体系）を活用することがあります。これにより、データの意味的な関連性や階層構造を考慮して検索が行われます。
  4. 自然言語処理と機械学習の応用
     - セマンティック検索では、自然言語処理（NLP）や機械学習の技術が活用され、テキストの意味解析や文脈理解、関連性の推定が行われます。

  #### 検索例
  - 例えば、キーワード「ロンドンの天気」でセマンティック検索を行うと、従来のキーワードベースの検索では「ロンドンの天気」というキーワードの一致や近さに基づいて結果が表示されましたが、セマンティック検索では「ロンドンの現在の気象情報」や「ロンドンの天気予報」など、関連性の高い情報やより正確な情報を提供することが可能になります。

## 1. Parent Document Retriever
- 参考URL
  - https://python.langchain.com/docs/modules/data_connection/retrievers/parent_document_retriever
  - https://api.python.langchain.com/en/stable/retrievers/langchain.retrievers.parent_document_retriever.ParentDocumentRetriever.html
  - https://community.fullstackretrieval.com/index/parent-document-retriever
  - https://github.com/gkamradt/langchain-tutorials/blob/main/data_generation/Advanced%20Retrieval%20With%20LangChain.ipynb
  - https://qiita.com/shimajiroxyz/items/facf409b81f59bb68775
  - https://qiita.com/Naoki_Ishihara/items/9f1b852917de19141847
  - https://community.fullstackretrieval.com/index/parent-document-retriever
- > When splitting documents for retrieval, there are often conflicting desires:
  > 1. You may want to have small documents, so that their embeddings can most accurately reflect their meaning. If too long, then the embeddings can lose meaning.
  > 2. You want to have long enough documents that the context of each chunk is retained.
   > 
   > The `ParentDocumentRetriever` strikes that balance by splitting and storing small chunks of data. During retrieval, it first fetches the small chunks but then looks up the parent ids for those chunks and returns those larger documents.
   >
   > Note that “parent document” refers to the document that a small chunk originated from. This can either be the whole raw document OR a larger chunk.
- 背景
  - デフォルト(?)ではRetrievalは分割されているChunk(文章)を取得してそのままPromptに渡すけど、そのChunk(文章)に必要な情報がすべて含まれているとは限らない(文字数でChunkを分割しているため)。
  - 改行とかセッションなどを元に分割することもできるが、対象ドキュメントのフォーマットが統一されてないと何を基準に分割するかも難しい。だとしてChunkサイズを大きくすると１つのChunkにいろんな情報が入って、Embeddingモデルが質問と関連性の高いChunkを取得するのに影響してしまう。また、Embeddingモデルの最大トークン数で無限に大きくすることはできない。
- Parent Document Retrieverを使えばChunkサイズを小さくしておいて、質問と関連するデータ取得にはChunkを使って、LLMに投げるPromptに含めるデータにはChunkが含まれている元のドキュメントをすべて含めることで、回答に必要な完全なデータをPromptに含めることができる。
  - 元のドキュメントをすべて含めることもできるけど、元のドキュメントが大きすぎる場合は、類似度検索に使うChunkサイズを別々としておくこともできる
      - https://python.langchain.com/docs/modules/data_connection/retrievers/parent_document_retriever
  - 元のドキュメントを保存しておくStoreとして使えるストレージ一覧
    - https://python.langchain.com/docs/integrations/stores/
#### ■ 使い方
- (1)元のドキュメントをFullで取得してLLMに渡す方法、(2)類似度検索に使うChunk(小さいChunk → `child_splitter`パラメータでサイズ指定)とLLMに渡すChunk(大きいサイズ → `parent_splitter`パラメータでサイズ指定)
- `parent_splitter`パラメータを省略した場合は(1)の元のドキュメントをすべてLLMに渡す
- **(ベクトル化する)Chunkと原本のフルドキュメントを保存する**  
  ~~~python
  from langchain.storage import LocalFileStore
  from langchain.storage._lc_store import create_kv_docstore
  from langchain.retrievers import ParentDocumentRetriever
  from langchain.text_splitter import RecursiveCharacterTextSplitter
  from langchain_community.document_loaders import PyPDFLoader

  files_location = "/opt/rag/documents" ## ドキュメントファイルが格納されるディレクトリ
  fs = LocalFileStore(files_location)
  store = create_kv_docstore(fs)
  text_splitter = RecursiveCharacterTextSplitter(
      chunk_size=400,
      chunk_overlap=0,
  )

  parent_splitter = RecursiveCharacterTextSplitter(
      chunk_size=4000,
      chunk_overlap=200,
  )

  loaders = PyPDFLoader("/<some_dir>/some.pdf")
  docs = loader.load()

  embeddings = BedrockEmbeddings(
      model_id = "cohere.embed-multilingual-v3"
  )

  db = OpenSearchVectorSearch(
      index_name="some_index",
      embedding_function=embeddings,
      opensearch_url="https://opensearch:9200",
      http_auth=("admin", "admin"),
      use_ssl = False,
      verify_certs = False,
      ssl_assert_hostname = False,
      ssl_show_warn = False,
  )

  ## Promptに参照元ドキュメント全文を入れる方法（parent_splitterパラメータを省略）
  retriever = ParentDocumentRetriever(
      vectorstore=db,
      docstore=store,  ## 元のドキュメントを保存する場所
      child_splitter=text_splitter,
  )

  ## Promptに全文ではなく、parent_splitterで指定した大きいChunkを入れる方法（parent_splitterパラメータを指定）
  retriever = ParentDocumentRetriever(
      vectorstore=db,
      docstore=store,  ## 元のドキュメントを保存する場所
      child_splitter=text_splitter,
      parent_splitter=parent_splitter,
  )

  retriever.add_documents(docs, ids=None) ## Adds documents to the docstore and vectorstores.
  ~~~
  - **Vector Storeに同じParent Documentを持つSub Docはすべて同じ`doc_id`をmetadataとして持つ。  
    また、`docstore`に指定したところに(上記の例だとローカルディスクの`/opt/rag/documents`)同じ`doc_id`のファイル名のファイル(フルドキュメント)が保存される**
  - **この`doc_id`をキーとしてParent Documentを取得してくる**
- **類似度の高いChunkとそのChunkの原本(フル)ドキュメントを取得してLLMに投げる**  
  ~~~python
  ## Promptに参照元ドキュメント全文を入れる方法（parent_splitterパラメータを省略）
  retriever = ParentDocumentRetriever(
      vectorstore=db,
      docstore=store, ## 元のドキュメントの場所
      child_splitter=text_splitter,
      search_kwargs = {
          "k": 3, ## docstoreからの(原本のドキュメント)最大取得件数
      }
  )

  ## Promptに全文ではなく、parent_splitterで指定した大きいChunkを入れる方法（parent_splitterパラメータを指定）
  retriever = ParentDocumentRetriever(
      vectorstore=db,
      docstore=store,  ## 元のドキュメントを保存する場所
      child_splitter=text_splitter,
      parent_splitter=parent_splitter,
  )

  qa_chain = ConversationalRetrievalChain.from_llm(
      retriever=retriever,
      llm=llm,
      memory=memory,
      chain_type="stuff",
      verbose=True ## LLMへの最終的なPromptを表示するかどうか
  )
  ~~~
  - **`search_kwargs`の`k`を１にすると類似度の最も高いChunkが含まれている元の(フル)ドキュメントを１つだけ取得し、複数(2以上)を指定した場合は類似度の高い複数のChunkが含まれているそれぞれの複数の元の(フル)ドキュメントを取得する。(動作確認結果、kの個数分必ず取得するわけではなく、最大取得件数っぽい)**

#### ■ Parent Document Retriever時の注意点
- Document Loaderの`load()`メソッドでロードすること！`load_and_split()`メソッドでロードするとParent Documentも分割されてしまう！
- Document Store(`docstore`パラメータで指定)として使えるストレージは以下
  - https://python.langchain.com/docs/integrations/stores/
  - 現状永久的に保存できるものとしては`LocalFileStore`くらい(ローカルディスクに保存)
    - `LocalFileStore`を使う時、`create_kv_docstore`でディレクトリを変換する必要がある  
      ~~~python
      from langchain.storage import LocalFileStore
      from langchain.storage._lc_store import create_kv_docstore
      
      files_location = "/opt/rag/documents" ## ドキュメントファイルが格納されるディレクトリ
      fs = LocalFileStore(files_location)
      store = create_kv_docstore(fs)
      ~~~
- **参照元ドキュメントの全文をPromptに入れると (特に参照元ドキュメントのサイズが大きいときは) Promptのサイズが大きくなりすぎて、応答までより時間がかかったり、Token数も多くなって利用料が上がったり、API側のスロットリングに引っ掛かる可能性があるため、parent_splitterを使って類似度検索で使うChunkとPromptに含めるChunkサイズを分ける方式の方が現実的な気がする。**

## 2. Multi-Query Retriever
- １つのクエリーをLLMで多角的な観点から複数のクエリーに変換して、それぞれのクエリーで類似度検索を行い、取得した複数のChunkをまとめてLLMに渡す手法

## 3. Self-Querying Retriever
- LangChain が提供するRetrieverの一種で、ユーザーの自然言語クエリを解析して セマンティック検索 + メタデータフィルタリング を自動で組み合わせてくれる仕組み
- ユーザーの質問をLLMに解釈させ、以下を自動で抽出し、検索に利用する。
  1. 検索クエリ部分（埋め込み検索用のテキスト）
  2. フィルタリング条件（メタデータに基づく制約）
- 例
  - クエリ：「2021年以降に出版されたPythonに関する論文を探して」
  - 「Python 論文」で埋め込み検索 ＋ `year >= 2021` でメタデータをフィルタリング

## 4. Time-weighted
- 時系列データを扱う場合に、最新の情報に重みを置いて類似度検索を行う手法

## その他
#### チャンク分割（チャンキング戦略）

#### メタデータによるフィルタリング
- https://zenn.dev/pipon_tech_blog/articles/8cdb27830236c5#2.-%E3%83%A1%E3%82%BF%E3%83%87%E3%83%BC%E3%82%BF%E3%81%A7%E3%81%AE%E3%83%95%E3%82%A3%E3%83%AB%E3%82%BF%E3%83%AA%E3%83%B3%E3%82%B0  
  ~~~python
  # metadataのlangが'en'のDocumentsオブジェクトのみを抽出
   retriever = vector_store.as_retriever(search_kwargs={"filter": {"lang": "en"}})
  ~~~