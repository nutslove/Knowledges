- 参考URL
  - https://docs.langchain.com/oss/python/langchain/middleware/built-in#summarization

# SummarizationMiddleware 詳細ガイド

LangChainの`SummarizationMiddleware`は、会話履歴が長くなった際に自動的に要約を行い、トークン数を管理するミドルウェア。

## 基本的な使い方

```python
from langchain.agents.middleware import SummarizationMiddleware

middleware = SummarizationMiddleware(
    model=model,
    trigger=("tokens", 30000),      # 要約をトリガーする条件
    keep=("messages", 10),          # 要約後に保持するメッセージ数
    trim_tokens_to_summarize=4000,  # 要約LLMに渡す最大トークン数（デフォルト: 4000）
)
```

## パラメータ詳細

### `trigger` - 要約トリガー条件

**要約を実行するタイミング**を指定する。

| 形式 | 説明 | 例 |
|------|------|-----|
| `("tokens", N)` | Nトークンに達したら要約 | `("tokens", 30000)` |
| `("messages", N)` | Nメッセージに達したら要約 | `("messages", 50)` |
| `("fraction", F)` | モデルの最大入力トークンのF%で要約 | `("fraction", 0.8)` |
| リスト | 複数条件（いずれかを満たしたら） | `[("tokens", 30000), ("messages", 50)]` |

> [!IMPORTANT]   
> `trigger`を設定しないと要約は**実行されない**。`keep`だけ設定しても意味がない。

> [!CAUTION]  
> `fraction`を使用する場合、モデルのprofile情報（`max_input_tokens`）が必要。profile情報がないモデルでは`ValueError`が発生する。その場合は`tokens`または`messages`を使用するか、モデル初期化時にprofileを指定する：  
> ```python
> ChatModel(..., profile={"max_input_tokens": 200000})
> ```

### `keep` - 保持するコンテキスト量

**要約後に保持するメッセージ量**を指定する（デフォルト: `("messages", 20)`）。

| 形式 | 説明 | 例 |
|------|------|-----|
| `("messages", N)` | 最新N件のメッセージを保持 | `("messages", 10)` |
| `("tokens", N)` | 最新Nトークン分を保持 | `("tokens", 5000)` |
| `("fraction", F)` | モデルの最大入力トークンのF%を保持 | `("fraction", 0.3)` |

> [!CAUTION]  
> `keep`でも`fraction`を使用する場合、同様にモデルのprofile情報が必要。

### `trim_tokens_to_summarize` - 要約用トリムサイズ

**要約LLMに渡すメッセージの最大トークン数**を指定する（デフォルト: `4000`）。

```python
trim_tokens_to_summarize=4000  # デフォルト値
```

## 処理フロー

```
1. メッセージ合計がtrigger条件を満たす
   ↓
2. keep で指定した最新メッセージを保持
   ↓
3. 残り（古いメッセージ）= 要約対象
   ↓
4. 要約対象を trim_tokens_to_summarize でトリム
   ↓
5. トリムしたメッセージをLLMに渡して要約生成
   ↓
6. 古いメッセージを削除し、要約 + 保持メッセージで置き換え
```

### 具体例

```
設定: trigger=("tokens", 30000), keep=("messages", 10), trim_tokens_to_summarize=4000

メッセージ合計: 30000トークン（trigger発動）
        ↓
最新10件を保持（例: 5000トークン）
        ↓
要約対象 = 残り（例: 25000トークン）
        ↓
25000トークンを4000トークンにトリム ← ここで大部分が捨てられる
        ↓
4000トークン分だけをLLMで要約
```

## "Previous conversation was too long to summarize." エラー

### 発生条件

`trim_tokens_to_summarize`でトリムした結果、**HumanMessageが含まれない**場合に発生する。

内部では`trim_messages`関数が以下の設定で呼ばれる：
- `strategy="last"` - 末尾からトークンを取得
- `start_on="human"` - 結果の先頭がHumanMessageである必要がある

### 具体例

```
要約対象: [HumanMessage(100)] [AIMessage(5000)] [ToolMessage(3000)] [AIMessage(4000)]
                                                                    ←─────────────→
                                                                    4000トークン範囲
                                                                   （HumanMessageがない）
```

末尾から4000トークン分を取得しても、その範囲にHumanMessageがないため、空リストになる。

### 結果

```python
HumanMessage(
    content="Here is a summary of the conversation to date:\n\nPrevious conversation was too long to summarize.",
    additional_kwargs={"lc_source": "summarization"},
)
```

**実質的に意味のある要約がないまま、古いメッセージは削除される。**

### 解決方法

`trim_tokens_to_summarize`を増やして、HumanMessageが含まれる範囲まで広げる：

```python
SummarizationMiddleware(
    model=model,
    trigger=("tokens", 30000),
    keep=("messages", 10),
    trim_tokens_to_summarize=15000,  # 4000から増やす
)
```

## AI/ToolMessage ペア保持ロジック

SummarizationMiddlewareは、AIMessageとそれに対応するToolMessageのペアが分割されないようにカットオフポイントを調整する。

```python
# _find_safe_cutoff_point メソッド
# ToolMessageがカットオフ位置にある場合、対応するAIMessageまで遡る
```

これにより、保持されるメッセージ数が`keep`で指定した値より多くなることがある。

## 推奨設定

### 基本設定

```python
SummarizationMiddleware(
    model=model,
    trigger=("tokens", 30000),
    keep=("messages", 10),
    trim_tokens_to_summarize=15000,  # triggerの50%程度
)
```

### 設定のポイント

| パラメータ | 推奨 | 理由 |
|-----------|------|------|
| `trigger` | 必ず設定する | 未設定だと要約が実行されない |
| `keep` | 用途に応じて | 少なすぎると文脈が失われる |
| `trim_tokens_to_summarize` | `trigger`の50-80% | 小さすぎるとHumanMessageに到達できない |

### Claude Sonnet使用時の例

Claude Sonnetは200Kトークン入力可能なので、余裕を持った設定が可能：

```python
SummarizationMiddleware(
    model=ChatBedrockConverse(model_id="us.anthropic.claude-sonnet-4-5-20250929-v1:0"),
    trigger=("tokens", 40000),
    keep=("messages", 10),
    trim_tokens_to_summarize=20000,
)
```

## 要約メッセージの識別

要約により生成されるHumanMessageには、以下のマーカーが付与される：

```python
HumanMessage(
    content="Here is a summary of the conversation to date:\n\n{要約内容}",
    additional_kwargs={"lc_source": "summarization"},
)
```

このマーカーを使って、要約メッセージを識別・フィルタリングできる：

```python
def is_summary_message(message: BaseMessage) -> bool:
    return message.additional_kwargs.get("lc_source") == "summarization"
```

## トラブルシューティング

### 要約が実行されない

1. `trigger`が設定されているか確認
2. `trigger`条件を満たしているか確認（トークン数/メッセージ数）

### メッセージが`keep`より多く残る

AI/ToolMessageペア保持ロジックにより、ペアが分割されないようにカットオフが調整される。これは正常な動作。

### "Previous conversation was too long to summarize." が頻発

`trim_tokens_to_summarize`を増やす。デフォルトの4000は、大きなToolMessage/AIMessageが連続する場合に不十分。

### `fraction`使用時にValueErrorが発生

モデルのprofile情報がない。以下のいずれかで対応：
- `tokens`または`messages`形式を使用する
- モデル初期化時に`profile={"max_input_tokens": N}`を指定する