## 集合（set）
- **重複しない**要素の集まりを表すデータ型
- ミュータブル（変更可能）な`set`型とイミュータブル（変更不可能）な`frozenset`型がある
- `{}`（※空の場合は dict になる点に注意）または `set()` で作成できる
- 主な操作
  - 要素の追加: `add()`, `update()`
  - 要素の削除: `remove()`, `discard()`, `pop()`
  - 集合演算: 和集合（`|`）、積集合（`&`）、差集合（`-`）、対称差集合（`^`）
- 例  
  ```python
  numbers = {1, 2, 3, 4, 5}
  print(numbers)  # {1, 2, 3, 4, 5}
  numbers.add(6)
  print(numbers)  # {1, 2, 3, 4, 5, 6}
  numbers.remove(3)
  print(numbers)  # {1, 2, 4, 5, 6}
  evens = {2, 4, 6, 8}
  intersection = numbers & evens
  print(intersection)  # {2, 4, 6}

  a = {1, 2, 3}
  b = {3, 4, 5}
  a | b  # 和集合 {1, 2, 3, 4, 5}
  a & b  # 積集合 {3}
  a - b  # 差集合 {1, 2}
  a ^ b  # 対称差集合 {1, 2, 4, 5}
  ```

### `remove()`と`discard()`の違い
- `remove(elem)`: 指定した要素が存在しない場合、`KeyError`例外を発生させる
- `discard(elem)`: 指定した要素が存在しなくても例外を発生させない

```python
s = {1, 2, 3}

s.remove(4)   # KeyError
s.discard(4)  # 何も起きない
```

## listとの違い
| 特徴 | set | list |
| --- | --- | --- |
| 重複要素 | 不可 | 可 |
| 順序 | 無し（順序保証されない） | 有り（挿入順に保持） |
| 主な用途 | 要素の存在確認、集合演算 | 順序付きデータの保持、インデックスアクセス |
| パフォーマンス | 高速な要素存在確認（平均O(1)） | 線形探索（O(n)） |
| ミュータブル/イミュータブル | `set`（ミュータブル）、`frozenset`（イミュータブル） | `list`（ミュータブル）、`tuple`（イミュータブル） |

### 存在チェック
```python
lst = list(range(1000000))
st = set(lst)

10 in lst  # 遅い
10 in st   # 速い
```