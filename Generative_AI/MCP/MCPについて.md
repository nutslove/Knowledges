# MCP（Model Context Protocol）
- https://modelcontextprotocol.io/introduction  
  ![](./image/mcp_arch_1.jpg)
  ![](./image/mcp_arch_2.jpg)
- App/AgentはMCP ClientとServer両方になれる

## MCP Server
- https://modelcontextprotocol.io/quickstart/server
![](./image/mcp_server_1.jpg)
- __*SSE（Server Sent Events）* を通じてMCP Serverを遠隔起動させる（Run remotely via SSE）こともできる__
- Dockerコンテナとして動かすこともできる

## MCP Client
- MCP ClientはMCP Hostの内部に存在する
- MCP ClientとMCP Serverは１対１の関係
  - １つのMCP Clientで複数のMCP Serverとやりとりすることはできない