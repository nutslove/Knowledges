# SDK
## v2
- https://langfuse.com/docs/sdk/python/low-level-sdk
> [!CAUTION] 
> - Lambdaなどの短命な環境では、`langfuse.flush()`を呼び出すことで、SDKがバックグラウンドで実行しているリクエストが完了するのを待つ必要がある。これを行わないと、イベントが失われる可能性がある。
> > ### Shutdown behavior
> > The Langfuse SDK executes network requests in the background on a separate thread for better performance of your application. This can lead to lost events in short lived environments like NextJs cloud functions or AWS Lambda functions when the Python process is terminated before the SDK sent all events to our backend.
> >
> > To avoid this, ensure that the langfuse.flush() function is called before termination. This method is waiting for all tasks to have completed, hence it is blocking.
> >
> > ```python
> > langfuse.flush()
> > ```
> - 例  
> ```python
> from langfuse import Langfuse
>
> langfuse = Langfuse()
>
> def attach_tag_to_trace(trace_id: str, *args):
>   try:
>     langfuse.trace(id=trace_id, tags=list(args)) # argsはタプルで渡されるため、listに変換する
>     langfuse.flush()
>   except Exception as e:
>     print(f"Error occurred when attaching tag to trace in Langfuse: {e}")
>     raise e
> ```

# API
- API一覧
  - https://api.reference.langfuse.com/
- API時の認証はTokenを使ったBasic Auth
  - https://langfuse.com/faq/all/api-authentication
  - 例: `curl -u "<public_key>:<secret_key>" https://cloud.langfuse.com/api/public/projects`

## LangGraphのinvoke時にTagを付与
- https://langfuse.com/docs/observability/sdk/python/instrumentation
- `config`の`metadata`の`langfuse_tags`にTagのListを渡す  
  ```python
  final_state = graph.invoke({
    "messages": [("user", error_message)],
    "final_response": Response(analysis_results="", final_command="")
  }, config={
    "recursion_limit": 120,
    "callbacks": [langfuse_handler],
    "run_id": predefined_run_id,
    "metadata": {
      "langfuse_tags": ["<system名など>","<alert_source名など>"] # should be a list of strings
    }
  })
  ```