## テンプレート(ファイル)のディレクトリ
- Djangoではテンプレートの配置戦略(?)は主に2つある
  1. アプリケーションごとに、アプリケーションのディレクトリ内に配置
  2. Project全体で共有するテンプレートとしてProjectルートに配置

### 1. アプリケーションごとに、アプリケーションのディレクトリ内に配置
- `<Appディレクトリ>/templates/<App名>`でディレクトリを作成し、そのディレクトリの中にhtmlを格納
  - *myapp*というapplicationを作成した場合  
    ```csharp
    myproject/
        myapp/
            templates/
                myapp/
                    base.html
                    index.html
    ```
- **この構成の場合、`settings.py`の`TEMPLATES`設定で`APP_DIRS`を`True`にするだけで、Djangoは自動的に各アプリケーションのtemplatesディレクトリを探す**
- 参考URL
  - https://docs.djangoproject.com/ja/4.1/intro/tutorial03/#a-shortcut-render
- `views.py`では`return render(request, <templates以降のパス>)`を指定
  ~~~python
  def home(request):
    return render(request, 'privilege/index.html')
  ~~~

### 2. Project全体で共有するテンプレートとしてProjectルートに配置
- `templates`ディレクトリをプロジェクトのルートに置く
  ```csharp
  myproject/
      myapp/
      myproject/
          __init__.py
          settings.py
          urls.py
          wsgi.py
      templates/
          base.html
          index.html
      manage.py
  ```
- **プロジェクト全体のテンプレートディレクトリを指定するには、`settings.py`の`TEMPLATES`設定の`DIRS`オプションに`BASE_DIR/'templates'`の設定が必要**
  ```python
  # settings.py
  from pathlib import Path

  BASE_DIR = Path(__file__).resolve().parent.parent

  TEMPLATES = [
      {
          'BACKEND': 'django.template.backends.django.DjangoTemplates',
          'DIRS': [BASE_DIR / 'templates'],  # プロジェクト全体のテンプレートディレクトリ
          'APP_DIRS': True, # アプリケーション固有(ごと)のテンプレートディレクトリも検索
          'OPTIONS': {
              'context_processors': [
                  'django.template.context_processors.debug',
                  'django.template.context_processors.request',
                  'django.contrib.auth.context_processors.auth',
                  'django.contrib.messages.context_processors.messages',
              ],
          },
      },
  ]
  ```

## テンプレート内での条件文
- htmlのtemplateの中のif文にandとorを混ぜて使うことも可能
  - **`and`の方が優先される**
    - 以下のif文は`if (athlete_list and coach_list) or cheerleader_list`と同じ
      - `{% if athlete_list and coach_list or cheerleader_list %}`
  - https://docs.djangoproject.com/ja/4.2/ref/templates/builtins/#boolean-operators

## 