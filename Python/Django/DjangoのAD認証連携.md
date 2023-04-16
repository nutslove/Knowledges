- `yum install python-devel openldap-devel`
- `pip3 install django-auth-ldap`
- django-auth-ldapはsettings.pyの`INSTALLED_APPS`への追記が不要
- 参考URL
  - https://coderbook.com/@marcus/how-to-add-ldap-and-active-directory-authentication-to-django/
  - https://github.com/shalomz/HowTo/blob/master/Django%20with%20AD%20(django-auth-ldap).md
  - https://techexpert.tips/django/django-ldap-authentication-active-directory/
- **ADはログインIDとして`sAMAccountName`を使用する**

#### `django-auth-ldap`インストール後の設定
- `settings.py`にて以下を追加
  ~~~python
              ・
              ・
              ・
  #### AD認証のための設定 Start ####
  AUTHENTICATION_BACKENDS = [
    "django.contrib.auth.backends.ModelBackend",
    "django_auth_ldap.backend.LDAPBackend",
  ]

  import ldap
  from django_auth_ldap.config import LDAPSearch, NestedActiveDirectoryGroupType

  AUTH_LDAP_SERVER_URI = 'ldap://<ADのIPまたはドメイン>:389'
  AUTH_LDAP_BIND_DN = '<BINDユーザのDN>'
  AUTH_LDAP_BIND_PASSWORD = '<BINDユーザのPW>'
  AUTH_LDAP_USER_SEARCH = LDAPSearch(
      '<ユーザが格納されるディレクトリ(e.g. OU=Users,OU=lee,DC=lee,DC=test,DC=ad)>',
      ldap.SCOPE_SUBTREE,
      '(sAMAccountName=%(user)s)',
  )
  AUTH_LDAP_USER_ATTR_MAP = {
      'username': 'sAMAccountName',
      'first_name': 'givenName',
      'last_name': 'sn',
      'email': 'mail',
  }
  #### AD認証のための設定 End   ####
              ・
              ・
              ・
  ~~~
- `views.py`にて`from django.contrib.auth import authenticate, login`し、`authenticate`でログイン
- 認証に成功したら以下例の`user`にADユーザ名が入って、失敗したら`None`が入る
  - 設定例
  ~~~python
  from django.contrib.auth import authenticate, login

  def form(request):
      user_id = request.POST['ID']
      user_pw = request.POST['PASSWORD']
      user = authenticate(request, username=user_id, password=user_pw)
      print(user)
      if user is not None:
          login(request, user)

          params = {
                  'ID':user_id,
                  'PW':user_pw,
          }
      else:
          params = {
                  'ID':'Invalid ID',
                  'PW':'Invalid PW',
          }
      return render(request, 'privilege/login.html', params)
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