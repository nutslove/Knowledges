# メモリの種類
- 参考URL
  - https://zenn.dev/pharmax/articles/26be245e159590
  - https://docs.langchain.com/oss/python/langgraph/add-memory
- LangGraphには以下の2種類のメモリがある  
  - **CheckPointer**（短期記憶 (Short-term Memory)）: 1つのスレッド（会話）のStateの情報を保存するメモリ。  
  - **Store**（長期記憶 (Long-term Memory)）: 複数のスレッド（会話）にまたがって情報を保存するメモリ。

## CheckPointer
- ワークフローの実行中に**特定の時点のステート**をスナップショットのように保存する機能
### チェックポイントのメリット(目的)
1. **ステートの永続化**
    - ワークフローの実行状態を保存し、あとでその時点のステートの状態から再開できる
2. **エラー回復**
    - 処理中にエラーが発生した場合、直前のチェックポイントから再開できる
3. **デバック**
    - ワークフローの実行過程を追跡し、問題の原因を特定しやすくする

### チェックポイントのデータ構造
- LangGraphの処理ステップごとに`CheckpointTuple`というデータ構造で保存される

## Store
