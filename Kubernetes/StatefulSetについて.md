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
- StatefulSetでは、各Podに対して一意の識別子（**ordinal index**）が割り当てられる。例えば、`my-statefulset` という名前のStatefulSetで3つのレプリカを持つ場合、Podの名前は以下のようになる  
  ```
  my-statefulset-0
  my-statefulset-1
  my-statefulset-2
  ```
- この**Pod名（ordinal index）をキーとして、PersistentVolumeClaim (PVC) と PersistentVolume (PV) がマッピングされる**ため、Podが削除・再作成されても同じPVがアタッチされる。

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
2. 例えば、`my-statefulset-0` が起動すると、自動的に`my-volume-my-statefulset-0` (`volumeClaimTemplates`の`metadata.name` + Pod名 + 番号) というPVCが作成される
3. `PersistentVolumeClaim`が作成されるとき、`StorageClass`の設定に基づいて`PersistentVolume`が動的にプロビジョニングされる
4. `PersistentVolume` (PV) はこのPVCにバインドされ、PodはそのPVを利用する
5. `my-statefulset-0` が削除された後、再作成されても `my-volume-my-statefulset-0` のPVCが存続しているため、同じPVがマウントされる