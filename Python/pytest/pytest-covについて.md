# `pytest-cov` (pytest coverage)
- テストがコードのどれくらいをカバーしているか（カバレッジ）を測定するpytestプラグイン

## 基本的な使い方
```bash
# インストール
pip install pytest-cov

# 実行（myappパッケージのカバレッジを測定）
pytest --cov=myapp

# HTMLレポートを生成
pytest --cov=myapp --cov-report=html

## 実行結果の例
---------- coverage: platform linux, python 3.11.0 -----------
Name                      Stmts   Miss  Cover
---------------------------------------------
myapp/__init__.py             2      0   100%
myapp/user_service.py        15      3    80%
myapp/order_service.py       20     10    50%
---------------------------------------------
TOTAL                        37     13    65%
```