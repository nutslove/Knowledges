# Pythonプロジェクトのディレクトリ構成ベストプラクティス（uv版）

uvを使ったPythonプロジェクトの推奨構成と運用方法のまとめ。

## 目次

1. **基本構成（src layout）** — ディレクトリの基本形と、なぜ `src/` を挟むのか
2. **アプリケーション系のレイヤー構成** — 通常のWeb API / AI Agent の構成と違い
3. **`__init__.py`について** — パッケージ化のマーカーファイル
4. **pyproject.tomlの設定例** — 各テーブルに何をどう書くか
5. **必須の補助ファイル** — .gitignore / .python-version / .env.example / pre-commit
6. **テストの構成**
7. **Dockerでの利用**
8. **フロントエンドとバックエンドを1リポジトリで管理する場合（モノレポ）** — `backend/` と `frontend/` の構成、uv workspace の使いどころ
9. **各サブディレクトリの役割と実装例** — AI Agentを題材にしたコード詳解（設定だけ知りたい人は飛ばしてOK）
10. **ベストプラクティスのまとめ**

## 基本構成（src layout）

PyPA（Python Packaging Authority）が推奨する **src layout** を採用する。

```
my_project/
├── src/
│   └── my_package/
│       ├── __init__.py
│       ├── main.py
│       ├── core/
│       │   ├── __init__.py
│       │   └── logic.py
│       └── utils/
│           ├── __init__.py
│           └── helpers.py
├── tests/
│   ├── conftest.py          # tests/配下に__init__.pyは置かない（後述）
│   ├── unit/
│   │   └── test_logic.py
│   └── integration/
│       └── test_api.py
├── docs/
│   └── index.md
├── scripts/
│   └── deploy.sh
├── .gitignore
├── .python-version
├── .env.example
├── pyproject.toml
├── uv.lock
├── README.md
└── LICENSE
```

`src/`配下にパッケージを置くことで、**インストールしないままソースを import してしまう事故を防げる**。ただしそれだけでは「配布物（wheel）に必要なファイルが入っているか」までは検証されない。パッケージング設定のミスまで検出するには、**editable ではない通常インストールでテストする**必要がある（後述）。

> [!NOTE]
> **この Tree の `core/` `utils/` について**
> これは src layout の配置を示すための**簡略化した最小例**で、`core/` `utils/` はロジックの置き場をざっくり表したもの。
> - **`core/`** … そのアプリ"ならでは"の中心的なロジック（ドメインモデル・ビジネスルール・計算・判定など）。できるだけ外部依存（DB・HTTP・フレームワーク）を持たない純粋なコードにする。
> - **`utils/`** … ドメインに依存しない汎用ヘルパー（文字列整形・日付変換など）。別プロジェクトにも流用できる処理。
>
> 実際のアプリ設計では、この `core/` の役割は規模に応じて **`domain/`（純粋なモデル・ルール）＋ `services/`（ユースケース＝処理の流れ）** に分割されることが多い。後述の「通常のバックエンドWeb API」「AI Agent」の Tree に `core/` が無く `domain/` `services/` があるのはこのため。小さいうちは `core/` にまとめ、ロジックが育って「何であるか（モデル・ルール）」と「どう処理するか（ユースケース）」を分けたくなったら割り直す、という育て方でよい。

### なぜ `src/` の中にさらに同名のディレクトリを作るのか

`my_project/src/my_package/` のように外側と内側で同名になるのは偶然ではなく、**`uv init` がプロジェクト名からパッケージ名（`src/` 直下のディレクトリ名）を自動で決める**ため。ビルドバックエンドも、デフォルトではこの名前の一致を前提にしている。ただし、同名なのは「名前」だけで、**両者の役割は別物**。

- **名前の変換ルール**: `uv init --package my-app` のようにハイフンやドットを含む名前を渡しても、パッケージ名は小文字・アンダースコアに正規化される（`my-app` / `My.App` / `my_app` はいずれも `src/my_app/` になる。手元で確認）。`pyproject.toml` の `name` は `my-app` のようにハイフン区切りに正規化されるが、`import` するときは `my_app` になる。
- **一致が必要な理由**: `uv_build`（uv のデフォルト）は `src/<正規化したプロジェクト名>/__init__.py` を探す。ディレクトリを `src/renamed/` に改名してビルドすると `Expected a Python module at: src/<プロジェクト名>/__init__.py` で失敗する（手元で確認）。hatchling も同様に `src/<プロジェクト名>/` を自動検出し、一致しない場合は明示設定が必要（後述）。
- **ずらしたいとき**: プロジェクト名（配布名）と `import` 名を意図的に変えたい場合は、`uv_build` なら `[tool.uv.build-backend] module-name = "..."`、hatchling なら `[tool.hatch.build.targets.wheel] packages = [...]` で指定する。

```
my_app/                  ← プロジェクト（リポジトリ）全体の入れ物
├── pyproject.toml       #   設定・テスト・docsなどを含む
├── README.md
├── tests/
└── src/                 ← ソースコードを隔離する境界
    └── my_app/          ← 実際にimport / インストールされるパッケージ本体
        ├── __init__.py
        └── main.py
```

| ディレクトリ | 役割 |
|---|---|
| `my_app/`（外側） | プロジェクト／リポジトリ全体（`pyproject.toml`・`tests/`・`docs/`などを含む入れ物） |
| `src/` | ソースコードを隔離する境界 |
| `src/my_app/`（内側） | `import my_app` で読み込まれる実際のパッケージ本体（`__init__.py`を持つ） |

`import my_app` の対象になるのは内側の `src/my_app/` のみ。外側は単なる入れ物。

#### 内側に同名パッケージを置く理由

1. **import可能な「パッケージ」の単位だから**  
   `__init__.py` を持つ内側のディレクトリがパッケージの実体。`import my_app` で読み込まれるのはこの内側であり、外側は単なる入れ物にすぎない。

