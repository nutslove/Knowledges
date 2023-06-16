- https://inuplace.tistory.com/602
- https://losskatsu.github.io/it-infra/django-inspectdb/#1-db-%ED%85%8C%EC%9D%B4%EB%B8%94-%EC%9E%A5%EA%B3%A0%EB%A1%9C-%EB%B6%88%EB%9F%AC%EC%98%A4%EA%B8%B0

## `python manage.py makemigrations`と`python manage.py migrate`は何をするもの？
- 以下Chat-GPT回答
> python manage.py makemigrations and python manage.py migrate are Django commands used for handling database schema changes. They are a part of Django's database migration system. Let's take a closer look at both commands.
> 
> 1. **`python manage.py makemigrations`:**
>``` 
> This command is responsible for creating new migrations based on the changes you have made to your Django models. 
> When you run it, Django checks all your models and compares them to the previous state of your models stored in migration files.
> If Django detects any changes, it will create new migration files, each of which is timestamped and contains the details of the changes.
>
> Each migration file is like a version control system for your database schema.
> makemigrations does not apply these changes to the database,
> but it generates the SQL commands needed to apply or unapply these changes.
>```
> 2. **`python manage.py migrate`:**
>```
> This command is responsible for applying and unapplying migrations.
> When you run it, Django will execute the SQL commands in the migration files to apply the changes to your database schema.
> This can include operations like creating new tables, adding fields to existing tables, or deleting fields.
>
> If you provide migrate with a migration name, it will apply all migrations up to and including that migration.
> If you don't provide a migration name, migrate will apply all unapplied migrations.
>
> If you want to unapply all migrations, you can run python manage.py migrate zero.
> This will unapply all migrations and leave your database schema in its initial state.
>```
>
> In summary, you use **`makemigrations`** to **create migrations for your changes**, and **`migrate`** to **apply those changes to your database**. *This two-step process gives you the opportunity to review the SQL that Django generates for your model changes before applying it to your database*.

## models.pyを修正した後にDBに反映する方法
- Djangoをコンテナで動かしてる場合はDjangoコンテナに入る
- `python manage.py makemigrations <アプリ名>`を実行
  - e.g. `python manage.py makemigrations privilege`
- `python manage.py migrate <アプリ名>`を実行
  - e.g. `python manage.py migrate privilege`
- DockerfileのCMDに`python manage.py makemigrations && python manage.py migrate`を含めて最初起動時にDBと同期をとるようにする

## views.pyでModelで定義したテーブルデータを操作する方法
- views.pyにて`from .models import <Model名>[, <Model名>, <Model名>,・・・]`でModelをimport
  - 例えば、作成したModelが「System」,「Dbuserpassword」,「Userprivilegestate」３つの場合、`from .models import System, Dbuserpassword, Userprivilegestate`
### DBデータ取得
- `get`と`filter`２種類がある
- 2つの違いについてChat-GPTからの回答
  > Djangoのモデルで使用される get と filter メソッドは、データベースからデータを取得するためのものですが、動作が若干異なります。
  >
  > ##### get メソッド：
  > - getは一つのレコードだけを返すメソッドです。つまり、getは一意の結果を期待します。
  > - 一意な結果が存在しない場合、つまり該当するレコードが存在しないか、複数のレコードが存在する場合、getはエラーを発生します。具体的には、該当するレコードがない場合はDoesNotExistエラー、複数のレコードが該当する場合はMultipleObjectsReturnedエラーを返します。
  > ##### filter メソッド：
  > - filterは一つ以上のレコードを返すことができるメソッドで、結果が複数でもエラーにはなりません。複数のオブジェクトが該当する場合、それらすべてを含むクエリセットが返されます。
  > - filterは該当するレコードがない場合でもエラーを発生させず、単に空のクエリセットを返します。
  >  
  > 例えば、特定のユーザーを名前で検索する場合、その名前がユニーク（一意）であることが分かっているなら get メソッドを使用するのが良いでしょう。しかし、特定の条件を満たすすべてのユーザー（例えば、特定の都市に住んでいる全てのユーザー）を取得したい場合には、 filter メソッドを使用するのが良いです。
#### `get`
- `<Model名>.objects.get(<カラム名>=<検索値>)`
- 条件にマッチするレコードが複数(例えば2レコード)ある場合、以下のようなエラーが返ってくる
  - `get() returned more than one Userprivilegestate -- it returned 2!`
