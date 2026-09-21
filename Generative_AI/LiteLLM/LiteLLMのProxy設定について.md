# LiteLLMのProxy設定について

Claude CodeをローカルLiteLLMプロキシ経由で使う際の接続設定、Headroom(コンテキスト圧縮)連携、認証パターン、ユーザー識別についてまとめる。

テレメトリ(Prometheus/OTel/Tempo/Loki/Grafana)の仕組み自体は [`LiteLLMのObservability設定について.md`](./LiteLLMのObservability設定について.md) を参照。本ファイルでは重複を避け、Claude Code接続・Headroom連携に固有の話題のみ扱う。

---

## 1. Claude CodeをローカルLiteLLMに接続

### 1.1 settings.jsonの基本(APIキー直運用パターン)

`~/.claude/settings.json`(グローバル)または`.claude/settings.json`(プロジェクト単位)。

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://localhost:4000",
    "ANTHROPIC_AUTH_TOKEN": "sk-xxxx",
    "ANTHROPIC_MODEL": "claude-sonnet-5"
  }
}
```

`ANTHROPIC_BASE_URL`は「どこに送るか」、`ANTHROPIC_MODEL`は「どのモデルを使うか」で別軸。両方必要。

> [!CAUTION]
> シェルの環境変数(`export`)はsettings.jsonより**優先される**ため、二重設定に注意。

### 1.2 docker-composeへのconfig.yamlマウント

```yaml
services:
  litellm:
    image: docker.litellm.ai/berriai/litellm:latest
    ports:
      - "4000:4000"
    volumes:
      - ./config.yaml:/app/config.yaml
    command: ["--config", "/app/config.yaml"]
    environment:
      LITELLM_MASTER_KEY: sk-xxxx
      LITELLM_SALT_KEY: sk-xxxx
      DATABASE_URL: postgresql://litellm:litellm@db:5432/litellm
      STORE_MODEL_IN_DB: "True"
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16
    environment:
      POSTGRES_USER: litellm
      POSTGRES_PASSWORD: litellm
      POSTGRES_DB: litellm
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U litellm"]
      interval: 5s
      timeout: 5s
      retries: 10
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

> [!NOTE]
> 本リポジトリの実際のdocker-compose.ymlはイメージをバージョン固定タグにしており、Prometheus/OTel関連の追加サービスも含む。詳細は[`LiteLLMのObservability設定について.md`](./LiteLLMのObservability設定について.md)参照。

> [!CAUTION]
> `volumes`/`command`などサービス定義自体を変更した場合、`docker compose restart`では反映されない。`docker compose up -d`でコンテナ再作成が必要(反映されないケースの詳細も[`LiteLLMのObservability設定について.md`](./LiteLLMのObservability設定について.md)の運用注意点を参照)。

### 1.3 config.yaml側のmodel_list(現行版)

```yaml
# config.yaml
model_list:
  - model_name: claude-sonnet-5
    litellm_params:
      model: anthropic/claude-sonnet-5

  - model_name: claude-opus-5
    litellm_params:
      model: anthropic/claude-opus-5

  - model_name: claude-fable-5-1
    litellm_params:
      model: anthropic/claude-fable-5-1

general_settings:
  master_key: os.environ/LITELLM_MASTER_KEY
  forward_client_headers_to_llm_api: true
```

> [!CAUTION]
> `model_name`(左側)はClaude Code側の`ANTHROPIC_MODEL`と一致させる必要がある。不一致だと`Invalid model name`エラー。
>
> virtual keyに`models`制限をかけている場合、そのリストにも新しいモデル名を追加すること。

---

## 2. 認証方式: APIキー直運用 vs `/login`(サブスクリプション)

| | virtual keyの置き場所 | Anthropicへの課金 |
|---|---|---|
| パターンA(APIキー直運用) | `ANTHROPIC_AUTH_TOKEN` | LiteLLM設定の`ANTHROPIC_API_KEY`(従量課金) |
| パターンB(`/login`サブスク運用) | `ANTHROPIC_CUSTOM_HEADERS`の`x-litellm-api-key` | Claude Codeの`/login`セッション(サブスク) |

### パターンB: `/login`を使う場合のconfig.yaml

```yaml
# config.yaml
model_list:
  - model_name: anthropic-claude
    litellm_params:
      model: anthropic/claude-sonnet-5

general_settings:
  master_key: os.environ/LITELLM_MASTER_KEY
  forward_client_headers_to_llm_api: true   # 必須。OAuthトークンをAnthropicへ転送
```

