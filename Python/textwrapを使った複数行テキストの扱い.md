## 概要
- `"""`を使って複数行のテキストを変数に代入する場合、関数やクラスの中でコードのインデントに合わせて書くと、そのインデントも文字列に含まれてしまう。しかし、`textwrap.dedent()`を使えば、コード上ではインデントを揃えて書いても、実際の出力時にはインデントが除去される。

### 例
```python
import textwrap

def example():
    # 問題のあるパターン
    bad = """これは
    複数行の
    テキストです"""

    # textwrap.dedent()を使った解決策
    good = textwrap.dedent("""\
        これは
        複数行の
        テキストです
        """).strip()

    # textwrapを使わずに行の始めのスペースなしで出力する方法
    good_without_textwrap = """これは
複数行の
テキストです"""

    print("Bad:")
    print(bad)
    print("\nGood:")
    print(good)
    print("\nGood without textwrap:")
    print(good_without_textwrap)

example()
```
- 出力  
  ```shell
  Bad:
  これは
      複数行の
      テキストです

  Good:
  これは
  複数行の
  テキストです

  Good without textwrap:
  これは
  複数行の
  テキストです
  ```