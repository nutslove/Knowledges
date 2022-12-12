### Pluginのinstall
- インストールするPluginを記述したファイルを用意してJenkins Dockerイメージビルド時にそのファイルを読み込ませてインストールすることができる
  - 前のバージョンでは`install-plugins.sh`が用意されていてそれを使って`install-plugins.sh < pluginlist.txt`のようにインストールしていたが、最近のバージョンでは`install-plugins.sh`が無くなり、代わりに`jenkins-plugin-cli`が用意されて`jenkins-plugin-cli -f pluginlist.txt`のようにインストールする  
    ※pluginlist.txtはインストールするPluginリストが書かれているファイル(ファイル名は任意)
- JenkinsでインストールされているPluginの名前やバージョンは`Jenkinsの管理` - `スクリプトコンソール`にて以下を実行したら出力される
  ~~~groovy
  Jenkins.instance.pluginManager.plugins.each{
  plugin -> println("${plugin.getShortName()}, ${plugin.getLongName()}, ${plugin.getVersion()}")
  }
  ~~~

## 各種Plugin
### Pipeline Utility Steps
- 参考URL
  - https://www.jenkins.io/doc/pipeline/steps/pipeline-utility-steps/#readjson-read-json-from-files-in-the-workspace
  - https://plugins.jenkins.io/pipeline-utility-steps/
- このPluginを使うとJSONやYAMLデータを扱うことができる
  - JSONの例
    ~~~groovy
    def props = readJSON file: 'dir/input.json'
    assert props['attr1'] == 'One'
    assert props.attr1 == 'One'

    def props = readJSON text: '{ "key": "value" }'
    assert props['key'] == 'value'
    assert props.key == 'value'

    def props = readJSON text: '[ "a", "b"]'
    assert props[0] == 'a'
    assert props[1] == 'b'

    def props = readJSON text: '{ "key": null, "a": "b" }', returnPojo: true
    assert props['key'] == null
    props.each { key, value ->
        echo "Walked through key $key and value $value"
    }
    ~~~
- `new groovy.json.JsonSlurperClassic().parseText(<JSONデータ>)`でJSONファイルをロードすることもできるが、Jenkinsでは`Script Console`でapproveされてないとエラーになる