# マルチテナントの設定
- https://grafana.com/docs/loki/latest/operations/multi-tenancy/
- defaultでMulti tenantモードが有効になっている
  - `auth_enabled`を`false`に設定すればsingle tenantモードとして動いて、tenant idは`fake`が設定される  
    > When configured with `auth_enabled: false`, Loki uses a single tenant. The `X-Scope-OrgID` header is not required in Loki API requests. The single tenant ID will be the string `fake`.
- HTTPの`X-Scope-OrgID`ヘッダーの値でテナントを識別する

---

# Multi-tenant Queries
- https://grafana.com/docs/loki/latest/operations/multi-tenancy/
- 複数のテナントに渡ってクエリーを投げることができる
- **そのためにはQuerierの設定で`multi_tenant_queries_enabled`を`true`(defaultは`false`)に設定する必要がある**
- `multi_tenant_queries_enabled: true`にせずに、GrafanaでLokiデータセットで`X-Scope-OrgID: A|B`とか設定すると以下のようなエラーが出る  
  ![](./image/multi_tenant_queries_1.jpg)
  - Grafanaのログ  
    ```shell
    errorMessageID=plugin.requestFailureError error="client: failed to call resources: multiple org IDs present"
    ```

---

# テナントごとのLimit設定
- 1つのテナントがリソースを使いすぎるのを防ぐためにテナントごとにLimitを設定することができる
## `frontend`
- https://grafana.com/docs/loki/latest/configure/#frontend
- `max_outstanding_per_tenant`  
  ```yaml
  # Maximum number of outstanding requests per tenant per frontend; requests
  # beyond this error with HTTP 429.
  # CLI flag: -querier.max-outstanding-requests-per-tenant
  [max_outstanding_per_tenant: <int> | default = 2048]
  ```
## `index_gateway`
- https://grafana.com/docs/loki/latest/configure/#index_gateway
- `mode`  
  ```yaml
  # Defines in which mode the index gateway server will operate (default to
  # 'simple'). It supports two modes:
  # - 'simple': an index gateway server instance is responsible for handling,
  # storing and returning requests for all indices for all tenants.
  # - 'ring': an index gateway server instance is responsible for a subset of
  # tenants instead of all tenants.
  # CLI flag: -index-gateway.mode
  [mode: <string> | default = "simple"]
  ```
## `limits_config`
- https://grafana.com/docs/loki/latest/configure/#limits_config  
  > The `limits_config` block configures global and per-tenant limits in Loki. The values here can be overridden in the `overrides` section of the `runtime_config` file
- **`max_queriers_per_tenant`**  
  ```yaml
  # Maximum number of queriers that can handle requests for a single tenant.
  # If set to 0 or value higher than number of available queriers, *all* queriers will handle requests for the tenant. 
  # Each frontend (or query-scheduler, if used) will select the same set of queriers for the same tenant (given that all queriers are connected to all frontends / query-schedulers). 
  # This option only works with queriers connecting to the query-frontend / query-scheduler, not when using downstream URL.
  # CLI flag: -frontend.max-queriers-per-tenant
  [max_queriers_per_tenant: <int> | default = 0]
  ```
- **`max_query_capacity`**  
  ```yaml
  # How much of the available query capacity ("querier" components in distributed
  # mode, "read" components in SSD mode) can be used by a single tenant. Allowed
  # values are 0.0 to 1.0. For example, setting this to 0.5 would allow a tenant
  # to use half of the available queriers for processing the query workload. If
  # set to 0, query capacity is determined by frontend.max-queriers-per-tenant.
  # When both frontend.max-queriers-per-tenant and frontend.max-query-capacity are
  # configured, smaller value of the resulting querier replica count is
  # considered: min(frontend.max-queriers-per-tenant, ceil(querier_replicas *
  # frontend.max-query-capacity)). *All* queriers will handle requests for the
  # tenant if neither limits are applied. This option only works with queriers
  # connecting to the query-frontend / query-scheduler, not when using downstream
  # URL. Use this feature in a multi-tenant setup where you need to limit query
  # capacity for certain tenants.
  # CLI flag: -frontend.max-query-capacity
  [max_query_capacity: <float> | default = 0]
  ```
## `querier`
- https://grafana.com/docs/loki/latest/configure/#querier
- `multi_tenant_queries_enabled`  
  ```yaml
  # When true, allow queries to span multiple tenants.
  # CLI flag: -querier.multi-tenant-queries-enabled
  [multi_tenant_queries_enabled: <boolean> | default = false]
  ```
## `query_scheduler`
- https://grafana.com/docs/loki/latest/configure/#query_scheduler
- `max_outstanding_requests_per_tenant`  
  ```yaml
  # Maximum number of outstanding requests per tenant per query-scheduler.
  # In-flight requests above this limit will fail with HTTP response status code
  # 429.
  # CLI flag: -query-scheduler.max-outstanding-requests-per-tenant
  [max_outstanding_requests_per_tenant: <int> | default = 32000]
  ```
## `overrides`
- https://grafana.com/docs/loki/latest/configure/#runtime-configuration-file
- `overrides`ブロックでテナントごとの値を設定できる  
  ```yaml
  overrides:
    tenant1:
      ingestion_rate_mb: 10
      max_streams_per_user: 100000
      max_chunks_per_query: 100000
    tenant2:
      max_streams_per_user: 1000000
      max_chunks_per_query: 1000000
  ```