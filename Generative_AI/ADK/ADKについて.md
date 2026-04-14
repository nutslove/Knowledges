# ADK のディレクトリ構成について
- ADK は特定のディレクトリ構造を前提としており、この構造は単なる慣習ではなく、ADKが正しく機能するための要件である。

```shell
my_project/
└── my_agent/              ← パッケージ名がエージェント名になる
    ├── __init__.py        ← `from . import agent` が必須
    ├── agent.py           ← root_agent を定義
    └── .env               ← APIキー等
```

- `__init__.py` に agent アセットが UI デバッグ時に正しくロードされるよう `import` 文を書く
  - `from . import agent`