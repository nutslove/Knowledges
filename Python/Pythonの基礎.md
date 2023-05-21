- PythonでもGoと同じように使わない変数を`_`にすることができる
  - e.g.
    ~~~python
    display = []
    for _ in range(len(chosen_word)):
      display.append("_")
    ~~~