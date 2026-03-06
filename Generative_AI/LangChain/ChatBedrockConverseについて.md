# ChatBedrock vs ChatBedrockConverse 比較まとめ

## 概要
LangChain AWS (`langchain-aws`) には Bedrock を使うためのチャットモデルクラスが2つある。

| クラス | 使用AWS API | 位置づけ |
|---|---|---|
| `ChatBedrock` | InvokeModel API（モデル固有） | レガシー（将来廃止予定） |
| `ChatBedrockConverse` | Bedrock Converse API（統一） | **推奨・将来の標準** |

---

## 最大の違い：モデル間インターフェースの差異を吸収してくれる

**ChatBedrockConverse が推奨される最も大きな理由**は、AWS Bedrock Converse API がモデルプロバイダーごとのリクエスト/レスポンス形式の差異をAPI側で吸収してくれる点にある。

### InvokeModel API（ChatBedrock）の場合

各プロバイダーが独自のリクエスト形式を定義しており、Bedrock はそのペイロードをそのままモデルに渡す。

```
Anthropic Claude  →  { "messages": [...], "anthropic_version": "bedrock-2023-05-31", ... }
Meta Llama        →  { "prompt": "<s>[INST] ... [/INST]", "max_gen_len": ... }
Mistral           →  { "prompt": "[INST] ... [/INST]", "max_tokens": ... }
```

モデルを切り替えると**リクエスト形式も変わる**ため、コードの修正が必要になる。
LangChain の `ChatBedrock` は内部で `ChatPromptAdapter` を使ってこの変換を吸収しているが、完全ではなく、`id` フィールドの付与など細かい差異でエラーが起きることがある（後述）。

### Converse API（ChatBedrockConverse）の場合

すべてのモデルに対して統一されたリクエスト/レスポンス形式を提供する。

```python
# モデルを変えてもコードは同じ
llm = ChatBedrockConverse(model="anthropic.claude-3-5-sonnet-...")
llm = ChatBedrockConverse(model="meta.llama3-...")
llm = ChatBedrockConverse(model="mistral.mistral-large-...")
```

モデルプロバイダー固有の差異はConverse API側が吸収するため、**コードを一切変えずにモデルを切り替えられる**。
ToolMessage の `id` フィールドのような細かいフォーマット差異も同様に吸収される。

---

## 基盤となるAWS APIの違い

### ChatBedrock（旧来）
- AWS Bedrock の **InvokeModel API** を使用
- モデルプロバイダーごとに異なるリクエスト/レスポンス形式を持つ
- 内部で `ChatPromptAdapter` を使って各モデル向けにプロンプトを変換する
- ペイロードは **Anthropic Messages API 形式**（Anthropicモデルの場合）

### ChatBedrockConverse（新）
- AWS Bedrock の **Converse API** を使用
- すべてのモデルに対して **統一されたインターフェース** を提供
- プロンプトの変換はAPI側が吸収するため、`ChatPromptAdapter` は不使用
- `ChatBedrock` を将来的に置き換える実装として位置づけられている

---

## 実際に発生したエラー事例：MCPツール + ChatBedrock

### エラー内容
```
An error occurred (ValidationException) when calling the InvokeModel operation:
messages.18.content.1.tool_result.content.0.text.id: Extra inputs are not permitted
```

### 原因の詳細
MCPツール（`langchain-mcp-adapters`）のレスポンスが Bedrock InvokeModel API の厳格なフォーマット要件を満たさないために発生する。

```
[発生する流れ]

1. MCPツールの実行結果
   → _convert_mcp_content_to_lc_block()
   → create_text_block(text=...) で変換

2. create_text_block() は自動的に id フィールドを付与
   → {"type": "text", "text": "...", "id": "lc_xxxxxxxx"}

3. ToolMessage の content がリスト形式になる
   → [{"type": "text", "text": "...", "id": "lc_xxxxxxxx"}]

4. ChatBedrock (InvokeModel API = Anthropic Messages API 形式) へ送信
   → tool_result の content block に id フィールドは許可されない
   → ValidationException エラー ❌
```

### なぜ ChatBedrockConverse では起きないのか
Converse API はリクエストの正規化・フォーマット変換を API 側で行うため、`id` フィールドの有無による差異を吸収する。
これは「モデル間インターフェースの差異を吸収する」という Converse API の設計思想と同じ仕組みによるもの。