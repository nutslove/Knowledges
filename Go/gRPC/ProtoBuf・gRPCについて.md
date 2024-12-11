## Protocol Buffers（protobuf）
- **スキーマ言語**であり、gRPCのデータフォーマットとして使われる。
  - 「*スキーマ言語*」とは、要素や属性などの構造を定義するための言語
- 様々なプログラミング言語に対応
- **バイナリ形式**でデータのサイズが小さく、通信コストとストレージコストを削減できる。
- データ構造(スキーマ)は **`.proto`** ファイルで定義され、このスキーマを基にしてデータが(バイナリ形式に)**シリアライズ**(およびデシリアライズ)される。これにより、データの一貫性と明確な構造が保たれる。
  - *シリアライズ(Serialization)*
    - データ構造やオブジェクト状態を、保存や送信に適したフォーマット（通常はバイト列）に変換するプロセス。このプロセスは、メモリ内のデータ構造をファイル、データベース、またはネットワークを通じて送信できる形式に変換するために使われる
    - シリアライズの例
      - オブジェクトの状態をJSONやXMLのようなテキストベースのフォーマットに変換
      - Protocol BuffersやApache Avroのようなバイナリフォーマットを使用して、データをよりコンパクトにする
  - *デシリアライズ(Deserialization)*
    - デシリアライズは、シリアライズの逆プロセスであり、保存されたデータやネットワークを介して受信したデータを、元のデータ構造やオブジェクトに再構築するプロセス。これにより、データを使ってプログラム内で操作が可能になる。
    - デシリアライズの例
      - JSONやXMLファイルを読み込み、それをプログラム内のオブジェクトに変換
      - バイナリデータを受信し、それをメモリ内のデータ構造に再構築
- **型**があり、**型安全性**が確保され、エラーの可能性が減少
- Jsonに変換することも可能
- `.proto`ファイルの1行目には`syntax = "proto3";`とprotobufのバージョンを指定
  - 記述しない場合はproto2バージョンが使用される
  - proto2とproto3は互換性がないので注意
- "`//`"（１行）と"`/* ~ */`"（複数行）でコメントを入れることもできる

### message
- **データ構造を定義するための基本的な単位**
- **一連のフィールドを持ち、それぞれのフィールドは特定の型（基本型や他のmessage型）を持つ**
  - それぞれのフィールドは**スカラ型**もしくは**コンポジット型**を設定することができる
- **各言語のコードとしてコンパイルすると、構造体やクラスとして変換される**
- Messageはバイナリ形式にシリアライズされ、簡単にデータ転送や保存ができる。  
  受信側ではデシリアライズされ、元のデータ構造に戻される。
- **型**、**フィールドの名前**、**識別番号(タグ番号)** を指定
- **protobufでは各フィールドは**フィールド名ではなく、**タグ番号によって識別される**
  - タグ番号は一意である必要がある
  - 19000 ~ 19999はProtobufによって予約されていて使用不可
  - 連番にする必要はない
- 1つのprotoファイルに複数の`message`を定義することも可能
- messageの例
  ~~~html
  syntax = "proto3"; --> 末尾に";"が必要

  message Person {
    string name = 1; <span style="color: red;">★ --> 1は値の代入ではなくタグ</span>
    int32 id = 2;
    string email = 3;

    enum PhoneType {
      MOBILE = 0;
      HOME = 1;
      WORK = 2;
    }

    message PhoneNumber {
      string number = 1;
      PhoneType type = 2;
    }

    repeated PhoneNumber phones = 4;
  }
  ~~~
- messageをネストさせることもできる
  - 例
    ~~~
    message Person {
      string name = 1;
      int32 id = 2;

      message Address {
        string street = 1;
        string city = 2;
        string state = 3;
        string country = 4;
      }

      Address address = 3;
    }
    ~~~
  - ネストされたメッセージは、他のメッセージからアクセスする場合は、完全な修飾名（上記の例では`Person.Address`）を使用

##### ■ **`enum`**
- 列挙型
  - 列挙した値のいずれかであることを要求する型
  - 特定のフィールドで許可される値の範囲を制限
  - 固定された一連の定数値を含む特別なデータ型
- 型の定義は不要
- すべて大文字で定義
- messageとは異なり、数字はタグ番号ではなく、実際の値(定数)
- 名前で識別される
- **必ず最初にデフォルト値の`0`を定義する必要がある**
  - `0`以外の`1`などの特定の値をスキップすることは可能
  - `0`は慣例的に`UNKNOWN`にすることが多い
##### ■ **`repeated`フィールド**
- 配列やリストのような動作をするフィールド
- 文法
  - `repeated <型> <フィールド名> = <タグ番号>;`
- `repeated`フィールドは、同じ型の要素を複数持つ
- 要素の数は動的であり、メッセージが使用される際に任意の数の要素を含めることができる
- プログラミング言語によっては、`repeated`フィールドは配列やリストとして表現され、その要素にはインデックスでアクセスできる
- 例
  ~~~
  message Person {
    repeated string phone_numbers = 1;
  }
  ~~~
- **`repeated`フィールドの注意点**
  - repeatedフィールドの要素は、シリアライズされたメッセージ内で定義された順序で格納される
  - 要素が0個(空のリスト)の場合、repeatedフィールドはシリアライズされたメッセージに含まれない
  - repeatedフィールドの各要素は、同じタグ番号を使用してエンコードされる
##### ■ `map`フィールド
- keyとvalueを持つフィールド
- 文法
  - `map<<keyの型>, <valueの型>> <フィールド名> = <タグ番号>;`
