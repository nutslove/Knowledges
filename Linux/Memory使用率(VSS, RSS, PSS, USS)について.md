- 参考URL
  - https://www.lineo.co.jp/blog/linux/sol01-processmemory.html

#### 前提知識
- 仮想メモリ
  - HDDやSSDの記憶デバイスの領域を物理メモリの一部であるかのように見せかけて提供される仮想的なメモリ領域
  - 物理メモリにいつでも作業を引き継げるようにスタンバイ
  - 優先度の低い作業をする
  - 作業スピードが遅い
- 物理メモリ
  - **RAMメモリ**のこと
  - 優先するべき作業を行う
  - 作業スピードが早い  

- ページング方式
  - メモリ領域を`ページ`と呼ばれる一定の大きさの領域に分割し管理する方式のこと
  - プログラムを`ページ`という単位に分割して仮想記憶(補助記憶装置)に記憶
  - 必要な時に必要なページだけを補助記憶装置から主記憶装置(メインメモリ)に読み込む  
    → __ページイン__  
  - 主記憶装置(メインメモリ)に空きが無くなったら主記憶装置(メインメモリ)から補助記憶装置にページを追い出して空き領域を確保する  
    → __ページアウト__
  - 参考URL
    - https://medium-company.com/%E3%83%9A%E3%83%BC%E3%82%B8%E3%83%B3%E3%82%B0%E6%96%B9%E5%BC%8F/

- スワッピング ≒ ページング
  - スワッピング
    - **プロセス(プログラム)単位**でメインメモリと補助記憶装置間でやりとり
  - ページング
    - **ページ単位**でメインメモリと補助記憶装置間でやりとり

### VSS (Virtual Set Size)
- プロセスがアクセスできるメモリ領域サイズの総和。VSSには仮想メモリ上にのみ確保している領域も計上されるため、プロセスがまだ使用していないメモリ領域も含まれる。プロセスが実際にどれだけ物理メモリを使用しているかについては、VSSだけでは判別できない。  

### RSS (Resident Set Size)
- プロセスが確保している物理メモリの使用量
- 物理メモリの使用量の指標
> **Note**
> RSSは複数のプロセス間で共有されているメモリ領域も合計して算出されるので、同じライブラリを使う複数のプロセスがあると`プロセス数*ライブラリ容量`が重複されて算出されてしまう

![RSS](https://github.com/nutslove/Knowledges/blob/main/Linux/image/RSS.jpg)  

### PSS (Proportional Set Size)
- RSS のうち共有メモリの使用量をプロセス間で等分することで得られる物理メモリの使用量

### USS (Unique Set Size)
- プロセスが確保している物理メモリ(RSS)のうち、他のどのプロセスとも共有していない領域の合計サイズ

### WSS (Working Set Size)
- WSS is how much memory an application needs to keep working.
  >Your application may have 100 Gbytes of main memory allocated and page mapped, but it is only touching 50 Mbytes each second to do its job. That's the working set size: the "hot" memory that is frequently used. It is useful to know for capacity planning and scalability analysis.
- 参考URL
  - https://www.brendangregg.com/wss.html
