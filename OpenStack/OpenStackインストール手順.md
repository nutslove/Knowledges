# MicroStackによるインストール手順
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
- 以下コマンドでOpenStackを払い出す
  - `sunbeam cluster bootstrap`
- 以下コマンドを実行し、上記URLの通り入力/選択する
  - `sunbeam configure --openrc demo-openrc`
  - demo-openrcファイルが生成されていることを確認
    - horizonの認証情報などが記載されているファイル
- 以下コマンドでhorizonのURLを確認し、上で確認した認証情報でログインする
  - `sunbeam dashboard-url`

# DevStackによるインストール手順
- https://docs.openstack.org/devstack/latest/
- https://cloud5.jp/openstack-install/
- https://stackoverflow.com/questions/42973688/error-opt-stack-logs-error-log-no-such-file-or-directory-devstack-deploy