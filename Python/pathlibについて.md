- `pathlib`は、ファイルシステムのパスを**オブジェクト指向**で扱うための標準ライブラリ（Python 3.4以降）。
- 従来の`os.path`（文字列ベース）と異なり、パスを`Path`オブジェクトとして扱うため、コードが直感的で読みやすくなる。
- OSの違い（Windowsの`\`、Unixの`/`）を意識せずにパス操作ができる。

## `os.path`との比較
- `os.path`は文字列ベースで関数を組み合わせて使うのに対し、`pathlib`はメソッドチェーンやオペレーターで簡潔に書ける。

| やりたいこと | `os.path`（従来） | `pathlib`（推奨） |
| --- | --- | --- |
| パスの結合 | `os.path.join(a, b)` | `Path(a) / b` |
| ファイル名取得 | `os.path.basename(p)` | `Path(p).name` |
| 拡張子取得 | `os.path.splitext(p)[1]` | `Path(p).suffix` |
| 親ディレクトリ | `os.path.dirname(p)` | `Path(p).parent` |
| 絶対パス化 | `os.path.abspath(p)` | `Path(p).resolve()` |
| 存在確認 | `os.path.exists(p)` | `Path(p).exists()` |
| ファイル読み込み | `open(p).read()` | `Path(p).read_text()` |

## `Path`オブジェクトの生成
```python
from pathlib import Path

# カレントディレクトリ
p = Path(".")

# 文字列からパスを生成
p = Path("/usr/local/bin")

# カレントディレクトリ・ホームディレクトリ
Path.cwd()    # 現在の作業ディレクトリ
Path.home()   # ホームディレクトリ（~）
```

## パスの結合（`/`演算子）
- `pathlib`の最大の特徴は、`/`演算子でパスを結合できること。
```python
from pathlib import Path

base = Path("/home/user")
config = base / "app" / "config.yaml"
print(config)   # /home/user/app/config.yaml

# 文字列とPathを混在させてもOK（最初がPathオブジェクトであればよい）
log_dir = Path("logs") / "2026" / "06"
```

## パスの各種属性（プロパティ）
- 例として`Path("/home/user/app/config.yaml")`を使う。

| 属性 | 説明 | 結果 |
| --- | --- | --- |
| `.name` | ファイル名（拡張子含む） | `config.yaml` |
| `.stem` | 拡張子を除いたファイル名 | `config` |
| `.suffix` | 拡張子 | `.yaml` |
| `.suffixes` | 複数拡張子のリスト | `['.yaml']`（`a.tar.gz`なら`['.tar', '.gz']`） |
| `.parent` | 親ディレクトリ | `/home/user/app` |
| `.parents` | 上位ディレクトリの一覧（イテレータ） | `/home/user/app`, `/home/user`, ... |
| `.parts` | パスを要素ごとに分解したタプル | `('/', 'home', 'user', 'app', 'config.yaml')` |
| `.anchor` | ルート部分 | `/` |

```python
from pathlib import Path

p = Path("/home/user/app/config.yaml")
print(p.name)     # config.yaml
print(p.stem)     # config
print(p.suffix)   # .yaml
print(p.parent)   # /home/user/app
```

## パスの変更
```python
from pathlib import Path

p = Path("/home/user/config.yaml")

# ファイル名を変更
p.with_name("settings.yaml")    # /home/user/settings.yaml

# 拡張子を変更
p.with_suffix(".json")          # /home/user/config.json

# stem（拡張子以外の名前）を変更（Python 3.9以降）
p.with_stem("backup")           # /home/user/backup.yaml
```

## 絶対パス・正規化
```python
from pathlib import Path

p = Path("../app/./config.yaml")

# 絶対パスに変換し、シンボリックリンクや「..」を解決
p.resolve()       # /home/user/app/config.yaml（実際のCWD依存）

# 絶対パスに変換（存在しなくてもOK、「..」は解決しない）
p.absolute()
```
- `resolve()`はファイルが存在しなくても動作する（Python 3.6以降）。シンボリックリンクの解決まで行うのが`absolute()`との違い。

## 存在確認・種別判定
```python
from pathlib import Path

p = Path("/home/user/config.yaml")

p.exists()      # 存在するか
p.is_file()     # ファイルか
p.is_dir()      # ディレクトリか
p.is_symlink()  # シンボリックリンクか
```

## ディレクトリの作成・削除
```python
from pathlib import Path

d = Path("logs/2026/06")

# ディレクトリ作成
# parents=True   : 中間ディレクトリも作成（mkdir -p 相当）
# exist_ok=True  : 既に存在していてもエラーにしない
d.mkdir(parents=True, exist_ok=True)

# 空ディレクトリの削除
d.rmdir()

# ファイルの削除
Path("temp.txt").unlink()              # 存在しないとエラー
Path("temp.txt").unlink(missing_ok=True)  # 存在しなくてもOK（Python 3.8以降）
```

## ファイルの読み書き
- `open()`を使わずに、簡潔にファイルの読み書きができる。
```python
from pathlib import Path

p = Path("config.yaml")

