- Amazon Linux 2023にはデフォルトでPython3.9がインストールされている
~~~
yum groupinstall -y "Development Tools"
pip3 install django==4.1
yum install python3-devel
yum install python-psycopg2
yum install postgresql-devel
pip3 install psycopg2
yum install libffi-devel
~~~ 
- Django設定
  ~~~
  django-admin startproject mysite ※mysiteは任意の名前
  cd mysite/mysite
  vi settings.py
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
  - Djangoが実行できることを確認する
    `python3 manage.py runserver 0.0.0.0:8080`
