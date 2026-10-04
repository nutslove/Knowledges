# このページの構成
- **第1部: Agent Skills（標準仕様・共通）**: 製品に依らない、Agent Skillsそのものの話。どのAgent製品で使う場合でも当てはまる
- **第2部: Claude Code / Claude Agent SDK固有**: Claude CodeとClaude Agent SDKだけに当てはまる話（配置場所、独自のFrontmatter、SDKでの使い方など）。他のAgent製品では当てはまらない

---

# 第1部: Agent Skills（標準仕様・共通）

## Agent Skillsとは
- Agent Skillsは、AI Agentに**特定のタスクのやり方（手順・知識・スクリプト）をパッケージして渡す**ためのオープンなフォーマット。元々Anthropicが開発し、オープン標準として公開された
- 実体は **`SKILL.md`を含むただのフォルダ**。Agentは起動時に各Skillの`name`と`description`だけを読み、タスクに合うSkillがあれば本文を読み込む（Progressive Disclosure。後述）
- プロンプト（その会話だけの指示）と違って、**一度作れば毎回書かなくてよい**。使われる時にだけ本文がContextに入るので、たくさん置いてもContextをほとんど消費しない
- 用途の例: 社内の作業手順、レビューの観点、特定フォーマットのドキュメント生成（pptx、xlsxなど）、特定のツール・APIの使い方
- 今は、Claude Code、Claude.ai、OpenAI Codex、GitHub Copilot、VS Code、Cursor、Gemini CLI、OpenCode、Kiroなど、多くのAgent製品が対応している（一覧: https://agentskills.io/clients）
  - そのため、**一度作ったSkillを複数のAgent製品で使い回せる**。ただし、製品独自のFrontmatterなどは他の製品では効かない

### 最小のSkill
`commit-message/SKILL.md`
```markdown
---
name: commit-message
description: Generates commit messages in Conventional Commits format by analyzing git diffs. Use when the user asks for help writing a commit message or reviewing staged changes.
---

# Commit message

1. `git diff --staged`で変更内容を確認する
2. 1行目は`type(scope): 要約`の形式で書く（typeは feat / fix / docs / chore など）
3. 必要なら空行を空けて、理由を箇条書きで補足する
```

> [!NOTE]  
> - このフォルダを**どこに置くか**は、Agent製品ごとに違う。Claude Codeでの置き場所と試し方は、第2部を参照。

## 仕様
- https://agentskills.io/specification
- skillのディレクトリに少なくても **`SKILL.md`は必ず必要**

## ディレクトリ
```shell
my-skill/
├── SKILL.md          # Required: instructions + metadata
├── scripts/          # Optional: executable code
├── references/       # Optional: documentation
└── assets/           # Optional: templates, resources
```

> [!NOTE]  
> - 仕様が挙げているディレクトリは`scripts/`、`references/`、`assets/`の3つだけ（仕様上も、これらは「推奨される整理の仕方」という位置づけ）
> - 上記以外のファイルやディレクトリも自由に追加可能
>   - https://github.com/anthropics/skills/tree/main/skills/theme-factory (`themes`というディレクトリもある)
>   - https://github.com/anthropics/skills/tree/main/skills/pptx (ルートディレクトリに`editing.md`などもある)

## Frontmatter
- https://agentskills.io/specification#frontmatter-required
- `name`と`description`は必須

| Field | Required | Constraint |
|-------|----------|------------|
| `name` | Yes | Max 64 characters. Lowercase letters, numbers, and hyphens only. Must not start or end with a hyphen. Must not contain consecutive hyphens (`--`). Must match the parent directory name. |
| `description` | Yes | Max 1024 characters. Non-empty. Describes what the skill does and when to use it. |
| `license` | No | License name or reference to a bundled license file. |
| `compatibility` | No | Max 500 characters. Indicates environment requirements (intended product, system packages, network access, etc.). |
| `metadata` | No | Arbitrary key-value mapping for additional metadata. |
| `allowed-tools` | No | Space-separated string of pre-approved tools the skill may use. e.g. `Bash(git:*) Bash(jq:*) Read` (Experimental) |

> [!IMPORTANT]  
> - 最初は`name`と`description`だけがAgentのContext Windowに入って、どのSkillを使うかは`description`だけで判断されるため、`description`は特に重要。できるだけ具体的に、どんな時にどのタスクに使うべきか、具体的に記述すること。

> [!NOTE]  
> - `allowed-tools`は**Experimental**で、対応状況はAgent製品によって異なる。
> - 製品によっては、上記以外の独自のFrontmatterを追加で解釈する（Claude Codeの例は第2部を参照）。**他の製品でも使い回すSkillでは、仕様のFieldだけにしておくのが安全**。

## Instructions
- `SKILL.md`の、**Frontmatterより下の部分**（Markdown本文）のこと。Agentがタスクを実行する時に読む「手順書」にあたる
- 書き方に決まりはなく、Markdownで自由に書ける。ただし、次の3つを入れるのが推奨されている
  - 手順（Step-by-step）
  - 入出力の例
  - よくあるエッジケース
