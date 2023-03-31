- まずLinuxにDjangoをインストールする
- 以下のようにDjango設定をしておく
  ~~~
  django-admin startproject app
  cd app
  vi app/settings.py
  ~~~
  - Databaseの部分を以下のように書き換える
    ~~~python
    DATABASES = {
        'default': {
            'ENGINE': 'django.db.backends.postgresql',
            'NAME': '<DBテーブル名(e.g. postgres)>',
            'USER': 'postgres',
            'PASSWORD': '<postgresのPW>',
            'HOST': '<Auroraのエンドポイント(e.g. django-postgresql.cluster-*****.ap-northeast-1.rds.amazonaws.com)>',
            'PORT': '5432',
        }
    }
    ~~~
  - `ALLOWED_HOSTS`の部分を`[]`→`["*"]`に修正する
- `python3 manage.py startapp <アプリ名>`でアプリを作成
- `views.py`などアプリのコード/設定を作成
- (manage.pyがあるディレクトリで)`Dockerfile`と`requirements.txt`を以下の通り作成する
  - Dockerfile
    ~~~
    FROM python:3.11-alpine
    ENV PYTHONUNBUFFERED 1

    WORKDIR /app
    COPY requirements.txt .
    RUN apk add g++ python3-dev gcc openldap-dev
    RUN pip install -r requirements.txt
    COPY . .
    CMD python manage.py runserver 0.0.0.0:80
    ~~~
  - requirements.txt (他にも必要なPackageがあれば追加)
    ~~~
    Django==4.1
    psycopg2-binary>=2.8.6 ## Python3.9以降では2.8.6以上でないといけないらしい(https://qiita.com/tamanobi/items/18a46fb8614b53d2fb6c)
    ~~~
- `docker build -t django:v1 .`
- `docker run -d --name django -p 8080:80 django:v1`