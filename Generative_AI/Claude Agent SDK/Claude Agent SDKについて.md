## Claude Agent SDKとは

- https://code.claude.com/docs/ko/agent-sdk/overview

**Claude Agent SDK**は、Claude Codeを支えているエージェントループ・ツール実行・コンテキスト管理などの仕組みを、そのままライブラリとして自分のアプリケーションに組み込めるようにしたSDK。Python / TypeScriptに対応している。

> [!NOTE]
> 元々は `claude-code-sdk` という名前だったが、2025年9月29日に `claude-agent-sdk` へリブランドされた。コーディング用途に留まらず、法務アシスタントやファイナンスアドバイザー、SREボット、セキュリティレビュアーなど幅広い用途で使われるようになったことが理由とされている。

- Python: `claude-agent-sdk`（Python 3.10+）
- TypeScript/JavaScript: `@anthropic-ai/claude-agent-sdk`（Node.js 18+）

ツール実行ループを自前で実装する必要がなく、Read/Edit/Bash/WebSearchなどのビルトインツール、パーミッション制御、サブエージェント、セッション管理、MCP連携などをそのまま利用できるのが最大の特徴。

## 他のClaude系ツールとの比較

| やりたいこと | 使うもの | 理由 |
|---|---|---|
| ツールループを自前実装せずにエージェントを作る | **Claude Agent SDK** | Python/TypeScriptのライブラリとして、自分のプロセス内でエージェントループを実行 |
| ターミナルで対話的に使う・単発タスクを実行する | **Claude Code**（CLI） | 日常利用向けのターミナルインターフェース |
| APIを直接叩いてツールループも自分で実装する | **Client SDKs**（Claude API） | Claude APIへの直接アクセス。Python/TypeScript/C#/Go/Java/PHP/Ruby対応の汎用Messages APIクライアント。「Claude」を冠さない正式名称 |
| 自前でサンドボックス/セッション基盤を持たずに長時間・非同期のエージェントを動かす | **Claude Managed Agents**（パブリックベータ、2026年4月〜） | Anthropicがエージェントとサンドボックスをホスティングするサーバーレス製品（Agent SDKとは別製品。REST API課金はトークン代＋$0.08/セッション時間） |

SDK自体はPython/TypeScript限定。他言語から同じループを使いたい場合は、CLIを `-p --output-format json` オプション付きでサブプロセスとして起動する（headlessモード）。

### Claude Agent SDKとClaude Managed Agentsは「似たもの」か

「自律的にツールを使って作業するClaudeエージェント」という**中身（できること）はほぼ同じ**：どちらもファイル読み書き・コマンド実行・Web検索/取得・MCP連携といった能力を持つエージェントループを提供する。実際、公式ドキュメントでも「Claude Code・Claude Agent SDK・Claude Managed Agentsはいずれもエージェントループ／ツール実行／ランタイムを提供する、より高レベルな仕組み」として同じグループにまとめられている。

ただし実装・運用面では別物：

| | Claude Agent SDK | Claude Managed Agents |
|---|---|---|
| 実行主体 | `claude` CLIサブプロセスを**自分のインフラ**で動かす | Anthropicのサンドボックスで**Anthropicが**動かす（アプリはEventを送るだけ） |
| インターフェース | `query()`のメッセージストリーム（Pythonは`ClaudeSDKClient`による永続セッションも可） | Agent/Environment/Session/EventのREST API＋SSE |
| 機能の充実度 | Subagents・Hooks・Skills・Compaction制御など詳細に制御可能 | ツール群はやや簡略化。代わりにScheduled deployments（cron実行）やMCPトンネル等の運用向け機能あり |
| 提供状況 | GA（一般提供） | パブリックベータ |

まとめると、**「何を実現するか」はほぼ同じ自律エージェントだが、「どこで・誰が動かすか」と機能の細かさが違う姉妹製品**、という理解が近い。

## インストール & クイックスタート

### TypeScript（新規プロジェクト）

```bash
npm init -y
npm pkg set type=module
npm install @anthropic-ai/claude-agent-sdk
npm install --save-dev tsx
```

### Python（uv推奨）

```bash
uv init
uv add claude-agent-sdk
```

