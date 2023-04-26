- `<Appディレクトリ>/templates/<App名>`でディレクトリを作成し、そのディレクトリの中にhtmlを格納
  - privilegeというapplicationを作成しindex.htmlを格納した場合、`privilege/templates/privilege/index.html`
- 参考URL
  - https://docs.djangoproject.com/ja/4.1/intro/tutorial03/#a-shortcut-render
- `views.py`では`return render(request, <templates以降のパス>)`を指定
  ~~~python
  def home(request):
    return render(request, 'privilege/index.html')
  ~~~

- htmlのtemplateの中のif文にandとorを混ぜて使うことも可能
  - **`and`の方が優先される**
    - 以下のif文は`if (athlete_list and coach_list) or cheerleader_list`と同じ
      - `{% if athlete_list and coach_list or cheerleader_list %}`
  - https://docs.djangoproject.com/ja/4.2/ref/templates/builtins/#boolean-operators