- Skillが選ばれると、この本文が**丸ごと**Contextに入る。そのため、長くなったら`references/`に分ける（後述）
- 仕様の記述:
  - https://agentskills.io/specification#body-content  
  > The Markdown body after the frontmatter contains the skill instructions. There are no format restrictions. Write whatever helps agents perform the task effectively.
  > Recommended sections:
  > - Step-by-step instructions
  > - Examples of inputs and outputs
  > - Common edge cases
  > Note that the agent will load this entire file once it’s decided to activate a skill. Consider splitting longer SKILL.md content into referenced files.

## Progressive Disclosure
- 日本語では「段階的な開示（読み込み）」。要するに、**必要になるまで読み込まない**仕組み
- 本に例えると、目次 → 本文 → 付録の順に、必要な所だけ開くイメージ

| 段階 | 読み込まれるタイミング | サイズの目安 | Contextへの入り方 |
|---|---|---|---|
| 1. Metadata | 起動時に、**全Skill分** | 1 Skillあたり約100トークン（実装ガイドでは50〜100） | `name`と`description`が入る（実装によっては`SKILL.md`の場所も） |
| 2. Instructions | そのSkillが**選ばれた時** | 5000トークン未満・500行以内（推奨） | `SKILL.md`の本文が**丸ごと**入る（Frontmatterも含めるかは実装による） |
| 3. Resources | 本文の指示で**必要になった時だけ** | 数値の目安なし。1ファイルは焦点を絞って小さくする | **読んだ範囲だけ**入る。`scripts/`は、**実行した場合は出力だけ**、読んだ場合はコードも入る |

- Resourcesは、`scripts/`、`references/`、`assets/`のファイルのこと

> [!NOTE]  
> - 段階1の「Metadata」は、Frontmatter全体ではなく、`name`と`description`を指す（仕様の記述）。Frontmatterの`metadata`フィールドとは別物。
> - Optionalのフィールドが起動時に読み込まれるかは、仕様では定められておらず、実装による。Claude Codeでは、`when_to_use`が`description`に連結されて一覧に入る（第2部を参照）。

- 例えばSkillが20個あっても、常にContextを使うのは、20個分の`name`と`description`だけ。使うSkillの本文と、その中で必要になったファイルだけが、後から追加で入る
- 仕様の記述:
  - https://agentskills.io/specification#progressive-disclosure  
  > Skills should be structured for efficient use of context:
  > 1. **Metadata** (~100 tokens): The `name` and `description` fields are loaded at startup for all skills
  > 2. **Instructions** (< 5000 tokens recommended): The full `SKILL.md` body is loaded when the skill is activated
  > 3. **Resources** (as needed): Files (e.g. those in `scripts/`, `references/`, or `assets/`) are loaded only when required
  >
  > Keep your main `SKILL.md` under 500 lines. Move detailed reference material to separate files.

### Skillの動作の流れ
- https://agentskills.io/what-are-skills

1. **Discovery**: 起動時に、全Skillの`name`と`description`だけを読み込む
2. **Activation**: Agent（モデル）が`description`を見てタスクに関連すると判断したら、`SKILL.md`全体をContextに読み込む（多くの実装は、キーワード照合ではなく、モデルの判断で選ぶ）
3. **Execution**: `SKILL.md`の指示に従い、必要に応じて`scripts/`の実行や`references/`・`assets/`の読み込みを行う

> [!NOTE]  
> - ユーザからのInputがどのSkillにもマッチしない場合、Agentは普通のLLMのモデルが学習した知識をもとに回答する。
> - １回のユーザからのinputで複数のSkillが読み込まれることもある

## `scripts/` `references/` `assets/`の違い
- どれも「`SKILL.md`に全部書くと長くなるので、別ファイルに分ける」ためのディレクトリ。**中身の種類と、Agentの使い方が違う**

| | 入れるもの | Agentの使い方 | Contextへの入り方 | 例 |
|---|---|---|---|---|
| `scripts/` | 実行できるコード | **実行する**（読んで参照することもある） | 実行した場合は、コードは入らず**出力だけ**入る | `validate.py` |
| `references/` | 詳しい説明文書 | **読む** | 読んだ時に中身が入る | `api-spec.md`、`finance.md` |
| `assets/` | テンプレート、画像、データファイル | 処理や出力に使う素材として参照する | 素材の種類による | `report-template.docx`、`schema.json` |

> [!NOTE]  
> - `assets/`の使い方とContextへの入り方は、仕様に細かい記述がない。上の表は目安として読むこと。

## `scripts/`
- 決まった処理（検証、変換、整形など）を、Agentに毎回コードを書かせる代わりに、**あらかじめ用意したコードとして実行させる**ためのディレクトリ
- 仕様の記述:
  - https://agentskills.io/specification#scripts/  
  > Contains executable code that agents can run. Scripts should:
  > - Be self-contained or clearly document dependencies
  > - Include helpful error messages
  > - Handle edge cases gracefully
  >
  > Supported languages depend on the agent implementation. Common options include Python, Bash, and JavaScript.

