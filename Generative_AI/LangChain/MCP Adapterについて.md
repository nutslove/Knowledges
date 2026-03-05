# MCP Adapterについて
- https://github.com/langchain-ai/langchain-mcp-adapters

# LangChain MCP AdaptersのToolのエラーハンドリング
- LangChain MCP Adaptersライブラリでは、StructuredTool作成時にhandle_tool_errorパラメータを渡していない（2026-03-03時点）
  - https://github.com/langchain-ai/langchain-mcp-adapters/issues/263
  - https://github.com/langchain-ai/langchain-mcp-adapters/pull/264 （未マージ）
- そのため、MCP Toolでエラーが発生すると`ToolException`がraiseされ、Agentが異常終了する
  - `langchain_mcp_adapters/tools.py`の`_convert_call_tool_result()`で`raise ToolException(error_msg)`している
- LangGraph側でも`ToolException`を適切にキャッチできていない問題がある
  - https://github.com/langchain-ai/langgraph/issues/6449
  - https://github.com/langchain-ai/langgraph/pull/6508 （未マージ）
  - これがマージされれば、`ToolException`が自動的にエラーメッセージに変換される
- 現時点の対策として、MCPツールをラップしてエラーを文字列として返すようにする必要がある

```python
from langchain_core.tools import StructuredTool

def wrap_mcp_tools_with_error_handling(mcp_tools: list) -> list:
    """MCPツールをエラーハンドリングでラップする"""
    wrapped_tools = []
    for tool in mcp_tools:
        original_coroutine = tool.coroutine

        async def wrapped_coroutine(*args, _original=original_coroutine, _name=tool.name, **kwargs):
            try:
                return await _original(*args, **kwargs)
            except Exception as e:
                return f"Tool '{_name}' failed: {repr(e)}"

        wrapped_tool = StructuredTool(
            name=tool.name,
            description=tool.description,
            coroutine=wrapped_coroutine,
            args_schema=tool.args_schema,
        )
        wrapped_tools.append(wrapped_tool)
    return wrapped_tools

# 使用例
client = _get_mcp_client()
mcp_tools = await client.get_tools()
mcp_tools = wrap_mcp_tools_with_error_handling(mcp_tools)  # ラップを適用
native_tools = [save_user_memory, get_user_memory, delete_user_memory]
tools = mcp_tools + native_tools
```

---

# `MultiServerMCPClient`の`tool_interceptors`について

## 概要
`tool_interceptors` は、`langchain-mcp-adapters` の `MultiServerMCPClient` において、MCP ツール呼び出しの**前後に処理を差し込むミドルウェア（インターセプター）** のリスト。

「オニオンパターン」で構成され、リストの **最初のインターセプターが最も外側** となる。

## 必要なライブラリ

| パッケージ | 用途 |
|---|---|
| `langchain-mcp-adapters` | `MultiServerMCPClient`, `MCPToolCallRequest` 等 |
| `mcp` | `CallToolResult`, `TextContent` 等（`langchain-mcp-adapters` の依存で自動インストールされる） |

> `fastmcp` をインストールしている場合も `mcp` は推移的依存で入るため、追加インストールは不要。

## インターセプターのインターフェース

### `ToolCallInterceptor` プロトコル

```python
from langchain_mcp_adapters.interceptors import MCPToolCallRequest

async def my_interceptor(
    request: MCPToolCallRequest,
    handler,  # Callable[[MCPToolCallRequest], Awaitable[MCPToolCallResult]]
) -> MCPToolCallResult:
    ...
```

- **`request`**: ツール呼び出しリクエスト（詳細は後述「`MCPToolCallRequest` の正体」を参照）
- **`handler`**: 次のインターセプター、または最終的なツール実行を呼ぶ関数

### `MCPToolCallRequest` の正体

`MCPToolCallRequest` は **MCP SDK (`mcp` パッケージ) の型ではなく、`langchain-mcp-adapters` 独自の `@dataclass`** である。

「LLM が生成した tool_call の情報 ＋ MCP サーバーへの接続コンテキスト」を1つにまとめたラッパーオブジェクトで、インターセプターがリクエストの前処理・書き換えをするためのインターフェースとして設計されている。

#### 生成される箇所（`tools.py`）

```python
# langchain_mcp_adapters/tools.py より
request = MCPToolCallRequest(
    name=tool.name,                        # MCP Tool の名前
    args=arguments,                        # LLM が生成したツール引数（dict）
    server_name=server_name or "unknown",  # MCP サーバー名（後述）
    headers=None,                          # HTTP ヘッダー（初期値は None）
    runtime=runtime,                       # LangGraph ランタイムコンテキスト
)
```

