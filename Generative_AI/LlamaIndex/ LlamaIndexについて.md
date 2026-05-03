# LlamaIndex とは

RAG (Retrieval-Augmented Generation) とエージェント構築のための Python フレームワーク。
「ドキュメント → 分割 → 埋め込み → ベクトルストア → 検索 → LLM で合成」という
パイプライン、および **Workflows** という event-driven なオーケストレーションを提供する。

- 公式: <https://developers.llamaindex.ai/python/framework/>
- API リファレンス: <https://developers.llamaindex.ai/python/framework/api_reference/>
- 統合一覧 (LlamaHub): <https://llamahub.ai/>
- 最新版: `llama-index-core` **0.14.21** (2026-04 時点)
- Python 要件: **>=3.10, <4.0** （3.9 は 2026-03 に廃止）

> [!IMPORTANT]
> 0.10 でモノレポからパッケージ分割。**0.12〜0.14 系で旧 Agent / QueryPipeline / ServiceContext が deprecated** になった。
> 古いブログ記事 (`from llama_index import ...`, `ReActAgent.from_tools(...)`, `ServiceContext.from_defaults(...)`) はそのままでは動かないことが多い。

---

## 1. 全体像（パイプライン）

```
Document → NodeParser → Node(=chunk) → Embedding → VectorStore
                                                       │
                                            Query → Retriever → Nodes
                                                                  │
                                                  ResponseSynthesizer (LLM)
                                                                  │
                                                              Response
```

| 抽象 | 役割 |
|---|---|
| `Document` | 生テキスト + メタデータ |
| `NodeParser` / `TextSplitter` | Document を Node (chunk) に分割 |
| `Node` (= `TextNode`) | 検索単位。テキスト・メタデータ・関係を持つ |
| `Embedding` | テキスト → ベクトル変換 |
| `VectorStore` | ベクトル + メタデータの永続化層 (ES, Chroma, Qdrant, …) |
| `StorageContext` | VectorStore / DocStore / IndexStore を束ねる |
| `Index` | 検索可能なデータ構造 (`VectorStoreIndex` 等) |
| `Retriever` | クエリに対して Node を返す |
| `ResponseSynthesizer` | Node 群と質問から LLM で回答を生成 |
| `QueryEngine` | Retriever + Synthesizer を束ねたラッパー |
| `ChatEngine` | QueryEngine + 会話履歴 |
| `Workflow` / `AgentWorkflow` | event-driven にステップを繋ぐ。エージェントの推奨形 |
| `Settings` | グローバル設定 (LLM, Embedding, chunk_size, …) |

---

## 2. 最低限のセットアップ

```python
from llama_index.core import Settings
from llama_index.embeddings.openai_like import OpenAILikeEmbedding
from llama_index.llms.openai_like import OpenAILike

Settings.embed_model = OpenAILikeEmbedding(
    model_name="text-embedding-bge-m3",
    api_base="http://localhost:1234/v1",
    api_key="lm-studio",
    embed_batch_size=16,
)
Settings.llm = OpenAILike(
    model="gemma-4-e4b-it",
    api_base="http://localhost:1234/v1",
    api_key="lm-studio",
    is_chat_model=True,
    temperature=0.3,
)
# よく触る他の設定
Settings.chunk_size = 512        # NodeParser のデフォルトに反映
Settings.chunk_overlap = 50
Settings.context_window = 8192   # LLM のコンテキスト長
Settings.num_output = 512        # 出力トークン上限
```

> [!NOTE]
> - **`Settings` はプロセスグローバル**。
> - テストや並行処理では明示的に`embed_model=`, `llm=` を引数で渡すほうが安全。
> - **`ServiceContext` は deprecated** ── 新規コードでは `Settings` を使う。

---

## 3. 代表的な使い方

### 3.1 文書を取り込んでインデックス化

```python
from llama_index.core import Document, StorageContext, VectorStoreIndex

docs = [Document(text="...", metadata={"source": "a.md"})]

storage_context = StorageContext.from_defaults(vector_store=vector_store)
index = VectorStoreIndex.from_documents(docs, storage_context=storage_context)
```

ファイル/ディレクトリから取り込む場合:

```python
from llama_index.core import SimpleDirectoryReader

docs = SimpleDirectoryReader("./data", recursive=True).load_data()
```

PDF / DOCX などは `llama-index-readers-file` 系の追加リーダーが必要。

### 3.2 既存の VectorStore を再利用する

