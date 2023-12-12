## CORS（Cross-Origin Resource Sharing）
- **https://developer.mozilla.org/ko/docs/Web/HTTP/CORS**
- SOP(Single Origin Policy)が異なるOrigin間のデータリクエストを禁止するもので、CORSがそれを許可するためのもの
- CORSがある理由
  - 信頼できないサイト(URL)からきたリクエストをブロックするため
  - 例えばセッションIDのためのCookieがブラウザに保存されている状態で、フィッシングサイトに騙されて悪意のあるサイトからJavascriptがブラウザにロードされて、そのJavaScriptがセッションIDのCookieを悪用してyahooなどのセッションが保持されているサイトからデータを取得たり更新したりすることができるため、事前に許可されているサイト(URL)からのみリクエストを許可するもの
- SOPによってCORSエラーを出す主体は **ブラウザ(e.g. Chrome)**
  - ブラウザの開発者ツールから以下のようなCORSエラーを確認できる
    ~~~
    Access to fetch at 'http://localhost:8000/' from origin 'http://localhost:8080' has been blocked by CORS policy: Response to preflight request doesn't pass access control check: No 'Access-Control-Allow-Origin' header is present on the requested resource. If an opaque response serves your needs, set the request's mode to 'no-cors' to fetch the resource with CORS disabled.
    ~~~
- 通常、ブラウザは異なるオリジン(サイト/URL)からのスクリプト(Javascript)によるリクエストを制限するけど、サーバーは特定のヘッダーを使用して、特定のオリジンからのリクエストを許可することができる。
例えば、あるサイトが別のオリジン（例えばAPIサーバー）からのリクエストを許可したい場合、その **サーバーはCORSヘッダー（`Access-Control-Allow-Origin`など）を含むレスポンスを返すことで、ブラウザに対してそのリクエストを許可することを伝える。**
- `GET`以外のメソッド(e.g. `POST`)は最初にブラウザーが`OPTIONS`リクエストメソッドを用いて、あらかじめリクエストの「**preflight**」 (サーバーから対応するメソッドの一覧を収集すること) を行い、サーバーの「認可」のもとに実際のリクエスト(e.g. `POST`)を送信する
  - なのでサーバ側で`POST`メソッドに対しては`OPTIONS`メソッドに対してもCORS設定をする必要がある