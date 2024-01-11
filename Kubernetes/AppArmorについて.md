## AppArmorとは
- https://gitlab.com/apparmor/apparmor/-/wikis/home
  > AppArmor is an effective and easy-to-use Linux application security system. AppArmor proactively protects the operating system and applications from external or internal threats, even zero-day attacks,by enforcing good behavior and preventing even unknown application flaws from being exploited. AppArmor security policies completely define what system resources individual applications can access, and with what privileges. A number of default policies are included with AppArmor, and using a combination of advanced static analysis and learning-based tools, AppArmor policies for even very complex applications can be deployed successfully in a matter of hours.
- Ubuntu 7.10以降、Ubuntuにデフォルトで含まれている重要なセキュリティ機能
  - 主にDebian系のLinuxディストリビューション、特にUbuntuで広く使われている
  - RedHat系のLinuxではSELinux(Security-Enhanced Linux)が使われている
- Profileでルールを定義し、各Profileは、特定のプログラムやプロセスがアクセスできるリソース（ファイル、ディレクトリ、ネットワークインターフェースなど）を定義し、システムへの不正なアクセスや悪意のある活動を防ぐ
  - 1 Profile = 1 Process
    - 各Profileは特定のプロセスに割り当てられる
  - ProfileをAppArmorサービスによって解析され、適切なセキュリティポリシーがKernelのセキュリティモジュールに適用される
  - KernelはAppArmor Profileに基づいてアクセス権をチェックし、Profileに定義されたルールに従ってアクセスを許可または拒否する
- ３つのmodeがある
  - **enforce**
    - ruleに違反する行為を拒否する
  - **complain**
    - ruleに違反しても拒否はされないけど、ログに記録される
  - **unconfined**
    - ruleに違反しても拒否もされないし、ログにも残らない