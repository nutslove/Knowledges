## Retrieval
- RAGのためのモジュール
- `ConversationalRetrievalChain`クラスを使う
  - RAGで単発の会話は`RetrievalQA`クラスを、会話履歴を踏まえて回答させる場合は`ConversationalRetrievalChain`クラスを使う
- `ConversationalRetrievalChain`クラスの`from_llm`メソッドでchainを定義して、`ConversationalRetrievalChain.from_llm.run`でVector Storeからの情報を含めてLLMに投げる
  - **`verbose=True`にするとLLMへの最終的なPromptを出力させることができる**
  - **`chain_type="stuff"` (defaultが"stuff")の場合、Vector Storeから取得した複数のデータをすべて最終的なPromptに入れてLLMに投げる**
- **Vector StoreへのEmbedding(ベクトル化)に使うLLMとユーザからの質問をEmbedding(ベクトル化)に使うLLMは同じじゃないといけない？要確認！！**
- 参考URL
  - **https://blog.serverworks.co.jp/langchain-bedrock-memo-1**
- **Knowledge BaseのOpenSearchでメタデータにscoreがあって、質問と取得したデータの類似度を表す**
  - **Knowledge Baseだからメターデータが保存できるのか、OpenSearchだからできるのかは要確認！**
- retrieverの **`get_relevant_documents`** でVector Storeから取得してデータの件数と取得したデータの中身、metadataなどを確認できる  
  ~~~python
  retriever = AmazonKnowledgeBasesRetriever(knowledge_base_id=knowledge_base_for_bedrock_id,retrieval_config=retrieval_config)

  # retrieverで取得してきたデータを確認 (上のnumberOfResultsで指定した件数のデータが取得)
  context_docs = retriever.get_relevant_documents(user_input)
  print(f"len = {len(context_docs)}")
  for context_doc in context_docs:
      print(f"metadata = {context_doc.metadata}")
      print("score = " + str(context_doc.metadata["score"]))
      print(f"page_content = {context_doc.page_content}") ## Vector Storeから取得したデータ
  ~~~
- Retrievalは以下の具体的なステップに分かれる
   - **Document loaders**
   - **Text Splitter**
   - **Text embedding models**
   - **Vector Store**
   - **Retrievers**

### Document loaders
- Vector Storeに保存するドキュメントを取り込む段階  
  > load documents from many different sources. LangChain provides over 100 different document loaders as well as integrations with other major providers in the space, like AirByte and Unstructured. LangChain provides integrations to load all types of documents (HTML, PDF, code) from all types of locations (private S3 buckets, public websites).
- ファイル種類(e.g. PDF,txt,csv)ごとにライブラリが用意されている
  - PDF -> `PyPDFLoader`
    - https://python.langchain.com/docs/modules/data_connection/document_loaders/pdf
  - txt -> `TextLoader`
    - https://python.langchain.com/docs/modules/data_connection/document_loaders/

### Text Splitter
- ドキュメントを複数のチャンクに分割する方法
- やり方が以下の２つある
  1. 各種Loaderライブラリ(e.g. PyPDFLoader)の`load()`とTextSplitterライブラリ(e.g. RecursiveCharacterTextSplitter)の`split_documents()`メソッドで分割する  
     ~~~python
     loader = PyPDFLoader(tmp_location)
     documents = loader.load()

     text_splitter = RecursiveCharacterTextSplitter(
       chunk_size = 1000,
       chunk_overlap = 20,
       length_function = len,
       is_separator_regex = False,
     )

     texts = text_splitter.split_documents(documents)
     ~~~
  2. 各種Loaderライブラリ(e.g. PyPDFLoader)の`load_and_split()`メソッドの引数にTextSplitterライブラリのパラメータを渡す方法
     - https://api.python.langchain.com/en/latest/document_loaders/langchain_community.document_loaders.pdf.PyMuPDFLoader.html#
     ~~~python
     from langchain.vectorstores import OpenSearchVectorSearch
     import boto3

     vector = OpenSearchVectorSearch(
       embedding_function = embeddings,
       index_name = 'aoss-index',
       http_auth = awsauth,
       use_ssl = True,
       verify_certs = True,
       http_compress = True, # enables gzip compression for request bodies
       connection_class = RequestsHttpConnection,
       opensearch_url=opensearch_url="https://aoss-example-1234.us-east-1.aoss.amazonaws.com"
     )

     loader = S3FileLoader(bucket_name, file_key)
     text_splitter = RecursiveCharacterTextSplitter(
       chunk_size = 1000,
       chunk_overlap = 20,
       length_function = len,
       is_separator_regex = False,
     )
     pages = loader.load_and_split(text_splitter=text_splitter)

     vector.add_documents(
       documents = pages,
       vector_field = "osha_vector"
     )
     ~~~
- **Parent Document Retrieverを使う時は、`load()`メソッドを使うこと！！`load_and_split()`メソッドでやるとParent Documentも分割されたものになってしまう**


## `ConversationalRetrievalChain` Classの`from_llm` Methodについて
- https://api.python.langchain.com/en/stable/chains/langchain.chains.conversational_retrieval.base.ConversationalRetrievalChain.html
- `from_llm`メソッドは、`llm`パラメータに指定するLLMを利用して質問と回答を生成する
  - 質問も会話履歴を踏まえてユーザからの質問を書き換えて、最終的にLLMに投げる
  - `verbose`パラメータを`True`にしてログ(標準出力)から確認できる

## OpenSearchへの(ベクトル化した)ドキュメントデータの投入
- `OpenSearchVectorSearch`クラスの`from_documents`を使用
  - https://python.langchain.com/docs/integrations/vectorstores/opensearch
- 例
  ~~~python
  from langchain_community.document_loaders import PyPDFLoader
  from langchain_community.vectorstores import OpenSearchVectorSearch
  from langchain.text_splitter import RecursiveCharacterTextSplitter
  from langchain_community.embeddings import BedrockEmbeddings

  embeddings = BedrockEmbeddings(
      model_id = "cohere.embed-multilingual-v3"
  )

  loader = PyPDFLoader(<PDFファイルのパス>)
  text_splitter = RecursiveCharacterTextSplitter(
      chunk_size=int(chunk_size), ## １つのchunkサイズ
      chunk_overlap=int(chunk_overlap), ## chunk間で重複させる範囲
      length_function=len, ## 文字数で分割
      is_separator_regex=False,
  )
  docs = loader.load_and_split(text_splitter=text_splitter)

  ##「ベクトル化されたデータ」と「ドキュメントのメタデータ(e.g. ファイル名、pageなど)」がOpenSearchに格納される
  docsearch = OpenSearchVectorSearch.from_documents(
      docs,
      embeddings,
      opensearch_url="https://opensearch:9200",
      index_name=index_name_for_load,
      http_auth=("admin", "admin"),
      use_ssl = False,
      verify_certs = False,
      ssl_assert_hostname = False,
      ssl_show_warn = False,
  )
  ~~~