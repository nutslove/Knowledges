- 参考URL
  - https://www.jenkins.io/doc/book/pipeline/shared-libraries/
  - https://swet.dena.com/entry/2021/01/18/200000

### Jenkins設定
- `Jenkinsの管理` → `システムの設定` → `Global Pipeline Libraries`

|  設定項目名  |  設定内容  | 備考 |
| ---- | ---- | ---- |
|  Name  |  ライブラリ名.<br>Pipelineでインポートする際に利用  |    |
|  Default version  |  TD  |    |
|  Retrieval method  |  TD  |    |
|  Source Code Management |  TD  |    |
|  Project Repository  |  TD  |    |
|  Credentials  |  TD  |    |
|  Behaviors  |  TD  |    |
|  Library Path (optional)  |  TD  |    |

### Directory構造
- src
  - パッケージに分けて独自のクラスを定義可能
- vars
  - pipeline jobで利用可能な変数定義（.groovy）とヘルプファイル（.txt）  
  - xxx.groovyのファイル名を変数として利用可能でありそのファイル内に定義されているメソッドを {ファイル名}.{メソッド名} で呼び出し可能
- resources
  - groovyではないファイル（xxx.jsonやxxx.sh等）を格納  
  - libraryResourceを利用してメソッド内で利用が可能