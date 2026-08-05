- RAGの回答精度などを評価するために、いくつかツールがある
- 参考サイト
  - https://zenn.dev/pipon_tech_blog/articles/98eb5ea00bcc52
  - https://www.ai-shift.co.jp/techblog/4003

# Ragas
- https://docs.ragas.io/en/latest/index.html

## 概要
- RAGAS（**R**etrieval **A**ugmented **G**eneration **As**sessment）は、RAGパイプラインを評価するためのOSSフレームワーク
- 人手によるラベル付き正解データ（ground truth）が無くても評価できるのが特徴（LLMを評価者＝judgeとして使う「LLM-as-a-judge」方式のメトリクスが中心）
- 「Retrieval（検索）」と「Generation（生成）」それぞれのステップを分離して評価できるため、精度が低い場合にボトルネックが検索側にあるのか生成側にあるのかを切り分けやすい
- LangChain・LlamaIndexなど主要なRAGフレームワークと連携可能

## 評価に必要なデータ
評価対象のメトリクスによって必要なデータが異なるが、基本的には以下を1レコードとして評価用データセットを作る
- `question`（`user_input`）：ユーザーの質問
- `answer`（`response`）：RAGが生成した回答
- `contexts`（`retrieved_contexts`）：検索によって取得されたチャンク（複数）
- `ground_truth`（`reference`）：（あれば）人が用意した正解の回答

## 主な評価指標（メトリクス）
### Retrieval（検索）に関する指標
- **Context Precision**：検索で取得したコンテキストのうち、質問に関連するものがどれだけ上位に集まっているか（ノイズが少ないか）を評価
- **Context Recall**：正解に必要な情報が、検索されたコンテキストにどれだけ網羅されているかを評価（ground_truthが必要）
- **Context Entities Recall**：正解に含まれる固有表現（エンティティ）が、検索されたコンテキストにどれだけ含まれているかを評価

### Generation（生成）に関する指標
- **Faithfulness（忠実性）**：生成された回答が、取得したコンテキストの内容にどれだけ忠実か（ハルシネーションしていないか）を評価
- **Answer Relevancy（回答の関連性）**：生成された回答が、質問に対してどれだけ的確に答えられているかを評価
- **Answer Correctness（回答の正確性）**：生成された回答が正解（ground_truth）とどれだけ意味的・事実的に一致しているかを評価
- **Answer Semantic Similarity（意味的類似度）**：生成された回答と正解の意味的な類似度を埋め込みベクトルで評価

### 統合的な指標
- **Answer Correctness** はFaithfulnessとRelevancyなど複数観点を組み合わせたスコアとして扱われることもある

---

## 使い方（イメージ）
```python
from datasets import Dataset
from ragas import evaluate
from ragas.metrics import (
    faithfulness,
    answer_relevancy,
    context_precision,
    context_recall,
)

# 評価用データセットを作成
data = {
    "question": ["...質問..."],
    "answer": ["...RAGの回答..."],
    "contexts": [["...取得したチャンク1...", "...取得したチャンク2..."]],
    "ground_truth": ["...正解..."],
}
dataset = Dataset.from_dict(data)

result = evaluate(
    dataset,
    metrics=[faithfulness, answer_relevancy, context_precision, context_recall],
)
print(result)
```
- 内部的にはLLM（デフォルトはOpenAIのモデルだが、他のLLMにも差し替え可能）にプロンプトを投げてスコアリングさせている
- 評価用のテストデータ（質問・正解ペア）が無い場合、Ragasの`TestsetGenerator`機能でドキュメントから合成テストデータを自動生成することも可能（詳細は後述）

## メリット
- 人手によるラベリングコストを抑えつつ、複数観点からRAGの精度を定量評価できる
- Retrieval / Generationを分離して評価できるため、改善箇所の切り分けがしやすい
- CI/CDに組み込んでRAGパイプラインの品質を継続的にモニタリングすることも可能

## 注意点・デメリット
- LLM-as-a-judge方式のため、評価に使うLLMの性能・クセに評価結果が影響される
- 評価のたびにLLM APIを呼び出すため、コスト・時間がかかる
- スコアの絶対値よりも、改善前後での相対比較（トレンド）として使う方が実用的

---

## Synthetic Test Data Generation（合成テストデータ生成）
- https://docs.ragas.io/en/v0.1.21/concepts/testset_generation.html
- 評価（`evaluate`）にはquestion/ground_truthを含むテストデータセットが必要だが、これを人手で大量に作るのはコストが高い
- Ragasは元となるドキュメント群から、LLMを使って質問・正解ペアを自動生成する`TestsetGenerator`機能を提供している

### 仕組み
1. 元ドキュメント（LangChainの`Document`やLlamaIndexの`Node`など）を入力として与える
2. ドキュメント群から **Knowledge Graph（ナレッジグラフ）** を構築する
   - ドキュメントをチャンク単位のノードに分割し、要約・見出し・エンティティなどをノードに付与（transforms）
   - チャンク間の類似度や共通エンティティなどを元にノード同士をエッジで関連付ける（似た内容・関連する内容のチャンクをつなげる）
3. ナレッジグラフ上でノード（＋関係）をサンプリングし、**Query Synthesizer**がLLMを使って質問と正解を生成する
   - **Single-hop**：1つのチャンク（ノード）だけから答えられる質問
   - **Multi-hop（Specific / Abstract）**：複数のチャンク（ノード）にまたがる情報を組み合わせないと答えられない、より複雑な質問
   - デフォルトでは複数のSynthesizerを組み合わせ、シナリオ（single-hop / multi-hop）の比率を指定して生成する
4. （オプション） **Persona（ペルソナ）** を定義することで、「初心者ユーザー」「専門家」など異なる立場・関心を持つ質問者を想定した質問を生成させることも可能

### 使い方（イメージ）
```python
from ragas.testset import TestsetGenerator
from langchain_openai import ChatOpenAI, OpenAIEmbeddings

generator_llm = ChatOpenAI(model="gpt-4o")
embeddings = OpenAIEmbeddings()

generator = TestsetGenerator(llm=generator_llm, embedding_model=embeddings)

# docsはLangChainのDocumentのリストなど
testset = generator.generate_with_langchain_docs(
    docs,
    testset_size=10,  # 生成する質問数
)

df = testset.to_pandas()
```
- 生成されたテストセットには、質問（`user_input`）・正解（`reference`）・参照コンテキスト（`reference_contexts`）などが含まれ、そのまま`evaluate`の入力データとして利用できる

### メリット
- 人手でのQ&A作成コストを大幅に削減でき、多様なタイプ（単純・複合・異なるペルソナ視点）の質問を網羅的に用意しやすい

### 注意点
- あくまでLLMによる自動生成のため、不自然な質問や事実と異なる正解が混じることがある → 生成後に人がレビュー・修正するのが推奨
- ドキュメントの質・量に生成結果が左右される（ノイズの多いドキュメントだと質問の質も下がりやすい）
