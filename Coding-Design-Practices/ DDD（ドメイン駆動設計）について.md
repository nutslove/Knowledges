# Domain-Driven Design (DDD)

## 1. 概要

Eric Evans が2003年の著書 "Domain-Driven Design: Tackling Complexity in the Heart of Software" で提唱した設計手法。

核心は **「ソフトウェアの構造をビジネスドメイン（業務領域）の構造に合わせる」** こと。
技術的な都合ではなく、ビジネスの専門家が使う言葉や概念をそのままコードに反映させる。

---

## 2. 解決する問題

ソフトウェア開発で最も難しいのは、技術ではなく**ビジネスの複雑さを正しくモデル化すること**。

DDD が解決する問題:
- 開発者とビジネス担当者の間で用語がずれ、認識の齟齬が生まれる
- DB テーブル構造がそのままコードの構造になり、ビジネスルールが散在する
- 同じ概念が場所によって異なる意味で使われる
- システムが大きくなると、変更の影響範囲が予測できなくなる

---

## 3. 戦略的設計と戦術的設計

DDD は2つのレベルに分かれる。

### 戦略的設計 (Strategic Design)

大きな視点: **システム全体をどう分割し、チーム間でどう整合性を保つか**。

- ユビキタス言語 (Ubiquitous Language)
- 境界づけられたコンテキスト (Bounded Context)
- コンテキストマップ (Context Map)

### 戦術的設計 (Tactical Design)

小さな視点: **1つのコンテキスト内部をどうモデリングするか**。

- エンティティ (Entity)
- 値オブジェクト (Value Object)
- 集約 (Aggregate)
- リポジトリ (Repository)
- ファクトリ (Factory)
- ドメインイベント (Domain Event)
- ドメインサービス (Domain Service)

---

## 4. 戦略的設計

### 4.1 ユビキタス言語 (Ubiquitous Language)

**DDD で最も重要な概念。**

開発者とビジネス担当者が**共通の語彙**を使い、コード・ドキュメント・会話すべてで同じ言葉を使うこと。

悪い例:
```
ビジネス担当者: 「チケットのステータスを "対応中" にして」
開発者:         ticket.state = "in_progress"  ← 用語が違う
```

良い例:
```
ビジネス担当者: 「インシデントを acknowledge して」
開発者:         incident.acknowledge()  ← 同じ言葉
```

**ルール**:
- コード内のクラス名、メソッド名、変数名にビジネス用語をそのまま使う
- 技術用語 (`data`, `info`, `manager`, `handler`) ではなくドメイン用語を使う
- 新しい用語が出たら定義を合意し、全員が統一して使う

### 4.2 境界づけられたコンテキスト (Bounded Context)

**同じ言葉でもコンテキストによって意味が異なる**ことを認め、明確な境界を設ける。

例: 「ユーザー」の意味
- 認証コンテキスト: ログイン情報を持つアカウント
- インシデント管理コンテキスト: インシデントの担当者
- 請求コンテキスト: 契約と支払い情報を持つ顧客

これらを1つの `User` クラスにまとめると、全コンテキストの要求が混ざり合い複雑化する。

**解決策**: コンテキストごとに独立したモデルを持つ。

```
[認証コンテキスト]          [インシデント管理コンテキスト]
 Account                    Assignee
 - email                    - assignee_id
 - password_hash            - assignee_name
 - last_login               - assignee_email
```

### 4.3 コンテキストマップ (Context Map)

複数の Bounded Context がどう関係するかを図示したもの。

コンテキスト間の関係パターン:

| パターン | 説明 |
|---------|------|
| Shared Kernel | 2つのコンテキストが共通のモデルを共有する |
| Customer-Supplier | 上流 (供給者) が下流 (顧客) の要求に応じてモデルを提供する |
| Conformist | 下流が上流のモデルにそのまま従う (交渉力がない場合) |
| Anti-Corruption Layer (ACL) | 外部モデルを自コンテキストのモデルに変換する防御層 |
| Open Host Service | 公開 API として標準的なプロトコルを提供する |
| Published Language | 共通のスキーマ (JSON, Protobuf 等) でやり取りする |

---

## 5. 戦術的設計

### 5.1 エンティティ (Entity)

