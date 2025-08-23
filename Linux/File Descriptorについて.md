# File Descriptor（FD）とは
- https://ja.wikipedia.org/wiki/%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E8%A8%98%E8%BF%B0%E5%AD%90  
  > POSIXでは、全てのプロセスが持つべき3つのfile descriptorを定義している  
  > - 0 : stdin
  > - 1 : stdout
  > - 2 : stderr
  > 
  > 一般にfile descriptorは、オープン中のファイルの詳細を記録するカーネル内データ構造（配列）へのインデックスである。POSIXでは、これをfile descriptor tableと呼び、各プロセスが自身のfile descriptor tableを持つ。ユーザーアプリケーションは抽象キー（＝file descriptor）をシステムコール経由でカーネルに渡し、カーネルはそのキーに対応するファイルにアクセスする。アプリケーション自身はfile descriptor tableを直接読み書きできない。
- OS（主にUNIX/Linux系）における開いている(オープン中の)ファイルを表す整数番号
- ファイルだけでなく、ソケット、パイプなど「入出力に使えるリソース」全般を抽象化して表現
- プログラムがファイルを開いたりソケットを作成したときに、OSは「ハンドル番号（FD）」を返し、以降はその番号を使って read/write などの操作を行う
- 各プロセスが同時に開けるファイルディスクリプタの数には制限がある
- 各プロセスには「File Descriptor Table」があり、最大数はOSや設定（`ulimit -n`で確認可能）で決まっている
- FDを閉じ忘れると「ファイルディスクリプタリーク」が発生し、`Too many open files` エラーになることがある
- 使い終わったら `close(fd);`（C言語）で解放するのが必須