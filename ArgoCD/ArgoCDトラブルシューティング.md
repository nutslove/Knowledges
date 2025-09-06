## Redis HA Serverが起動しない
### 前提
- EKS Auto Modeを使っていて、21日ごとにWorkerNodeが置き換わる

### 事象
- 3つのRedis HA Server Podのうち、1つが`2/3 Running`状態でになって起動しない
- そのRedis HA Server PodがMasterだった
- そのPodと他の2つのRedis HA Server Podから、Masterにアクセスできないみたいなエラーが出ていた

### 原因
- トリガーとなったのは、EKS Auto ModeによるWorker Nodeの置き換え。問題となったPodがそのNodeにスケジューリングされていた。
- 真の原因は不明・・・。

### 対応
- 問題のPodを削除して再作成したり、色々試したがダメだった
- **結局のところ、`kubectl delete statefulset argocd-redis-ha-server -n argocd`で完全に `argocd-redis-ha-server` を削除して、再度デプロイしたら直った・・・。**