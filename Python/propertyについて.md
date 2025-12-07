# `property`デコレータ
- クラスの属性へのアクセスをメソッド経由で制御するための機能
- **属性の取得、設定、削除の振る舞いをカスタマイズできる**

## `@property`デコレータを使うメリット
- **カプセル化**: 内部実装を隠蔽しつつ、シンプルなインターフェースを提供
- **バリデーション**: 属性の設定時に値の検証や変換が可能
- **読み取り専用属性**: getterのみを定義して、属性を読み取り専用にできる
- **計算プロパティ**: アクセス時に動的に値を計算できる
- **後方互換性**: 既存の属性アクセスを変更せずに、内部実装を変更可能

## 例
### 基本的な使用例  
```python
class User:
    def __init__(self, name: str, age: int):
        self._name = name
        self._age = age  # 内部では_age（プライベート）として保持
    
    @property
    def age(self) -> int:
        """getter: obj.age でアクセス時に呼ばれる"""
        return self._age
    
    @age.setter
    def age(self, value: int):
        """setter: obj.age = value で代入時に呼ばれる"""
        if value < 0:
            raise ValueError("Age cannot be negative")
        self._age = value
    
    @age.deleter
    def age(self):
        """deleter: del obj.age で削除時に呼ばれる"""
        print("Deleting age...")
        del self._age


# 使用例
user = User("Alice", 30)

print(user.age)      # 30 （getterが呼ばれる）

user.age = 25        # setterが呼ばれる
print(user.age)      # 25

user.age = -5        # ValueError: Age cannot be negative

del user.age         # "Deleting age..." と表示される
```

### 計算プロパティの例  
```python
class Rectangle:
    def __init__(self, width: float, height: float):
        self.width = width
        self.height = height
    
    @property
    def area(self) -> float:
        """面積は動的に計算される（読み取り専用）"""
        return self.width * self.height
    
    @property
    def perimeter(self) -> float:
        """周囲長も動的に計算"""
        return 2 * (self.width + self.height)


rect = Rectangle(10, 5)
print(rect.area)       # 50
print(rect.perimeter)  # 30

rect.width = 20
print(rect.area)       # 100 （自動的に再計算される）
```

### 後方互換性の例  
- シナリオ： 既存コードを壊さずにロジックを追加したい
##### Phase 1: 最初のシンプルな実装

最初は単純な属性として公開していたとする。

```python
# user.py (初期バージョン)
class User:
    def __init__(self, email: str):
        self.email = email
```

この時点で、他のチームや外部のコードがこのクラスを使っている。

```python
# 他のチームのコード（変更できない）
user = User("alice@example.com")
print(user.email)              # 直接アクセス
user.email = "bob@example.com" # 直接代入
```

##### Phase 2: 要件追加「メールアドレスのバリデーションが必要になった」

ここで問題が発生する。
**propertyがなかったら？**
メソッドに変更するしかない。

```python
# ❌ 悪い例：インターフェースが変わってしまう
class User:
    def __init__(self, email: str):
        self._email = self._validate(email)
    
    def get_email(self) -> str:
        return self._email
    
    def set_email(self, email: str):
        self._email = self._validate(email)
    
    def _validate(self, email: str) -> str:
        if "@" not in email:
            raise ValueError("Invalid email")
        return email.lower()
```

すると、外部のコードが**すべて壊れる**。

```python
# 他のチームのコード → すべて書き換えが必要 😱
user = User("alice@example.com")
print(user.get_email())              # user.email → user.get_email()
user.set_email("bob@example.com")    # user.email = ... → user.set_email(...)
```

##### Phase 2（正解）: propertyを使う

```python
# ✅ 良い例：propertyでインターフェースを維持
class User:
    def __init__(self, email: str):
        self.email = email  # setterを経由する
    
    @property
    def email(self) -> str:
        return self._email
    
    @email.setter
    def email(self, value: str):
        if "@" not in value:
            raise ValueError("Invalid email")
        self._email = value.lower()
```

外部のコードは**一切変更不要**で、そのまま動く。

```python
# 他のチームのコード → 変更なしで動く！ 🎉
user = User("alice@example.com")
print(user.email)              # そのまま動く（内部でgetterが呼ばれる）
user.email = "bob@example.com" # そのまま動く（内部でsetterが呼ばれる）
user.email = "invalid"         # ValueError（バリデーションが効く）
```

#### より実践的な例：内部実装の変更

##### Phase 1: 名前を単一フィールドで保持

```python
class Employee:
    def __init__(self, name: str):
        self.name = name

# 外部コード
emp = Employee("Taro Yamada")
print(emp.name)  # "Taro Yamada"
```

##### Phase 2: first_name/last_nameに分割したくなった

内部的には分割したいが、`emp.name`でのアクセスは維持したい。

```python
class Employee:
    def __init__(self, first_name: str, last_name: str):
        self.first_name = first_name
        self.last_name = last_name
    
    @property
    def name(self) -> str:
        """後方互換性のため維持"""
        return f"{self.first_name} {self.last_name}"
    
    @name.setter
    def name(self, value: str):
        """後方互換性のため維持"""
        parts = value.split(" ", 1)
        self.first_name = parts[0]
        self.last_name = parts[1] if len(parts) > 1 else ""


# 新しいコード（新APIを使用）
emp = Employee("Taro", "Yamada")
print(emp.first_name)  # "Taro"
print(emp.last_name)   # "Yamada"

# 古いコード（既存APIも動く）
print(emp.name)        # "Taro Yamada"
emp.name = "Jiro Suzuki"
print(emp.first_name)  # "Jiro"
```

## まとめ

| 状況 | propertyなし | propertyあり |
|------|-------------|--------------|
| ロジック追加時 | 外部コードの修正が必要 | 外部コードは変更不要 |
| 内部実装の変更 | APIが変わる | APIを維持できる |
| 段階的な移行 | 困難 | 新旧API共存可能 |
