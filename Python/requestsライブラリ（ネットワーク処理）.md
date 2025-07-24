- `requests`は、内部で`urllib3`を使用しており、HTTPリクエストの送信やレスポンスの処理を行う。
- `requests.post()`メソッド単独ではリトライのためのパラメータは提供されていないが、`urllib3.util.retry`モジュールを使用して、リトライの設定を行うことができる。
- 例  
  ```python
  import requests
  from requests.adapters import HTTPAdapter
  from urllib3.util.retry import Retry

  def get_bearer_token_with_retry(client: str) -> str:
      # 재시도 정책 설정
      retry_strategy = Retry(
          total=3,  # 총 재시도 횟수
          status_forcelist=[429, 500, 502, 503, 504],  # 재시도할 HTTP 상태 코드
          backoff_factor=1,  # 재시도 간 대기 시간 (1, 2, 4초...)
          allowed_methods=["POST"]  # 재시도를 허용할 HTTP 메서드
      )
      
      # 세션 생성 및 어댑터 설정
      session = requests.Session()
      adapter = HTTPAdapter(max_retries=retry_strategy)
      session.mount("http://", adapter)
      session.mount("https://", adapter)
      
      # 기존 코드...
      response = session.post(token_endpoint, headers=headers, data=data, timeout=10)
      return response.json()["access_token"]
  ```