```python
index = VectorStoreIndex.from_vector_store(vector_store=vector_store)
```

→ 文書は永続化済みなので再 ingest 不要。

### 3.3 類似検索のみ（LLM を使わない）

```python
retriever = index.as_retriever(similarity_top_k=5)
nodes = retriever.retrieve("検索したい文")
for n in nodes:
    print(n.score, n.metadata.get("source"))
    print(n.get_content())
```

### 3.4 RAG（検索 + LLM 回答）

```python
qe = index.as_query_engine(similarity_top_k=3)
resp = qe.query("質問文")
print(str(resp))             # 回答テキスト
for n in resp.source_nodes:  # 参照されたノード
    print(n.metadata, n.score)
```

`as_query_engine()` の主な引数:

| 引数 | 内容 |
|---|---|
| `similarity_top_k` | 上位何件を検索するか |
| `response_mode` | 合成モード（後述） |
| `streaming` | True でトークンストリーム |
| `filters` | メタデータフィルタ (`MetadataFilters`) |
| `node_postprocessors` | 取得後の後処理 / 再ランカ |
| `llm`, `embed_model` | Settings を上書き |

### 3.5 チャット（会話履歴つき）

```python
chat = index.as_chat_engine(chat_mode="condense_plus_context")
print(chat.chat("最初の質問"))
print(chat.chat("それを踏まえてもう少し詳しく"))
```

`chat_mode` の主な選択肢:

| mode | 内容 |
|---|---|
| `simple` | 検索なし。LLM とそのまま会話 |
| `context` | 毎ターン検索してコンテキストに入れる |
| `condense_question` | 履歴を1文に要約してから検索 |
| `condense_plus_context` | 上記 + コンテキスト注入（汎用的にこれ） |

> ツール呼び出しを含むエージェント風の対話は **AgentWorkflow** に移行（§11 参照）。

### 3.6 ストリーミング応答

```python
qe = index.as_query_engine(streaming=True)
resp = qe.query("...")
for token in resp.response_gen:
    print(token, end="", flush=True)
# response_gen を使い切ったあとで resp.source_nodes が読める
```

### 3.7 メタデータフィルタ付き検索

```python
from llama_index.core.vector_stores import (
    MetadataFilters, MetadataFilter, FilterOperator,
)

filters = MetadataFilters(filters=[
    MetadataFilter(key="lang", value="ja", operator=FilterOperator.EQ),
])
qe = index.as_query_engine(similarity_top_k=3, filters=filters)
```

VectorStore 側がフィルタをサポートしていることが前提（ES, Pinecone, Qdrant, … は対応）。

### 3.8 非同期（推奨）

```python
resp = await qe.aquery("...")
nodes = await retriever.aretrieve("...")
```

`Workflows` は async-first。async 版を使うと埋め込み呼び出しなどが並列化される。

---

## 4. インデックスの種類

| Index | 用途 |
|---|---|
| `VectorStoreIndex` | 標準。ベクトル類似検索 |
| `SummaryIndex` (旧 ListIndex) | 全 Node を順番に LLM に渡す。要約向け |
| `KeywordTableIndex` | キーワード抽出ベース。ベクトル不要 |
| `TreeIndex` | 階層的に要約を作って木構造で検索 |
| `KnowledgeGraphIndex` (旧) → **`PropertyGraphIndex`** が後継 | KG ベースの検索 |
| `DocumentSummaryIndex` | ドキュメント単位のサマリで一次選別 |
| `ComposableGraph` | 複数 Index を子として束ねる |

迷ったら `VectorStoreIndex`。要約タスクは `SummaryIndex` を併用。
グラフ検索は新しい `PropertyGraphIndex` を使う。

---

## 5. NodeParser（チャンク分割）

```python
from llama_index.core.node_parser import (
    SentenceSplitter,            # 文境界優先（デフォルト）
    TokenTextSplitter,           # トークン数で機械的に
    SentenceWindowNodeParser,    # 1文ずつ + 前後 N 文を context に保持
    SemanticSplitterNodeParser,  # 埋め込みの差分で意味境界を検出
    HierarchicalNodeParser,      # 親子関係つきの多階層 chunk
    MarkdownNodeParser,          # Markdown 見出し構造で分割
    CodeSplitter,                # 言語別 AST 分割
)

splitter = SentenceSplitter(chunk_size=512, chunk_overlap=50)
nodes = splitter.get_nodes_from_documents(docs)
index = VectorStoreIndex(nodes, storage_context=storage_context)
```

