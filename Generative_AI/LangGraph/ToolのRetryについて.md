- 参考URL
  - https://docs.langchain.com/oss/python/langgraph/use-graph-api#add-retry-policies
  - https://github.com/langchain-ai/langgraph/blob/main/libs/langgraph/langgraph/types.py#L100C7-L100C18  
    ```python
    class RetryPolicy(NamedTuple):
        """Configuration for retrying nodes.

        !!! version-added "Added in version 0.2.24"
        """

        initial_interval: float = 0.5
        """Amount of time that must elapse before the first retry occurs. In seconds."""
        backoff_factor: float = 2.0
        """Multiplier by which the interval increases after each retry."""
        max_interval: float = 128.0
        """Maximum amount of time that may elapse between retries. In seconds."""
        max_attempts: int = 3
        """Maximum number of attempts to make before giving up, including the first."""
        jitter: bool = True
        """Whether to add random jitter to the interval between retries."""
        retry_on: (
            type[Exception] | Sequence[type[Exception]] | Callable[[Exception], bool]
        ) = default_retry_on
        """List of exception classes that should trigger a retry, or a callable that returns `True` for exceptions that should trigger a retry."""
    ```

## `RetryPolicy`でリトライ設定を行う方法
- 基本的な例  
  ```python
  from langgraph.types import RetryPolicy

  builder.add_node(
      "node_name",
      node_function,
      retry_policy=RetryPolicy(),
  )
  ```
- `max_attempts`でリトライ回数の指定可能
- `retry_on`でリトライ対象のExceptionを指定可能