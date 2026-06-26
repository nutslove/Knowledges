- 参考URL
  - https://docs.langchain.com/oss/python/deepagents/overview

# 概要
- DeepAgentsはLangChainのReact Agentを拡張したもので、複数のbuilt-inのMiddleWareが最初から組み込まれている
- DeepAgentsもSubAgentsもLangchainの`create_agent`で作られていて、どちらもReact Agentである
  - https://github.com/langchain-ai/deepagents/blob/master/libs/deepagents/deepagents/graph.py#L149  

    ```python
    from langchain.agents import create_agent

    ・・・中略・・・

    def create_deep_agent(
        model: str | BaseChatModel | None = None,
        tools: Sequence[BaseTool | Callable | dict[str, Any]] | None = None,
        *,
        system_prompt: str | None = None,
        middleware: Sequence[AgentMiddleware] = (),
        subagents: list[SubAgent | CompiledSubAgent] | None = None,
        response_format: ResponseFormat | None = None,
        context_schema: type[Any] | None = None,
        checkpointer: Checkpointer | None = None,
        store: BaseStore | None = None,
        backend: BackendProtocol | BackendFactory | None = None,
        interrupt_on: dict[str, bool | InterruptOnConfig] | None = None,
        debug: bool = False,
        name: str | None = None,
        cache: BaseCache | None = None,
    ) -> CompiledStateGraph:

        ・・・中略・・・

        return create_agent(
            model,
            system_prompt=system_prompt + "\n\n" + BASE_AGENT_PROMPT if system_prompt else BASE_AGENT_PROMPT,
            tools=tools,
            middleware=deepagent_middleware,
            response_format=response_format,
            context_schema=context_schema,
            checkpointer=checkpointer,
            store=store,
            debug=debug,
            name=name,
            cache=cache,
        ).with_config({"recursion_limit": 1000})
    ```
  - https://github.com/langchain-ai/deepagents/blob/master/libs/deepagents/deepagents/middleware/subagents.py#L244

---

# ToolのOutputがでかい場合の退避について

## 概要

DeepAgents（`create_deep_agent` / `FilesystemMiddleware`）は、ツール出力が大きすぎる場合に **原文をバックエンドに退避し、LLM にはプレビュー（head+tail）＋退避先パスだけを返す**。LLM は必要に応じて `read_file` / `grep` で原文を取得する。

## 目的

主に 2 つ:

1. **コンテキストウィンドウの上限超過を防ぐ（最重要）**: 巨大なツール出力をそのまま渡すとモデルの入力トークン上限を超え、リクエスト自体が失敗する。退避はこれを構造的に回避する。
2. **入力トークンの節約**: 上限内であっても、不要な原文を毎回モデルに渡さないことでトークン消費（コスト）と処理負荷を抑える。

## 退避の発火条件

- 対象は **1 回の ToolMessage の content**。
- しきい値: `tool_token_limit_before_evict`（デフォルト **20,000 トークン**）× `NUM_CHARS_PER_TOKEN`（**4**）= **約 80,000 文字**。
- これを超えると退避（eviction）が発火する。
- 発火箇所は `FilesystemMiddleware.wrap_tool_call` / `awrap_tool_call`。**除外ツール以外のすべてのツール**に効く。
- 除外ツール（退避されない）: `ls`, `glob`, `grep`, `read_file`, `edit_file`, `write_file`。

### 閾値の変更可否

- `FilesystemMiddleware` 自体は閾値を引数で受け付ける:
  - `tool_token_limit_before_evict`（デフォルト `20000`、`None` で退避を無効化）
  - `human_message_token_limit_before_evict`（デフォルト `50000`、HumanMessage 用の別枠）
- **ただし `create_deep_agent` は `FilesystemMiddleware` を内部でハードコード生成しており（`backend` / `custom_tool_descriptions` / `_permissions` のみ渡す）、閾値を設定したインスタンスを差し込む口が無い。** `create_deep_agent` 自身にも閾値の引数は無い。
- したがって正しい理解は:
  > **ライブラリの能力としては変更可能。ただし `create_deep_agent` というファサード越しでは塞がれており、現状の構成では変更できない。**
- 変えたい場合の選択肢:
  1. `create_deep_agent` を使わず、`create_agent` + 自前の middleware スタックで `FilesystemMiddleware(tool_token_limit_before_evict=...)` を明示的に生成する（正攻法）。
  2. しきい値はそのままに、MCP 段階の interceptor 等でモデルに渡る量を手前で調整する（退避ではなく切り捨て/圧縮）。
  3. deepagents のバージョンを上げ、`create_deep_agent` に閾値引数が追加されていないか確認する。

## 退避の動作

1. 原文を `backend.write` でバックエンドに保存。保存パスは `/large_tool_results/<tool_call_id>`。
2. LLM 向けの ToolMessage を次の文言に置換する:
   > `Tool result too large, the result of this tool call <id> was saved in the filesystem at this path: /large_tool_results/<id>`（＋ head/tail プレビュー）
