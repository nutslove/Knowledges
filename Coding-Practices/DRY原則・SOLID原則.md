## DRY原則とは
- Don't Repeat Yourselfの略で、「繰り返しを避ける」という意味
- 同じ機能やロジックを複数箇所で繰り返し実装することを避けて、共通の部分を一箇所にまとめることを推奨する原則
- コードの重複を減らし、再利用性を高めることで保守性や可読性が向上し、バグの発生を防ぐことができる

## SOLID原則とは
- オブジェクト指向プログラミングにおいて、**変更しやすく・理解しやすく・再利用しやすい**モジュール設計のための5つの原則
- 以下の5つの原則から構成される
  - 単一責任の原則（Single Responsibility Principle, SRP）
  - オープン・クローズドの原則（Open/Closed Principle, OCP）
  - リスコフの置換原則（Liskov Substitution Principle, LSP）
  - インターフェース分離の原則（Interface Segregation Principle, ISP）
  - 依存関係逆転の原則（Dependency Inversion Principle, DIP）

### 単一責任の原則（Single Responsibility Principle, SRP）
- 1つのクラス（モジュール）は1つの責任・機能のみを持つべき。クラス（モジュール）は「変更理由」が一つだけであるべき。

### オープン・クローズドの原則（Open/Closed Principle, OCP）
- ソフトウェア要素（クラスや関数）は「拡張に対しては開かれて」いなければならないが、「修正に対しては閉じられて」いるべき。つまり既存コードを変更せずに機能追加できる設計。

### リスコフの置換原則（Liskov Substitution Principle, LSP）
- サブタイプはスーパークラスの代わりに使える（置換してもプログラムの正当性が保たれる）ように設計すること。契約（前提・事後条件）を破らない。
- 例えば、`Bird`クラスを継承した`Penguin`クラスが`fly()`メソッドを持つと、`Bird`型の変数に`Penguin`を代入した際に問題が発生する。  
   ```python
   class Bird:
       def fly(self):
           pass

   class Penguin(Bird):
       def fly(self):
           raise NotImplementedError("ペンギンは飛べません")

   def make_bird_fly(bird: Bird):
       bird.fly()  # ここでエラーになる可能性がある

   penguin = Penguin()
   make_bird_fly(penguin)  # LSP違反
   ```

### インターフェース分離の原則（Interface Segregation Principle, ISP）
- クライアントはそれが使用しないメソッドに依存してはならない。大きなインターフェースを小さく分割し、特定のクライアントに必要なインターフェースのみを提供すること。
- 悪い例  
  ```python
  # bad_isp.py
  class Printer:
      def print(self, doc): ...
      def scan(self, doc): ...
      def fax(self, doc): ...
  # 複合機の全機能を期待するインターフェースだが、
  # 単純に印刷だけしたいクライアントが不必要なメソッドを実装する必要がある
  ```
- 改善（機能別に分割）  
  ```python
  # good_isp.py
  from abc import ABC, abstractmethod

  class PrinterInterface(ABC):
      @abstractmethod
      def print(self, doc): pass

  class ScannerInterface(ABC):
      @abstractmethod
      def scan(self, doc): pass

  class FaxInterface(ABC):
      @abstractmethod
      def fax(self, doc): pass

  class SimplePrinter(PrinterInterface):
      def print(self, doc):
          print("printing", doc)

  class MultiFunctionDevice(PrinterInterface, ScannerInterface, FaxInterface):
      def print(self, doc):
          print("printing", doc)
      def scan(self, doc):
          print("scanning", doc)
      def fax(self, doc):
          print("faxing", doc)
  ```

### 依存関係逆転の原則（Dependency Inversion Principle, DIP）
1. 上位モジュールは下位モジュールに依存してはならない。両者とも抽象に依存すべき
2. 抽象は実装の詳細に依存してはならない。実装の詳細が抽象に依存すべき
   - NG  
     ![DIP NG](./images/DIP_NG.jpg)
   - OK  
     ![DIP OK](./images/DIP_OK.jpg)

