## `time.Unix()`の基本
- フォーマット
  - `func Unix(sec int64, nsec int64) Time`
- Unixタイムスタンプ（エポック時間）からGoの`Time`型を生成する関数
- パラメータ
  - `sec`: 秒数 (1970年1月1日からの経過秒数)
  - `nsec`: ナノ秒数 (秒の小数部分をナノ秒で指定)
- 例  
  1.  
  ```go
  // 1673424000は2023年1月11日12:00:00 UTC
  t := time.Unix(1673424000, 0)
  fmt.Println(t) // 2023-01-11 12:00:00 +0000 UTC
  ```

  2.  
  ```go
  // 現在のUnixタイムを取得して変換する例
  now := time.Now().Unix()      // 現在のUnixタイム（秒）
  t := time.Unix(now, 0)
  fmt.Println(Format("2006-01-02 15:04:05"))        // 読みやすい形式に整形（+0000 UTCの部分がない）
  ```