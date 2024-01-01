## Seccomp（Secure Computing Mode）
- defaultでUser Space上のプログラムはすべてのsyscallを使える
- seccompはLinux Kernelのセキュリティ機能の一つで、プロセスが実行できるsyscallを制限する
- seccompを使用すると、アプリケーションは「WhiteList」に基づいて、許可されたsyscallのみを実行できる

#### seccompの主要なモード
1. **Strict Mode**
   - これは非常に制限されたモードで、プロセスはread()、write()、_exit()、およびsigreturn()のみのシステムコールを使用できます。これは最初のSeccompの実装であり、非常に制約が強いため、現在ではあまり使用されていません。

2. **Filter Mode**
   - これはより柔軟なモードで、開発者はBPF（Berkeley Packet Filter）を使用して、プロセスによって実行されるシステムコールをカスタマイズできます。Filter Modeでは、プロセスは必要に応じてシステムコールのWhiteList/BlackListを定義でき、セキュリティと機能のバランスを取ることが可能です。
   - WhiteList方式とBlackList方式がある
   - syscallのWhiteListもしくはBlackListが定義されているファイルをprofileという
     - profileを適用して許可/拒否するsyscallやdefault actionを反映する

#### その他
- 使用しているLinuxがseccompに対応しているかは以下のコマンドで確認可能
  - `grep -i seccomp /boot/config-$(uname -r)`
- あるプロセスがどのseccompモードで動いているかは`/proc/<pid>/status`の`Seccomp`項目の値で確認できる
  - `0`: Disabled
  - `1`: Strict Mode
  - `2`: Filter Mode
- DockerはdefaultでseccompをFilter Modeで有効化されている
  - defaultでDockerのseccompのprofileで無効化されているsyscall
    - https://docs.docker.com/engine/security/seccomp/
    - https://matsuand.github.io/docs.docker.jp.onthefly/engine/security/seccomp/
- Kubernetesではdefaultではseccompは無効化(disabled)されている
  - https://kubernetes.io/docs/tutorials/security/seccomp/
  - マニフェストファイル(`spec.securityContext.seccompProfile`)でseccompの有効化することもできる  
    ~~~yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: default-pod
      labels:
        app: default-pod
    spec:
      securityContext:
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: test-container
        image: hashicorp/http-echo:1.0
        args:
        - "-text=just made some more syscalls!"
        securityContext:
          allowPrivilegeEscalation: false
    ~~~

## Capabilities
- CapabilitiesはLinux Kernelの機能でrootユーザの権限を細分化してもの
- https://dockerlabs.collabnix.com/advanced/security/capabilities/  
  > The Linux kernel is able to break down the privileges of the root user into distinct units referred to as capabilities. For example, the CAP_CHOWN capability is what allows the root use to make arbitrary changes to file UIDs and GIDs. The CAP_DAC_OVERRIDE capability allows the root user to bypass kernel permission checks on file read, write and execute operations. Almost all of the special powers associated with the Linux root user are broken down into individual capabilities.
- Dockerでデフォルトで有効化されているcapabilitiesがあって、明示的に追加/除外できる
- seccompとcapabilitiesは別物
- seccompで許可しているsyscallでもcapabilitiesで拒否しているものは操作できない（逆も同じ）