#### 最終的な使われ方

インターセプターチェーンを通過した後、最内部の `execute_tool` 関数内で `request.name` と `request.args` が取り出され、**実際の MCP SDK の `session.call_tool()` に渡される**：

```python
# langchain_mcp_adapters/tools.py の execute_tool 内部
async def execute_tool(request: MCPToolCallRequest) -> MCPToolCallResult:
    tool_name = request.name      # ← request から取り出す
    tool_args = request.args      # ← request から取り出す
    # ...
    call_tool_result = await tool_session.call_tool(
        tool_name,
        tool_args,
        # ...
    )
```

つまり、インターセプターで `request.override(name=..., args=..., headers=...)` を使って書き換えると、**実際の MCP 呼び出し内容がそのまま変わる**。

#### まとめ

| 観点 | 説明 |
|---|---|
| MCP SDK の型ではない | `langchain-mcp-adapters` 独自の `@dataclass` |
| 中身は MCP Tool 呼び出しに必要な情報 | `name`（ツール名）と `args`（引数）は最終的にそのまま `session.call_tool(name, args)` に渡される |
| 追加のコンテキスト情報も含む | `server_name`、`headers`、`runtime` など、MCP SDK 単体では持たないメタ情報も付与されている |
| インターセプターで書き換え可能 | `request.override(...)` で変更すると、実際の MCP 呼び出し内容が変わる |

### `MCPToolCallRequest` のフィールド

| フィールド | 種別 | 説明 |
|---|---|---|
| `name` | 変更可能 | ツール名��最終的に `session.call_tool()` の第1引数になる） |
| `args` | 変更可能 | ツール引数 dict（最終的に `session.call_tool()` の第2引数になる） |
| `headers` | 変更可能 | HTTPヘッダー（SSE/HTTPトランスポート用） |
| `server_name` | 読み取り専用 | MCPサーバー名（後述「`server_name` の由来」を参照） |
| `runtime` | 読み取り専用 | LangGraph ランタイムコンテキスト |

変更は `request.override(args={...})` でイミュータブルに行う。

### `server_name` に `MultiServerMCPClient` のどこの値が入るか

`request.server_name` には、**`MultiServerMCPClient` のコンストラクタに渡す `connections` ディクショナリの「キー名」** がそのまま入る。

#### データの流れ

```
MultiServerMCPClient({
    "observability": {...},     ← dict のキー名
    "aws_cli": {...},
})
    │
    ▼  get_tools() で self.connections.items() をイテレート
    │
    name = "observability"      ← dict のキー名がそのまま
    │
    ▼  load_mcp_tools(server_name=name) に渡される
    │
    ▼  convert_mcp_tool_to_langchain_tool(server_name=name) に渡される
    │
    ▼  MCPToolCallRequest(server_name=server_name or "unknown")
    │
    request.server_name == "observability"  ✅
```

#### ソースコード上の流れ

**Step 1: コンストラクタで `connections` を保持**

```python
# langchain_mcp_adapters/client.py
self.connections: dict[str, Connection] = (
    connections if connections is not None else {}
)
```

**Step 2: `get_tools()` でキー名を `server_name` として渡す**

```python
# langchain_mcp_adapters/client.py
for name, connection in self.connections.items():
    load_mcp_tool_task = asyncio.create_task(
        load_mcp_tools(
            None,
            connection=connection,
            server_name=name,              # ← dict のキー名
            tool_interceptors=self.tool_interceptors,
            # ...
        )
    )
```

**Step 3: ツール呼び出し時に `MCPToolCallRequest` にセット**

```python
# langchain_mcp_adapters/tools.py
request = MCPToolCallRequest(
    name=tool.name,
    args=arguments,
    server_name=server_name or "unknown",  # ← ここで最終的にセット
    headers=None,
    runtime=runtime,
)
```

#### 具体例

```python
_client = MultiServerMCPClient(
    {
        "observability": {...},   # → request.server_name == "observability"
        "aws_cli": {...},         # → request.server_name == "aws_cli"
    },
    tool_interceptors=[truncate_interceptor],
)
```

> **補足**: `load_mcp_tools` を直接使って `server_name` 引数を省略した場合は、フォールバックとして `"unknown"` が設定される。

### `handler` の戻り値（`MCPToolCallResult`）

