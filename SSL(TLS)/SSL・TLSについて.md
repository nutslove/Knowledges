##### 参照URL
- https://blog.naver.com/alice_k106/221468341565 （KR）
- https://go-journey.club/archives/7519

##### その他
- PKI (Public Key Infrastructure)とは、公開鍵暗号方式に基づいて、電子署名や相手認証等を実現するための技術基盤
- 現在はTLS 1.3が主流
- 共通鍵暗号方式(対称鍵暗号方式)の方が公開鍵暗号方式(非対称鍵暗号方式)より処理速度が早い。  
  なのでTLS handshakeにて最初に公開鍵暗号方式でサーバ/クライアント間で共通鍵を生成/交換した後、実際のデータのやり取り時は共通鍵を使って暗号化/復号化を行う。
- Root証明書 = CA(認証局)の公開鍵
- 暗号化/復号化は公開鍵/秘密鍵どちらでもできる
  - 公開鍵で暗号化して秘密鍵で復号化、秘密鍵で暗号化して公開鍵で復号化、どちらも可

##### Hash関数
- 一方向性関数
  - inputからoutputを出すことはできるけど、outputからinputを導き出すことはできない(難しい)関数
- いかなる長さのデータを入力しても(inputデータ短くても長くても)固定長の擬似乱数データを出力(output)する関数
- データの完全性(データが改ざんされてないこと)を確認するのに良く使われる
- Hash関数で得られた値をHash値(もしくはMD(Message Digest))という

## ディジタル証明書(サーバ証明書)の発行からディジタル証明書を使った通信の流れ
#### ディジタル証明書(サーバ証明書)の発行
1. まずサーバ(事業者)側で秘密鍵/公開鍵を生成する
2. 使用したいドメイン(ex. amazon.com)と公開鍵をCA(認証局)に送ってディジタル証明書を要請する
3. CAは事業者を検証し本当にその事業者からの要請か確認してから、ディジタル証明書を生成して自身(CA)の秘密鍵でディジタル証明書のHash値を署名(暗号化)する
   - このディジタル証明書には**1.ドメイン(ex. amazon.com)などの情報**と**2.サーバ(事業者)の公開鍵**と**3.CAの秘密鍵で暗号化された1.+ 2.のHash値**が含まれる
   - ディジタル証明書の形式(規格)が`X.509`
   - ディジタル証明書の有効期限は6ヶ月~1年程度
#### ディジタル証明書を使った通信の流れ（TLS handshake）
1. クライアントがWebサイト(ex. amazon.com)に接続するとサーバからディジタル証明書(サーバ証明書)が送られる
2. ディジタル証明書を受け取ったクライアントはブラウザに内蔵されている(またはWindowsに保存されている)該当CAの公開鍵(Root証明書)でディジタル証明書の中のHash値を復号化し、再計算したHash値と一致することを確認することでディジタル証明書が改ざんされてないこととその証明書が該当CAによって署名/発行されたことを確認する
   > **Note**  
   > 各ブラウザには信頼できるCAのリストと各CAの公開鍵(Root証明書)が内蔵されている  
   > ※*各CAから公開鍵(Root証明書)をダウンロードしてブラウザに登録することもできる（オンプレ上のブラウザ等）*
   - *CAの公開鍵(Root証明書)も定期的に再発行される。そうなると既存のRoot証明書は使えないのでRoot証明書を更新する必要がある。ただ、Windows UpdateでWindows内のRoot証明書が更新されるので普通は気にすることはないけど、Internetに繋がってない(Windows Updateを行わない)Windows環境では手動で該当CAサイトからRoot証明書をダウンロードしてWindowsにImportする必要がある*
3. クライアントは任意のデータをサーバに送る
4. サーバはクライアントから送られてきた任意のデータを自身の秘密鍵で暗号化してクライアントに返す
5. クライアントはサーバのディジタル証明書に含まれている公開鍵でデータを復号化し、復号化できれば通信相手が偽物ではないことを確認できる
6. クライアントは共通鍵を生成して、ディジタル証明書に含まれている公開鍵で暗号化しサーバに送る
7. サーバはクライアントから送られてきた暗号化されてる共通鍵を自身の秘密鍵で復号化する
   > **Note**  
   > クライアントが生成した共通鍵を**Session Key**とも呼ぶ
8.  以降クライアントとサーバはこの共通鍵(Session Key)を使ってデータを暗号化/復号化してやり取りをする

    ##### TLS handshakeについて
    - 参照URL
        - https://www.cloudflare.com/ja-jp/learning/ssl/what-happens-in-a-tls-handshake/
    - TLS handshakeはTLS暗号化を使った通信セッションを始めるプロセス
    - TLS handshakeの間通信する二者がメッセージをやり取りして互いを認識し、検証し、使用する暗号化アルゴリズムを決定し、Session Keyについて合意する
    - 使用するTLSのバージョン（TLS 1.2、1.3など）を指定
    - 使用する暗号スイート[^1]を決定
    - サーバー公開鍵とSSL証明書認証局のデジタル署名を介して、サーバーのIDを認証
    - handshake完了後に対称暗号化[^2]に使うためのSession Keyを生成  
    [^1]:鍵交換や暗号化方式、ハッシュ関数など暗号化通信に使われる各種アルゴリズムの組合せ
    [^2]:対称暗号化方式 ⇒ 共通鍵暗号方式  
    非対称暗号化方式 ⇒ 公開鍵暗号方式
    - TLS handshakeはTCP handshakeでTCP接続が開かれた後に行われる
    ![TLS handshake](https://github.com/nutslove/Knowledges/blob/main/SSL(TLS)/image/TLS-handshake.jpg)

#### ディジタル証明書の種類
- 証明書は持ち主をどこまで証明するかによって3つに分類される
- DV → OV → EVの順で証明する内容が多くなり、信頼性・証明書価格も高くなる

1. **ドメイン認証(DV)**
    - ドメインの所有者と、SSL証明書の申請者が一致していることを証明

2. **企業実在認証・組織認証(OV)**
    - ドメイン認証に加えて、企業が実在していることを証明

3. **EV認証(EV)**
    - 上記2つの認証に加えて更に厳格な審査を受ける

#### 中間証明書について
- 参考URL
  - https://ssl.sakura.ad.jp/column/difference-in-ssl/
・・要整理・・

## KubernetesでのSSL/TLS
- 参考URL (KR)
  - https://ikcoo.tistory.com/25
- Public Key
  - `*.crt`もしくは`*.pem`
- Private Key
  - `*.key`もしくは`*.key.pem`
