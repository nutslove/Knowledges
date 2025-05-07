# 手順（Windows）
- uvをインストール
  - https://docs.astral.sh/uv/getting-started/installation/

- 以下コマンドを実行
```Powershell
uv init mcp-server-demo
cd mcp-server-demo
uv add "mcp[cli]"
new-item server.py # server.pyファイルが生成される
```

- 以下コマンドでCursorを起動
```Powershell
cursor .
```

- `server.py`を修正後、以下コマンド実行
```Powershell
uv run mcp install server.py # claude_desktop_config.jsonが更新される
```
  - `server.py`の例  
    ```python
    from mcp.server.fastmcp import FastMCP

    # Create an MCP server
    mcp = FastMCP("MCP_Demo")

    @mcp.resource("data://mcp_architecture")
    def get_mcp_architecture() -> str:
        """Get the MCP architecture"""
        mcp_architecture = '''
    # Core architecture

    > Understand how MCP connects clients, servers, and LLMs

    The Model Context Protocol (MCP) is built on a flexible, extensible architecture that enables seamless communication between LLM applications and integrations. This document covers the core architectural components and concepts.

    ## Overview

    MCP follows a client-server architecture where:

    * **Hosts** are LLM applications (like Claude Desktop or IDEs) that initiate connections
    * **Clients** maintain 1:1 connections with servers, inside the host application
    * **Servers** provide context, tools, and prompts to clients

    ```mermaid
    flowchart LR
        subgraph "Host"
            client1[MCP Client]
            client2[MCP Client]
        end
        subgraph "Server Process"
            server1[MCP Server]
        end
        subgraph "Server Process"
            server2[MCP Server]
        end

        client1 <-->|Transport Layer| server1
        client2 <-->|Transport Layer| server2
    ```

    ## Core components

    ### Protocol layer

    The protocol layer handles message framing, request/response linking, and high-level communication patterns.

    <Tabs>
      <Tab title="TypeScript">
        ```typescript
        class Protocol<Request, Notification, Result> {
            // Handle incoming requests
            setRequestHandler<T>(schema: T, handler: (request: T, extra: RequestHandlerExtra) => Promise<Result>): void

            // Handle incoming notifications
            setNotificationHandler<T>(schema: T, handler: (notification: T) => Promise<void>): void

            // Send requests and await responses
            request<T>(request: Request, schema: T, options?: RequestOptions): Promise<T>

            // Send one-way notifications
            notification(notification: Notification): Promise<void>
        }
        ```
      </Tab>

      <Tab title="Python">
        ```python
        class Session(BaseSession[RequestT, NotificationT, ResultT]):
            async def send_request(
                self,
                request: RequestT,
                result_type: type[Result]
            ) -> Result:
                """Send request and wait for response. Raises McpError if response contains error."""
                # Request handling implementation

            async def send_notification(
                self,
                notification: NotificationT
            ) -> None:
                """Send one-way notification that doesn't expect response."""
                # Notification handling implementation

            async def _received_request(
                self,
                responder: RequestResponder[ReceiveRequestT, ResultT]
            ) -> None:
                """Handle incoming request from other side."""
                # Request handling implementation

            async def _received_notification(
                self,
                notification: ReceiveNotificationT
            ) -> None:
                """Handle incoming notification from other side."""
                # Notification handling implementation
        ```
      </Tab>
    </Tabs>

    Key classes include:

    * `Protocol`
    * `Client`
    * `Server`

    ### Transport layer

    The transport layer handles the actual communication between clients and servers. MCP supports multiple transport mechanisms:

    1. **Stdio transport**
       * Uses standard input/output for communication
       * Ideal for local processes

    2. **HTTP with SSE transport**
       * Uses Server-Sent Events for server-to-client messages
       * HTTP POST for client-to-server messages

    All transports use [JSON-RPC](https://www.jsonrpc.org/) 2.0 to exchange messages. See the [specification](/specification/) for detailed information about the Model Context Protocol message format.

    ### Message types

    MCP has these main types of messages:

    3. **Requests** expect a response from the other side:
       ```typescript
       interface Request {
         method: string;
         params?: { ... };
       }
       ```

    4. **Results** are successful responses to requests:
       ```typescript
       interface Result {
         [key: string]: unknown;
       }
       ```

    5. **Errors** indicate that a request failed:
       ```typescript
       interface Error {
         code: number;
         message: string;
         data?: unknown;
       }
       ```

    6. **Notifications** are one-way messages that don't expect a response:
       ```typescript
       interface Notification {
         method: string;
         params?: { ... };
       }
       ```

    ## Connection lifecycle

    ### 1. Initialization

    ```mermaid
    sequenceDiagram
        participant Client
        participant Server

        Client->>Server: initialize request
        Server->>Client: initialize response
        Client->>Server: initialized notification

        Note over Client,Server: Connection ready for use
    ```

    7. Client sends `initialize` request with protocol version and capabilities
    8. Server responds with its protocol version and capabilities
    9. Client sends `initialized` notification as acknowledgment
    10. Normal message exchange begins

    ### 2. Message exchange

    After initialization, the following patterns are supported:

    * **Request-Response**: Client or server sends requests, the other responds
    * **Notifications**: Either party sends one-way messages

    ### 3. Termination

    Either party can terminate the connection:

    * Clean shutdown via `close()`
    * Transport disconnection
    * Error conditions

    ## Error handling

    MCP defines these standard error codes:

    ```typescript
    enum ErrorCode {
      // Standard JSON-RPC error codes
      ParseError = -32700,
      InvalidRequest = -32600,
      MethodNotFound = -32601,
      InvalidParams = -32602,
      InternalError = -32603
    }
    ```

    SDKs and applications can define their own error codes above -32000.

    Errors are propagated through:

    * Error responses to requests
    * Error events on transports
    * Protocol-level error handlers

    ## Implementation example

    Here's a basic example of implementing an MCP server:

    <Tabs>
      <Tab title="TypeScript">
        ```typescript
        import { Server } from "@modelcontextprotocol/sdk/server/index.js";
        import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

        const server = new Server({
          name: "example-server",
          version: "1.0.0"
        }, {
          capabilities: {
            resources: {}
          }
        });

        // Handle requests
        server.setRequestHandler(ListResourcesRequestSchema, async () => {
          return {
            resources: [
              {
                uri: "example://resource",
                name: "Example Resource"
              }
            ]
          };
        });

        // Connect transport
        const transport = new StdioServerTransport();
        await server.connect(transport);
        ```
      </Tab>

      <Tab title="Python">
        ```python
        import asyncio
        import mcp.types as types
        from mcp.server import Server
        from mcp.server.stdio import stdio_server

        app = Server("example-server")

        @app.list_resources()
        async def list_resources() -> list[types.Resource]:
            return [
                types.Resource(
                    uri="example://resource",
                    name="Example Resource"
                )
            ]

        async def main():
            async with stdio_server() as streams:
                await app.run(
                    streams[0],
                    streams[1],
                    app.create_initialization_options()
                )

        if __name__ == "__main__":
            asyncio.run(main())
        ```
      </Tab>
    </Tabs>

    ## Best practices

    ### Transport selection

    1. **Local communication**
       * Use stdio transport for local processes
       * Efficient for same-machine communication
       * Simple process management

    2. **Remote communication**
       * Use SSE for scenarios requiring HTTP compatibility
       * Consider security implications including authentication and authorization

    ### Message handling

    3. **Request processing**
       * Validate inputs thoroughly
       * Use type-safe schemas
       * Handle errors gracefully
       * Implement timeouts

    4. **Progress reporting**
       * Use progress tokens for long operations
       * Report progress incrementally
       * Include total progress when known

    5. **Error management**
       * Use appropriate error codes
       * Include helpful error messages
       * Clean up resources on errors

    ## Security considerations

    6. **Transport security**
       * Use TLS for remote connections
       * Validate connection origins
       * Implement authentication when needed

    7. **Message validation**
       * Validate all incoming messages
       * Sanitize inputs
       * Check message size limits
       * Verify JSON-RPC format

    8. **Resource protection**
       * Implement access controls
       * Validate resource paths
       * Monitor resource usage
       * Rate limit requests

    9. **Error handling**
       * Don't leak sensitive information
       * Log security-relevant errors
       * Implement proper cleanup
       * Handle DoS scenarios

    ## Debugging and monitoring

    10. **Logging**
       * Log protocol events
       * Track message flow
       * Monitor performance
       * Record errors

    11. **Diagnostics**
       * Implement health checks
       * Monitor connection state
       * Track resource usage
       * Profile performance

    12. **Testing**
       * Test different transports
       * Verify error handling
       * Check edge cases
       * Load test servers
    '''

        return mcp_architecture


    @mcp.resource("data://mcp_specification")
    def get_mcp_specification() -> str:
        """Get the MCP specification"""
        mcp_specification = """
    # Specification

    [Model Context Protocol](https://modelcontextprotocol.io) (MCP) is an open protocol that
    enables seamless integration between LLM applications and external data sources and
    tools. Whether you're building an AI-powered IDE, enhancing a chat interface, or creating
    custom AI workflows, MCP provides a standardized way to connect LLMs with the context
    they need.

    This specification defines the authoritative protocol requirements, based on the
    TypeScript schema in
    [schema.ts](https://github.com/modelcontextprotocol/specification/blob/main/schema/2025-03-26/schema.ts).

    For implementation guides and examples, visit
    [modelcontextprotocol.io](https://modelcontextprotocol.io).

    The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD
    NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
    interpreted as described in [BCP 14](https://datatracker.ietf.org/doc/html/bcp14)
    \[[RFC2119](https://datatracker.ietf.org/doc/html/rfc2119)]
    \[[RFC8174](https://datatracker.ietf.org/doc/html/rfc8174)] when, and only when, they
    appear in all capitals, as shown here.

    ## Overview

    MCP provides a standardized way for applications to:

    * Share contextual information with language models
    * Expose tools and capabilities to AI systems
    * Build composable integrations and workflows

    The protocol uses [JSON-RPC](https://www.jsonrpc.org/) 2.0 messages to establish
    communication between:

    * **Hosts**: LLM applications that initiate connections
    * **Clients**: Connectors within the host application
    * **Servers**: Services that provide context and capabilities

    MCP takes some inspiration from the
    [Language Server Protocol](https://microsoft.github.io/language-server-protocol/), which
    standardizes how to add support for programming languages across a whole ecosystem of
    development tools. In a similar way, MCP standardizes how to integrate additional context
    and tools into the ecosystem of AI applications.

    ## Key Details

    ### Base Protocol

    * [JSON-RPC](https://www.jsonrpc.org/) message format
    * Stateful connections
    * Server and client capability negotiation

    ### Features

    Servers offer any of the following features to clients:

    * **Resources**: Context and data, for the user or the AI model to use
    * **Prompts**: Templated messages and workflows for users
    * **Tools**: Functions for the AI model to execute

    Clients may offer the following feature to servers:

    * **Sampling**: Server-initiated agentic behaviors and recursive LLM interactions

    ### Additional Utilities

    * Configuration
    * Progress tracking
    * Cancellation
    * Error reporting
    * Logging

    ## Security and Trust & Safety

    The Model Context Protocol enables powerful capabilities through arbitrary data access
    and code execution paths. With this power comes important security and trust
    considerations that all implementors must carefully address.

    ### Key Principles

    1. **User Consent and Control**

       * Users must explicitly consent to and understand all data access and operations
       * Users must retain control over what data is shared and what actions are taken
       * Implementors should provide clear UIs for reviewing and authorizing activities

    2. **Data Privacy**

       * Hosts must obtain explicit user consent before exposing user data to servers
       * Hosts must not transmit resource data elsewhere without user consent
       * User data should be protected with appropriate access controls

    3. **Tool Safety**

       * Tools represent arbitrary code execution and must be treated with appropriate
         caution.
         * In particular, descriptions of tool behavior such as annotations should be
           considered untrusted, unless obtained from a trusted server.
       * Hosts must obtain explicit user consent before invoking any tool
       * Users should understand what each tool does before authorizing its use

    4. **LLM Sampling Controls**
       * Users must explicitly approve any LLM sampling requests
       * Users should control:
         * Whether sampling occurs at all
         * The actual prompt that will be sent
         * What results the server can see
       * The protocol intentionally limits server visibility into prompts

    ### Implementation Guidelines

    While MCP itself cannot enforce these security principles at the protocol level,
    implementors **SHOULD**:

    1. Build robust consent and authorization flows into their applications
    2. Provide clear documentation of security implications
    3. Implement appropriate access controls and data protections
    4. Follow security best practices in their integrations
    5. Consider privacy implications in their feature designs

    ## Learn More

    Explore the detailed specification for each protocol component:

    <CardGroup cols={5}>
      <Card title="Architecture" icon="sitemap" href="architecture" />

      <Card title="Base Protocol" icon="code" href="basic" />

      <Card title="Server Features" icon="server" href="server" />

      <Card title="Client Features" icon="user" href="client" />

      <Card title="Contributing" icon="pencil" href="contributing" />
    </CardGroup>
    """
        return mcp_specification

    @mcp.prompt()
    def respond_in_japanese() -> str:
        return "日本語で回答してください。語尾に'やん'をつけてください。"

    # dynamic resource
    @mcp.resource("greeting://{name}")
    def get_greeting(name: str) -> str:
        """Get a personalized greeting"""
        return f"Hello, {name}!"
    ```

- `claude_desktop_config.json`に`server.py`の`FastMCP("<任意の名前>")`の「任意の名前」の設定が反映されていることを確認
  - 例えば`server.py`で`FastMCP("Demo")`としている場合、`claude_desktop_config.json`に以下のような設定が追加される  
    ```json
    {
      "mcpServers": {
        "Demo": {
          "command": "uv",
          "args": [
            "run",
            "--with",
            "mcp[cli]",
            "mcp",
            "run",
            "C:\\Users\\JoonkiLee\\mcp-server-demo\\server.py"
          ]
        }
      }
    }
    ```

- Claude Desktopを再起動
  - Claude Desktopの「＋」ボタンで追加したMCPのResourcesとPromptsが表示/選択できることを確認

## 参考URL
- https://github.com/modelcontextprotocol/python-sdk
- https://modelcontextprotocol.io/docs/concepts/resources
- https://modelcontextprotocol.io/docs/concepts/prompts