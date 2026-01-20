# MCP（Model Context Protocol）
- https://modelcontextprotocol.io/introduction  
  ![](./image/mcp_arch_1.jpg)
  ![](./image/mcp_arch_2.jpg)
- App/AgentはMCP ClientとServer両方になれる

## MCPの仕様
- https://modelcontextprotocol.io/specification/

## Tool Call
- MCPでは、ツールの引数は常にJSON bodyとして送信される。LLMがツールを呼び出す際、引数はJSON Schemaに基づいたJSONオブジェクトとして渡される。

## 非同期通信（async）
- MCP自体は非同期通信のみを前提として設計されている（明記されている）わけではなさそうだが、LangChain MCP AdaptersやMCP公式のPython SDKなどは非同期通信を前提としている
  - LangChain MCP Adapters
    - https://github.com/langchain-ai/langchain-mcp-adapters
    - https://langchain-ai.github.io/langgraph/agents/mcp/
  - MCP公式のPython SDK
    - https://github.com/modelcontextprotocol/python-sdk
- https://www.issoh.co.jp/tech/details/5984/

---

# MCP Server
- https://modelcontextprotocol.io/quickstart/server
![](./image/mcp_server_1.jpg)
- MCP ServerはClientと同じマシン上（ローカル）で動かすことも、リモートマシン（Clientと異なるマシン）上で動かす（Run remotely via SSE）こともできる
- ClientとServerとの通信方式はStdIOとSSEの2つのタイプがある
  - https://medium.com/@vkrishnan9074/mcp-clients-stdio-vs-sse-a53843d9aabb
- Dockerコンテナとして動かすこともできる

## stdio
- MCP Serverがローカル（Host/Clientと同じマシン上）で動いている場合のClientとServerとの通信方式
- 標準入力/出力を通じて通信

## SSE（Server Sent Events）
- MCP Serverがリモート（Host/Clientと異なるマシン上）で動いている場合のClientとServerとの通信方式
- HTTPベースの一方向通信を使用
> [!CAUTION]
> SSEは **_Streamable HTTP_** に置き換わる予定

---

# MCP Client
- MCP ClientはMCP Hostの内部に存在する
- MCP Serverと1対1の接続を行うプロトコルクライアント
  - **MCP ClientとMCP Serverは１対１の関係**
    - **１つのMCP Clientで複数のMCP Serverとやりとりすることはできない**

---

# MCP ServerとMCP Clientの通信
- https://modelcontextprotocol.io/docs/concepts/transports

## Server-Sent Events (SSE)
- Streamable HTTPに置き換わる予定

## Standard Input/Output (stdio)
- MCP Client(Host)とMCP Serverが同じサーバ上にある場合の通信方式

# `llms.txt`、`llms-full.txt`について
- https://llmstxt.org/
- langchain-ai.github.io/langgraph/llms-txt-overview/
## 概要
- A standard file which is designed to help LLMs and AI agents to better understand and process web site content.
- These files allow large language models (LLMs) and agents to access programming documentation and APIs, particularly useful within integrated development environments (IDEs).
- `llms.txt` is a website index for LLMs, providing background information, guidance, and links to detailed markdown files. IDEs like Cursor and Windsurf or apps like Claude Code/Desktop can use `llms.txt` to retrieve context for tasks.
- このファイルは一般的にWebサイトのルートディレクトリに配置されるケースが多い(らしい)
  - provides a concise summary of the site's most important content and structure in a machine readable markdown format.
  - It will improve the AI accuracy when extracting the information from the website.

## mcpdoc
- https://github.com/langchain-ai/mcpdoc