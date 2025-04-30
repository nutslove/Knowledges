- https://grafana.com/docs/k6/latest/extensions/build-k6-binary-using-go/
- https://github.com/grafana/xk6

## 概要
- k6を使ってLokiに対して、Write/Read両方の負荷試験ができる

## 手順
- 参考URL
  - https://grafana.com/docs/loki/latest/send-data/k6/
  - https://github.com/grafana/xk6-loki
- まず以下で`xk6`をインストールする必要がある  
  ```shell
  go install go.k6.io/xk6/cmd/xk6@latest
  ```
  - 問題なくインストールできていれば`which xk6`でxk6のパスが確認できる（`GOPATH`の設定が必要）
- その後、`xk6-loki`のextensionをインストールする（ビルドに成功すると`k6`バイナリが生成される）
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
- テスト仕様を記載する`<任意の名前>.js`を作成
  - 書き方
    - https://github.com/grafana/xk6-loki
    - https://github.com/grafana/xk6-loki/tree/main/examples
  - 例  
    ```javascript
    import loki from 'k6/x/loki';

    const timeout = 5000; // ms
    const labels = loki.Labels({
      "format": ["logfmt"], // must contain at least one of the supported log formats
      "os": ["linux"],
      "source": ["k6"],
    });

    const conf = loki.Config("http://fake@10.1.2.107:3100", timeout, 1.0, {}, labels);
    const client = loki.Client(conf);

    export default () => {
       client.pushParameterized(2, 512*1024, 1024*1024);
    };
    ```
    - tenantは`loki.Config`のURLの前に`<テナント名>@`をつける
    - Labelも`loki.Config`の最後の引数で指定できる
- `k6 run <任意の名前>.js`でテストを実行