## Factory関数とは
- オブジェクトのインスタンスを作成して返す関数。  
  特にGo言語では、コンストラクタの概念が直接サポートされていないため、Factory関数がオブジェクトの初期化とインスタンス作成のためによく使用される。これにより、コードの再利用性が向上し、より維持しやすくなる。  
  Factory関数は、特定のinterfaceを満たすオブジェクトを動的に作成する際に特に有用。これにより、プログラムの柔軟性が向上し、異なるコンテキストで異なる実装を簡単に切り替えることができる。  

  例えば、以下のようにinterface `Animal`があり、それを実装する構造体`Dog`と`Cat`がある場合、`Animal `interfaceを満たす`Dog`または`Cat`のインスタンスを返すFactory関数を定義することができる。
  ~~~go
  package main

  import "fmt"

  // Animal インターフェース
  type Animal interface {
      Speak() string
  }

  // Dog 構造体
  type Dog struct{}

  // Dog が Animal インターフェースの Speak メソッドを実装
  func (d Dog) Speak() string {
      return "Woof!"
  }

  // Cat 構造体
  type Cat struct{}

  // Cat が Animal インターフェースの Speak メソッドを実装
  func (c Cat) Speak() string {
      return "Meow!"
  }

  // AnimalFactory は引数に応じて Dog または Cat のインスタンスを返すファクトリ関数
  func AnimalFactory(animalType string) Animal {
      switch animalType {
      case "dog":
          return Dog{}
      case "cat":
          return Cat{}
      default:
          return nil // 未知の型の場合は nil を返す
      }
  }

  func main() {
      animal := AnimalFactory("dog")
      fmt.Println(animal.Speak()) // "Woof!"

      animal = AnimalFactory("cat")
      fmt.Println(animal.Speak()) // "Meow!"
  }
  ~~~
  - この例では、`AnimalFactory`関数がFactory関数として機能しており、`animalType`の値に基づいて`Dog`または`Cat`のインスタンスを作成し、それを`Animal`interface型として返している。
- **Goでは一般的に生成したい構造体(struct)の前に`New`を付けた名前でFactory関数を作成する**
  - 例）  
    ~~~go
    type IItemRepository interface {
      Findll() (*[]Item, error)
    }

    type ItemMemoryRepository struct {
      items []Item
    }

    func NewItemMemoryRepository(items []Item) IItemRepository { // これがFactory関数
      return &ItemMemoryRepository{items: items}
    }

    func (r *ItemMemoryRepository) FindAll() (*[]Item, error) {
      return &r.items, nil
    }
    ~~~

#### コンストラクタ(constructor)とは
- コンストラクタは、オブジェクト指向プログラミング（OOP）において、クラスのインスタンス（オブジェクト）が作成される際に自動的に呼び出される特別なメソッドまたは関数のこと。  
  コンストラクタの主な目的は、新しいオブジェクトの初期化を行うこと。これには、メンバ変数（プロパティ）の初期値の設定や、オブジェクト作成時に必要なリソースの割り当てなどが含まれる。