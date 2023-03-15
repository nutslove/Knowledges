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