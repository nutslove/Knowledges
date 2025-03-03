# finalizersとは
- **https://kubernetes.io/ja/docs/concepts/overview/working-with-objects/finalizers/**
  > ファイナライザーは、削除対象としてマークされたリソースを完全に削除する前に、特定の条件が満たされるまでKubernetesを待機させるための名前空間付きのキー
  > Kubernetesにファイナライザーが指定されたオブジェクトを削除するように指示すると、
  > Kubernetes APIはそのオブジェクトに`.metadata.deletionTimestamp`を追加し削除対象としてマークして、ステータスコード`202`(HTTP "Accepted")を返します。 
  > コントロールプレーンやその他のコンポーネントがファイナライザーによって定義されたアクションを実行している間、対象のオブジェクトは終了中の状態のまま残っています。
  > それらのアクションが完了したら、そのコントローラーは関係しているファイナライザーを対象のオブジェクトから削除します。 
  > **`metadata.finalizers`フィールドが空になったら、Kubernetesは削除が完了したと判断しオブジェクトを削除します。**

## finalizersはどのように動作するか
> マニフェストファイルを使ってリソースを作るとき、`metadata.finalizers`フィールドの中でファイナライザーを指定することができます。 リソースを削除しようとするとき、削除リクエストを扱うAPIサーバーは`finalizers`フィールドの値を確認し、以下のように扱います。
>
> - 削除を開始した時間をオブジェクトの`metadata.deletionTimestamp`フィールドに設定します。
> - `metadata.finalizers`フィールドが空になるまでオブジェクトが削除されるのを阻止します。
> - ステータスコード`202`(HTTP "Accepted")を返します。
> ファイナライザーを管理しているコントローラーは、オブジェクトの削除がリクエストされたことを示す`metadata.deletionTimestamp`がオブジェクトに設定されたことを検知します。 するとコントローラーはリソースに指定されたファイナライザーの要求を満たそうとします。 ファイナライザーの条件が満たされるたびに、そのコントローラーはリソースの`finalizers`フィールドの対象のキーを削除します。 `finalizers`フィールドが空になったとき、`deletionTimestamp`フィールドが設定されたオブジェクトは自動的に削除されます。管理外のリソース削除を防ぐためにファイナライザーを利用することもできます。
>
> ファイナライザーの一般的な例は`kubernetes.io/pv-protection`で、これは `PersistentVolume`オブジェクトが誤って削除されるのを防ぐためのものです。 `PersistentVolume`オブジェクトをPodが利用中の場合、Kubernetesは `pv-protection` ファイナライザーを追加します。 `PersistentVolume`を削除しようとすると`Terminating`ステータスになりますが、ファイナライザーが存在しているためコントローラーはボリュームを削除することができません。 Podが`PersistentVolume`の利用を停止するとKubernetesは`pv-protection`ファイナライザーを削除し、コントローラーがボリュームを削除します。