## *Function Calling*
- OpenAIモデルが外部の機能(e.g. 関数)やサービスを呼び出して利用する機能
  - LangChain Agentと似たような概念
- 例えば、ユーザからの質問に対して、インターネットから最新のデータや専門的な情報を検索したり、Pythonを実行して統計的な分析やグラフの作成などを行うことができる
- どの関数を使うかもOpenAIモデルが判断
- 関数の実行はOpenAIではなく、プログラム
  - OpenAIは必要な関数を判断してプログラムに指示を出す(呼び出す)だけ
- 参考URL
  - https://qiita.com/yu-Matsu/items/12b686fe4cab343f50b3