**同一性 (Identity) を持つオブジェクト**。属性が変わっても、IDが同じなら同じものとみなす。

特徴:
- 一意の識別子 (ID) を持つ
- ライフサイクルがある (生成 → 変更 → 削除)
- 属性が変わっても同一性は変わらない
- ミュータブル (変更可能)

```python
class Incident:
    id: IncidentId           # ← 同一性の基準
    title: str               # ← 変わっても同じインシデント
    status: IncidentStatus   # ← 変わっても同じインシデント

# incident_a.id == incident_b.id なら同じインシデント
# (たとえ title や status が異なっていても)
```

### 5.2 値オブジェクト (Value Object)

**属性の値そのものが重要なオブジェクト**。同一性を持たない。

特徴:
- IDを持たない
- 属性の値が等しければ同じとみなす
- イミュータブル (不変)
- 自己検証する (不正な値を受け付けない)
- 副作用のないメソッドを持てる

```python
class EmailAddress:
    value: str
    # "user@example.com" == "user@example.com" (値が同じなら等しい)
    # 生成時にフォーマットを検証する

class Money:
    amount: Decimal
    currency: str
    def add(self, other: Money) -> Money:
        # 新しい Money を返す (元の値は変えない = イミュータブル)
        return Money(amount=self.amount + other.amount, currency=self.currency)
```

### エンティティ vs 値オブジェクトの判断基準

| 観点 | エンティティ | 値オブジェクト |
|------|------------|-------------|
| 同一性 | IDで識別する | 値で等価判定する |
| 可変性 | ミュータブル | イミュータブル |
| ライフサイクル | ある (生成 → 変更 → 削除) | ない (作ったら変えない、新しく作る) |
| 例 | ユーザー, 注文, インシデント | メールアドレス, 金額, 住所, ステータス |

**判断のコツ**: 「2つのオブジェクトの属性がすべて同じとき、交換可能か？」
→ Yes なら値オブジェクト。No (区別する必要がある) ならエンティティ。

### 5.3 集約 (Aggregate)

**関連するエンティティと値オブジェクトをまとめた整合性の境界**。

```
┌─ Aggregate ──────────────────────────┐
│                                      │
│  [Aggregate Root: Order] ◄── 外部からのアクセスは必ずここを経由
│      │
│      ├── OrderLine (Entity)
│      │      ├── ProductId (Value Object)
│      │      └── Quantity (Value Object)
│      │
│      ├── OrderLine (Entity)
│      │      ├── ...
│      │
│      └── ShippingAddress (Value Object)
│                                      │
└──────────────────────────────────────┘
```

**ルール**:

1. **集約ルート (Aggregate Root)**
   - 集約の入り口。外部から集約内部のオブジェクトに直接アクセスしてはいけない
   - 集約内の整合性を保証する責務を持つ

2. **トランザクション境界**
   - 1つのトランザクションで変更できるのは1つの集約のみ
   - 複数の集約をまたぐ変更はドメインイベント等で結果整合性を使う

3. **他の集約への参照は ID のみ**
   - 集約間はオブジェクト参照ではなく ID で参照する

```python
class Order:  # Aggregate Root
    order_id: OrderId
    lines: List[OrderLine]  # 集約内のエンティティ (直接保持)
    customer_id: CustomerId  # 他の集約は ID のみ (Customer オブジェクトではない)

    def add_line(self, product_id, quantity):
        # 集約ルートを通じてのみ内部を変更できる
        # ビジネスルール (上限チェック等) をここで強制
        if len(self.lines) >= 100:
            raise ValueError("Order line limit exceeded")
        self.lines.append(OrderLine(product_id=product_id, quantity=quantity))
```

**集約の大きさの設計指針**:
- 小さくする。必要最小限のエンティティのみ含める
- 大きな集約はパフォーマンス問題やロック競合を招く
- 迷ったら小さく始め、必要に応じて統合する

### 5.4 リポジトリ (Repository)

**集約の永続化と取得を抽象化するインターフェース**。

- 集約ルートごとに1つのリポジトリ
- コレクションのように振る舞う (add, find, remove)
- DB の詳細 (SQL, テーブル構造) を隠蔽する
- ドメイン層ではインターフェースのみ定義し、実装はインフラ層で行う

