- CoreDNSに **`externalName`に指定したドメイン名と`Endpoints`に指定したIPアドレスでAレコードが登録されて、さらに`<service名>.<namespace名>.svc.cluster.local`と`externalName`に指定したドメイン名でCNAMEレコードが登録される**
- `<service名>.<namespace名>.svc.cluster.local`でアクセスするとCNAMEレコード → Aレコード → IPアドレスの順で最終的にIPアドレスを取得できる
- `ExternalName`は`EXTERNAL-IP`のところにIPアドレスの代わりにドメインが表示される
- 例  
  ```yaml
  ---
  apiVersion: v1
  kind: Service
  metadata:
    labels:
      app: nginx
    name: nginx
  spec:
    type: ExternalName
    externalName: www.iij.ad.jp
  ---
  apiVersion: v1
  kind: Endpoints
  metadata:
    labels:
      app: nginx
    name: nginx
  subsets:
  - addresses:
    - ip: 202.232.2.180
  ```

  ```shell
  / # dig nginx.default.svc.cluster.local.

  ; <<>> DiG 9.11.6-P1 <<>> nginx.default.svc.cluster.local.
  ;; global options: +cmd
  ;; Got answer:
  ;; WARNING: .local is reserved for Multicast DNS
  ;; You are currently testing what happens when an mDNS query is leaked to DNS
  ;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 45955
  ;; flags: qr aa rd; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1
  ;; WARNING: recursion requested but not available

  ;; OPT PSEUDOSECTION:
  ; EDNS: version: 0, flags:; udp: 1232
  ; COOKIE: 9e8f60d765ce465d (echoed)
  ;; QUESTION SECTION:
  ;nginx.default.svc.cluster.local. IN    A

  ;; ANSWER SECTION:
  nginx.default.svc.cluster.local. 5 IN   CNAME   www.iij.ad.jp.
  www.iij.ad.jp.          5       IN      A       202.232.2.180
  ```