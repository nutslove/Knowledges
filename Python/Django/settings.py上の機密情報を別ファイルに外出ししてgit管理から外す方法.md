- 参考URL
  - https://chigusa-web.com/blog/django-secret/

## 前提
- `settings.py`には`SECRET`など機密情報が書いてあり、そのままGitに上げるわけにないかない
- 別ファイルを作成して、そこに`SECRET`などの機密情報を移して`.gitignore`にその別ファイルを記載してgitには連携されないようにする

## 別ファイルをimportする方法
- Djangoは`manage.py`ファイルがあるディレクトリをルートディレクトリと認識するため、`settings.py`と同じ階層に外だしファイルを作成した場合、`settings.py`では`import <プロジェクト名.外出しファイル名>`でimportする必要がある。
- 例えばProject名が`app`で外出しファイル`settings_local.py`が`settings.py`と同じ階層にあるとした場合の`settings.py`でのimportは以下になる
  ~~~python
  import app.settings_local as secret
  SECRET = secret.SECRET
  ~~~

## `.gitignore`について
- gitに連携したくないファイルがあるディレクトリに`.gitignore`ファイルを作成し、中にgitに連携しないファイル名を記載するとgitに連動されない
- 正規表現なども使える
  - https://qiita.com/inabe49/items/16ee3d9d1ce68daa9fff