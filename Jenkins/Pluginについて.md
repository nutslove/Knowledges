- インストールするPluginを記述したファイルを用意してJenkins Dockerイメージビルド時にそのファイルを読み込ませてインストールすることができる
  - 前のバージョンでは`install-plugins.sh`が用意されていてそれを使って`install-plugins.sh < pluginlist.txt`のようにインストールしていたが、最近のバージョンでは`install-plugins.sh`が無くなり、代わりに`jenkins-plugin-cli`が用意されて`jenkins-plugin-cli -f pluginlist.txt`のようにインストールする  
    ※pluginlist.txtはインストールするPluginリストが書かれているファイル(ファイル名は任意)
- JenkinsでインストールされているPluginの名前やバージョンは`Jenkinsの管理` - `スクリプトコンソール`にて以下を実行したら出力される
  ~~~groovy
  Jenkins.instance.pluginManager.plugins.each{
  plugin -> println("${plugin.getShortName()}, ${plugin.getLongName()}, ${plugin.getVersion()}")
  }
  ~~~