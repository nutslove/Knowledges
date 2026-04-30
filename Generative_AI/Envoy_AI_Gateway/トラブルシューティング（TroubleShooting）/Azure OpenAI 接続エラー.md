# Envoy AI Gateway × Azure OpenAI 接続エラー トラブルシュートメモ

## 概要

| 項目 | 内容 |
|------|------|
| 日時 | 2026-04-30 |
| 環境 | EKS (ap-northeast-1a) + Envoy AI Gateway v0.5.0 |
| 対象プロバイダー | Azure OpenAI Service (`<your-resource-name>.openai.azure.com`, Japan East) |
| 認証方式 | Microsoft Entra ID (Service Principal / Client Secret) |
| 問題 | Envoy AI Gateway 経由で Azure OpenAI を呼ぶと常に 502 エラー |
| 結論 | Envoy が upstream に HTTP/2 で接続し、Azure 側が RST_STREAM で切断していた |
| 対処 | EnvoyPatchPolicy で対象 cluster を HTTP/1.1 強制 |

備考: AWS Bedrock / GCP Vertex AI は同じ Envoy AI Gateway 経由で正常動作。Azure のみ問題。

---

## 発生した事象

### エラー内容

```bash
$ curl -H "Content-Type: application/json" \
    -d '{ "model": "gpt-4o", "messages": [{"role": "user", "content": "hi"}]}' \
    $GATEWAY_URL/v1/chat/completions

{"type":"error","error":{"type":"OpenAIBackendError","code":"502","message":"upstream connect error or disconnect/reset before headers. reset reason: protocol error"}}
```

### Envoy アクセスログ

```json
{
  "response_code": 502,
  "response_code_details": "upstream_reset_before_response_started{protocol_error}",
  "response_flags": "UPE",
  "upstream_cluster": "httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/7",
  "upstream_host": "20.191.161.227:443",
  "x-envoy-origin-path": "/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview"
}
```

### ポイント
- `upstream_reset_before_response_started{protocol_error}` → upstream(Azure) へリクエスト送信中にプロトコルエラーで切断
- `UPE` フラグ = Upstream Protocol Error
- TCP 接続自体は成功している (upstream_host が確定している)

---

## 切り分け手順

### Step 1: ExtProc が正しく動作しているか確認

ExtProc を debug ログに変更して確認:

```bash
kubectl logs -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -c ai-gateway-extproc --tail=200
```

**結果**: ExtProc は正常動作していた:
- 正しいパス組み立て: `/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview`
- 正しい `Authorization: Bearer <JWT>` ヘッダー設定
- JWT トークンの中身も妥当 (aud=cognitiveservices.azure.com, appid, tid 全て正しい)

→ ExtProc は問題なし。Envoy 本体から upstream への接続段階で失敗している。

### Step 2: Azure OpenAI 自体の正常性確認

EC2 から直接 Azure OpenAI を叩いて、Azure 側・API キー・デプロイメント名・API バージョンが正しいか確認。

```bash
# APIキー方式
curl -v "https://<your-resource-name>.openai.azure.com/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview" \
  -H "Content-Type: application/json" \
  -H "api-key: $AZURE_API_KEY" \
  -d '{"messages":[{"role":"user","content":"hi"}],"max_tokens":50}'
```

**結果**: 200 OK で正常応答。Azure 側・URL・デプロイメント名・API バージョン全て正しい。

> [!NOTE]
> APIキーはAzure OpenAIの「Resource Management」の「Keys and Endpoint」から確認可能

### Step 3: Entra ID 認証の単独確認

Envoy AI Gateway と同じ Service Principal でアクセストークン取得 → そのトークンで叩いてみる:

```bash
# トークン取得
ACCESS_TOKEN=$(curl -s -X POST \
  "https://login.microsoftonline.com/$TENANT_ID/oauth2/v2.0/token" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "scope=https://cognitiveservices.azure.com/.default" \
  | jq -r .access_token)

# Bearerトークンで呼び出し
curl -v "https://<your-resource-name>.openai.azure.com/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"messages":[{"role":"user","content":"hi"}],"max_tokens":50}'
```