3. LLM は `read_file`（offset/limit でページング）や `/large_tool_results/` 配下の `grep` で原文を参照する。

## 退避先（バックエンド）

退避された **ツール出力の原文がどこに保存されるか** は、バックエンドと checkpointer の設定で決まる。

- `create_deep_agent` に `backend=` を渡さない場合、デフォルトは **`StateBackend`**。
- **デフォルト（`StateBackend`）の場合 → 原文はメモリに入る。**
  - 原文は LangGraph の state の **`files` チャネル（メモリ上の dict）** に格納される。ローカルディスクのファイルは作られない。
  - **checkpointer を設定していなければ、原文はメモリ上だけに存在し、実行が終わると消える（永続化されない）。**
- **`checkpointer`（`AsyncPostgresSaver`）を設定している場合 → 原文はそこ（RDS）に入る。**
  - state（`files` チャネル＝退避されたツール出力の原文を含む）が checkpointer 経由で **RDS（PostgreSQL）にシリアライズ永続化**される。
  - つまり「退避されたツール出力の原文の最終的な保存先 = checkpointer のストレージ（このプロジェクトでは RDS）」になる。
  - 同じ `thread_id` の範囲では checkpoint 経由で復元される。別スレッドには引き継がれない。

## 退避先を別 backend に分離する（CompositeBackend）

`create_deep_agent` の `backend=` には backend を 1 つしか渡せないが、**`CompositeBackend`** を使えばパスのプレフィックスで複数 backend に振り分けられる。これを使って **ツール出力の退避先だけを別ストレージに逃がす**ことができる。

### 退避パスの決まり方

`FilesystemMiddleware` は退避パスを次のように組み立てる（[filesystem.py:864-866]）:

```python
artifacts_root = backend.artifacts_root if isinstance(backend, CompositeBackend) else "/"
_root = artifacts_root.rstrip("/")
self._large_tool_results_prefix = f"{_root}/large_tool_results"
```

- `CompositeBackend` **以外**（= デフォルトの `StateBackend`）→ `artifacts_root` は強制的に `"/"` → 退避先は **`/large_tool_results/<id>`**。
- `CompositeBackend` → そのインスタンスの `artifacts_root`（デフォルト `"/"`）が使われる。

### ルーティング規則

`CompositeBackend` は **最長プレフィックス一致**でルーティングし、どの route にも一致しないパスは `default` に流す（[composite.py:104-115]）。

### 注意: route を足すだけでは退避先は変わらない

例えば次のように `/memories/` だけを Store にしても:

```python
CompositeBackend(default=StateBackend(), routes={"/memories/": StoreBackend()})
```

- `artifacts_root` は `"/"` のままなので退避パスは `/large_tool_results/<id>`。
- これは `/memories/` に一致しないため **`default`（= StateBackend）に退避される**（= 現状と同じく state → RDS `checkpoint_writes`）。

### 退避物を別 backend に逃がす書き方

`artifacts_root` を設定し、それにマッチする route を張る:

```python
from deepagents.backends import CompositeBackend, StateBackend, StoreBackend

backend = CompositeBackend(
    default=StateBackend(),
    routes={
        "/cache/": StoreBackend(namespace=lambda rt: (rt.server_info.user.identity, "cache")),
    },
    artifacts_root="/cache/",   # ← 退避物のルートをここに変える
)
agent = create_deep_agent(..., backend=backend)
```

- 退避パスが **`/cache/large_tool_results/<id>`** になり、
- `/cache/` route → **StoreBackend** にルーティングされる。
- → **退避原文が checkpoint（RDS `checkpoint_writes`）ではなく store 側に保存される**。これにより DeltaChannel / checkpoint 肥大化を緩和できる。

### まとめ

| 構成 | ツール出力の退避先 |
|---|---|
| 現状（backend 未指定 = StateBackend）| `/large_tool_results/` → **state → RDS `checkpoint_writes`** |
| Composite（`artifacts_root="/"`、退避用 route なし）| 同上（**default = StateBackend**）|
| Composite（`artifacts_root="/cache/"` + `/cache/`→Store）| **`/cache/large_tool_results/` → StoreBackend** |

> [!CAUTION]
> 退避物を Store に逃がすと「短期メモリ（スレッド単位で消える想定）」ではなく **store の永続データ**になる。不要になった `large_tool_results` の削除（クリーンアップ）を別途設計する必要がある。

### backend 別のスコープ（参考）

| backend | スコープ | 永続性 |
|---|---|---|
| `StateBackend` | スレッド（`thread_id`）単位 | checkpointer 経由で同一スレッド内のターンをまたいで保持。別スレッドには引き継がれない。checkpointer 無しなら 1 実行で消える（短期メモリ）|
| `StoreBackend` | namespace 単位（スレッドまたぎ）| store に永続。namespace に user_id 等を入れてユーザー横断の長期メモリにできる |

## RDS 上での保存場所（重要な注意点）

