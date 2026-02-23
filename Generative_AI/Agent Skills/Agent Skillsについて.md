# 仕様
- https://agentskills.io/specification
- skillのディレクトリに少なくても `SKILL.md`は必ず必要

# ディレクトリ
```shell
my-skill/
├── SKILL.md          # Required: instructions + metadata
├── scripts/          # Optional: executable code
├── references/       # Optional: documentation
└── assets/           # Optional: templates, resources
```

> [!NOTE]  
> - 上記以外のファイルやディレクトリも自由に追加可能
>   - https://github.com/anthropics/skills/tree/main/skills/theme-factory (`themes`というディレクトリもある)
>   - https://github.com/anthropics/skills/tree/main/skills/pptx (ルートディレクトリに`editing.md`などもある)

# Frontmatter
- https://agentskills.io/specification#frontmatter-required
- `name`と`description`は必須

| Field | Required | Constraint |
|-------|----------|------------|
| `name` | Yes | Max 64 characters. Lowercase letters, numbers, and hyphens only. Must not start or end with a hyphen. |
| `description` | Yes | Max 1024 characters. Non-empty. Describes what the skill does and when to use it. |
| `license` | No | License name or reference to a bundled license file. |
| `compatibility` | No | Max 500 characters. Indicates environment requirements (intended product, system packages, network access, etc.). |
| `metadata` | No | Arbitrary key-value mapping for additional metadata. |
| `allowed-tools` | No | Space-delimited list of pre-approved tools the skill may use. (Experimental) |

# Instructions
- https://agentskills.io/specification#body-content  
  > The Markdown body after the frontmatter contains the skill instructions. There are no format restrictions. Write whatever helps agents perform the task effectively.
  > Recommended sections:
  > - Step-by-step instructions
  > - Examples of inputs and outputs
  > - Common edge cases
  > Note that the agent will load this entire file once it’s decided to activate a skill. Consider splitting longer SKILL.md content into referenced files.

# Progressive Disclosure
- https://agentskills.io/specification#progressive-disclosure  
  > Skills should be structured for efficient use of context:
  > 1. **Metadata** (~100 tokens): The `name` and `description` fields are loaded at startup for all skills
  > 2. **Instructions** (< 5000 tokens recommended): The full `SKILL.md` body is loaded when the skill is activated
  > 3. **Resources** (as needed): Files (e.g. those in `scripts/`, `references/`, or `assets/`) are loaded only when required
  >
  > Keep your main `SKILL.md` under 500 lines. Move detailed reference material to separate files.
