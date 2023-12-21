- Nginx Open Source公式Doc
  - https://nginx.org/en/docs/  
- `nginx -t`コマンドで設定ファイルの文法チェックができる

## Configuration
##### `proxy_pass`にIPアドレスではなくFQDNを指定する時の注意点
- 参考URL
  - https://www.subthread.co.jp/blog/20190424/
  - https://www.ponkotsu-log.com/entry/2017/07/02/002353
  - http://nginx.org/en/docs/http/ngx_http_core_module.html#resolver
- 基本、nginxの起動時の解決IPアドレスをずっと使い続けるらしい。なのでAutoScaling等でIPアドレスが変わった時に対応できない。  
それを解決するために`resolver`でDNSサーバを指定し、ホスト名(FQDN)を変数にして`proxy_pass`には変数を指定すれば定期的に名前解決を行うらしい。
    ~~~
    location ~ /hoge/(.*) {
        resolver 172.31.0.2 valid=5s;
        set $endpoint hoge.example.com;
        proxy_pass https://$endpoint;
    }
    ~~~

##### `server_name`にサーバ(ドメイン)名の代わりにforward先のIPアドレスを書いても良い
- 参考URL
  - https://obel.hatenablog.jp/entry/20180813/1534133423
- 例
  - `http://192.168.123.123:12345`に接続すると`http://192.168.123.124:54321`にプロキシされる
      ~~~
      server {
      listen 12345;
    server_name 192.168.123.123;

      location / {
        proxy_pass http://192.168.123.124:54321;
    ~~~

##### `location`ブロックについて
- 参考URL
  - https://www.digitalocean.com/community/tutorials/understanding-nginx-server-and-location-block-selection-algorithms
  - https://www.thegeekstuff.com/2017/05/nginx-location-examples/
  - [日本語でまとまってるページ](https://heartbeats.jp/hbblog/2012/04/nginx05.html)
  - [Nginx公式Doc](https://nginx.org/en/docs/http/ngx_http_core_module.html#location)
- `location`のevaluationロジック
  > Nginx evaluates the possible location contexts by comparing the request URI to each of the locations. It does this using the following algorithm:
  >
  > - Nginx begins by checking all prefix-based location matches (all location types not involving a regular expression). It checks each location against the complete request URI.
  > - First, Nginx looks for an exact match. If a location block using the = modifier is found to match the request URI exactly, this location block is immediately selected to serve the request.
  > - If no exact (with the`=`modifier) location block matches are found, Nginx then moves on to evaluating non-exact prefixes. It discovers the longest matching prefix location for the given request URI, which it then evaluates as follows:
  >     - If the longest matching prefix location has the`^~`modifier, then Nginx will immediately end its search and select this location to serve the request.
  >     - If the longest matching prefix location does not use the`^~`modifier, the match is stored by Nginx for the moment so that the focus of the search can be shifted.
  > - After the longest matching prefix location is determined and stored, Nginx moves on to evaluating the regular expression locations (both case sensitive and insensitive). If there are any regular expression locations within the longest matching prefix location, Nginx will move those to the top of its list of regex locations to check. Nginx then tries to match against the regular expression locations sequentially. The first regular expression location that matches the request URI is immediately selected to serve the request.
  > - If no regular expression locations are found that match the request URI, the previously stored prefix location is selected to serve the request.
  >>
  > It is important to understand that, by default, Nginx will serve regular expression matches in preference to prefix matches. However, it evaluates prefix locations first, allowing for the administer to override this tendency by specifying locations using the`=`and`^~`modifiers.
  >
  > It is also important to note that, while prefix locations generally select based on the longest, most specific match, regular expression evaluation is stopped when the first matching location is found. This means that positioning within the configuration has vast implications for regular expression locations.

- `^~`や`=`条件に一致したらそれ以降のlocationで定義した正規表現が評価されない。  
  例えば下記の場合、1個目の条件ですべて一致するので、2個目の条件は評価されず、すべてallowされちゃう。  
  2つ目以降の`location`が評価されるようにしたい場合は`location / { allow ・・・ }`で記載。  
  2つ目以降の`location`は基本`~*`を使えば問題なさそう。
  ~~~
  location ^~ / { 
    allow all;
    proxy_pass  http://10.10.10.10:8000;
  }

  location ~* /ja-JP/app/.+/search {
    deny all;
  }
  ~~~

- `~`と`~*`は条件文の中の大文字/小文字を区別するかどうかの違い。  
  例えば下記のような設定でdenyしても`/ja-jp/~`の小文字でアクセス出来てしまう。  
  `~*`にすると`/ja-jp/~`でもdenyされる。
  ~~~
  location ~ /ja-JP/app/.+/search {
    deny all;
  }
  ~~~

- __locationでURLパス単位のdeny等の設定時の注意点__
  - ブラウザでアクセス時に見えるURLパスと実際内部で処理されるURLが異なる場合がある。そういう場合はブラウザから見えるURLパスで`location`を設定しても想定通りに制御されない。
  - ブラウザの開発者ツールの`Network`タブにて内部で処理されるURLパスやjavascript等を確認してそれを基に`location`を設定すれば想定通りに動く場合があるので要確認

##### `server`ブロック
- 以下のように1つのconfファイル内に複数の`server`ブロックの指定ができる
  ~~~
  server {
    listen    800;
   
    location / {
      allow all;
      proxy_set_header Host $http_host;
      proxy_pass  http://OPS-GRAFANA-NLB-asasdasd.elb.ap-northeast-1.amazonaws.com;
    }
  }
   
  server {
    listen    8082;
   
    location / {
      allow all;
      proxy_set_header Host $http_host;
      proxy_pass  http://CHAT-GRAFANA-NLB-adfqwegre.elb.ap-northeast-1.amazonaws.com;
    }
  }
   
  server {
    listen    8083;
   
    location / {
      client_max_body_size 50m;
      allow all;
      proxy_set_header Host $http_host;
      proxy_pass  http://HELP-NLB-asdwqtrgh.elb.ap-northeast-1.amazonaws.com;
    }
  }
   
  server {
    listen    8084;
   
    location / {
      allow all;
      proxy_set_header Host $http_host;
      proxy_pass  http://OMG-NLB-asdqwght.elb.ap-northeast-1.amazonaws.com;
    }
    location ~* /d/. {
      deny all;
    }
    location ~* /explore {
      deny all;
    }
    location ~* /dashboards/uid {
      deny all;
    }
    location ~* /public/build/DashboardListPage.*\.js$ {
      deny all;
    }
    location ~* /public/build/DashboardPage.*\.js$ {
      deny all;
    }
  }
  ~~~