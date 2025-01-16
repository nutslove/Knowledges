

Port number『7946』 is what loki uses to communicate with other members of the ring to share state






My org is using grafana managed loki in prod. and currently we are facing Too Many Outstanding Requests  on different loki panels.
I want to  increase the max_outstanding_requests_per_tenant, but i am not getting this configuration in grafana loki datasource configuration.

we have asked the grafana cloud loki team to set the following config:
They haven’t made the changes yet.
query_scheduler:
  max_outstanding_requests_per_tenant: 4096
frontend:
  max_outstanding_per_tenant: 4096
query_range:
  parallelise_shardable_queries: true

hi,just wanted you to know it works like a charm!, thank you so much!