**結果**: 200 OK で正常応答。
- Service Principal の権限 OK
- Client Secret の値も正しい
- Azure OpenAI への IAM ロール (Cognitive Services OpenAI User) も付与済み

→ 認証情報自体に問題なし。

> [!NOTE]
> CLIENT_SECRETはAzureのEntraIDの「App registrations」から該当アプリを選択 → 「Certificates & secrets」タブ → 「Client secrets」の「Value」を確認

### Step 4: K8s Secret の値確認

ExternalSecret 経由で AWS Secrets Manager から取得した値を確認:

```bash
kubectl get secret -n envoy-ai-gateway-system envoy-ai-gateway-azure-client-secret \
  -o jsonpath='{.data.client-secret}' | base64 -d
```

**結果**: 手元の値と完全一致。Secret 内のキー名 (`client-secret`) も ExtProc が期待する形式と一致。

### Step 5: Pod 再起動で Secret 再読み込み

```bash
kubectl delete pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway
```

**結果**: 変化なし。引き続き 502。

### Step 6: Envoy Admin API で cluster stats を確認

```bash
ENVOY_POD=$(kubectl get pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -o jsonpath='{.items[0].metadata.name}')

kubectl port-forward -n envoy-gateway-system $ENVOY_POD 19000:19000 &

# 該当 cluster の stats
curl -s http://localhost:19000/stats | grep "rule/7" | grep -v ": 0$"
```

**初回結果**: rule/7 cluster の stats が全て 0 → この Pod にはリクエストが到達していなかった。

### Step 7: 複数 Pod 問題に気づき、もう一方の Pod で確認

`envoyDeployment.replicas: 2` だったため、2 つ目の Pod で再確認:

```bash
POD2=$(kubectl get pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -o jsonpath='{.items[1].metadata.name}')

kubectl port-forward -n envoy-gateway-system $POD2 19000:19000 &
curl -s http://localhost:19000/stats | grep "rule/7" | grep -v ": 0$"
```

**結果**: 重要な情報が判明:

```
upstream_cx_http2_total: 2          ← Envoy は HTTP/2 で接続
upstream_cx_http1_total: 0          ← HTTP/1.1 は使われていない
http2.rx_reset: 2                   ← HTTP/2 stream を Azure が RST で切断
upstream_cx_destroy_remote: 2       ← リモート(Azure)側からの切断
upstream_rq_tx_reset: 2             ← 送信側 reset
ssl.handshake: 2                    ← TLSハンドシェイク自体は成功
ssl.versions.TLSv1.2: 2             ← TLS 1.2 で接続
```

→ **TLS ハンドシェイクは成立しているが、HTTP/2 の stream が直後に RST で切断されている**

### Step 8: HTTP version の影響を切り分け

EC2 から HTTP/1.1 と HTTP/2 で同じエンドポイントを叩いて比較:

```bash
# HTTP/1.1
curl --http1.1 -v "https://rca-llm.openai.azure.com/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview" \
  -H "api-key: $AZURE_API_KEY" -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"hi"}]}'
# → 200 OK

# HTTP/2  
curl --http2 -v "https://rca-llm.openai.azure.com/openai/deployments/gpt-4o/chat/completions?api-version=2025-01-01-preview" \
  -H "api-key: $AZURE_API_KEY" -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"hi"}]}'
# → 200 OK (curlの場合)
```

**結果**: curl からは HTTP/1.1 でも HTTP/2 でも動く。
→ Envoy 特有の HTTP/2 フレーム送信方法で Azure 側に拒否されている可能性。

---

## 原因

**Envoy → Azure OpenAI 間の TLS ALPN ネゴシエーションで HTTP/2 が選択され、Azure 側が Envoy の HTTP/2 リクエストを RST_STREAM で切断していた。**