- 日本語: `SentenceSplitter` の `paragraph_separator` / `chunking_tokenizer_fn` を調整。
- コードや Markdown は専用パーサのほうが切れ目が綺麗。

---

## 6. Retriever の種類

| Retriever | 概要 |
|---|---|
| `VectorIndexRetriever` | `index.as_retriever()` のデフォルト |
| `BM25Retriever` | 語彙一致。日本語は分かち書きが必要 |
| `QueryFusionRetriever` | 複数 Retriever を RRF などで統合（ハイブリッド検索） |
| `AutoMergingRetriever` | 子ノードヒット → 親ノードに「マージ」して返す |
| `RecursiveRetriever` | ノード → 別 Index へ再帰的に降りる |
| `RouterRetriever` | LLM が複数 Retriever を選択 |

### 再ランキング（NodePostprocessor）

```python
from llama_index.postprocessor.cohere_rerank import CohereRerank
from llama_index.core.postprocessor import SimilarityPostprocessor

qe = index.as_query_engine(
    similarity_top_k=20,
    node_postprocessors=[
        SimilarityPostprocessor(similarity_cutoff=0.7),  # しきい値
        CohereRerank(top_n=5),                           # 再ランク
    ],
)
```

`top_k` を大きめに取って **postprocessor で絞る** のが定石。

---

## 7. ResponseSynthesizer モード

`response_mode` で挙動が変わる。

| mode | 挙動 | 向き |
|---|---|---|
| `compact` (default) | コンテキストを詰めて1回 / 不足分のみ refine | 一般用途 |
| `refine` | ノードごとに refine を繰り返す | 精度重視 |
| `tree_summarize` | ノードを再帰的に要約 | 要約・長文 |
| `simple_summarize` | 全部結合して1回呼ぶ | 短文のみ |
| `accumulate` | ノードごとに独立回答 → 連結 | Q&A 列挙 |
| `no_text` | LLM 呼ばずに source だけ返す | デバッグ |

```python
qe = index.as_query_engine(response_mode="tree_summarize")
```

明示的に組み立てる場合:

```python
from llama_index.core.response_synthesizers import get_response_synthesizer
from llama_index.core.query_engine import RetrieverQueryEngine

synth = get_response_synthesizer(response_mode="compact")
qe = RetrieverQueryEngine(retriever=retriever, response_synthesizer=synth)
```

---

## 8. プロンプトのカスタマイズ

```python
from llama_index.core import PromptTemplate

qa_tmpl = PromptTemplate(
    "以下のコンテキストを使い、日本語で簡潔に答えてください。\n"
    "コンテキスト:\n{context_str}\n"
    "質問: {query_str}\n"
    "回答:"
)
qe.update_prompts({"response_synthesizer:text_qa_template": qa_tmpl})
```

差し替えポイントは `qe.get_prompts()` で確認できる。
（`refine_template`、`summary_template` 等もある）

---

## 9. 永続化

### 9.1 VectorStore に書く（推奨）

ES / Chroma / Qdrant 等を `StorageContext.from_defaults(vector_store=...)` に渡せば
ベクトルもメタデータもそこに残る。

### 9.2 ローカルファイルにダンプ

VectorStore を使わない構成では:

```python
index.storage_context.persist(persist_dir="./storage")

from llama_index.core import StorageContext, load_index_from_storage
storage_context = StorageContext.from_defaults(persist_dir="./storage")
index = load_index_from_storage(storage_context)
```

DocStore / IndexStore / VectorStore それぞれの JSON が出力される。

### 9.3 差分更新

```python
# upsert: doc_id を揃えて入れ直すと既存を置き換える
doc = Document(text="...", doc_id="manual:section-3")
index.refresh_ref_docs([doc])  # 変更があったものだけ再 embed

index.delete_ref_doc("manual:section-3", delete_from_docstore=True)
```

---

## 10. ハイブリッド検索（語彙 + ベクトル）

```python
from llama_index.retrievers.bm25 import BM25Retriever
from llama_index.core.retrievers import QueryFusionRetriever

vec = index.as_retriever(similarity_top_k=10)
bm25 = BM25Retriever.from_defaults(docstore=index.docstore, similarity_top_k=10)

fusion = QueryFusionRetriever(
    retrievers=[vec, bm25],
    similarity_top_k=10,
    num_queries=1,            # >1 で multi-query 書き換え
    mode="reciprocal_rerank", # RRF
    use_async=True,
)
```

