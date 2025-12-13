## テストファイル名、テスト関数名、テストClass名の命名規則
- テストファイル名とテスト関数名は`test_` で始まる必要がある
  - ファイル名は`*_test.py`も可能だが、`test_*.py`が一般的
- 例: `test_example.py`, `def test_function():`
- テストClass名は`Test`で始まる必要がある
  - 例: `class TestExample:`
- `tests`フォルダ内の`test_*.py`ファイルも自動的にテスト対象になる

## `pytest.ini`ファイル
- pytestのデフォルトの振る舞いを変更できるメインの設定ファイル。
- このファイルが保存されているディレクトリがpytestのルートディレクトリになる。

## アサーション（assertion）
- アサーションは、テストが期待通りの結果を得ているかを確認するためのもの
- 例: `assert foo() == 42`
- `assert`はpythonの組み込みキーワードであり、`assert`の後に続く式が`False`の場合、`AssertionError`を発生させる

## 想定される例外のテスト
- `pytest.raises()`を使って、特定の例外が発生することをテストできる
- 例（`ZeroDivisionError`が発生することをテスト）  
  ```python
  import pytest

  def test_zero_division():
      with pytest.raises(ZeroDivisionError):
          1 / 0
  ```
  - **`with`ブロック内で`ZeroDivisionError`が発生しない場合、テストは失敗する**
  - **`with`ブロック内で`ZeroDivisionError`以外の例外が発生した場合、テストは失敗する**

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
- pytest の fixture は、テスト関数の引数に指定することで実行され、その戻り値（または `yield` の値）がテスト関数の引数に渡される。戻り値がない場合は副作用のみ利用できる。
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

### `params`を使ったパラメータ化
- フィクスチャに`params`引数を渡すことで、同じフィクスチャを使って複数のパラメータでテストを実行できる
- `request`は固定の引数名で、`request.param`で現在のパラメータを取得できる
- 一般的な例  
  ```python
  import pytest

  @pytest.fixture(params=[
      (2, 3, 5),
      (0, 0, 0),
      (-1, 1, 0),
  ])
  def add_test_data(request):
      return request.param  # 各パラメータが順番に渡される

  def test_add(add_test_data):
      a, b, expected = add_test_data
      assert add(a, b) == expected
  ```
- idも付けられる  
  ```python
  import pytest

  @pytest.fixture(params=[
      pytest.param({"a": 2, "b": 3, "expected": 5}, id="positive"),
      pytest.param({"a": -1, "b": 1, "expected": 0}, id="negative"),
      pytest.param({"a": 0, "b": 0, "expected": 0}, id="zero"),
  ])
  def add_test_data(request):
      return request.param

  def test_add(add_test_data):
      assert add(add_test_data["a"], add_test_data["b"]) == add_test_data["expected"]
  ```


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
- `yield`の前が前処理（セットアップ）、後が後処理（ティアダウン）になる
- `yield`の戻り値がテスト関数に渡される

> [!IMPORTANT]  
> フィクスチャ関数に`yield`が含まれている場合、`yield`のタイミングでフィクスチャ関数が停止され、テスト関数が実行される。そして、テスト関数の実行が終わると、フィクスチャ関数が再開されて`yield`の後のコードが実行される
> また、`yield`の後のコードは、テスト関数が例外を発生させても必ず実行される

- 例１  
  ```python
  import pytest

  @pytest.fixture
  def resource():
      print("セットアップ")
      res = {"connection": "DB接続"}
      yield res   # ← ここで res がテスト関数に渡される
      print("ティアダウン")  # テスト関数の後に必ず実行される

  # テスト関数
  def test_resource(resource):
      print("テスト実行中")
      assert resource["connection"] == "DB接続"
  ```
- 例２
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

### `monkeypatch`フィクスチャ
- 本番コードを直接変更する必要なく、テストの中で**一時的に**特定の関数・変数・環境変数などを書き換えることができるfixture
  - テストが終わると、自動的に元の状態に戻る
#### サポートされるメソッド
- `setattr`: オブジェクトの属性を一時的に変更
- `delattr`: オブジェクトの属性を一時的に削除
- `setitem`: 辞書のキーと値を一時的に変更
- `delitem`: 辞書のキーと値を一時的に削除
- `setenv`: 環境変数を一時的に設定
- `delenv`: 環境変数を一時的に削除
- `syspath_prepend`: `sys.path`にパスを一時的に追加
- `chdir`: カレントディレクトリを一時的に変更
#### 例１: 関数の差し替え（外部APIを呼び出す関数をモックする）
- `app.py`（本番コード）  
  ```python
  import requests

  def get_user_name(user_id):
      response = requests.get(f"https://api.example.com/users/{user_id}")
      return response.json()["name"]
  ```
- `test_app.py`（テストコード）  
  ```python
  from app import get_user_name

  def test_get_user_name(monkeypatch):
      # モック用のレスポンスクラス
      class MockResponse:
          def json(self):
              return {"name": "テスト太郎"}

      # requests.getを差し替え
      def mock_get(url):
          return MockResponse()

      monkeypatch.setattr("requests.get", mock_get)

      result = get_user_name(123)
      assert result == "テスト太郎"
  ```
#### 例２: 環境変数の差し替え
- `config.py`（本番コード）  
  ```python
  import os

  def get_database_url():
      return os.getenv("DATABASE_URL", "sqlite:///default.db")
  ```