```python
# ドメイン層 (インターフェース)
class IOrderRepository(ABC):
    def find_by_id(self, order_id: OrderId) -> Order | None: ...
    def save(self, order: Order) -> None: ...
    def delete(self, order_id: OrderId) -> None: ...

# インフラ層 (実装)
class SQLOrderRepository(IOrderRepository):
    def find_by_id(self, order_id: OrderId) -> Order | None:
        row = self.session.query(OrderORM).filter(...).first()
        return self._to_domain(row)  # ORM → ドメインモデル変換
```

### 5.5 ファクトリ (Factory)

**複雑なオブジェクトの生成ロジックをカプセル化する**。

コンストラクタが複雑になる場合や、生成時にビジネスルールが絡む場合に使う。

```python
class OrderFactory:
    def create(self, customer_id, items) -> Order:
        order_id = generate_new_id()
        lines = [OrderLine(product_id=item.id, quantity=item.qty) for item in items]
        return Order(
            order_id=order_id,
            lines=lines,
            status=OrderStatus.DRAFT,
            created_at=datetime.now(),
        )
```

### 5.6 ドメインイベント (Domain Event)

**ドメイン内で発生した重要な出来事を表すオブジェクト**。

- 過去形で命名する (「〜が起きた」)
- イミュータブル
- 集約間の結果整合性を実現する手段

```python
class OrderPlaced:       # 「注文が確定された」
    order_id: OrderId
    customer_id: CustomerId
    placed_at: datetime

class PaymentReceived:   # 「支払いが受領された」
    order_id: OrderId
    amount: Money
    received_at: datetime
```

**用途**:
- 集約間の連携 (1トランザクション = 1集約の制約を守りつつ)
- 監査ログ
- 外部システムへの通知
- イベントソーシング

```
[注文コンテキスト]                    [通知コンテキスト]
 Order.place()                        
   └─ publish(OrderPlaced) ──────→  on OrderPlaced:
                                       send_confirmation_email()
```

### 5.7 ドメインサービス (Domain Service)

**特定のエンティティに属さないビジネスロジック**を配置する場所。

使う基準:
- 複数の集約にまたがるロジック
- エンティティの責務としては不自然な操作

```python
class TransferService:
    def transfer(self, from_account: Account, to_account: Account, amount: Money):
        # この操作は Account 単体の責務ではない
        from_account.withdraw(amount)
        to_account.deposit(amount)
```

**注意**: ドメインサービスを多用しすぎると「貧血ドメインモデル」になる。
まずエンティティ自身にロジックを持たせることを検討し、
どうしても収まらないものだけドメインサービスにする。

---

## 6. 貧血ドメインモデル (Anemic Domain Model) アンチパターン

エンティティがデータの入れ物 (getter/setter のみ) になり、
ビジネスロジックがすべてサービス層にある状態。Martin Fowler が「アンチパターン」と呼んだ。

```python
# アンチパターン: 貧血ドメインモデル
class Order:
    id: str
    status: str
    total: float
    # ロジックなし、ただのデータ入れ物

class OrderService:
    def place_order(self, order: Order):
        if order.status != "draft":
            raise Error("...")
        order.status = "placed"     # ← ビジネスルールがサービスに漏れている
        order.total = self.calculate_total(order)
```

```python
# リッチドメインモデル (DDD が推奨)
class Order:
    id: OrderId
    status: OrderStatus
    lines: List[OrderLine]

    def place(self):
        if self.status != OrderStatus.DRAFT:
            raise DomainError("Only draft orders can be placed")
        self.status = OrderStatus.PLACED
        # ビジネスルールがエンティティ自身にある
```

---

## 7. 集約設計のガイドライン

Vaughn Vernon の "Implementing Domain-Driven Design" で示された実践的ルール:

### ルール 1: 真の不変条件を集約境界内で保護する

整合性を**常に**保証する必要があるデータだけを同じ集約に入れる。

### ルール 2: 小さな集約を設計する

大きな集約は:
- トランザクションの競合が発生しやすい
- メモリ消費が大きい
- ロード時間が長い

### ルール 3: 他の集約は ID で参照する

オブジェクト参照ではなく ID を持つ。これにより集約間の結合を防ぐ。

