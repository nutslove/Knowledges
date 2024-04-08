- https://access.redhat.com/documentation/ja-jp/openshift_container_platform/4.6/html/post-installation_configuration/using-machineconfigs-to-change-machines
- **`MachineConfig`で設定したファイルなどを手動で修正してはいけない。`MachineConfigPool`がチェックしてて手動で修正すると`MachineConfig`の設定が正常に反映されなかったりする**

## MachineConfig
- OpenShiftにおけるノードの設定を定義するためのリソースで、これを使用することでノードのOSレベルの設定を宣言的に管理することができる。
- `MachineConfig`の主な役割
  - ノードの設定を定義する
    - `MachineConfig`を使って、ノードのファイルシステム、systemdユニット、ネットワーク設定などを定義できる。これにより、ノードの構成を一元的に管理できる。
  - 設定の適用を自動化する
    - `MachineConfig`を更新すると、それに関連付けられたノードが自動的に再起動され、新しい設定が適用される。これにより、ノードの設定変更を自動化できる。
  - 設定のバージョン管理
    - `MachineConfig`はバージョン管理されており、変更履歴を追跡できる。これにより、設定の変更を追跡し、必要に応じてロールバックすることができる。

## MachineConfigPoolとMachineConfig
- `MachineConfigPool`で管理対象のノード、設定を適用するノードを定義
  - **`spec.nodeSelector`**
    - **`MachineConfigPool`が管理するノード（Machine）を指定。つまり、`spec.nodeSelector`は`MachineConfigPool`が "このラベルを持つノードは私の管理下にある" と識別するために使われる。**
  - **`spec.machineConfigSelector`**
    - **`MachineConfigPool`によって適用されるべき`MachineConfig`オブジェクトを特定するために使用される。これは、`MachineConfig`オブジェクトのラベルに基づいて、どの設定がこのプールに属するノードに適用されるべきかを選択するために使われる。`spec.machineConfigSelector`は`MachineConfigPool`が "このラベルを持つ`MachineConfig`は私が管理するノードに適用する" と識別するために使われる。**
- `MachineConfig`の変更を適用すると、MachineConfigPoolに属するノードは順次再起動されて新しい設定が反映される
- `MachineConfigPool`の例
  ```yaml
  apiVersion: machineconfiguration.openshift.io/v1
  kind: MachineConfigPool
  metadata:
    name: worker
  spec:
    machineConfigSelector:
      matchExpressions:
        - {key: machineconfiguration.openshift.io/role, operator: In, values: [worker]}
    nodeSelector:
      matchLabels:
        node-role.kubernetes.io/worker: "" --> node-role.kubernetes.io/workerラベルを持つノードを指定
  ```
- `MachineConfig`の例
  ```yaml
  apiVersion: machineconfiguration.openshift.io/v1
  kind: MachineConfig
  metadata:
    name: worker-config
    labels:
      machineconfiguration.openshift.io/role: worker
  spec:
    config:
      ignition:
        version: 3.2.0
      storage:
        files:
        - path: /etc/my-config-file
          mode: 0644
          contents:
            source: data:text/plain;base64,SGVsbG8gV29ybGQhCg==
  ```