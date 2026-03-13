## `"type": "markdown"` ブロックのテーブル制約
- Slack Block Kitの`"type": "markdown"`ブロックは、Markdown記法のテーブル（`| col | col |`）を内部的に`"type": "table"`ブロックに変換してレンダリングする
- Slack APIの仕様として、**1メッセージにつきテーブルブロックは1つまで**という制約がある
  - 参考: https://docs.slack.dev/reference/block-kit/blocks/table-block/
- `final_response`などのテキスト内にMarkdownテーブルが2つ以上含まれた状態で`chat_postMessage`を呼び出すと以下のエラーが発生する
```
  {'ok': False, 'error': 'invalid_blocks', 'errors': ['only_one_table_allowed'], 'response_metadata': {'messages': ['[ERROR] only_one_table_allowed']}}
```
- ブロック定義側にテーブルを明示的に書いていなくても（`"type": "markdown"`でも）、テキスト内容に複数のMarkdownテーブルがあれば発生する点に注意