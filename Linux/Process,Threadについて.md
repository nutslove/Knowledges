## ProcessとThreadの違い
- __Process__
  - 実行中のプログラム

- __Thread__
  - ここで言うThreadはSoftware観点のThread
  - Processの中の実行単位(Flow)
  - 1つのProcessの中に1つ以上のThreadが存在する
    - 1つのProcessの中に2つ以上のThreadが存在する場合、`Multi Threading`という
    - 各Threadは同時に独立して動く
- Process間はMemory spaceを共有しない
  - 各Processは独立したMemory space上で動いて、他のProcessのMemory spaceには干渉できない
- 同じProcess内のThread間はMemory space(Heap)を共有する
  - ThreadsはProcessが持つMemory spaceしか使うことができない
- 参考URL
  - https://stackoverflow.com/questions/34689709/java-threads-and-number-of-cores
  - https://www.geeksforgeeks.org/difference-between-java-threads-and-os-threads/
  - https://www.youtube.com/watch?v=x-Lp-h_pf9Q&t=918s

### 各Threadで共有するもの、個別に持つもの
#### 各Threadで共有するもの
- **メモリ空間**
  - ヒープ領域（動的に確保されるメモリ領域）、グローバル変数、静的変数などを共有
- **ファイルハンドラ**
  - プロセスが開いているファイルやソケットの情報を共有

#### 各Threadごとに個別に持つもの
- **スタック領域**（以下はスタック領域に格納されるもの）
  - 関数のローカル変数の値
  - 関数の戻りアドレス（呼び出し元の関数に戻るためのメモリ位置）
  - 関数パラメータ（引数）の値
  - etc.
- **レジスタ**（CPU内の一時的なデータ格納場所）
- **プログラムカウンタ**（次に実行する命令のアドレスを指すポインタ）

---

## Hardware(CPU) ThreadとSoftware(Program) Threadについて
- Hardware(CPU) Thread
  - CPUが命令を実行できる単位
  - 例えばCPU Threadが2つあるCPU Coreは同時に2つの命令を実行できる
- Software(Program) Thread
  - 上で書いた通り、1つのProgramで同時に実行できる独立して実行される実行単位(Flow)
  - 上記の同時とはCPU Thread数によって、**並列処理(parallelism)** または **並行処理(concurrency)** になる
    - 例えばCPU Threadが2つあって、Programの中のThreadが3つある場合、  
      2つのProgramの中のThreadは並列に実行されるけど、残り1つはContext Switchingで入れ替えされるので  
      全体の観点で見るとこのProgramは並行処理と言える