### パターンB: Claude Code側settings.json

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://localhost:4000",
    "ANTHROPIC_MODEL": "anthropic-claude",
    "ANTHROPIC_CUSTOM_HEADERS": "x-litellm-api-key: Bearer sk-xxxx"
  }
}
```

> [!CAUTION]
> `ANTHROPIC_AUTH_TOKEN`/`ANTHROPIC_API_KEY`を設定すると、`/login`のOAuthセッションより優先されてしまい、サブスク認証が機能しなくなる。`Authorization`ヘッダはOAuth専用に空けておき、LiteLLM向けキーは`ANTHROPIC_CUSTOM_HEADERS`の`x-litellm-api-key`に逃がすこと。

---

## 3. メトリクスにユーザー(ID/メール)を出す方法

### 3.1 `end_user`ラベル(推奨・個別キー発行不要)

リクエストヘッダーで都度指定する方式。事前のユーザー登録不要、自動的にDBへupsertされる。

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://localhost:4000",
    "ANTHROPIC_MODEL": "claude-sonnet-5",
    "ANTHROPIC_CUSTOM_HEADERS": "x-litellm-api-key: Bearer sk-xxxx\nx-litellm-end-user-id: alice@example.com"
  }
}
```

値は自由文字列なので、メールアドレスをそのまま入れることで人間が読める識別子になる。

### 3.2 Prometheusで`end_user`を有効化する追加設定(opt-in)

```yaml
# config.yaml
litellm_settings:
  callbacks: ["prometheus"]
  enable_end_user_cost_tracking_prometheus_only: true
```

この設定自体の詳細は[`LiteLLMのObservability設定について.md`](./LiteLLMのObservability設定について.md)の「LiteLLM側の設定」セクションにも記載。

> [!CAUTION]
> Prometheusは`end_user`をデフォルトでは追跡しない仕様(カーディナリティ対策のため)。スペンドログ(DB/Admin UI)は影響を受けないが、Prometheusの`end_user`ラベルだけは上記設定を明示しないと空になる。

### 3.3 セキュリティ上の注意

> [!CAUTION]
> メールアドレスをラベルにすると、`/metrics`にアクセスできる人全員に見えてしまう。認証なし公開(`require_auth_for_metrics_endpoint: false`)にしている場合は特に注意。
>
> ユーザー数が多いとラベルの組み合わせ(カーディナリティ)が増え、Prometheus/Grafanaの負荷に影響する(カーディナリティの注意点全般は[`LiteLLMのObservability設定について.md`](./LiteLLMのObservability設定について.md)参照)。

---

## 4. Headroom(コンテキスト圧縮)連携

### 4.1 概要とアーキテクチャ

Headroomはツール出力・RAGペイロードなどをLLMに渡す前に圧縮するサービス。LiteLLMの **guardrail(pre_call)** として動作する。

```
Client → LiteLLM Gateway(そのまま)
              │
              ├─ pre_call時にHeadroomを内部呼び出しして圧縮
              └─ 圧縮後のペイロードを上流LLMへ転送
```

クライアント・上流LLMともにHeadroomと直接通信しない(サイドカー方式)。

### 4.2 config.yaml設定(運用主体が別チームでもOK)

```yaml
# config.yaml
guardrails:
  - guardrail_name: headroom-compression
    litellm_params:
      guardrail: headroom
      mode: pre_call
      api_base: https://your-headroom-service   # ネットワーク越しのURLでOK
      api_key: os.environ/HEADROOM_API_KEY       # [OPTIONAL] 認証
      # default_on: true                          # [OPTIONAL] 全リクエスト圧縮するか
```

通信は`{api_base}/v1/compress`へのシンプルなHTTP POST。運用チームが別でも`api_base`+`api_key`のインターフェースで疎結合に接続可能。

> [!NOTE]
> デフォルトでは`user`/`system`ロールのメッセージや`cache_control`付きメッセージは圧縮対象外(`HEADROOM_COMPRESS_USER_MESSAGES=1`で変更可)。

### 4.3 メトリクスとの整合性(圧縮後の値で計上されるか)

**結論: LiteLLM側のメトリクス(トークン数・コスト)は圧縮後(実際に送られた・実際に課金された)の値で計上される。実態との乖離はない。**

理由: `pre_call`フックは「LLM API呼び出しに実際に渡されるデータ」を書き換える。LiteLLMのコスト計算はLLMプロバイダが返す実際の`usage`を元にしているため。

> [!NOTE]
> 「圧縮でどれだけ節約できたか」というBefore/After比較はLiteLLM側には残らない。Headroom自身の`/metrics`(`headroom_tokens_saved_total`など)やJSON statsエンドポイントを別途参照する必要がある。
>
> 特定リクエストで圧縮が走ったかは、レスポンスヘッダ`x-litellm-applied-guardrails`やAdmin UIの`guardrail_information`で確認可能。
