- あるページへのアクセスにログインを強制する方法
  - https://docs.djangoproject.com/ja/4.1/topics/auth/default/#the-login-required-decorator
  - `login_required`decoratorを使うことであるページへのアクセスの前にログインを強制することができる
    > login_required() は下記の処理を行います:
    > - もしユーザがログインしていなければ、settings.LOGIN_URL にリダイレクトし、クエリ文字列に現在の絶対パスを渡します。リダイレクト先の例: /accounts/login/?next=/polls/3/
    > - もしユーザがログインしていれば、通常通りビューを処理します。ビューのコードの中ではユーザがログインしているかを意識しなくて良いのです。

- htmlのform(post method)でログイン機能を実装した時、  
  IDやPWを入力せずにsubmit(ログイン)ボタンを押すと`django.utils.datastructures.MultiValueDictKeyError: '<入力パラメータ>'`エラーが出る。  
  `request.POST`の代わりに`request.POST.get`を使えばデフォルト値を設定できて`MultiValueDictKeyError`エラーを防ぐことができる
  ~~~python
  from django.shortcuts import render
  from django.contrib.auth.decorators import login_required
  from django.http import HttpResponse
  from django.contrib.auth import authenticate, login

  # Create your views here.
  def home(request):
      return render(request, 'privilege/index.html')

  def form(request):
      user_id = request.POST.get('ID', '') ## form(ID)に何も入力せずにsubmitした場合、default値として空白を設定
      user_pw = request.POST.get('PASSWORD', '') ## form(PASSWORD)に何も入力せずにsubmitした場合、default値として空白を設定
      if user_id == '' or user_pw == '':
          return HttpResponse("先にログインしてください！")
      else:
          user = authenticate(request, username=user_id, password=user_pw) ## authenticateでログイン実現
          print(user) ## userにはログインに成功したらユーザ名が、失敗したらNoneが入る
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
          return render(request, 'privilege/index_after_login.html', params)
  ~~~
  - `django.utils.datastructures.MultiValueDictKeyError: '<入力パラメータ>'`エラーについてChat-GPTの回答
    > The error you're getting is MultiValueDictKeyError, which means that the key you're trying to access in the POST dictionary does not exist. Specifically, the error message says that the key 'ID' is not found.
    >
    > There are a few possible reasons why this error could be happening. Here are some things you can check:
    >
    > 1. Make sure that the form you're submitting has a field with the name attribute set to 'ID'. If this field is missing or has a different name, then the POST dictionary will not contain the 'ID' key.
    >
    > 2. Check that the form method is set to POST in your HTML. If the form method is set to GET, then the data will not be submitted in the POST dictionary.
    >
    > 3. Verify that the form data is being submitted correctly. You can do this by adding print(request.POST) before the line that's causing the error. This will print out the entire POST dictionary so you can see what keys are present.
    >
    > 4. **Consider using the get method of the POST dictionary instead of the index notation. The get method allows you to specify a default value if the key is not found, which can help prevent MultiValueDictKeyError exceptions. For example, instead of request.POST['ID'], you can use request.POST.get('ID', '').**

## 手順
- 参考URL
  - https://bio-info.tistory.com/171
- accounts appを作成
  `python[3] manage.py startapp accounts`
- `settings.py`の`INSTALLED_APPS`に`accounts`を追加