> [!NOTE]
> Python/TypeScript両SDKとも、ネイティブのClaude Codeバイナリを同梱しているため、基本的に別途Claude Code本体をインストールする必要はない。ただし`pip`がソースディストリビューションを入れた場合（ARM64 Windows等）や、`npm ci --omit=optional`のようにoptional dependenciesを省略した場合はバイナリが同梱されないため、Claude Codeのネイティブインストールが別途必要になる。

### 認証

```bash
export ANTHROPIC_API_KEY=your-api-key
```

- `.env`は自動読み込みされないので、必要なら`dotenv`等で明示的にロードする
- Amazon Bedrock / Claude Platform on AWS / Google Cloud Vertex / Microsoft Foundry経由の認証にも対応（各種環境変数を設定）
- サードパーティ製品でのclaude.aiログイン代用は原則不可。APIキー認証を使うこと

### 最小サンプル（バグ修正エージェント）

```python
import asyncio
from claude_agent_sdk import query, ClaudeAgentOptions, AssistantMessage, ResultMessage

async def main():
    async for message in query(
        prompt="Review utils.py for bugs that would cause crashes. Fix any issues you find.",
        options=ClaudeAgentOptions(
            allowed_tools=["Read", "Edit", "Glob"],
            permission_mode="acceptEdits",
        ),
    ):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if hasattr(block, "text"):
                    print(block.text)
        elif isinstance(message, ResultMessage):
            print(f"Done: {message.subtype}")

asyncio.run(main())
```

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

for await (const message of query({
  prompt: "Review utils.py for bugs that would cause crashes. Fix any issues you find.",
  options: {
    allowedTools: ["Read", "Edit", "Glob"],
    permissionMode: "acceptEdits"
  }
})) {
  if (message.type === "assistant" && message.message?.content) {
    // ...
  } else if (message.type === "result") {
    console.log(`Done: ${message.subtype}`);
  }
}
```

`query()` が async iterator を返し、Claudeの思考・ツール呼び出し・ツール結果・最終結果がメッセージとしてストリーミングされてくる。

## Single Message Input vs Streaming Input Mode（`query()` と `ClaudeSDKClient`）

SDK全体としては「**Single Message Input**（単発）」と「**Streaming Input Mode**（永続セッション、公式推奨）」という2つの入力モードがあり、Pythonではこれが `query()` / `ClaudeSDKClient` という別々のAPIとして提供されている。

| 項目 | `query()` | `ClaudeSDKClient` |
|---|---|---|
| セッション | 呼び出すたびに新規（デフォルト） | 同一セッションを保持 |
| 会話継続 | `continue_conversation`/`resume`で手動指定 | 自動的に文脈が継続 |
| ストリーミング入力 | 対応（AsyncIterableを渡せる） | 対応 |
| 割り込み（interrupt） | ❌ 非対応 | ✅ `interrupt()`で対応 |
| 画像の途中添付 | ❌ 非対応 | ✅ 対応 |
| 動的なメッセージキューイング | ❌ 非対応 | ✅ 対応 |
| セッション中の設定変更 | ❌ | ✅ `set_permission_mode()` / `set_model()` / `rewind_files()` / `toggle_mcp_server()` 等 |
| 接続管理 | 自動 | `connect()`/`disconnect()`（`async with`で自動化可） |
| 向いている用途 | 単発タスク・バッチ処理・ステートレス環境（Lambda等） | 対話型アプリ・チャットボット・途中で制御が必要な長時間セッション |

```python
# query(): 呼び出すたびに独立（文脈なし）
async for message in query(prompt="Pythonとは？"):
    ...
async for message in query(prompt="どうやってインストールする？"):
    # 前の質問を知らない
    ...
```

```python
# ClaudeSDKClient: 同一セッション内で文脈を保持
async with ClaudeSDKClient() as client:
    await client.query("Pythonとは？")
    async for message in client.receive_response():
        ...
    await client.query("どうやってインストールする？")  # 前の質問の文脈を保持
    async for message in client.receive_response():
        ...