詳細:
- TLS ALPN で Azure サーバーが `h2` を提案 → Envoy が受け入れて HTTP/2 で接続
- Envoy が POST リクエスト送信
- Azure 側がリクエストを受け取り、ヘッダーを処理する段階で RST_STREAM フレームを返す
- Envoy が `protocol error` として応答

curl の HTTP/2 では動作するが、Envoy の HTTP/2 では動作しない。フレーム送信方式・ヘッダー圧縮 (HPACK)・タイミングなどの差異によるものと推測。

### 補足: なぜ AWS Bedrock と GCP Vertex AI では動いたか

- AWS Bedrock: HTTP/1.1 / HTTP/2 どちらでも問題なく応答
- GCP Vertex AI: HTTP/1.1 / HTTP/2 どちらでも問題なく応答
- Azure OpenAI: Envoy の HTTP/2 リクエストのみ拒否 (curl からの HTTP/2 は応答)

---

## 実施した対応

### EnvoyPatchPolicy で該当 cluster の HTTP/1.1 強制

事前確認: `EnvoyPatchPolicy` が有効化されているか:

```bash
kubectl get configmap -n envoy-gateway-system envoy-gateway-config -o yaml | grep -i patch
# enableEnvoyPatchPolicy: true ← 有効化済み
```

以下の EnvoyPatchPolicy を適用:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: envoy-ai-gateway-azure-http1
  namespace: envoy-gateway-system
spec:
  type: JSONPatch
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: envoy-ai-gateway
  jsonPatches:
    - type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
      name: httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/4
      operation:
        op: add
        path: "/typed_extension_protocol_options"
        value:
          envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
            "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
            explicit_http_config:
              http_protocol_options: {}
    - type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
      name: httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/5
      operation:
        op: add
        path: "/typed_extension_protocol_options"
        value:
          envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
            "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
            explicit_http_config:
              http_protocol_options: {}
    - type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
      name: httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/6
      operation:
        op: add
        path: "/typed_extension_protocol_options"
        value:
          envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
            "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
            explicit_http_config:
              http_protocol_options: {}
    - type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
      name: httproute/envoy-ai-gateway-system/envoy-ai-gateway/rule/7
      operation:
        op: add
        path: "/typed_extension_protocol_options"
        value:
          envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
            "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
            explicit_http_config:
              http_protocol_options: {}
```

> [!NOTE]
> - `rule/4 ~ rule/7` は AIGatewayRoute で `envoy-ai-gateway-azure` Backend を指す 4 つのルール (gpt-5.4 / gpt-5.4-mini / gpt-5.4-nano / gpt-4o)
> - **`typed_extension_protocol_options` で `explicit_http_config.http_protocol_options: {}` を指定すると、HTTP/1.1 強制になる**

### 動作確認

```bash
$ curl -H "Content-Type: application/json" \
    -d '{ "model": "gpt-4o", "messages": [{"role": "user", "content": "hi"}]}' \
    $GATEWAY_URL/v1/chat/completions

{"choices":[{"...","message":{"content":"Hello! 😊 How can I assist you today?",...}}],...}
# → 200 OK で正常動作
```

stats でも HTTP/1.1 接続を確認:

```
upstream_cx_http1_total: 1          ← 新規接続はHTTP/1.1
upstream_cx_http2_total: 2          ← パッチ適用前のカウンター(増えていない)
upstream_rq_total: 3                ← 全3リクエスト処理
```

---

## 残課題と運用上の注意点

### 課題 1: rule 番号が AIGatewayRoute の順番に依存

`rule/N` の番号は `AIGatewayRoute` のルール定義順で決まる。

**影響**:
- AIGatewayRoute の先頭付近にルール追加 → Azure より前の rule 番号がずれる
- Azure の新モデル追加 → 新しい rule 番号にもパッチ追加が必要

**確認コマンド**:

```bash
# 現在の cluster 一覧と Azure backend を参照しているものを確認
kubectl port-forward -n envoy-gateway-system $POD2 19000:19000 &
curl -s http://localhost:19000/clusters | grep "rca-llm" | head -10

