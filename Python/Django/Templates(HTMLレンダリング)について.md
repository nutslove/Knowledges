- `<Appディレクトリ>/templates/<App名>`でディレクトリを作成し、そのディレクトリの中にhtmlを格納
  - privilegeというapplicationを作成しindex.htmlを格納した場合、`privilege/templates/privilege/index.html`
- 参考URL
  - https://docs.djangoproject.com/ja/4.1/intro/tutorial03/#a-shortcut-render
- `views.py`では`return render(request, <templates以降のパス>)`を指定
  ~~~python
  def home(request):
    return render(request, 'privilege/index.html')
  ~~~

- css,image,javascriptファイルは、静的ファイルはstaticディレクトリ配下に配置
  - **imageファイル(staticディレクトリ)はtemplatesと同様に`<Appディレクトリ>/static/<App名>`を作成し、その配下に格納**
    - privilegeというapplicationを作成しicon.jpgを格納した場合、`privilege/static/privilege/icon.jpg`
  - **cssファイルは<App名>から更に`css`ディレクトリを作成し、`css`配下格納**
    - login.cssを格納した場合、`privilege/static/privilege/css/login.css`

- htmlのtemplateの中のif文にandとorを混ぜて使うことも可能
  - **`and`の方が優先される**
    - 以下のif文は`if (athlete_list and coach_list) or cheerleader_list`と同じ
      - `{% if athlete_list and coach_list or cheerleader_list %}`
  - https://docs.djangoproject.com/ja/4.2/ref/templates/builtins/#boolean-operators