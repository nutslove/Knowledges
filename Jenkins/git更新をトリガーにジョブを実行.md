- ジョブの設定で`ビルドトリガ`の`SCMをポーリング`にチェックを入れて、スケジュールの方にpollingする日時(間隔)をCron形式に記述
  - `?`マークを押すとフォーマットや設定例などを確認できる  
  ![](images/polling_1.jpg)  
  ![](images/polling_2.jpg)
- 対象リポジトリは`パイプライン`の方で指定  
  ![](images/additional_exec_1.jpg)
- `追加処理`の`Polling ignores commits in certain paths`で特定のパスorファイル(以外)が更新された場合のみジョブが実行されるようにすることもできそう。  
  → **まだ未検証。試してみる！**  
  ![](images/additional_exec_2.jpg)