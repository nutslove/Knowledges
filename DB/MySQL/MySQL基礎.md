# DB操作
- DB接続
  - `mysql -p` -> password入力後ログイン
- DB作成
  - `create database <DB名>;`
- DB削除
  - `drop database <DB名>;`
- DB一覧確認
  - `show databases;`
- DB切り替え
  - `use <DB名>;`
- Table一覧確認
  - `show tables;`
- Table作成
  - `create table <Table名>(<カラム名> <データ型> [制約][,<カラム名> <データ型>])`
  - 例１  
    `create table test(id INT, name VARCHAR(10));`
  - 例２  
    ```
    CREATE TABLE customers (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(100) UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    ```
  - 例３  
    ```
    CREATE TABLE orders (
        order_id INT AUTO_INCREMENT PRIMARY KEY,
        customer_id INT,
        order_date DATE,
        total_amount DECIMAL(10, 2),
        FOREIGN KEY (customer_id) REFERENCES customers(id)
    );
    ```
  - 例４  
    ```
    CREATE TABLE products (
        product_code VARCHAR(20),
        product_name VARCHAR(100) NOT NULL,
        category VARCHAR(50),
        price DECIMAL(8, 2),
        stock INT DEFAULT 0,
        PRIMARY KEY (product_code, product_name)
    );
    ```