## `assets/`
- 出力や処理に使う**素材**（テンプレート、画像、データファイルなど）を置くディレクトリ。説明文書は`references/`、コードは`scripts/`に置き、それ以外の素材がここ
- 仕様の記述:
  - https://agentskills.io/specification#assets/  
  > Contains static resources:
  > - Templates (document templates, configuration templates)
  > - Images (diagrams, examples)
  > - Data files (lookup tables, schemas)

## `references/`
- `SKILL.md`に書くには長い、**詳しい説明文書**を置くディレクトリ。Agentは、必要になった時にだけ読む
- 分野ごとにファイルを分けておくと、関係のある分野だけ読み込めて、Contextの節約になる（例: 売上の話なら`sales.md`だけ読み、`finance.md`は読まない）
- 仕様の記述:
  - https://agentskills.io/specification#references/  
  > Contains additional documentation that agents can read when needed:
  > - `REFERENCE.md` - Detailed technical reference
  > - `FORMS.md` - Form templates or structured data formats
  > - Domain-specific files (`finance.md`, `legal.md`, etc.)
  >
  > Keep individual reference files focused. Agents load these on demand, so smaller files mean less use of context.

## 具体例（請求書を作るSkill）
ディレクトリ構成
```shell
invoice-creating/
├── SKILL.md
├── references/
│   └── tax-rules.md          # 税率や端数処理のルール（詳しい説明文書）
├── scripts/
│   └── validate.py           # 作成した請求書の検証（実行するコード）
└── assets/
    └── invoice-template.docx # 請求書のひな形（素材）
```

`SKILL.md`の本文（Frontmatterは省略）
```markdown
# Invoice creating

1. `assets/invoice-template.docx`をコピーして、請求書を作る
2. 税額の計算は、[tax-rules.md](references/tax-rules.md)のルールに従う
3. 作成後、`scripts/validate.py <作成したファイル>`を実行して検証する
   - エラーが出たら、メッセージに従って直し、再度実行する
```

- Agentの動き
  - Agentが`description`を見て関連すると判断したら、この`SKILL.md`の本文が読み込まれる
  - 手順2に来た時に初めて`tax-rules.md`が読まれる。**税の計算が不要な依頼なら、読まれない**
  - 手順3の`validate.py`は、**コードの中身は読まれず、実行結果（OKやエラーメッセージ）だけ**がContextに入る

## ファイル参照のルール
- https://agentskills.io/specification#file-references
- `SKILL.md`から他のファイルを参照する時は、**Skillのルートからの相対パス**で書く
- 参照は`SKILL.md`から**1階層まで**に留める（`SKILL.md` → `references/a.md` → `references/b.md`のように深くネストした参照チェーンは避ける）

> [!IMPORTANT]  
> - `SKILL.md`内に「どんな時にどのファイルを読むか」を書いておく。
> - 仕様は`references/`などのディレクトリに特別な意味を与えておらず、Agentが自動で読むとは定めていない。
> - 実装によっては、有効化時にファイル一覧が渡されるが、その場合も「いつ読むか」はAgentの判断になる。
> - 「詳細は`references/`を参照」ではなく、「APIが200以外を返したら`references/api-errors.md`を読む」のように、**条件とファイル名を具体的に書く**。

```markdown
See [the reference guide](references/REFERENCE.md) for details.

Run the extraction script:
scripts/extract.py
```

> [!NOTE]  
> - `scripts/`のコードは、**実行した場合は**中身がContext Windowに入らない（出力だけが入る）。そのため、同じ処理をAgentに毎回コードとして生成させるより、scriptにしておく方がトークン効率も再現性も良い。
> - 「どのファイルのどこをどれだけ読むか」は、基本的に**LLM（Agent）が決める**。参照が入れ子になっていると、`head -100`のようにファイルの一部だけ読んでしまうことがある。また、Agentを動かす側（ハーネス）が、ツールの出力を一定量（例: 10〜30K文字）で切り詰めることもある。そのため「入るのは、読んだ範囲だけ」であり、ファイルの全体が必ず入るわけではない。
> - スクリプトは、**実行前に中身を読まなくてよい**ように、用途と使い方を`SKILL.md`に書いておく（例: 「`scripts/validate.sh` — 設定ファイルを検証する」）。さらに、スクリプトに`--help`（使い方、オプション、例）を用意しておくと、Agentは`--help`を実行するだけで使い方が分かる（Contextに入るのは`--help`の出力だけ。簡潔に書くこと）。`SKILL.md`や`--help`の説明が足りないと、Agentがスクリプトの中身を読んでしまい、コードもContextに入る。
> - 「実行する」のか「読んで参照する」のかは、`SKILL.md`に書き分けておく（例: 「`analyze_form.py`を実行してフィールドを抽出する」は実行、「抽出のアルゴリズムは`analyze_form.py`を参照」は読む）。読んだ場合は、コードもContextに入る。