2. **`src/` を挟むことで「インストールせずに誤ってローカルのソースをimportする事故」を防げる（src layout最大の利点）**  
   `src/`を挟まないフラットレイアウト（`my_app/my_app/`）だと、プロジェクトルートで作業しているときカレントディレクトリの `my_app/` がそのままimportできてしまう（Pythonはカレントディレクトリを `sys.path` の先頭に入れるため）。すると、インストールしていなくてもテストが通ってしまい、「パッケージング設定が壊れていてインストールでは動かないのに、開発環境ではたまたま動いていた」という不整合に気づけない。`src/`を挟むと、ルートから直接importできなくなり、**インストールしていない状態ではimportに失敗する**ので、「インストールし忘れ」にはすぐ気づける。PyPA公式も、src layoutは「開発中のコードを誤って使うことを防ぐ」ものとして説明している（[src layout vs flat layout](https://packaging.python.org/en/latest/discussions/src-layout-vs-flat-layout/)）。

3. **ただし「配布物が正しい」ことまでは保証されない（editable install の落とし穴）**  
   `uv sync` や `pip install -e .` は **editable install** で、`src/` ディレクトリそのものを指す形でインストールされる。そのため、`pyproject.toml` のパッケージ指定ミスで wheel にファイルが入らなくなっていても、**editable ではテストが通ってしまう**。手元で、`secret.py` を wheel から除外する設定ミスを入れて確認した結果は次のとおり:

   | 状況 | flat layout | src layout |
   |---|---|---|
   | インストールせず `python -m pytest` | 通る（見逃す） | import エラー（気づける） |
   | editable install（`uv sync` のデフォルト）でテスト | — | **通る（見逃す）** |
   | 通常インストール（`uv sync --no-editable` / wheel）でテスト | — | **失敗（気づける）** |

   つまり、src layout が防ぐのは「インストールせずに動いてしまう」ことで、「**wheel の中身が正しいか**」を検証するには別途、**editable ではない通常インストールの状態でテストを走らせる**必要がある。CI では次のようにするとよい。

   ```bash
   uv sync --no-editable    # プロジェクトを editable ではなく通常インストールする
   uv run --no-sync pytest  # 再同期（editable への戻し）を避けてテストする
   ```

   **`--no-sync` は必須**。付けないと `uv run` が環境を再同期して **editable に戻してしまい**、テストが通ってしまう（手元で、`uv sync --no-editable` 後に `--no-sync` なしで `uv run pytest` すると通り、`direct_url.json` が `"editable": true` になっていることを確認した）。`--no-sync` を付ければ、wheel から欠落したファイルをテストで検出できる。ローカル開発は editable（高速・変更が即反映）、CI は `--no-editable` という使い分けが現実的。

4. **ルートが散らからない**  
   ルート直下にパッケージコードを置くと、`tests/`・`docs/`・設定ファイルとコードが混在する。`src/`に閉じ込めることでコードと周辺ファイルが明確に分離される。

> [!NOTE]
> シンプルなスクリプトや小さなツールであれば `src/` を挟まないフラットレイアウトでもよい。**ライブラリとして配布する・きちんとテストしたい場合**に src layout が推奨される（PyPA公式も推奨）。

### 1つのプロジェクトに複数のパッケージを入れられるか

「複数のパッケージ」が何を指すかで、方法は3通りに分かれる（いずれも手元でビルドして確認）。

| やりたいこと | 方法 | 追加設定 |
|---|---|---|
| 機能ごとにディレクトリを分けたい | **サブパッケージ**（`src/my_app/` の下に増やす） | 不要 |
| `import foo` / `import bar` のようにトップレベルを複数持ちたい | **1プロジェクトにトップレベルパッケージを複数置く** | 必要（下記） |
| パッケージごとに別々にバージョン管理・配布したい | **複数プロジェクトを uv workspace で束ねる** | 各 `pyproject.toml` が必要（モノレポ章のパターンB） |

#### ① サブパッケージ（普通のやり方）

`src/my_app/api/`、`src/my_app/services/`、`src/my_app/domain/` のように、**1つのパッケージの下にディレクトリを増やす**。この記事のレイヤー構成がこれで、設定は要らない。`my_app/services/__init__.py` があれば wheel に含まれ、`import my_app.services` で使える。

#### ② トップレベルパッケージを複数置く

`src/foo/` と `src/bar/` を並べる形。`import foo` と `import bar` が別々にできる。

```text
my_project/
└── src/
    ├── my_project/   ← プロジェクト名と同じ（デフォルトで wheel に入る）
    ├── foo/
    └── bar/
```

> [!WARNING]
> `uv_build` はデフォルトだと、**プロジェクト名と一致する1つのパッケージしか wheel に入れない**。`src/foo` や `src/bar` を足しても、ビルドは成功するのに wheel には `my_project/` しか入らず、インストールすると `import foo` が失敗する（手元で確認）。ビルドが成功するため気づきにくい。

使っているビルドバックエンドに応じて、含めるパッケージを**明示**すると、上の Warning の問題（`foo` / `bar` が wheel に入らない）は解消する。

```toml
# uv_build の場合（リストで指定。手元で3つとも wheel に入り、import できることを確認）
[tool.uv.build-backend]
module-name = ["my_project", "foo", "bar"]
```

```toml
# hatchling の場合
[tool.hatch.build.targets.wheel]
packages = ["src/foo", "src/bar"]
```

ただし、**明示すれば絶対に安全というわけではない**。手元で試した結果は次のとおり。

| 状況 | `uv_build` | hatchling |
|---|---|---|
| 設定に書いた名前が正しい | 全部 wheel に入る | 全部 wheel に入る |
| リストに**書き忘れた**パッケージがある | **ビルドは成功し、wheel には入らない** | **ビルドは成功し、wheel には入らない** |
| **存在しない**名前・パスを書いた | **ビルド失敗**（`Expected a Python module at: ...`） | **ビルドは成功し、wheel は空**（`dist-info` しか入らない） |
| あとから `src/baz/` を足したが設定に追記し忘れた | wheel に入らず、`import baz` が失敗 | 同左 |

- つまり、明示しても「**書き忘れ**」は防げず、しかもビルドは成功してしまう。hatchling は**綴りを間違えても気づけない**（`uv_build` は失敗するので気づける）。
- 対策は、前述の **`uv sync --no-editable` + `uv run --no-sync pytest`** のように、**wheel をインストールした状態で `import` するテスト**を CI に入れること。`src/foo` を足したのに設定へ追記し忘れた場合も、`import foo` の失敗として検出できる。

> [!NOTE]
> `uv_build` の `module-name` にリストを渡せることは、uv 公式の [Build backend](https://docs.astral.sh/uv/concepts/build-backend/) に記載がある。トップレベル名は、他のライブラリと衝突すると `import` が壊れるため、汎用的な名前（`utils`、`common` など）は避ける。**名前空間を分けたいだけなら、②よりも①のサブパッケージの方が素直**。

#### ③ パッケージごとに独立して管理・配布する

別々にバージョンを付けて配布したい、依存を分けたい、という場合は、1つのプロジェクトではなく**複数のプロジェクトを uv workspace で束ねる**。パッケージごとに `pyproject.toml` を持ち、`uv.lock` は1つにまとまる。詳しくは「フロントエンドとバックエンドを1リポジトリで管理する場合（モノレポ）」章のパターンBを参照。

#### どれを選ぶか

- **普通のアプリは①で十分**。
- **②は、同じ配布物にまとめたいが `import` 名は分けたい場合**（既存コードの移行途中など）に限る。
- **③は、別々にバージョンを付けて配布したい場合**。

## アプリケーション系のレイヤー構成

ライブラリではなくアプリケーションの場合、ドメインごとにレイヤーを分ける。基本は **api → services → infrastructure** の単方向依存で、`services`層が外部の実装詳細（DBドライバや外部APIクライアント）を直接知らないようにする。

ここでは「通常のバックエンドWeb API」と「AI Agent」の2パターンを示す。違いは主に **`agents/`レイヤーの有無** と **`infrastructure/`が繋ぐ相手（DB中心か、LLM・ツール中心か）**。

### 通常のバックエンドWeb API（FastAPI + DB）

ユーザー管理APIを例にした、DBを永続化層に持つ典型的な構成。

```
my_api/
├── src/
│   └── my_api/
│       ├── __init__.py
│       ├── main.py              # FastAPIエントリポイント
│       ├── config.py            # 設定（pydantic-settings）
│       ├── api/                 # ルーター層（HTTPエンドポイント）
│       │   ├── __init__.py
│       │   ├── routes.py
│       │   └── deps.py          # 依存性注入（DBセッション等）
│       ├── services/            # ビジネスロジック（ユースケース）
│       │   ├── __init__.py
│       │   └── user_service.py
│       ├── domain/              # ドメインモデル（純粋なビジネスモデル）
│       │   ├── __init__.py
│       │   └── models.py
│       ├── repositories/        # 永続化（DBアクセスの抽象化）
│       │   ├── __init__.py
│       │   └── user_repository.py
│       ├── infrastructure/      # 外部連携（DB接続、外部APIクライアント）
│       │   ├── __init__.py
│       │   └── database.py
│       └── schemas/             # Pydanticスキーマ（リクエスト/レスポンス）
│           ├── __init__.py
│           └── user_schemas.py
├── tests/
│   ├── conftest.py
│   ├── unit/
│   └── integration/
├── pyproject.toml
├── uv.lock
└── README.md
```

依存方向:

```
api  →  services  →  repositories  →  infrastructure（DB）
              ↓
           domain
```

- `api`はHTTPの受け口。リクエストを`schemas`で受け取り、`services`を呼ぶだけ
- `services`がユースケースを実装し、`repositories`経由でデータを読み書きする
- `repositories`がDBアクセスを抽象化し、SQLやORMの詳細を`infrastructure`に隠す
- データ系のアプリでは外部連携の主役がDBなので、**`repositories/`** が中心になる

### AI Agent（LangGraph + LLM）

LangGraphベースのRCA（根本原因分析）Agentを例にした構成。通常のWeb APIに対して **`agents/`レイヤーが加わり**、`repositories/`（DB）の代わりに **`infrastructure/`がLLMや観測ツールに繋ぐ** のが特徴。

```
my_app/
├── src/
│   └── my_app/
│       ├── __init__.py
│       ├── main.py              # エントリポイント（FastAPI / CLI など）
│       ├── config.py            # 設定（pydantic-settings）
│       ├── api/                 # ルーター層（HTTPで公開する場合）
│       │   ├── __init__.py
│       │   ├── routes.py
│       │   └── deps.py
│       ├── services/            # ビジネスロジック（APIとエージェントの橋渡し）
│       │   ├── __init__.py
│       │   └── rca_service.py
│       ├── domain/              # ドメインモデル（純粋なビジネスモデル）
│       │   ├── __init__.py
│       │   └── models.py
│       ├── agents/              # ★LangGraphエージェント定義（このレイヤーが追加される）
│       │   ├── __init__.py
│       │   ├── graph.py         #   グラフの組み立て
│       │   ├── nodes.py         #   各ノード（LLM呼び出し等）
│       │   ├── state.py         #   グラフのState定義
│       │   └── tools.py         #   エージェントが使うTool
│       ├── infrastructure/      # 外部連携（LLM・観測・データソース）
│       │   ├── __init__.py
│       │   ├── bedrock_client.py
│       │   ├── prometheus_client.py
│       │   └── langfuse_client.py
│       └── schemas/             # Pydanticスキーマ（API I/O + LLM構造化出力）
│           ├── __init__.py
│           ├── api_schemas.py
│           └── agent_schemas.py
├── tests/
│   ├── conftest.py
│   ├── unit/
│   └── integration/
├── Dockerfile              # プロジェクトルートに置く（後述）
├── .dockerignore
├── .env.example
├── pyproject.toml
├── uv.lock
└── README.md
```

依存方向:

```
api  →  services  →  agents  →  infrastructure（LLM・ツール）
              ↓           ↓
           domain     schemas
              ↑
          config（全レイヤーから参照可）
```

### 両者の違い

| 観点 | 通常のWeb API | AI Agent |
|---|---|---|
| `agents/`レイヤー | なし | **あり**（LangGraphのgraph/nodes/tools/state） |
| 永続化 | `repositories/`でDBを読み書き（中心的） | 状態は基本的に短命。永続化が要るなら同様に`repositories/`を足す |
| `infrastructure/`が繋ぐ相手 | DB、外部API | LLM（Bedrock等）、ベクトルDB、観測ツール（Langfuse等） |
| `schemas/` | リクエスト/レスポンス | それに加え**LLMの構造化出力用スキーマ**（`agent_schemas.py`） |
| `services/`の役割 | ユースケースを実装し`repositories`を呼ぶ | ユースケースを実装し`agents`（グラフ）を呼んで結果を変換 |

共通の骨格は **api / services / domain / schemas / infrastructure**。AI Agentはそこに`agents/`が乗るだけ、と捉えると分かりやすい。

## `__init__.py`について

`__init__.py`は、そのディレクトリを **Pythonパッケージとして扱うためのマーカーファイル**。空ファイルでも構わない。

### 厳密には必須ではない

Python 3.3以降、**Namespace Packages**（PEP 420）が導入され、`__init__.py`がなくてもパッケージとして機能するようになった。ただし実プロジェクトでは **明示的に置く（Regular Package にする）のがベストプラクティス**。

### 置くべき理由

1. **ツールの挙動が安定する**
   pytest、mypy、IDE（VSCode、PyCharm）は`__init__.py`があるディレクトリを明確にパッケージとして認識する。Namespace Packageだとツールによってはパッケージ検出に失敗したり、import解決がおかしくなることがある。特にpytestでは`__init__.py`の有無でテスト収集の挙動（rootdir判定、conftest.pyのスコープ）が変わる。

2. **意図せぬディレクトリがパッケージ化されるのを防ぐ**
   Namespace Packageだと`__init__.py`がないディレクトリも勝手にパッケージとして認識される。これがimportの曖昧さやバグの原因になることがある。

3. **パッケージレベルの初期化処理が書ける**
   パッケージ読み込み時に実行したい処理を記述できる。

   ```python
   # src/my_app/__init__.py
   from importlib.metadata import version

   __version__ = version("my_app")  # pyproject.toml の version を単一の情報源にする
   ```

> [!CAUTION]
> `__init__.py` で設定の読み込み（`Settings()` の生成）やログ初期化のような副作用の大きい処理をすると、`import my_app` した時点で実行される。必須の環境変数（例: Langfuseのキー）が無い環境（テスト・CI・`uv build` など）では **import だけで落ちる**ので、そうした処理は `main.py` の起動時や `get_settings()` の呼び出し側に置く。

4. **パブリックAPIを定義できる**
   `__all__` や再エクスポートで、パッケージの外向けインターフェースを明示できる。

   ```python
   # src/my_app/services/__init__.py
   from my_app.services.rca_service import RCAService
   from my_app.services.summary_service import SummaryService

   __all__ = ["RCAService", "SummaryService"]
   ```

   これで利用側は `from my_app.services import RCAService` と簡潔に書ける。

### `tests/`配下の`__init__.py`は別

`tests/`ディレクトリには **`__init__.py`を置かないのが推奨**。テストをパッケージ化しないことで余計な依存関係が発生せず、src layoutの「インストール済みパッケージを対象にテストする」という利点を濁さない。

注意点として、pytestデフォルトの `prepend` importモードでは、テストファイルのベース名が一意でないと衝突する（例: `tests/unit/test_logic.py` と `tests/integration/test_logic.py` はどちらも `test_logic` でモジュール名が衝突しエラーになる）。これを避けるには各ディレクトリに `__init__.py` を置く方法もあるが、**`pyproject.toml`で `--import-mode=importlib` を指定するのが現代的な推奨**（前掲の `[tool.pytest.ini_options]` 参照）。importlibモードなら **`__init__.py` を一切置かずに同名衝突も回避できる**ので、`tests/`は常に `__init__.py` なしで統一できる。

### まとめ

- `src/my_package/`配下のディレクトリには **必ず`__init__.py`を置く**（空でOK）
- `tests/`配下は **置かない**。`--import-mode=importlib` を設定すれば名前衝突も起きないので常に不要（デフォルトモードのまま同名ファイルを衝突させたくない場合のみ置く）
- 必須ではないが、ツール互換性と意図の明確化のために（`src/`配下では）置くのが現代的なベストプラクティス

## pyproject.tomlの設定例

`setup.py`、`setup.cfg`、`requirements.txt`は使わず、すべて`pyproject.toml`に集約する。1ファイルで「パッケージのメタ情報・依存・ビルド方法・各ツール設定」をまとめて管理できる。

`pyproject.toml`は **テーブル（`[...]`の見出し）ごとに役割が分かれている**のがポイント。まず全体像を示し、その後で各テーブルに何をどう書くかを解説する。

### `pyproject.toml`は手動作成ではなく `uv init` で自動作成される

uvを使う場合、**`uv init` を実行すると、必要最小限の `pyproject.toml` が自動作成される**（あわせて `.python-version` / `README.md` / `.gitignore` / `src/<name>/__init__.py` なども作成される）。下の「全体像」は、そこから `uv add` などで依存やツール設定を育てていった結果のイメージだと捉えるとよい。

`uv init` のモードによって作成される内容が変わる:

| コマンド | レイアウト | `[build-system]` | 用途 |
|---|---|---|---|
| `uv init`（= `--app`、デフォルト） | `src/` レイアウト | あり（`uv_build`）＋ `[project.scripts]` | Webサーバ・CLI・バッチなど、実行するアプリ |
| `uv init --package` | `src/` レイアウト | あり（`uv_build`）＋ `[project.scripts]` | デフォルトとほぼ同じ（パッケージ化を明示したいとき） |
| `uv init --lib` | `src/` レイアウト | あり（`uv_build`）＋ `py.typed`（`[project.scripts]` なし） | ライブラリとして配布する |
| `uv init --no-package` | ルート直下に `main.py` | なし（非パッケージ） | 配布・インストールしない簡易スクリプト |
| `uv init --bare` | なし | なし | `pyproject.toml` だけを作る最小構成 |

> [!NOTE]
> uv 0.12より前は、素の `uv init`（アプリ）は `[build-system]` を作らず `main.py` をルート直下に置いていた。現在はデフォルトでも `src/` レイアウト＋ビルドシステムありになり、旧来の挙動は `--no-package` で得られる（`--build-backend` を指定すると `--package` 扱いになる）。
>
> このドキュメントの構成（src layout・テスト同梱）なら、素の `uv init` / `uv init --package` / `uv init --lib` のいずれでもよい。

#### `--package` と `--lib` の使い分け

判断軸は「**自分で動かすもの**」か「**他人に import させるもの**」か。

- **`--package`（アプリ）** … 最終的に "実行する" もの。Webサーバ（FastAPI/Django）、CLIツール、バッチ・ジョブ、社内サービスなど。`src/`レイアウトでちゃんと作りたいが、PyPI等に配って他人に import させるわけではないケース。
- **`--lib`（ライブラリ）** … 最終的に `import` されることが目的のもの。PyPI公開パッケージ、複数プロジェクトから使われる社内共通ライブラリ、SDK・ユーティリティ群など。実行エントリポイントより**公開APIの提供**が主目的のケース。

```
そのコードのゴールは？
├─ 実行する（サーバ/CLI/バッチ） ……… --package
└─ import される（共通部品/SDK/公開pkg） … --lib
```

両者の構成（src layout＋`[build-system]`あり）はほぼ同じで、**実質的な違いは `py.typed` の有無と `[project.scripts]`（実行コマンド）の有無の2点**。`--package` には `[project.scripts]` が生成され、`--lib` には生成されない。`py.typed` は「このパッケージは型ヒント付き」というマーカーで、**import側の型チェッカー（mypy等）が型情報を読み取れるようにする**もの。他人に import される前提の `--lib` には付き、自分で動かす `--package` には付かない。

> [!TIP]
> 「アプリだが他プロジェクトからも import される」中間ケースは、`--package` で作って後から `py.typed` を手で追加すればよい（`[project.scripts]` はそのまま残して使える）。

`uv init --package my_app` 直後に自動作成される `pyproject.toml` は、おおむね次のような **最小の状態**になっている:

```toml
[project]
name = "my_app"
version = "0.1.0"
description = "Add your description here"
readme = "README.md"
requires-python = ">=3.12"   # 実行環境の.python-versionに応じた値
dependencies = []            # 依存はまだ空

[project.scripts]
my-app = "my_app:main"

[build-system]
requires = ["uv_build>=0.13.6,<0.14.0"]   # 値は uv のバージョンに応じて変わる
build-backend = "uv_build"
```

- **ビルドバックエンド**: 最近のuvがデフォルトで入れるのは `uv_build`。`hatchling` を使いたい場合は `--build-backend hatchling` を付けて `uv init` するか、あとから `[build-system]` を書き換える（このドキュメントの全体像では `hatchling` の例を採用している）。
- `dependencies` や各ツール設定（`[tool.ruff]` など）は **初期状態では入っていない**。次の「自動作成後の修正方法」で育てていく。

### `uv init` で自動作成された後の修正方法（特に依存関係）

自動作成された `pyproject.toml` は、**基本はCLIコマンドで編集し、`pyproject.toml`は自動で書き換わる**。手で直接編集してもよいが、その場合は最後に `uv sync` で環境とロックファイルを合わせる。

- **本番依存を追加**: `uv add fastapi` → `[project].dependencies` に `"fastapi>=0.115"` のように **自動で追記される**。同時に `uv.lock` 更新と `.venv` へのインストールも行われる
- **バージョン制約を付けて追加**: `uv add "fastapi>=0.115"`（PEP 508の指定子。範囲は `"pkg>=1.2,<2"`、extrasは `"pkg[a,b]"`、環境マーカーは `"pkg; python_version < '3.10'"`）
- **開発依存を追加**: `uv add --dev pytest` → `[dependency-groups].dev` に追記。本番ビルドでは `uv sync --no-dev` で除外できる
- **任意のグループに追加**: `uv add --group lint ruff` → `[dependency-groups].lint` に追記（`docs` など独自グループも作れる）
- **optional依存（extras）を追加**: `uv add httpx --optional network` → `[project.optional-dependencies]` に追記（配布パッケージ向け）
- **依存を削除**: `uv remove fastapi`（`--dev` / `--group <name>` / `--optional <extra>` で対象セクションを指定）
- **バージョン制約を変更**: 新しい制約で `uv add` をやり直す（例 `uv add "fastapi>=0.116"`）。最新へ上げたいだけなら `uv add --upgrade-package fastapi`
- **`requires-python` や `[tool.*]` 設定**: これらはCLIでは触れないので **`pyproject.toml`を直接編集**し、依存に関わる変更なら `uv sync` を実行して反映する

> [!TIP]
> まとめ: **依存は `uv add` / `uv remove`（手書きしない）、ツール設定やメタ情報（`requires-python`・`[tool.ruff]`など）は直接編集 → `uv sync`**。どちらの場合も `uv.lock` は自動でメンテされるのでコミットすればよい。

以下が、こうして育てた `pyproject.toml` の全体像:

```toml
[project]
name = "my_app"
version = "0.1.0"
description = "AI Agent for Root Cause Analysis"
requires-python = ">=3.12"
dependencies = [
    "fastapi>=0.115",
    "uvicorn>=0.54.0",       # fastapi 単体では入らない（DockerのCMD・main.pyで使用）
    "httpx>=0.28.1",         # infrastructure/prometheus_client.py で使用
    "langfuse>=5.1.0",       # infrastructure/langfuse_client.py で使用
    "pydantic>=2.9",
    "pydantic-settings>=2.5",
    "langgraph>=0.2",
    "langchain-aws>=0.2",
]

[dependency-groups]
dev = [
    "pytest>=8.3",
    "pytest-asyncio>=0.24",
    "pytest-cov>=5.0",
    "ruff>=0.7",
    "mypy>=1.13",
    "pre-commit>=4.0",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP", "B", "SIM", "RUF"]

[tool.mypy]
strict = true
python_version = "3.12"

[tool.pytest.ini_options]
testpaths = ["tests"]
asyncio_mode = "auto"
addopts = "--import-mode=importlib --cov=my_app --cov-report=term-missing"
```

### 各テーブルに何を・どう書くか

どのテーブルに何を設定するかを役割ごとに整理する。

#### `[project]` — パッケージのメタ情報と本番依存

- **何を**: パッケージ名・バージョン・対応Pythonバージョン（`requires-python`）・**本番で必要な依存**（`dependencies`）
- **どう**: `dependencies`は手書きせず `uv add <pkg>` で追加すると自動でここに追記される。`requires-python`は対応する最小バージョンを書く（例 `>=3.12`）
- **補足**: CLIコマンドを提供したいなら `[project.scripts]` でエントリポイントを定義する

  ```toml
  [project.scripts]
  my-app = "my_app.main:main"   # `my-app`コマンドで my_app/main.py の main() を起動
  ```

#### `[dependency-groups]` — 開発用依存（PEP 735）

- **何を**: テスト・lint・型チェックなど **本番には不要な開発ツール**
- **どう**: `uv add --dev <pkg>` で `dev` グループに自動追記される。本番ビルドでは `uv sync --no-dev` で除外できる
- **補足**: 旧来の `[tool.uv] dev-dependencies` は非推奨。今は `[dependency-groups]` が標準。`dev` 以外のグループ（例 `docs`, `lint`）も作れる

#### `[build-system]` — ビルドバックエンド

- **何を**: パッケージをビルドするツール（バックエンド）の指定
- **どう**: uvがデフォルトで生成するのは `uv_build`（uv公式のビルドバックエンド）。`hatchling` もシンプルで広く使われており、`uv init --build-backend hatchling` で選べる。いずれも `uv init`（`--package`/`--lib`）が自動生成するので通常はそのままでよい（この全体像の例では `hatchling` を採用）
- **補足（src layoutの注意・hatchlingの場合）**: **プロジェクト名（正規化後）とパッケージのディレクトリ名が一致していれば** hatchlingは `src/<name>/` を自動検出するので追加設定は不要。一致しない場合（例: `name = "my-cool-app"` だがディレクトリが `src/coolapp/`）は明示が必要

  ```toml
  [tool.hatch.build.targets.wheel]
  packages = ["src/coolapp"]
  ```

#### `[tool.ruff]` / `[tool.ruff.lint]` — リンタ&フォーマッタ

- **何を**: `line-length`・`target-version`は `[tool.ruff]` に、**有効化するルールの選択（`select`）は `[tool.ruff.lint]`** に書く
- **どう**: `select`で使うルールのカテゴリを選ぶ（`E`/`F`=基本、`I`=import整列、`UP`=構文の近代化、`B`=bugbear、`SIM`=簡約、`N`=命名 など）
- **補足**: ruff 0.2.0以降、lint設定は `[tool.ruff.lint]` 配下に移動した。トップレベルの `[tool.ruff] select` は非推奨

#### `[tool.mypy]` — 型チェック

- **何を**: 型チェックの厳格さ（`strict = true`）、対象Pythonバージョン（`python_version`）
- **どう**: まず `strict = true` で始め、厳しすぎる箇所はモジュール単位で個別に緩める運用が定番

  ```toml
  [[tool.mypy.overrides]]
  module = ["some_untyped_lib.*"]
  ignore_missing_imports = true   # 型情報のないサードパーティを個別に許容
  ```

#### `[tool.pytest.ini_options]` — テスト

- **何を**: テストの探索パス（`testpaths`）、importモード（`--import-mode`）、デフォルトオプション（`addopts`）、pytest-asyncioのモード（`asyncio_mode`）
- **どう**: `addopts` に常用フラグ（カバレッジ等）を入れておくと毎回打たずに済む。カバレッジは **`--cov=src`（パス指定）ではなく `--cov=my_app`（パッケージ名指定）** にする。src layoutでは「インストール済みパッケージ」を対象にテストするため、パスではなく `import` 名で計測対象を指す方がインストール先に追従して正確（CIやDockerで実際にwheelを入れてテストしてもズレない）
- **`--import-mode=importlib` を推奨**: src layoutではpytest公式（Good Integration Practices）が `importlib` モードを推奨。これにより **`tests/`配下に `__init__.py` を一切置かなくても、同名のテストファイル（`unit/test_x.py` と `integration/test_x.py`）が衝突しない**。デフォルトの `prepend` モードだと衝突回避のために `__init__.py` が必要になる

> [!TIP]
> どこに書くか迷ったら原則: **ツール固有の設定は `[tool.<ツール名>]`、パッケージ自体の情報は `[project]`、依存は `[project]`（本番）か `[dependency-groups]`（開発）**。

## 必須の補助ファイル

### .gitignore

```gitignore
# Python
__pycache__/
*.py[cod]
*$py.class
*.egg-info/
dist/
build/

# venv
.venv/

# 環境変数
.env
.env.local

# IDE
.vscode/
.idea/

# テスト・カバレッジ
.pytest_cache/
.coverage
htmlcov/
.mypy_cache/
.ruff_cache/
```

`.python-version`と`uv.lock`は **コミット対象**。除外しないように注意。

### .python-version

Pythonバージョンを固定するファイル（uvが自動生成）。

```
3.12
```

### .env.example

実際の`.env`はコミットせず、テンプレートだけ共有する。

```bash
# AWS
AWS_REGION=ap-northeast-1
BEDROCK_MODEL_ID=global.anthropic.claude-sonnet-5-5

# Langfuse
LANGFUSE_PUBLIC_KEY=
LANGFUSE_SECRET_KEY=
LANGFUSE_BASE_URL=https://cloud.langfuse.com

# Application
LOG_LEVEL=INFO
```

### pre-commit設定（.pre-commit-config.yaml）

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.17.3
    hooks:
      - id: ruff-check
        args: [--fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v2.4.1
    hooks:
      - id: mypy
        additional_dependencies: [pydantic, types-requests]
```

> [!NOTE]
> リンタのフックIDは ruff-pre-commit v0.11.10（2025年5月）以降 **`ruff-check`** にリネームされた（それ以前の `rev` には `ruff-check` が存在せず、指定するとエラーになる）。旧来の `ruff` はレガシーエイリアスとして今も動くが、現行は `ruff-check` を使う。`rev` は上記の値（記事執筆時点の最新）に固定されるので、定期的に `uv run pre-commit autoupdate` で更新する。`--fix`を使う場合は `ruff-check` を `ruff-format` より**前**に置くこと。

セットアップ:

```bash
uv add --dev pre-commit
uv run pre-commit install
```

#### `.pre-commit-config.yaml` は手動で作成する

`.pre-commit-config.yaml` は `pip install` などで自動生成されるファイルではなく、**プロジェクトルートに自分で作成する**。使いたいフックを記述していく（各ツールの README に pre-commit 用の設定例が載っていることが多いので、それをコピペして組み立てるのが実際の流れ）。

ゼロから書くのが面倒な場合は、最小の雛形を生成するコマンドもある:

```bash
pre-commit sample-config > .pre-commit-config.yaml
```

ただし結局そこから自分で編集して使うフックを足していくことになる。

> [!WARNING]
> 設定ファイルが無い状態で `pre-commit install` してもフックは何も実行されない（むしろ commit 時にエラーになる）。**「設定ファイルを手動で用意」→「install で Git に組み込む」** の2段構え。

#### `pre-commit install` が具体的にすること

`pre-commit install` は、**`.pre-commit-config.yaml` に書いたチェックを `git commit` のたびに自動実行するよう Git に登録する**コマンド。

- 実行すると、リポジトリ内の **`.git/hooks/pre-commit`** スクリプトを生成（上書き）する。
- Git にはもともと「commit 前に `.git/hooks/pre-commit` があれば実行する」仕組みがあり、そこに pre-commit ツールを呼び出すスクリプトを書き込む。
- 以降 `git commit` するたびに、変更ファイルに対して `ruff` / `mypy` などのチェックが自動で走り、**失敗すると commit が中断される**。

```
git commit
   ↓ .git/hooks/pre-commit が起動
   ↓ pre-commit が .pre-commit-config.yaml を読む
   ↓ ruff / mypy などを変更ファイルに実行
全部パス → commit 成立 ／ 失敗 → commit が中断（修正してやり直し）
```

ポイント:

- **1回だけ実行すればよい**。リポジトリを clone した直後などに1度実行すれば以降ずっと有効。
- **`.git/hooks/` は Git 管理外**なので、clone した人は各自 `pre-commit install` を実行する必要がある（README に書いておく）。
- `pip install pre-commit`（＝ツール本体の導入）とは別物。`pre-commit install` は**そのツールを今のリポジトリの Git フックに組み込む**操作。
- 解除は `pre-commit uninstall`。
- commit せず手動で全ファイルをチェックするには `uv run pre-commit run --all-files`。
- push 時に走らせたい場合は `pre-commit install --hook-type pre-push`。

## テストの構成

```
tests/
├── conftest.py          # pytest共通fixture
├── unit/                # 単体テスト（外部依存なし）
│   ├── test_logic.py
│   └── services/
│       └── test_rca_service.py
├── integration/         # 統合テスト（DB、API呼び出しあり）
│   └── test_bedrock.py
└── e2e/                 # E2Eテスト
    └── test_full_flow.py
```

テストはプロダクションコードを **ミラーする構造** にすると見通しが良い。`src/my_app/services/foo.py` に対して `tests/unit/services/test_foo.py` のように対応させる。

## Dockerでの利用

**`Dockerfile` はプロジェクトルート（`pyproject.toml` と同じ階層）に置く**のが標準。イメージが1つならわざわざ `docker/` を切らず直下に置く。こうすると `docker build -t my_app .` だけでよく、`COPY pyproject.toml uv.lock ./` や `COPY src/ ./src/` がコンテキスト（＝ルート）基準でそのまま解決される。

```text
my_app/
├── src/
├── tests/
├── pyproject.toml
├── uv.lock
└── Dockerfile          ← ルート直下
```

uvの公式イメージを使うと、本番デプロイ用のイメージも軽量に作れる。

```dockerfile
FROM python:3.12-slim

# uvをインストール
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

WORKDIR /app

# 依存関係を先にコピー（キャッシュ効率化）
COPY pyproject.toml uv.lock ./

# 依存のみインストール（プロジェクト本体はまだ入れない）
RUN uv sync --frozen --no-dev --no-install-project

# アプリケーションコードをコピー（README.md は pyproject.toml の readme が参照するため必須）
COPY README.md ./
COPY src/ ./src/

# ソースを入れた後にプロジェクト本体をインストール
RUN uv sync --frozen --no-dev

# 仮想環境のPythonを使う
ENV PATH="/app/.venv/bin:$PATH"

# 本番起動はuvicornを直接呼ぶ（reloadは付けない）
CMD ["uvicorn", "my_app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

ポイント:

- `--frozen`は`uv.lock`を変更しないモード、`--no-dev`は開発依存をスキップする
- **`uv sync`はデフォルトでプロジェクト自身も（editableで）インストールしようとする**ため、ソースをコピーする前に実行すると失敗する。そこで依存だけ入れる層で`--no-install-project`を付け、ソースをコピーした後にもう一度`uv sync`してプロジェクト本体を入れる（uv公式のDocker統合ガイド推奨の2段構成）。これにより依存関係の層がソース変更で無効化されず、ビルドキャッシュが効く
- **`README.md` を `COPY` しないとビルドが失敗する**。`uv init` が生成する `pyproject.toml` には `readme = "README.md"` があり、プロジェクト本体のインストール（2回目の `uv sync`）時にビルドバックエンドがこのファイルを読むため、無いと `failed to open file '/app/README.md'`（hatchling なら `Readme file does not exist: README.md`）で落ちる。`.dockerignore` に `*.md` を入れている場合は `!README.md` で除外を解除しないと、`COPY README.md` 自体が失敗する。依存だけの1回目（`--no-install-project`）は README が無くても通るので、**ソースをコピーした後にだけ発覚する**点に注意。

### .dockerignore

Dockerはビルド時に**コンテキスト（＝ルート）全体をデーモンに送る**ため、`.venv/`や`.git/`を除外しないとビルドが遅くなり、ホストの`.venv`がイメージに紛れ込む事故も起きる。`.gitignore`とは別に **`.dockerignore` を必ず置く**。

```dockerignore
.venv/
.git/
__pycache__/
*.pyc
.pytest_cache/
.mypy_cache/
.ruff_cache/
.env
tests/
docs/
*.md
# pyproject.toml の readme が参照するので README.md は除外しない
!README.md
```

## フロントエンドとバックエンドを1リポジトリで管理する場合（モノレポ）

フロントエンド（React / Vue など）とバックエンド（Python）を1つのリポジトリに同居させる構成。

### まず「モノレポにすべきか」

「uv を使うからモノレポ」ではなく、**変更が両方にまたがるか／別々にデプロイするか**で決める。

| | モノレポが向く | 別リポジトリが向く |
|---|---|---|
| 変更の単位 | 1つの機能で API と UI を同時に変えることが多い | 片方だけ変わることがほとんど |
| デプロイ | 一緒に出す、または同じパイプラインで出し分けられる | リリース周期・担当チームが別 |
| 型の共有 | OpenAPI から生成したクライアントを同じ PR で更新できる | 別リポジトリだと生成物のバージョン管理が別途必要 |

- 1人〜小さなチームで、API と UI が密結合なら、モノレポの方が PR・レビュー・ブランチが1つで済む。
- モノレポでも**デプロイは独立にできる**（CI で出し分ければよい）。逆に「A だけデプロイすると B が壊れる」状態なら、リポジトリを分けても境界は偽物（分散モノリス）。

### 前提: uv が扱うのは Python だけ

- **uv が管理するのは Python の依存だけ**。フロントエンドの依存は npm / pnpm / bun など JS 側のツールで別に管理する（uv は `package.json` を扱わない）。
- uv の **workspace は「複数の Python パッケージ」を1つのロックファイルで束ねる機能**で、フロントエンドを束ねる機能ではない。
- したがって `frontend/` と `backend/` は**それぞれのツールのプロジェクトとして並べる**だけで、uv 側から見て `frontend/` は存在しないものとして扱う。

### 構成の選択肢

**どちらも実在する構成で、優劣ではなく「Python パッケージがいくつあるか」で選ぶ。**

| | パターンA（独立プロジェクト） | パターンB（ルートに workspace） |
|---|---|---|
| Python パッケージの数 | 1つで足りる | 複数（API / ワーカー / 共有ライブラリなど）、または将来増える見込み |
| `pyproject.toml` / `uv.lock` / `.venv` | すべて `backend/` | `pyproject.toml`（workspace 定義）・`uv.lock`・`.venv` はルート、各 member に `pyproject.toml` |
| コマンドの実行場所 | `backend/`（ルートからは `uv run --project backend`） | ルートでも member のディレクトリ内でもよい |
| Dockerfile | `backend/` に置き、ビルドコンテキストも `backend/` | `backend/` に置いても、**ビルドコンテキストはルート**にする |
| 採用例 | — | FastAPI 公式の [full-stack-fastapi-template](https://github.com/fastapi/full-stack-fastapi-template) |

> [!TIP]
> 迷ったらパターンA で始め、Python パッケージが2つ目（ワーカー・共有ライブラリなど）になった時点でパターンB に移行する。移行は「ルートに `[tool.uv.workspace]` だけの `pyproject.toml` を足す」程度で済む。

### パターンA: `backend/` を独立した uv プロジェクトにする

`backend/` の中に前述の構成（`pyproject.toml` / `uv.lock` / `src/` / `tests/`）をそのまま置く。

```text
my_product/
├── backend/                    ← uv プロジェクト（この記事のここまでの構成がそのまま入る）
│   ├── src/
│   │   └── my_app/
│   ├── tests/
│   ├── pyproject.toml
│   ├── uv.lock
│   ├── .python-version
│   ├── README.md
│   └── Dockerfile
├── frontend/                   ← JS 側のプロジェクト（uv は関与しない）
│   ├── src/
│   ├── package.json
│   ├── bun.lock                （npm なら package-lock.json、pnpm なら pnpm-lock.yaml）
│   └── Dockerfile
├── compose.yml                 ← 両方をまとめて起動
├── .env.example
├── .gitignore                  ← .venv/ と node_modules/ の両方を除外
├── .pre-commit-config.yaml
└── README.md
```

- **コマンドは `backend/` に入って実行**するのが基本（`cd backend && uv sync && uv run pytest`）。ルートから実行したいときは `uv run --project backend ...`（手元で動作確認済み）。
- Dockerfile はそれぞれのディレクトリに置き、`compose.yml` で `build: ./backend` / `build: ./frontend` と指定する。**ビルドコンテキストが `backend/` になるので、前章の Dockerfile をそのまま使える**。
- フロントとバックエンドを別々にデプロイする（フロントは静的ホスティング、バックエンドはコンテナ、など）場合に素直。

### パターンB: ルートに uv workspace を置く

Python パッケージが複数ある場合や、ロックファイルを1つにまとめたい場合は、リポジトリルートを workspace のルートにする。

```text
my_product/
├── pyproject.toml              ← workspace の定義だけ（[tool.uv.workspace]）
├── uv.lock                     ← ロックファイルはルートに1つだけ
├── .python-version
├── backend/                    ← workspace member（Python パッケージ）
│   ├── src/
│   │   └── my_app/
│   ├── tests/
│   ├── pyproject.toml
│   ├── README.md
│   └── Dockerfile              ← 置き場所は backend/ でよい（ビルドコンテキストはルート）
├── frontend/                   ← members に含めない
│   ├── src/
│   ├── package.json
│   └── Dockerfile
├── compose.yml
├── .env.example
└── README.md
```

ルートの `pyproject.toml`:

```toml
# my_product/pyproject.toml（workspace のルート）
[tool.uv.workspace]
members = ["backend"]           # Python パッケージだけを明示する
```

`backend/pyproject.toml` は通常の uv プロジェクトと同じ（`uv init --package backend` で作れる）。

> [!NOTE]
> uv 公式ドキュメントの例はすべてルートにも `[project]` を持たせており、上のような「`[project]` を書かないルート」を説明するページは見当たらない。ただし FastAPI 公式テンプレートがこの形で、手元の uv 0.13.6 でも `uv sync` / `uv lock` / `uv run` が問題なく動いた。ルートも Python パッケージにしたい場合は、通常どおり `[project]` を足し、`members` にルート自身を含める（ルートは常に member になる）。

```bash
uv sync                           # ルートで実行。.venv と uv.lock はルート直下に作られる
uv sync --package my-app          # 特定の member（と、その依存する member）だけ同期する
uv run --package my-app pytest    # member を指定して実行
uv add --package my-app fastapi   # member に依存を追加
uv add --package my-app shared    # member 同士の依存。`{ workspace = true }` が自動で付く
```

手元で複数 member（`api` / `worker` / `shared`）を作って確認した動作:

- **ルートで素の `uv sync` を実行すると、全 member がインストールされる**（`--package` を付けなくてよい）。`--package` は「一部だけ入れたい」ときに使う（例: `--package api` だと `worker` は入らない）。
- `backend/` の中に入って `uv sync` / `uv run` を実行しても、**`.venv` と `uv.lock` は常にルートのものが使われる**（`backend/.venv` は作られない）。
- member の `[dependency-groups]` の `dev`（pytest など）は、ルートでの `uv sync` でデフォルトで入る。本番イメージでは `--no-dev` で外す。
- `uv add --package api shared` のように member 同士の依存を足すと、`api/pyproject.toml` に `shared = { workspace = true }` が自動で書かれる（手書き不要）。

> [!IMPORTANT]
> - `--package` に渡すのは**ディレクトリ名ではなく `pyproject.toml` の `[project].name`**。`uv init --package backend --name my_app` のようにディレクトリ名と `name` が違うなら `--package my_app`（`--package backend` だと `The workspace does not have a member backend` になる）。
> - workspace 全体で **`requires-python` は全 member の共通部分**が使われ、依存も1つの解決結果に揃う。member 同士で依存が衝突する、member ごとに別の仮想環境が必要、という場合は workspace を避け、各パッケージを独立プロジェクトにして `path` 依存でつなぐ（`uv run --package` は使えなくなり、各パッケージのディレクトリ内で実行することになる）。
> - Python は依存を隔離しないので、uv は member が**他の member だけが宣言している依存を import してしまう**ことを防げない。member ごとに依存をきちんと宣言すること。

#### `frontend/` を members に一致させない

`members` のパターンに一致したディレクトリは `pyproject.toml` を必須とするため、`package.json` しか無い `frontend/` が一致すると `uv lock` が失敗する（手元で確認済み）。

```text
error: Workspace member `.../apps/web` missing a `pyproject.toml` (matches: `apps/*`)
```

回避策は3つ（いずれも手元で動作確認済み）:

1. **`members` に Python のディレクトリだけを明示する**（`members = ["backend"]`）。最も単純で、上の構成はこれ。
2. **`frontend/` を `apps/` の外に置く**。`members = ["apps/*"]` のようにグロブを使いたいときは、フロントを `apps/` 配下に置かない。
3. **グロブを使いつつ `exclude` で除外する**: `members = ["apps/*"]` + `exclude = ["apps/web"]`。

#### フロントエンドも含めて1つのイメージにする Dockerfile 例（パターンB）

**ビルドコンテキストはリポジトリルート**にする（`uv.lock` と workspace の `pyproject.toml` がルートにあるため）。Dockerfile 自体は `backend/` に置いてよく、その場合は `docker build -f backend/Dockerfile .` のようにコンテキストだけをルートにする。フロントエンドをビルドして成果物を Python 側のイメージに取り込むマルチステージの形:

```dockerfile
# --- ステージ1: フロントエンドをビルド ---
FROM oven/bun:1 AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY frontend/ ./
RUN bun run build                       # 成果物は dist/ に出力される想定

# --- ステージ2: Python アプリ ---
FROM python:3.12-slim
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy
ENV PATH="/app/.venv/bin:$PATH"
WORKDIR /app

# 依存だけ先にインストール（member 自身は入れない → コード変更でキャッシュが効く）
COPY pyproject.toml uv.lock ./
COPY backend/pyproject.toml ./backend/
RUN uv sync --frozen --no-dev --no-install-workspace --package my_app

# アプリ本体とフロントの成果物をコピーして、member 自身をインストール
COPY backend/README.md ./backend/          # pyproject の readme を参照するため必須
COPY backend/src ./backend/src
COPY --from=frontend-build /app/frontend/dist ./backend/src/my_app/static
RUN uv sync --frozen --no-dev --package my_app

CMD ["uvicorn", "my_app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

> [!NOTE]
> - workspace では前章の `--no-install-project` ではなく **`--no-install-workspace`** を使う（ルートだけでなく member も除外して、依存だけの層を作る）。`--package my_app` で必要な member だけを同期する。
> - **1回目は `--frozen`、2回目は `--locked`（または `--frozen`）**。uv 公式は「member の `pyproject.toml` が無いと `uv.lock` が最新か検証できないので、最初の sync は `--locked` ではなく `--frozen` を使う」と説明している。手元でも、member の `pyproject.toml` が欠けた状態の `--locked` は `missing a pyproject.toml` で失敗した。全 member の `pyproject.toml` を `COPY` した後なら `--locked` が使え、ロックと `pyproject.toml` のずれを検出できる。
> - `pyproject.toml` に `readme = "README.md"` がある（`uv init` のデフォルト）ので、**2回目の `uv sync` の前に `backend/README.md` もコピーする**。コピーしないと `failed to open file '/app/backend/README.md'` でビルドが失敗する（実際に踏んだ）。`.dockerignore` に `*.md` を入れている場合は `!backend/README.md` で除外を解除すること。
> - この Dockerfile は手元でビルドして、コンテナ起動後に `/openapi.json` が HTTP 200 を返し、フロントのビルド成果物が `static/` に入っていることを確認した。フロントのビルド出力先や静的ファイルの配信方法（FastAPI の `StaticFiles` か nginx か）はプロジェクトに合わせて調整すること。
> - 実運用の例は FastAPI 公式テンプレートの `backend/Dockerfile`。そちらは `fastapi run` を使っている。FastAPI 公式ドキュメントは、`CMD` を必ず **exec 形式**（`["fastapi", "run", ...]`）で書くこと（シェル形式だと graceful shutdown や lifespan イベントが正しく動かない）、ロードバランサ / nginx の背後では `--proxy-headers` を付けることを案内している。

### 型・API クライアントの共有

Python と TypeScript はコードを直接共有できないので、**FastAPI が出力する OpenAPI スキーマから TypeScript のクライアント／型を生成**するのが定番。

- 手順は「バックエンドの OpenAPI JSON を出力 → フロントのクライアントを生成 → **生成物をコミット**」。FastAPI 公式テンプレートは `scripts/generate-client.sh` でこれをやっていて、`app.openapi()` を `uv run python -c` で出力し（サーバや DB の起動は不要）、フロントの `generate-client` スクリプトを実行している。
- 生成ツールは `@hey-api/openapi-ts`（SDK と TanStack Query のオプションまで生成）や `openapi-typescript` + `openapi-fetch`（型だけ生成）などが使われている。
- **CI で「再生成して差分が出たら失敗」させる**と、バックエンドの API を変えたのにフロントの型を更新し忘れる事故を防げる。ここがモノレポの最大の利点を活かすところ。
- FastAPI の OpenAPI には `servers` が含まれず、生成されたクライアントに `baseUrl` が入らないので、フロント側で設定する。

### 共通の注意点

- **`.gitignore`**: Python 側（`.venv/` など）に加えて、`node_modules/`、フロントのビルド成果物（`dist/`）も除外する。
- **`.env` / `.env.example`**: バックエンド用の変数（Bedrock・Langfuse のキーなど）はサーバ側だけに置く。Vite は `VITE_` で始まる変数をビルド時にバンドルへ埋め込み、**ブラウザから見えてしまう**（Vite 公式も「API キーなどの機密情報を入れてはいけない」としている）。フロントに渡す変数に秘密情報を入れない。
- **pre-commit**: 設定ファイルはルートに1つ置き、ruff / mypy のフックには `files: ^backend/` を付けて対象を絞る（手元で、`frontend/` の違反が無視され `backend/` だけ検出されることを確認）。フロント用のフック（prettier / eslint など）も同じファイルに並べられる。
- **タスクランナー**: `cd backend && uv run ...` と `cd frontend && bun run ...` を行き来するのが面倒になったら、ルートに `Makefile` や `justfile` を置いて `make test` / `make dev` にまとめるとよい。

### CI（GitHub Actions）

**変更パスで出し分けたいからといって、workflow 全体に `on.pull_request.paths` を付けてはいけない。**

> [!CAUTION]
> パスフィルタ・ブランチフィルタ・コミットメッセージで **workflow 全体がスキップされると、そのチェックは「Pending」のまま**になり、そのチェックを必須（required）にしている PR は**マージできなくなる**（[GitHub 公式](https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions)）。一方、workflow 内の**ジョブが `if:` でスキップされた場合は「Success」扱い**になる。

推奨される形:

1. **workflow には `paths:` を付けず、常に起動させる**。
2. 最初のジョブで変更ファイルを判定する（[dorny/paths-filter](https://github.com/dorny/paths-filter) など）。
3. `backend` / `frontend` のジョブは、判定結果で **ジョブレベルの `if:`** を使って実行／スキップする。
4. **必須チェックは、常に走る「集約ジョブ」（例: `ci-gate`）1つだけ**にし、他のジョブが成功またはスキップのときに成功させる。

```yaml
# .github/workflows/ci.yml（骨子）
on:
  pull_request:
  push:
    branches: [main]            # workflow に paths: は付けない

jobs:
  changes:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: read       # paths-filter が PR の変更ファイル一覧を API で取得するのに必要
    outputs:
      backend: ${{ steps.f.outputs.backend }}
      frontend: ${{ steps.f.outputs.frontend }}
    steps:
      - uses: actions/checkout@v7
      - uses: dorny/paths-filter@v4
        id: f
        with:
          filters: |
            backend:
              - 'backend/**'
              - 'pyproject.toml'
              - 'uv.lock'
            frontend:
              - 'frontend/**'

  backend:
    needs: changes
    if: needs.changes.outputs.backend == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: astral-sh/setup-uv@v10
        with:
          enable-cache: true
      - run: uv sync --locked          # パターンB はルート、パターンA は working-directory: backend
      - run: uv run --package my_app pytest

  frontend:
    needs: changes
    if: needs.changes.outputs.frontend == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile && bun run build && bun test

  ci-gate:                          # ← ブランチ保護の必須チェックはこれだけ
    if: always()
    needs: [backend, frontend]
    runs-on: ubuntu-latest
    steps:
      - run: |
          [[ "${{ contains(needs.*.result, 'failure') }}" == "false" && "${{ contains(needs.*.result, 'cancelled') }}" == "false" ]]
```

> [!NOTE]
> - 上の YAML は骨子で、アクションのバージョンは記事執筆時点の最新メジャー（`checkout` v7、`setup-uv` v10、`paths-filter` v4）。実運用では commit SHA 固定が推奨される（uv 公式ガイドも `setup-uv` を SHA 固定している）。FastAPI 公式テンプレートの `test-backend.yml` は `paths:` を使わず、常に全テストを走らせている。**小規模なら出し分けずに常に全部走らせる**のも立派な選択。
> - `astral-sh/setup-uv` の `cache-dependency-glob` はデフォルトで `**/uv.lock` や `**/pyproject.toml` などを拾うので、モノレポの深い階層の lockfile も対象になる。`working-directory` で `pyproject.toml` を探すディレクトリを変えられる。

### Dependabot

`.github/dependabot.yml` で、バックエンド（uv）とフロントエンド（npm など）を**別々の ecosystem エントリ**として登録する。

```yaml
version: 2
updates:
  - package-ecosystem: "uv"
    directory: "/backend"           # パターンB は "/"（workspace ルートに uv.lock がある）
    schedule:
      interval: "weekly"
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
```

- `directory` は **`pyproject.toml` と `uv.lock` がある場所**を指す。uv workspace は `uv.lock` がルートに1つだけなので、**member ごとに並べず、workspace ルートの1エントリ**にする。
- 複数ディレクトリは `directories:`（複数形）で書け、グロブも使える（単数形の `directory` はグロブ不可）。同じ ecosystem のエントリ同士でディレクトリが重複しないようにする。
- uv workspace まわりは Dependabot 側にまだ既知の不具合報告がある（[dependabot-core #12072](https://github.com/dependabot/dependabot-core/issues/12072)）ので、導入したら実際に PR が作られるかを確認すること。

## 各サブディレクトリの役割と実装例

> [!NOTE]
> この章は **AI Agentを題材にしたコードの詳解**。各レイヤーに実際どんなコードを書くかを具体例で示す。**設定や全体像だけ知りたい場合はこの章を飛ばして「pyproject.tomlの設定例」以降へ進んでよい。**

以降は **AI Agentの例** を題材に各レイヤーのコードを見ていく（`agents/`以外は通常のWeb APIでもほぼ同じ）。`services`は`infrastructure`を呼ぶが逆はNG、`domain`は誰にも依存しない純粋なモデル層、という依存方向は両構成に共通する。

### `config.py` — 設定管理

環境変数や設定値を一元管理する。`pydantic-settings`を使うのが定番。

```python
# src/my_app/config.py
from functools import lru_cache
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # AWS
    aws_region: str = "ap-northeast-1"
    bedrock_model_id: str = "global.anthropic.claude-sonnet-5-5"

    # Langfuse
    langfuse_public_key: str
    langfuse_secret_key: str
    langfuse_base_url: str = "https://cloud.langfuse.com"

    # Prometheus
    prometheus_url: str = "http://prometheus:9090"

    # App
    log_level: str = "INFO"
    max_iterations: int = Field(default=10, ge=1, le=50)


@lru_cache
def get_settings() -> Settings:
    return Settings()
```

ポイント:

- `@lru_cache`でシングルトン化
- 利用側は `from my_app.config import get_settings` で取得

### `domain/` — ドメインモデル

ビジネスの本質を表現する層。**外部ライブラリやフレームワークに依存しない** 純粋なPythonコード。

```python
# src/my_app/domain/models.py
from dataclasses import dataclass
from datetime import datetime
from enum import StrEnum


class IncidentSeverity(StrEnum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


@dataclass(frozen=True)
class Incident:
    """インシデントを表すドメインモデル"""
    id: str
    title: str
    severity: IncidentSeverity
    started_at: datetime
    affected_services: list[str]

    def is_critical(self) -> bool:
        return self.severity == IncidentSeverity.CRITICAL


@dataclass(frozen=True)
class RootCause:
    """RCA分析の結果"""
    incident_id: str
    summary: str
    contributing_factors: list[str]
    confidence_score: float  # 0.0 - 1.0

    def is_high_confidence(self) -> bool:
        return self.confidence_score >= 0.8
```

ポイント:

- `@dataclass(frozen=True)`で不変オブジェクトに
- ビジネスルール（`is_critical`、`is_high_confidence`）はモデル自身に持たせる
- ここでPydanticを使ってもよいが、I/O境界（API/DB）とドメインを混ぜたくない場合は分ける

### `schemas/` — Pydanticスキーマ（I/O境界）

API入出力やLLM出力のバリデーション用。**domainとは別物** として扱う。

```python
# src/my_app/schemas/api_schemas.py
from datetime import datetime
from pydantic import BaseModel, Field


class RCARequest(BaseModel):
    """RCA分析リクエスト"""
    incident_id: str = Field(..., description="インシデントID")
    time_range_minutes: int = Field(default=60, ge=5, le=1440)
    services: list[str] = Field(default_factory=list)


class RCAResponse(BaseModel):
    """RCA分析レスポンス"""
    incident_id: str
    summary: str
    contributing_factors: list[str]
    confidence_score: float
    analyzed_at: datetime
```

```python
# src/my_app/schemas/agent_schemas.py
from pydantic import BaseModel, Field


class RCASummary(BaseModel):
    """LLMに構造化出力させるためのスキーマ"""
    root_cause: str = Field(..., description="根本原因の説明")
    evidence: list[str] = Field(..., description="根拠となる観測事実")
    recommended_actions: list[str] = Field(..., description="推奨される対応策")
    confidence: float = Field(..., ge=0.0, le=1.0)
```

### `domain` と `schemas` の違い

両者は「データを表すクラス」という見た目が似ているため混同しやすいが、**役割（誰のための形か）がまったく違う**。

| 観点 | `domain`（ドメインモデル） | `schemas`（スキーマ） |
|---|---|---|
| 何のための形か | **アプリ内部**でビジネスを表現する形 | **外部との境界（I/O）** でやり取りする形 |
| 形を決めるのは誰か | 自分たちのビジネスルール | 外の都合（API仕様、LLMの出力形式、DBの列） |
| 主な責務 | 状態 + **振る舞い**（ビジネスロジック） | **バリデーション**とシリアライズ（データの入れ物） |
| 依存 | 何にも依存しない純粋なPython | Pydantic等のI/Oライブラリに依存 |
| 変わりやすさ | 安定（仕様の本質なので滅多に変わらない） | 変わりやすい（API改版・LLM変更で頻繁に変わる） |
| 典型例 | `Incident`, `RootCause`（`is_critical()`等を持つ） | `RCARequest`, `RCAResponse`, `RCASummary` |

#### イメージ：同じ「概念」でも形が複数ある

たとえば「RCAの結果」という1つの概念でも、登場する場所ごとに別の形を持つ:

- `RCASummary`（schema）… **LLMに構造化出力させる**ための形。LLMのプロンプト都合で決まる
- `RootCause`（domain）… アプリ内部で扱う**ビジネス上の本質**。`is_high_confidence()`のような判断ロジックを持つ
- `RCAResponse`（schema）… **APIで外に返す**ための形。クライアントとの契約で決まる

`services`層がこれらを変換する（LLM出力`RCASummary` → ドメイン`RootCause` → API応答`RCAResponse`）。

#### なぜ分けるのか

- **外の変更を内部に波及させない**: API仕様やLLMの出力形式が変わっても、`schemas`を直すだけで済み、`domain`とビジネスロジックは無傷
- **ビジネスロジックの置き場所が明確になる**: 判断や計算は`domain`のメソッドに集約され、`schemas`は単なるデータの入れ物に保てる
- **テストしやすい**: `domain`はPydanticにもフレームワークにも依存しないので、純粋な単体テストが書ける

> [!TIP]
> 小規模なアプリなら、`domain`を省いて`schemas`（Pydantic）だけで済ませることも多い。**ビジネスロジックが薄いうち**はそれで十分で、ロジックが育って「APIの形」と「内部の本質」を分けたくなった段階で`domain`を導入すればよい。

### `infrastructure/` — 外部連携

外部API、DB、メッセージキュー等とのやり取りを担当。**実装の詳細を隠蔽** することが目的。

```python
# src/my_app/infrastructure/bedrock_client.py
from langchain_aws import ChatBedrockConverse
from my_app.config import get_settings


def create_bedrock_chat() -> ChatBedrockConverse:
    """LangChain用のBedrockクライアントを生成"""
    settings = get_settings()
    return ChatBedrockConverse(
        model=settings.bedrock_model_id,
        region_name=settings.aws_region,
        max_tokens=4096,
    )
```

> [!NOTE]
> **`ChatBedrockConverse` を使う理由**: LangChain公式はConverse APIベースの `ChatBedrockConverse` を推奨しており、従来の `ChatBedrock`（InvokeModel API）はカスタムモデルが必要な場合向け。引数名も `model_id` ではなく `model` になる。

> [!CAUTION]
> Claude 5系（Sonnet 5 / 5.5 など）は `temperature` / `top_p` / `top_k` を**デフォルト値以外で指定すると400エラー**になるため、`temperature: 0.0` は渡さない。また `global.` はクロスリージョン推論プロファイルのID（モデルID単体では呼べない）。Sonnet 5.5 には `apac.` / `jp.` プロファイルが無く、東京リージョンからも `global.` で呼べる。

```python
# src/my_app/infrastructure/prometheus_client.py
import httpx
from datetime import datetime
from my_app.config import get_settings


class PrometheusClient:
    def __init__(self, base_url: str | None = None) -> None:
        self.base_url = base_url or get_settings().prometheus_url
        self.client = httpx.AsyncClient(base_url=self.base_url, timeout=30.0)

    async def query_range(
        self,
        query: str,
        start: datetime,
        end: datetime,
        step: str = "30s",
    ) -> dict:
        """PromQLクエリをレンジで実行"""
        response = await self.client.get(
            "/api/v1/query_range",
            params={
                "query": query,
                "start": start.timestamp(),
                "end": end.timestamp(),
                "step": step,
            },
        )
        response.raise_for_status()
        return response.json()

    async def close(self) -> None:
        await self.client.aclose()
```

```python
# src/my_app/infrastructure/langfuse_client.py
from langfuse import Langfuse
from my_app.config import get_settings


def create_langfuse() -> Langfuse:
    settings = get_settings()
    return Langfuse(
        public_key=settings.langfuse_public_key,
        secret_key=settings.langfuse_secret_key,
        base_url=settings.langfuse_base_url,
    )
```

ポイント:

- 外部ライブラリ（boto3、httpx、langfuse）への依存はここに閉じ込める
- `services`や`agents`からは抽象化されたインターフェースだけ見せる

### `agents/` — LangGraphエージェント定義

エージェントのグラフ構造、ノード、ツールを置く。

#### `state.py` — グラフのState定義

グラフ全体で受け渡す状態を定義する。`MessagesState` を継承して `messages` を引き継ぎつつ、独自フィールド（`summary`・`iteration` など）を足す。

```python
# src/my_app/agents/state.py
from langgraph.graph import MessagesState
from my_app.schemas.agent_schemas import RCASummary


class RCAState(MessagesState):
    """messages（MessagesState由来）に加え、RCA固有の状態を持つ"""
    incident_id: str
    iteration: int          # ループ回数（should_continueの打ち切り判定に使う）
    summary: RCASummary | None
```

#### `tools.py` — Toolの定義

```python
# src/my_app/agents/tools.py
from datetime import datetime, timedelta
from langchain_core.tools import tool
from my_app.infrastructure.prometheus_client import PrometheusClient


@tool
async def query_metrics(promql: str, minutes_ago: int = 60) -> str:
    """Prometheusにメトリクスクエリを実行する。

    Args:
        promql: PromQLクエリ式
        minutes_ago: 何分前から取得するか（デフォルト60分）
    """
    client = PrometheusClient()
    try:
        end = datetime.now()
        start = end - timedelta(minutes=minutes_ago)
        result = await client.query_range(promql, start, end)
        return _format_metrics_result(result)
    finally:
        await client.close()


def _format_metrics_result(result: dict) -> str:
    # LLMに渡しやすい形式に整形
    ...


# search_logs も同様に @tool で定義する（ログ検索ツール。本書では省略）
```

#### `nodes.py` — グラフのノード

```python
# src/my_app/agents/nodes.py
from langchain_core.messages import SystemMessage
from my_app.infrastructure.bedrock_client import create_bedrock_chat
from my_app.agents.state import RCAState
from my_app.agents.tools import query_metrics, search_logs
from my_app.schemas.agent_schemas import RCASummary


def planner_node(state: RCAState) -> dict:
    """次に取るべきアクションを計画する"""
    llm = create_bedrock_chat().bind_tools([query_metrics, search_logs])
    system = SystemMessage(content=(
        "あなたはSREエンジニアです。インシデントの根本原因を特定するため、"
        "必要なメトリクスやログを順次調査してください。"
    ))
    response = llm.invoke([system, *state["messages"]])
    # iterationを進める（should_continueの打ち切り判定が機能するように）
    return {"messages": [response], "iteration": state.get("iteration", 0) + 1}


def summarizer_node(state: RCAState) -> dict:
    """調査結果をまとめる"""
    llm = create_bedrock_chat().with_structured_output(RCASummary)
    summary = llm.invoke(state["messages"])
    return {"summary": summary}
```

#### `graph.py` — グラフの組み立て

```python
# src/my_app/agents/graph.py
from langchain_core.messages import AIMessage
from langgraph.graph import StateGraph, START, END
from langgraph.prebuilt import ToolNode
from my_app.agents.nodes import planner_node, summarizer_node
from my_app.agents.tools import query_metrics, search_logs
from my_app.agents.state import RCAState


def should_continue(state: RCAState) -> str:
    """ループ継続判定"""
    # 反復上限に達したら、ツール呼び出しが残っていても打ち切ってまとめへ（無限ループ防止）
    if state.get("iteration", 0) >= 10:
        return "summarize"
    # ツール呼び出しが残っていればツール実行へ、なければまとめへ
    last_message = state["messages"][-1]
    # messages の要素は AIMessage 以外も取り得るため、isinstance で絞ってから tool_calls を見る（mypy strict 対応）
    if isinstance(last_message, AIMessage) and last_message.tool_calls:
        return "tools"
    return "summarize"


def build_rca_graph():
    workflow = StateGraph(RCAState)

    workflow.add_node("planner", planner_node)
    workflow.add_node("tools", ToolNode([query_metrics, search_logs]))
    workflow.add_node("summarize", summarizer_node)

    workflow.add_edge(START, "planner")
    workflow.add_conditional_edges("planner", should_continue)
    workflow.add_edge("tools", "planner")
    workflow.add_edge("summarize", END)

    return workflow.compile()
```

### `services/` — ビジネスロジック

ユースケースを実装する層。**APIとエージェントの橋渡し** をする。

```python
# src/my_app/services/rca_service.py
from datetime import datetime
from my_app.agents.graph import build_rca_graph
from my_app.domain.models import RootCause
from my_app.schemas.api_schemas import RCARequest, RCAResponse


class RCAService:
    def __init__(self) -> None:
        self.graph = build_rca_graph()

    async def analyze(self, request: RCARequest) -> RCAResponse:
        # 1. リクエストをエージェント入力に変換
        initial_message = self._build_initial_prompt(request)

        # 2. エージェント実行
        result = await self.graph.ainvoke({
            "messages": [initial_message],
            "incident_id": request.incident_id,
        })

        # 3. ドメインモデルへ変換
        summary = result["summary"]
        root_cause = RootCause(
            incident_id=request.incident_id,
            summary=summary.root_cause,
            contributing_factors=summary.evidence,
            confidence_score=summary.confidence,
        )

        # 4. APIレスポンス形式に変換
        return RCAResponse(
            incident_id=root_cause.incident_id,
            summary=root_cause.summary,
            contributing_factors=root_cause.contributing_factors,
            confidence_score=root_cause.confidence_score,
            analyzed_at=datetime.now(),
        )

    def _build_initial_prompt(self, request: RCARequest) -> str:
        services = ", ".join(request.services) or "全サービス"
        return (
            f"インシデント {request.incident_id} の根本原因を分析してください。"
            f"対象サービス: {services}、"
            f"調査範囲: 過去{request.time_range_minutes}分。"
        )
```

ポイント:

- `services`は **ユースケース単位** で分ける（`rca_service.py`、`alert_service.py`等）
- API層とエージェント層を疎結合にする変換役
- ドメインモデル（`RootCause`）を経由することで、エージェント出力の変更がAPIに直接波及しないようにする

### `api/` — ルーター層（FastAPI）

HTTPエンドポイントの定義。**ロジックは持たず、`services`を呼ぶだけ**。

#### `deps.py` — 依存性注入

```python
# src/my_app/api/deps.py
from functools import lru_cache
from my_app.services.rca_service import RCAService


@lru_cache
def get_rca_service() -> RCAService:
    return RCAService()
```

#### `routes.py` — エンドポイント

```python
# src/my_app/api/routes.py
from fastapi import APIRouter, Depends, HTTPException
from my_app.api.deps import get_rca_service
from my_app.schemas.api_schemas import RCARequest, RCAResponse
from my_app.services.rca_service import RCAService

router = APIRouter(prefix="/api/v1")


@router.post("/rca/analyze", response_model=RCAResponse)
async def analyze_incident(
    request: RCARequest,
    service: RCAService = Depends(get_rca_service),
) -> RCAResponse:
    try:
        return await service.analyze(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)) from e


@router.get("/health")
async def health() -> dict:
    return {"status": "ok"}
```

### `main.py` — エントリポイント

アプリケーションの起動。

```python
# src/my_app/main.py
from contextlib import asynccontextmanager
from fastapi import FastAPI
from my_app.api.routes import router
from my_app.config import get_settings


@asynccontextmanager
async def lifespan(app: FastAPI):
    # 起動時の処理
    settings = get_settings()
    print(f"Starting app in {settings.aws_region}")
    yield
    # 終了時の処理


app = FastAPI(
    title="RCA Agent API",
    version="0.1.0",
    lifespan=lifespan,
)
app.include_router(router)


if __name__ == "__main__":
    # ローカル開発用の起動（reload=Trueはホットリロード）。
    # 本番はDockerのCMDでuvicornを直接起動し、reloadは付けない。
    import uvicorn
    uvicorn.run("my_app.main:app", host="0.0.0.0", port=8000, reload=True)
```

### レイヤー設計の効果

この分け方により以下が実現できる。

1. **テストしやすさ** — `services`をテストする際、`infrastructure`をモックすればLLMやPrometheusに繋がずテスト可能
2. **変更の局所化** — Bedrockから別のLLMに切り替えても、影響は`infrastructure`内に閉じる
3. **再利用性** — 同じ`RCAService`をHTTP API経由でもCLI経由でもSlack Bot経由でも呼べる
4. **責務の明確化** — レビュー時に「このコードはここに置くべきか？」の判断がしやすい

## ベストプラクティスのまとめ

新規プロジェクトを始めるときの推奨フロー:

1. `uv init --package my_project`（アプリ）/ `uv init --lib my_project`（配布ライブラリ）でsrc layoutで作成
2. `uv python pin 3.12` でPythonバージョン固定
3. `pyproject.toml`にruff/mypy/pytestの設定を集約
4. `uv add --dev pytest ruff mypy pre-commit` で開発ツール導入
5. `pre-commit install` でコミット時の自動チェック設定
6. `.env.example`、`.gitignore`、`README.md`を整備
7. `uv.lock`は **必ずコミット** する

これで再現性のある開発環境と高速なツールチェーンが整う。アプリケーション系（FastAPI、LangGraphなど）の場合は、上記の「アプリケーション系」のレイヤー構成を採用することで、ドメインロジックと外部依存を綺麗に分離できる。