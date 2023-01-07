

Port number『7946』 is what loki uses to communicate with other members of the ring to share state



『Loki: Internal Server Error. 500. too many unhealthy instances in the ring』
I encountered this. You could see your ring status through (replace with your host):
http://loki:3100/ring
There you can "forget" the unhealthy instance and it should work.
Having said that, you could have this option in your loki config under:
common:
    ring:
        autoforget_unhealthy: true
ingester:
    lifecycler:
        readiness_check_ring_health: false





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