# EnvoyPatchPolicyの適用ステータス
kubectl describe envoypatchpolicy -n envoy-gateway-system envoy-ai-gateway-azure-http1
# Conditions.Programmed: True であればOK
```

### 課題 2: 監視

- `upstream_cx_http2_total` が増えていないか定期チェック (誤って HTTP/2 で繋がる場合がないか)
- `upstream_rq_502` の急増アラート

### 改善案 (検討中)

1. **AIGatewayRoute を Azure 専用に分離**: `httproute/envoy-ai-gateway-azure-route/rule/N` のような独立した名前空間にすれば、他プロバイダーのルール変更の影響を受けない
2. **Envoy AI Gateway へ Issue 起票**: `BackendSecurityPolicy.AzureCredentials` か `Backend` リソースで `forceHttp1` のような明示的なオプションを追加してほしい

---

## 参考情報

### 動いた組み合わせ

| プロバイダー | 認証方式 | プロトコル | 備考 |
|-------------|---------|-----------|------|
| AWS Bedrock | Pod Identity / IRSA | HTTP/2 (デフォルト) | 問題なし |
| GCP Vertex AI | Service Account Key (Secrets Manager 経由) | HTTP/2 (デフォルト) | `region: global` で全 Gemini モデル動作 |
| Azure OpenAI | Entra ID Service Principal (Secrets Manager 経由) | **HTTP/1.1 強制** | EnvoyPatchPolicy 必要 |

### 切り分けに有効だったコマンド集

```bash
# ExtProc debug ログ
kubectl logs -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -c ai-gateway-extproc --tail=200

# Envoy Admin API (ポートフォワード)
ENVOY_POD=$(kubectl get pod -n envoy-gateway-system \
  -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway \
  -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n envoy-gateway-system $ENVOY_POD 19000:19000 &

# 該当 cluster の stats(0以外を抽出)
curl -s http://localhost:19000/stats | grep "rule/N" | grep -v ": 0$"

# cluster一覧
curl -s http://localhost:19000/clusters | grep observability_name

# 注意: replicas が複数の場合は全Podでstatsを確認すること
```

### Envoy Admin API での重要な stat

| stat | 意味 |
|------|------|
| `upstream_cx_http1_total` | upstream への HTTP/1.1 接続数 |
| `upstream_cx_http2_total` | upstream への HTTP/2 接続数 |
| `upstream_cx_destroy_remote` | リモート側からの接続切断数 |
| `upstream_cx_protocol_error` | プロトコルエラー数 |
| `upstream_rq_tx_reset` | 送信側 reset 数 |
| `http2.rx_reset` | HTTP/2 stream の RST 受信数 |
| `ssl.handshake` | TLS ハンドシェイク成功数 |

---

## 学んだこと (個人メモ)

1. **EKS で `replicas` が複数の場合、Envoy stats は Pod ごと**に持っている。1 つの Pod で stats が 0 でも、もう一方で発生している可能性があるので両方見る必要がある
2. **`response_flags: UPE` (Upstream Protocol Error)** はネットワーク到達はしている前提で、TLS 後の HTTP プロトコル層で問題が起きているサイン
3. **TLS ALPN ネゴシエーション**で HTTP/2 が選ばれることがあり、curl では動いても Envoy では動かないケースがある (フレーム送信の細かな差異)
4. **EnvoyPatchPolicy** は強力だが cluster 名 (rule/N) が動的に決まるため、運用上 fragile になりやすい。長期的にはコミュニティへ機能要望を上げるのが正解
5. **切り分け順序**: Envoy AI Gateway を介さない直接 curl → 認証単独テスト → Envoy 経由 → Pod ごとの stats 確認、と段階的に絞り込むと早い