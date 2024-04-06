## インストール手順
- https://microstack.run/docs/single-node-guided
- `apt install openssh-server`でsshをインストーr
- 以下コマンドは一般ユーザで実行
  - `sudo snap install openstack --channel 2023.1`
  - `sunbeam prepare-node-script | bash -x && newgrp snap_daemon`
- rootにスイッチし、以下を実行(事前準備)
  ```shell
  ssh-keygen (デフォルトでEnter)
  cp -p ~/.ssh/id_rsa.pub ~/.ssh/authorized_keys
  chmod 600 ~/.ssh/authorized_keys
  ```
