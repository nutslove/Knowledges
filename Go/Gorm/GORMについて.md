- GORMはGolangのORM（Object-Relational Mapping）の１つ

# 使い方
## tagを使った`struct`(構造体)の定義
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
  - その他指定できるスキーマは https://gorm.io/ja_JP/docs/models.html から確認可能

### type(データ型)について
- tagで`type`でデータ型を明示的に指定することもできる
- 例  
  ```go
  type User struct {
      Name string `gorm:"type:varchar(100)"`
      Age int `gorm:"type:int"`
      IsActive bool `gorm:"type:boolean"`
      CreatedAt time.Time `gorm:"type:datetime"`
      Description string `gorm:"type:text"`
  }
  ```
- 明示的に指定しない場合は、gormがGoの型からデータベースのデータ型を推測して設定してくれる
  - e.g. `string` → `VARCHAR`、`int` → `INTEGER`、`time.Time` → `DATETIME`

## DBへの接続
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

## `AutoMigrate()`メソッドでDBにテーブルを追加できる
- https://gorm.io/ja_JP/docs/migration.html
> **AutoMigrate はテーブル、外部キー、制約、カラム、インデックスを作成します。 カラムのサイズ、精度、null可否などが変更された場合、既存のカラムの型を変更します。 しかし、データを守るために、使われなくなったカラムの削除は実行されません。**

```go
type User struct {
    gorm.Model
    Name  string
    Email string
}

// テーブル作成/更新
db.AutoMigrate(&User{})
```

- テーブルがすでに存在していて、構造体に変更がなければ`AutoMigrate()`は何もしない

- **`AutoMigrate()`メソッドはテーブルの追加だけではなく、tagで定義したスキーマを実際データベースに反映するもの（Migrate the schema）**
- `gorm.Model`を埋め込むことで、`ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`フィールドが自動的に追加される。
  - `ID`が自動的に主キーとして設定され、`AUTOINCREMENT`が設定される
- デフォルトでは、GORMはテーブル名の末尾に`s`を付けて複数形で作成する（e.g. User → Users）  
  これを防ぐためにはGORM内蔵の`TableName()`メソッドで明示的にテーブル名を`return`に指定する。  
  例えば、以下の例ではデフォルトではテーブル名は`dbuserpasswords`になるけど、`TableName()`メソッドの`return`に`dbuserpassword`と指定することで単数形として作られる。  
  ```go
  type Dbuserpassword struct {
          System        string `gorm:"foreignKey:system_id;references:system_id;constraint:OnDelete:CASCADE;column:system_id"`
          Dbuser        string `gorm:"size:50;column:dbuser"`
          Password      string `gorm:"size:30;column:dbuserpw"`
  }

  func (Dbuserpassword) TableName() string {
          return "dbuserpassword"
  }
  ```

## レコードの追加(Insert)は`Create`、削除は`Delete`、更新(Update)は`Update`や`Save`メソッドを使用
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

### autoIncrement
- 主キーのフィールド(カラム)の型が`int`もしくは`uint`の場合、明示的に`autoIncrement`オプションを指定しなくても自動的に`autoIncrement`が適用される
- `autoIncrement`を無効にしたい場合は`autoIncrement:false`を指定する  
  ```go
  type Product struct {
      ProductID int    `gorm:"primaryKey;autoIncrement:false"`
      Name      string
  }
  ```

- `autoIncrement`が有効になっているフィールド(カラム)はInsert時、指定しなくても自動的に最後のレコードの値＋１の値で挿入される。  
  また、Gormはレコードの挿入後に自動的に挿入された各カラムの値を構造体に反映してくれて、そこから`autoIncrement`で自動で挿入された値を確認/取得することができる
  ```go
  type CareerBoard struct {
    Number int       `gorm:"primaryKey;column:num"`
    Title  string    `gorm:"size:100;column:title"`
    Author string    `gorm:"size:30;column:author"`
    Date   time.Time `gorm:"type:datetime;column:date"`
    Count  int       `gorm:"column:count"`
  }

  addedPost := models.CareerBoard{
    // Numberは指定してない
    Title:  "テスト",
    Author: username.(string),
    Date:   time.Now(),
    Count:  0,
  }

  result := db.Create(&addedPost)
  if result.Error != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "success": false,
      "message": result.Error,
    })
    return
  }

  // addedPos構造体に挿入されたNumberも反映されて、取得できる
  fmt.Println("挿入されたレコードのNumber:", addedPost.Number)
  ```

