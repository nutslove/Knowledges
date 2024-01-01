## seccomp（Secure Computing Mode）
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

#### その他
- 使用しているLinuxがseccompに対応しているかは以下のコマンドで確認可能
  - `grep -i seccomp /boot/config-$(uname -r)`
- あるプロセスがどのseccompモードを使っているかは`/proc/<pid>/status`の`Seccomp`項目の値で確認できる
  - `0`: Disabled
  - `1`: Strict Mode
  - `2`: Filter Mode
- DockerはdefaultでseccompをFilter Modeで有効化されている
  - defaultでDockerのseccompのprofileで無効化されているsyscall
    - https://docs.docker.com/engine/security/seccomp/

## Capabilities
- CapabilitiesはLinux Kernelの機能でrootユーザの権限を細分化してもの
- https://dockerlabs.collabnix.com/advanced/security/capabilities/  
  > The Linux kernel is able to break down the privileges of the root user into distinct units referred to as capabilities. For example, the CAP_CHOWN capability is what allows the root use to make arbitrary changes to file UIDs and GIDs. The CAP_DAC_OVERRIDE capability allows the root user to bypass kernel permission checks on file read, write and execute operations. Almost all of the special powers associated with the Linux root user are broken down into individual capabilities.
- Dockerでデフォルトで有効化されているcapabilitiesがあって、明示的に追加/除外できる
- seccompとcapabilitiesは別物
- seccompで許可しているsyscallでもcapabilitiesで拒否しているものは操作できない（逆も同じ）