VectorStore 側のハイブリッド機能を使う場合（例: Postgres, Qdrant）は
`hybrid_search=True` を VectorStore に渡し、
`vector_store_query_mode="hybrid"` でクエリエンジンを構築する。

固有名詞・型番のような語彙一致が効くデータでは BM25 を混ぜると精度が出やすい。

---

## 11. Agent / Workflows ★最新

> **旧 `ReActAgent.from_tools(...)` / `AgentRunner` / `AgentWorker` / `QueryPipeline` は deprecated**。
> 新規コードは **Workflows / AgentWorkflow** で書く。

### 11.1 シングルエージェント

```python
from llama_index.core.agent.workflow import FunctionAgent, ReActAgent
from llama_index.core.tools import FunctionTool

def get_weather(city: str) -> str:
    """Get current weather for a city."""
    return f"{city}: 晴れ"

tool = FunctionTool.from_defaults(fn=get_weather)

# Function calling 対応 LLM なら FunctionAgent、非対応なら ReActAgent
agent = FunctionAgent(
    tools=[tool],
    llm=Settings.llm,
    system_prompt="あなたは天気アシスタントです。",
)

resp = await agent.run("東京の天気は？")
print(str(resp))
```

> 旧 `from llama_index.core.agent import ReActAgent` ではなく
> **`from llama_index.core.agent.workflow import FunctionAgent, ReActAgent`** を使う点に注意。

### 11.2 マルチエージェント（AgentWorkflow）

```python
from llama_index.core.agent.workflow import AgentWorkflow

research_agent = FunctionAgent(name="researcher", tools=[...], llm=Settings.llm,
                               system_prompt="...", can_handoff_to=["writer"])
writer_agent   = FunctionAgent(name="writer",     tools=[...], llm=Settings.llm,
                               system_prompt="...", can_handoff_to=["reviewer"])
reviewer_agent = FunctionAgent(name="reviewer",   tools=[...], llm=Settings.llm,
                               system_prompt="...")

workflow = AgentWorkflow(
    agents=[research_agent, writer_agent, reviewer_agent],
    root_agent="researcher",
)
resp = await workflow.run(user_msg="生成 AI の市場動向レポートを書いて")
```

エージェント間は `can_handoff_to` でハンドオフできる。
状態は `Context` に保持され、複数 run で共有可能。

### 11.3 RAG をツール化してエージェントに持たせる

```python
from llama_index.core.tools import QueryEngineTool

rag_tool = QueryEngineTool.from_defaults(
    query_engine=qe,
    name="docs",
    description="社内ドキュメントを検索する。",
)
agent = FunctionAgent(tools=[rag_tool], llm=Settings.llm)
```

### 11.4 自前 Workflow（Step ベース）

`llama-index-workflows` パッケージ（2026-02 公開）は LlamaIndex の中核。
イベント駆動で `@step` デコレータを並べる:

```python
from llama_index.core.workflow import (
    Workflow, StartEvent, StopEvent, Event, step, Context,
)

class RetrievedEvent(Event):
    nodes: list

class MyRAG(Workflow):
    @step
    async def retrieve(self, ev: StartEvent, ctx: Context) -> RetrievedEvent:
        nodes = await retriever.aretrieve(ev.query)
        await ctx.set("query", ev.query)
        return RetrievedEvent(nodes=nodes)

    @step
    async def synthesize(self, ev: RetrievedEvent, ctx: Context) -> StopEvent:
        query = await ctx.get("query")
        resp = await synth.asynthesize(query, ev.nodes)
        return StopEvent(result=str(resp))

resp = await MyRAG(timeout=60).run(query="...")
```

ループ・分岐・並列・人間の確認ステップ等を素直に書ける。

---

## 12. 観測・デバッグ

```python
# 旧来の global handler は残っているが新規は Instrumentation を推奨
from llama_index.core import set_global_handler
set_global_handler("simple")  # プロンプトと応答を標準出力へ

# Instrumentation API（推奨）
from llama_index.core.instrumentation import get_dispatcher
dispatcher = get_dispatcher()
# event handler / span handler を登録できる
```

外部ツール連携: `arize-phoenix`, `langfuse`, `wandb`, `traceloop` 等のインテグレーションあり。

`response.source_nodes` を必ず確認する習慣をつけると、
「LLM の幻覚なのか / Retriever の取りこぼしなのか」を切り分けやすい。