```

> [!NOTE]
> TypeScript SDKには`ClaudeSDKClient`に相当する別クラスは存在しない。`query()`一つで両モードをカバーしており、`prompt`に文字列を渡せば単発、`AsyncIterable`（ジェネレータ）を渡せばストリーミング入力になる。割り込みは`AbortController`で行い、セッション一覧取得（`listSessions()`）やリネーム（`renameSession()`）などはクライアントのメソッドではなく個別の関数として提供される。

**使い分けの指針（公式ドキュメント準拠）**

- 単発・ステートレスでよい場合（1回きりの分析、バッチ処理、Lambda等）→ `query()`（Single Message Input）
- 対話的・画像添付・途中制御が必要な場合（チャットUI、長時間の共同作業セッション）→ `ClaudeSDKClient`（Streaming Input Mode。公式にも"推奨"と明記されている）

## エージェントループの仕組み

1. **プロンプト受信**：システムプロンプト・ツール定義・会話履歴とともにClaudeに送信。`SystemMessage(subtype="init")`がセッションメタデータとして返る
2. **評価と応答**：Claudeがテキスト応答／ツール呼び出し／その両方を返す（`AssistantMessage`）
3. **ツール実行**：SDKが要求されたツールを実行し、結果を次のターンにフィードバック（`hooks`でツール呼び出しを横取り・改変・ブロック可能）
4. **繰り返し**：ツール呼び出しが無くなるまで2〜3を繰り返す（1サイクル＝1ターン）
5. **結果返却**：最終`AssistantMessage`＋`ResultMessage`（最終テキスト・トークン使用量・コスト・セッションID）

### 具体例：ループが実際にどう動くか

「`auth.ts`の失敗しているテストを直して」という1つのプロンプトが、内部では複数ターンに分かれて処理される例：

1. **ターン1**：Claudeが`Bash`で`npm test`を実行 → `AssistantMessage`（ツール呼び出し）→ 実行結果（失敗3件）が`UserMessage`として返る
2. **ターン2**：Claudeが`Read`で`auth.ts`と`auth.test.ts`を読む → ファイル内容が返る
3. **ターン3**：Claudeが`Edit`で`auth.ts`を修正し、再度`Bash`で`npm test` → 全テスト成功
4. **最終ターン**：ツール呼び出しのないテキストのみの応答 →「Fixed the auth bug, all three tests pass now.」→ 続けて`ResultMessage`

これを`ClaudeSDKClient`（Streaming Input Mode）で受け取ると、各ターンのツール呼び出し名まで見える：

```python
import asyncio
from claude_agent_sdk import ClaudeSDKClient, ClaudeAgentOptions, AssistantMessage, ResultMessage

async def main():
    options = ClaudeAgentOptions(
        allowed_tools=["Read", "Edit", "Bash"],
        permission_mode="acceptEdits",
    )
    async with ClaudeSDKClient(options=options) as client:
        await client.query("Fix the failing tests in auth.ts")
        async for message in client.receive_response():
            if isinstance(message, AssistantMessage):
                for block in message.content:
                    if hasattr(block, "text") and block.text:
                        print(f"[thinking] {block.text}")
                    elif hasattr(block, "name"):
                        print(f"[tool call] {block.name}({getattr(block, 'input', {})})")
            elif isinstance(message, ResultMessage):
                print(f"[done] subtype={message.subtype} turns={message.num_turns} cost=${message.total_cost_usd}")

asyncio.run(main())
```

実行すると、`[tool call] Bash(...)` → `[tool call] Read(...)` → `[tool call] Edit(...)` → `[tool call] Bash(...)` → `[thinking] Fixed the auth bug...` → `[done] subtype=success turns=4 ...` のように、ターンを追って出力される。

### メッセージ種別

- `SystemMessage`：セッションライフサイクル（`init` / `compact_boundary` / `informational` / `worker_shutting_down`）
- `AssistantMessage`：各ターンのClaude応答（テキスト＋ツール呼び出し）
- `UserMessage`：ツール実行結果（ユーザーからの追加入力もここに乗る）
- `StreamEvent`：部分メッセージ有効時のみ。生のストリーミングイベント
- `ResultMessage`：ループ終了。最終テキスト・コスト・トークン使用量・`session_id`

**種別ごとの判定方法**：Pythonは`isinstance()`、TypeScriptは`message.type`文字列で判定する。

```python
import asyncio
from claude_agent_sdk import ClaudeSDKClient, SystemMessage, AssistantMessage, UserMessage, ResultMessage

async def main():
    async with ClaudeSDKClient() as client:
        await client.query("Summarize this project")
        # receive_response()は最初のResultMessageで終了する
        async for message in client.receive_response():
            if isinstance(message, SystemMessage):
                if message.subtype == "init":
                    print(f"Session started: {message.data.get('session_id')}")
                elif message.subtype == "compact_boundary":
                    print("Context was compacted")
            elif isinstance(message, AssistantMessage):
                print(f"Turn completed: {len(message.content)} content blocks")
            elif isinstance(message, UserMessage):
                print("Tool result received")
            elif isinstance(message, ResultMessage):
                if message.subtype == "success":
                    print(message.result)
                else:
                    print(f"Stopped: {message.subtype}")

