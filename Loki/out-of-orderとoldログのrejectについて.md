- 参考URL
  - **https://grafana.com/blog/2024/01/04/the-concise-guide-to-loki-how-to-work-with-out-of-order-and-older-logs/**  
    ![alt text](image.png)
  - https://grafana.com/docs/loki/latest/configure/#accept-out-of-order-writes  
    > How far into the past accepted out-of-order log entries may be is configurable with `max_chunk_age`. `max_chunk_age` defaults to 2 hour. Loki calculates the earliest time that out-of-order entries may have and be accepted with  
    > ```shell
    > time_of_most_recent_line - (max_chunk_age/2)
    > ```
    > Log entries with timestamps that are after this earliest time are accepted. Log entries further back in time return an out-of-order error.
    >
    > **For example, if `max_chunk_age` is 2 hours and the stream `{foo="bar"}` has one entry at `8:00`, Loki will accept data for that stream as far back in time as `7:00`. If another log line is written at `10:00`, Loki will accept data for that stream as far back in time as `9:00`.**