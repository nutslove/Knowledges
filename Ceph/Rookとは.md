- https://github.com/rook/rook
- https://rook.io/

## Rookとは
- https://rook.io/docs/rook/latest-release/Getting-Started/intro/
  > Rook is an open source cloud-native storage orchestrator, providing the platform, framework, and support for Ceph storage to natively integrate with cloud-native environments.
- RookはCephをKubernetes上にデプロイ/管理などをするためのOperator

### AWS上のOpenShiftにRookを入れるためにやったこと（とりあえずメモ）
うまくいかない・・・
- 以下ページを参考
  - https://rook.io/docs/rook/latest-release/Getting-Started/ceph-openshift/
- https://github.com/rook/rook/tree/release-1.14/deploy/examples配下のものをデプロイ
- デフォルトでは`rook-ceph` namespace上にデプロイされる
- 上記Rookページに書いてあるもの以外に以下をデプロイ
  ```shell
  oc create -f toolbox.yaml
  oc apply -f csi/rbd/storageclass.yaml
  ```
  - `toolbox.yaml`をデプロイすると`rook-ceph-tools-NNN` Podが作成されて、このPod内で以下のように`ceph` CLIが使える  
    ```shell
    oc exec -it rook-ceph-tools-58c6857df4-zr78l -n rook-ceph -- ceph status
    oc exec -it rook-ceph-tools-58c6857df4-zr78l -n rook-ceph -- ceph device ls
    oc exec -it rook-ceph-tools-58c6857df4-zr78l -n rook-ceph -- ceph osd status
    ```