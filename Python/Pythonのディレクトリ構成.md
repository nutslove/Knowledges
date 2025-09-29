## よくある一般的なディレクトリ構成例  
  ```bash
my_project/
├── src/                # 本体コード
│   ├── __init__.py
│   └── my_module.py
├── tests/              # テストコード
│   ├── __init__.py     # 空でOK（必須ではない）
│   ├── test_my_module.py
│   └── conftest.py     # fixture定義用
├── pyproject.toml or setup.cfg or pytest.ini
└── requirements.txt
```
