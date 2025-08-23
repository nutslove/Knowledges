# UDS（Unix Domain Socket）とは
- 同一マシン内でのプロセス間通信（IPC: Inter-Process Communication）を行うためのソケット通信方式。ネットワークを経由せず、カーネル内で直接通信が行われる。
- UDSのソケットの実体はファイルシステム上にファイルだけど、通常のファイルとは違う**特殊ファイル**
  - ファイルタイプが`s`  
    ```shell
    $ ls -la /var/run/docker.sock
    srw-rw----. 1 root docker 0 Aug 23 10:30 /var/run/docker.sock
    ```
  - ソケットファイルのサイズは常に0バイト
    - データは実際にはファイルに保存されず、カーネル内のバッファで管理される
  - readやwriteシステムコールでは通信できない
    - `cat`や`echo "someting" >`などできない（エラーになる）
## メリット
- ネットワークスタックを経由しないため、TCP/IPより高速
- 同一マシン内でのみ通信可能、外部からアクセス不可（セキュア）
- ファイルシステムの権限でアクセス制御可能
- オーバーヘッドが少ない（ネットワークプロトコルの処理が不要）
## 制約
- 同一マシン内でのみ利用可能（ネットワーク越しの通信はできない）

### `docker.sock`
- DockerクライアントとDockerデーモン間の通信に使われるUnix Domain Socket
- 通常`/var/run/docker.sock`に作成される
- `docker ps`コマンドを実行すると内部的には`/var/run/docker.sock`を通じてDocker daemonと通信し、コンテナ一覧を取得している