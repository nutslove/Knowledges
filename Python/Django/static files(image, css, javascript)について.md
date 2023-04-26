- css,image,javascriptファイルは、静的ファイルはstaticディレクトリ配下に配置
  - **imageファイル(staticディレクトリ)はtemplatesと同様に`<Appディレクトリ>/static/<App名>`を作成し、その配下に格納**
    - privilegeというapplicationを作成しicon.jpgを格納した場合、`privilege/static/privilege/icon.jpg`
  - **cssファイルは<App名>から更に`css`ディレクトリを作成し、`css`配下格納**
    - login.cssを格納した場合、`privilege/static/privilege/css/login.css`
- **DEBUG=Trueにしている時は自動的にstatic filesを読み込んでくれるけど、DEBUG=Falseにすると手動で読み込まないといけない**

## static filesを読み込む方法 (DEBUG=Falseにした場合)
- 参考URL
  - **https://www.jujens.eu/posts/en/2021/Mar/31/manage-static-k8s-django/**
- 以下の３つの方法がある
  1. WhiteNoiseというライブラリを使う
     - https://whitenoise.readthedocs.io/en/stable/
     - https://whitenoise.readthedocs.io/en/stable/django.html
  2. static filesをS3など外部リポジトリに置き、そこから取得するようにする
  3. NginxなどWebサーバを使って、Webサーバの方に配置する
#### WhiteNoiseを使って設定する方法
- `pip install whitenoise`でwhitenoiseをインストール
- `settings.py`に`STATIC_ROOT = BASE_DIR / "staticfiles"`を追加
- `settings.py`の`MIDDLEWARE`にて`django.middleware.security.SecurityMiddleware`のすぐ下に`whitenoise.middleware.WhiteNoiseMiddleware`を追加
  ~~~python
  MIDDLEWARE = [
    # ...
    "django.middleware.security.SecurityMiddleware",
    "whitenoise.middleware.WhiteNoiseMiddleware",
    # ...
  ]
  ~~~
- `settings.py`に以下を追加
  ~~~python
  STORAGES = {
    # ...
    "staticfiles": {
        "BACKEND": "whitenoise.storage.CompressedManifestStaticFilesStorage",
    },
  }
  ~~~
  - v4.2より前のDjangoの場合は`STATICFILES_STORAGE = "whitenoise.storage.CompressedManifestStaticFilesStorage"`
- `python manage.py runserver`でDjangoを起動する前に`python manage.py collectstatic`を実行する
  - Dockerの場合CMDに`python manage.py collectstatic && python manage.py runserver 0.0.0.0:80`のように書けばOK
  > `python manage.py collectstatic` command collects all the static files from your various apps and puts them into a single directory (STATIC_ROOT) that you have specified in your settings file. If you haven't specified a STATIC_ROOT, Django will use a default directory called staticfiles.
