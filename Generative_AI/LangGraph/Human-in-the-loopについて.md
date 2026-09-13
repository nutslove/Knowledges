# 前提
- LangGraphのHuman-in-the-loopは、LangGraphの`interrupt`を利用して実装されている。
  - https://docs.langchain.com/oss/python/langgraph/interrupts
  - https://reference.langchain.com/python/langgraph/types/interrupt

> [!IMPORTANT]
> To use an interrupt, you must enable a **checkpointer**, as the feature relies on persisting the graph state.

# Human-in-the-loopの実装例

> [!IMPORTANT]
> 再開時に、thread_id (`config["configurable"]["thread_id"]`) を前回と同じ値に合わせる必要がある。それにより、同じ会話・同じ一時停止状態にアクセスでき、再開できる。
  
## 例１
```python
"""
LangGraph の interrupt() を試すためのサンプルコード。

参考:
- https://reference.langchain.com/python/langgraph/types/interrupt
- https://langchain-ai.github.io/langgraph/how-tos/human_in_the_loop/wait-user-input/

ポイント:
1. interrupt() はノードの中で呼ぶと、その時点でグラフの実行を一時停止し、
   引数に渡した値（今回は質問内容の dict）を呼び出し元に返す。
2. 一時停止した状態は checkpointer（ここでは MemorySaver）によって保存される。
   → checkpointer を設定していないと interrupt は使えない。
3. 再開するときは `graph.invoke(Command(resume=<再開値>), config)` のように
   Command(resume=...) を渡す。すると interrupt() の戻り値としてその値が返り、
   ノードは最初から再実行される（interrupt() より前の処理も再実行される点に注意）。
4. 同じ thread_id (config["configurable"]["thread_id"]) を使い続けることで、
   同じ会話・同じ一時停止状態にアクセスできる。
"""

from typing import TypedDict

from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, START, StateGraph
from langgraph.types import Command, interrupt


class State(TypedDict):
    input: str
    approved: bool
    result: str


def human_approval(state: State) -> State:
    """人間の承認を interrupt() で待つノード。

    このノードは「呼ばれない」わけではなく、必ず先頭から実行される。
    ただし interrupt() の行で処理が止まる（GraphInterrupt が発生してノードの
    残りのコードが実行されないまま graph.invoke() に制御が戻る）ので、
    見かけ上「ここで一時停止しているように」ノードの中身が二段階に分かれる。

    1回目の graph.invoke(initial_state, config) 時の流れ:
      - ここが呼ばれる → 下の print() は実行される（ログに出る）
      - interrupt() の行で停止 → これより下（2つ目の print() と return）は実行されない

    2回目の graph.invoke(Command(resume=...), config) 時の流れ:
      - ノードは最初から再実行される → 下の print() が「もう一度」実行される
      - interrupt() が今度は resume で渡した値を戻り値として返す（停止しない）
      - そのまま最後まで実行され、2つ目の print() と return が実行される
    """
    # ← 1回目・2回目、どちらの呼び出しでも必ず実行される（2回表示される）
    print(f"[human_approval] ノード実行開始 (input={state['input']!r})")

    # 1回目の呼び出し: ここで GraphInterrupt が発生し、関数はここで中断される。
    #   → value の内容が graph.invoke() の戻り値（__interrupt__ キー）として返る。
    # 2回目の呼び出し（resume 後）: 中断せずに resume で渡した値がそのまま返る。
    answer = interrupt(
        {
            "question": f"「{state['input']}」を実行してよいですか？ (yes/no)",
        }
    )

    # ← 1回目の呼び出しでは実行されない。2回目（resume 後）のみ実行される。
    print(f"[human_approval] interrupt から再開。受け取った値: {answer!r}")
    return {"approved": str(answer).strip().lower() == "yes"}


def process(state: State) -> State:
    """承認結果に応じて処理を分岐するノード。"""
    if state["approved"]:
        result = f"「{state['input']}」を実行しました。"
    else:
        result = "処理はキャンセルされました。"
    print(f"[process] {result}")
    return {"result": result}


def build_graph():
    builder = StateGraph(State)
    builder.add_node("human_approval", human_approval)
    builder.add_node("process", process)
    builder.add_edge(START, "human_approval")
    builder.add_edge("human_approval", "process")
    builder.add_edge("process", END)

    # interrupt を使うには checkpointer が必須。
    checkpointer = MemorySaver()
    return builder.compile(checkpointer=checkpointer)


def main():
    graph = build_graph()

    # thread_id ごとに会話（実行状態）が管理される。
    config = {"configurable": {"thread_id": "demo-thread-1"}}

    initial_state: State = {
        "input": "本番サーバーを再起動する",
        "approved": False,
        "result": "",
    }

    print("=" * 60)
    print("1回目の実行: interrupt が発生するところまで進む")
    print("=" * 60)
    # ここで human_approval ノードは実行される（START → human_approval と進む）。
    # ただしノードの中の interrupt() で停止するので、process ノードまでは到達しない。
    result = graph.invoke(initial_state, config=config)
    print("\n--- graph.invoke() の戻り値 ---")
    print(result)

    # interrupt が発生すると、戻り値の "__interrupt__" キーに
    # Interrupt オブジェクトのタプルが入っている。
    if "__interrupt__" in result:
        interrupt_obj = result["__interrupt__"][0]
        print("\n>>> interrupt が発生しました")
        print(">>> interrupt_obj:", interrupt_obj)
        print(">>> interrupt.value:", interrupt_obj.value)

        # ここで実際に人間からの入力を受け取る。
        user_answer = input(
            f"\n{interrupt_obj.value['question']} > "
        )

        print("\n" + "=" * 60)
        print("2回目の実行: Command(resume=...) で再開する")
        print("=" * 60)
        # human_approval ノードが「最初から」再実行される（ノード実行開始のログが
        # もう一度出るのはそのため）→ interrupt() が user_answer を返す → process へ進む。
        result2 = graph.invoke(Command(resume=user_answer), config=config)
        print("\n--- 再開後の graph.invoke() の戻り値 ---")
        print(result2)
    else:
        print("\ninterrupt は発生しませんでした。")


if __name__ == "__main__":
    main()
```
- `no`の時の出力  
  ```shell
  ============================================================
  1回目の実行: interrupt が発生するところまで進む
  ============================================================
  [human_approval] ノード実行開始 (input='本番サーバーを再起動する')

  --- graph.invoke() の戻り値 ---
  {'input': '本番サーバーを再起動する', 'approved': False, 'result': '', '__interrupt__': [Interrupt(value={'question': '「本番サーバーを再起動する」を実行してよいですか？ (yes/no)'}, id='6c0decec5ba19c7c8a91b61940866424')]}

  >>> interrupt が発生しました
  >>> interrupt.value: {'question': '「本番サーバーを再起動する」を実行してよいですか？ (yes/no)'}

  「本番サーバーを再起動する」を実行してよいですか？ (yes/no) > no

  ============================================================
  2回目の実行: Command(resume=...) で再開する
  ============================================================
  [human_approval] ノード実行開始 (input='本番サーバーを再起動する')
  [human_approval] interrupt から再開。受け取った値: 'no'
  [process] 処理はキャンセルされました。

  --- 再開後の graph.invoke() の戻り値 ---
  {'input': '本番サーバーを再起動する', 'approved': False, 'result': '処理はキャンセルされました。'}
  ```