#### 依存性逆転の原則に違反している例
```python
# 下位モジュール（具体的な実装）
class MySQLDatabase:
    def save_user(self, user_data):
        print(f"MySQLにユーザーデータを保存: {user_data}")
        # MySQL固有の処理...
    
    def get_user(self, user_id):
        print(f"MySQLからユーザーID {user_id} を取得")
        return {"id": user_id, "name": "John Doe"}

class EmailService:
    def send_email(self, to, subject, body):
        print(f"メール送信: {to} - {subject}")
        # メール送信の具体的な処理...

# 上位モジュール（ビジネスロジック）
class UserService:
    def __init__(self):
        # ❌ DIP違反: 高レベルモジュールが低レベルモジュールに直接依存
        self.database = MySQLDatabase()  # 具象クラスに依存
        self.email_service = EmailService()  # 具象クラスに依存
    
    def create_user(self, user_data):
        # ユーザー作成のビジネスロジック
        self.database.save_user(user_data)
        self.email_service.send_email(
            user_data['email'], 
            "Welcome!", 
            "アカウントが作成されました"
        )
    
    def get_user(self, user_id):
        return self.database.get_user(user_id)

# 使用例
user_service = UserService()
user_service.create_user({
    "id": 1, 
    "name": "田中太郎", 
    "email": "tanaka@example.com"
})
user = user_service.get_user(1)
```
- **上記コードの問題点**
  - `UserService`（上位モジュール、ビジネスロジック層）が `MySQLDatabase`,`EmailService`（下位モジュール、具体的な永続化実装）に直接依存している。
  - 「抽象」ではなく「具体」に依存しているので、もしDBをMySQLからPostgreSQLやファイル保存に変えたい場合、`UserService`コード自体を修正しなければならない。
- **依存性逆転の原則に違反による影響**
  1. テストが困難
     - `UserService`をテストする際に、実際のMySQLデータベースやメールサービスが必要になる
     - モックやスタブを使ったユニットテストが困難（モックDBを注入できないため、ユニットテスト時に本物のDBが必要になる）
  2. 柔軟性の欠如
     - データベースをMySQLからPostgreSQLに変更したい場合、`UserService`のコードを変更する必要がある
     - メール送信方法を変更（例：SMTP → AWS SES）する際も同様
  3. 保守性の低下
     - 下位モジュールの変更が上位モジュールに伝播してしまい、修正範囲が大きくなる（変更に弱い設計）
  4. 再利用性の低下
     - UserServiceは特定の実装（MySQL、特定のEmailService）にしか使えない