# 読み込み
text = p.read_text(encoding="utf-8")   # テキストとして読み込み
data = p.read_bytes()                  # バイナリとして読み込み

# 書き込み（上書き）
p.write_text("hello", encoding="utf-8")
p.write_bytes(b"\x00\x01")

# 従来通りopen()も使える（withと組み合わせ）
with p.open("r", encoding="utf-8") as f:
    for line in f:
        print(line)
```

## ファイル・ディレクトリの一覧取得
```python
from pathlib import Path

d = Path("/home/user/app")

# 直下の要素を列挙
for item in d.iterdir():
    print(item)

# パターンマッチ（直下のみ）
for py in d.glob("*.py"):
    print(py)

# 再帰的にパターンマッチ（サブディレクトリも含む）
for py in d.rglob("*.py"):
    print(py)

# rglob("*.py") は glob("**/*.py") と同じ
```
- `glob()` / `rglob()`はジェネレータを返すため、リスト化したい場合は`list(d.glob("*.py"))`とする。

## パスの相対・包含関係
```python
from pathlib import Path

p = Path("/home/user/app/config.yaml")

# あるディレクトリからの相対パスを取得
p.relative_to("/home/user")    # app/config.yaml

# パスが特定のパターンに一致するか
p.match("*.yaml")              # True
p.match("app/*.yaml")          # True
```

## ファイル情報の取得（`stat()`）
```python
from pathlib import Path

p = Path("config.yaml")
st = p.stat()

st.st_size      # ファイルサイズ（バイト）
st.st_mtime     # 最終更新時刻（UNIXタイムスタンプ）
```

## `PurePath`について
- `Path`は実際のファイルシステムにアクセスする（`exists()`や`read_text()`など）。
- `PurePath`はファイルシステムにアクセスせず、**パス文字列の操作のみ**を行うクラス。
  - 異なるOS向けのパスを扱いたい場合（例: Linux上でWindowsパスを操作）に使う。
  - `PurePosixPath`（Unix系）、`PureWindowsPath`（Windows）がある。
```python
from pathlib import PureWindowsPath

p = PureWindowsPath("C:/Users/test/file.txt")
print(p.name)    # file.txt
# 実ファイルの存在確認やI/Oはできない
```

## よく使うパターン
- **スクリプト自身のディレクトリを基準にパスを組み立てる**
  ```python
  from pathlib import Path

  # 実行中のファイルのあるディレクトリ
  BASE_DIR = Path(__file__).resolve().parent

  # そこからの相対パスでファイルを指定
  config_path = BASE_DIR / "config" / "settings.yaml"
  ```

### `Path(__file__).resolve().parent.parent` の分解
- `__file__`からファイルの位置を基準にディレクトリを特定する定番パターン。`scripts/ingest.py`のような階層の深いスクリプトから、プロジェクトルートを求めたいときに使う。
- 左から順に分解すると以下の通り（`scripts/ingest.py`を例とする）。

| 式 | 意味 | 結果（例） |
| --- | --- | --- |
| `__file__` | 現在実行中のスクリプトのパス。相対パスのこともある | `scripts/ingest.py` |
| `.resolve()` | シンボリックリンクを解決し、絶対パスに変換 | `/Users/nutslove/money-forward-test/scripts/ingest.py` |
| `.parent` | 1つ上のディレクトリ | `/Users/nutslove/money-forward-test/scripts` |
| `.parent`（2回目） | さらに1つ上 | `/Users/nutslove/money-forward-test` |

- つまり全体で「このスクリプトファイルから見て2つ上のディレクトリ（＝プロジェクトのルート）の絶対パス」を取得している。
- `scripts/ingest.py`が`scripts/`ディレクトリにあるので、`.parent`を2回呼ぶことでプロジェクトルートにたどり着ける。

**なぜこう書くのか**
- `.resolve()`で絶対パスにしておくと、スクリプトをどのディレクトリから実行しても常に同じ正しいパスを指す。
  - `python scripts/ingest.py`でも `cd scripts && python ingest.py`でも結果が変わらない。
  - これがカレントディレクトリ依存の`os.getcwd()`などより安全な理由。
- この後`BASE_DIR / "data"`のように書けば、実行場所に関係なく確実に`プロジェクトルート/data`を参照できる。
  ```python
  from pathlib import Path

  BASE_DIR = Path(__file__).resolve().parent.parent  # プロジェクトルート
  DATA_DIR = BASE_DIR / "data"                        # プロジェクトルート/data
  ```
- **文字列が必要な場面では`str()`で変換**
  ```python
  import subprocess
  from pathlib import Path

  p = Path("/usr/bin/python3")
  subprocess.run([str(p), "--version"])
  ```
  - 多くの標準ライブラリは`Path`オブジェクトを直接受け取れる（`os.fspath`プロトコル対応）が、対応していない関数には`str()`で渡す。

## まとめ
- 新規コードでは`os.path`より`pathlib`を使うのが推奨される。
- `/`演算子によるパス結合、豊富なプロパティ、ファイルI/Oメソッドにより、可読性が高く保守しやすいコードが書ける。
- パス文字列の操作だけなら`PurePath`、実ファイルへのアクセスを伴うなら`Path`を使う。