asyncio.run(main())
```

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

try {
  for await (const message of query({ prompt: "Summarize this project" })) {
    switch (message.type) {
      case "system":
        if (message.subtype === "init") {
          console.log(`Session started: ${message.session_id}`);
        }
        break;
      case "assistant":
        // AssistantMessage/UserMessageは生のAPIメッセージを.messageにラップしている点に注意
        console.log(`Turn completed: ${message.message.content.length} content blocks`);
        break;
      case "user":
        console.log("Tool result received");
        break;
      case "result":
        if (message.subtype === "success") {
          console.log(message.result);
        } else {
          console.log(`Stopped: ${message.subtype}`);
        }
        break;
    }
  }
} catch (error) {
  console.log(`Session ended with an error: ${error}`);
}
```

### ループの制御オプション

| オプション | 内容 | デフォルト |
|---|---|---|
| `max_turns` / `maxTurns` | ツール使用ターンの上限 | 無制限 |
| `max_budget_usd` / `maxBudgetUsd` | コスト上限（サブエージェントの支出も合算） | 無制限 |
| `effort` | 推論の深さ（`low`/`medium`/`high`/`xhigh`/`max`） | モデルのデフォルト |
| `permission_mode` / `permissionMode` | 承認フローの制御（下記参照） | `"default"` |
| `model` | 使用モデルの明示指定（例: `"claude-sonnet-5"`） | 認証方法・契約プランに依存 |

制限に達すると`ResultMessage`の`subtype`が`error_max_turns`や`error_max_budget_usd`になる。

**設定・使用例**（本番運用を想定し、ターン数上限・推論レベル・プロジェクト設定読み込みをまとめて指定し、結果のsubtypeごとに分岐する）：

```python
import asyncio
from claude_agent_sdk import ClaudeSDKClient, ClaudeAgentOptions, ResultMessage

async def run_agent():
    session_id = None
    options = ClaudeAgentOptions(
        allowed_tools=["Read", "Edit", "Bash", "Glob", "Grep"],  # 列挙したツールは自動承認
        setting_sources=["project"],  # カレントディレクトリのCLAUDE.md/skills/hooksを読み込む
        max_turns=30,       # 暴走防止
        max_budget_usd=1.0, # コスト上限（$1を超えたら停止）
        effort="high",      # デバッグなので厚めに推論させる
        model="claude-sonnet-5",
        permission_mode="acceptEdits",
    )
    async with ClaudeSDKClient(options=options) as client:
        await client.query("Find and fix the bug causing test failures in the auth module")
        async for message in client.receive_response():
            if isinstance(message, ResultMessage):
                session_id = message.session_id
                if message.subtype == "success":
                    print(f"Done: {message.result}")
                elif message.subtype == "error_max_turns":
                    print(f"Hit turn limit. Resume session {session_id} to continue.")
                elif message.subtype == "error_max_budget_usd":
                    print("Hit budget limit.")
                else:
                    print(f"Stopped: {message.subtype}")
                if message.total_cost_usd is not None:
                    print(f"Cost: ${message.total_cost_usd:.4f}")

asyncio.run(run_agent())
```

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

let sessionId: string | undefined;

