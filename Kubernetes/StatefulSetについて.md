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