- `test_config.py`（テストコード）  
  ```python
  from config import get_database_url

  def test_database_url_from_env(monkeypatch):
      monkeypatch.setenv("DATABASE_URL", "postgresql://localhost/testdb")
      
      assert get_database_url() == "postgresql://localhost/testdb"

  def test_database_url_default(monkeypatch):
      monkeypatch.delenv("DATABASE_URL", raising=False)
      
      assert get_database_url() == "sqlite:///default.db"
  ```

> [!NOTE]
> - `raising=False`を指定すると、環境変数が存在しない場合でも例外が発生しない
#### 例３: クラスメソッドの差し替え
- `service.py`（本番コード）  
  ```python
  class PaymentService:
      def charge(self, amount):
          # 実際には外部決済APIを呼ぶ
          return {"status": "charged", "amount": amount}

  def process_order(amount):
      service = PaymentService()
      return service.charge(amount)
  ```
- `test_service.py`（テストコード）  
  ```python
  from service import PaymentService, process_order

  def test_process_order(monkeypatch):
      def mock_charge(self, amount):
          return {"status": "mocked", "amount": amount}

      monkeypatch.setattr(PaymentService, "charge", mock_charge)

      result = process_order(1000)
      assert result["status"] == "mocked"
  ```

### その他fixtureのオプション
#### `autouse`オプション
- `autouse=True`を指定すると、明示的にテスト関数の引数に指定しなくても自動的に実行される（デフォルトは`False`）
  - `autouse: False`の場合  
    ```python
    import pytest

    @pytest.fixture
    def setup_env():
        print("\n--- setup_env fixture 実行 ---")

    def test_one(setup_env):
        print("test_one 実行")

    def test_two():
        print("test_two 実行")
    
    # 実行結果
    ## --- setup_env fixture 実行 ---
    ## test_one 実行
    ## test_two 実行
    ```
  - `autouse: True`の場合  
    ```python
    import pytest

    @pytest.fixture(autouse=True)
    def setup_env():
        print("\n--- setup_env fixture 実行 ---")

    def test_one():
        print("test_one 実行")

    def test_two():
        print("test_two 実行")

    # 実行結果
    ## --- setup_env fixture 実行 ---
    ## test_one 実行
    ## --- setup_env fixture 実行 ---
    ## test_two 実行
    ```

## パラメータ化
- 同じテスト関数を異なる引数（複数のテストケース）で繰り返し実行したい場合に使う仕組み
- ３つの方法がある
  1. `@pytest.mark.parametrize`デコレータを使う方法
  2. フィクスチャでパラメータ化する方法
  3. `pytest_generate_tests`フック関数を使う方法

### `@pytest.mark.parametrize`デコレータを使う方法
- 最も一般的な方法
- 複数の引数セットを指定して、同じテスト関数を繰り返し実行できる
  - 引数セットごとにテスト関数が実行される
- **セットアップ/ティアダウンが不要な単純なテストに適している**
- 例  
  ```python
  import pytest

  @pytest.mark.parametrize("input,expected", [
      (1, 2),
      (2, 3),
      (3, 4),
  ])
  def test_increment(input, expected):
      assert input + 1 == expected
  ```

### フィクスチャでパラメータ化する方法
- フィクスチャに`params`引数を指定してパラメータ化できる
- 各パラメータセットごとにフィクスチャが実行され、テスト関数に渡される
- セットアップ/ティアダウンが必要な場合に適している
- `request.param`は現在のパラメータセットの値を表す
- 例  
  ```python
  import pytest

  @pytest.fixture(params=["mysql", "postgresql", "sqlite"])
  def database(request):
      db = connect_to_db(request.param)
      yield db
      db.close()

  def test_query(database):
      result = database.execute("SELECT 1")
      assert result is not None
  ```

### `pytest_generate_tests`フック関数を使う方法
- より高度なパラメータ化が必要な場合に使う
- テスト収集フェーズで動的にパラメータを生成
- 外部ファイルからパラメータを読み込んだり、条件に応じてパラメータを変えたい場合に有効
- `metafunc`オブジェクトに様々な情報が含まれていて、`metafunc`オブジェクトを使って、テスト関数の引数名やパラメータセットを操作できる
  - `metafunc.fixturenames`: テスト関数の引数名のリスト
  - `metafunc.parametrize(argnames, argvalues)`: 引数名と引数値のリストを指定してパラメータ化する
- 例  
  ```python
  import pytest

  def pytest_generate_tests(metafunc):
      if "input" in metafunc.fixturenames:
          test_data = [(1, 2), (2, 3), (3, 4)]
          metafunc.parametrize("input,expected", test_data)

  def test_increment(input, expected):
      assert input + 1 == expected
  ```

## マーカー（marker）
- 特定のテストが何か特別であることをpytestに伝えるためのラベル/タグのようなもの

## テスト関数の構造化
- Arrange-Act-Assert（Given-When-Then）パターンを使うと、テスト関数の構造が明確になる
  1. **Arrange（Given）（準備）**: テストに必要なデータや状態を準備する
  2. **Act（When）（実行）**: テスト対象のコードを実行する
  3. **Assert（Then）（検証）**: 結果が期待通りかどうかを検証する