| 型 | 説明 |
|---|---|
| `CallToolResult` | MCP SDK の標準戻り値 |
| `ToolMessage` | LangChain のメッセージ形式 |
| `Command` | LangGraph の状態更新（langgraph インストール時のみ） |

## `CallToolResult.content` に入りうる型

| 型 | 内容 | `.text` 属性 |
|---|---|---|
| `TextContent` | テキスト | ✅ あり |
| `ImageContent` | Base64画像 | ❌ なし |
| `AudioContent` | 音声 | ❌ なし |
| `ResourceLink` | リソースリンク | ❌ なし |
| `EmbeddedResource` | 埋め込みリソース | ❌ なし |

> `isinstance(block, TextContent)` でチェックするのは、TextContent 以外に `.text` を呼ぶと `AttributeError` になるための防御的コーディング。

## 特定サーバーにだけ適用する方法

`tool_interceptors` はグローバルに全サーバーに適用されるが、**インターセプター内部で `request.server_name` をチェック**することで特定サーバーのみに適用可能。

```python
async def my_interceptor(request, handler):
    if request.server_name != "observability":
        return await handler(request)  # 素通り

    # observability サーバーのみの処理
    result = await handler(request)
    # ... 加工 ...
    return result
```

## インターセプターのパターン集

### 1. ログ記録

```python
async def logging_interceptor(request, handler):
    print(f"Before: {request.name}({request.args})")
    result = await handler(request)
    print(f"After: {request.name}")
    return result
```

### 2. キャッシュ（handler をスキップ）

```python
cache = {}

async def caching_interceptor(request, handler):
    cache_key = f"{request.name}:{request.args}"
    if cache_key in cache:
        return cache[cache_key]  # ショートサーキット
    result = await handler(request)
    cache[cache_key] = result
    return result
```

### 3. リクエスト書き換え

```python
async def modify_args_interceptor(request, handler):
    new_request = request.override(args={"a": request.args["a"] * 2})
    return await handler(new_request)
```

### 4. リトライ

```python
async def retry_interceptor(request, handler):
    last_error = None
    for attempt in range(3):
        try:
            return await handler(request)
        except Exception as e:
            last_error = e
    raise last_error
```

### 5. トークンリフレッシュ（HTTPヘッダー上書き）

```python
async def token_refresh_interceptor(request, handler):
    if request.server_name != "observability":
        return await handler(request)
    fresh_token = await get_fresh_token()
    new_request = request.override(headers={"Authorization": f"Bearer {fresh_token}"})
    return await handler(new_request)
```

> `request.override(headers=...)` で渡したヘッダーは、内部で既存の `connection["headers"]` にマージされる。

### 6. 出力の切り詰め

```python
from mcp.types import CallToolResult, TextContent
from langchain_mcp_adapters.interceptors import MCPToolCallRequest

MAX_OUTPUT_LENGTH = 500000

async def truncate_interceptor(request: MCPToolCallRequest, handler):
    if request.server_name != "observability":
        return await handler(request)

    result = await handler(request)

    if not isinstance(result, CallToolResult):
        return result

    truncated_content = []
    for block in result.content:
        if isinstance(block, TextContent) and len(block.text) > MAX_OUTPUT_LENGTH:
            truncated_text = (
                block.text[:MAX_OUTPUT_LENGTH]
                + f"  ... (truncated: {len(block.text)} -> {MAX_OUTPUT_LENGTH} chars)"
            )
            truncated_content.append(TextContent(type="text", text=truncated_text))
        else:
            truncated_content.append(block)

    return CallToolResult(
        content=truncated_content,
        isError=result.isError,
        structuredContent=result.structuredContent,
    )
```

### 7. LangGraph Command を返す（状態更新）

```python
from langgraph.types import Command
from langchain_core.messages import ToolMessage

async def counter_interceptor(request, handler):
    tool_runtime = request.runtime
    return Command(
        update={
            "counter": 42,
            "messages": [
                ToolMessage(content="Updated!", tool_call_id=tool_runtime.tool_call_id)
            ],
        },
        goto="__end__",
    )
```

## 複数インターセプターの合成（オニオンパターン）

```python
client = MultiServerMCPClient(
    {...},
    tool_interceptors=[interceptor_1, interceptor_2],
)
```

実行順序:

```
before_1 → before_2 → ツール実行 → after_2 → after_1
```

リストの最初が最も外側のレイヤーとなる。

### `before` / `after` とは何か