- 上記の問題を解決するために、**依存性注入（Dependency Injection, DI）** を使って、`UserService`が依存する具体的な実装を外部から注入できるようにする（以下は上記コードをDI対応に書き換えた例）  
  1. 抽象化（インターフェース）の導入  
     ```python
     class DatabaseInterface(ABC):
         @abstractmethod
         def save_user(self, user_data: Dict[str, Any]) -> None:
             pass
     ```
       - `ABC`（Abstract Base Class）を使用してインターフェースを定義
       - 具象実装の詳細ではなく、契約（メソッドシグネチャ）を定義
  2. 依存性注入（Dependency Injection）
     ```python
     class UserService:
         def __init__(self, database: DatabaseInterface, email_service: EmailInterface):
             self.database = database        # 抽象化に依存
             self.email_service = email_service
     ```
     - コンストラクタで依存関係を外部から注入
     - 具象クラスを直接インスタンス化しない
  3. 抽象化への依存
     - `UserService`は`DatabaseInterface`と`EmailInterface`に依存し、具体的な実装(e.g. `MySQLDatabase`, `SMTPEmailService`)には依存しない
     - これにより、異なるDBやメールサービスの実装を簡単に差し替え可能

  - 全体コード  
    ```python
    # DIPに準拠した修正版コード
    from abc import ABC, abstractmethod
    from typing import Dict, Any, Optional

    # 抽象化（インターフェース）の定義
    class DatabaseInterface(ABC):
        """データベース操作の抽象インターフェース"""
        
        @abstractmethod
        def save_user(self, user_data: Dict[str, Any]) -> None:
            pass
        
        @abstractmethod
        def get_user(self, user_id: int) -> Optional[Dict[str, Any]]:
            pass

    class EmailInterface(ABC):
        """メール送信の抽象インターフェース"""
        
        @abstractmethod
        def send_email(self, to: str, subject: str, body: str) -> None:
            pass

    # 低レベルモジュール（具象実装）
    class MySQLDatabase(DatabaseInterface):
        """MySQL実装"""
        
        def save_user(self, user_data: Dict[str, Any]) -> None:
            print(f"MySQLにユーザーデータを保存: {user_data}")
            # MySQL固有の処理...
        
        def get_user(self, user_id: int) -> Optional[Dict[str, Any]]:
            print(f"MySQLからユーザーID {user_id} を取得")
            return {"id": user_id, "name": "John Doe"}

    class PostgreSQLDatabase(DatabaseInterface):
        """PostgreSQL実装（新しい実装も簡単に追加可能）"""
        
        def save_user(self, user_data: Dict[str, Any]) -> None:
            print(f"PostgreSQLにユーザーデータを保存: {user_data}")
            # PostgreSQL固有の処理...
        
        def get_user(self, user_id: int) -> Optional[Dict[str, Any]]:
            print(f"PostgreSQLからユーザーID {user_id} を取得")
            return {"id": user_id, "name": "Jane Smith"}

    class SMTPEmailService(EmailInterface):
        """SMTP実装"""
        
        def send_email(self, to: str, subject: str, body: str) -> None:
            print(f"SMTP経由でメール送信: {to} - {subject}")
            # SMTP固有の処理...

    class AWSEmailService(EmailInterface):
        """AWS SES実装（新しい実装も簡単に追加可能）"""
        
        def send_email(self, to: str, subject: str, body: str) -> None:
            print(f"AWS SES経由でメール送信: {to} - {subject}")
            # AWS SES固有の処理...

    # 高レベルモジュール（ビジネスロジック）
    class UserService:
        """✅ DIP準拠: 抽象化に依存し、具象実装は外部から注入"""
        
        def __init__(self, database: DatabaseInterface, email_service: EmailInterface):
            # 抽象化（インターフェース）に依存
            self.database = database
            self.email_service = email_service
        
        def create_user(self, user_data: Dict[str, Any]) -> None:
            """ユーザー作成のビジネスロジック"""
            # ビジネスロジックは変更せず、具象実装は差し替え可能
            self.database.save_user(user_data)
            self.email_service.send_email(
                user_data['email'], 
                "Welcome!", 
                "アカウントが作成されました"
            )
        
        def get_user(self, user_id: int) -> Optional[Dict[str, Any]]:
            return self.database.get_user(user_id)

    # 使用例1: MySQL + SMTP
    print("=== MySQL + SMTP ===")
    mysql_db = MySQLDatabase()
    smtp_email = SMTPEmailService()
    user_service1 = UserService(mysql_db, smtp_email)

    user_service1.create_user({
        "id": 1, 
        "name": "田中太郎", 
        "email": "tanaka@example.com"
    })

    # 使用例2: PostgreSQL + AWS SES（実装を簡単に切り替え可能）
    print("\n=== PostgreSQL + AWS SES ===")
    postgres_db = PostgreSQLDatabase()
    aws_email = AWSEmailService()
    user_service2 = UserService(postgres_db, aws_email)

    user_service2.create_user({
        "id": 2, 
        "name": "佐藤花子", 
        "email": "sato@example.com"
    })

    # テスト用のモック実装例
    class MockDatabase(DatabaseInterface):
        """テスト用のモック実装"""
        
        def __init__(self):
            self.saved_users = []
        
        def save_user(self, user_data: Dict[str, Any]) -> None:
            self.saved_users.append(user_data)
            print(f"モックDB: ユーザーデータを保存 {user_data}")
        
        def get_user(self, user_id: int) -> Optional[Dict[str, Any]]:
            print(f"モックDB: ユーザーID {user_id} を取得")
            return {"id": user_id, "name": "Test User"}

    class MockEmailService(EmailInterface):
        """テスト用のモック実装"""
        
        def __init__(self):
            self.sent_emails = []
        
        def send_email(self, to: str, subject: str, body: str) -> None:
            email_data = {"to": to, "subject": subject, "body": body}
            self.sent_emails.append(email_data)
            print(f"モックメール: {email_data}")

    # テスト例
    print("\n=== テスト用モック ===")
    mock_db = MockDatabase()
    mock_email = MockEmailService()
    test_service = UserService(mock_db, mock_email)

    test_service.create_user({
        "id": 999, 
        "name": "テストユーザー", 
        "email": "test@example.com"
    })

    print(f"保存されたユーザー数: {len(mock_db.saved_users)}")
    print(f"送信されたメール数: {len(mock_email.sent_emails)}")
    ```