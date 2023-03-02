- Python3.11/Django4.1のインストール
  - 参考URL：https://tech.fragment.co.jp/infrastructure/server/amazon-linux-2%E3%81%ABpython3%E3%82%92%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB/
  ~~~
  yum remove -y openssl openssl-devel
  yum install -y openssl11 openssl11-devel
  vi /root/.bash_profile
  export CFLAGS=$(pkg-config --cflags openssl11)
  export LDFLAGS=$(pkg-config --libs openssl11)
  :wq
  yum groupinstall -y "Development Tools"
  yum install -y kernel-devel kernel-headers bzip2-devel libdb-devel libffi-devel
  yum install -y gdbm-devel xz-devel ncurses-devel readline-devel tk-devel uuid-devel
  wget https://zlib.net/zlib-1.2.13.tar.gz
  tar -xzvf zlib-1.2.13.tar.gz
  cd zlib-1.2.13
  ./configure
  make -j4
  make install
  wget https://www.python.org/ftp/python/3.11.2/Python-3.11.2.tgz
  tar -xzvf Python-3.11.2.tgz
  cd Python-3.11.2
  ./configure
  make
  make altinstall
  ln -sf /usr/local/bin/python3.11 /usr/bin/python3
  ln -sf /usr/local/bin/pip3.11 /usr/bin/pip3

  python3 -m venv django4
  source django4/bin/activate
  pip3 install django==4.1
  python3 -m django --version
  ~~~
- AWS Aurora PostgreSQLを作成し、接続エンドポイントとpostgresユーザのPWを押さえておく
- postgres利用のための設定
  ~~~
  yum install python-psycopg2
  yum install postgresql-devel
  pip3 install psycopg2
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
  - Djangoが実行できることを確認する
    `python3 manage.py runserver 0.0.0.0:8080`