---

## 13. よくある落とし穴 / 注意点

- **`Settings` はグローバル**。複数モデルを同時に使うときはコンポーネントごとに
  `embed_model=` / `llm=` を明示する。
- **埋め込みモデルを途中で変えると既存ベクトルが無効化**される。次元数や分布が変わるため、
  モデル切替時は **再 ingest 必須**。`metadata` にモデル名を残しておくとよい。
- **`from_documents` は毎回 embedding を呼ぶ**。再 ingest は API コストとレイテンシに直結。
  既存 VectorStore からは `from_vector_store` で読み込む。
- **`chunk_size` は埋め込みモデルの最大トークン以下**にする。bge-m3 は 8192 までいけるが、
  検索精度の観点では 256〜1024 程度が無難。
- **メタデータも埋め込みに含まれる**（デフォルト）。秘匿フィールドや巨大な文字列は
  `excluded_embed_metadata_keys` / `excluded_llm_metadata_keys` で除外する。
- **`similarity_top_k` を増やしすぎない**。LLM のコンテキストを圧迫し精度が落ちる。
  20 → rerank → 5 のように二段にするのが現代的。
- **score の意味は VectorStore 依存**。コサイン類似度 / 内積 / L2 距離 で大小関係が逆になる。
  しきい値を決めるときは実データで分布を見る。
- **メタデータフィルタは VectorStore の機能**。サポートしないストアでは無視されるか例外。
- **Insert は 2048 件単位のバッチ**。`insert_batch_size=` で調整可。
- **非同期**: `aquery`, `aretrieve`, `arun` を使うと並列化できる。`asyncio` イベントループ内で
  同期版を呼ぶとブロックする。
- **古い import 文に注意**。
  - `from llama_index import ...` ❌ → `from llama_index.core import ...` ✅
  - `ServiceContext.from_defaults(...)` ❌ → `Settings` ✅
  - `from llama_index.core.agent import ReActAgent` ❌ → `from llama_index.core.agent.workflow import ReActAgent` ✅
  - `QueryPipeline` ❌ → `Workflow` ✅
- **`llama-index-legacy` は廃止済**。依存に残っていたら削除する。

---

## 14. パッケージ分割（v0.10+）

最低限よく入れるもの:

```bash
uv add llama-index-core
uv add llama-index-embeddings-openai-like     # or huggingface, ollama, …
uv add llama-index-llms-openai-like           # or ollama, anthropic, …
uv add llama-index-vector-stores-elasticsearch
uv add llama-index-readers-file               # PDF/DOCX などのリーダー
uv add llama-index-workflows                  # Workflows コア
```

- 埋め込みと LLM は別パッケージ。
- VectorStore は使うバックエンドの分だけ追加。
- Reader / Postprocessor / Tools も個別配布。
- メタパッケージ `llama-index` は主要パッケージをまとめて入れるが、
  サイズが大きいので CI では個別 install が無難。

---

## 15. クイックレシピ集

### 15.1 質問の前にクエリを書き換える（HyDE）

```python
from llama_index.core.query_engine import TransformQueryEngine
from llama_index.core.indices.query.query_transform import HyDEQueryTransform

hyde = HyDEQueryTransform(include_original=True)  # 仮想回答を作って検索
qe2 = TransformQueryEngine(qe, query_transform=hyde)
```

### 15.2 ルーティング（質問の種類で Index を切り替え）

```python
from llama_index.core.query_engine import RouterQueryEngine
from llama_index.core.tools import QueryEngineTool

router = RouterQueryEngine.from_defaults(
    query_engine_tools=[
        QueryEngineTool.from_defaults(query_engine=qe_docs, name="docs",
                                       description="社内ドキュメント"),
        QueryEngineTool.from_defaults(query_engine=qe_code, name="code",
                                       description="コードベース"),
    ],
)
```

### 15.3 構造化出力（Pydantic）

```python
from pydantic import BaseModel

class Answer(BaseModel):
    summary: str
    confidence: float

resp = Settings.llm.as_structured_llm(Answer).complete("...")
print(resp.raw)  # Answer インスタンス
```

### 15.4 トークン数 / コスト見積もり

```python
from llama_index.core.callbacks import CallbackManager, TokenCountingHandler

token_counter = TokenCountingHandler()
Settings.callback_manager = CallbackManager([token_counter])
# クエリを実行
print(token_counter.total_llm_token_count)
print(token_counter.total_embedding_token_count)
```
