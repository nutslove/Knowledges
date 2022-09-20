https://blog.naver.com/alice_k106/221468341565  
https://go-journey.club/archives/7519

### TLS handshake
- 参照URL  
  https://www.cloudflare.com/ja-jp/learning/ssl/what-happens-in-a-tls-handshake/
- TLS handshakeはTLS暗号化を使った通信セッションを始めるプロセス
- TLS handshakeの間通信する二者がメッセージをやり取りして互いを認識し、検証し、使用する暗号化アルゴリズムを決定し、セッションキーについて合意する
  - 使用するTLSのバージョン（TLS 1.0、1.2、1.3など）を指定
  - 使用する暗号スイート[^1]を決定
  - サーバー公開鍵とSSL証明書認証局のデジタル署名を介して、サーバーのIDを認証
  - handshake完了後に対称暗号化[^2]に使うためのセッションキーを生成  
    [^1]:鍵交換や暗号化方式、ハッシュ関数など暗号化通信に使われる各種アルゴリズムの組合せ
    [^2]:対称暗号化方式 ⇒ 共通鍵暗号方式  
    非対称暗号化方式 ⇒ 公開鍵暗号方式
- TLS handshakeはTCP handshakeでTCP接続が開かれた後に行われる
  ![TLS handshake](https://github.com/nutslove/all_I_need/blob/master/Knowledges/SSL(TLS)/image/TLS handshake.jpg)

  #### TLS handshakeの具体的な手順
   使用される鍵交換アルゴリズムの種類と、両側でサポートする暗号スイートによって異なる。<br>ここでは最も頻繁に使われる**RSA鍵交換アルゴリズム**のケースで説明  
  1.  **「Client Hello」メッセージ**  
     クライアントがサーバーに「Hello」というメッセージを送信することによってhandshakeを開始する。このメッセージには、クライアントがサポートするTLSのバージョン、対応する暗号スイート、「クライアントランダム」というランダムなバイト文字列が含まれている。
  2.  **「Server Hello」メッセージ**  
     Client Helloメッセージへの返答として、サーバーがメッセージを送る。このメッセージには、サーバーのSSL証明書、選んだ暗号スイート、サーバーが生成した別のバイト文字列「サーバーランダム」が含まれている。
  3. **認証**  
     クライアントはサーバーのSSL証明書を発行元の認証局に確認する。これにより、サーバーが自称する本人に間違いなく、クライアントはそのドメインの実際の所有者とやりとりしていることが確認される。
  4. **プレマスタシークレット**  
     

### 証明書の構造
1. Root