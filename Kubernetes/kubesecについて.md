## kubesecとは
- https://kubesec.io/
- kubernetesのマニフェストファイルのセキュリティに関する静的解析をしてくれるツール
- 使い方
  - `kubesec scan <マニフェストファイル>`

### kubesectを使った`Secret`の暗号化
- AWS KMSやGoogle Cloud KMSなどで作成した暗号化カギと組み合わせて使う必要がある
- https://nobelabo.hatenablog.com/entry/2023/03/27/085020