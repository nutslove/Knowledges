# 注意点
## 1. `--label`によるExternal Labelsの設定
- Receiverで最低限１つ以上の`--label`フラグでExternal Labelsを設定することが必須。１もない場合は、Receiver起動時にエラーになる。

> [!IMPORTANT]  
> - receiverの起動時に`--label`フラグを使用して、各Receiverに一意のExternal Label（例: Pod名）を設定することが重要。
> - これにより、各Receiverが異なるExternal Labelsを持つようになり、CompactorがBlock overlapを防止できる。
> - 詳細については同じディレクトリの「Thanosについて.md」の「### Overlaps」セクションを参照！