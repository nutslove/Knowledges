# MCP（Model Context Protocol）
- https://modelcontextprotocol.io/introduction  
  ![](./image/mcp_arch_1.jpg)
  ![](./image/mcp_arch_2.jpg)

## MCP Server
- https://modelcontextprotocol.io/quickstart/server
![](./image/mcp_server_1.jpg)
> [!NOTE]  
> 2025/04/13 現在、MCP ServerはLocal Hostでしか動かせない
![](./image/mcp_server_limitation.jpg)

## MCP Client
- MCP ClientはMCP Hostの内部に存在する
- MCP ClientとMCP Serverは１対１の関係
  - １つのMCP Clientで複数のMCP Serverとやりとりすることはできない