## StatefulSetの特徴
1. Pod名の末尾に0から始まる数字が付く
   - Pod再作成などでもPod名は変わらない
   - `**-0`、`**-1`、`**-N`のようにカウントしていく
2. Pod名による名前解決で同じPodへのアクセスを実現
   - 「Headless Servicesについて.md」参照
3. Pod作成/削除/更新時、1つずつ完全にReady/Terminatedされた後に次のPodが作成/削除される
4. PodごとにPVCが作成され、１回特定のPVとバインドされた後はずっと同じPVを使い続ける
   - StatefulSetの`spec.volumeClaimTemplates`フィールドに`accessModes`や必要な容量などを指定
   - StatefulSetによって作成されたPVCはStatefulSetと独立したライフサイクルを持ち、（defaultでは）StatefulSetが削除されてもPVC（＋Dynamic Provisioningの場合、PVCによって作成されたPV）は削除されない
     - `spec.persistentVolumeClaimRetentionPolicy`フィールドでStatefulSetが削除された場合・スケールダウンした場合のPVCの挙動（defaultでは`Retain`）を設定できる
     - https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/

### StatefulSetのPodとPVの紐づけ
- StatefulSetでは、各PodにPersistentVolumeClaim（PVC）が`volumeClaimTemplates`に基づいて作成され、PVCの名前は `<volumeClaimTemplate名>-<StatefulSet名>-<ordinal>` という形式で命名される
- 一度PVCがPVにバインドされると、Podが再起動・再スケジュールされても同じPVCが再利用されるため、結果的に同じPVにマウントされる
- これにより、StatefulSetの各Podは一貫したストレージを持ち、データの永続性が確保される

#### PVCのライフサイクル
- Podを削除してもPVCは削除されない
  - StatefulSetのPodが削除されても、対応するPVCは自動削除されない。これは意図的な設計で、データ保護のため
- StatefulSet自体を削除してもPVCは残る
  - 手動でPVCを削除しない限り、PVとのバインドは維持される

#### 具体的な流れ
1. 例えば、以下のようなStatefulSet内の`volumeClaimTemplates`があるとする  
  ```yaml
  volumeClaimTemplates:
  - metadata:
      name: my-volume
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
  ```
2. 例えば、`my-statefulset-0` が起動すると、自動的に`my-volume-my-statefulset-0` (`<volumeClaimTemplates.metadata.name>-<statefulset名>-<ordinal番号>`) というPVCが作成される
3. `PersistentVolumeClaim`が作成されるとき、`StorageClass`の設定に基づいて`PersistentVolume`が動的にプロビジョニングされる
4. `PersistentVolume` (PV) はこのPVCにバインドされ、PodはそのPVを利用する
5. `my-statefulset-0` が削除された後、再作成されても `my-volume-my-statefulset-0` のPVCが存続しているため、同じPVがマウントされる