各インターセプター内の **`await handler(request)` の呼び出しを境にした、その前後のコード** を指す。

```python
async def logging_interceptor_1(request, handler):
    # -------- before_1（handler 呼び出し前のコード）--------
    execution_order.append("before_1")

    result = await handler(request)   # ← ここが境界：内側（interceptor_2）へ進む

    # -------- after_1（handler 呼び出し後のコード）--------
    execution_order.append("after_1")
    return result

async def logging_interceptor_2(request, handler):
    # -------- before_2（handler 呼び出し前のコード）--------
    execution_order.append("before_2")

    result = await handler(request)   # ← ここが境界：実際のツール実行へ進む

    # -------- after_2（handler 呼び出し後のコード）--------
    execution_order.append("after_2")
    return result
```

### 実行フローの図解

```
interceptor_1 に入る
  ├── 【before_1】 handler() 呼び出し前のコード
  │
  ├── await handler(request)  ─→  interceptor_2 に入る
  │                                 ├── 【before_2】 handler() 呼び出し前のコード
  │                                 │
  │                                 ├── await handler(request)  ─→  実際の MCP ツール実行
  │                                 │                               └── CallToolResult を返す
  │                                 │
  │                                 ├── 【after_2】 handler() 呼び出し後のコード
  │                                 └── result を返す
  │
  ├── 【after_1】 handler() 呼び出し後のコード
  └── result を返す
```

**「オニオン（玉ねぎ）」と呼ばれる理由は、外側のインターセプターが内側を完全に包み込む構造** だからである。

### 内部実装（`_build_interceptor_chain`）

```python
# langchain_mcp_adapters/tools.py
def _build_interceptor_chain(base_handler, tool_interceptors):
    handler = base_handler

    if tool_interceptors:
        for interceptor in reversed(tool_interceptors):  # ← 逆順でラップ
            current_handler = handler

            async def wrapped_handler(
                req, _interceptor=interceptor, _handler=current_handler,
            ):
                return await _interceptor(req, _handler)

            handler = wrapped_handler

    return handler
```

`reversed()` で逆順にラップすることで、リスト先頭のインターセプターが最も外側になる。

## 実装例（実際の使用コード）

```python
from mcp.types import CallToolResult, TextContent
from langchain_mcp_adapters.interceptors import MCPToolCallRequest
from langchain_mcp_adapters.client import MultiServerMCPClient

MAX_OUTPUT_LENGTH = 500000

async def truncate_interceptor(request: MCPToolCallRequest, handler):
    """observability サーバーのツール出力が長すぎる場合に切り詰めるインターセプター。"""
    if request.server_name != "observability":
        return await handler(request)

    result = await handler(request)

    if not isinstance(result, CallToolResult):
        return result

    truncated_content = []
    for block in result.content:
        if isinstance(block, TextContent) and len(block.text) > MAX_OUTPUT_LENGTH:
            truncated_text = (
                block.text[:MAX_OUTPUT_LENGTH]
                + f"  ... (truncated: {len(block.text)} -> {MAX_OUTPUT_LENGTH} chars)"
            )
            truncated_content.append(TextContent(type="text", text=truncated_text))
        else:
            truncated_content.append(block)

    return CallToolResult(
        content=truncated_content,
        isError=result.isError,
        structuredContent=result.structuredContent,
    )

_client = MultiServerMCPClient(
    {
        "observability": {                   # → request.server_name == "observability"
            "transport": "http",
            "url": "http://example.com/mcp",
            "headers": {"Authorization": "Bearer xxx"},
        },
        "aws_cli": {                         # → request.server_name == "aws_cli"
            "transport": "stdio",
            "command": "python",
            "args": ["mcp_aws_cli_tool.py"],
        },
        # ... 他のサーバー
    },
    tool_interceptors=[truncate_interceptor],
)
```

## 参考リンク

- [interceptors.py（インターフェース定義）](https://github.com/langchain-ai/langchain-mcp-adapters/blob/main/langchain_mcp_adapters/interceptors.py)
- [tools.py（インターセプターチェーンの構築・実行）](https://github.com/langchain-ai/langchain-mcp-adapters/blob/main/langchain_mcp_adapters/tools.py)
- [test_interceptors.py（公式テストコード）](https://github.com/langchain-ai/langchain-mcp-adapters/blob/main/tests/test_interceptors.py)
- [client.py（MultiServerMCPClient）](https://github.com/langchain-ai/langchain-mcp-adapters/blob/main/langchain_mcp_adapters/client.py)