## AppArmorとは
- Linux用のsecurity moduleであり、プログラムに対して特定のリソースへのアクセスを許可または拒否するためのprofileを使用する
  - seccompと同様にprofileで何を適用するか管理(定義)する
- プログラムごとにprofileを作成し、適用する
  - デフォルトのAppArmor Profileディレクトリは`/etc/apparmor.d/`
  - `/etc/apparmor.d/disable/`ディレクトリ配下のprofileは適用されない
- https://apparmor.net/
- AppArmorには３つのmodeがある
  - **enforce**
    - ruleを強制
  - **complain**
    - 制約なし、ログに出力のみ
  - **unconfined**
    - 制約なし、ログへの出力もなし
- `aa-status`でAppArmorのload状態を確認できる
- KubernetesのPodでもAppArmorを使える
  - https://kubernetes.io/docs/tutorials/security/apparmor/
  - 2024/01/02の時点でまだBeta
  - `metadata.annotations`配下に`container.apparmor.security.beta.kubernetes.io/<AppArmorを適用するPod名>: localhost/<適用するAppArmor名>`を記述する  
    ~~~yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: hello-apparmor
      annotations:
        # Tell Kubernetes to apply the AppArmor profile "k8s-apparmor-example-deny-write".
        container.apparmor.security.beta.kubernetes.io/hello: localhost/k8s-apparmor-example-deny-write
    spec:
      containers:
      - name: hello
        image: busybox:1.28
        command: [ "sh", "-c", "echo 'Hello AppArmor!' && sleep 1h" ]
    ~~~