```python
# NG: オブジェクト参照
class Order:
    customer: Customer  # ← Customer 集約への直接参照

# OK: ID 参照
class Order:
    customer_id: CustomerId  # ← ID のみ
```

### ルール 4: 集約間は結果整合性を使う

1つのトランザクションで複数の集約を変更しない。
ドメインイベントで非同期に整合性を取る。

```
[トランザクション1]          [トランザクション2]
 Order.place()               Inventory.reserve()
   └─ publish(OrderPlaced)       ↑
          └─────────────────────┘
              イベント経由で連携
```

---

## 8. DDD と Clean Architecture の関係

DDD と Clean Architecture は独立した概念だが、相性が非常に良い。

| DDD の要素 | Clean Architecture での配置 |
|-----------|---------------------------|
| エンティティ, 値オブジェクト | エンティティ層 (最内層) |
| 集約, ドメインイベント | エンティティ層 |
| リポジトリインターフェース | エンティティ層 (またはユースケース層) |
| リポジトリ実装 | アダプター層 |
| ユースケース / アプリケーションサービス | ユースケース層 |
| ドメインサービス | エンティティ層 (またはユースケース層) |
| ファクトリ | 生成ロジックの複雑さに応じて各層 |

```
Clean Architecture の層構造
┌───────────────────────────────────────────┐
│ Frameworks (FastAPI, SQLAlchemy, Slack SDK) │
│ ┌───────────────────────────────────────┐ │
│ │ Adapters (Repository実装, Controller)  │ │
│ │ ┌───────────────────────────────────┐ │ │
│ │ │ Use Cases (Interactor)            │ │ │
│ │ │ ┌───────────────────────────────┐ │ │ │
│ │ │ │ Entities ← DDDのモデルが住む場所│ │ │ │
│ │ │ │  Entity, Value Object,        │ │ │ │
│ │ │ │  Aggregate, Domain Event,     │ │ │ │
│ │ │ │  IFRepository                 │ │ │ │
│ │ │ └───────────────────────────────┘ │ │ │
│ │ └───────────────────────────────────┘ │ │
│ └───────────────────────────────────────┘ │
└───────────────────────────────────────────┘
```

- **Clean Architecture** は「依存性の方向」を定義する → **構造のルール**
- **DDD** は「内側の層で何をモデリングするか」を定義する → **モデリングの手法**

両者を組み合わせることで:
- Clean Architecture がビジネスロジックを外部から保護し
- DDD がそのビジネスロジックを豊かにモデリングする

---

## 9. DDD 導入の判断基準

### 向いているケース

- ビジネスロジックが複雑 (条件分岐、状態遷移、計算ルールが多い)
- ドメインエキスパート (業務に詳しい人) にアクセスできる
- 長期的にメンテナンスされるシステム
- チームがドメイン知識を深く理解する必要がある

### 向いていないケース

- 純粋な CRUD アプリケーション (ビジネスロジックがほぼない)
- 技術的な課題が中心 (データパイプライン、インフラツール等)
- 短期間のプロトタイプ
- ドメインエキスパートと対話する機会がない

### 段階的導入

DDD は all-or-nothing ではない。戦術的パターン（値オブジェクト、リポジトリ）だけでも価値がある。

推奨の導入順序:
1. **値オブジェクト** → すぐに効果が出る。プリミティブ型の乱用を防ぐ
2. **エンティティ** → ビジネスロジックをモデルに寄せる
3. **リポジトリ** → データアクセスを抽象化する
4. **集約** → 整合性の境界を明確にする
5. **ドメインイベント** → 集約間の連携を疎結合にする
6. **境界づけられたコンテキスト** → システム全体の分割に取り組む

---

## 10. 参考文献

- Eric Evans, "Domain-Driven Design: Tackling Complexity in the Heart of Software" (2003) — 原典
- Vaughn Vernon, "Implementing Domain-Driven Design" (2013) — 実践的な適用方法
- Vaughn Vernon, "Domain-Driven Design Distilled" (2016) — DDD の簡潔な入門
- Martin Fowler, "AnemicDomainModel" (blog post, 2003)
  https://martinfowler.com/bliki/AnemicDomainModel.html
