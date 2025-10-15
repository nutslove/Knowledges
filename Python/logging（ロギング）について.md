# Pythonの主要 logging（ロギング）ライブラリ
## １．`logging` ライブラリ
- Pythonの標準ライブラリで、ログの出力や管理を行うための機能を提供。

### `logging` の基本的な使い方
```python
import logging

# 基本的なログ設定
logging.basicConfig(level=logging.DEBUG)

# ログの出力
logging.debug('デバッグメッセージ')
logging.info('情報メッセージ')
logging.warning('警告メッセージ')
logging.error('エラーメッセージ')
logging.critical('致命的なエラー')
```

## ２．`loguru` ライブラリ
- https://github.com/Delgan/loguru
- `loguru`は、Pythonのサードパーティ製のロギングライブラリで、使いやすさと柔軟性を提供。