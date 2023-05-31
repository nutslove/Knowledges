## metrics-generatorを有効にしたらTempoが起動しなくなった件
- 事象
  - Tempo v2.1.1でmonolithic modeで動かして問題なかったけど、`metrics_generator`設定を追加したらpanicが起きてTempoが起動しなくなった
    - Tempo Logs
      ~~~
      github.com/grafana/tempo/modules/generator/registry/registry.go:122 +0x9db
      created by github.com/grafana/tempo/modules/generator/registry.New
      github.com/grafana/tempo/modules/generator/registry/job.go:11 +0x4c
      github.com/grafana/tempo/modules/generator/registry.job({0x29a8920, 0xc00016e140}, 0xc001e46450, 0xc001e46460)
      time/tick.go:24 +0x10f
      time.NewTicker(0x29a8920?)
      goroutine 2376 [running]:
      panic: non-positive interval for NewTicker
      ~~~
    - Configuration
      ~~~
      server:
        http_listen_port: 3200
       
      distributor:
        receivers:
            otlp:
              protocols:
                http:
                grpc:
       
      compactor:
        compaction:
          block_retention: 744h                # configure total trace retention here
       
      storage:
        trace:
          backend: s3
          s3:
            endpoint: s3.ap-northeast-1.amazonaws.com
            bucket: <S3 Bucket Name>
            forcepathstyle: true
            #set to true if endpoint is https
            insecure: true
          wal:
            path: /tmp/tempo/wal         # where to store the the wal locally
          local:
            path: /tmp/tempo/blocks
       
      overrides:
        metrics_generator_processors:
          - span-metrics
       
      metrics_generator:
        ring:
          kvstore:
        processor:
          service_graphs:
          span_metrics:
            intrinsic_dimensions:
            dimensions:
              - "db.statement"
        registry:
        storage:
          path: /opt/tempo/wal
          wal:
          remote_write:
            - url: <Remote Write URL>      
      ~~~
- 原因
  - `metrics_generator.registry`がemptyだとその配下の項目(e.g. `collection_interval`)の設定値がdefault値になるのではなく、全部`0`が設定されるとのこと
    ![](img/registry_trouble.jpg)
- 対処
  - `registry` blockを削除するか、以下のように明示的に`registry`配下の項目を設定する
    ~~~yaml
    metrics_generator:
        registry:
            collection_interval: 15s
            stale_duration: 15m
            max_label_name_length: 1024
            max_label_value_length: 2048
    ~~~