### `Update`と`Save`の違いについて
- https://gorm.io/docs/update.html
- `Save`は構造体の変更された(変更がないフィールドも含めて)すべてのフィールドを一括で更新  
  > `Save` will save all fields when performing the Updating SQL  

  ```go
  db.First(&user)

  user.Name = "jinzhu 2"
  user.Age = 100
  db.Save(&user)
  // UPDATE users SET name='jinzhu 2', age=100, birthday='2016-01-01', updated_at = '2013-11-17 21:34:10' WHERE id=111;
  ```

  > `Save` is a combination function. If save value does not contain primary key, it will execute `Create`, otherwise it will execute `Update` (with all fields).

- `Update`は指定された(特定の)フィールドのみを更新  

  ```go
  // Update with conditions
  db.Model(&User{}).Where("active = ?", true).Update("name", "hello")
  // UPDATE users SET name='hello', updated_at='2013-11-17 21:34:10' WHERE active=true;

  // User's ID is `111`:
  db.Model(&user).Update("name", "hello")
  // UPDATE users SET name='hello', updated_at='2013-11-17 21:34:10' WHERE id=111;

  // Update attributes with `struct`, will only update non-zero fields
  db.Model(&user).Updates(User{Name: "hello", Age: 18, Active: false})
  // UPDATE users SET name='hello', age=18, updated_at = '2013-11-17 21:34:10' WHERE id = 111;

  // Update attributes with `map`
  db.Model(&user).Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})
  // UPDATE users SET name='hello', age=18, active=false, updated_at='2013-11-17 21:34:10' WHERE id=111;
  ```

  > `Updates` supports updating with `struct` or `map[string]interface{}`, when updating with `struct` it will only update non-zero fields by default

## ユーザーを検索するには、`First`や`Find`メソッドを使用
- SQLの`SELECT`文に相当。`First`メソッドは`LIMIT 1`を使用して最初のレコードのみを取得し、`Find`メソッドは条件に一致するすべてのレコードを取得
- `Where`、`Order`、`Limit`、`Offset`などのメソッドを使用することで、より詳細な条件やソート、ページネーションなどを実現できる
  - `Offest`メソッドは、取得するレコードのオフセット（スキップする数）を指定するために使用
  - `Limit`メソッドは、取得するレコードの最大数を指定するために使用
### `First`メソッド
- 指定された条件に一致する最初のレコードを取得
- レコードが見つからない場合は、`ErrRecordNotFound`エラーが返される
### `Find`メソッド
- 指定された条件に一致するすべてのレコードを取得
- 条件を指定しない場合は、テーブルのすべてのレコードが取得される
- 取得したレコードは、スライスまたは構造体のポインタのスライスに格納される
### 例
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

## その他

### レコード数の取得
- `Count`メソッドを使う  
  > The Count method in GORM is used to retrieve the number of records that match a given query. It’s a useful feature for understanding the size of a dataset, particularly in scenarios involving conditional queries or data analysis.
- https://gorm.io/ja_JP/docs/advanced_query.html#Count
- あるテーブル内のすべてのレコード数を取得する  
  ```go
  var count int64
  db.Model(&YourModel{}).Count(&count)
  ```
- 他の例  
  ```go
  var count int64

  // Counting users with specific names
  db.Model(&User{}).Where("name = ?", "jinzhu").Or("name = ?", "jinzhu 2").Count(&count)
  // SQL: SELECT count(1) FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2'

  // Counting users with a single name condition
  db.Model(&User{}).Where("name = ?", "jinzhu").Count(&count)
  // SQL: SELECT count(1) FROM users WHERE name = 'jinzhu'

  // Counting records in a different table
  db.Table("deleted_users").Count(&count)
  // SQL: SELECT count(1) FROM deleted_users
  ```

### `Order`について
- `Order`であるカラムの値で降順、昇順(default)に並べ替えることができる
- 例  
  ```go
  type CareerBoard struct {
    Number int       `gorm:"primaryKey;column:num"`
    Title  string    `gorm:"size:100;column:title"`
    Author string    `gorm:"size:30;column:author"`
    Date   time.Time `gorm:"type:datetime;column:date"`
    Count  int       `gorm:"column:count"`
  }
  var posts []CareerBoard
  var db *gorm.DB
  db.Order("num desc").Find(&posts) // ★structでcolumnに指定しているカラム名を指定
  ```

### `Offset`と`Limit`について
- `Limit`は取得するレコード数を制限する
- `Offset`は何件目のレコードから取得するかを指定（indexは0からスタート (1件目のデータのindex=0) ）
- 例  
  ```go
  db.Order("num desc").Offset((page - 1) * 15).Limit(15).Find(&posts)
  ```