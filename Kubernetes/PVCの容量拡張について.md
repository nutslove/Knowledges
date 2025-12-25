- http://kubernetes.io/docs/concepts/storage/storage-classes/
- https://kubernetes.io/docs/concepts/storage/persistent-volumes/
- `StorageClass`で`allowVolumeExpansion`を`true`に設定すると、途中でPVCの容量を拡張することができる（デフォルト値は`false`）  
  ```yaml
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: ebs-sc
  provisioner: ebs.csi.aws.com
  allowVolumeExpansion: true  # これが必要
  parameters:
    type: gp3
  ```
  - **`allowVolumeExpansion`を`true`に設定している場合、後から`kubectl patch pvc <pvc-name> -p '{"spec":{"resources":{"requests":{"storage":"10Gi"}}}}'`のようにPVCの容量を拡張できる**