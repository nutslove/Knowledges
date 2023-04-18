- https://docs.djangoproject.com/en/4.2/topics/http/sessions/
- https://developer.mozilla.org/ko/docs/Learn/Server-side/Django/Sessions
- https://ssungkang.tistory.com/entry/Django%EB%A1%9C%EA%B7%B8%EC%9D%B8-%EC%9C%A0%EC%A7%80%ED%95%98%EA%B8%B0-%EC%BF%A0%ED%82%A4%EC%99%80-%EC%84%B8%EC%85%98
- https://sinyblog.com/django/user_session/
- https://valuefactory.tistory.com/708

- デフォルトではユーザセッションはDBで管理されるけど、ファイルやCacheで管理することもできる
  - https://docs.djangoproject.com/en/4.2/topics/http/sessions/#configuring-the-session-engine
    > By default, Django stores sessions in your database (using the model django.contrib.sessions.models.Session). Though this is convenient, in some setups it’s faster to store session data elsewhere, so Django can be configured to store session data on your filesystem or in your cache.
- Session IDは`request.session.session_key`に入る
  - https://ssungkang.tistory.com/entry/Django%EB%A1%9C%EA%B7%B8%EC%9D%B8-%EC%9C%A0%EC%A7%80%ED%95%98%EA%B8%B0-%EC%BF%A0%ED%82%A4%EC%99%80-%EC%84%B8%EC%85%98
- Sessionの設定は`set_cookie("<セッションKey名>", <セッションValue>, None|max_age=<セッションのTTL>)`
  - e.g. `response.set_cookie("user_id", user_id, max_age=600)`
- Sessionの取得は`request.COOKIES.get('<セッションKey名>')`
  - e.g. `request.COOKIES.get('user_id')`

### 手順
- `settings.py`の`INSTALLED_APPS`と`MIDDLEWARE`に以下を追加(すでに追加されている場合は不要)
  ~~~python
  INSTALLED_APPS = [
    ...
    'django.contrib.sessions',
    ....

  MIDDLEWARE = [
    ...
    'django.contrib.sessions.middleware.SessionMiddleware',
    ....
  ~~~
- run `manage.py migrate` to install the single database table that stores session data.