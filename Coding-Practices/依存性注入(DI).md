## 依存性注入（Dependency Injection, DI）とは
- **https://qiita.com/okazuki/items/a0f2fb0a63ca88340ff6**
- SOLID原則の一つである「依存関係逆転の原則（Dependency Inversion Principle, DIP）」を実現するための設計パターン
- 依存性注入（DI）は、クラスやモジュールが必要とする依存オブジェクトを自分で生成するのではなく、外部から提供（注入）することを指す
  - これにより、クラスやモジュールは具体的な実装に依存せず、抽象（インターフェースや抽象クラス）に依存するようになる
  - その結果、コードの柔軟性、再利用性、テスト容易性が向上する

### 依存性注入（DI）の種類
- 依存性注入（DI）には大きく分けて3種類ある
- **Goでは「コンストラクタ注入（Constructor Injection）」が一般的に使われて、Pythonでは３つとも使われることがある**

#### 1. コンストラクタ注入（Constructor Injection）
- 依存オブジェクトを コンストラクタ（生成時の引数）で渡す方法
- **Goの場合は `NewXxx(...)` のようなコンストラクタ関数を定義して、そこで依存を渡すのが一般的**
- Goの例  
  ```go
  type Repository interface {
      Find(id int) string
  }

  type Service struct {
      repo Repository
  }

  // コンストラクタ注入
  func NewService(r Repository) *Service {
      return &Service{repo: r}
  }
  ```
#### 2. セッター注入（Setter Injection）
- 依存をインスタンス生成後にメソッドでセットする方法
- Pythonの例  
  ```python
  class Service:
      def __init__(self):
          self.repo = None

      # セッター注入
      def set_repository(self, repo):
          self.repo = repo

  service = Service()
  service.set_repository(SomeRepository())  # セッター注入
  ```

#### 3. フィールド注入（Field/Property Injection）
- 依存を直接フィールド（メンバ変数）に代入する方法
- Pythonではシンプルに `obj.repo = repo` のように書ける
- Pythonの例  
  ```python
  class Service:
      def __init__(self):
          self.repo = None

  service = Service()
  service.repo = SomeRepository()  # フィールド注入
  ```