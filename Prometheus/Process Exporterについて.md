- https://github.com/ncabatoff/process-exporter
- 基本的な(default)configの書き方
  - `{{.Comm}}` contains the basename of the original executable, i.e. 2nd field in `/proc/<pid>/stat`
  - `/proc/<pid>/stat`の第2フィールドには最大15文字までしか入らず、15文字を超える部分は切れてしまう
    ~~~yaml
    process_names:
      - name: "{{.Comm}}"
        cmdline:
        - '.+'
    ~~~
- ただ、上記のデフォルトの設定だとjavaやpythonなど、`/proc/<pid>/stat`の第2フィールドは同じ文字列が入っているプロセスは区別がつかないので、そういう場合は`comm`と`cmdline`を組み合わせて特定のプロセスを独立した`groupname`として取得する
  - `comm`
    - `/proc/<pid>/stat`の第2フィールドに入っている文字列
  - `cmdline`
    - コマンド実行時に指定されている引数（e.g. `python3 manage.py`の`manage.py`、`java app.jar`の`app.jar`の部分）
  - `python3 manage.py`と`java app.jar`のプロセスを区別して独立したprocess(メトリクス内の`groupname`ラベル)として扱う例  
    - **デフォルトの以下の設定は一番下に定義すること！**  
      **以下のデフォルトの設定を一番上に書くとその下の`comm`と`cmdline`を組み合わせた設定がうまく動作しない** 
        ~~~yaml
        process_names:
        - name: "{{.Comm}}"
          cmdline:
          - '.+'
        ~~~
    ~~~yaml
    process_names:
    - name: django --> これはgroupnameラベルの値になる
      comm:
      - python3
      cmdline:
      - manage.py
    - name: javaapp --> これはgroupnameラベルの値になる
      comm:
      - java
      cmdline:
      - app.jar
    - name: "{{.Comm}}"
      cmdline:
      - '.+'
    ~~~