- `yes`の時の出力  
  ```shell
  ============================================================
  1回目の実行: interrupt が発生するところまで進む
  ============================================================
  [human_approval] ノード実行開始 (input='本番サーバーを再起動する')

  --- graph.invoke() の戻り値 ---
  {'input': '本番サーバーを再起動する', 'approved': False, 'result': '', '__interrupt__': [Interrupt(value={'question': '「本番サーバーを再起動する」を実行してよいですか？ (yes/no)'}, id='825de82ee07cc78c8d3eebe255e1c21f')]}

  >>> interrupt が発生しました
  >>> interrupt_obj: Interrupt(value={'question': '「本番サーバーを再起動する」を実行してよいですか？ (yes/no)'}, id='825de82ee07cc78c8d3eebe255e1c21f')
  >>> interrupt.value: {'question': '「本番サーバーを再起動する」を実行してよいですか？ (yes/no)'}

  「本番サーバーを再起動する」を実行してよいですか？ (yes/no) > yes

  ============================================================
  2回目の実行: Command(resume=...) で再開する
  ============================================================
  [human_approval] ノード実行開始 (input='本番サーバーを再起動する')
  [human_approval] interrupt から再開。受け取った値: 'yes'
  [process] 「本番サーバーを再起動する」を実行しました。

  --- 再開後の graph.invoke() の戻り値 ---
  {'input': '本番サーバーを再起動する', 'approved': True, 'result': '「本番サーバーを再起動する」を実行しました。'}
  ```

