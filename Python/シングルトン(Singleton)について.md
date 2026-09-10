# シングルトン（Singleton）について
- シングルトンとは、あるクラスのインスタンスがアプリケーション全体で**常に1つしか存在しないこと**を保証するデザインパターン
  - 設定情報の管理、DBコネクションプール、ロガー、キャッシュなど「共有された単一の状態」を持たせたい場合に使われる
- Pythonでは言語仕様として複数の実装方法があり、それぞれ挙動やスレッドセーフ性、テストのしやすさが異なる

## 1. モジュールレベルでの実装（Pythonで最も一般的・推奨される方法）
- Pythonの**モジュールはimport時に1度しか実行されず、2回目以降のimportはキャッシュ（`sys.modules`）が使われる**ため、モジュール自体が自然にシングルトンとして機能する
- クラスや`__new__`のオーバーライドが不要でシンプル・可読性が高いため、Pythonicな方法としてよく推奨される
- 例
  ```python
  # config.py
  class _Config:
      def __init__(self):
          self.value = "default"

  # モジュールがimportされた時点でインスタンスが1つ生成される
  config = _Config()
  ```
  ```python
  # main.py
  from config import config

  config.value = "changed"
  ```
  ```python
  # other_module.py
  from config import config

  print(config.value)  # ★"changed" が出力される（同じインスタンスを参照しているため）
  ```

## 2. `__new__`をオーバーライドする方法
- クラスに対して明示的にシングルトンであることを示したい場合によく使われる古典的な実装
- `__new__`は「インスタンスを生成する」メソッド（`__init__`は「生成されたインスタンスを初期化する」メソッド）であることを利用し、既にインスタンスが存在する場合はそれを返す
  ```python
  class Singleton:
      _instance = None

      def __new__(cls, *args, **kwargs):
          if cls._instance is None:
              cls._instance = super().__new__(cls)
          return cls._instance

      def __init__(self, value=None):
          # ★注意: __init__はインスタンス取得のたびに毎回呼ばれる
          #   （__new__で既存インスタンスを返しても__init__は再実行される）
          if not hasattr(self, "_initialized"):
              self.value = value
              self._initialized = True

  a = Singleton("first")
  b = Singleton("second")

  print(a is b)      # True（同一インスタンス）
  print(a.value)      # "first"（__init__の再実行をガードしているため上書きされない）
  ```
> [!CAUTION]
> `__init__`は`__new__`が既存インスタンスを返した場合でも呼ばれる（上記例のように`_initialized`フラグ等でガードしないと、生成のたびに初期化処理が走ってしまう）

## 3. デコレータで実装する方法
- クラスをデコレータでラップし、インスタンス生成をキャッシュする方法
  ```python
  from functools import wraps

  def singleton(cls):
      instances = {}

      @wraps(cls)
      def get_instance(*args, **kwargs):
          if cls not in instances:
              instances[cls] = cls(*args, **kwargs)
          return instances[cls]

      return get_instance

  @singleton
  class Database:
      def __init__(self):
          print("DBコネクションを初期化")

  db1 = Database()
  db2 = Database()
  print(db1 is db2)  # True
  ```
> [!CAUTION]
> デコレータ適用後の`Database`は「関数」（`get_instance`）になるため、`isinstance(db1, Database)`のようなクラスとしての型チェックができなくなる
> - `type(db1)`は`get_instance`ではなく元の`Database`クラスになる点は問題ないが、`Database`自体を型として参照するコードとは相性が悪い

## 4. メタクラス（metaclass）を使う方法
- クラスの「生成のされ方」自体をカスタマイズすることでシングルトンを実現する方法
- 複数のクラスにシングルトンの性質を持たせたい場合、継承だけで再利用できるのが利点
  ```python
  class SingletonMeta(type):
      _instances = {}

      def __call__(cls, *args, **kwargs):
          if cls not in cls._instances:
              cls._instances[cls] = super().__call__(*args, **kwargs)
          return cls._instances[cls]

  class Logger(metaclass=SingletonMeta):
      def __init__(self):
          print("Loggerを初期化")

  l1 = Logger()
  l2 = Logger()
  print(l1 is l2)  # True
  ```
- `isinstance`によるチェックも問題なく機能するため、デコレータ方式より型安全
- 継承関係やメタクラスの競合（他のメタクラスと併用する場合）に注意が必要

