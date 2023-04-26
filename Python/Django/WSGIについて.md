- 参考URL
  - https://medium.com/analytics-vidhya/what-is-wsgi-web-server-gateway-interface-ed2d290449e

## WSGIとは
- Web Server(e.g. Nginx/Apache)とPythonのWebフレームワーク(e.g. Django/Flask)の間に位置し、  
  Web ServerとPython Webフレームワークがコミュニケーションできるようにするインターフェース
- Web Server Gateway Interfaceの頭文字
- **Python Webフレームワークの前段にWeb Serverを配置しない場合は不要**
- WSGIのツールには以下のようなものがある
  - Gunicorn
    - https://gunicorn.org/#quickstart
  - uWSGI
    - https://uwsgi-docs.readthedocs.io/en/latest/