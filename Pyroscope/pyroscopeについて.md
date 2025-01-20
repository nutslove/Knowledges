# install
- Dockerで動かす場合
  - https://grafana.com/docs/pyroscope/latest/get-started/
- Helmでk8s上で動かす場合
  - https://grafana.com/docs/pyroscope/latest/deploy-kubernetes/helm/

# 実行方法
## Java
- https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/
- コードを修正せずに`javaagent`として実行する方法とjavaコードを修正する方法がある
### `javaagent`として実行する方法
- 以下URLから`pyroscope.jar`をダウンロード
  - https://github.com/pyroscope-io/pyroscope-java/releases
- 環境変数を設定して`javaagent`に`pyroscope.jar`を指定  
  ```shell
  export PYROSCOPE_APPLICATION_NAME=my.java.app
  export PYROSCOPE_SERVER_ADDRESS=http://pyroscope-server:4040

  java -javaagent:pyroscope.jar -jar app.jar
  ```