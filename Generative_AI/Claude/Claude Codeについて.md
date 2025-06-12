# Claude CodeのBest Practices
- https://www.anthropic.com/engineering/claude-code-best-practices

# Claude CodeのCHANGELOG
- https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md

# Claude Code Action
- Claude CodeとGithubリポジトリを統合して、自動コードレビュー、PR管理などができる機能
- 参考URL
  - https://github.com/anthropics/claude-code-action
  - https://docs.anthropic.com/ja/docs/claude-code/github-actions
  - https://azukiazusa.dev/blog/claude-code-action-github-integration/

# Claude Code使用時注意事項
## Bedrock利用時、`AWS_REGION`環境変数とモデルのCross Region識別子(e.g. `apac`、`us`)を一致させる必要がある
- 2つが不一致すると以下のエラーが出る  
  ```shell
  API Error (apac.anthropic.claude-sonnet-4-20250514-v1:0): 400 The provided model identifier is invalid.
  ```
- https://github.com/anthropics/claude-code/issues/1434