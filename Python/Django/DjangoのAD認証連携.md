- `yum install python-devel openldap-devel`
- `pip3 install django-auth-ldap`
- django-auth-ldapはsettings.pyの`INSTALLED_APPS`への追記が不要
- 参考URL
  - https://techexpert.tips/django/django-ldap-authentication-active-directory/
  - https://coderbook.com/@marcus/how-to-add-ldap-and-active-directory-authentication-to-django/

#### `django-auth-ldap`インストール後の設定
- `settings.py`にて以下を追加
  ~~~python
  
  AUTHENTICATION_BACKENDS = [
    "django.contrib.auth.backends.ModelBackend",
    "django_auth_ldap.backend.LDAPBackend",
  ]
  ~~~

#### `pip3 install django-auth-ldap`で`ModuleNotFoundError: No module named '_ctypes'`エラーが出る場合と対処
- インストールしたpython(e.g. /usr/local/bin/python3.11)とpip(e.g. /usr/local/bin/pip3.11)を削除する
- `yum -y install libffi-devel`でlibffi-develをインストールする
- 以下手順でPythonを再度インストールする
  ~~~
  wget https://www.python.org/ftp/python/3.11.2/Python-3.11.2.tgz
  tar -xzvf Python-3.11.2.tgz
  cd Python-3.11.2
  ./configure
  make
  make altinstall
  ln -sf /usr/local/bin/python3.11 /usr/bin/python3
  ln -sf /usr/local/bin/pip3.11 /usr/bin/pip3
  ~~~