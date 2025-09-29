## テストファイル名、テスト関数名、テストClass名の命名規則
- テストファイル名とテスト関数名は`test_` で始まる必要がある
  - ファイル名は`*_test.py`も可能だが、`test_*.py`が一般的
- 例: `test_example.py`, `def test_function():`
- テストClass名は`Test`で始まる必要がある
  - 例: `class TestExample:`
- `tests`フォルダ内の`test_*.py`ファイルも自動的にテスト対象になる

## `pytest.ini`ファイル
- プロジェクトルートに`pytest.ini`ファイルを作成して、pytestの設定を行うことができる

## アサーション（assertion）
- アサーションは、テストが期待通りの結果を得ているかを確認するためのもの
- 例: `assert foo() == 42`
- `assert`はpythonの組み込みキーワードであり、`assert`の後に続く式が`False`の場合、`AssertionError`を発生させる

## テストの実行
- ターミナルで`pytest`コマンドを実行する
  - 例: `pytest test_example.py`
- `-v`オプションを付けると、詳細な出力が得られる
  - 例: `pytest -v test_example.py`
- ファイル名を指定しない場合、カレントディレクトリとサブディレクトリ配下のすべてのテストファイル（`test_*.py`または`*_test.py`）が実行される
- ファイルを複数指定したり、ディレクトリを指定することも可能
- 特定のテスト関数だけを実行したい場合、`::`で区切って関数名を指定する
  - 例: `pytest test_example.py::test_function`

### Traceback（トレースバック）
- テストが失敗した場合、失敗した場所とその周辺のコードが表示されて、これを**Traceback（トレースバック）**という
- `--tb=no`オプションを付けると、Tracebackが表示されない
  - 例: `pytest --tb=no test_example.py`
- 例（`test_failing`の部分）  
  ```shell
  ================================================= test session starts ==================================================
  platform linux -- Python 3.12.3, pytest-8.4.2, pluggy-1.6.0
  rootdir: /home/nutslove/pytest_book_code
  configfile: pytest.ini
  plugins: langsmith-0.4.4, anyio-4.9.0
  collected 1 item

  test_two.py F                                                                                                    [100%]

  ======================================================= FAILURES =======================================================
  _____________________________________________________ test_failing _____________________________________________________

      def test_failing():
  >       assert (1, 2, 3) == (3, 2, 1)
  E       assert (1, 2, 3) == (3, 2, 1)
  E
  E         At index 0 diff: 1 != 3
  E         Use -v to get more diff

  test_two.py:2: AssertionError
  =============================================== short test summary info ================================================
  FAILED test_two.py::test_failing - assert (1, 2, 3) == (3, 2, 1)
  ================================================== 1 failed in 0.03s ===================================================
  ```

## テスト結果
- 以下のような記号でテスト結果が表示される
  - `.`（PASSED）: テスト成功
  - `F`（FAILED）: テスト失敗
  - `E`（ERROR）: 例外がテスト関数の外（e.g. フィクスチャ、フック関数）で発生したことを意味
  - `s`（SKIPPED）: テストがスキップされたことを意味
  - `x`（XFAIL）: 失敗するはずのテストが、想定通りに失敗したことを意味（xfail）
  - `X`（XPASS）: xfailマーカーが付いたテストが想定に反して成功したことを意味（xpass）

## フィクスチャ（fixture）
- テストで繰り返し使う「事前準備」「後片付け」「共通データの提供」を行う仕組み
- 実際のテスト関数の実行に先立って（場合によってはそのあとに）pytestが実行する関数
  - 例: テストで使うデータセットの取得、DBのセットアップ、モックサーバーの起動など
- フィクスチャは`@pytest.fixture`デコレータを使って定義する
- テスト関数の引数にフィクスチャ名を指定すると、そのフィクスチャが実行され、その戻り値が引数として渡される
- 例  
  ```python
  import pytest

  # fixtureの定義
  @pytest.fixture
  def sample_data():
      return [1, 2, 3, 4, 5]

  # fixtureを使うテスト
  def test_sum(sample_data):
      assert sum(sample_data) == 15

  def test_length(sample_data):
      assert len(sample_data) == 5
  ```

> [!NOTE]  
> テスト関数とフィクスチャ関数で、例外が発生した場合の挙動は異なる
> - テスト関数で例外が発生した場合は、**FAILED**になる
> - フィクスチャ関数で例外が発生した場合は、**ERROR**になる
> なので、テストが失敗した場合に、フィクスチャ関数で例外が発生しているのか、テスト関数で例外が発生しているのかを区別することができる

### `conftest.py`ファイル
- フィクスチャをプロジェクト全体で共有したい場合、`conftest.py`ファイルにフィクスチャを定義する
- `conftest.py`ファイルにフィクスチャを定義すると、そのディレクトリとサブディレクトリ配下のすべてのテストから利用可能になる
  - 例えば、以下のように`tests/conftest.py`にフィクスチャを定義すると、プロジェクトルートで`pytest`をすれば、`tests/`配下のすべてのテストファイルからそのフィクスチャを利用できる  
    ```bash
    my_project/
    ├── src/                # 本体コード
    │   ├── __init__.py
    │   └── my_module.py
    ├── tests/              # テストコード
    │   ├── __init__.py     # 空でOK（必須ではない）
    │   ├── test_my_module.py
    │   └── conftest.py     # fixture定義用
    ├── pyproject.toml or setup.cfg or pytest.ini
    └── requirements.txt
    ```

### 後処理を含むfixture
- `yield`を使って、セットアップ & ティアダウン（後片付け）を行うことができる
- `yield`の前が事前処理、後が後処理になる
- 例
  ```python
  import pytest

  @pytest.fixture
  def file_handle(tmp_path):
      # セットアップ
      file = tmp_path / "test.txt"
      f = open(file, "w+")
      yield f
      # ティアダウン（後片付け）
      f.close()

  # fixtureを使うテスト
  def test_file_write(file_handle):
      file_handle.write("hello")
      file_handle.seek(0)
      assert file_handle.read() == "hello"
  ```

### fixtureのスコープ
- fixtureがどの単位で実行されるかを制御できる
- スコープには以下の種類がある
  - `function`（デフォルト）: 各テスト関数ごとに実行される
  - `class`: 各テストクラスごとに1回実行される
  - `module`: 各テストモジュール（ファイル）ごとに1回実行される
  - `package`: 各パッケージ（ディレクトリ）ごとに1回実行される
  - `session`: テストセッション全体で1回だけ実行される
- 例: スコープを`module`に設定する場合  
  ```python
  @pytest.fixture(scope="module")
  def db_connection():
      conn = create_db_connection()
      yield conn
      conn.close()
  ```