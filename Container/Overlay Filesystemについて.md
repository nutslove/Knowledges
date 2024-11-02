## OverlayFS
- https://docs.docker.jp/storage/storagedriver/overlayfs-driver.html
- *Union Filesystem*とも呼ばれる
- **Overlay Filesystemは Lower layer、Upper layer、Overlay layer の３つのLayerで構成されている**
- **Lower layerとUpper layerのディレクトリ/データを合わせて（マージして）Overlay layerのディレクトリ/データが構成される**  
  ![overlayfs](./image/overlayfs.jpg)  
  - **Lower layerとUpper layerに同じパスのファイルやディレクトリが存在する場合、Upper layerの内容が優先され、Overlay layer（マージされたファイルシステム）に反映される**
- **OverlayFSは変更をUpper layerに記録し、Lower layerを不変の基盤として扱う設計になっている**
- **Upper layerは通常、読み書き可能であり、Lower layerは読み取り専用として扱われる**
  - **その結果、ファイルの追加、更新、削除はすべてUpper layerで行われ、同じパスのLower layerのファイルやディレクトリは隠蔽される**
  - **Overlay layer上にあるLower layerからのファイルを更新/削除しても、Lower layer上のファイルをそのままで、Upper layer上で反映される**
- Lower layerに複数のディレクトリを指定することもできる

### Layerとは
- それぞれのLayerはrootファイルシステムとデータを持っている
- Layerごとに独立したファイルシステムのスナップショット（断面）を持つ
  - Dockerの場合、`/var/lib/docker/overlay2/<レイヤーID>`に各レイヤーのファイルシステム断面が保持される
- 各レイヤーが積み重なることで最終的なイメージのファイルシステムが構築される
- Dockerfileで各命令は命令ごとにレイヤーを作成する
  - **レイヤーを作成するのは`FROM`、`COPY`、`ADD`、`RUN`で、`EXPOSE`、`WORKDIR`、`ENV`、`CMD`、`ENTRYPOINT`はレイヤーを作成しない**

## なぜOverlay Filesystemが必要なのか
#### 1. レイヤー構造の活用
コンテナイメージは複数のレイヤーから構成されており、overlayファイルシステムはこれらのレイヤーを一つの統一されたファイルシステムとしてマウントする。

#### 2.ストレージの節約
同じベースイメージを共有する複数のコンテナが存在する場合、overlayファイルシステムを使うことで、重複するデータを保存する必要がなくなり、ストレージの使用量を大幅に削減できる。

#### 3. コピーオンライト（Copy-on-Write）
overlayファイルシステムはコピーオンライトの仕組みを提供する。これにより、ファイルが変更された場合にのみ、その差分が保存され、未変更のファイルは元のレイヤーから読み取られるため、効率的。

#### 4. 高速なコンテナ起動
ファイルシステムのレイヤー化とコピーオンライトにより、新しいコンテナを起動する際に全てのデータをコピーする必要がないため、コンテナの起動が高速化される。

#### 5. 簡単なイメージの更新と配布
レイヤーごとにイメージを管理できるため、更新や配布が容易になる。変更があったレイヤーだけを再配布すればよいので、ネットワーク帯域の節約にもなる。