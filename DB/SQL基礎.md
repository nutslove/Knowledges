## PostgreSQL
- tableのカラムとデータ型の確認方法
  - `SELECT * FROM information_schema.columns WHERE table_bane='<確認したTable名>';`
- table内の制約確認方法
  - `SELECT table_name, constraint_name, constraint_type FROM information_schema.table_constraints WHERE table_name='<確認したTable名>';`
- tableの削除
  - `DROP TABLE <削除したいTable名> [,<削除したいTable名>,<削除したいTable名>,・・・];`
- 新しいレコードの挿入(INSERT)
  - `INSERT INTO <対象Table名> [('カラム名','カラム名'[,'カラム名',・・・])] VALUES ('<値>','<値>'[,'<値>','<値>',・・・]);`
  - `INSERT INTO <対象Table名> [('カラム名','カラム名'[,'カラム名',・・・])] SELECT文;'`
- 既存レコードの更新(UPDATE)
  - `UPDATE <対象Table名> SET <カラム名> = <更新後の値> WHERE <条件式>;`
  - `UPDATE <対象Table名> SET ( <カラム名>, <カラム名> [,<カラム名>,・・・])  = ( <更新後の値>, <更新後の値> [,<更新後の値>,・・・] ) WHERE <条件式>;`