try {
  for await (const message of query({
    prompt: "Find and fix the bug causing test failures in the auth module",
    options: {
      allowedTools: ["Read", "Edit", "Bash", "Glob", "Grep"],
      settingSources: ["project"],
      maxTurns: 30,
      maxBudgetUsd: 1.0,
      effort: "high",
      model: "claude-sonnet-5",
      permissionMode: "acceptEdits"
    }
  })) {
    if (message.type === "system" && message.subtype === "init") {
      sessionId = message.session_id;
    }
    if (message.type === "result") {
      if (message.subtype === "success") {
        console.log(`Done: ${message.result}`);
      } else if (message.subtype === "error_max_turns") {
        console.log(`Hit turn limit. Resume session ${sessionId} to continue.`);
      } else if (message.subtype === "error_max_budget_usd") {
        console.log("Hit budget limit.");
      } else {
        console.log(`Stopped: ${message.subtype}`);
      }
      console.log(`Cost: $${message.total_cost_usd.toFixed(4)}`);
    }
  }
} catch (error) {
  console.log(`Session ended with an error: ${error}`);
}
```

`max_budget_usd`にはサブエージェントの支出も合算されるため、上限に達すると新規サブエージェントの起動は`Budget limit reached`で失敗し、実行中のバックグラウンドのサブエージェントも停止する。

### コンテキストウィンドウと自動圧縮（Compaction）

- コンテキストはターンをまたいでリセットされず蓄積し続ける（システムプロンプト・CLAUDE.md・ツール定義・会話履歴・ツール入出力）
- 同一内容部分（システムプロンプト・ツール定義・CLAUDE.md）はプロンプトキャッシュされ、コスト・レイテンシを削減
- コンテキストが上限に近づくと自動的に**Compaction**（古い履歴を要約して圧縮）が発生。`system/compact_boundary`メッセージが発火
- Compaction時に保持したい内容はCLAUDE.mdに書いておく（初期プロンプトは要約で失われる可能性がある）
- `PreCompact`フックで圧縮前にトランスクリプトをアーカイブ可能。`/compact`を手動送信して任意タイミングで圧縮も可能

### コンテキストを効率よく使うコツ

- サブエージェントに委譲する（サブエージェントは新規コンテキストで開始し、親には最終応答の要約だけが返る）
- サブエージェントに必要最小限のツールだけを持たせる
- MCPサーバーのツールスキーマは`Tool Search`によりデフォルトで遅延ロードされる（コンテキスト消費を抑制）
- 単純作業には`effort: "low"`を指定してトークン消費を抑える

## ビルトインツール

| カテゴリ | ツール | 内容 |
|---|---|---|
| ファイル操作 | `Read`, `Edit`, `Write` | ファイルの読み取り・変更・作成 |
| 検索 | `Glob`, `Grep` | パターン検索・正規表現によるコンテンツ検索 |
| 実行 | `Bash` | シェルコマンド・git操作等 |
| Web | `WebSearch`, `WebFetch` | Web検索・ページ取得 |
| 発見 | `ToolSearch` | ツールを事前ロードせず動的に検索・ロード |
| オーケストレーション | `Agent`, `Skill`, `AskUserQuestion`, `TaskCreate`, `TaskUpdate` | サブエージェント起動・Skill呼び出し・ユーザーへの質問・タスク管理 |

これに加えて、MCPサーバー経由での外部サービス連携、カスタムツールハンドラの定義も可能。

### 並列実行

`Read`/`Glob`/`Grep`など読み取り専用ツール（およびMCPの`readOnlyHint`が付いたツール）は並列実行され、`Edit`/`Write`/`Bash`など状態を変更するツールは競合回避のため逐次実行される。

## パーミッション（Permission Mode）

| モード | 挙動 | 用途 |
|---|---|---|
| `default` | allowルール外のツールは`canUseTool`コールバックへ（無ければ拒否） | カスタム承認フローを持つ対話型アプリ |
| `acceptEdits` | ファイル編集と一般的なファイルシステムコマンドを自動承認 | Claudeの編集を信頼して高速反復したい場合 |
| `plan` | 探索・計画のみ、編集は常にcanUseToolへ | レビュー前提で変更内容だけ提案させたい場合 |
| `dontAsk` | 一切プロンプトなし。許可ルールにあるものだけ実行、他は拒否 | headlessエージェントでツール範囲を固定したい場合 |
| `auto` | モデルによる分類器が承認/拒否を自動判定 | ガードレール付きの自律エージェント |
| `bypassPermissions` | 明示的なaskルール以外は無条件実行 | CI・コンテナ等の隔離環境限定 |

`allowed_tools` / `disallowed_tools`、および `"Bash(npm *)"` のような個別ツールへのルール指定も可能。

> [!NOTE]
> `bypassPermissions`はTypeScript SDKでは`options`に`allowDangerouslySkipPermissions: true`も明示しないと有効にならない。またUnix環境でroot実行時には使用不可。エージェントの操作が影響してよい隔離環境でのみ使うこと。

## ClaudeAgentOptions（Options）全体マップ

`query()`/`ClaudeSDKClient`（TSは`query()`）に渡す設定オブジェクト。Pythonは`snake_case`のdataclass（`ClaudeAgentOptions`）、TypeScriptは`camelCase`のオブジェクトリテラル（`Options`型）で、フィールドはほぼ1対1対応する。既出の`allowed_tools`/`disallowed_tools`/`setting_sources`/`max_turns`/`max_budget_usd`/`effort`/`model`/`permission_mode`以外の主要フィールドを整理する。

### システムプロンプト・実行環境

| Python | TypeScript | 内容 |
|---|---|---|
| `system_prompt` | `systemPrompt` | カスタムプロンプト文字列、または`{"type": "preset", "preset": "claude_code", "append": "..."}`でClaude Codeのデフォルトプロンプトに追記 |
| `cwd` | `cwd` | 作業ディレクトリ |
| `add_dirs` | `addDirs` | アクセスを許可する追加ディレクトリ |
| `env` | `env` | 子プロセスに渡す環境変数 |
| `cli_path` | `cliPath` | 使用するClaude Code CLI実行ファイルのパス |
| `extra_args` | `extraArgs` | CLIへの追加引数 |
| `stderr` | `stderr` | CLIのstderr出力を受け取るコールバック |

```python
options = ClaudeAgentOptions(
    system_prompt={"type": "preset", "preset": "claude_code", "append": "Always write tests first."},
    cwd="/path/to/project",
    env={"NODE_ENV": "development"},
)
```

```typescript
const options = {
  systemPrompt: { type: "preset", preset: "claude_code", append: "Always write tests first." },
  cwd: "/path/to/project",
  env: { NODE_ENV: "development" },
};
```

### セッション管理

| Python | TypeScript | 内容 |
|---|---|---|
| `continue_conversation` | `continueConversation` | 直近セッションを自動的に継続 |
| `resume` | `resume` | 指定セッションIDから再開 |
| `fork_session` | `forkSession` | 再開時、元セッションを変更せず新IDに分岐 |
| `session_store` | `sessionStore` | ステートレス環境向けにトランスクリプトを外部バックエンドへミラーリング |

### MCP・エージェント・スキル

| Python | TypeScript | 内容 |
|---|---|---|
| `mcp_servers` | `mcpServers` | MCPサーバー構成（詳細は[MCP連携](#mcpmodel-context-protocol連携)節） |
| `agents` | `agents` | `AgentDefinition`によるサブエージェントのプログラム的定義 |
| `skills` | `skills` | ロードするSkillの指定（`"all"`または名前リスト） |
| `hooks` | `hooks` | フック登録（詳細は[フック](#フックhooks)節） |

`AgentDefinition`の例：

```python
agents = {
    "code-reviewer": AgentDefinition(
        description="Reviews code changes",
        prompt="You are a meticulous code reviewer...",
        tools=["Read", "Grep"],
        model="claude-sonnet-5",
    )
}
options = ClaudeAgentOptions(agents=agents)
```

```typescript
const options = {
  agents: {
    "code-reviewer": {
      description: "Reviews code changes",
      prompt: "You are a meticulous code reviewer...",
      tools: ["Read", "Grep"],
      model: "claude-sonnet-5",
    },
  },
};
```

### 権限・出力制御

| Python | TypeScript | 内容 |
|---|---|---|
| `can_use_tool` | `canUseTool` | ツール実行可否を判定する自前コールバック（`permission_mode="default"`でallowルール外のツールが渡る） |
| `permission_prompt_tool_name` | `permissionPromptToolName` | 権限確認に使うMCPツール名を指定 |
| `include_partial_messages` | `includePartialMessages` | ストリーミングの生イベント（`StreamEvent`）を含めるか |
| `output_format` | `outputFormat` | JSON Schemaによる構造化出力の指定 |

`can_use_tool`の例（Pythonでは`can_use_tool`未指定時、allowルール外は既定でユーザー確認が発生する点に注意）：

```python
async def approve_reads_only(request):
    if request.tool_name in ("Read", "Grep", "Glob"):
        return {"approved": True}
    return {"approved": False, "reason": "Only read-only tools are auto-approved"}

