### 作成手順
- 以下のコマンドで`kubernetes.io/dockerconfigjson` Typeの`regcred`という名前でSecretを作成（emailは省略可）  
  ```
  kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
  ```
- Pod/Deploymentのマニフェストファイルの`spec.imagePullSecrets`フィールドに作成したSecretを指定  
  ```yaml
  apiVersion: v1
  kind: Pod
  metadata:
    name: private-reg
  spec:
    containers:
    - name: private-reg-container
      image: <your-private-image>
    imagePullSecrets:
    - name: regcred
  ```

- 参照URL
  - https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
  - https://stackoverflow.com/questions/64066461/kubernetes-failed-to-pull-image-no-basic-auth-credentials