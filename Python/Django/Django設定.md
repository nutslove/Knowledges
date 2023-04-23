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
- Templatesディレクトリを作成
  - `<Project名>/<アプリケーション名>/templates/<アプリケーション名>`
    - e.g. Project名がidでAPP名がprivilegeの場合のTemplatesディレクトリ → `id/privilege/templates/privilege/`

## `settings.py`
https://docs.djangoproject.com/en/4.1/ref/settings/#databases
- `TIME_ZONE`を`UTC`→`Asia/Tokyo`に変更
- `LANGUAGE_CODE`を`en-us`→`ja`に変更
- `INSTALLED_APPS`に`manage.py startapp`で作成したApp名を追加する
  > INSTALLED_APPS → A list of strings designating all applications that are enabled in this Django installation.

## <Projectフォルダ>/`urls.py`
- リクエストを受け付けるURLの定義
- Project全体のURLとAppのURLのマッピングを定義
  - `urlpatterns`中に`path('<ProjectのURL>',include('<マッピングするApp名.urls>'))`
- 例
  ~~~python
  from django.contrib import admin
  from django.urls import path, include ## includeはデフォルトではないので追加でimport

  urlpatterns = [
    path('', include('privilege.urls')), ## root URLとprivilegeというAppとマッピング
    path('admin/', admin.site.urls), ## デフォルトでぞんざいするadmin画面
  ]
  ~~~

## <Appフォルダ>/`urls.py`
- URLとviews.pyの関数/Classのマッピングを定義
- デフォルトでは存在しないので作成する必要がある
- <Projectフォルダ>/`urls.py`に定義しているURL Path＋<Appフォルダ>/`urls.py`に定義しているURL PathでViewのURL Pathが決まる
  - 例えば上のProjectのurls.pyの例と以下のAppのurls.pyがある場合、ブラウザで`/test`にアクセスした時、privilegeアプリの中の`views.py`の`test`関数が実行される
    ~~~python
    from django.urls import path

    from . import views

    urlpatterns = [
      path('', views.index, name='index'),
      path('test', views.test, name='test'),
    ]
    ~~~

## `views.py`
#### 関数で定義する方法と、Classで定義する方法がある
##### ■ 関数で定義する方法
- Format
  ~~~
  def <関数名>(request):
    <処理>
  ~~~
- `views.py`の`<関数名>`は`urls.py`の`path('<ULRパス>',views.<関数名>)`と一致させる必要がある