options = ClaudeAgentOptions(can_use_tool=approve_reads_only)
```

### その他

| Python | TypeScript | 内容 |
|---|---|---|
| `thinking` | `thinking` | 拡張思考の有効化・トークン予算（例:`{"type": "enabled", "budget_tokens": 10000}`） |
| `enable_file_checkpointing` | `enableFileCheckpointing` | ファイル変更を追跡し、`rewind_files()`等でのリワインドを可能にする |
| `settings` | `settings` | 読み込む設定ファイルのパス |
| `betas` | `betas` | ベータ機能フラグの有効化 |

## フック（Hooks）

エージェントループの特定タイミングで自分のプロセス内でコールバックを実行できる（コンテキストは消費しない）。`ClaudeAgentOptions`（TSは`options`）の`hooks`フィールドに、イベント名をキーとした辞書で登録する。

### 主要フックイベント

| フック | 発火タイミング | 用途例 |
|---|---|---|
| `PreToolUse` | ツール実行前（ブロック・入力改変可能） | 入力検証、危険なコマンドのブロック |
| `PostToolUse` | ツール実行結果後 | 出力監査、副作用のトリガー |
| `PostToolUseFailure` | ツール実行がエラーになった時 | エラーハンドリング、ロギング |
| `UserPromptSubmit` | プロンプト送信時 | 追加コンテキストの注入 |
| `Stop` | エージェント終了時 | 結果検証、セッション状態の保存 |
| `SubagentStart` / `SubagentStop` | サブエージェント開始/終了時 | 並列タスク結果の追跡・集約 |
| `PreCompact` | コンテキスト圧縮前 | 全トランスクリプトのアーカイブ |
| `PermissionRequest` | ツール呼び出しが権限判定を必要とする時 | カスタム権限ハンドリング |
| `Notification` | エージェントのステータス通知時 | Slack/PagerDuty等への転送 |

> [!NOTE]
> 上記に加えてTypeScript SDKには`SessionStart`/`SessionEnd`（セッション開始/終了）、`PreModelSwitch`/`PostModelSwitch`（モデル切替前後）、`PermissionDenied`、`ConfigChange`、`FileChanged`など、より粒度の細かいフックイベントが多数存在する（バージョンにより追加される可能性があるため詳細は公式リファレンス参照）。Python SDKでは主要フック（上表）が中心。

### 登録方法（`HookMatcher` / `matcher`）

- キー：フックイベント名（例:`"PreToolUse"`）
- 値：`HookMatcher`（TSはオブジェクトリテラル）のリスト。`matcher`でツール名にマッチさせ（未指定なら全イベント対象）、`hooks`にコールバック関数の配列を渡す
- ツール系フックの`matcher`はツール名の文字列（`"Bash"`、`"Write|Edit"`のような正規表現も可）。MCPツールは`mcp__<サーバー名>__<ツール名>`形式でマッチさせる

```python
from claude_agent_sdk import ClaudeAgentOptions, HookMatcher

