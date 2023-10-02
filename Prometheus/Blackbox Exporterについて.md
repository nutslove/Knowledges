- AWSのEC2に対してPingで正常性確認を行っている場合、テストのために一時的にSecurity GroupからICMPを削除する時、Blackbox Exporterも再起動する必要がある。じゃないとセッションが生きててPingが通り続ける。
  - 逆のパターン(拒否→許可)の場合はBlackbox Exporter再起動不要

## Blackbox Exporter Configファイル例
- とりあえず全moduleを定義しておいて、Prometheus側で必要なmoduleを使う
~~~yaml
modules:
  http_2xx_proxy:
    prober: http
    timeout: 5s
    http:
      tls_config:
        insecure_skip_verify: true
      preferred_ip_protocol: "ip4"
      ip_protocol_fallback: false
      proxy_url: "http://192.168.0.6:60080"
      skip_resolve_phase_with_proxy: true
  http_2xx:
    prober: http
    timeout: 5s
    http:
      tls_config:
        insecure_skip_verify: true
      preferred_ip_protocol: "ip4"
      ip_protocol_fallback: false
  http_post_2xx:
    prober: http
    http:
      method: POST
  tcp_connect:
    prober: tcp
  pop3s_banner:
    prober: tcp
    tcp:
      query_response:
      - expect: "^+OK"
      tls: true
      tls_config:
        insecure_skip_verify: false
  ssh_banner:
    prober: tcp
    tcp:
      query_response:
      - expect: "^SSH-2.0-"
      - send: "SSH-2.0-blackbox-ssh-check"
  irc_banner:
    prober: tcp
    tcp:
      query_response:
      - send: "NICK prober"
      - send: "USER prober prober prober :prober"
      - expect: "PING :([^ ]+)"
        send: "PONG ${1}"
      - expect: "^:[^ ]+ 001"
  icmp:
    prober: icmp
    icmp:
      preferred_ip_protocol: "ip4"
~~~

#### tcp接続確認(port番号指定でアクセス確認)
- Prometheus側の設定
  ~~~yaml
  - job_name: 'some_service_tcp_Blackbox_Exporter'
    metrics_path: /probe
    params:
      module:
        - tcp_connect
    static_configs:
      - targets: 
        - 10.10.10.0:5000
        - 10.10.10.1:5005
        - 10.10.10.2:5006
        - 10.10.10.3:5007
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: blackbox-exporter.monitoring.svc.cluster.local:9115
    scrape_interval: 10s
  ~~~

#### proxy超しのHTTPエンドポイントに対して正常性確認を行う時は`http_2xx`ではなく`http_2xx_proxy` moduleを使う
- blackbox exporterの`http_2xx_proxy` module側
  ~~~yaml
  modules:
    http_2xx_proxy:
      prober: http
      timeout: 5s
      http:
        tls_config:
          insecure_skip_verify: true
        preferred_ip_protocol: "ip4"
        ip_protocol_fallback: false
        proxy_url: "http://<ProxyサーバのIP>:<ProxyサーバのPort>"
        skip_resolve_phase_with_proxy: true
  ~~~
- Prometheus側の設定
  ~~~yaml
  - job_name: 'some_system_Blackbox_Exporter_with_proxy'
    metrics_path: /probe
    params:
      module:
        - http_2xx_proxy
    static_configs:
      - targets: 
        - https://some-service.com/test1/test.html
        - https://some-service.com/test1/test2.html
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: blackbox-exporter.monitoring.svc.cluster.local:9115
    scrape_interval: 10s
  ~~~