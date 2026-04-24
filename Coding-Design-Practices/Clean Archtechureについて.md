# Clean Architecture

## 1. 概要

Robert C. Martin (Uncle Bob) が2012年に提唱したソフトウェアアーキテクチャ設計原則。
核心は **「ビジネスロジックを外部の詳細 (DB, Web, UI, フレームワーク) から独立させる」** こと。

特定の言語やフレームワークに依存しない、汎用的な設計思想である。

---

## 2. 解決する問題

従来のレイヤードアーキテクチャ（Controller → Service → Repository → DB）では:

- ビジネスロジックが DB の構造に引きずられる
- フレームワークを変えるとビジネスロジックまで書き直しになる
- UI の変更が DB 層まで波及する
- テスト時に DB やネットワークが必要になる

Clean Architecture はこれらの問題を**依存性の方向を制御する**ことで解決する。

---

## 3. 同心円モデル

Clean Architecture は同心円で表現される。**外側の円は内側の円に依存してよいが、内側は外側を知らない。**

```
┌─────────────────────────────────────────────────────────┐
│                Frameworks & Drivers (最外層)              │
│  Web フレームワーク, DB ドライバ, UI, 外部サービス          │
│                                                         │
│  ┌─────────────────────────────────────────────────┐    │
│  │          Interface Adapters (アダプター層)         │    │
│  │  Controller, Presenter, Gateway, Repository 実装  │    │
│  │                                                  │    │
│  │  ┌──────────────────────────────────────────┐    │    │
│  │  │       Application Business Rules          │    │    │
│  │  │       (ユースケース層)                      │    │    │
│  │  │                                          │    │    │
│  │  │  ┌──────────────────────────────────┐    │    │    │
│  │  │  │   Enterprise Business Rules       │    │    │    │
│  │  │  │   (エンティティ層)                  │    │    │    │
│  │  │  └──────────────────────────────────┘    │    │    │
│  │  └──────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

---

## 4. 各層の役割

### (1) Enterprise Business Rules (エンティティ層) - 最内層

- **ビジネスの核心となるルールとデータ構造**
- アプリケーション固有ではなく、企業全体で共通するルール
- 技術的な関心事を一切含まない（DB, HTTP, フレームワーク等を知らない）
- 最も変更されにくい層

例:
- 「インシデントには triggered / acknowledged / resolved のステータスがある」
- 「urgency は severity_source と severity_value から導出される」

### (2) Application Business Rules (ユースケース層)

- **アプリケーション固有のビジネスルール**
- エンティティを操作して1つのユースケースを実行する
- データの流れ（入力 → 処理 → 出力）を制御する
- エンティティ層にのみ依存する

例:
- 「インシデントを更新する際は、まず権限を確認し、楽観ロックを検証し、更新後に通知を送る」

### (3) Interface Adapters (アダプター層)

- **外部と内部のデータ形式を変換する**
- Controller: 外部入力 → ユースケースの入力形式に変換
- Presenter: ユースケースの出力 → 外部出力形式に変換
- Gateway/Repository 実装: DB の行 → エンティティに変換

例:
- HTTP JSON → ユースケースの InputData に変換
- DB の行データ → ドメインエンティティに変換

### (4) Frameworks & Drivers (最外層)

- **具体的な技術の詳細**
- Web フレームワーク (FastAPI, Django, Express...)
- DB ドライバ (SQLAlchemy, Prisma...)
- 外部 API クライアント (Slack SDK, AWS SDK...)
- この層は「糊」であり、ビジネスロジックは含まない

---

## 5. 依存性ルール (The Dependency Rule)

**Clean Architecture で最も重要なルール。**

> ソースコードの依存性は、常に内側の円に向かわなければならない。
> 内側の円は、外側の円のことを何も知ってはならない。

```
Frameworks → Adapters → Use Cases → Entities
                 依存の方向 →→→→→→→
```

### なぜ重要か

依存の方向が統一されていることで:
- 内側の変更は外側に波及しない
- 外側の変更（DB 変更、フレームワーク変更）は内側に波及しない
- 内側だけでテストできる

### 依存性逆転の原則 (DIP) との関係

自然に書くと、ユースケースは DB に依存する:

```
UseCase → DatabaseRepository → Database
```

しかし依存性ルールにより、ユースケースは外側の DB を知ってはいけない。
ここで **依存性逆転 (Dependency Inversion Principle)** を使う:

```
UseCase → IRepository (インターフェース、内側で定義)
               ↑ 実装
         DatabaseRepository (外側で定義)