- keyで使える型は`string`, `int32`, `bool`
- `map`フィールド内のエントリの順序は保証されない
- 同じkeyを持つエントリは1つだけ存在できる  
  もし同じkeyで複数の値が設定された場合、最後に設定された値が保持される
- 例
  ~~~
  message Person {
    map<string, string> phones = 1;
  }
  ~~~

##### ■ `oneof`ブロック
- `oneof`ブロック内の各フィールドは排他的であり、複数のフィールドの中から一つのフィールドだけ値を持つことができる
- 例えば、以下の例では`name`フィールドが値を持つと、残りの`id`と`is_active`フィールドは値を持つことができない
  ~~~
  message SampleMessage {
    oneof test_oneof {
      string name = 1;
      int32 id = 2;
      bool is_active = 3;
    }
  }
  ~~~
- 使用されていないフィールドにはメモリが割り当てられず。メモリ使用量を節約できる
- １つの`oneof`ブロック内に異なる型のデータを定義できるため、より柔軟なデータ構造を定義できる

##### 各型のデフォルト値
- `string`
  - 空の文字列
- `bytes`
  - 空のbyte
- `bool`
  - `false`
- 整数型/浮動小数点数
  - 0
- 列挙型(`enum`)
  - タグ番号0の値
- `repeated`
  - 空のリスト

### `import`と`package`
#### ■ `import`
- `import`ステートメントを使用して、他の`.proto`ファイルの中に定義されているmessageを使うことができる
- 例
  - importされる側
    ~~~
    // address.proto
    message Address {
      string street = 1;
      string city = 2;
      string state = 3;
      string country = 4;
    }
    ~~~
  - importする側
    ~~~
    // person.proto
    import "address.proto";

    message Person {
      string name = 1;
      int32 id = 2;
      Address address = 3;
    }
    ~~~
- `import`ステートメントでは、importするファイルの相対パスまたは絶対パスを指定

#### ■ `package`
- `package`ステートメントを使っている場合、importする側はimport元のmessageを使う時に`<import元package名>.<message名>`で指定する必要がある
- 例
  - importされる側
    ~~~
    // address.proto
    package address

    message Address {
      string street = 1;
      string city = 2;
      string state = 3;
      string country = 4;
    }
    ~~~
  - importする側
    ~~~
    // person.proto
    package person

    import "address.proto";

    message Person {
      string name = 1;
      int32 id = 2;
      address.Address address = 3; --> ここ
    }
    ~~~

### protoファイルのコンパイル
- `.proto`ファイルは、protobufコンパイラ`protoc`で特定のプログラミング言語用のソースコードにコンパイルする必要がある
  - protobufは多言語対応であり、`.proto`ファイルから、特定のプログラミング言語に適したソースコードを生成する必要がある
- golangの場合
  - `protoc --go_out=<outputディレクトリ> <inputとなるprotoファイル> [<inputとなるprotoファイル2> <inputとなるprotoファイル3> ・・・]`
  - golangはプラグインで追加する必要がある
  - コンパイルに成功したらgoファイル生成される
  - messageで定義した内容はgoの`struct`に変換される
- pythonの場合
  - `protoc --python_out=<outputディレクトリ> <inputとなるprotoファイル>`
- inputファイルは`*.proto`のように複数指定することもできる
- `.proto`ファイルの中にて`import`しているものは`protoc`実行時`-I`オプションで指定する必要がある
  - `protoc -I<importする.protoファイルがある(絶対/相対)パス> --go_out=<outputディレクトリ> <inputとなるprotoファイル>`
  - 複数の`-I`オプションを使用できる。または`:`区切り(e.g. `-I./test:./dev`)で複数のパスも記述できる
  - `-I`オプションを省略した場合はカレントディレクトリ`-I.`が設定される
- golangの場合、`.proto`ファイルに`option go_package = <パッケージ名>`オプションでGoのパッケージ名を指定する必要がある。  
  これは、生成されたGoファイル内での`package`ステートメントに反映される。

## gRPC
### gRPC開発の流れ
1. protoファイルを作成
2. protoファイルをコンパイルし、サーバ / クライアントの雛形コードを作成
3. 雛形コードを使用してサーバ / クライアントを実装

### gRPCの通信方式
1. **Unary RPC**
   - 1リクエスト・1レスポンス方式
   - 通常の関数コールのような扱い
   - Unary RPCのService定義
     ~~~
     ~~~
2. **Server Streaming RPC**
   - 1リクエスト・複数レスポンス方式
   - クライアントはサーバから送信完了の信号が送信されるまで、ストリームのメッセージを読み続ける
   - サーバからのプッシュ通知などで使われる
   - Server Streaming RPCのService定義
     ~~~
     ~~~
3. **Client Streaming RPC**
   - 複数リクエスト・1レスポンス方式
   - サーバはクライアントからリクエスト完了の信号が送信されるまで、ストリームメッセージを読み続け、レスポンスを返さない
   - 大きなファイルのアップロードなどに使われる
   - Client Streaming RPCのService定義
     ~~~
     ~~~
4. **Bidirectional Streaming RPC**
   - 複数リクエスト・複数レスポンス方式
   - クライアントとサーバのストリームが独立していて、リクエストとレスポンスに順序は守らなくても良い
     - クライアントから複数リクエストを送り終わった後にサーバから複数レスポンスを返すのではなく、クライアントが送信している間にサーバからレスポンスを返しても良い
     - チャットなどで使われる
   - Client Streaming RPCのService定義
     ~~~
     ~~~
