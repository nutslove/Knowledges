- 参考URL
  - https://grafana.com/docs/alloy/latest/set-up/migrate/from-promtail/
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.file/

# Positionファイルとは
- Alloyがログを収集する際に、どこまでログを読み取ったかを記録するためのファイル。
- これにより、Alloyが再起動した場合や、ログの収集が一時的に停止した場合でも、前回の位置からログの収集を再開することができる。

# AlloyでのPositionファイル
- Alloyでは、`loki.source.kubernetes.pod_logs`、`loki.source.journal.dataplane`、`loki.source.kubernetes_events.kubernetes_events`などのログソースコンポーネントごとにディレクトリが作成され、その中にPositionファイルが作成される。つまり、ログソースコンポーネントごとにPositionファイルが管理される。
- positionファイルは、Helmでデプロイした場合、`tmp/alloy`配下に`loki.source.journal.dataplane`や`loki.source.kubernetes.pod_logs`などのディレクトリが作成され、その中にPositionファイルが作成される。
  - ディレクトリは`alloy.storagePath`で変更可能
    - https://github.com/grafana/alloy/tree/main/operations/helm/charts/alloy
- 例  
  ```shell
  alloy@plat-alloy-2mh9g:/$ ls -l /tmp/alloy/
  total 20
  -rw-r--r--. 1 alloy alloy  112 Apr  2 07:18 alloy_seed.json
  drwxr-x---. 2 alloy alloy 6144 Apr  2 12:37 loki.source.journal.dataplane
  drwxr-x---. 2 alloy alloy 6144 Apr  2 12:37 loki.source.kubernetes.pod_logs
  drwxr-x---. 2 alloy alloy 6144 Apr  2 12:37 loki.source.kubernetes_events.kubernetes_events
  drwxr-x---. 2 alloy alloy 6144 Apr  2 07:18 remotecfg

  alloy@plat-alloy-2mh9g:/$ ls -l /tmp/alloy/loki.source.kubernetes.pod_logs/
  total 56
  -rw-------. 1 alloy alloy 53987 Apr  2 12:37 positions.yml
  alloy@plat-alloy-2mh9g:/$ 
  alloy@plat-alloy-2mh9g:/$ ls -l /tmp/alloy/loki.source.kubernetes_events.kubernetes_events/
  total 4
  -rw-------. 1 alloy alloy 73 Apr  2 12:38 positions.yml
  ```

- AlloyをDaemonSetでデプロイしている場合、EFSで共有ストレージを使用することで、複数のPod間でPositionファイルを共有することができる。
  - PV/PVCのマニフェストの例  
    ```yaml  
    apiVersion: v1
    kind: PersistentVolume
    metadata:
      name: alloy-positions-efs-pv
    spec:
      capacity:
        storage: 1Gi
      volumeMode: Filesystem
      accessModes:
        - ReadWriteMany
      persistentVolumeReclaimPolicy: Delete # PVが削除されるだけで、その(EFSの)中のデータは消えない
      storageClassName: efs-sc
      claimRef:
        namespace: monitoring
        name: alloy-positions-pvc
      csi:
        driver: efs.csi.aws.com
        volumeHandle: fs-xxxxxx::fsap-xxxxx  # FileSystem::AccessPoint
    ---
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: alloy-positions-pvc
      namespace: monitoring
    spec:
      accessModes:
        - ReadWriteMany
      storageClassName: efs-sc
      resources:
        requests:
          storage: 1Gi
    ```

  - Helmのvalues.ymlの例  
    ```yaml
    alloy:
      storagePath: /tmp/alloy # これはデフォルト値なので、明示的に設定する必要はない
      mounts:
        varlog: true
        extra:
          - name: alloy-data
            mountPath: /tmp/alloy

    controller:
      volumes:
        extra:
          - name: alloy-data
            persistentVolumeClaim:
              claimName: alloy-positions-pvc
    ```

> [!NOTE]
> DaemonSetでは`volumeClaimTemplates`でDynamic Provisioningを使用して、各PodにPersistent Volume Claimを割り当てることができない。`volumeClaimTemplates`はStatefulSet専用の機能であるため、DaemonSetでは使用できない。