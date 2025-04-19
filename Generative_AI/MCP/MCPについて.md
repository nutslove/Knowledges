# MCP（Model Context Protocol）
- https://modelcontextprotocol.io/introduction  
  ![](./image/mcp_arch_1.jpg)
  ![](./image/mcp_arch_2.jpg)
- App/AgentはMCP ClientとServer両方になれる

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

# MCP Client
- MCP ClientはMCP Hostの内部に存在する
- MCP ClientとMCP Serverは１対１の関係
  - １つのMCP Clientで複数のMCP Serverとやりとりすることはできない