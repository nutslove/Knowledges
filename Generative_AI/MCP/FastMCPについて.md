## FastMCPとは
- https://gofastmcp.com/getting-started/welcome
- https://github.com/jlowin/fastmcp
- MCP serversとclientsを実装するためのPythonフレームワーク 

## FastMCPのインストール
- まず`uv`をインストール
  - https://docs.astral.sh/uv/getting-started/installation/  
    ```shell
    curl -LsSf https://astral.sh/uv/install.sh | sh
    ```

> [!NOTE]  
> uvはRustで開発されたPythonのパッケージ & Project Manager  
> https://docs.astral.sh/uv/

- uvでFastMCPをインストール  
  ```shell
  uv pip install fastmcp
  ```

## FastMCPを使ったMCP Serverの開発
- https://gofastmcp.com/servers/fastmcp