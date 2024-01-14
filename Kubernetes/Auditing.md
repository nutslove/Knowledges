- https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
- API Serverに対するアクションの監査ログ

### **各リクエストは以下の４つのStageで記録される**
- `RequestReceived`
  - リクエストを受信して処理が行われる前の時点
- `ResponseStarted`
  - レスポンスヘッダーが送信された後、レスポンスボディが送信される前の時点
  - 長時間実行されるリクエスト(`watch`など)でのみ発生
- `ResponseComplete`
  - リクエストに対して応答(レスポンスボディの送信)が完了した時点
- `Panic`
  - パニックが起きた時

### API Serverの実行時に指定するflag
- `--audit-policy-file`
  - auditログ取得ルール(e.g. どのNameSpaceのどのResourceのどのアクション(e.g. delete)に対してログを残すか/残さないか)を定義してファイルを指定
  - Auditファイルの例
    ~~~yaml
    apiVersion: audit.k8s.io/v1 # This is required.
    kind: Policy
    # Don't generate audit events for all requests in RequestReceived stage.
    omitStages:
      - "RequestReceived"
    rules:
      # Log pod changes at RequestResponse level
      - level: RequestResponse
        resources:
        - group: ""
          # Resource "pods" doesn't match requests to any subresource of pods,
          # which is consistent with the RBAC policy.
          resources: ["pods"]
      # Log "pods/log", "pods/status" at Metadata level
      - level: Metadata
        resources:
        - group: ""
          resources: ["pods/log", "pods/status"]

      # Don't log requests to a configmap called "controller-leader"
      - level: None
        resources:
        - group: ""
          resources: ["configmaps"]
          resourceNames: ["controller-leader"]

      # Don't log watch requests by the "system:kube-proxy" on endpoints or services
      - level: None
        users: ["system:kube-proxy"]
        verbs: ["watch"]
        resources:
        - group: "" # core API group
          resources: ["endpoints", "services"]

      # Don't log authenticated requests to certain non-resource URL paths.
      - level: None
        userGroups: ["system:authenticated"]
        nonResourceURLs:
        - "/api*" # Wildcard matching.
        - "/version"

      # Log the request body of configmap changes in kube-system.
      - level: Request
        resources:
        - group: "" # core API group
          resources: ["configmaps"]
        # This rule only applies to resources in the "kube-system" namespace.
        # The empty string "" can be used to select non-namespaced resources.
        namespaces: ["kube-system"]

      # Log configmap and secret changes in all other namespaces at the Metadata level.
      - level: Metadata
        resources:
        - group: "" # core API group
          resources: ["secrets", "configmaps"]

      # Log all other resources in core and extensions at the Request level.
      - level: Request
        resources:
        - group: "" # core API group
        - group: "extensions" # Version of group should NOT be included.

      # A catch-all rule to log all other requests at the Metadata level.
      - level: Metadata
        # Long-running requests like watches that fall under this rule will not
        # generate an audit event in RequestReceived.
        omitStages:
          - "RequestReceived"
    ~~~
- `--audit-log-maxage`
  - auditログ最大保存日数
- `--audit-log-path`
  - auditログを書き出すファイルのフルパスを指定

### 監査レベル
- 該当イベントが発生した時にログに記録するか、記録する場合はどの情報まで記録するかを定義
- 以下４つのレベルがある
  - `None`
    - ログに記録しない
  - `Metadata`
    - リクエストのメタデータ(リクエストしたユーザー、タイムスタンプ、Resource、Verbなど)を記録するが、リクエストやレスポンスのボディは記録しない
  - `Request`
    - リクエストのメタデータとボディは記録されるけど、レスポンスボディは記録しない
  - `RequestResponse`
    - メタデータ、リクエストとレスポンスのボディを記録