## 5. `@lru_cache`を使う方法
- 引数なし（または同じ引数）で呼び出す関数をキャッシュする`functools.lru_cache`（もしくは`functools.cache`）を利用して、簡易的にシングルトンを実現する方法
- FastAPIの依存性注入（`Depends`）でシングルトン的にインスタンスを共有したい場合などによく使われる
  ```python
  from functools import lru_cache

  class Settings:
      def __init__(self):
          print("設定を読み込み")
          self.value = "default"

  @lru_cache
  def get_settings() -> Settings:
      return Settings()

  s1 = get_settings()
  s2 = get_settings()
  print(s1 is s2)  # True（2回目以降はキャッシュされた同一インスタンスが返る）
  ```
> [!CAUTION]
> - 引数が異なると別のキャッシュエントリになる（＝厳密には「関数の呼び出し結果のキャッシュ」であり、クラスそのものをシングルトン化しているわけではない）
> - テスト時にキャッシュをクリアしたい場合は`get_settings.cache_clear()`を呼ぶ必要がある

## 6. Borg（Monostate）パターン
- 「インスタンスは複数存在してよいが、状態（属性）だけを共有する」という発想の変則的なシングルトン
- 同一性（`is`で比較して同じオブジェクトか）は保証しないが、状態の共有という目的だけを達成したい場合に使われる
  ```python
  class Borg:
      _shared_state = {}

      def __init__(self):
          self.__dict__ = self._shared_state

  a = Borg()
  b = Borg()

  a.value = "shared"
  print(b.value)   # ★"shared" が出力される（状態は共有されている）
  print(a is b)     # False（インスタンス自体は別物）
  ```

## スレッドセーフについての注意
> [!CAUTION]
> 上記の`__new__`・デコレータ・メタクラスによる実装は、マルチスレッド環境では複数のスレッドが同時にインスタンス生成条件（`if cls._instance is None`等）を通過してしまい、複数インスタンスが生成されるリスクがある

- スレッドセーフにするには`threading.Lock`を使ったダブルチェックロッキングが必要
  ```python
  import threading

  class SingletonMeta(type):
      _instances = {}
      _lock = threading.Lock()

      def __call__(cls, *args, **kwargs):
          if cls not in cls._instances:
              with cls._lock:
                  # ロック取得後に再度チェック（ダブルチェックロッキング）
                  if cls not in cls._instances:
                      cls._instances[cls] = super().__call__(*args, **kwargs)
          return cls._instances[cls]
  ```
- モジュールレベルの実装（1.）はimport機構自体が排他制御されているため、この問題が発生しにくい

## テストにおける注意点
> [!CAUTION]
> シングルトンは「グローバルな状態」を持つため、テスト間で状態が残ってしまいテストの独立性が壊れやすい
> - 例: あるテストでシングルトンの属性を変更すると、後続のテストにその変更が影響してしまう

- 対策
  - `pytest`のfixtureで各テスト後にシングルトンのインスタンス（`_instance`やキャッシュ辞書）をリセットする
  - `@lru_cache`を使っている場合は`.cache_clear()`をteardownで呼ぶ
  - そもそもシングルトンに依存しすぎず、依存性注入（DI）でテスト時に差し替え可能にしておく

## まとめ・使い分けの指針
| 方法 | 特徴 | 向いているケース |
| --- | --- | --- |
| モジュールレベル | 最もシンプル・Pythonic | 通常はこれで十分 |
| `__new__`オーバーライド | クラスとして明示的にシングルトンを表現 | OOP的な設計を重視したい場合 |
| デコレータ | 既存クラスに後付けしやすい | 型チェック（`isinstance`）を使わない場合 |
| メタクラス | 継承で複数クラスに適用可能・型安全 | 複数クラスをシングルトン化したい場合 |
| `@lru_cache` | 関数ベースで手軽 | FastAPIの`Depends`など依存性注入と組み合わせる場合 |
| Borg | 同一性は不要、状態共有のみでよい | 複数インスタンスを許容しつつ状態だけ共有したい場合 |

- ⭐️ 一般的にPythonではJavaやC++のような厳密なシングルトン実装は過剰とされることが多く、**「モジュールレベルの実装」や「`@lru_cache`を使った関数ベースの実装」で十分なケースが多い**
- グローバルな状態を持つこと自体がテスタビリティや依存関係の見通しを悪くするため、本当にシングルトンが必要かどうか（DIで代替できないか）を検討することも重要