- **https://office54.net/python/django/orm-database-operate**
- https://qiita.com/NOIZE/items/a50afe3af644a55d37e7
#### `filter`
- `<Model名>.objects.filter(<カラム名>=<検索値>)`


### DjangoのModelとDBデータ型のマッピング
- https://qiita.com/okoppe8/items/13ad7b7d52fd6a3fd3fc

### Modelで`primary_key=True`を指定しない場合、Djangoが自動的に`id`というPrimary Keyを作成する
- https://docs.djangoproject.com/en/4.2/topics/db/models/#automatic-primary-key-fields
- `SELECT nextval('<テーブル名>_id_seq');`で次に振られるid番号を確認できる
  - `privilege_userprivilegestate`というテーブルで次のidを取得する場合、`SELECT nextval('privilege_userprivilegestate_id_seq');`
- 以下のように`primary_key=True`のないModelで作成した時、作成されるDBテーブル例
  - Django側(models.py)
    ~~~python
    from django.db import models

    # Create your models here.
    class Userprivilegestate(models.Model):
      userid = models.CharField(max_length=30, help_text='ユーザID')
      combinationid = models.ForeignKey(Dbuserpassword, on_delete=models.CASCADE)
      endtimestamp = models.DateTimeField(help_text='特権利用終了日時')

      class Meta:
        constraints = [
            models.UniqueConstraint(fields=['userid', 'combinationid'], name='unique_user')
        ]
    ~~~
  - DB側
    ~~~
    postgres=> \d privilege_userprivilegestate
                              Table "public.privilege_userprivilegestate"
      Column      |           Type           | Collation | Nullable |             Default
    ------------------+--------------------------+-----------+----------+----------------------------------
    id               | bigint                   |           | not null | generated by default as identity
    userid           | character varying(30)    |           | not null |
    endtimestamp     | timestamp with time zone |           | not null |
    combinationid_id | character varying(30)    |           | not null |
    Indexes:
        "privilege_userprivilegestate_pkey" PRIMARY KEY, btree (id)
        "unique_user" UNIQUE CONSTRAINT, btree (userid, combinationid_id)
        "privilege_userprivilegestate_combinationid_id_d9d618ef" btree (combinationid_id)
        "privilege_userprivilegestate_combinationid_id_d9d618ef_like" btree (combinationid_id varchar_pattern_ops)
    Foreign-key constraints:
        "privilege_userprivil_combinationid_id_d9d618ef_fk_privilege" FOREIGN KEY (combinationid_id) REFERENCES privilege_dbuserpassword(combinationid) DEFERRABLE INITIALLY DEFERRED
    ~~~

### Djangoには復号主キー機能はないらしい
- https://zenn.dev/shimakaze_soft/scraps/22dcea1acd133a
- その代わりに、複合ユニーク制約の機能を使う

## Modelから作成されたTableを削除した場合、再migrationする方法
#### 正攻法
1. `django_migrations`テーブルからアプリ名のレコードを削除
   - 削除したいアプリのidが19の場合
     - `delete from django_migrations where id = 19;`
2. Modelから作成したテーブルを`DROP TABLE <テーブル名>;`ですべて削除する
3. `<アプリ名>/migrations`フォルダをフォルダごと削除する
4. `python manage.py makemigrations <アプリ名>`を実行
5. `python manage.py migrate <アプリ名>`を実行
 
- 参考URL
  - https://stackoverflow.com/questions/33259477/how-to-recreate-a-deleted-table-with-django-migrations

#### Modelに定義されているTableをすべて手動で削除してから再度migrationする方法
1. Modelに定義されているDB上のTableを手動ですべて削除する
2. `python manage.py makemigrations <アプリ名>`
3. `python manage.py migrate <アプリ名> zero --fake`
4. `python manage.py migrate <アプリ名>`
- 上記のオプション`python manage.py migrate <アプリ名> zero`コマンドと`--fake`オプションについて
  - `python manage.py migrate <アプリ名> zero`コマンド
    > 指定されたアプリケーションの全てのマイグレーションをロールバック（つまり、Undo）します。zeroは、適用するべきマイグレーションの数を0にする、という意味です。そのため、このコマンドを実行すると、そのアプリケーションのデータベーススキーマは、まったくマイグレーションが適用されていない状態に戻ります。
  - `--fake`
    > Djangoに対して、指定されたマイグレーションをデータベースに適用することなく、そのマイグレーションが適用されたと記録するよう指示します。つまり、Djangoはそのマイグレーションが適用されたという記録を保持しますが、実際のデータベーススキーマは変更されません。