```

- インターフェースは内側（ユースケース層 or エンティティ層）で定義する
- 実装は外側（アダプター層）で行う
- 実行時に DI コンテナ等で実装を注入する

これにより **依存の方向はインターフェースに向かう（内向き）** が、
**制御の流れは外から内、内から外** と自由に流れる。

---

## 6. データの流れ

リクエストからレスポンスまでのデータ変換:

```
HTTP Request (JSON)
    │
    ▼ Controller が変換
Use Case Input (InputData / Request Model)
    │
    ▼ Use Case が処理
Entity (ドメインモデル)
    │
    ▼ Use Case が変換
Use Case Output (OutputData / Response Model)
    │
    ▼ Presenter が変換
HTTP Response (JSON)
```

**各層の境界でデータ型が変わる**のが特徴。
同じ「インシデント」でも:

| 層 | データ型 | 目的 |
|----|---------|------|
| 最外層 | HTTP JSON / Form | 通信プロトコルに適した形式 |
| アダプター層 | Request DTO / Response DTO | 入出力の検証と変換 |
| ユースケース層 | InputData / OutputData | ユースケースに必要な情報の過不足ない定義 |
| エンティティ層 | Entity / Value Object | ビジネスルールの表現に適した構造 |
| アダプター層 (DB) | ORM / Row | DB 格納に適した構造 |

一見冗長だが、層間の結合を防ぎ、各層が独立して変更可能になる。

---

## 7. 関連するアーキテクチャとの比較

Clean Architecture は以下のアーキテクチャを統合・一般化したもの:

| アーキテクチャ | 提唱者 | 共通点 |
|-------------|--------|--------|
| Hexagonal Architecture (Ports & Adapters) | Alistair Cockburn | ビジネスロジックと外部の分離、Port = インターフェース |
| Onion Architecture | Jeffrey Palermo | 同心円構造、依存性は内向き |
| BCE (Boundary-Control-Entity) | Ivar Jacobson | Boundary = Adapter, Control = Use Case, Entity = Entity |

これらはすべて **「外部詳細からビジネスロジックを守る」** という同じ目的を持つ。
Clean Architecture はこれらの共通原則を明文化したもの。

---

## 8. メリットと代償

### メリット

| メリット | 説明 |
|---------|------|
| フレームワーク独立 | ビジネスロジックがフレームワークに依存しない。移行が容易 |
| テスト容易性 | 内側の層を外部依存なしでユニットテスト可能 |
| UI 独立 | Web → CLI → gRPC に変更してもビジネスロジックは不変 |
| DB 独立 | PostgreSQL → MongoDB に変えてもユースケース層は不変 |
| 変更の局所化 | 各層の変更が他の層に波及しにくい |

### 代償

| 代償 | 説明 |
|------|------|
| ファイル数の増加 | 1 機能あたりの必要ファイル数が多い |
| 変換コスト | 層の境界ごとにデータ変換が発生する |
| 学習コスト | 設計パターンの理解と遵守が必要 |
| 過剰設計のリスク | 小規模プロジェクトでは YAGNI に反する場合がある |
| 間接化の増加 | 処理を追うのに複数ファイルを横断する必要がある |

### 適用の判断基準

**向いているケース**:
- 長期間メンテナンスされるプロダクト
- ビジネスロジックが複雑
- 外部システム（DB, API）が将来変わる可能性がある
- チーム開発で責務分担を明確にしたい

**向いていないケース**:
- プロトタイプ、PoC
- CRUD 中心でビジネスロジックが薄い
- 短期間で破棄されるもの

---

## 9. よくある誤解

### 「層の数は4つでなければならない」

→ 同心円の数は固定ではない。プロジェクトの規模に応じて増減してよい。
重要なのは数ではなく **依存性ルール（内向きの依存）が守られていること**。

### 「Clean Architecture = フォルダ構成」

→ フォルダを domain / application / infrastructure に分けただけでは Clean Architecture にならない。
重要なのはフォルダ構成ではなく **import の方向**。
domain/ が infrastructure/ を import していたら、フォルダ構成に関係なく設計は壊れている。

### 「すべての層を必ず通らなければならない」

→ 単純な CRUD は層をスキップしてもよい。
ただし依存性ルールは常に守る。

### 「DB のインターフェースを Domain 層で定義 = DB を簡単に差し替えられる」

→ インターフェースがあっても、実際に DB を差し替えるのは大きな作業。
真のメリットは差し替えの容易さではなく、**テスト時のモック差し替え**と**ビジネスロジックの DB 非依存性**。

---

## 10. 参考文献

- Robert C. Martin, "Clean Architecture: A Craftsman's Guide to Software Structure and Design" (2017)
- Robert C. Martin, "The Clean Architecture" (blog post, 2012)
  https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
