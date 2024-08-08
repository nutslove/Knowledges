## TLS設定
### OpenSearch
- 参照URL
  - https://opensearch.org/docs/latest/security/configuration/tls/
  - https://opensearch.org/docs/latest/troubleshoot/tls/#bad-configuration
  - https://opensearch.org/docs/latest/security/configuration/generate-certificates/
- OpenSearchでは内部クラスタリングのための`transport`はTLSが必須になっていて、外部からのアクセスのための`http`はデフォルトではTLSが無効になっている
- `http`でTLSを有効にしてクライアントからのアクセスをHTTPS化して、OpenSearch（**multi-node clusterの場合は最初にクライアントからのアクセスを受け付けるClientノードでのみ**）でTLSを終端させることができる
- 設定例  
  ```yaml
  opensearch.yml: |
    plugins:
      security:
        ssl:
          transport:
            pemcert_filepath: esnode.pem
            pemkey_filepath: esnode-key.pem
            pemtrustedcas_filepath: root-ca.pem
            enforce_hostname_verification: false
          http:
            enabled: true
            pemcert_filepath: node-for-api.pem
            pemkey_filepath: node-key-for-api.pem
            pemtrustedcas_filepath: root-ca-for-api.pem
        allow_unsafe_democertificates: true
        allow_default_init_securityindex: true
        authcz:
          admin_dn:
            - CN=kirk,OU=client,O=client,L=test,C=de
        nodes_dn:
            - 'CN=*.es.lee-test.com,OU=Career Cloud Develop Dept\.,O=KDDI,L=Chiyoda-Ku,ST=Tokyo,C=JP'
  ```

### OpenSearch Dashboard
- 参照URL
  - https://opensearch.org/docs/latest/install-and-configure/install-dashboards/tls/
- 前段のLBなどでTLSを終端させてもいいし、OpenSearch Dashboardで直接TLSを終端させることもできる
  - `server.ssl`配下のパラメータで証明書を指定
- **OpenSearch（Client）側がHTTPSになっていて、オレオレ証明書の場合、エラーになるため、`opensearch.ssl.verificationMode`を`none`にする必要がある**
- 設定例  
  ```yaml
  source:
    helm:
      values: |
        opensearchHosts: "https://lee-test-client.opensearch.svc:9200"
        config:
          opensearch_dashboards.yml: |
            server:
              ssl:
                enabled: true
                certificate: /usr/share/opensearch-dashboards/config/node-for-gui.pem
                key: /usr/share/opensearch-dashboards/config/node-key-for-gui.pem
                certificateAuthorities: /usr/share/opensearch-dashboards/config/root-ca-for-gui.pem
            opensearch:
              ssl:
                verificationMode: none
  ```