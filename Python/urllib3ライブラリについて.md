- urllib3はPythonで広く使用されているHTTPクライアントライブラリの一つ。Python標準のurllibライブラリよりも機能が豊富で、接続プール / スレッドセーフ / ヘッダー管理 / クッキー管理 / ファイルアップロード / 自動リダイレクトなど、多くの機能を提供。

## urllib3の`PoolManager`クラス
- HTTPリクエストを送信するための高レベルのインターフェースを提供
- 複数のHTTPリクエストにわたって接続を再利用し、管理する機能を持っていて、ネットワーク遅延が減少し、アプリケーションのパフォーマンスが向上させる。
- `PoolManager`クラスの機能
  1. **接続プール**
     - 同じホストへの複数のリクエストに対して、TCP接続をプールして再利用。これにより、接続のオーバーヘッドが削減され、リクエストのレスポンスタイムが改善される。
  2. **スレッドセーフ**
     - 複数のスレッドから同時にリクエストを安全に発行できる。
       - **データの整合性**
          - 複数のスレッドが同時にHTTPリクエストを行っても、それぞれのリクエストとレスポンスが正しくマッチし、データの不整合が発生しないようにする。つまり、あるスレッドのリクエスト結果が他のスレッドに誤って割り当てられることはない。
       - **排他制御**
          - 必要に応じて、リソースへの同時アクセスを制御し、一度に一つのスレッドだけが特定の操作を行えるようにすることで、データの競合や破壊を防ぐ。
       - **デッドロックの防止**
          - リソースの要求順序などを適切に管理することで、複数のスレッドが互いにリソースを待ち続けるデッドロック状態を防ぐ。
  3. **自動リダイレクトの管理**
     - HTTPリダイレクトに自動的に従い、新しいリクエストを発行する。
  4. **リトライロジック**
     - 失敗したリクエストを自動的に再試行する機能を提供。