- `files` フィールドは **`DeltaChannel(snapshot_frequency=50)`** という特殊チャネル。通常チャネルのように毎ステップの値を `checkpoint_blobs` に書かない。
- **退避された原文の実体は `checkpoint_writes` テーブル**（`channel='files'`, `type='msgpack'`）に delta（PendingWrite）として入る。退避が起きたステップで書かれる。
- **通常ステップ**では `DeltaChannel.checkpoint()` が `MISSING`（センチネル）を返すため、`files` は `channel_values` に含まれず `checkpoint_blobs` に行が作られない。状態は ancestor の writes を辿って再構築する（`get_delta_channel_history`）。
- **スナップショット発火時のみ** `checkpoint_blobs` に `files` の行（`_DeltaSnapshot` blob = `files` 全体）が作られる（発火条件は「コスト・運用上の注意」参照）。
  - → スナップショットが一度も発火していなければ、退避が起きていても `SELECT ... FROM checkpoint_blobs WHERE channel='files'` は **0 件**になる（実際にこの状態だった。原文は `checkpoint_writes` 側にある）。
- DB は接続文字列の `CHECKPOINT_POSTGRES_DB`。

## 原文を確認する方法（SQL）

```sql
-- 退避が起きたか（messages 側のマーカー）
SELECT count(*) FROM checkpoint_blobs
WHERE channel='messages'
  AND position('large_tool_results' in encode(blob,'escape')) > 0;

-- 退避された原文をサイズ順に（実体は checkpoint_writes）
SELECT thread_id, octet_length(blob) AS bytes
FROM checkpoint_writes
WHERE channel='files'
ORDER BY bytes DESC;

-- 中身を読む（JSON 部分は読める。前後に msgpack のフレーミングバイトが混ざる）
SELECT encode(blob,'escape')
FROM checkpoint_writes
WHERE channel='files' AND thread_id='<thread_id>'
ORDER BY octet_length(blob) DESC LIMIT 1;
```

完全に構造化して取り出したい場合は、**コンパイル済みエージェントの `aget_state()`** を使う:

```python
config = {"configurable": {"thread_id": "<thread_id>"}}
state = await agent.aget_state(config)
content = state.values["files"]["/large_tool_results/<id>"]["content"]
```

> [!CAUTION]
> 素の `checkpointer.aget(config)["channel_values"]["files"]` では取れない。`files` は `DeltaChannel` で、非スナップショットステップでは `channel_values` にセンチネルしか入らず、delta の再構築（ancestor writes を辿る `get_delta_channel_history`）は Pregel ループ / `aget_state()` 側でしか行われないため。SQL で直接読む場合も再構築は自前で行う必要があり、上記 SQL は「個々の write の生バイト」を見るもの（複数 write にまたがる場合は最新の状態 = 全 write を reducer で畳み込んだ結果になる点に注意）。

## コスト・運用上の注意

- 退避は **入力トークン（LLM 課金）は節約する**（退避原文はモデルには毎回渡らず、プレビューのみ）。一方で **DB I/O・ストレージのコストは残る**。
- checkpointer の読み書き粒度:
  - **読み込み（DB→メモリ）**: 各ユーザーターン（`ainvoke`）の開始時に 1 回だけ最新 state をロード・デシリアライズして復元する。1 回の `ainvoke` 実行中は state はメモリ上に保持され、ステップごとの再ロードはしない。
  - **書き込み（メモリ→DB）**: 実行の最後に 1 回ではなく、**superstep（ノード境界）ごとに checkpoint を永続化**する（耐障害性・resume のため）。
- `files`（退避原文）固有のコスト:
  - 原文の書き込みは退避が起きたそのステップで `checkpoint_writes` に書かれる。以降 `files` が更新されなければ delta は追加されない（毎ステップ巨大データを書き直すわけではない）。
  - スナップショット（`files` 全体を `_DeltaSnapshot` として `checkpoint_blobs` に再シリアライズ）は **次のいずれか**で発火する（`snapshot_frequency=50`）:
    - `files` チャネルへの**更新回数が 50 回**たまったとき（= 退避が 50 回累積。**「50 ステップごと」ではない**。`files` を更新しない superstep は更新回数を増やさない）、または
    - 最後のスナップショットからの**経過 superstep 数が `DELTA_MAX_SUPERSTEPS_SINCE_SNAPSHOT`（デフォルト 5000、env `LANGGRAPH_DELTA_MAX_SUPERSTEPS_SINCE_SNAPSHOT` で変更可）に達したとき**。
  - したがって通常の会話（退避が数回・総ステップ 5000 未満）では**スナップショットは発火せず**、退避原文の保存・復元はもっぱら `checkpoint_writes` 経由になる。
- BYTEA の上限は 1GB なので数十万〜数百万文字でも文字数制限には引っかからない。大きい値は TOAST で圧縮・別領域保存される。
- 古いバージョンや終了スレッドの writes/blobs は自動削除されない。長期運用ではスレッドの定期クリーンアップ（`checkpointer.adelete_thread` 等）を検討する。
