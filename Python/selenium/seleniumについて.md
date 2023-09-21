- `pip install selenium`でインストール
  - `pip install selenium==3.141.0`のようにバージョン指定も可能
- SeleniumはWebDriverを仲介してブラウザを操作するため、WebDriverのインストールが必要。  
  ※WebDriverは各ブラウザの固有のもの(e.g. Chrome, Edge)を用意する必要がある。
- `webdriver.<ブラウザ種類>()`で対象ブラウザのドライバーを読み込んで、用意されているメソッドで操作する
  - `Options()`メソッドでブラウザのオプションの設定も可能
  - 設定例
    ~~~python
    from selenium import webdriver
    from selenium.webdriver.edge.service import Service
    from selenium.webdriver.edge.options import Options
    from selenium.webdriver.common.by import By
    from selenium.webdriver.common.keys import Keys
    from selenium.webdriver.common.action_chains import ActionChains

    # Edgeの設定を調整
    options = Options()
    options.add_experimental_option("prefs", {
      "download.default_directory": DOWNLOAD_DIR,
      "download.prompt_for_download": False,
    })
     
    # WebDriverのパスを指定
    driver = webdriver.Edge(service=Service('D:/work/test/edgedriver_win32/msedgedriver.exe'), options=options)
    driver.get("https://test.s3.ap-northeast-1.amazonaws.com/form.html?ABORT_MAIL=false")
    ~~~