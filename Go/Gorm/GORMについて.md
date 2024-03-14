- GORMはGolangのORM（Object-Relational Mapping）の１つ

## 使い方
### tagを使った`struct`(構造体)の定義
- GORMではtagを使用して、構造体(`struct`)のフィールドとDBのカラムをマッピングすることができる。  
- tagは、バッククオート(`` ` ``)で囲み、キーと値をコロン(`:`)で区切り、複数のスキーマを指定する場合は、セミコロン(`;`)で区切る。
- 例
  ~~~go
  type Privilege_Dbuserpassword struct {
          CombinationID int    `gorm:"size:30;primaryKey;column:combinationid"`
          System        string `gorm:"foreignKey:system_id;references:system_id;constraint:OnDelete:CASCADE;column:system_id"`
          Dbuser        string `gorm:"size:50;column:dbuser"`
          Password      string `gorm:"size:30;column:dbuserpw"`
  }
  ~~~
  - `size:<サイズ>`: カラムのサイズ(最大長)を指定
  - `primaryKey`: このカラムが主キーであることを示している
  - `column:<カラム名>`: カラム名を指定
  - `foreignKey:<外部キー>`: このカラムが外部キーであることを示している
  - `references:<参照先のカラム名>`: 参照先テーブルのカラム名を指定
    - 参照先テーブルはGROMが推測。`references:<参照先のテーブル名>.<参照先のカラム名>`のように明示的に参照先テーブルを指定することもできる。
  - `constraint:OnDelete:CASCADE`: 参照先のレコードが削除された場合、このレコードも一緒に削除される

### DBへの接続
- `Open`メソッドでDBに接続する
  ```go
  package main

  import (
      "gorm.io/driver/postgres"
      "gorm.io/gorm"
  )

  func main() {
      // データベース接続
      dsn := "port=5432 sslmode=disable TimeZone=Asia/Tokyo host=" + PostgresHost + " user=" + PostgresUser + " password=" + PostgresPassword + " dbname=" + PostgresDatabase

      db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
      if err != nil {
          panic("failed to connect database")
      }
      // ...
  }
  ```

### `AutoMigrate`メソッドでDBにテーブルを追加できる
```go
type User struct {
    gorm.Model
    Name  string
    Email string
}

// テーブル作成
db.AutoMigrate(&User{})
```
- **`AutoMigrate`メソッドはテーブルの追加だけではなく、tagで定義したスキーマを実際データベースに反映するもの**
- `gorm.Model`を埋め込むことで、`ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`フィールドが自動的に追加される。
  - `ID`が自動的に主キーとして設定される

### レコードの追加(Insert)は`Create`、削除は`Delete`、更新(Update)は`Save`メソッドを使用
```go
// ユーザー作成
user := User{Name: "John Doe", Email: "john@example.com"}
result := db.Create(&user)
if result.Error != nil {
    // エラー処理
}
fmt.Println("User ID:", user.ID)

// ユーザー更新
user.Name = "Updated Name"
db.Save(&user)

// ユーザー削除
db.Delete(&user)
```

### ユーザーを検索するには、`First`や`Find`メソッドを使用
- SQLの`SELECT`文に相当。`First`メソッドは`LIMIT 1`を使用して最初のレコードのみを取得し、`Find`メソッドは条件に一致するすべてのレコードを取得
- `Where`、`Order`、`Limit`、`Offset`などのメソッドを使用することで、より詳細な条件やソート、ページネーションなどを実現できる
  - `Offest`メソッドは、取得するレコードのオフセット（スキップする数）を指定するために使用
  - `Limit`メソッドは、取得するレコードの最大数を指定するために使用
#### `First`メソッド
- 指定された条件に一致する最初のレコードを取得
- レコードが見つからない場合は、`ErrRecordNotFound`エラーが返される
#### `Find`メソッド
- 指定された条件に一致するすべてのレコードを取得
- 条件を指定しない場合は、テーブルのすべてのレコードが取得される
- 取得したレコードは、スライスまたは構造体のポインタのスライスに格納される
#### 例
```go
// 主キーを使用してレコードを取得
var user User
db.First(&user, 1)

// 条件を指定してレコードを取得
var users []User
db.Find(&users, "email = ?", "john@example.com")

// すべてのレコードを取得
var allUsers []User
db.Find(&allUsers)

// Firstメソッドで特定のカラムの値を条件とする例
var user User
db.First(&user, "email = ?", "john@example.com")

// Whereメソッドの例（ageが18より大きいユーザーレコードを取得）
var users []User
db.Where("age > ?", 18).Find(&users)

// Orderメソッドの例（created_atカラムの値を降順（DESC）でソート）
var users []User
db.Order("created_at DESC").Find(&users)

// Limitメソッドの例（最大10件のユーザーレコードを取得）
var users []User
db.Limit(10).Find(&users)

// Offsetメソッドの例（最初の20件のレコードをスキップし、その次の10件のユーザーレコードを取得）
var users []User
db.Offset(20).Limit(10).Find(&users)

// これらのメソッドを組み合わせた例
var users []User
db.Where("age > ?", 18).Order("created_at DESC").Limit(10).Offset(20).Find(&users)
```
