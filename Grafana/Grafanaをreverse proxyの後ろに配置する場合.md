- 参考URL
  - https://grafana.com/tutorials/run-grafana-behind-a-proxy/

- **複数のOrgを使っている場合、`root_url`にGrafanaに接続するためのReverse ProxyのIPとPortを設定する必要がある**
  - 例えばGrafanaのNginx接続アドレスが`10.0.0.8:800`の場合、`root_url`(環境変数は`GF_SERVER_ROOT_URL`)に`http://10.0.0.8:800`を設定する

- nginxのConfigに`proxy_set_header Host $http_host;`を追加する必要がある