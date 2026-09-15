- 参考URL
  - https://langchain-ai.github.io/langgraph/concepts/low_level/#send
  - https://langchain-ai.github.io/langgraph/how-tos/graph-api/#map-reduce-and-the-send-api

## 概要
- **Stateの内容に応じて、動的に複数のNodeを並列実行するための機能（map-reduceパターン）**
  - 事前に並列実行する数が分からない場合（例: リストの要素数によって処理数が変わる場合）に使用する
  - 通常の`add_conditional_edges`は「次にどのNodeに進むか」しか指定できないが、`Send`を使うと **「そのNodeにどんなState（入力）を渡すか」も同時に、かつ複数個** 指定できる
  - `Send(<遷移先のNode名>, <そのNodeに渡すState>)`という形式で使用する

  > https://langchain-ai.github.io/langgraph/concepts/low_level/#send
  > A common pattern in LangGraph is to create multiple copies of a node for each item in a list, and have them return the outputs to the parent graph state. ... You can use the `Send` object to accomplish this. `Send` lets you dynamically create edges and pass custom state to each of these nodes as part of the conditional edge function.

## 使い方
- `add_conditional_edges`のルーティング関数（分岐関数）で、Node名（str）の代わりに`Send`オブジェクトの**リスト**を返す
  - `Send`の第一引数: 遷移先のNode名
  - `Send`の第二引数: そのNodeに渡すState（dict）
- リストの要素数分だけ、指定したNodeが並列に（別々のStateを持って）実行される

```python
from typing import Annotated, TypedDict
from operator import add

from langgraph.graph import StateGraph, START, END
from langgraph.types import Send

# Graph全体のState
class OverallState(TypedDict):
    subjects: list[str]
    jokes: Annotated[list[str], add]  # 各並列Nodeの結果をここに集約する（add Reducer）

# 各並列Nodeに渡すState（OverallStateとは別に定義できる）
class JokeState(TypedDict):
    subject: str

def generate_joke(state: JokeState):
    return {"jokes": [f"Joke about {state['subject']}"]}

# subjectsの要素数だけgenerate_jokeを並列実行するためのルーティング関数
def continue_to_jokes(state: OverallState):
    return [Send("generate_joke", {"subject": s}) for s in state["subjects"]]

builder = StateGraph(OverallState)
builder.add_node("generate_joke", generate_joke)
builder.add_conditional_edges(START, continue_to_jokes, ["generate_joke"])
builder.add_edge("generate_joke", END)

graph = builder.compile()
graph.invoke({"subjects": ["cats", "dogs"]})
# -> generate_jokeが"cats"用と"dogs"用の2つ、並列に実行される
# -> それぞれの結果がjokesフィールドにadd Reducerでまとめられる
```

### forを使わず、固定で複数指定する例
- `Send`はStateの要素数分をforでループして生成する必要はなく、**固定の数・固定の内容**で複数個そのまま返してもよい
  - 例えば「常に2つの固定Subjectについて並列実行したい」場合など

```python
# 常に"cats"と"dogs"の2つを並列実行するルーティング関数（forループなし）
def continue_to_fixed_jokes(state: OverallState):
    return [
        Send("generate_joke", {"subject": "cats"}),
        Send("generate_joke", {"subject": "dogs"}),
    ]
```

- 遷移先のNode名も同じである必要はなく、**異なるNodeへ固定で並列に振り分ける**こともできる

```python
# node_bとnode_cへ、それぞれ異なるStateを渡して固定で並列実行するルーティング関数
def route_to_fixed_nodes(state: OverallState):
    return [
        Send("node_b", {"foo": "b"}),
        Send("node_c", {"foo": "c"}),
    ]

builder.add_conditional_edges(START, route_to_fixed_nodes, ["node_b", "node_c"])
```

## ポイント・注意点
- `continue_to_jokes`のようなルーティング関数は、`START`や通常のNodeからの`add_conditional_edges`の第2引数に指定する
  - `add_conditional_edges`の第3引数には、遷移しうるNode名のリスト（またはdict）を渡しておく（グラフ描画のため）
- 各並列実行の結果を1つのStateに集約するため、集約先のフィールド（例の`jokes`）は**add Reducer**（`Annotated[list, add]`）にしておく必要がある
  - デフォルトのReducer（上書き）のままだと、並列実行した結果が競合して正しく集約されない
  - Reducerについては`Reducerについて.md`を参照
- 各並列Nodeに渡すState（`JokeState`）は、Graph全体のState（`OverallState`）と同じフィールドを持つ必要はない（必要なフィールドだけを渡せば良い）

## `Command`との違い
- `Command`: **1つのNode内**で「State更新」と「次のNodeへのルーティング」を同時に行う機能（遷移先は基本的に1つ）
- `Send`: **1つのルーティング関数**から、動的な数のNodeを並列に呼び出し、それぞれに異なるStateを渡す機能（map-reduceパターン）
- `Command`については`LangGraphについて.md`の`Command`セクションを参照
