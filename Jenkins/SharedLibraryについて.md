- 参考URL
  - https://www.jenkins.io/doc/book/pipeline/shared-libraries/
  - https://swet.dena.com/entry/2021/01/18/200000

### Jenkins設定
- `Jenkinsの管理` → `システムの設定` → `Global Pipeline Libraries`

|  設定項目名  |  設定内容  |  必須/任意  |  設定例  |  備考  |
| ---- | ---- | ---- | ---- | ---- |
|  Name  |  ライブラリ名.<br>Pipelineでインポートする際に利用  |  必須  |  id-ope  |    |
|  Default version  |  Gitのブランチ名やタグを指定  |  任意  |  master  |    |
|  Retrieval method  |  利用するSCM(Source Code Management)  |  必須  |  Modern SCM  |    |
|  Source Code Management |  利用するSCMサービス  |  必須  |  Git  |    |
|  Project Repository  |  Git URL  |  必須  |  HTTPS/SSH Git URL  |    |
|  Credentials  |  Git認証に使うCredential  |  必須  |  設定済みのCredentialから選択  |    |
|  Behaviors  |  Gitに対する動作  |  必須  |  Discover branches  |    |
|  Library Path (optional)  |  src,resources,varsディレクトリがあるパス  |  任意  |  shared-libraries  |    |

### Directory構造
- src
  - パッケージに分けて独自のクラスを定義可能
- vars
  - pipeline jobで利用可能な変数定義（.groovy）とヘルプファイル（.txt）  
  - xxx.groovyのファイル名を変数として利用可能でありそのファイル内に定義されているメソッドを {ファイル名}.{メソッド名} で呼び出し可能
- resources
  - groovyではないファイル（xxx.jsonやxxx.sh等）を格納  
  - libraryResourceを利用してメソッド内で利用が可能

### 使い方