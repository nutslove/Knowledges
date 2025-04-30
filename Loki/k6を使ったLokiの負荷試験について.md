- https://grafana.com/docs/k6/latest/extensions/build-k6-binary-using-go/
- https://github.com/grafana/xk6

## 手順
- 参考URL
  - https://grafana.com/docs/loki/latest/send-data/k6/
  - https://github.com/grafana/xk6-loki
- まず以下で`xk6`をインストールする必要がある  
  ```shell
  go install go.k6.io/xk6/cmd/xk6@latest
  ```
  - 問題なくインストールできていれば`which xk6`でxk6のパスが確認できる（`GOPATH`の設定が必要）
- その後、`xk6-loki`のextensionをインストールする
  - 上記のLokiのページでは以下のように`make`でビルドする手順が書いてるけど、以下の`xk6 build`で`github.com/grafana/xk6-loki`を指定するやり方でもビルドできた  
    - `make`  
      ```shell
      git clone https://github.com/grafana/xk6-loki
      cd xk6-loki
      make k6
      ```
    - `xk6 build`  
      ```shell
      xk6 build latest --with github.com/grafana/xk6-loki
      ```