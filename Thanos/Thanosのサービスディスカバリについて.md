- https://thanos.io/tip/thanos/service-discovery.md/#dns-service-discovery  
  > # DNS Service Discovery
  > DNS Service Discovery is another mechanism for finding components that can be used in conjunction with Static Flags or File SD. With DNS SD, a domain name can be specified and it will be periodically queried to discover a list of IPs.
  >
  > To use DNS SD, just add one of the following prefixes to the domain name in your configuration:
  > - `dns+` - the domain name after this prefix will be looked up as an A/AAAA query. A port is required for this query type. An example using this lookup with a static flag:
  >
  > ```
  > --endpoint=dns+stores.thanos.mycompany.org:9090
  > ```
  >
  >
  > - `dnssrv+` - the domain name after this prefix will be looked up as a SRV query, and then each SRV record will be looked up as an A/AAAA query. You do not need to specify a port as the one from the query results will be used. For example:
  > ```
  > --endpoint=dnssrv+_thanosstores._tcp.mycompany.org
  > ```

## `dns+`、`dnssrv+`
- `dns+` や `dnssrv+` は主に Thanos（およびその周辺エコシステム）が採用しているサービスディスカバリの記法
- この `dns+` / `dnssrv+` というプレフィックス記法は、Thanos が内部で使っている Go のライブラリ（もともと Thanos プロジェクトから生まれた github.com/thanos-io/thanos/pkg/discovery/dns）が定義しているもの

### `dns+`
- `dns+` は DNS の **A/AAAA レコード** を使ってサービスディスカバリを行う
- IPアドレスの一覧は返ってくるが、ポート番号は含まれないため、ポート番号を明示的に指定する必要がある
- 例: `dns+stores.thanos.mycompany.org:9090`

### `dnssrv+`
- `dnssrv+` は DNS の **SRV レコード** を使ってサービスディスカバリを行う
- SRVレコードは「どのホストの、どのポートでサービスが動いているか」を返すDNSレコードタイプ
- そのため、`dnssrv+` を使う場合はポート番号を指定する必要はない
- 例: `dnssrv+_thanosstores._tcp.mycompany.org`

## Thanosでの利用例
> [!IMPORTANT]  
> - **Querierで、endpointに（Ingesting）Receiverを指定するとき、`--endpoint=thanos-ingesting-receiver.monitoring.svc`のように普通のService名で指定すると、名前解決される１つのReceiverにしか接続されない（かつ、Thanosは複数のReceiverにメトリクスデータが分散される）ため、クエリーを実行するたびに取得されるデータにばらつきが出る。**  
> そのため、**`dns+`（`--endpoint=dns+thanos-ingesting-receiver.monitoring.svc:10901`）か`dnssrv+`を使って、Serviceに紐づいているすべてのPodに接続できるようにする必要がある**

## `dns+`が指定された時の挙動
### 1. `dns+` プレフィックスを解析
`provider.go` の `Resolve` 関数が `dns+` プレフィックスを検出すると、`QType("dns")`（= Aレコード lookup）として処理する

### 2. DNS Aレコードクエリを実行
Headless Serviceに対してAレコードを引くと、**紐づいている全PodのIPアドレスのリスト**が返ってくる

### 3. 全IPをエンドポイントとして登録
`Provider` が解決した全IPアドレスをキャッシュに保存し、`Addresses()` メソッドで返す

### 4. Querierが全エンドポイントに接続
Querierは `Addresses()` から得た全IPを **それぞれ独立したStore APIピア** として認識し、クエリ時に**全ピアに並行してリクエストを投げ、結果をマージ**する

### 補足
つまり、`dns+` 自体は「DNSで名前解決して全IPを取得する」ところまでの責務で、「全Podにクエリを投げてマージする」のはQuerierのロジック。`dns+` がなければDNSは1つのClusterIPしか返さないので、Querierは1つのピアしか認識できない、という構造。