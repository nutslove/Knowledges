## Reducer
- **Stateを更新するための関数**
- 各Nodeの出力をGraphの全体Stateに統合する役割を持つ
- 別途Reducerを指定しない場合のデフォルトのReducerは、該当Stateの前の値を上書きする
- add Reducerは以前の値に新しい値を追加する（リスト）
  - MessageStateのReducerは、add Reducerを使用している
  - `typing`の`Annotated`と`operator`モジュールの`add`関数を使用  
  - add Reducerの例  
    ```python
    from typing import Annotated, TypedDict, List
    from langgraph.graph import StateGraph, START, END
    from operator import add

    Class ReducerState(TypedDict):
        query: str
        documents: Annotated[List[str], add]
    ```
- 重複排除などユーザが独自のReducerを定義することも可能
  - Custom Reducer

## Custom Reducerの例
```python
from typing import Annotated, TypedDict, List

def reduce_unique_documents(left: list | None, right: list | None) -> list:
  """Combines two lists of documents, removing duplicates."""
    if not left:
        left = []
    if not right:
        right = []
    return list(set(left + right))

class CustomReducerState(TypedDict):
    query: str
    documents: Annotated[List[str], reduce_unique_documents]
```