## 例２
```python
"""
graph.stream() を使って interrupt を扱うサンプル。

invoke() では戻り値の "__interrupt__" キーに Interrupt が入っていたが、
stream() では途中のイベントとして {"__interrupt__": (Interrupt(...), ...)} が流れてくる。
CLI で対話しながら、何度でも承認/却下を繰り返せるループにしてある。
"""

from typing import TypedDict

from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, START, StateGraph
from langgraph.types import Command, interrupt


class State(TypedDict):
    task: str
    history: list[str]


def ask_human(state: State) -> State:
    answer = interrupt(
        {
            "question": f"次のタスクを実行しますか？: {state['task']}",
            "choices": ["yes", "no", "quit"],
        }
    )
    return {"history": state["history"] + [f"human said: {answer}"]}


def run_task(state: State) -> State:
    print(f"  -> タスク実行: {state['task']}")
    return {"history": state["history"] + [f"executed: {state['task']}"]}


def skip_task(state: State) -> State:
    print(f"  -> タスクをスキップ: {state['task']}")
    return {"history": state["history"] + [f"skipped: {state['task']}"]}


def route(state: State):
    last = state["history"][-1]
    if "yes" in last:
        return "run_task"
    return "skip_task"


def build_graph():
    builder = StateGraph(State)
    builder.add_node("ask_human", ask_human)
    builder.add_node("run_task", run_task)
    builder.add_node("skip_task", skip_task)

    builder.add_edge(START, "ask_human")
    builder.add_conditional_edges("ask_human", route, ["run_task", "skip_task"])
    builder.add_edge("run_task", END)
    builder.add_edge("skip_task", END)

    return builder.compile(checkpointer=MemorySaver())


def stream_until_interrupt(graph, input_or_command, config):
    """stream を回して、途中で interrupt が来たらその Interrupt を返す。"""
    for event in graph.stream(input_or_command, config=config, stream_mode="updates"):
        if "__interrupt__" in event:
            return event["__interrupt__"][0]
        print(f"  [event] {event}")
    return None


def main():
    graph = build_graph()

    tasks = ["ログをローテーションする", "本番DBをリストアする"]

    for i, task in enumerate(tasks):
        config = {"configurable": {"thread_id": f"stream-demo-{i}"}}
        print("\n" + "=" * 60)
        print(f"タスク {i + 1}: {task}")
        print("=" * 60)

        interrupt_obj = stream_until_interrupt(
            graph, {"task": task, "history": []}, config
        )

        while interrupt_obj is not None:
            print(f">>> interrupt: {interrupt_obj.value}")
            answer = input(f"{interrupt_obj.value['question']} [yes/no/quit] > ")
            if answer == "quit":
                print("中断します。")
                break
            interrupt_obj = stream_until_interrupt(
                graph, Command(resume=answer), config
            )

    print("\n全タスク終了。")


if __name__ == "__main__":
    main()
```