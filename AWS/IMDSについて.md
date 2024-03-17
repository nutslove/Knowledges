## IMDSとは
- Instance MetaData Serviceの略
- リンクローカルアドレスの`169.254.169.254`のエンドポイントを持ち、EC2インスタンス内部からのみアクセス可能な特殊なエンドポイントで、インスタンスに関する情報(メタデータ)を取得できる
  - インスタンスID、AMI ID、IAMロール情報、ネットワーク設定など
- IMDSはEC2をホスティングしているホスト(Xen hypervisor)上で動いているサービス
  - https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/using-instance-addressing.html#link-local-addresses
  - https://syedhasan010.medium.com/aws-instance-metadata-service-a-quick-refresher-4b61ed9af23a
- IMDSはv1とv2がある
- AWS SDKはデフォルトでv2が使われる
  - https://docs.aws.amazon.com/ja_jp/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
- **IMDSv2ではIMDSのPUTリクエストに対するレスポンスのHop数(TTL)をデフォルトでは１に制限している。**
  - https://docs.aws.amazon.com/ja_jp/AWSEC2/latest/UserGuide/instance-metadata-v2-how-it-works.html  
    > デフォルトで、PUT リクエストに対するレスポンスには IP プロトコルレベルで 1 のレスポンスホップリミット (有効期限) があります。より大きなホップリミットが必要な場合は、modify-instance-metadata-options AWS CLI コマンドを使って調整できます。
  - https://docs.aws.amazon.com/ja_jp/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html#imds-considerations  
    > ##### ・コンテナ環境では、ホップ制限を 2 に設定することをお勧めします。
    > AWS SDK はデフォルトで IMDSv2 コールを使用します。IMDSv2 呼び出しに応答がない場合、SDK は呼び出しを再試行し、それでも失敗した場合は、IMDSv1 を使用します。これにより、特にコンテナ環境では、遅延が発生することがあります。コンテナ環境では、ホップ制限が 1 の場合、コンテナへの到達は余分なネットワークホップと見なされるため、IMDSv2 応答は返されません。IMDSv1 へのフォールバックプロセスとその結果として生じる遅延を回避するために、コンテナ環境でホップ制限を 2 に設定することをお勧めします。
  - つまり、コンテナ(Docker単体/k8s Pod)からAWS SDKを使ってIMDSへPUTメソッドでその後GETメソッドでIMDSからメタデータを取得するために使用するために必要なトークン値を取得するリクエストに対して、IMDSが返すレスポンスのHop limit(TTL)が1に設定されていて、Linux BridgeでTTLが減らされてコンテナまでにパケットが届かない
  - 解決策としては以下２つくらい？
    1. Dockerコンテナ/k8s podがホストのネットワークを使うようにする
    2. インスタンスメタデータオプションでホップ数を引き上げる  
  - この問題に関するGithub issue
    - https://github.com/aws/aws-sdk-go/issues/2972