## 検証 (Validation)
- https://agentskills.io/specification#validation
- 参照実装の[skills-ref](https://github.com/agentskills/agentskills/tree/main/skills-ref)で、Frontmatterや命名規則が仕様に沿っているかを検証できる
```bash
skills-ref validate ./my-skill
```

## Skill vs 常時読み込まれる指示ファイル（CLAUDE.md、AGENTS.mdなど）
| | 常時読み込まれる指示ファイル | Skill |
|---|---|---|
| 読み込み | **常に**Contextに入る | `description`だけ常駐。本文は使われる時だけ |
| 向いているもの | 事実・規約（「このRepoはpnpmを使う」など） | 手順・チェックリスト・長い参考資料 |

## 配置場所
- 仕様では、Skillフォルダを**どこに置くかは定められていない**。Agent製品ごとに異なる
- 対応製品と、各製品のドキュメント（置き場所の説明を含む）へのリンクは、https://agentskills.io/clients にまとまっている
- Claude Code / Claude Agent SDKの場合は、第2部を参照

## 設計のベストプラクティス
- **`description`に「何をするか」と「いつ使うか」の両方を書く**。ユーザが実際に使いそうなキーワードを含める
  - 👎 Bad: `Helps with PDFs.`
  - 👍 Good: `Extracts text and tables from PDF files, fills PDF forms, and merges multiple PDFs. Use when working with PDF documents or when the user mentions PDFs, forms, or document extraction.`
- **`description`は三人称で書く**（`Processes Excel files and generates reports`）。`I can help you...`や`You can use this to...`は避ける。`description`はシステムプロンプトに注入されるので、視点が混ざると発動に悪影響が出る
- **`name`は動名詞形（verb + -ing）にすると何をするSkillか分かりやすい**（`processing-pdfs`、`analyzing-spreadsheets`、`testing-code`）。`pdf-processing`のような名詞句や、`process-pdfs`のような動詞形でも可。`helper`、`utils`、`tools`のような曖昧な名前や、`documents`、`data`のような広すぎる名前は避ける。Skillが複数ある場合は命名パターンを統一する
- 1つのSkillは**1つの責務**に絞る。責務が絞られていると、`description`を具体的に書きやすい
- `SKILL.md`は500行以内・5,000トークン以内を目安にし、詳細は`references/`に分割する
- 決まった処理（変換、検証、整形など）は文章で説明せず`scripts/`に置いて実行させる。エラーメッセージは、Agentが自力で直せるくらい具体的にする
- 手順（Step-by-step）、入出力の例、よくあるエッジケースを書く
- 「モデルが既に知っていること」は書かない。**モデルが知らない、このプロジェクト・組織固有の知識**に絞る
- **指示の自由度は、作業の壊れやすさに合わせる**
  - 高い自由度（文章での指示）: 複数のやり方があり、状況で判断が変わる作業（コードレビューなど）
  - 中程度（パラメータ付きのテンプレート・疑似コード）: 推奨パターンはあるが、多少の違いは許容できる作業
  - 低い自由度（決まったscriptをそのまま実行）: 壊れやすく、手順の順序や一貫性が重要な作業（DBマイグレーションなど）
- **100行を超える`references/`のファイルは、先頭に目次を付ける**。Agentがファイルの一部だけ読んだ場合でも、全体の構成が分かるようにするため
- **ファイルパスは常にスラッシュ区切りにする**（`scripts/helper.py`）。`scripts\helper.py`のようなバックスラッシュは、Unix系の環境でエラーになる
- **時期に依存する情報は本文に書かない**（「2025年8月より前はv1、後はv2」など）。現在の方法を本文に書き、古い方法は「Old patterns」のような別の節に分ける
- **選択肢を並べすぎない**。「pypdfでも、pdfplumberでも、PyMuPDFでも…」ではなく、デフォルトを1つ示し、例外の時だけ別の方法を示す
- 用語は統一する（「API endpoint」と「URL」と「path」を混在させない）
- 出力形式が重要な場合は、テンプレートや入出力の例を載せる。説明文よりも、例のほうが求める形式とレベル感が伝わる
- 重要な指示は`SKILL.md`の先頭に書く

## トラブルシュート
| 症状 | 対処 |
|---|---|
| Skillが発動しない | `description`に、ユーザが実際に使いそうなキーワードを足す / `description`が「何をするか」と「いつ使うか」の両方を含んでいるか確認する |
| 発動しすぎる | `description`をより具体的にする |
| 途中から指示を守らない | `SKILL.md`の先頭に重要な指示を置く / 指示を短く保つ |
| Skillがロードされない | FrontmatterがYAMLとして正しいか確認する。仕様上、ディレクトリ名と`name`は一致している必要がある |

> [!NOTE]  
> - 具体的な確認方法（`--debug`、`claude plugin validate`など）はAgent製品ごとに異なる。Claude Codeの場合は第2部を参照。

## 評価・メンテナンス
- **ベースライン比較**: 現実的なプロンプトを、Skillあり/なしの新しいセッションで実行して差を見る
- **評価を先に作る**: ドキュメントを書き込む前に、「Skillがない状態でAgentがどこで失敗するか」を確認し、そのギャップを埋める最小限の内容から書き始める。想像上の要件を先回りして書かない
- 複数のシナリオを用意し、**使う予定のモデルすべてで**試す（モデルによって、必要な説明の量が変わる）
- 実際の利用で、Agentが想定通りの順番でファイルを読むか、参照すべきファイルを見落とさないか、使われないファイルがないかを観察し、構成を直す

## 共有・配布
- Skillはただのフォルダなので、**Gitで管理して共有する**のが基本
- 公開されているSkillの例
  - https://github.com/anthropics/skills （Anthropic公式。`pptx`、`docx`、`pdf`、`xlsx`、`skill-creator`など）
- 配布の仕組み（Pluginなど）は、Agent製品ごとに異なる。Claude Codeの場合は第2部を参照

## セキュリティ上の注意
> [!WARNING]  
> - Skillには`scripts/`として**任意のコードを含められる**ので、出所の不明なSkillは、使う前に`SKILL.md`と`scripts/`の中身を必ず確認すること（プロンプトインジェクションやデータ流出の経路になり得る）。
> - 外部のURLからデータを取得するSkillは特にリスクが高い。取得した内容に悪意のある指示が含まれる可能性があり、信頼できるSkillでも、依存先が後から変わる可能性がある。

---

# 第2部: Claude Code / Claude Agent SDK固有
> [!NOTE]  
> - この部は、Claude CodeとClaude Agent SDKだけに当てはまる。他のAgent製品では当てはまらない。
> - Claude Agent SDKは、内部でClaude Codeを動かしているので、SKILL.mdの書き方（Frontmatter、引数、動的コンテキスト注入など）は**Claude Codeと同じ**。SDK固有の話は、最後の「Claude Agent SDKでの使い方」にまとめている。
> - 公式ドキュメント: https://code.claude.com/docs/en/skills （Claude Code）、https://code.claude.com/docs/en/agent-sdk/skills （Agent SDK）

## Claude Codeでの配置場所
- https://code.claude.com/docs/en/skills

| 場所 | パス | スコープ |
|---|---|---|
| Enterprise | managed settingsディレクトリ内の`.claude/skills/<name>/SKILL.md` | 組織の全ユーザ |
| Personal | `~/.claude/skills/<name>/SKILL.md` | 自分の全プロジェクト |
| Project | `.claude/skills/<name>/SKILL.md` | そのRepoのみ（Gitで共有できる） |
| Nested | `<サブディレクトリ>/.claude/skills/...` | そのディレクトリ配下のファイルを触る時に読み込まれる（monorepo向け） |
| Plugin | `<plugin>/skills/<name>/SKILL.md` | `/plugin:skill`のように名前空間が付くので衝突しない |

- 同名のSkillがある場合の優先順位は **Enterprise > Personal > Project**
- Skillのディレクトリは**シンボリックリンク**でもよい
- `~/.claude/skills/`や`.claude/skills/`配下の編集は**セッション中に自動で反映**される（最上位の`skills/`ディレクトリが起動時に存在しなかった場合のみ`/reload-skills`が必要）
- `/skills`で一覧表示・有効/無効の切り替えができる

### 最初のSkillを作って試す手順
1. `mkdir -p ~/.claude/skills/commit-message`でディレクトリを作る（**ディレクトリ名と`name`は揃える**）
2. 第1部の「最小のSkill」の内容を、`SKILL.md`として保存する
3. Claude Codeで「コミットメッセージを考えて」のように頼む。`description`にマッチすれば自動で使われる。`/commit-message`で直接呼び出すこともできる
4. 使われない場合は、`description`に実際に使いそうなキーワードを足す（後述の「Claude Code向けトラブルシュート」参照）

## Skill vs スラッシュコマンド
- Claude Codeでは、カスタムスラッシュコマンドはSkillに統合されている。`.claude/commands/deploy.md`と`.claude/skills/deploy/SKILL.md`はどちらも`/deploy`になる（同名の場合はSkillが優先）
- Skillは、コマンドファイルと違って、補助ファイル・呼び出し制御・自動ロードが使える

## Claude CodeのFrontmatter
- 仕様（agentskills.io）にはないが、Claude Codeが追加で解釈するフィールド。https://code.claude.com/docs/en/skills#frontmatter-fields

> [!NOTE]  
> - **Claude Codeでは全フィールドが任意**（仕様では`name`と`description`が必須）。`name`を省略するとディレクトリ名が使われ、`description`を省略すると本文の最初の行が使われる。ただし`description`は書くことが推奨されている。
> - `description`は、仕様では1024文字までだが、Claude Codeは`when_to_use`と合わせて1,536文字まで許容する。**他のAgent製品でも使い回すつもりなら、`description`は1024文字以内にしておく**こと。
> - Frontmatterは**1行目から**始まる必要がある。YAMLが壊れていると、メタデータなしで読み込まれる。
> - 未知のフィールドは無視される。

| Field | 内容 |
|---|---|
| `when_to_use` | 発動条件の補足。`description`に連結される（両方合わせて1,536文字まで） |
| `argument-hint` | `/`補完時に出すヒント。例: `[issue-number]` |
| `arguments` | `$name`で参照する名前付き引数の宣言 |
| `disable-model-invocation` | `true`で**ユーザだけ**が呼べる（Claudeは自動で使わない）。`description`もContextから消える |
| `user-invocable` | `false`で`/`メニューから隠し、**Claudeだけ**が使える |
| `allowed-tools` | そのSkillを呼んだターンの間、許可確認を省略するツール（次のユーザのメッセージで切れる）。**許可確認を省くだけで、他のツールを使えなくする制限ではない**。`Bash(git:*)`（仕様の書式）や`Bash(git diff *)`（Claude Codeの書式）のようにパターン指定も可 |
| `disallowed-tools` | Skillがアクティブな間、使えるツールから外す |
| `model` / `effort` | そのSkill実行中のモデル・推論effortを上書き |
| `context: fork` | Skillを**サブエージェント**で実行する |
| `agent` | `context: fork`の時のサブエージェント種別（`Explore`、`Plan`、`general-purpose`、カスタムAgent） |
| `hooks` | Skill呼び出し時に登録されるHooks |
| `paths` | globで指定したファイルを触る時だけ自動発動させる |

> [!WARNING]  
> - claude.aiへのアップロードやSkills APIでは、使えるFrontmatterは仕様上の`name` / `description` / `license` / `compatibility` / `metadata` / `allowed-tools`のみ。**それ以外を含めるとエラーになる**。
> - Claude Code独自のフィールドを使ったSkillは、Claude Code専用と考えること。

### 呼び出し制御（誰が起動できるか）
| Frontmatter | ユーザ（`/name`） | Claude（自動） | Contextへの影響 |
|---|---|---|---|
| デフォルト | ○ | ○ | `description`は常駐、本文は呼び出し時 |
| `disable-model-invocation: true` | ○ | × | `description`も入らない |
| `user-invocable: false` | × | ○ | `description`は常駐 |

- deploy、commit、メール送信など**副作用のあるワークフロー**は`disable-model-invocation: true`にして、Claudeが勝手に実行しないようにする

## 引数と変数
- https://code.claude.com/docs/en/skills#arguments-and-substitutions

| 記法 | 内容 |
|---|---|
| `$ARGUMENTS` | `/skill-name`の後ろに渡された引数全体。本文に`$ARGUMENTS`がない場合は末尾に`ARGUMENTS: <値>`として付く |
| `$0`, `$1`, `$ARGUMENTS[N]` | 0始まりの位置引数 |
| `$name` | `arguments`で宣言した名前付き引数 |
| `${CLAUDE_SKILL_DIR}` | そのSkillのディレクトリの絶対パス（`scripts/`を呼ぶ時に便利。カレントディレクトリに依存しない） |
| `${CLAUDE_SESSION_ID}` | セッションID |
| `${CLAUDE_PROJECT_DIR}` | プロジェクトのルート |

```yaml
---
name: migrate-component
description: Migrate a component between frameworks. Use when ...
---
Migrate the $0 component from $1 to $2.
```

## 動的コンテキスト注入 (`` !`command` ``)
- 本文中の `` !`command` `` は、**Claudeが読む前に**シェルで実行され、出力で置換される（Claudeに実行させるのではなく、事前の前処理）
- 複数行の場合は` ```! `のfenced blockを使う
```markdown
---
description: Summarizes uncommitted changes and flags anything risky. Use when the user asks what changed.
allowed-tools: Bash(git diff *)
---
## Current changes
!`git diff HEAD`

## Instructions
Summarize the changes above in two or three bullets, then list risks.
```

> [!WARNING]  
> - コマンドが**失敗すると、そのSkillの呼び出し全体が中止**される（grep系のexit code 1は許容。それ以外は`|| true`を付ける）
> - コマンドは権限ルールの対象。許可されていないものは中止されるので、`allowed-tools`で事前許可しておく
> - 設定の`disableSkillShellExecution: true`で、この機能自体を無効化できる

## サブエージェントでの実行 (`context: fork`)
- `context: fork`にすると、Skillの本文が**サブエージェントのプロンプト**として分離された環境で実行される
- サブエージェントは**会話履歴を見られない**ので、Skill本文に具体的なタスクを明記する必要がある
- 大量のファイルを調べる調査系など、**途中経過でメインのContextを汚したくない**場合に有効
```yaml
---
name: deep-research
description: Research a topic thoroughly across the codebase.
context: fork
agent: Explore
---
Research $ARGUMENTS thoroughly...
```

## Contextでの扱い（ライフサイクル）
- 呼び出されたSkillの本文は、**会話に1回だけ入り、以降はそのまま残る**（毎ターン読み直されない）
- Auto-compaction（会話の自動要約）後は、直近に使ったSkillの**先頭5,000トークン**が再添付される（全Skill合計25,000トークンまで）
  - → **重要な指示は`SKILL.md`の先頭に書く**
- 「Skillを途中から守らなくなった」場合は、再度`/skill-name`で呼び直す。**絶対に守らせたいルール**はSkillではなくHooksで強制する

## 権限・制限
- 権限ルール`Skill`で全Skillを拒否、`Skill(name)`や`Skill(name *)`で個別に許可/拒否できる
- 設定の`skillOverrides`で、Skillごとに`on` / `name-only` / `user-invocable-only` / `off`を指定できる
- Skillの`description`は、Contextの一定割合（Context windowの約1%）の文字数予算の中にまとめて載る。Skillが多すぎると超過分の`description`が切り捨てられるので、**不要なSkillは無効化する**

## Claude Code向けトラブルシュート
| 症状 | 対処 |
|---|---|
| Skillが発動しない | 「What skills are available?」と聞いて一覧に出るか確認 / `/skill-name`で直接呼ぶ / `--debug`でYAMLのパースエラーを確認 / `claude plugin validate <dir>`で検証 |
| 発動しすぎる | `disable-model-invocation: true`にする |
| 途中から指示を守らない | 再度`/skill-name`で呼び出す / 必須のルールはHooksにする（Skillのfrontmatterの`hooks`でもよい） |
| Skillがロードされない | Frontmatterが**1行目から**始まっているか確認（YAMLが壊れているとメタデータなしで読み込まれる） |

## Claude Code向けの評価・メンテナンスツール
- **skill-creator**: Skillの作成・評価（evals）・ベンチマーク・`description`の最適化をしてくれるSkill。`/plugin install skill-creator@claude-plugins-official`
- **`/skill-doctor`**: 各SkillのContext消費量と使用状況をレポートし、使われていないSkillを見つけられる

## Claude Codeでの共有・配布
- **Project**: `.claude/skills/`をGitにコミットしてチームで共有
- **Plugin**: Pluginの`skills/`ディレクトリにまとめて配布（`/plugin:skill`のように名前空間が付く）
- **Enterprise**: managed settingsで組織全体に配信
- **Claude.ai / API**: ZIPにしてアップロード（上記の通り、使えるFrontmatterが制限される）

## Claude Agent SDKでの使い方
- https://code.claude.com/docs/en/agent-sdk/skills
- SDKでも、Skillは**ファイルとして作成**する（`.claude/skills/<name>/SKILL.md`）。`agents`オプションでプログラムから定義できるサブエージェントと違い、**Skillをプログラムから登録するAPIはない**
- Skillの`name`と`description`は起動時に検出され、Claudeが呼び出した時に本文が読み込まれる。`/<name>`をプロンプトに含めて、直接呼び出すこともできる

### Skillの読み込み元（`settingSources` / `setting_sources`）
- SDKは、`settingSources`（TypeScript）/ `setting_sources`（Python）に含まれるソースから、Skillを読み込む
- **デフォルトの`query()`オプションでは、`user`と`project`を読み込む**。そのため、`~/.claude/skills/`、`<cwd>/.claude/skills/`、`cwd`から**リポジトリのルートまでの親ディレクトリ**の`.claude/skills/`が対象になる
- `additionalDirectories`（TypeScript）/ `add_dirs`（Python）で渡したディレクトリの`.claude/skills/`も、`project`ソースで読まれる
- **`settingSources`を明示的に指定する場合は、`'user'`と`'project'`を含めること**。含めないと、Skillは読み込まれない（`settingSources: []`ではSkillが読み込まれない）
- 特定のパスからSkillを読み込みたい場合は、`plugins`オプションを使う

### 使えるSkillの絞り込み（`skills`オプション）
| 指定 | 動作 |
|---|---|
| 省略 | 検出された全Skillが有効。Skillツールも使える（CLIと同じ動作） |
| `"all"` | 検出された全Skillを呼び出せる |
| `["pdf", "docx"]` | 指定した名前のSkillだけ呼び出せる（`name`フィールドかディレクトリ名。Pluginのskillは`plugin:skill`） |
| `[]` | どのSkillも呼び出せない |

- `skills`を指定すると、**SDKが`Skill`ツールを`allowedTools`に自動で追加する**。`tools`リストを明示的に渡す場合は、そこに`"Skill"`を含めること
- リストにないSkillは、**モデルから見えず、Skillツールも拒否する**。ただしSkillのファイル自体はディスク上に残っているので、ReadやBashでは到達できる
- リストには**完全一致の名前だけ**を指定できる。`"docs:*"`のようなワイルドカード、空文字、前後に空白があるもの、括弧やカンマを含むものは、`query()`がセッション開始前にエラーにする
- `/<name>`での直接呼び出しは、`skills`オプションの影響を受けない（リストに含めなくても動く）

```python
# Python
import os
from claude_agent_sdk import query, ClaudeAgentOptions

options = ClaudeAgentOptions(
    cwd=os.getcwd(),                      # .claude/skills/ はここか親ディレクトリにあること
    setting_sources=["user", "project"],  # Skillをファイルシステムから読み込む
    skills="all",                         # 検出された全Skillを呼び出せる
    allowed_tools=["Read", "Write", "Bash"],
)

async for message in query(prompt="Help me process this PDF document", options=options):
    print(message)
```

```typescript
// TypeScript
import { query } from "@anthropic-ai/claude-agent-sdk";

for await (const message of query({
  prompt: "Help me process this PDF document",
  options: {
    cwd: process.cwd(),                    // .claude/skills/ はここか親ディレクトリにあること
    settingSources: ["user", "project"],   // Skillをファイルシステムから読み込む
    skills: "all",                         // 検出された全Skillを呼び出せる
    allowedTools: ["Read", "Write", "Bash"]
  }
})) {
  console.log(message);
}
```

### 読み込まれたSkillの確認
- ストリームの先頭に、subtypeが`init`の`system`メッセージが来る。その`skills`配列で、Skillが読み込まれたか確認できる
- `skills`配列に載るのは、**ユーザが呼び出せる**（`user-invocable: false`でない）Skillと、Claude Codeにバンドルされたスキル。`user-invocable: false`のSkillは、読み込まれてClaudeからは使えるが、この配列には載らない
- `init`メッセージの`slash_commands`には、`/<name>`で呼び出せるコマンド（組み込みコマンド、バンドルされたSkill、自分のSkill、`.claude/commands/`のファイル）が載る

### ツールの事前許可
- SDKでは、Skillの`allowed-tools`（frontmatter）か、クエリの`allowedTools`（Pythonでは`allowed_tools`）オプションで、ツールを事前許可できる
- どちらも**許可確認を省くだけで、他のツールを使えなくする制限ではない**
- 組織が`allowManagedPermissionRulesOnly`をmanaged settingsで設定している場合は、どちらも無視される

### SDKでのトラブルシュート
| 症状 | 確認すること |
|---|---|
| Skillが見つからない | `settingSources` / `setting_sources`に`user`と`project`が含まれているか / `cwd`が`.claude/skills/`のあるディレクトリか、その配下になっているか（同じリポジトリ内に限る） / `ls .claude/skills/*/SKILL.md`と`ls ~/.claude/skills/*/SKILL.md`でファイルがあるか |
| Skillが使われない | `skills`リストに、そのSkillの名前が入っているか（リストにない場合、Skillツールは`Skill <name> is not in this session's skills allowlist`を返す） / `description`が具体的でキーワードを含んでいるか |
| `Invalid skill name`エラー | `skills`リストの名前が、完全一致の名前になっているか（ワイルドカードや空文字は不可） |

- YAMLの構文エラーなど、SDKに限らない一般的なSkillのトラブルは、上の「Claude Code向けトラブルシュート」を参照

### Skillsのscriptsと、MCPの使い分け
- 複数のSkillで同じ処理を使いたい時、各Skillの`scripts/`にコードを置く代わりに、**MCPサーバとして切り出し**、Skillからそのツールを呼ぶ方法がある
- 「常にMCPのほうがよい」わけではない。**共通の処理が、複数のSkillやAgentから呼ばれ、認証や状態を持つようなら、MCPサーバが向いている**。単純な処理なら、`scripts/`のほうが手軽
- 公式ドキュメントに、「`scripts/`とMCPのどちらを使うべきか」を直接比べた記述は見つけられなかった。下の表は、各機能の公式な仕様から整理した違い

| | `scripts/` | MCPサーバ |
|---|---|---|
| Contextへの入り方 | コードは入らず、**実行結果の出力だけ**入る | **ツール定義**と結果が入る（SDKでは、ツール検索がデフォルトで有効で、定義がContextを消費する量を抑える） |
| 共有 | 各Skillに複製する。Pluginにまとめれば、`${CLAUDE_PLUGIN_ROOT}`で、Plugin内のSkill間でファイルを共有できる | 1か所で管理できる。Skill以外のAgentやアプリからも使える |
| 運用 | 不要（フォルダを置くだけ） | サーバの運用（認証、バージョン管理など）が必要 |
| 失敗の仕方 | スクリプトのエラー | 接続の失敗もある（`init`メッセージのstatusが`failed`や`needs-auth`） |
| 許可 | `Bash`の許可が必要 | `allowedTools`に`mcp__<server>__<tool>`の許可が必要 |

> [!NOTE]  
> - MCPサーバは、**外部でホスティングしなくてもよい**。SDKの`createSdkMcpServer()`で、アプリ内（同じプロセス）にMCPサーバを持てる。他にも、stdio（ローカルプロセス）やHTTP（リモート）で接続できる。
> - SDKでMCPツールを使うには、`mcpServers`で接続を設定し、`allowedTools`で`mcp__<server>__<tool>`（`mcp__<server>__*`でそのサーバの全ツール）を許可する必要がある。許可がないと、ツールは見えても呼べない。
> - SDKは、OAuthの対話的な認証をしない。認証が必要なリモートのMCPサーバには、アプリ側でトークンを取得して、`headers`で渡す。
> - ツールの出力が25,000トークンを超えると、ファイルに保存され、エラーメッセージにファイルのパスが入る。
> - Skillの`allowed-tools`でも、MCPツールは`mcp__<server>__<tool>`の形で指定する（Claude CodeのMCPドキュメントの、プラグインのMCPサーバの記述に基づく。通常のサーバでも同じ形かは、明示的な記述を確認できていない）。