async def block_dangerous_bash(input_data, tool_use_id, context):
    if input_data["tool_name"] == "Bash" and "rm -rf" in input_data["tool_input"].get("command", ""):
        return {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": "Dangerous command blocked",
            }
        }
    return {}

options = ClaudeAgentOptions(
    hooks={
        "PreToolUse": [HookMatcher(matcher="Bash", hooks=[block_dangerous_bash])],
    }
)
```

```typescript
import { query, HookCallback, PreToolUseHookInput } from "@anthropic-ai/claude-agent-sdk";

const blockDangerousBash: HookCallback = async (input) => {
  const preInput = input as PreToolUseHookInput;
  const command = (preInput.tool_input as { command?: string }).command ?? "";
  if (preInput.tool_name === "Bash" && command.includes("rm -rf")) {
    return {
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: "Dangerous command blocked",
      },
    };
  }
  return {};
};

for await (const message of query({
  prompt: "Clean up the temp directory",
  options: { hooks: { PreToolUse: [{ matcher: "Bash", hooks: [blockDangerousBash] }] } },
})) {
  // ...
}
```

### 入出力の型

- **入力（`HookInput`）共通フィールド**：`session_id`、`cwd`、`hook_event_name`。`PreToolUse`/`PostToolUse`はさらに`tool_name`・`tool_input`（`PostToolUse`は`tool_output`も）・`tool_use_id`を持つ
- **出力（`HookOutput`）**：`systemMessage`（ユーザーへの表示メッセージ）、`continue_`/`continue`（エージェント継続可否）、`hookSpecificOutput`（イベント固有の内容）を返せる。何も制御しない場合は`{}`を返せばよい

### `PreToolUse`でのパーミッション制御

`hookSpecificOutput.permissionDecision`に以下を指定してツール実行を制御できる：

| 値 | 挙動 |
|---|---|
| `"allow"` | プロンプトなしで自動承認 |
| `"deny"` | ツール実行をブロック（Claudeには拒否メッセージが返る） |
| `"ask"` | ユーザーに確認を求める（未指定時のデフォルト） |

`updatedInput`を一緒に返すと、ツールへの入力そのものを改変してから実行させることも可能（例：書き込み先パスをサンドボックス配下にリダイレクト）。

## サブエージェント（Subagents）

- 専門特化したタスクを親から切り離して実行する子エージェント
- 新規の会話履歴で開始し、親のターンは見えない。最終応答のみが親にツール結果として返る（親のコンテキストは要約分しか増えない）
- 親の権限を自動継承しない（`PreToolUse`フックや権限ルールで個別に許可設定が必要）
- `AgentDefinition`の`tools`フィールドでツールを絞り込み、`effort`でセッション全体とは別の推論レベルを指定可能
- `max_budget_usd`の上限にはサブエージェントの支出も合算される

## MCP（Model Context Protocol）連携

- 外部ツール・データソース（DB、ブラウザ、API等）をMCPサーバー経由で接続
- MCPツール名は必ず `mcp__<サーバー名>__<アクション>` の形式（例: `mcp__playwright__browser_screenshot`）
- デフォルトでMCPツールのスキーマは遅延ロード（Tool Search）され、コンテキスト消費を抑える。未対応モデル/環境では事前ロードにフォールバック

## セッション管理

- `ResultMessage.session_id`（TypeScriptは`init`のSystemMessageにも直接あり）でセッションIDを取得し、後で resume 可能
- resume時は過去ターンの全コンテキスト（読み取ったファイル・実施した分析や操作）が復元される
- セッションをforkして、元セッションを変更せずに別アプローチへ分岐することも可能
- ステートレスなコンテナ/サーバーレス環境では`session_store`アダプタで自前バックエンドにトランスクリプトをミラーリング可能

## Skills / Commands / Memory / Plugins

- プロジェクトの`.claude/`、ユーザーの`~/.claude/`から、Claude Codeと同様にSkills・Commands・Memory（CLAUDE.md等）を自動ロード
- Pluginsとして、Skills・Agents・Hooks・MCPサーバーをパッケージ化し、ローカルパスからロード可能

## 料金・課金（2026年9月時点）

- 2026年6月15日より導入予定だった「Claude Agent SDK専用の月額クレジット」制度（Pro: $20 / Max 5x: $100 / Max 20x: $200、Team/Enterpriseにも別枠あり）は、公式ヘルプセンターの告知によれば**導入前に一時停止（"Update June 15: We're pausing the changes..."）** されている
- 現状（2026年9月時点）も変更前のままで、Claude Agent SDK・`claude -p`・サードパーティアプリ経由の利用はすべてサブスクリプションの通常利用枠（usage limits）から消費される。専用クレジット制度は未提供
- Anthropicはサブスクリプションベースでの Claude Agent SDK 利用をより適切にサポートするための改定を検討中とアナウンスしており、変更前に改めて告知予定
- （参考）別製品の**Claude Managed Agents**は従量課金で、モデル推論はClaude Platform標準トークン単価、加えて稼働中のセッション時間あたり$0.08/session-hourが発生（月額固定費なし）

## ユースケース例

- コードのバグ発見・自動修正エージェント
- テスト失敗の自動診断・修正
- リサーチエージェント、メールアシスタントなど（公式のexample集: `claude-agent-sdk-demos`）
- CI/CDやDocker上でのheadless自動化エージェント

## 参考リンク

- [Agent SDK Overview（公式）](https://code.claude.com/docs/en/agent-sdk/overview)
- [Quickstart（公式）](https://code.claude.com/docs/en/agent-sdk/quickstart)
- [Agent Loopの仕組み（公式）](https://code.claude.com/docs/en/agent-sdk/agent-loop)
- [TypeScript SDK Reference](https://code.claude.com/docs/en/agent-sdk/typescript)
- [Python SDK Reference](https://code.claude.com/docs/en/agent-sdk/python)
- [TypeScript SDK GitHub / Changelog](https://github.com/anthropics/claude-agent-sdk-typescript)
- [Python SDK GitHub / Changelog](https://github.com/anthropics/claude-agent-sdk-python)
- [Example agents（デモ集）](https://github.com/anthropics/claude-agent-sdk-demos)
- [Agent SDK課金についてのHelp Center記事](https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan)
