## 初期設定
- Projectの作成  
  `django-admin startproject <Project名>`
- Appの作成（Projectディレクトリの中で実施）  
  `python[3] manage.py startapp <アプリケーション名>`
- スーパーユーザの作成（`/admin`で使うAdminユーザ）  
  `python[3] manage.py createsuperuser --username=admin --email=<メアド>`
  - 以下のようなエラーが出る場合は`python[3] manage.py migrate`を実行してDBの同期をとった後に再度試してみること ([参考URL](https://genchan.net/it/programming/python/django/11480/))
    > django.db.utils.ProgrammingError: relation "auth_user" does not exist  
    > LINE 1: ...user"."is_active", "auth_user"."date_joined" FROM "auth_user...
  - スーパーユーザのPW変更  
    `python[3] manage.py changepassword <スーパーユーザ名>`

## `settings.py`
- `TIME_ZONE`を`UTC`→`Asia/Tokyo`に変更
- `LANGUAGE_CODE`を`en-us`→`ja`に変更

## `urls.py`
- URLの追加

## `views.py`
